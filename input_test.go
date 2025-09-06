package main

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"
)

func TestReadStdinWithPipedInput(t *testing.T) {
	// Create a pipe to simulate piped input
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Save original stdin and restore it later
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()

	// Replace stdin with our pipe reader
	os.Stdin = r

	// Write test data to the pipe
	testInput := "Line 1\nLine 2\nLine 3\n"
	go func() {
		defer w.Close()
		w.WriteString(testInput)
	}()

	// Test ReadStdin
	result := ReadStdin()

	expected := "Line 1\nLine 2\nLine 3"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestReadStdinWithEmptyInput(t *testing.T) {
	// Create a pipe with empty input
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Save original stdin and restore it later
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()

	// Replace stdin with our pipe reader
	os.Stdin = r

	// Close writer immediately (empty input)
	w.Close()

	// Test ReadStdin
	result := ReadStdin()

	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestReadStdinWithSingleLine(t *testing.T) {
	// Create a pipe with single line input
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Save original stdin and restore it later
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()

	// Replace stdin with our pipe reader
	os.Stdin = r

	// Write single line
	testInput := "Single line input"
	go func() {
		defer w.Close()
		w.WriteString(testInput)
	}()

	// Test ReadStdin
	result := ReadStdin()

	if result != testInput {
		t.Errorf("Expected %q, got %q", testInput, result)
	}
}

func TestReadStdinWithTrailingSpaces(t *testing.T) {
	// Create a pipe with input that has trailing spaces/newlines
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Save original stdin and restore it later
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()

	// Replace stdin with our pipe reader
	os.Stdin = r

	// Write input with trailing spaces and newlines
	testInput := "  Content with spaces  \n\n  \n"
	go func() {
		defer w.Close()
		w.WriteString(testInput)
	}()

	// Test ReadStdin
	result := ReadStdin()

	// TrimSpace removes leading and trailing whitespace, including newlines
	expected := "Content with spaces"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// Note: Testing terminal mode (when input is from terminal, not pipe) is tricky
// because it requires actual terminal interaction. For now, we'll focus on
// testing the piped input scenario which is the main use case for CLI tools.

func TestReadStdinLargeInput(t *testing.T) {
	// Test with larger input to ensure buffering works correctly
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Save original stdin and restore it later
	originalStdin := os.Stdin
	defer func() { os.Stdin = originalStdin }()

	// Replace stdin with our pipe reader
	os.Stdin = r

	// Generate large input (multiple lines)
	var testLines []string
	for i := 0; i < 100; i++ {
		testLines = append(testLines, "Line number "+strings.Repeat("x", 50))
	}
	testInput := strings.Join(testLines, "\n") + "\n"

	go func() {
		defer w.Close()
		// Write in chunks to simulate real input
		writer := bufio.NewWriter(w)
		writer.WriteString(testInput)
		writer.Flush()
	}()

	// Test ReadStdin
	result := ReadStdin()

	expected := strings.Join(testLines, "\n")
	if result != expected {
		t.Errorf("Large input test failed. Expected length %d, got length %d",
			len(expected), len(result))
		if len(result) < 100 {
			t.Errorf("Got: %q", result)
		}
	}
}

// Helper function to create a mock stdin for testing
func createMockStdin(content string) (*os.File, func()) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	originalStdin := os.Stdin
	os.Stdin = r

	go func() {
		defer w.Close()
		io.WriteString(w, content)
	}()

	cleanup := func() {
		os.Stdin = originalStdin
		r.Close()
	}

	return r, cleanup
}

func TestReadStdinIntegration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "multiline text",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "text with trailing newline",
			input:    "Content\n",
			expected: "Content",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   \n  \n   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := createMockStdin(tt.input)
			defer cleanup()

			result := ReadStdin()
			if result != tt.expected {
				t.Errorf("ReadStdin() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
