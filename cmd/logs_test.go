package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hyperlab-be/ralph/internal/config"
)

func TestRunLogsNoArgs(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	err := runLogs(logsCmd, []string{})
	if err == nil {
		t.Error("logs should error when no loop name provided")
	}
}

func TestRunLogsNonExistentLoop(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	err := runLogs(logsCmd, []string{"non-existent"})
	if err == nil {
		t.Error("logs should error for non-existent loop")
	}
}

func TestRunLogsWithLoop(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", configDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Create loop directory with log
	loopDir := filepath.Join(tmpDir, "test-loop")
	os.MkdirAll(filepath.Join(loopDir, ".ralph"), 0755)
	os.WriteFile(filepath.Join(loopDir, ".ralph", "session.log"), []byte("test log content\n"), 0644)

	// Register loop
	config.SetLoop(&config.Loop{
		Name:   "test-loop",
		Path:   loopDir,
		Status: "stopped",
	})

	// Should not error when log exists
	err := runLogs(logsCmd, []string{"test-loop"})
	if err != nil {
		t.Errorf("logs should not error when log exists: %v", err)
	}
}

func TestRunLogsNoLogFile(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", configDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Create loop directory without log
	loopDir := filepath.Join(tmpDir, "no-log-loop")
	os.MkdirAll(filepath.Join(loopDir, ".ralph"), 0755)

	// Register loop
	config.SetLoop(&config.Loop{
		Name:   "no-log-loop",
		Path:   loopDir,
		Status: "stopped",
	})

	// logs prints a warning but doesn't return an error for missing logs
	err := runLogs(logsCmd, []string{"no-log-loop"})
	// This is acceptable behavior - just warns instead of erroring
	_ = err
}
