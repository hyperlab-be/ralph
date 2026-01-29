package prd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	prd, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Expected no error for non-existent PRD, got: %v", err)
	}
	if prd != nil {
		t.Error("Expected nil PRD for non-existent file")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()

	prd := &PRD{
		Name:        "Test Project",
		Description: "A test project",
		UserStories: []Story{
			{
				ID:                 "1",
				Title:              "First Story",
				Description:        "Do something",
				AcceptanceCriteria: []string{"It works"},
				Passes:             false,
			},
		},
	}

	// Save
	err := Save(tmpDir, prd)
	if err != nil {
		t.Fatalf("Failed to save PRD: %v", err)
	}

	// Verify file exists
	path := PRDPath(tmpDir)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("PRD file was not created")
	}

	// Load
	loaded, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load PRD: %v", err)
	}

	if loaded.Name != prd.Name {
		t.Errorf("Expected name '%s', got '%s'", prd.Name, loaded.Name)
	}

	if len(loaded.UserStories) != 1 {
		t.Errorf("Expected 1 story, got %d", len(loaded.UserStories))
	}
}

func TestGetCurrentStory(t *testing.T) {
	prd := &PRD{
		UserStories: []Story{
			{ID: "1", Title: "First", Passes: true},
			{ID: "2", Title: "Second", Passes: false},
			{ID: "3", Title: "Third", Passes: false},
		},
	}

	current := prd.GetCurrentStory()
	if current == nil {
		t.Fatal("Expected current story, got nil")
	}
	if current.ID != "2" {
		t.Errorf("Expected story 2, got %s", current.ID)
	}
}

func TestGetCurrentStoryAllComplete(t *testing.T) {
	prd := &PRD{
		UserStories: []Story{
			{ID: "1", Title: "First", Passes: true},
			{ID: "2", Title: "Second", Passes: true},
		},
	}

	current := prd.GetCurrentStory()
	if current != nil {
		t.Error("Expected nil when all stories complete")
	}
}

func TestProgress(t *testing.T) {
	tests := []struct {
		name     string
		stories  []Story
		expected string
	}{
		{
			name:     "empty",
			stories:  []Story{},
			expected: "0/0",
		},
		{
			name: "none complete",
			stories: []Story{
				{Passes: false},
				{Passes: false},
			},
			expected: "0/2",
		},
		{
			name: "some complete",
			stories: []Story{
				{Passes: true},
				{Passes: false},
				{Passes: true},
			},
			expected: "2/3",
		},
		{
			name: "all complete",
			stories: []Story{
				{Passes: true},
				{Passes: true},
			},
			expected: "2/2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prd := &PRD{UserStories: tt.stories}
			if got := prd.Progress(); got != tt.expected {
				t.Errorf("Progress() = %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestProgressPercent(t *testing.T) {
	tests := []struct {
		name     string
		stories  []Story
		expected int
	}{
		{
			name:     "empty",
			stories:  []Story{},
			expected: 0,
		},
		{
			name: "50%",
			stories: []Story{
				{Passes: true},
				{Passes: false},
			},
			expected: 50,
		},
		{
			name: "100%",
			stories: []Story{
				{Passes: true},
				{Passes: true},
			},
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prd := &PRD{UserStories: tt.stories}
			if got := prd.ProgressPercent(); got != tt.expected {
				t.Errorf("ProgressPercent() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestMarkStoryComplete(t *testing.T) {
	prd := &PRD{
		UserStories: []Story{
			{ID: "1", Passes: false},
			{ID: "2", Passes: false},
		},
	}

	result := prd.MarkStoryComplete("1")
	if !result {
		t.Error("Expected true when marking existing story")
	}
	if !prd.UserStories[0].Passes {
		t.Error("Story 1 should be marked as complete")
	}

	result = prd.MarkStoryComplete("999")
	if result {
		t.Error("Expected false when marking non-existent story")
	}
}

func TestAddStory(t *testing.T) {
	prd := &PRD{}

	prd.AddStory(Story{Title: "First Story"})
	prd.AddStory(Story{Title: "Second Story"})

	if len(prd.UserStories) != 2 {
		t.Errorf("Expected 2 stories, got %d", len(prd.UserStories))
	}

	// Auto-generated IDs
	if prd.UserStories[0].ID != "1" {
		t.Errorf("Expected ID '1', got '%s'", prd.UserStories[0].ID)
	}
	if prd.UserStories[1].ID != "2" {
		t.Errorf("Expected ID '2', got '%s'", prd.UserStories[1].ID)
	}
}

func TestIsComplete(t *testing.T) {
	tests := []struct {
		name     string
		prd      *PRD
		expected bool
	}{
		{
			name:     "empty",
			prd:      &PRD{},
			expected: false,
		},
		{
			name: "not complete",
			prd: &PRD{
				UserStories: []Story{{Passes: false}},
			},
			expected: false,
		},
		{
			name: "complete",
			prd: &PRD{
				UserStories: []Story{{Passes: true}},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.prd.IsComplete(); got != tt.expected {
				t.Errorf("IsComplete() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPRDPath(t *testing.T) {
	path := PRDPath("/project")
	expected := filepath.Join("/project", ".ralph", "prd.json")
	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}
}
