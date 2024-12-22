// Package config provides functionality for managing persistent settings in JSON configuration files.
// It supports organizing settings into sections, similar to INI files, but using JSON as the storage
// format. Each section is a top-level key in the JSON object containing key-value pairs.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config manages persistent settings in a JSON configuration file. It provides
// section-based organization of settings where each section is a top-level key
// in a JSON object. This allows logical grouping of related settings while
// maintaining a simple, flat storage structure.
type Config struct {
	path string
	data map[string]map[string]string
}

// Section provides access to a specific configuration section. Each section
// is a collection of key-value pairs under a common namespace (the section name).
// This allows for organizing related settings together and avoiding key collisions.
type Section struct {
	config *Config
	key    string
}

// New creates a new Config instance. If no path is provided (empty string),
// it defaults to ".sandworm" in the current directory. The function will
// attempt to load existing configuration from the specified path but won't
// fail if the file doesn't exist - this allows for gradual configuration
// building where settings are added as needed.
//
// The function will return an error only if the configuration file exists
// but cannot be read or contains invalid JSON.
func New(path string) (*Config, error) {
	if path == "" {
		path = ".sandworm"
	}

	config := &Config{
		path: path,
		data: make(map[string]map[string]string),
	}

	if err := config.load(); err != nil {
		// We don't treat a missing configuration file as an error because
		// it's a normal scenario when first running the application or when
		// a user wants to start with a fresh configuration. The file will
		// be created when the first settings are saved.
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	return config, nil
}

// GetSection returns a configuration section for the specified key. If the
// section doesn't exist, it's created automatically. This allows for
// adding new sections on demand without explicit initialization.
//
// Example:
//
//	config, _ := config.New("")
//	claudeSection := config.GetSection("claude")
//	claudeSection.Set("api_key", "abc123")
func (c *Config) GetSection(key string) *Section {
	if _, exists := c.data[key]; !exists {
		c.data[key] = make(map[string]string)
	}
	return &Section{config: c, key: key}
}

// load reads and parses the configuration file. It's called automatically
// by New() but can also be used to reload configuration from disk if needed.
// The function expects the file to contain a valid JSON object where each
// top-level key represents a section containing key-value pairs.
func (c *Config) load() error {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &c.data)
}

// Save writes the current configuration state to disk in JSON format.
// It ensures the target directory exists and creates it if necessary.
// The resulting file uses standard file permissions (0644) to ensure
// it's readable by the user and group but protected from other users.
func (c *Config) Save() error {
	data, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure the parent directory exists before attempting to write
	// This handles cases where the config file is in a subdirectory
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(c.path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Get retrieves a value from the section. If the key doesn't exist,
// it returns an empty string.
func (s *Section) Get(key string) string {
	return s.config.data[s.key][key]
}

// Set stores a value in the section. The change is only held in memory
// until Save() is called to persist it to disk.
func (s *Section) Set(key, value string) {
	s.config.data[s.key][key] = value
}

// Save persists the current configuration to disk. This is a convenience
// method that calls Save() on the underlying Config instance.
func (s *Section) Save() error {
	return s.config.Save()
}
