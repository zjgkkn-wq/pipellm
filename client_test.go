package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func TestNewClient(t *testing.T) {
	// This test now checks if the client is created without errors.
	// A valid API key is not needed for the basic client creation itself,
	// but requests will fail. We test requests separately.
	_, err := NewClient("test-api-key", "gemini-pro") // pragma: allowlist secret
	if err != nil {
		t.Fatalf("NewClient() error = %v, wantErr nil", err)
	}
}

// Helper struct for a more robust check of the request body
type GeminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

func TestClientSendPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "gemini-pro:generateContent") {
			t.Errorf("Expected path to contain 'gemini-pro:generateContent', got %s", r.URL.Path)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		var req GeminiRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}

		if len(req.Contents) != 1 || len(req.Contents[0].Parts) != 1 {
			t.Fatalf("Expected 1 content with 1 part, got %+v", req)
		}

		expectedContent := "Test prompt\n\nTest input"
		if req.Contents[0].Parts[0].Text != expectedContent {
			t.Errorf("Expected content %q, got %q", expectedContent, req.Contents[0].Parts[0].Text)
		}

		mockResponse := `{
			"candidates": [{
				"content": {
					"parts": [{"text": "Test response from AI"}],
					"role": "model"
				}
			}]
		}`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	ctx := context.Background()
	// Use WithEndpoint to redirect requests to our test server.
	client, err := genai.NewClient(ctx,
		option.WithAPIKey("test-api-key"),
		option.WithEndpoint(server.URL),
	)
	if err != nil {
		t.Fatalf("Failed to create test genai client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")
	pipellmClient := &Client{model: model}

	response, err := pipellmClient.SendPrompt("Test prompt", "Test input")
	if err != nil {
		t.Fatalf("SendPrompt failed: %v", err)
	}

	expectedResponse := "Test response from AI"
	if response != expectedResponse {
		t.Errorf("Expected response %q, got %q", expectedResponse, response)
	}
}

func TestClientSendPromptNoInput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		var req GeminiRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("Failed to unmarshal request body: %v", err)
		}

		if len(req.Contents) != 1 || len(req.Contents[0].Parts) != 1 {
			t.Fatalf("Expected 1 content with 1 part, got %+v", req)
		}

		expectedContent := "Test prompt only"
		if req.Contents[0].Parts[0].Text != expectedContent {
			t.Errorf("Expected content %q, got %q", expectedContent, req.Contents[0].Parts[0].Text)
		}

		mockResponse := `{
			"candidates": [{
				"content": {
					"parts": [{"text": "Response to prompt only"}],
					"role": "model"
				}
			}]
		}`
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey("test-api-key"), option.WithEndpoint(server.URL))
	if err != nil {
		t.Fatalf("Failed to create test genai client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")
	pipellmClient := &Client{model: model}

	response, err := pipellmClient.SendPrompt("Test prompt only", "")
	if err != nil {
		t.Fatalf("SendPrompt failed: %v", err)
	}

	expectedResponse := "Response to prompt only"
	if response != expectedResponse {
		t.Errorf("Expected response %q, got %q", expectedResponse, response)
	}
}

func TestClientSendPromptNoChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse := `{"candidates": []}` // Empty candidates
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey("test-api-key"), option.WithEndpoint(server.URL))
	if err != nil {
		t.Fatalf("Failed to create test genai client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")
	pipellmClient := &Client{model: model}

	_, err = pipellmClient.SendPrompt("Test prompt", "Test input")
	if err == nil {
		t.Fatal("Expected error when no choices in response, got nil")
	}

	if !strings.Contains(err.Error(), "no response from Gemini") {
		t.Errorf("Expected 'no response from Gemini' error, got %v", err)
	}
}

func TestClientSendPromptInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json response"))
	}))
	defer server.Close()

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey("test-api-key"), option.WithEndpoint(server.URL))
	if err != nil {
		t.Fatalf("Failed to create test genai client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")
	pipellmClient := &Client{model: model}

	_, err = pipellmClient.SendPrompt("Test prompt", "Test input")
	if err == nil {
		t.Fatal("Expected error when response is invalid JSON, got nil")
	}
}
