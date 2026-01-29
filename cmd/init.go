package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize rl in a project",
	Long:  `Initialize rl configuration in the current or specified project directory.`,
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
	configPath := filepath.Join(absPath, "rl.toml")
	if _, err := os.Stat(configPath); err == nil {
		printWarn("Project already initialized")
		return nil
	}

	projectName := filepath.Base(absPath)

	// Create rl.toml
	configContent := fmt.Sprintf(`# rl configuration for %s

[project]
name = "%s"

[worktree]
# Worktrees will be named: %s-<feature>
prefix = "%s"

[hooks]
# Commands to run after creating a worktree
# Available variables: $DB_NAME, $WORKTREE_PATH, $FEATURE
setup = """
# Example for Laravel:
# cp .env.example .env
# sed -i '' "s/DB_DATABASE=.*/DB_DATABASE=${DB_NAME}/" .env
# mysql -e "CREATE DATABASE ${DB_NAME}"
# php artisan migrate
"""

# Commands to run before removing a worktree
cleanup = """
# Example for Laravel:
# mysql -e "DROP DATABASE IF EXISTS ${DB_NAME}"
"""

[agent]
model = "claude-sonnet-4-20250514"
max_iterations = 10
# Custom prompt file (optional)
# prompt = ".rl/prompt.md"
`, projectName, projectName, projectName, projectName)

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create rl.toml: %w", err)
	}

	// Create .rl directory
	rlDir := filepath.Join(absPath, ".rl")
	if err := os.MkdirAll(rlDir, 0755); err != nil {
		return fmt.Errorf("failed to create .rl directory: %w", err)
	}

	printSuccess(fmt.Sprintf("Initialized rl in %s", absPath))
	printInfo("Edit rl.toml to configure hooks and settings")

	return nil
}
