package prd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PRD represents a Product Requirement Document
type PRD struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	UserStories []Story `json:"userStories"`
}

// Story represents a user story in the PRD
type Story struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
	Passes             bool     `json:"passes"`
}

// PRDPath returns the path to the PRD file for a project
func PRDPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".ralph", "prd.json")
}

// Load loads a PRD from disk
func Load(projectRoot string) (*PRD, error) {
	path := PRDPath(projectRoot)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read PRD: %w", err)
	}

	var prd PRD
	if err := json.Unmarshal(data, &prd); err != nil {
		return nil, fmt.Errorf("failed to parse PRD: %w", err)
	}

	return &prd, nil
}

// Save saves a PRD to disk
func Save(projectRoot string, prd *PRD) error {
	path := PRDPath(projectRoot)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(prd, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal PRD: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// GetCurrentStory returns the first non-completed story
func (p *PRD) GetCurrentStory() *Story {
	for i := range p.UserStories {
		if !p.UserStories[i].Passes {
			return &p.UserStories[i]
		}
	}
	return nil
}

// CurrentStory returns the title of the current story, or "none" if all complete
func (p *PRD) CurrentStory() string {
	story := p.GetCurrentStory()
	if story == nil {
		return "none"
	}
	return story.Title
}

// Progress returns the completion progress as "done/total"
func (p *PRD) Progress() string {
	done := 0
	for _, story := range p.UserStories {
		if story.Passes {
			done++
		}
	}
	return fmt.Sprintf("%d/%d", done, len(p.UserStories))
}

// ProgressPercent returns the completion percentage
func (p *PRD) ProgressPercent() int {
	if len(p.UserStories) == 0 {
		return 0
	}
	done := 0
	for _, story := range p.UserStories {
		if story.Passes {
			done++
		}
	}
	return (done * 100) / len(p.UserStories)
}

// MarkStoryComplete marks a story as complete
func (p *PRD) MarkStoryComplete(storyID string) bool {
	for i := range p.UserStories {
		if p.UserStories[i].ID == storyID {
			p.UserStories[i].Passes = true
			return true
		}
	}
	return false
}

// AddStory adds a new story to the PRD
func (p *PRD) AddStory(story Story) {
	// Generate ID if not provided
	if story.ID == "" {
		story.ID = fmt.Sprintf("%d", len(p.UserStories)+1)
	}
	p.UserStories = append(p.UserStories, story)
}

// IsComplete returns true if all stories are complete
func (p *PRD) IsComplete() bool {
	for _, story := range p.UserStories {
		if !story.Passes {
			return false
		}
	}
	return len(p.UserStories) > 0
}
