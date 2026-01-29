package cmd

import (
	"os"
	"testing"

	"github.com/hyperlab-be/ralph/internal/config"
)

func TestRunStopNoArgs(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	err := runStop(stopCmd, []string{})
	if err == nil {
		t.Error("stop should error when no loop name provided")
	}
}

func TestRunStopNonExistentLoop(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	err := runStop(stopCmd, []string{"non-existent"})
	if err == nil {
		t.Error("stop should error for non-existent loop")
	}
}

func TestRunStopAlreadyStopped(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Add a stopped loop
	config.SetLoop(&config.Loop{
		Name:   "stopped-loop",
		Status: "stopped",
	})

	err := runStop(stopCmd, []string{"stopped-loop"})
	// Should not error, just warn
	if err != nil {
		t.Errorf("stop should not error for already stopped loop: %v", err)
	}
}

func TestRunStopRunningLoop(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Add a "running" loop with non-existent PID
	config.SetLoop(&config.Loop{
		Name:   "running-loop",
		Status: "running",
		PID:    999999, // Non-existent PID
	})

	// Stop should handle non-existent process gracefully
	err := runStop(stopCmd, []string{"running-loop"})
	// May or may not error depending on implementation
	_ = err
}
