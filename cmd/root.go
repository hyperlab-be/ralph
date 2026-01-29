package cmd

import (
	"fmt"
	"os"

	"github.com/hyperlab-be/ralph/internal/config"
	"github.com/spf13/cobra"
)

var Version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "ralph",
	Short: "AI-powered development loop manager",
	Long: `ralph is a CLI tool for managing AI-powered development loops.

It helps you:
  - Create and manage git worktrees for features
  - Define PRDs (Product Requirement Documents) with user stories
  - Run AI agents to implement features autonomously
  - Monitor progress across multiple loops`,
	Version: Version,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

// Helper functions for output
func printSuccess(msg string) {
	fmt.Fprintf(os.Stdout, "\033[32m✓\033[0m %s\n", msg)
}

func printError(msg string) {
	fmt.Fprintf(os.Stderr, "\033[31m✗\033[0m %s\n", msg)
}

func printInfo(msg string) {
	fmt.Fprintf(os.Stdout, "\033[36mℹ\033[0m %s\n", msg)
}

func printWarn(msg string) {
	fmt.Fprintf(os.Stdout, "\033[33m⚠\033[0m %s\n", msg)
}

func printAvailableLoops() {
	registry, err := config.LoadLoops()
	if err != nil || len(registry.Loops) == 0 {
		fmt.Fprintln(os.Stderr, "  (no loops registered)")
		return
	}
	for _, loop := range registry.Loops {
		status := "⚫"
		if loop.Status == "running" {
			status = "🟢"
		}
		fmt.Fprintf(os.Stderr, "  %s %s\n", status, loop.Name)
	}
}
