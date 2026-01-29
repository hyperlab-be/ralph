package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hyperlab/ralph/internal/config"
	"github.com/hyperlab/ralph/internal/loop"
	"github.com/hyperlab/ralph/internal/prd"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:     "dashboard",
	Aliases: []string{"d"},
	Short:   "Live dashboard with auto-refresh",
	Long:    `Show a live dashboard that auto-refreshes every few seconds.`,
	RunE:    runDashboard,
}

var refreshInterval int

func init() {
	dashboardCmd.Flags().IntVarP(&refreshInterval, "interval", "i", 5, "Refresh interval in seconds")
	rootCmd.AddCommand(dashboardCmd)
}

func runDashboard(cmd *cobra.Command, args []string) error {
	// Handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(refreshInterval) * time.Second)
	defer ticker.Stop()

	// Initial render
	renderDashboard()

	for {
		select {
		case <-ticker.C:
			renderDashboard()
		case <-sigChan:
			fmt.Println("\nExiting dashboard...")
			return nil
		}
	}
}

func renderDashboard() {
	// Clear screen and move cursor to top
	fmt.Print("\033[2J\033[H")

	// Header
	fmt.Println("\033[1m\033[36m")
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║              🤖 rl - Live Dashboard                       ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println("\033[0m")

	loops, err := loop.ListAll()
	if err != nil {
		printError(fmt.Sprintf("Failed to list loops: %v", err))
		return
	}

	if len(loops) == 0 {
		fmt.Println("\033[2mNo loops registered.\033[0m")
	} else {
		for _, l := range loops {
			renderLoopCard(l)
		}
	}

	// Footer
	fmt.Println()
	fmt.Printf("\033[2m[Refreshing every %ds - Ctrl+C to exit]\033[0m\n", refreshInterval)
	fmt.Printf("\033[2mLast updated: %s\033[0m\n", time.Now().Format("15:04:05"))
}

func renderLoopCard(l *config.Loop) {
	status := loop.GetStatus(l)

	// Status colors
	var statusIcon, statusColor, progressBar string
	if status == "running" {
		statusIcon = "🟢"
		statusColor = "\033[32m"
	} else {
		statusIcon = "⚫"
		statusColor = "\033[31m"
	}

	// Load PRD for progress
	progress := "?/?"
	percent := 0
	var currentStory string

	if p, err := prd.Load(l.Path); err == nil && p != nil {
		progress = p.Progress()
		percent = p.ProgressPercent()

		if story := p.GetCurrentStory(); story != nil {
			currentStory = story.Title
		}
	}

	// Build progress bar
	barWidth := 20
	filled := (percent * barWidth) / 100
	progressBar = fmt.Sprintf("[%s%s]",
		repeatStr("█", filled),
		repeatStr("░", barWidth-filled))

	// Print card
	fmt.Printf("┌─────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│ %s \033[1m%-50s\033[0m │\n", statusIcon, l.Name)
	fmt.Printf("│   %sStatus: %-8s\033[0m   Progress: %s %s │\n",
		statusColor, status, progressBar, progress)

	if currentStory != "" && status == "running" {
		// Truncate long story titles
		if len(currentStory) > 45 {
			currentStory = currentStory[:42] + "..."
		}
		fmt.Printf("│   \033[36m→ %s\033[0m%s │\n", currentStory, repeatStr(" ", 45-len(currentStory)))
	}

	fmt.Printf("└─────────────────────────────────────────────────────────┘\n")
	fmt.Println()
}

func repeatStr(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
