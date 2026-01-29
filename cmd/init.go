package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize ralph in a project",
	Long:  `Initialize ralph configuration in the current or specified project directory.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	projectRoot := "."
	if len(args) > 0 {
		projectRoot = args[0]
	}

	// Get absolute path
	absPath, err := filepath.Abs(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if already initialized
	configPath := filepath.Join(absPath, "ralph.toml")
	if _, err := os.Stat(configPath); err == nil {
		printWarn("Project already initialized")
		return nil
	}

	projectName := filepath.Base(absPath)

	// Create ralph.toml
	configContent := fmt.Sprintf(`# ralph configuration for %s

[project]
name = "%s"

[worktree]
# Worktrees will be named: %s-<feature>
prefix = "%s"

[hooks]
# Commands to run after creating a worktree
# Available variables: $WORKTREE_PATH, $FEATURE
setup = """
# Example:
# cp .env.example .env
# npm install
"""

# Commands to run before removing a worktree
cleanup = """
# Example:
# rm -rf node_modules
"""

[agent]
model = "claude-sonnet-4-20250514"
max_iterations = 10
# Custom prompt file (optional)
# prompt = ".ralph/prompt.md"
`, projectName, projectName, projectName, projectName)

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create ralph.toml: %w", err)
	}

	// Create .ralph directory
	ralphDir := filepath.Join(absPath, ".ralph")
	if err := os.MkdirAll(ralphDir, 0755); err != nil {
		return fmt.Errorf("failed to create .ralph directory: %w", err)
	}

	// Add ralph artifacts to .gitignore if not already present
	gitignorePath := filepath.Join(absPath, ".gitignore")
	gitignoreEntry := "\n# Ralph tooling\n.ralph/\nprd.json\n"

	existingGitignore, _ := os.ReadFile(gitignorePath)
	if !strings.Contains(string(existingGitignore), ".ralph/") {
		f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			f.WriteString(gitignoreEntry)
			f.Close()
			printInfo("Added .ralph/ to .gitignore")
		}
	}

	printSuccess(fmt.Sprintf("Initialized ralph in %s", absPath))
	printInfo("Edit ralph.toml to configure hooks and settings")

	return nil
}
