package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"
	client := NewClient(apiKey)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.apiKey != apiKey {
		t.Errorf("Expected API key %q, got %q", apiKey, client.apiKey)
	}

	if client.httpClient == nil {
		t.Error("HTTP client should not be nil")
	}
}

func TestClientSendPrompt(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify headers
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got %q", contentType)
		}

		auth := r.Header.Get("Authorization")
		expectedAuth := "Bearer test-api-key"
		if auth != expectedAuth {
			t.Errorf("Expected Authorization %q, got %q", expectedAuth, auth)
		}

		// Read and verify request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		var req Request
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("Failed to parse request JSON: %v", err)
		}

		if req.Model != "gpt-3.5-turbo" {
			t.Errorf("Expected model 'gpt-3.5-turbo', got %q", req.Model)
		}

		if len(req.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(req.Messages))
		}

		if req.Messages[0].Role != "user" {
			t.Errorf("Expected role 'user', got %q", req.Messages[0].Role)
		}

		expectedContent := "Test prompt\n\nTest input"
		if req.Messages[0].Content != expectedContent {
			t.Errorf("Expected content %q, got %q", expectedContent, req.Messages[0].Content)
		}

		// Send mock response
		mockResponse := Response{
			Choices: []Choice{
				{
					Message: Message{
						Role:    "assistant",
						Content: "Test response from AI",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client and replace URL for testing
	client := NewClient("test-api-key")

	// We need to modify the client to use our test server
	// For this, we'll temporarily modify the SendPrompt method's URL
	originalTransport := client.httpClient.Transport
	client.httpClient.Transport = &mockTransport{
		server: server,
		orig:   originalTransport,
	}

	// Test SendPrompt
	response, err := client.SendPrompt("Test prompt", "Test input")
	if err != nil {
		t.Fatalf("SendPrompt failed: %v", err)
	}

	expectedResponse := "Test response from AI"
	if response != expectedResponse {
		t.Errorf("Expected response %q, got %q", expectedResponse, response)
	}
}

func TestClientSendPromptNoInput(t *testing.T) {
	// Mock server for testing prompt without input
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and verify request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read request body: %v", err)
		}

		var req Request
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("Failed to parse request JSON: %v", err)
		}

		// When no input is provided, content should be just the prompt
		expectedContent := "Test prompt only"
		if req.Messages[0].Content != expectedContent {
			t.Errorf("Expected content %q, got %q", expectedContent, req.Messages[0].Content)
		}

		// Send mock response
		mockResponse := Response{
			Choices: []Choice{
				{
					Message: Message{
						Role:    "assistant",
						Content: "Response to prompt only",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClient("test-api-key")
	client.httpClient.Transport = &mockTransport{
		server: server,
		orig:   client.httpClient.Transport,
	}

	// Test SendPrompt with empty input
	response, err := client.SendPrompt("Test prompt only", "")
	if err != nil {
		t.Fatalf("SendPrompt failed: %v", err)
	}

	expectedResponse := "Response to prompt only"
	if response != expectedResponse {
		t.Errorf("Expected response %q, got %q", expectedResponse, response)
	}
}

func TestClientSendPromptNoChoices(t *testing.T) {
	// Mock server that returns empty choices
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse := Response{
			Choices: []Choice{}, // Empty choices
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client := NewClient("test-api-key")
	client.httpClient.Transport = &mockTransport{
		server: server,
		orig:   client.httpClient.Transport,
	}

	// Test SendPrompt with no choices in response
	_, err := client.SendPrompt("Test prompt", "Test input")
	if err == nil {
		t.Fatal("Expected error when no choices in response, got nil")
	}

	if !strings.Contains(err.Error(), "no response from ChatGPT") {
		t.Errorf("Expected 'no response from ChatGPT' error, got %v", err)
	}
}

func TestClientSendPromptInvalidJSON(t *testing.T) {
	// Mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json response"))
	}))
	defer server.Close()

	client := NewClient("test-api-key")
	client.httpClient.Transport = &mockTransport{
		server: server,
		orig:   client.httpClient.Transport,
	}

	// Test SendPrompt with invalid JSON response
	_, err := client.SendPrompt("Test prompt", "Test input")
	if err == nil {
		t.Fatal("Expected error when response is invalid JSON, got nil")
	}
}

// mockTransport redirects requests to our test server
type mockTransport struct {
	server *httptest.Server
	orig   http.RoundTripper
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Replace the URL to point to our test server
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(t.server.URL, "http://")

	if t.orig != nil {
		return t.orig.RoundTrip(req)
	}
	return http.DefaultTransport.RoundTrip(req)
}
