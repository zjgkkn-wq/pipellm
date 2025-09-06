package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".pipellm.yaml")

	configContent := `api_key: test_api_key_12345
prompts:
- name: test1
  prompt: This is test prompt 1
- name: test2
  prompt: >
    This is a multi-line
    test prompt 2
- name: CaseSensitive
  prompt: Case test prompt
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Temporarily change HOME directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Test LoadConfig
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Test API key
	if config.APIKey != "test_api_key_12345" { // pragma: allowlist secret
		t.Errorf("Expected API key 'test_api_key_12345', got '%s'", config.APIKey)
	}

	// Test prompts count
	if len(config.Prompts) != 3 {
		t.Errorf("Expected 3 prompts, got %d", len(config.Prompts))
	}

	// Test first prompt
	if config.Prompts[0].Name != "test1" {
		t.Errorf("Expected first prompt name 'test1', got '%s'", config.Prompts[0].Name)
	}
	if config.Prompts[0].Prompt != "This is test prompt 1" {
		t.Errorf("Expected first prompt content 'This is test prompt 1', got '%s'", config.Prompts[0].Prompt)
	}

	// Test multi-line prompt - YAML library handles folded scalars correctly
	expectedMultiLine := "This is a multi-line test prompt 2"
	if strings.TrimSpace(config.Prompts[1].Prompt) != expectedMultiLine {
		t.Errorf("Expected multi-line prompt '%s', got '%s'", expectedMultiLine, strings.TrimSpace(config.Prompts[1].Prompt))
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	// Create empty temp directory
	tempDir := t.TempDir()

	// Temporarily change HOME directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Test LoadConfig with missing file
	_, err := LoadConfig()
	if err == nil {
		t.Fatal("Expected error when config file doesn't exist, got nil")
	}

	if !strings.Contains(err.Error(), "config file not found") {
		t.Errorf("Expected 'config file not found' error, got '%v'", err)
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	// Create a temporary invalid config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".pipellm.yaml")

	invalidYAML := `api_key: test_key
prompts:
  - name: test
    prompt: valid
  invalid_yaml_structure: [unclosed array
`

	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Temporarily change HOME directory
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Test LoadConfig with invalid YAML
	_, err = LoadConfig()
	if err == nil {
		t.Fatal("Expected error when YAML is invalid, got nil")
	}

	if !strings.Contains(err.Error(), "failed to parse config") {
		t.Errorf("Expected 'failed to parse config' error, got '%v'", err)
	}
}

func TestConfigFindPrompt(t *testing.T) {
	config := &Config{
		APIKey: "test_key",
		Prompts: []Prompt{
			{Name: "test1", Prompt: "First test prompt"},
			{Name: "Test2", Prompt: "Second test prompt"},
			{Name: "CAPS", Prompt: "Caps test prompt"},
			{Name: "  spaced  ", Prompt: "Spaced test prompt"},
		},
	}

	tests := []struct {
		name     string
		search   string
		expected string
	}{
		{
			name:     "exact match lowercase",
			search:   "test1",
			expected: "First test prompt",
		},
		{
			name:     "case insensitive match",
			search:   "test2",
			expected: "Second test prompt",
		},
		{
			name:     "case insensitive caps",
			search:   "caps",
			expected: "Caps test prompt",
		},
		{
			name:     "trimmed spaces",
			search:   "spaced",
			expected: "Spaced test prompt",
		},
		{
			name:     "not found",
			search:   "nonexistent",
			expected: "",
		},
		{
			name:     "empty search",
			search:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.FindPrompt(tt.search)
			if result != tt.expected {
				t.Errorf("FindPrompt(%q) = %q, expected %q", tt.search, result, tt.expected)
			}
		})
	}
}
