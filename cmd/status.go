package cmd

import (
	"fmt"

	"github.com/hyperlab/rl/internal/config"
	"github.com/hyperlab/rl/internal/loop"
	"github.com/hyperlab/rl/internal/prd"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Show status of loops",
	Long:  `Show the status of all registered loops or a specific loop.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Header
	fmt.Println("\033[1m\033[36m")
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                 🤖 rl - Loop Status                       ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println("\033[0m")

	loops, err := loop.ListAll()
	if err != nil {
		return fmt.Errorf("failed to list loops: %w", err)
	}

	if len(loops) == 0 {
		fmt.Println("\033[2mNo loops registered.\033[0m")
		fmt.Println()
		fmt.Println("Start a new project with:")
		fmt.Println("  cd ~/Code/your-project")
		fmt.Println("  rl init")
		fmt.Println("  rl new feature-name")
		return nil
	}

	filterName := ""
	if len(args) > 0 {
		filterName = args[0]
	}

	for _, l := range loops {
		if filterName != "" && l.Name != filterName {
			continue
		}

		printLoopStatus(l)
	}

	return nil
}

func printLoopStatus(l *config.Loop) {
	// Status indicator
	status := loop.GetStatus(l)
	var statusIcon, statusColor string
	if status == "running" {
		statusIcon = "🟢"
		statusColor = "\033[32m" // Green
	} else {
		statusIcon = "⚫"
		statusColor = "\033[31m" // Red
	}

	// Progress
	progress := "?/?"
	var currentStory string
	if p, err := prd.Load(l.Path); err == nil && p != nil {
		progress = p.Progress()
		if story := p.GetCurrentStory(); story != nil && status == "running" {
			currentStory = story.Title
		}
	}

	// Print
	fmt.Printf("%s \033[1m%s\033[0m\n", statusIcon, l.Name)
	fmt.Printf("   Status: %s%s\033[0m\n", statusColor, status)
	fmt.Printf("   Progress: %s stories\n", progress)
	fmt.Printf("   Path: \033[2m%s\033[0m\n", l.Path)

	if currentStory != "" {
		fmt.Printf("   Current: \033[36m%s\033[0m\n", currentStory)
	}

	fmt.Println()
}

// Aliases
func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "s",
		Short: "Alias for status",
		RunE:  runStatus,
		Args:  cobra.MaximumNArgs(1),
	})
}
