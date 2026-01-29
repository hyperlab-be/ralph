package loop

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/hyperlab/rl/internal/config"
)

// IsRunning checks if a loop is currently running
func IsRunning(loop *config.Loop) bool {
	if loop == nil || loop.PID == 0 {
		return false
	}

	process, err := os.FindProcess(loop.PID)
	if err != nil {
		return false
	}

	// On Unix, FindProcess always succeeds, so we need to send signal 0
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// GetStatus returns the current status of a loop
func GetStatus(loop *config.Loop) string {
	if IsRunning(loop) {
		return "running"
	}
	return "stopped"
}

// Start starts a loop
func Start(loop *config.Loop) error {
	if IsRunning(loop) {
		return fmt.Errorf("loop %s is already running", loop.Name)
	}

	// Build the command
	cmd := exec.Command("rl", "run")
	cmd.Dir = loop.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start in background
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start loop: %w", err)
	}

	// Update registry
	loop.PID = cmd.Process.Pid
	loop.Status = "running"
	if err := config.SetLoop(loop); err != nil {
		return fmt.Errorf("failed to update loop registry: %w", err)
	}

	return nil
}

// Stop stops a running loop
func Stop(loop *config.Loop) error {
	if !IsRunning(loop) {
		return nil
	}

	process, err := os.FindProcess(loop.PID)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop process: %w", err)
	}

	// Update registry
	loop.PID = 0
	loop.Status = "stopped"
	return config.SetLoop(loop)
}

// ListAll returns all registered loops
func ListAll() ([]*config.Loop, error) {
	registry, err := config.LoadLoops()
	if err != nil {
		return nil, err
	}

	loops := make([]*config.Loop, 0, len(registry.Loops))
	for _, loop := range registry.Loops {
		loops = append(loops, loop)
	}

	return loops, nil
}
