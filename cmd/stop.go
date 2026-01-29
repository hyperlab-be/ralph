package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/hyperlab-be/ralph/internal/config"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop a running loop",
	Long:  `Stop a running AI agent loop.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) error {
	var loopName string

	if len(args) > 0 {
		loopName = args[0]
	} else {
		// Use current directory
		cwd, _ := os.Getwd()
		projectRoot, err := config.FindProjectRoot(cwd)
		if err != nil {
			return fmt.Errorf("not in a rl project and no loop name provided")
		}
		loopName = filepath.Base(projectRoot)
	}

	// Get loop
	loop, err := config.GetLoop(loopName)
	if err != nil {
		return fmt.Errorf("failed to get loop: %w", err)
	}
	if loop == nil {
		return fmt.Errorf("loop not found: %s", loopName)
	}

	// Check if running
	if loop.PID == 0 {
		printWarn(fmt.Sprintf("Loop %s is not running", loopName))
		return nil
	}

	// Find process
	process, err := os.FindProcess(loop.PID)
	if err != nil {
		printWarn(fmt.Sprintf("Process %d not found", loop.PID))
		loop.PID = 0
		loop.Status = "stopped"
		config.SetLoop(loop)
		return nil
	}

	// Send SIGTERM
	printInfo(fmt.Sprintf("Stopping loop %s (PID %d)...", loopName, loop.PID))

	if err := process.Signal(syscall.SIGTERM); err != nil {
		// Process might already be dead
		printWarn(fmt.Sprintf("Failed to send signal: %v", err))
	}

	// Update status
	loop.PID = 0
	loop.Status = "stopped"
	if err := config.SetLoop(loop); err != nil {
		return fmt.Errorf("failed to update loop status: %w", err)
	}

	printSuccess(fmt.Sprintf("Stopped loop: %s", loopName))

	return nil
}
