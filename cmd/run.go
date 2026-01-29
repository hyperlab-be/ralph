package cmd

import (
	"bufio"
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
  - Read the PRD and find the next incomplete story
  - Work on implementing the story
  - Run tests and verify acceptance criteria
  - Commit changes and mark the story complete
  - Move to the next story`,
	RunE: runAgent,
}

var maxIterations int
var dryRun bool

func init() {
	runCmd.Flags().IntVarP(&maxIterations, "max-iterations", "m", 10, "Maximum iterations")
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without executing")
	rootCmd.AddCommand(runCmd)
}

func runAgent(cmd *cobra.Command, args []string) error {
	// Find project root
	cwd, _ := os.Getwd()
	projectRoot, err := config.FindProjectRoot(cwd)
	if err != nil {
		return fmt.Errorf("not in a rl project")
	}

	worktreeName := filepath.Base(projectRoot)

	// Load PRD
	p, err := prd.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load PRD: %w", err)
	}
	if p == nil {
		return fmt.Errorf("no PRD found. Create one with 'rl prd create'")
	}

	// Check if already running
	loop, _ := config.GetLoop(worktreeName)
	if loop != nil && loop.Status == "running" {
		return fmt.Errorf("loop is already running")
	}

	// Load config
	cfg, _ := config.LoadProjectConfig(projectRoot)
	model := "claude-sonnet-4-20250514"
	if cfg != nil && cfg.Agent.Model != "" {
		model = cfg.Agent.Model
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

	// Session log
	sessionLog := filepath.Join(projectRoot, ".rl", "session.log")
	logFile, _ := os.OpenFile(sessionLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer logFile.Close()

	fmt.Fprintf(logFile, "\n=== Session started %s ===\n", time.Now().Format(time.RFC3339))

	// Main loop
	for iteration := 1; iteration <= maxIterations; iteration++ {
		select {
		case <-ctx.Done():
			break
		default:
		}

		// Get current story
		story := p.GetCurrentStory()
		if story == nil {
			printSuccess("All stories complete!")
			break
		}

		fmt.Println()
		printInfo(fmt.Sprintf("Iteration %d/%d: Story %s - %s", iteration, maxIterations, story.ID, story.Title))
		fmt.Fprintf(logFile, "[%s] Iteration %d: %s\n", time.Now().Format("15:04:05"), iteration, story.Title)

		// Run agent iteration
		err := runAgentIteration(ctx, projectRoot, story, logFile)
		if err != nil {
			if ctx.Err() != nil {
				break // Interrupted
			}
			printError(fmt.Sprintf("Agent iteration failed: %v", err))
			fmt.Fprintf(logFile, "[%s] Error: %v\n", time.Now().Format("15:04:05"), err)
			continue
		}

		// Reload PRD to check if story completed
		p, _ = prd.Load(projectRoot)
		if p != nil {
			updatedStory := findStory(p, story.ID)
			if updatedStory != nil && updatedStory.Passes {
				printSuccess(fmt.Sprintf("Story %s completed!", story.ID))
				fmt.Fprintf(logFile, "[%s] Story %s completed\n", time.Now().Format("15:04:05"), story.ID)
			}
		}
	}

	// Update loop status
	loop.Status = "stopped"
	loop.Stopped = time.Now().Format(time.RFC3339)
	loop.PID = 0
	config.SetLoop(loop)

	fmt.Fprintf(logFile, "=== Session ended %s ===\n", time.Now().Format(time.RFC3339))

	// Final status
	if p != nil {
		printInfo(fmt.Sprintf("Final progress: %s stories", p.Progress()))
	}

	return nil
}

func runAgentIteration(ctx context.Context, projectRoot string, story *prd.Story, logFile *os.File) error {
	// Build task description
	var criteria strings.Builder
	for _, c := range story.AcceptanceCriteria {
		criteria.WriteString("- ")
		criteria.WriteString(c)
		criteria.WriteString("\n")
	}

	task := fmt.Sprintf(`You are working in: %s

## Current Story: %s

### Description
%s

### Acceptance Criteria
%s

### Instructions
1. Implement the story requirements
2. Write tests to verify the acceptance criteria
3. Run tests and ensure they pass
4. Commit changes with message: feat(story-%s): %s
5. Update .rl/prd.json to set passes: true for story %s

When complete, the story's "passes" field in .rl/prd.json must be true.
`, projectRoot, story.Title, story.Description, criteria.String(), story.ID, story.Title, story.ID)

	// Check for Claude CLI
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("Claude CLI not found. Install with: npm install -g @anthropic-ai/claude-code")
	}

	// Run Claude
	cmd := exec.CommandContext(ctx, claudePath, "--print", task)
	cmd.Dir = projectRoot
	cmd.Env = os.Environ()

	// Capture output
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Claude: %w", err)
	}

	// Stream output
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
			logFile.WriteString(line + "\n")
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprintln(os.Stderr, line)
			logFile.WriteString("[ERR] " + line + "\n")
		}
	}()

	return cmd.Wait()
}

func findStory(p *prd.PRD, id string) *prd.Story {
	for i := range p.UserStories {
		if p.UserStories[i].ID == id {
			return &p.UserStories[i]
		}
	}
	return nil
}
