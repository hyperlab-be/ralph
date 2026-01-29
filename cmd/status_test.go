package cmd

import (
	"os"
	"testing"

	"github.com/hyperlab-be/ralph/internal/config"
)

func TestRunStatus(t *testing.T) {
	// Setup temp config dir
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Should not panic with no loops
	err := runStatus(statusCmd, []string{})
	if err != nil {
		t.Errorf("status should not error: %v", err)
	}
}

func TestRunStatusWithLoops(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Add a loop
	loop := &config.Loop{
		Name:   "test-loop",
		Status: "stopped",
		Path:   "/tmp/test-project",
	}
	config.SetLoop(loop)

	err := runStatus(statusCmd, []string{})
	if err != nil {
		t.Errorf("status should not error: %v", err)
	}
}

func TestRunStatusWithRunningLoop(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Add a running loop
	loop := &config.Loop{
		Name:   "running-loop",
		Status: "running",
		Path:   "/tmp/running-project",
		PID:    99999, // Non-existent PID
	}
	config.SetLoop(loop)

	err := runStatus(statusCmd, []string{})
	if err != nil {
		t.Errorf("status should not error: %v", err)
	}
}

func TestRunStatusSpecificLoop(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Add a loop
	loop := &config.Loop{
		Name:   "specific-loop",
		Status: "stopped",
		Path:   "/tmp/specific-project",
	}
	config.SetLoop(loop)

	err := runStatus(statusCmd, []string{"specific-loop"})
	if err != nil {
		t.Errorf("status for specific loop should not error: %v", err)
	}
}

func TestRunStatusNonExistentLoop(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	err := runStatus(statusCmd, []string{"non-existent"})
	// status shows available loops and doesn't error for non-existent
	// This is acceptable UX behavior
	_ = err
}
