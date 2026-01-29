package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/hyperlab-be/ralph/internal/config"
)

func TestValidateFeatureName(t *testing.T) {
	tests := []struct {
		name    string
		feature string
		wantErr bool
	}{
		{"valid lowercase", "my-feature", false},
		{"valid with numbers", "feature-123", false},
		{"valid with underscore", "my_feature", false},
		{"valid uppercase", "MyFeature", false},
		{"invalid space", "my feature", true},
		{"invalid special char", "my@feature", true},
		{"invalid dot", "my.feature", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFeatureName(tt.feature)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFeatureName(%q) error = %v, wantErr %v", tt.feature, err, tt.wantErr)
			}
		})
	}
}

func TestPrintAvailableLoops(t *testing.T) {
	// Setup temp config dir
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Test with no loops
	printAvailableLoops()

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stderr = oldStderr

	output := buf.String()
	if output != "  (no loops registered)\n" {
		t.Errorf("Expected '(no loops registered)', got %q", output)
	}

	// Add a loop
	loop := &config.Loop{Name: "test-loop", Status: "stopped"}
	config.SetLoop(loop)

	// Test with loops
	r, w, _ = os.Pipe()
	os.Stderr = w
	printAvailableLoops()
	w.Close()
	buf.Reset()
	buf.ReadFrom(r)
	os.Stderr = oldStderr

	output = buf.String()
	if output == "  (no loops registered)\n" {
		t.Error("Should show loops when they exist")
	}
}

func TestRootCommand(t *testing.T) {
	// Test that root command exists and has correct name
	if rootCmd.Use != "ralph" {
		t.Errorf("Expected root command 'ralph', got '%s'", rootCmd.Use)
	}

	if rootCmd.Version != Version {
		t.Errorf("Expected version '%s', got '%s'", Version, rootCmd.Version)
	}
}

func TestInitCreatesFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp dir
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Run init
	err := runInit(nil, []string{})
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Check ralph.toml exists
	if _, err := os.Stat(filepath.Join(tmpDir, "ralph.toml")); os.IsNotExist(err) {
		t.Error("ralph.toml was not created")
	}

	// Check .ralph directory exists
	if _, err := os.Stat(filepath.Join(tmpDir, ".ralph")); os.IsNotExist(err) {
		t.Error(".ralph directory was not created")
	}
}

func TestInitAlreadyInitialized(t *testing.T) {
	tmpDir := t.TempDir()

	// Create ralph.toml
	os.WriteFile(filepath.Join(tmpDir, "ralph.toml"), []byte("test"), 0644)

	// Change to temp dir
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Run init - should not error, just warn
	err := runInit(nil, []string{})
	if err != nil {
		t.Errorf("init should not fail on already initialized project: %v", err)
	}
}

func TestDoctorCommand(t *testing.T) {
	// Doctor should not panic
	err := runDoctor(nil, []string{})
	// May fail if git/claude not installed, but should not panic
	_ = err
}

// Helper function to validate feature name (extracted for testing)
func validateFeatureName(feature string) error {
	if feature == "" {
		return os.ErrInvalid
	}
	for _, char := range feature {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_') {
			return os.ErrInvalid
		}
	}
	return nil
}
