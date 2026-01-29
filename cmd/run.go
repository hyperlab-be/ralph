package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/hyperlab-be/ralph/internal/config"
	"github.com/hyperlab-be/ralph/internal/prd"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:     "run",
	Aliases: []string{"r"},
	Short:   "Start the AI agent loop",
	Long: `Start the AI agent loop for the current project.

The agent will:
  - Read the PRD and choose the highest priority incomplete story
  - Work on implementing the story
  - Run tests and verify acceptance criteria
  - Commit changes and mark the story complete
  - Move to the next story

Sandbox modes provide isolation for safe operation:
  - docker: Docker sandbox (default, recommended)
  - mac:    macOS sandbox-exec (experimental)`,
	RunE: runAgent,
}

var (
	maxIterations int
	dryRun        bool
	sandbox       string
	once          bool
)

func init() {
	runCmd.Flags().IntVarP(&maxIterations, "max-iterations", "m", 10, "Maximum iterations")
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without executing")
	runCmd.Flags().StringVarP(&sandbox, "sandbox", "s", "docker", "Sandbox mode: docker, mac")
	runCmd.Flags().BoolVar(&once, "once", false, "Run single iteration (HITL mode)")
	rootCmd.AddCommand(runCmd)
}

func runAgent(cmd *cobra.Command, args []string) error {
	// Find project root
	cwd, _ := os.Getwd()
	projectRoot, err := config.FindProjectRoot(cwd)
	if err != nil {
		return fmt.Errorf("not in a ralph project")
	}

	worktreeName := filepath.Base(projectRoot)

	// Load PRD
	p, err := prd.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load PRD: %w", err)
	}
	if p == nil {
		return fmt.Errorf("no PRD found. Create one with 'ralph prd create'")
	}

	// Check if already running
	loop, _ := config.GetLoop(worktreeName)
	if loop != nil && loop.Status == "running" {
		return fmt.Errorf("loop is already running")
	}

	// Load config (project config overrides global config)
	globalCfg, _ := config.LoadGlobalConfig()
	projectCfg, _ := config.LoadProjectConfig(projectRoot)

	model := "claude-sonnet-4-20250514" // ultimate fallback
	if globalCfg != nil && globalCfg.Defaults.Model != "" {
		model = globalCfg.Defaults.Model
	}
	if projectCfg != nil && projectCfg.Agent.Model != "" {
		model = projectCfg.Agent.Model
	}

	// Use config max_iterations if flag wasn't explicitly set
	if !cmd.Flags().Changed("max-iterations") {
		if projectCfg != nil && projectCfg.Agent.MaxIterations > 0 {
			maxIterations = projectCfg.Agent.MaxIterations
		} else if globalCfg != nil && globalCfg.Defaults.MaxIterations > 0 {
			maxIterations = globalCfg.Defaults.MaxIterations
		}
	}

	// --once overrides max-iterations
	if once {
		maxIterations = 1
	}

	// Validate sandbox
	if sandbox != "docker" && sandbox != "mac" {
		return fmt.Errorf("invalid sandbox mode: %s (use: docker, mac)", sandbox)
	}

	// Check docker sandbox availability
	if sandbox == "docker" {
		if err := checkDockerSandbox(); err != nil {
			return err
		}
	}

	printInfo(fmt.Sprintf("Starting agent loop for %s", worktreeName))
	printInfo(fmt.Sprintf("Model: %s | Max iterations: %d | Sandbox: %s", model, maxIterations, sandbox))

	if dryRun {
		printWarn("Dry run mode - not executing")
		story := p.GetCurrentStory()
		if story != nil {
			fmt.Printf("\nWould work on: %s. %s\n", story.ID, story.Title)
		}
		return nil
	}

	// Create conversations directory
	conversationsDir := filepath.Join(projectRoot, ".ralph", "conversations")
	os.MkdirAll(conversationsDir, 0755)

	// Update loop status
	if loop == nil {
		loop = &config.Loop{
			Name:   worktreeName,
			Path:   projectRoot,
			Status: "running",
		}
	}
	loop.Status = "running"
	loop.Started = time.Now().Format(time.RFC3339)
	loop.PID = os.Getpid()
	config.SetLoop(loop)

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		printWarn("\nReceived interrupt, stopping...")
		cancel()
	}()

	// Session log (summary)
	sessionLog := filepath.Join(projectRoot, ".ralph", "session.log")
	logFile, _ := os.OpenFile(sessionLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer logFile.Close()

	fmt.Fprintf(logFile, "\n=== Session started %s ===\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(logFile, "Sandbox: %s\n", sandbox)

	// Main loop
	for iteration := 1; iteration <= maxIterations; iteration++ {
		select {
		case <-ctx.Done():
			break
		default:
		}

		// Reload PRD each iteration (agent may have updated it)
		p, _ = prd.Load(projectRoot)
		if p == nil || p.IsComplete() {
			printSuccess("All stories complete!")
			break
		}

		fmt.Println()
		fmt.Println(strings.Repeat("━", 60))
		printInfo(fmt.Sprintf("Iteration %d/%d", iteration, maxIterations))
		printInfo(fmt.Sprintf("Progress: %s", p.Progress()))
		fmt.Println(strings.Repeat("━", 60))

		// Create conversation log for this iteration
		convLogPath := filepath.Join(conversationsDir, fmt.Sprintf("iteration-%d.md", iteration))
		convLog, err := os.Create(convLogPath)
		if err != nil {
			printError(fmt.Sprintf("Failed to create conversation log: %v", err))
			continue
		}

		// Write conversation header
		fmt.Fprintf(convLog, "# Iteration %d\n\n", iteration)
		fmt.Fprintf(convLog, "**Started:** %s\n", time.Now().Format(time.RFC3339))
		fmt.Fprintf(convLog, "**Sandbox:** %s\n", sandbox)
		fmt.Fprintf(convLog, "**Progress before:** %s\n\n", p.Progress())

		fmt.Fprintf(logFile, "[%s] Iteration %d started\n", time.Now().Format("15:04:05"), iteration)

		// Run agent iteration
		err = runAgentIteration(ctx, projectRoot, p, sandbox, convLog)
		
		// Write conversation footer
		p, _ = prd.Load(projectRoot) // Reload to get updated progress
		progressAfter := "unknown"
		if p != nil {
			progressAfter = p.Progress()
		}
		fmt.Fprintf(convLog, "\n\n---\n")
		fmt.Fprintf(convLog, "**Ended:** %s\n", time.Now().Format(time.RFC3339))
		fmt.Fprintf(convLog, "**Progress after:** %s\n", progressAfter)
		convLog.Close()

		if err != nil {
			if ctx.Err() != nil {
				break // Interrupted
			}
			printError(fmt.Sprintf("Agent iteration failed: %v", err))
			fmt.Fprintf(logFile, "[%s] Error: %v\n", time.Now().Format("15:04:05"), err)
			continue
		}

		fmt.Fprintf(logFile, "[%s] Iteration %d completed, progress: %s\n", 
			time.Now().Format("15:04:05"), iteration, progressAfter)

		// Brief pause between iterations (unless single iteration)
		if iteration < maxIterations && !once {
			printInfo("Pausing 5s before next iteration...")
			time.Sleep(5 * time.Second)
		}
	}

	// Update loop status
	loop.Status = "stopped"
	loop.Stopped = time.Now().Format(time.RFC3339)
	loop.PID = 0
	config.SetLoop(loop)

	fmt.Fprintf(logFile, "=== Session ended %s ===\n", time.Now().Format(time.RFC3339))

	// Final status
	p, _ = prd.Load(projectRoot)
	if p != nil {
		fmt.Println()
		fmt.Println(strings.Repeat("━", 60))
		printInfo(fmt.Sprintf("Final progress: %s", p.Progress()))
		fmt.Println(strings.Repeat("━", 60))

		// Create PR if all stories complete
		if p.IsComplete() {
			printSuccess("All stories complete! Creating pull request...")
			if err := createPullRequest(projectRoot, p); err != nil {
				printWarn(fmt.Sprintf("Failed to create PR: %v", err))
			}
		}
	}

	return nil
}

func checkDockerSandbox() error {
	cmd := exec.Command("docker", "sandbox", "--help")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker sandbox not available. Install Docker Desktop with AI features enabled")
	}
	return nil
}

