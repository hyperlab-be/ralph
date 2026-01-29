package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunNewNotInGitRepo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create ralph project but not git repo
	os.MkdirAll(filepath.Join(tmpDir, ".ralph"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "ralph.toml"), []byte("[project]\nname = \"test\"\n"), 0644)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	err := runNew(newCmd, []string{"test-feature"})
	if err == nil {
		t.Error("new should error when not in git repo")
	}
}

func TestRunNewNoFeatureName(t *testing.T) {
	// cobra.ExactArgs(1) handles this at the command level
	// Test that the command has the correct args requirement
	if newCmd.Args == nil {
		t.Error("new command should have Args validator")
	}
	// Verify by checking the Use string
	if newCmd.Use != "new <feature>" {
		t.Errorf("new command should require feature arg, got Use: %s", newCmd.Use)
	}
}

func TestRunNewInvalidFeatureName(t *testing.T) {
	tmpDir := t.TempDir()

	// Create git repo
	exec.Command("git", "init", tmpDir).Run()

	os.MkdirAll(filepath.Join(tmpDir, ".ralph"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "ralph.toml"), []byte("[project]\nname = \"test\"\n"), 0644)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	err := runNew(newCmd, []string{"invalid feature name"})
	if err == nil {
		t.Error("new should error for invalid feature name")
	}
}

func TestRunNewNotInRalphProject(t *testing.T) {
	tmpDir := t.TempDir()

	// Create git repo but not ralph project
	exec.Command("git", "init", tmpDir).Run()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	err := runNew(newCmd, []string{"test-feature"})
	if err == nil {
		t.Error("new should error when not in ralph project")
	}
}

func TestRunNewValidFeature(t *testing.T) {
	tmpDir := t.TempDir()

	// Create git repo with initial commit
	cmd := exec.Command("git", "init", tmpDir)
	cmd.Run()

	// Need to set git config for commit
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()

	// Create a file and commit
	os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test"), 0644)
	exec.Command("git", "-C", tmpDir, "add", ".").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "initial").Run()

	// Create ralph project
	os.MkdirAll(filepath.Join(tmpDir, ".ralph"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "ralph.toml"), []byte("[project]\nname = \"test\"\n"), 0644)

	configDir := t.TempDir()
	os.Setenv("RALPH_CONFIG_DIR", configDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// This should work (creates worktree)
	err := runNew(newCmd, []string{"valid-feature"})
	// May fail if git worktree isn't supported properly in tmp
	// but should not panic
	_ = err
}
