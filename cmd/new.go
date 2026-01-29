package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hyperlab/ralph/internal/config"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:     "new <feature>",
	Aliases: []string{"n"},
	Short:   "Create a new worktree for a feature",
	Long: `Create a new git worktree for developing a feature.

This will:
  - Create a git worktree with a feature branch
  - Copy project configuration
  - Run setup hooks (if configured)
  - Register the loop`,
	Args: cobra.ExactArgs(1),
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	feature := args[0]

	// Find project root
	cwd, _ := os.Getwd()
	projectRoot, err := config.FindProjectRoot(cwd)
	if err != nil {
		return fmt.Errorf("not in a rl project. Run 'rl init' first")
	}

	// Load project config
	cfg, err := config.LoadProjectConfig(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load project config: %w", err)
	}

	projectName := filepath.Base(projectRoot)
	if cfg != nil && cfg.Project.Name != "" {
		projectName = cfg.Project.Name
	}

	worktreeName := fmt.Sprintf("%s-%s", projectName, feature)
	worktreePath := filepath.Join(filepath.Dir(projectRoot), worktreeName)
	branch := fmt.Sprintf("feature/%s", feature)

	// Check if worktree exists
	if _, err := os.Stat(worktreePath); err == nil {
		return fmt.Errorf("worktree already exists: %s", worktreePath)
	}

	printInfo(fmt.Sprintf("Creating worktree: %s", worktreeName))

	// Create git worktree
	gitCmd := exec.Command("git", "worktree", "add", worktreePath, "-b", branch)
	gitCmd.Dir = projectRoot
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr

	if err := gitCmd.Run(); err != nil {
		// Branch might exist, try without -b
		gitCmd = exec.Command("git", "worktree", "add", worktreePath, branch)
		gitCmd.Dir = projectRoot
		gitCmd.Stdout = os.Stdout
		gitCmd.Stderr = os.Stderr

		if err := gitCmd.Run(); err != nil {
			return fmt.Errorf("failed to create worktree: %w", err)
		}
	}

	printSuccess(fmt.Sprintf("Worktree created at %s", worktreePath))

	// Copy rl.toml if exists
	srcConfig := filepath.Join(projectRoot, "rl.toml")
	if _, err := os.Stat(srcConfig); err == nil {
		dstConfig := filepath.Join(worktreePath, "rl.toml")
		data, _ := os.ReadFile(srcConfig)
		os.WriteFile(dstConfig, data, 0644)
	}

	// Create .rl directory
	rlDir := filepath.Join(worktreePath, ".rl")
	os.MkdirAll(rlDir, 0755)

	// Initialize progress file
	progressContent := `# Progress Log

This file tracks progress across iterations.

## Patterns & Learnings

*Add reusable patterns discovered during development here.*

---

## Session Log

`
	os.WriteFile(filepath.Join(rlDir, "progress.md"), []byte(progressContent), 0644)

	// Run setup hook if defined
	if cfg != nil && cfg.Hooks.Setup != "" {
		printInfo("Running setup hook...")

		// Set environment variables
		dbName := fmt.Sprintf("%s_%s", strings.ReplaceAll(projectName, "-", "_"), strings.ReplaceAll(feature, "-", "_"))
		env := os.Environ()
		env = append(env, fmt.Sprintf("DB_NAME=%s", dbName))
		env = append(env, fmt.Sprintf("WORKTREE_PATH=%s", worktreePath))
		env = append(env, fmt.Sprintf("FEATURE=%s", feature))

		hookCmd := exec.Command("bash", "-c", cfg.Hooks.Setup)
		hookCmd.Dir = worktreePath
		hookCmd.Env = env
		hookCmd.Stdout = os.Stdout
		hookCmd.Stderr = os.Stderr

		if err := hookCmd.Run(); err != nil {
			printWarn(fmt.Sprintf("Setup hook failed: %v", err))
		}
	}

	// Register loop
	loop := &config.Loop{
		Name:    worktreeName,
		Path:    worktreePath,
		Project: projectName,
		Feature: feature,
		Branch:  branch,
		Status:  "created",
		Created: time.Now().Format(time.RFC3339),
	}

	if err := config.SetLoop(loop); err != nil {
		printWarn(fmt.Sprintf("Failed to register loop: %v", err))
	}

	printSuccess(fmt.Sprintf("Ready! cd %s", worktreePath))
	printInfo("Next: Create a PRD with 'rl prd create' then start with 'rl run'")

	return nil
}