func createPullRequest(projectRoot string, p *prd.PRD) error {
	// Check if gh is available
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not found - install from https://cli.github.com")
	}

	// Get current branch
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchCmd.Dir = projectRoot
	branchOut, err := branchCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get branch: %w", err)
	}
	branch := strings.TrimSpace(string(branchOut))

	// Don't create PR from main/master
	if branch == "main" || branch == "master" {
		return fmt.Errorf("cannot create PR from %s branch", branch)
	}

	// Check for uncommitted changes and commit them
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = projectRoot
	statusOut, _ := statusCmd.Output()
	if len(statusOut) > 0 {
		// Add tracked files only (excludes .ralph/, prd.json if in .gitignore)
		addCmd := exec.Command("git", "add", "-u")
		addCmd.Dir = projectRoot
		addCmd.Run()

		// Also add new files except ralph artifacts
		addNewCmd := exec.Command("git", "add", "--all", "--", ".", ":!.ralph/", ":!.ralph-tui/", ":!.rl/", ":!prd.json", ":!.ralph-*")
		addNewCmd.Dir = projectRoot
		addNewCmd.Run()

		commitCmd := exec.Command("git", "commit", "-m", fmt.Sprintf("feat: complete %s", p.Name))
		commitCmd.Dir = projectRoot
		commitCmd.Run()
	}

	// Push branch
	printInfo("Pushing branch...")
	pushCmd := exec.Command("git", "push", "-u", "origin", branch)
	pushCmd.Dir = projectRoot
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	// Build PR body
	var body strings.Builder
	body.WriteString(fmt.Sprintf("## %s\n\n", p.Name))
	if p.Description != "" {
		body.WriteString(p.Description)
		body.WriteString("\n\n")
	}
	body.WriteString("## Stories completed\n")
	for _, story := range p.UserStories {
		body.WriteString(fmt.Sprintf("- ✅ %s\n", story.Title))
	}
	body.WriteString("\n_Generated by ralph_ 🤖")

	// Create PR
	printInfo("Creating pull request...")
	prCmd := exec.Command("gh", "pr", "create",
		"--title", p.Name,
		"--body", body.String(),
	)
	prCmd.Dir = projectRoot
	prCmd.Stdout = os.Stdout
	prCmd.Stderr = os.Stderr

	if err := prCmd.Run(); err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	printSuccess("Pull request created!")
	return nil
}

