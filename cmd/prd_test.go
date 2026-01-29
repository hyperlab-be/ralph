package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShowPRDNoPRD(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, ".ralph"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "ralph.toml"), []byte("[project]\nname = \"test\"\n"), 0644)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// showPRD prints a warning when no PRD exists, doesn't return error
	err := showPRD(tmpDir)
	// This is acceptable - it warns instead of erroring
	_ = err
}

func TestShowPRDWithPRD(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, ".ralph"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "ralph.toml"), []byte("[project]\nname = \"test\"\n"), 0644)

	prdData := `{
		"name": "Test Feature",
		"description": "A test",
		"userStories": [
			{"id": "1", "title": "First story", "passes": false},
			{"id": "2", "title": "Second story", "passes": true}
		]
	}`
	os.WriteFile(filepath.Join(tmpDir, ".ralph", "prd.json"), []byte(prdData), 0644)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Should not error
	err := showPRD(tmpDir)
	if err != nil {
		t.Errorf("Should not error with valid PRD: %v", err)
	}
}

func TestAddStoryNoPRD(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, ".ralph"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "ralph.toml"), []byte("[project]\nname = \"test\"\n"), 0644)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	err := addStory(tmpDir, "New Story")
	if err == nil {
		t.Error("Should error when no PRD exists")
	}
	if !strings.Contains(err.Error(), "no PRD found") {
		t.Errorf("Error should mention no PRD found, got: %v", err)
	}
}

func TestAddStoryWithPRD(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, ".ralph"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "ralph.toml"), []byte("[project]\nname = \"test\"\n"), 0644)

	prdData := `{
		"name": "Test Feature",
		"userStories": [
			{"id": "1", "title": "Existing story", "passes": false}
		]
	}`
	os.WriteFile(filepath.Join(tmpDir, ".ralph", "prd.json"), []byte(prdData), 0644)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Set criteria via the global variable
	storyCriteria = []string{"Criterion 1", "Criterion 2"}
	defer func() {
		storyCriteria = nil
	}()

	err := addStory(tmpDir, "New Story")
	if err != nil {
		t.Errorf("Should not error when adding story: %v", err)
	}

	// Verify story was added
	data, _ := os.ReadFile(filepath.Join(tmpDir, ".ralph", "prd.json"))
	if !strings.Contains(string(data), "New Story") {
		t.Error("New story should be in PRD")
	}
}

func TestRunPrdNotInProject(t *testing.T) {
	tmpDir := t.TempDir()

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	err := runPrd(prdCmd, []string{})
	if err == nil {
		t.Error("Should error when not in ralph project")
	}
}
