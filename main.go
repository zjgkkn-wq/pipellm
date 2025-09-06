package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	APIKey  string   `yaml:"api_key"`
	Prompts []Prompt `yaml:"prompts"`
}

type Prompt struct {
	Name   string `yaml:"name"`
	Prompt string `yaml:"prompt"`
}

type ChatGPTRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatGPTResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--bash-alias" {
		generateAliases()
		return
	}

	var promptName string
	if len(os.Args) > 1 {
		// Called with alias name as argument
		promptName = os.Args[1]
	} else {
		// Called directly by binary name
		promptName = filepath.Base(os.Args[0])
	}

	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	prompt := findPrompt(config, promptName)
	if prompt == "" {
		fmt.Fprintf(os.Stderr, "No prompt found for name: %s\n", promptName)
		os.Exit(1)
	}

	input := readStdin()
	response, err := callChatGPT(config.APIKey, prompt, input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error calling ChatGPT: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(response)
}

func loadConfig() (*Config, error) {
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
	if err := parseYAML(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	return &config, nil
}

func parseYAML(data []byte, v interface{}) error {
	lines := strings.Split(string(data), "\n")
	config := v.(*Config)

	for _, line := range lines {
		originalLine := line
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "api_key:") {
			config.APIKey = strings.TrimSpace(strings.TrimPrefix(line, "api_key:"))
		} else if strings.HasPrefix(line, "- name:") {
			name := strings.TrimSpace(strings.TrimPrefix(line, "- name:"))
			prompt := Prompt{Name: name}
			config.Prompts = append(config.Prompts, prompt)
		} else if strings.HasPrefix(line, "prompt:") && strings.HasPrefix(originalLine, "  ") {
			if len(config.Prompts) > 0 {
				promptText := strings.TrimSpace(strings.TrimPrefix(line, "prompt:"))
				config.Prompts[len(config.Prompts)-1].Prompt = promptText
			}
		}
	}

	return nil
}

func findPrompt(config *Config, binaryName string) string {
	for _, p := range config.Prompts {
		if strings.ToLower(strings.TrimSpace(p.Name)) == strings.ToLower(strings.TrimSpace(binaryName)) {
			return p.Prompt
		}
	}
	return ""
}

func readStdin() string {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// Terminal mode - no piped input
		return ""
	}

	var input strings.Builder
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input.WriteString(scanner.Text() + "\n")
	}
	return strings.TrimSpace(input.String())
}

func callChatGPT(apiKey, prompt, input string) (string, error) {
	fullPrompt := prompt
	if input != "" {
		fullPrompt = prompt + "\n\n" + input
	}

	reqBody := ChatGPTRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "user", Content: fullPrompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var chatResp ChatGPTResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from ChatGPT")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func generateAliases() {
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting executable path: %v\n", err)
		os.Exit(1)
	}

	for _, prompt := range config.Prompts {
		alias := strings.ToLower(prompt.Name)
		fmt.Printf("alias %s='%s %s'\n", alias, execPath, alias)
	}
}
