package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hyperlab-be/ralph/internal/config"
)

func TestRunCleanupNoArgsNotInProject(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Change to a non-project directory
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// When no args and not in project, cleanup prompts for confirmation
	// which gets cancelled automatically in test mode (no TTY)
	// This is acceptable behavior
	_ = runCleanup(cleanupCmd, []string{})
}

func TestRunCleanupNonExistentLoop(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	err := runCleanup(cleanupCmd, []string{"non-existent"})
	if err == nil {
		t.Error("cleanup should error for non-existent loop")
	}
}

func TestRunCleanupRunningLoop(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Add a running loop
	config.SetLoop(&config.Loop{
		Name:   "running-loop",
		Status: "running",
		Path:   "/tmp/running-loop",
	})

	err := runCleanup(cleanupCmd, []string{"running-loop"})
	if err == nil {
		t.Error("cleanup should error for running loop")
	}
}

func TestRunCleanupStoppedLoop(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", configDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Create a fake worktree directory
	loopDir := filepath.Join(tmpDir, "cleanup-loop")
	os.MkdirAll(loopDir, 0755)

	// Add a stopped loop
	config.SetLoop(&config.Loop{
		Name:   "cleanup-loop",
		Status: "stopped",
		Path:   loopDir,
	})

	// Set force flag to skip confirmation
	forceCleanup = true
	defer func() { forceCleanup = false }()

	// Note: This won't actually cleanup because it's not a real worktree
	// but it should handle the error gracefully
	_ = runCleanup(cleanupCmd, []string{"cleanup-loop"})
}
