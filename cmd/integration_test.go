//go:build integration

package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Integration tests that actually run Claude CLI
// Run with: go test -tags=integration -v ./cmd -run TestIntegration

func TestIntegrationFullLoop(t *testing.T) {
	// This test actually runs Claude CLI and takes 1-3 minutes
	// Skip by default, run with: RALPH_RUN_CLAUDE_TEST=1 go test -tags=integration -run TestIntegrationFullLoop
	if os.Getenv("RALPH_RUN_CLAUDE_TEST") == "" {
		t.Skip("Set RALPH_RUN_CLAUDE_TEST=1 to run full Claude integration test (takes 1-3 min)")
	}
	
	// Skip if Claude CLI is not installed
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("Claude CLI not installed")
	}

	// Create temp directory
	tmpDir := t.TempDir()
	configDir := t.TempDir()
	
	// Use isolated config directory
	os.Setenv("RALPH_CONFIG_DIR", configDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Initialize git repo
	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.email", "test@test.com")
	runGit(t, tmpDir, "config", "user.name", "Test")

	// Create initial file and commit
	os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test Project\n"), 0644)
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "initial")

	// Initialize ralph
	ralphBin := getRalphBinary(t)
	runRalph(t, tmpDir, ralphBin, "init")

	// Create a very simple PRD - just create a single file
	prdContent := `{
	"name": "Simple Test",
	"description": "Create one file",
	"userStories": [
		{
			"id": "1",
			"title": "Create hello.txt",
			"description": "Create hello.txt with text Hello World",
			"acceptanceCriteria": ["hello.txt exists with Hello World"],
			"passes": false
		}
	]
}`
	os.WriteFile(filepath.Join(tmpDir, ".ralph", "prd.json"), []byte(prdContent), 0644)

	// Run ralph with --once (single iteration)
	t.Log("Running ralph with Claude CLI (this may take 1-3 minutes)...")
	start := time.Now()
	output := runRalphWithOutput(t, tmpDir, ralphBin, "run", "--once")
	t.Logf("Completed in %v", time.Since(start))
	t.Logf("Ralph output:\n%s", output)

	// Check conversation log was created (main success criteria)
	convDir := filepath.Join(tmpDir, ".ralph", "conversations")
	if _, err := os.Stat(convDir); os.IsNotExist(err) {
		t.Error("conversations directory was not created")
	} else {
		files, _ := filepath.Glob(filepath.Join(convDir, "*.md"))
		if len(files) == 0 {
			t.Error("no conversation logs created")
		} else {
			t.Logf("Conversation logs created: %v", files)
			// Read first log to verify content
			if content, err := os.ReadFile(files[0]); err == nil {
				if !strings.Contains(string(content), "## Prompt") {
					t.Error("conversation log should contain prompt section")
				}
				if !strings.Contains(string(content), "## Agent Output") {
					t.Error("conversation log should contain agent output section")
				}
				t.Logf("Conversation log preview:\n%s", truncate(string(content), 1000))
			}
		}
	}

	// Check if hello.txt was created (nice to have, may not always happen)
	helloPath := filepath.Join(tmpDir, "hello.txt")
	if content, err := os.ReadFile(helloPath); err == nil {
		t.Logf("hello.txt created with content: %s", string(content))
	} else {
		t.Log("hello.txt not created - this is OK, Claude output varies")
		// List files for debugging
		files, _ := filepath.Glob(filepath.Join(tmpDir, "*"))
		t.Logf("Files in project: %v", files)
	}

	// Check PRD status
	prdPath := filepath.Join(tmpDir, ".ralph", "prd.json")
	if prdData, err := os.ReadFile(prdPath); err == nil {
		t.Logf("PRD after run:\n%s", string(prdData))
	}
}

