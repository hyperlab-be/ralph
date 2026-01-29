package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir(t *testing.T) {
	// Test with environment variable
	os.Setenv("RALPH_CONFIG_DIR", "/tmp/test-ralph-config")
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	dir := ConfigDir()
	if dir != "/tmp/test-ralph-config" {
		t.Errorf("Expected /tmp/test-ralph-config, got %s", dir)
	}
}

func TestConfigDirDefault(t *testing.T) {
	os.Unsetenv("RALPH_CONFIG_DIR")

	dir := ConfigDir()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "ralph")

	if dir != expected {
		t.Errorf("Expected %s, got %s", expected, dir)
	}
}

func TestLoadLoopsEmpty(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	registry, err := LoadLoops()
	if err != nil {
		t.Fatalf("Failed to load loops: %v", err)
	}

	if len(registry.Loops) != 0 {
		t.Errorf("Expected empty loops, got %d", len(registry.Loops))
	}
}

func TestSetAndGetLoop(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	loop := &Loop{
		Name:    "test-loop",
		Path:    "/tmp/test-project",
		Project: "test",
		Feature: "feature",
		Branch:  "feature/test",
		Status:  "stopped",
	}

	// Save loop
	err := SetLoop(loop)
	if err != nil {
		t.Fatalf("Failed to set loop: %v", err)
	}

	// Get loop
	retrieved, err := GetLoop("test-loop")
	if err != nil {
		t.Fatalf("Failed to get loop: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected loop, got nil")
	}

	if retrieved.Name != loop.Name {
		t.Errorf("Expected name %s, got %s", loop.Name, retrieved.Name)
	}

	if retrieved.Path != loop.Path {
		t.Errorf("Expected path %s, got %s", loop.Path, retrieved.Path)
	}
}

func TestRemoveLoop(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	loop := &Loop{
		Name:   "to-remove",
		Path:   "/tmp/test",
		Status: "stopped",
	}

	// Add loop
	SetLoop(loop)

	// Remove loop
	err := RemoveLoop("to-remove")
	if err != nil {
		t.Fatalf("Failed to remove loop: %v", err)
	}

	// Verify removed
	retrieved, _ := GetLoop("to-remove")
	if retrieved != nil {
		t.Error("Expected nil after removal")
	}
}

func TestFindProjectRoot(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "myproject")
	subDir := filepath.Join(projectDir, "src", "pkg")
	os.MkdirAll(subDir, 0755)

	// Create ralph.toml
	ralphToml := filepath.Join(projectDir, "ralph.toml")
	os.WriteFile(ralphToml, []byte("[project]\nname = \"test\"\n"), 0644)

	// Find from subdirectory
	found, err := FindProjectRoot(subDir)
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	if found != projectDir {
		t.Errorf("Expected %s, got %s", projectDir, found)
	}
}

func TestFindProjectRootNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := FindProjectRoot(tmpDir)
	if err == nil {
		t.Error("Expected error when no project root found")
	}
}

func TestLoadProjectConfig(t *testing.T) {
	// Create temp project
	tmpDir := t.TempDir()
	configContent := `
[project]
name = "test-project"

[worktree]
prefix = "test"

[hooks]
setup = "echo setup"
cleanup = "echo cleanup"
`
	os.WriteFile(filepath.Join(tmpDir, "ralph.toml"), []byte(configContent), 0644)

	cfg, err := LoadProjectConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load project config: %v", err)
	}

	if cfg.Project.Name != "test-project" {
		t.Errorf("Expected project name 'test-project', got '%s'", cfg.Project.Name)
	}

	if cfg.Worktree.Prefix != "test" {
		t.Errorf("Expected worktree prefix 'test', got '%s'", cfg.Worktree.Prefix)
	}

	if cfg.Hooks.Setup != "echo setup" {
		t.Errorf("Unexpected setup hook: %s", cfg.Hooks.Setup)
	}
}

func TestLoadProjectConfigNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	cfg, err := LoadProjectConfig(tmpDir)
	if err != nil {
		t.Errorf("Expected no error for missing config, got: %v", err)
	}
	if cfg != nil {
		t.Error("Expected nil config for missing file")
	}
}

func TestLoadGlobalConfig(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Test default values when no config exists
	cfg, err := LoadGlobalConfig()
	if err != nil {
		t.Fatalf("Failed to load global config: %v", err)
	}

	// Only projects_dir should have a default now
	if cfg.Defaults.ProjectsDir != "~/Code" {
		t.Errorf("Expected default projects_dir '~/Code', got '%s'", cfg.Defaults.ProjectsDir)
	}
}

func TestLoopsFile(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	expected := filepath.Join(tmpDir, "loops.json")
	if got := LoopsFile(); got != expected {
		t.Errorf("Expected %s, got %s", expected, got)
	}
}

func TestGlobalConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	expected := filepath.Join(tmpDir, "config.toml")
	if got := GlobalConfigFile(); got != expected {
		t.Errorf("Expected %s, got %s", expected, got)
	}
}

func TestFindProjectRootByRalphDir(t *testing.T) {
	// Create temp directory structure with .ralph dir
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "myproject")
	ralphDir := filepath.Join(projectDir, ".ralph")
	subDir := filepath.Join(projectDir, "src")
	os.MkdirAll(ralphDir, 0755)
	os.MkdirAll(subDir, 0755)

	// Find from subdirectory (should find by .ralph dir)
	found, err := FindProjectRoot(subDir)
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	if found != projectDir {
		t.Errorf("Expected %s, got %s", projectDir, found)
	}
}

func TestGetLoopNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	loop, err := GetLoop("nonexistent")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if loop != nil {
		t.Error("Expected nil for nonexistent loop")
	}
}
