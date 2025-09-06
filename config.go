package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIKey  string   `yaml:"api_key"`
	Prompts []Prompt `yaml:"prompts"`
}

type Prompt struct {
	Name   string `yaml:"name"`
	Prompt string `yaml:"prompt"`
}

func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".pipellm.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("config file not found at %s", configPath)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	return &config, nil
}

func (c *Config) FindPrompt(name string) string {
	for _, p := range c.Prompts {
		if strings.EqualFold(strings.TrimSpace(p.Name), strings.TrimSpace(name)) {
			return p.Prompt
		}
	}
	return ""
}
