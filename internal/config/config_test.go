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

	configPath := filepath.Join(tmpDir, "config.json")

	t.Run("new config file", func(t *testing.T) {
		cfg, err := New(configPath)
		if err != nil {
			t.Fatalf("Failed to create config: %v", err)
		}

		section := cfg.GetSection("test")
		section.Set("key", "value")

		if err := section.Save(); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Verify file contents
		data, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config file: %v", err)
		}

		var config map[string]map[string]string
		if err := json.Unmarshal(data, &config); err != nil {
			t.Fatalf("Failed to parse config file: %v", err)
		}

		if config["test"]["key"] != "value" {
			t.Errorf("Expected value 'value', got '%s'", config["test"]["key"])
		}
	})

	t.Run("load existing config", func(t *testing.T) {
		cfg, err := New(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		section := cfg.GetSection("test")
		if got := section.Get("key"); got != "value" {
			t.Errorf("Expected value 'value', got '%s'", got)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		if err := os.WriteFile(configPath, []byte("invalid json"), 0644); err != nil {
			t.Fatalf("Failed to write invalid config: %v", err)
		}

		if _, err := New(configPath); err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
	})
}
