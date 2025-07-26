package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFromPath(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		configYAML  string
		expectError bool
		description string
	}{
		{
			name: "Valid config file",
			configYAML: `
defaultNamespace: "kube-system"
preferredContexts:
  - "prod-cluster"
  - "staging-cluster"
refreshInterval: 10
`,
			expectError: false,
			description: "Should load valid YAML config",
		},
		{
			name: "Invalid YAML",
			configYAML: `
defaultNamespace: "kube-system"
preferredContexts:
  - "prod-cluster"
  - "staging-cluster"
refreshInterval: "invalid"
`,
			expectError: true, // Invalid YAML should cause an error
			description: "Should handle invalid YAML gracefully",
		},
		{
			name:        "Empty file",
			configYAML:  "",
			expectError: false,
			description: "Should handle empty file",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test config file with unique name
			configPath := filepath.Join(tempDir, fmt.Sprintf("test-config-%d.yaml", i))

			// Always create the file, even if empty
			err := os.WriteFile(configPath, []byte(tt.configYAML), 0644)
			if err != nil {
				t.Fatalf("Failed to write test config file: %v", err)
			}

			// Test loading config
			cfg, err := LoadConfigFromPath(configPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify config was loaded
			if cfg == nil {
				t.Errorf("Config should not be nil")
			}

			// For valid config, check some values
			if tt.name == "Valid config file" {
				if cfg.DefaultNamespace != "kube-system" {
					t.Errorf("Expected DefaultNamespace 'kube-system', got '%s'", cfg.DefaultNamespace)
				}
				if len(cfg.PreferredContexts) != 2 {
					t.Errorf("Expected 2 preferred contexts, got %d", len(cfg.PreferredContexts))
				}
				if cfg.RefreshInterval != 10 {
					t.Errorf("Expected RefreshInterval 10, got %d", cfg.RefreshInterval)
				}
			}
		})
	}
}

func TestLoadConfigFromPath_NonExistentFile(t *testing.T) {
	// Test with non-existent file
	_, err := LoadConfigFromPath("/nonexistent/config.yaml")
	if err == nil {
		t.Errorf("Expected error for non-existent file, but got none")
	}

	expectedError := "config file not found: /nonexistent/config.yaml"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}
