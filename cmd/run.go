package cmd

import (
	"context"
	"fmt"
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
  - Move to the next story`,
	RunE: runAgent,
}

var (
	maxIterations int
	model         string
	dryRun        bool
	once          bool
)

func init() {
	runCmd.Flags().IntVarP(&maxIterations, "max-iterations", "m", 10, "Maximum iterations")
	runCmd.Flags().StringVar(&model, "model", "opus", "Model to use (opus, sonnet, etc)")
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without executing")
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
		return fmt.Errorf("no PRD found. Create one with 'ralph prd'")
	}

	// Check if already running
	loop, _ := config.GetLoop(worktreeName)
	if loop != nil && loop.Status == "running" {
		return fmt.Errorf("loop is already running")
	}

	// --once overrides max-iterations
	if once {
		maxIterations = 1
	}

	printInfo(fmt.Sprintf("Starting agent loop for %s", worktreeName))
	printInfo(fmt.Sprintf("Model: %s | Max iterations: %d", model, maxIterations))

	if dryRun {
		printWarn("Dry run mode - not executing")
		story := p.GetCurrentStory()
		if story != nil {
			fmt.Printf("\nWould work on: %s. %s\n", story.ID, story.Title)
		}
		return nil
	}

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

	// Live output log (streamed, for ralph logs -f)
	// Truncate at start of new loop so logs only show current session
	outputLog := filepath.Join(projectRoot, ".ralph", "output.log")
	outputFile, _ := os.OpenFile(outputLog, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	defer outputFile.Close()

	fmt.Fprintf(logFile, "\n=== Session started %s ===\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(logFile, "Model: %s\n", model)
	fmt.Fprintf(outputFile, "\n%s\n", strings.Repeat("═", 60))
	fmt.Fprintf(outputFile, "Session started: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(outputFile, "%s\n\n", strings.Repeat("═", 60))

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

		fmt.Fprintf(logFile, "[%s] Iteration %d started\n", time.Now().Format("15:04:05"), iteration)

		// Write to live output log
		fmt.Fprintf(outputFile, "━━━ Iteration %d/%d ━━━\n", iteration, maxIterations)
		fmt.Fprintf(outputFile, "Progress: %s | Story: %s\n\n", p.Progress(), p.CurrentStory())
		outputFile.Sync()

		// Run agent iteration
		err = runAgentIteration(ctx, projectRoot, p, outputFile)

		// Reload to get updated progress
		p, _ = prd.Load(projectRoot)
		progressAfter := "unknown"
		if p != nil {
			progressAfter = p.Progress()
		}

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

Now read .ralph/prd.json and .ralph/progress.txt, then begin work.
`, projectRoot, p.Name, p.Description, storiesList.String())
}

func runAgentIteration(ctx context.Context, projectRoot string, p *prd.PRD, outputLog *os.File) error {
	prompt := buildAgentPrompt(projectRoot, p)

	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude CLI not found")
	}

	// Always run with --dangerously-skip-permissions for autonomous operation
	cmd := exec.CommandContext(ctx, claudePath,
		"--dangerously-skip-permissions",
		"--model", model)

	cmd.Dir = projectRoot
	cmd.Env = os.Environ()

	// Set up stdin pipe for prompt
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Set up stdout/stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	// Write prompt to stdin and close
	go func() {
		stdin.Write([]byte(prompt))
		stdin.Close()
	}()

	// Read and stream output to stdout and output log
	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				os.Stdout.Write(buf[:n])
				outputLog.Write(buf[:n])
				outputLog.Sync() // Flush for live tailing
			}
			if err != nil {
				break
			}
		}
	}()

	// Also capture stderr
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				os.Stderr.Write(buf[:n])
				outputLog.Write(buf[:n])
				outputLog.Sync()
			}
			if err != nil {
				break
			}
		}
	}()

	// Wait for command to finish
	err = cmd.Wait()

	// Wait for stdout reader to finish
	<-done

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
