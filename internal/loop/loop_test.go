package loop

import (
	"os"
	"testing"

	"github.com/hyperlab-be/ralph/internal/config"
)

func TestIsRunning(t *testing.T) {
	// Test nil loop
	if IsRunning(nil) {
		t.Error("Expected false for nil loop")
	}

	// Test loop with PID 0
	loop := &config.Loop{PID: 0}
	if IsRunning(loop) {
		t.Error("Expected false for loop with PID 0")
	}

	// Test loop with non-existent PID
	loop = &config.Loop{PID: 99999999}
	if IsRunning(loop) {
		t.Error("Expected false for non-existent PID")
	}

	// Test loop with current process PID (should be running)
	loop = &config.Loop{PID: os.Getpid()}
	if !IsRunning(loop) {
		t.Error("Expected true for current process PID")
	}
}

func TestGetStatus(t *testing.T) {
	// Test stopped loop
	loop := &config.Loop{PID: 0}
	if status := GetStatus(loop); status != "stopped" {
		t.Errorf("Expected 'stopped', got '%s'", status)
	}

	// Test running loop (current process)
	loop = &config.Loop{PID: os.Getpid()}
	if status := GetStatus(loop); status != "running" {
		t.Errorf("Expected 'running', got '%s'", status)
	}
}

func TestListAll(t *testing.T) {
	// Create temp config dir
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Test empty list
	loops, err := ListAll()
	if err != nil {
		t.Fatalf("Failed to list loops: %v", err)
	}
	if len(loops) != 0 {
		t.Errorf("Expected 0 loops, got %d", len(loops))
	}

	// Add a loop
	testLoop := &config.Loop{
		Name:   "test-loop",
		Path:   "/tmp/test",
		Status: "stopped",
	}
	config.SetLoop(testLoop)

	// Test list with one loop
	loops, err = ListAll()
	if err != nil {
		t.Fatalf("Failed to list loops: %v", err)
	}
	if len(loops) != 1 {
		t.Errorf("Expected 1 loop, got %d", len(loops))
	}
	if loops[0].Name != "test-loop" {
		t.Errorf("Expected name 'test-loop', got '%s'", loops[0].Name)
	}
}

func TestStartAlreadyRunning(t *testing.T) {
	loop := &config.Loop{
		Name: "test",
		PID:  os.Getpid(), // Current process, so it's "running"
	}

	err := Start(loop)
	if err == nil {
		t.Error("Expected error when starting already running loop")
	}
}

func TestStopNotRunning(t *testing.T) {
	loop := &config.Loop{
		Name: "test",
		PID:  0,
	}

	// Should not error when stopping already stopped loop
	err := Stop(loop)
	if err != nil {
		t.Errorf("Unexpected error stopping already stopped loop: %v", err)
	}
}
