package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfig(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "sandworm-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "test-project")

	t.Run("new config file", func(t *testing.T) {
		cfg, err := New(configPath)
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		// Set some values
		if err := cfg.Set("claude.session_key", "global-value"); err != nil {
			t.Fatalf("Failed to set global value: %v", err)
		}
		if err := cfg.Set("claude.project_id", "project-value"); err != nil {
			t.Fatalf("Failed to set project value: %v", err)
		}

		// Check global config file
		globalData, err := os.ReadFile(cfg.globalPath)
		if err != nil {
			t.Fatalf("Failed to read global config file: %v", err)
		}

		var globalConfig map[string]map[string]string
		if err := json.Unmarshal(globalData, &globalConfig); err != nil {
			t.Fatalf("Failed to parse global config file: %v", err)
		}

		if globalConfig["claude"]["session_key"] != "global-value" {
			t.Errorf("Expected global value 'global-value', got '%s'", globalConfig["claude"]["session_key"])
		}

		// Check project config file
		projectData, err := os.ReadFile(cfg.projectPath)
		if err != nil {
			t.Fatalf("Failed to read project config file: %v", err)
		}

		var projectConfig map[string]map[string]string
		if err := json.Unmarshal(projectData, &projectConfig); err != nil {
			t.Fatalf("Failed to parse project config file: %v", err)
		}

		if projectConfig["claude"]["project_id"] != "project-value" {
			t.Errorf("Expected project value 'project-value', got '%s'", projectConfig["claude"]["project_id"])
		}
	})

	t.Run("load existing config", func(t *testing.T) {
		cfg, err := New(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Check values persist
		if !cfg.Has("claude.session_key") {
			t.Error("Expected to find global key")
		}
		if got := cfg.Get("claude.session_key"); got != "global-value" {
			t.Errorf("Expected global value 'global-value', got '%s'", got)
		}

		if !cfg.Has("claude.project_id") {
			t.Error("Expected to find project key")
		}
		if got := cfg.Get("claude.project_id"); got != "project-value" {
			t.Errorf("Expected project value 'project-value', got '%s'", got)
		}
	})

	t.Run("delete values", func(t *testing.T) {
		cfg, err := New(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Delete values
		if err := cfg.Delete("claude.session_key"); err != nil {
			t.Fatalf("Failed to delete global value: %v", err)
		}
		if err := cfg.Delete("claude.project_id"); err != nil {
			t.Fatalf("Failed to delete project value: %v", err)
		}

		// Check values are gone
		if cfg.Has("claude.session_key") {
			t.Error("Global key should be deleted")
		}
		if cfg.Has("claude.project_id") {
			t.Error("Project key should be deleted")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		badConfigPath := filepath.Join(tmpDir, "bad-config")
		// Create the directory first
		if err := os.MkdirAll(badConfigPath, 0o755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(
			filepath.Join(badConfigPath, ".sandworm"),
			[]byte("invalid json"),
			0o644,
		); err != nil {
			t.Fatalf("Failed to write invalid config: %v", err)
		}

		if _, err := New(badConfigPath); err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
	})

	t.Run("missing directory", func(t *testing.T) {
		cfg, err := New(filepath.Join(tmpDir, "nonexistent"))
		if err != nil {
			t.Fatalf("Failed to create config with nonexistent directory: %v", err)
		}

		// Should be able to set and get values
		if err := cfg.Set("claude.test_key", "test_value"); err != nil {
			t.Fatalf("Failed to set value: %v", err)
		}

		if got := cfg.Get("claude.test_key"); got != "test_value" {
			t.Errorf("Expected value 'test_value', got '%s'", got)
		}
	})

	t.Run("empty values", func(t *testing.T) {
		cfg, err := New(configPath)
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		// Get nonexistent value
		if got := cfg.Get("nonexistent.key"); got != "" {
			t.Errorf("Expected empty string for nonexistent key, got '%s'", got)
		}

		// Check nonexistent value
		if cfg.Has("nonexistent.key") {
			t.Error("Has() should return false for nonexistent key")
		}
	})
}
