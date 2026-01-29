package cmd

import (
	"os"
	"testing"

	"github.com/hyperlab-be/ralph/internal/config"
)

func TestRunList(t *testing.T) {
	// Setup temp config dir
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Should not panic with no loops
	err := runList(listCmd, []string{})
	if err != nil {
		t.Errorf("list should not error: %v", err)
	}
}

func TestRunListWithLoops(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Add loops
	config.SetLoop(&config.Loop{Name: "loop-1", Status: "stopped"})
	config.SetLoop(&config.Loop{Name: "loop-2", Status: "running"})

	err := runList(listCmd, []string{})
	if err != nil {
		t.Errorf("list should not error: %v", err)
	}
}