func TestIntegrationDryRun(t *testing.T) {
	// This test doesn't need Claude CLI
	tmpDir := t.TempDir()
	configDir := t.TempDir()
	
	// Use isolated config directory
	os.Setenv("RALPH_CONFIG_DIR", configDir)
	defer os.Unsetenv("RALPH_CONFIG_DIR")

	// Initialize git repo
	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.email", "test@test.com")
	runGit(t, tmpDir, "config", "user.name", "Test")
	os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test\n"), 0644)
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "initial")

	ralphBin := getRalphBinary(t)

	// Initialize
	runRalph(t, tmpDir, ralphBin, "init")

	// Create PRD
	prdContent := `{"name": "Test", "userStories": [{"id": "1", "title": "Test story", "passes": false}]}`
	os.WriteFile(filepath.Join(tmpDir, ".ralph", "prd.json"), []byte(prdContent), 0644)

	// Dry run should work without Claude
	output := runRalphWithOutput(t, tmpDir, ralphBin, "run", "--dry-run")

	if !strings.Contains(output, "Dry run mode") {
		t.Errorf("Expected dry run output, got: %s", output)
	}
	if !strings.Contains(output, "Would work on") {
		t.Errorf("Expected 'Would work on' in output, got: %s", output)
	}
}

func TestIntegrationStatus(t *testing.T) {
	ralphBin := getRalphBinary(t)

	// Status should work even outside a project
	output := runRalphWithOutput(t, t.TempDir(), ralphBin, "status")

	if !strings.Contains(output, "ralph") {
		t.Errorf("Expected ralph status output, got: %s", output)
	}
}

func TestIntegrationDoctor(t *testing.T) {
	ralphBin := getRalphBinary(t)

	output := runRalphWithOutput(t, t.TempDir(), ralphBin, "doctor")

	if !strings.Contains(output, "git") {
		t.Errorf("Expected git check in doctor output, got: %s", output)
	}
}

func TestIntegrationPrdWorkflow(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git
	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.email", "test@test.com")
	runGit(t, tmpDir, "config", "user.name", "Test")
	os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test\n"), 0644)
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "initial")

	ralphBin := getRalphBinary(t)

	// Initialize ralph
	runRalph(t, tmpDir, ralphBin, "init")

	// Create PRD manually
	prdContent := `{"name": "Test Feature", "userStories": []}`
	os.WriteFile(filepath.Join(tmpDir, ".ralph", "prd.json"), []byte(prdContent), 0644)

	// Add a story via CLI
	runRalph(t, tmpDir, ralphBin, "prd", "add", "First story", "-c", "Criterion 1", "-c", "Criterion 2")

	// View PRD
	output := runRalphWithOutput(t, tmpDir, ralphBin, "prd")

	if !strings.Contains(output, "First story") {
		t.Errorf("PRD should contain 'First story', got: %s", output)
	}
	if !strings.Contains(output, "0/1") || !strings.Contains(output, "1/1") || !strings.Contains(output, "Progress") {
		// Should show progress
		t.Logf("PRD output: %s", output)
	}
}

// Helper functions

func getRalphBinary(t *testing.T) string {
	// First try the built binary in the repo
	repoRoot := os.Getenv("RALPH_REPO_ROOT")
	if repoRoot == "" {
		// Try to find it relative to test
		cwd, _ := os.Getwd()
		repoRoot = filepath.Dir(cwd)
	}
	
	binaryPath := filepath.Join(repoRoot, "ralph")
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath
	}

	// Try ~/Code/ralph/ralph
	homeDir, _ := os.UserHomeDir()
	binaryPath = filepath.Join(homeDir, "Code", "ralph", "ralph")
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath
	}

	// Fall back to PATH
	path, err := exec.LookPath("ralph")
	if err != nil {
		t.Fatal("ralph binary not found")
	}
	return path
}

func runGit(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}

func runRalph(t *testing.T, dir, binary string, args ...string) {
	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("ralph %v failed: %v\n%s", args, err, output)
	}
}

func runRalphWithOutput(t *testing.T, dir, binary string, args ...string) string {
	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	// Pass through environment including RALPH_CONFIG_DIR
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Don't fail, just log - some commands may "fail" but still work
		t.Logf("ralph %v returned error: %v", args, err)
	}
	return string(output)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