// buildAgentPrompt creates the prompt that lets the agent choose and implement a story
func buildAgentPrompt(projectRoot string, p *prd.PRD) string {
	// Build stories list
	var storiesList strings.Builder
	for _, story := range p.UserStories {
		status := "⬜ INCOMPLETE"
		if story.Passes {
			status = "✅ COMPLETE"
		}
		storiesList.WriteString(fmt.Sprintf("- [%s] %s: %s\n", story.ID, status, story.Title))
		if story.Description != "" {
			storiesList.WriteString(fmt.Sprintf("  Description: %s\n", story.Description))
		}
		if len(story.AcceptanceCriteria) > 0 {
			storiesList.WriteString("  Criteria:\n")
			for _, c := range story.AcceptanceCriteria {
				storiesList.WriteString(fmt.Sprintf("    - %s\n", c))
			}
		}
	}

	return fmt.Sprintf(`You are an autonomous coding agent working through a PRD (Product Requirement Document).

## Working Directory
%s

## PRD: %s
%s

## User Stories
%s

## Your Task
1. Review the PRD and choose the HIGHEST PRIORITY incomplete story (passes: false)
   - Prioritize: architectural decisions > integrations > core features > polish
   - NOT necessarily the first in the list - use your judgment

2. Implement that ONE story fully:
   - Write clean, production-quality code
   - Follow existing patterns in the codebase
   - Write tests to verify acceptance criteria
   - Run all feedback loops (tests, types, lint)

3. After implementation:
   - Run tests and fix any failures
   - Commit changes with message: feat(story-ID): description
   - Update .ralph/prd.json to set passes: true for the completed story

4. Append to .ralph/progress.txt:
   - Story completed
   - Key decisions made
   - Files changed
   - Any notes for next iteration

## Rules
- Work on ONE story per iteration
- Do NOT commit if tests fail
- Be thorough - a story is only "done" when fully working
- If blocked, document in progress.txt and move to next story

## Completion Check
If ALL stories have passes: true, output exactly:
<promise>COMPLETE</promise>

Now read .ralph/prd.json and .ralph/progress.txt, then begin work.
`, projectRoot, p.Name, p.Description, storiesList.String())
}

func runAgentIteration(ctx context.Context, projectRoot string, p *prd.PRD, sandboxMode string, convLog *os.File) error {
	prompt := buildAgentPrompt(projectRoot, p)

	// Write prompt to conversation log
	fmt.Fprintf(convLog, "## Prompt\n\n```\n%s\n```\n\n", prompt)
	fmt.Fprintf(convLog, "## Agent Output\n\n```\n")

	var cmd *exec.Cmd

	switch sandboxMode {
	case "docker":
		// Docker sandbox - fully isolated, safe for AFK
		printInfo("[Docker Sandbox]")
		cmd = exec.CommandContext(ctx, "docker", "sandbox", "run", "claude", ".",
			"--", "--print", "--dangerously-skip-permissions", "-p", prompt)

	case "mac":
		// macOS sandbox - experimental
		printInfo("[macOS Sandbox]")
		claudePath, err := exec.LookPath("claude")
		if err != nil {
			return fmt.Errorf("claude CLI not found")
		}
		cmd = exec.CommandContext(ctx, "sandbox-exec", "-p", "(version 1)(allow default)",
			claudePath, "--print", "--dangerously-skip-permissions", "-p", prompt)

	default:
		return fmt.Errorf("invalid sandbox mode: %s", sandboxMode)
	}

	cmd.Dir = projectRoot
	cmd.Env = os.Environ()

	// Capture output
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	// Stream output to terminal and log
	go func() {
		reader := bufio.NewReader(stdout)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "stdout read error: %v\n", err)
				}
				break
			}
			fmt.Print(line)
			convLog.WriteString(line)
		}
	}()

	go func() {
		reader := bufio.NewReader(stderr)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "stderr read error: %v\n", err)
				}
				break
			}
			fmt.Fprint(os.Stderr, line)
			convLog.WriteString("[ERR] " + line)
		}
	}()

	err := cmd.Wait()
	fmt.Fprintf(convLog, "```\n")
	
	return err
}

func findStory(p *prd.PRD, id string) *prd.Story {
	for i := range p.UserStories {
		if p.UserStories[i].ID == id {
			return &p.UserStories[i]
		}
	}
	return nil
}
