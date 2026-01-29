package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// GlobalConfig represents the global ralph configuration
type GlobalConfig struct {
	Defaults DefaultsConfig `toml:"defaults"`
	Agent    AgentConfig    `toml:"agent"`
}

type DefaultsConfig struct {
	Model         string `toml:"model"`
	MaxIterations int    `toml:"max_iterations"`
	ProjectsDir   string `toml:"projects_dir"`
}

type AgentConfig struct {
	APIKey        string `toml:"api_key"`
	Model         string `toml:"model"`
	MaxIterations int    `toml:"max_iterations"`
	Prompt        string `toml:"prompt"`
}

// ProjectConfig represents project-specific configuration (ralph.toml)
type ProjectConfig struct {
	Project  ProjectInfo  `toml:"project"`
	Worktree WorktreeInfo `toml:"worktree"`
	Hooks    HooksConfig  `toml:"hooks"`
	Agent    AgentConfig  `toml:"agent"`
}

type ProjectInfo struct {
	Name string `toml:"name"`
}

type WorktreeInfo struct {
	Prefix string `toml:"prefix"`
}

type HooksConfig struct {
	Setup   string `toml:"setup"`
	Cleanup string `toml:"cleanup"`
}

// LoopsRegistry holds all registered loops
type LoopsRegistry struct {
	Loops map[string]*Loop `json:"loops"`
}

// Loop represents a single development loop
type Loop struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Project string `json:"project"`
	Feature string `json:"feature"`
	Branch  string `json:"branch"`
	Status  string `json:"status"`
	PID     int    `json:"pid,omitempty"`
	Created string `json:"created,omitempty"`
	Started string `json:"started,omitempty"`
	Stopped string `json:"stopped,omitempty"`
}

// Paths
func ConfigDir() string {
	dir := os.Getenv("RALPH_CONFIG_DIR")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config", "ralph")
	}
	return dir
}

func LoopsFile() string {
	return filepath.Join(ConfigDir(), "loops.json")
}

func GlobalConfigFile() string {
	return filepath.Join(ConfigDir(), "config.toml")
}

// LoadGlobalConfig loads the global configuration
func LoadGlobalConfig() (*GlobalConfig, error) {
	cfg := &GlobalConfig{
		Defaults: DefaultsConfig{
			Model:         "claude-sonnet-4-20250514",
			MaxIterations: 10,
			ProjectsDir:   "~/Code",
		},
	}

	path := GlobalConfigFile()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	_, err := toml.DecodeFile(path, cfg)
	return cfg, err
}

// LoadProjectConfig loads project configuration from ralph.toml
func LoadProjectConfig(projectRoot string) (*ProjectConfig, error) {
	cfg := &ProjectConfig{}
	path := filepath.Join(projectRoot, "ralph.toml")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	_, err := toml.DecodeFile(path, cfg)
	return cfg, err
}

// LoadLoops loads the loops registry
func LoadLoops() (*LoopsRegistry, error) {
	registry := &LoopsRegistry{
		Loops: make(map[string]*Loop),
	}

	path := LoopsFile()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return registry, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, registry)
	return registry, err
}

// SaveLoops saves the loops registry
func SaveLoops(registry *LoopsRegistry) error {
	path := LoopsFile()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// GetLoop returns a loop by name
func GetLoop(name string) (*Loop, error) {
	registry, err := LoadLoops()
	if err != nil {
		return nil, err
	}
	return registry.Loops[name], nil
}

// SetLoop updates or adds a loop
func SetLoop(loop *Loop) error {
	registry, err := LoadLoops()
	if err != nil {
		return err
	}
	registry.Loops[loop.Name] = loop
	return SaveLoops(registry)
}

// RemoveLoop removes a loop from the registry
func RemoveLoop(name string) error {
	registry, err := LoadLoops()
	if err != nil {
		return err
	}
	delete(registry.Loops, name)
	return SaveLoops(registry)
}

// FindProjectRoot finds the project root (directory with ralph.toml or .ralph/)
func FindProjectRoot(start string) (string, error) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "ralph.toml")); err == nil {
			return dir, nil
		}
		if _, err := os.Stat(filepath.Join(dir, ".ralph")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
