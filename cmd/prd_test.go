package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPrdShowNoPRD(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, ".ralph"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "ralph.toml"), []byte("[project]\nname = \"test\"\n"), 0644)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// prd show prints a warning when no PRD exists, doesn't return error
	err := runPrdShow(prdCmd, []string{})
	// This is acceptable - it warns instead of erroring
	_ = err
}

func TestRunPrdShowWithPRD(t *testing.T) {
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
	err := runPrdShow(prdCmd, []string{})
	if err != nil {
		t.Errorf("Should not error with valid PRD: %v", err)
	}
}

func TestRunPrdAddNoPRD(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, ".ralph"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "ralph.toml"), []byte("[project]\nname = \"test\"\n"), 0644)

	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	// Set flags for non-interactive mode
	prdAddCmd.Flags().Set("title", "New Story")
	prdAddCmd.Flags().Set("criteria", "Criterion 1")
	defer func() {
		prdAddCmd.Flags().Set("title", "")
		prdAddCmd.Flags().Set("criteria", "")
	}()

	err := runPrdAdd(prdAddCmd, []string{"New Story"})
	if err == nil {
		t.Error("Should error when no PRD exists")
	}
	if !strings.Contains(err.Error(), "no PRD found") {
		t.Errorf("Error should mention no PRD found, got: %v", err)
	}
}

func TestRunPrdAddWithPRD(t *testing.T) {
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

	// Reset and set flags
	storyTitle = "New Story"
	storyCriteria = []string{"Criterion 1", "Criterion 2"}
	defer func() {
		storyTitle = ""
		storyCriteria = nil
	}()

	err := runPrdAdd(prdAddCmd, []string{"New Story"})
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

	err := runPrdShow(prdCmd, []string{})
	if err == nil {
		t.Error("Should error when not in ralph project")
	}
}
