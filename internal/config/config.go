// Package config provides functionality for managing persistent settings in JSON configuration files.
// It supports organizing settings into sections, similar to INI files, but using JSON as the storage
// format. Each section is a top-level key in the JSON object containing key-value pairs.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Config manages application configuration, automatically storing values in either
// global or project-specific locations based on the key.
type Config struct {
	globalPath  string
	projectPath string
	global      map[string]map[string]string
	project     map[string]map[string]string
}

// Specify shared keys. These are stored in the global configuration file and are accessible
// to all sandworm projects.
var globalKeys = map[string]bool{
	"claude.session_key": true,
}

// New creates a new Config instance. If projectPath is empty, only global config
// is used. Global config is stored in ~/.config/sandworm/config.json, while
// project config is stored in .sandworm in the project directory.
func New(projectPath string) (*Config, error) {
	globalPath, err := getGlobalConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine global config path: %w", err)
	}

	config := &Config{
		globalPath:  filepath.Join(globalPath, "config.json"),
		projectPath: filepath.Join(projectPath, ".sandworm"),
		global:      make(map[string]map[string]string),
		project:     make(map[string]map[string]string),
	}

	// Load global config
	if err := config.loadGlobal(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	// Load project config if path provided
	if projectPath != "" {
		if err := config.loadProject(); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load project config: %w", err)
		}
	}

	return config, nil
}

// Has checks if a configuration key exists
func (c *Config) Has(key string) bool {
	section, subKey := splitKey(key)
	if globalKeys[key] {
		sectionData, exists := c.global[section]
		if !exists {
			return false
		}
		_, exists = sectionData[subKey]
		return exists
	}
	sectionData, exists := c.project[section]
	if !exists {
		return false
	}
	_, exists = sectionData[subKey]
	return exists
}

// Get retrieves a configuration value. Returns empty string if not found.
func (c *Config) Get(key string) string {
	section, subKey := splitKey(key)
	if globalKeys[key] {
		return c.global[section][subKey]
	}
	return c.project[section][subKey]
}

// Set stores a configuration value and persists it to the appropriate location
func (c *Config) Set(key, value string) error {
	section, subKey := splitKey(key)
	if globalKeys[key] {
		if _, exists := c.global[section]; !exists {
			c.global[section] = make(map[string]string)
		}
		c.global[section][subKey] = value
		return c.saveGlobal()
	}
	if _, exists := c.project[section]; !exists {
		c.project[section] = make(map[string]string)
	}
	c.project[section][subKey] = value
	return c.saveProject()
}

// Delete removes a configuration value
func (c *Config) Delete(key string) error {
	section, subKey := splitKey(key)
	if globalKeys[key] {
		if sectionData, exists := c.global[section]; exists {
			delete(sectionData, subKey)
		}
		return c.saveGlobal()
	}
	if sectionData, exists := c.project[section]; exists {
		delete(sectionData, subKey)
	}
	return c.saveProject()
}

// GetAllKeys returns all configuration keys as a slice of strings
func (c *Config) GetAllKeys() []string {
	var keys []string

	// Add global keys
	for section, sectionData := range c.global {
		for subKey := range sectionData {
			keys = append(keys, section+"."+subKey)
		}
	}

	// Add project keys
	for section, sectionData := range c.project {
		for subKey := range sectionData {
			keys = append(keys, section+"."+subKey)
		}
	}

	return keys
}

// IsGlobalKey checks if a key is stored in global config
func (c *Config) IsGlobalKey(key string) bool {
	return globalKeys[key]
}

// MARK: Internal helper functions

func splitKey(key string) (section, subKey string) {
	parts := strings.SplitN(key, ".", 2)
	if len(parts) != 2 {
		return "", key
	}
	return parts[0], parts[1]
}

func (c *Config) loadGlobal() error {
	return c.load(c.globalPath, c.global)
}

func (c *Config) loadProject() error {
	return c.load(c.projectPath, c.project)
}

func (c *Config) load(path string, data map[string]map[string]string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, &data)
}

func (c *Config) saveGlobal() error {
	return c.save(c.globalPath, c.global)
}

func (c *Config) saveProject() error {
	return c.save(c.projectPath, c.project)
}

func (c *Config) save(path string, data map[string]map[string]string) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func getGlobalConfigPath() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "darwin", "linux", "freebsd", "openbsd", "netbsd":
		// Check XDG_CONFIG_HOME first
		if xdgHome := os.Getenv("XDG_CONFIG_HOME"); xdgHome != "" {
			configDir = xdgHome
		} else {
			// Fall back to ~/.config
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get home directory: %w", err)
			}
			configDir = filepath.Join(home, ".config")
		}

	case "windows":
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}

	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return filepath.Join(configDir, "sandworm"), nil
}
