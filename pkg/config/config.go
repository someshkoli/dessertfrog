package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// KeyBindingConfig represents a single key binding in the config
type KeyBindingConfig struct {
	Key     string `yaml:"key"`
	Command string `yaml:"command"`
}

// KeyBindingsConfig holds all key bindings configuration
type KeyBindingsConfig struct {
	Global       []KeyBindingConfig `yaml:"global,omitempty"`
	Normal       []KeyBindingConfig `yaml:"normal,omitempty"`
	TableView    []KeyBindingConfig `yaml:"table_view,omitempty"`
	Search       []KeyBindingConfig `yaml:"search,omitempty"`
	InlineSearch []KeyBindingConfig `yaml:"inline_search,omitempty"`
	CommandMode  []KeyBindingConfig `yaml:"command_mode,omitempty"`
	CellEdit     []KeyBindingConfig `yaml:"cell_edit,omitempty"`
	SQLQuery     []KeyBindingConfig `yaml:"sql_query,omitempty"`
	CellPopup    []KeyBindingConfig `yaml:"cell_popup,omitempty"`
	RecordView   []KeyBindingConfig `yaml:"record_view,omitempty"`
	DebugPanel   []KeyBindingConfig `yaml:"debug_panel,omitempty"`
}

// Config represents the application configuration
type Config struct {
	KeyBindings KeyBindingsConfig `yaml:"keybindings,omitempty"`
}

// DefaultConfigPath returns the default config file path
func DefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "dessertfrog", "config.yaml")
}

// LoadConfig loads configuration from a file
// If the file doesn't exist, returns an empty config (will use defaults)
func LoadConfig(path string) (*Config, error) {
	// If no path provided, try default location
	if path == "" {
		path = DefaultConfigPath()
	}

	// If file doesn't exist, return empty config (use defaults)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("sk file not foundn", path)
		return &Config{}, nil
	}

	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// SaveConfig saves configuration to a file
func SaveConfig(path string, cfg *Config) error {
	// If no path provided, use default location
	if path == "" {
		path = DefaultConfigPath()
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
