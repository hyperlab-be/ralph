package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hyperlab-be/ralph/internal/config"
	"github.com/hyperlab-be/ralph/internal/loop"
	"github.com/hyperlab-be/ralph/internal/prd"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:     "status [name]",
	Aliases: []string{"s"},
	Short:   "Show status of loops",
	Long:    `Show the status of all registered loops or a specific loop.`,
	Args:    cobra.MaximumNArgs(1),
	RunE:    runStatus,
}

var watchStatus bool
var watchInterval int

func init() {
	statusCmd.Flags().BoolVarP(&watchStatus, "watch", "w", false, "Auto-refresh status")
	statusCmd.Flags().IntVarP(&watchInterval, "interval", "i", 5, "Refresh interval in seconds (with --watch)")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	filterName := ""
	if len(args) > 0 {
		filterName = args[0]
	}

	if watchStatus {
		return runStatusWatch(filterName)
	}

	return renderStatus(filterName)
}

func renderStatus(filterName string) error {
	// Header
	fmt.Println("\033[1m\033[36m")
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                 🤖 ralph - Loop Status                    ║")
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
		fmt.Println("  ralph init")
		fmt.Println("  ralph new feature-name")
		return nil
	}

	for _, l := range loops {
		if filterName != "" && l.Name != filterName {
			continue
		}

		printLoopStatus(l)
	}

	return nil
}

func runStatusWatch(filterName string) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(watchInterval) * time.Second)
	defer ticker.Stop()

	// Initial render
	renderStatusScreen(filterName)

	for {
		select {
		case <-ticker.C:
			renderStatusScreen(filterName)
		case <-sigChan:
			fmt.Println("\nExiting...")
			return nil
		}
	}
}

func renderStatusScreen(filterName string) {
	// Clear screen
	fmt.Print("\033[2J\033[H")
	renderStatus(filterName)
	fmt.Printf("\n\033[2m[Refreshing every %ds - Ctrl+C to exit]\033[0m\n", watchInterval)
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
