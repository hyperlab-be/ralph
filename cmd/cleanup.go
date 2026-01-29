package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hyperlab/ralph/internal/config"
	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:     "cleanup [feature]",
	Aliases: []string{"clean", "rm"},
	Short:   "Remove a worktree and clean up",
	Long: `Remove a worktree and run cleanup hooks.

This will:
  - Run cleanup hooks (if configured)
  - Remove the git worktree
  - Delete the feature branch (optional)
  - Unregister the loop`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCleanup,
}

var forceCleanup bool
var deleteBranch bool

func init() {
	cleanupCmd.Flags().BoolVarP(&forceCleanup, "force", "f", false, "Skip confirmation")
	cleanupCmd.Flags().BoolVarP(&deleteBranch, "delete-branch", "d", false, "Also delete the feature branch")
	rootCmd.AddCommand(cleanupCmd)
}

func runCleanup(cmd *cobra.Command, args []string) error {
	var worktreePath string
	var worktreeName string
	var loop *config.Loop

	if len(args) > 0 {
		feature := args[0]

		// Try to find the loop
		loop, _ = config.GetLoop(feature)
		if loop != nil {
			worktreePath = loop.Path
			worktreeName = loop.Name
		} else {
			// Try to construct path
			cwd, _ := os.Getwd()
			projectRoot, err := config.FindProjectRoot(cwd)
			if err != nil {
				return fmt.Errorf("not in a rl project")
			}

			projectName := filepath.Base(projectRoot)
			projectName = strings.Split(projectName, "-")[0] // Remove feature suffix if in worktree

			worktreeName = fmt.Sprintf("%s-%s", projectName, feature)
			worktreePath = filepath.Join(filepath.Dir(projectRoot), worktreeName)

			loop, _ = config.GetLoop(worktreeName)
		}
	} else {
		// Use current directory
		cwd, _ := os.Getwd()
		worktreePath = cwd
		worktreeName = filepath.Base(cwd)
		loop, _ = config.GetLoop(worktreeName)
	}

	// Verify worktree exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		return fmt.Errorf("worktree not found: %s", worktreePath)
	}

	// Confirmation
	if !forceCleanup {
		fmt.Println("\033[33mThis will remove:\033[0m")
		fmt.Printf("  - Worktree: %s\n", worktreePath)
		if loop != nil {
			fmt.Printf("  - Branch: %s\n", loop.Branch)
		}
		fmt.Println()
		fmt.Print("Are you sure? (y/N) ")

		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(response)) != "y" {
			printInfo("Cancelled")
			return nil
		}
	}

	// Run cleanup hook if defined
	cfg, _ := config.LoadProjectConfig(worktreePath)
	if cfg != nil && cfg.Hooks.Cleanup != "" {
		printInfo("Running cleanup hook...")

		// Set environment variables
		var dbName string
		if loop != nil {
			dbName = fmt.Sprintf("%s_%s", strings.ReplaceAll(loop.Project, "-", "_"), strings.ReplaceAll(loop.Feature, "-", "_"))
		}

		env := os.Environ()
		env = append(env, fmt.Sprintf("DB_NAME=%s", dbName))
		env = append(env, fmt.Sprintf("WORKTREE_PATH=%s", worktreePath))
		if loop != nil {
			env = append(env, fmt.Sprintf("FEATURE=%s", loop.Feature))
		}

		hookCmd := exec.Command("bash", "-c", cfg.Hooks.Cleanup)
		hookCmd.Dir = worktreePath
		hookCmd.Env = env
		hookCmd.Stdout = os.Stdout
		hookCmd.Stderr = os.Stderr

		if err := hookCmd.Run(); err != nil {
			printWarn(fmt.Sprintf("Cleanup hook failed: %v", err))
		}
	}

	// Find main repo
	gitCmd := exec.Command("git", "worktree", "list", "--porcelain")
	gitCmd.Dir = worktreePath
	output, _ := gitCmd.Output()

	var mainRepo string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			if path != worktreePath {
				mainRepo = path
				break
			}
		}
	}

	if mainRepo == "" {
		mainRepo = filepath.Dir(worktreePath)
	}

	// Remove worktree
	printInfo("Removing worktree...")

	removeCmd := exec.Command("git", "worktree", "remove", worktreePath, "--force")
	removeCmd.Dir = mainRepo
	removeCmd.Stdout = os.Stdout
	removeCmd.Stderr = os.Stderr

	if err := removeCmd.Run(); err != nil {
		// Try manual removal
		os.RemoveAll(worktreePath)

		// Prune worktrees
		pruneCmd := exec.Command("git", "worktree", "prune")
		pruneCmd.Dir = mainRepo
		pruneCmd.Run()
	}

	// Delete branch if requested
	if deleteBranch && loop != nil && loop.Branch != "" {
		printInfo(fmt.Sprintf("Deleting branch %s...", loop.Branch))
		branchCmd := exec.Command("git", "branch", "-D", loop.Branch)
		branchCmd.Dir = mainRepo
		branchCmd.Run()
	}

	// Remove from registry
	if loop != nil {
		config.RemoveLoop(loop.Name)
	}

	printSuccess(fmt.Sprintf("Cleaned up: %s", worktreeName))

	return nil
}
