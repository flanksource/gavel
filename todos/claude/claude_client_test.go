package claude

import (
	"strings"
	"testing"
)

func TestExtractChangedFiles(t *testing.T) {
	client := &ClaudeClient{}

	output := `
Starting implementation...
Created file: pkg/service/validator.go
Modified: pkg/service/user.go
Updated: pkg/service/user_test.go
All tests passing
IMPLEMENTATION_COMPLETE
`

	files := client.extractChangedFiles(output)

	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(files))
	}

	expectedFiles := []string{
		"pkg/service/validator.go",
		"pkg/service/user.go",
		"pkg/service/user_test.go",
	}

	for i, expected := range expectedFiles {
		if i >= len(files) {
			t.Errorf("Missing file: %s", expected)
			continue
		}
		if files[i] != expected {
			t.Errorf("Expected file %s, got %s", expected, files[i])
		}
	}
}

func TestExtractChangedFiles_NoFiles(t *testing.T) {
	client := &ClaudeClient{}

	output := `
Implementation complete
IMPLEMENTATION_COMPLETE
`

	files := client.extractChangedFiles(output)

	if len(files) != 0 {
		t.Errorf("Expected 0 files, got %d", len(files))
	}
}

func TestGetSessionLogs(t *testing.T) {
	client := &ClaudeClient{
		sessionLogs: []string{
			"Log entry 1",
			"Log entry 2",
			"Log entry 3",
		},
	}

	logs := client.GetSessionLogs()

	if !strings.Contains(logs, "Log entry 1") {
		t.Error("Expected logs to contain first entry")
	}
	if !strings.Contains(logs, "Log entry 2") {
		t.Error("Expected logs to contain second entry")
	}
	if !strings.Contains(logs, "Log entry 3") {
		t.Error("Expected logs to contain third entry")
	}
	if !strings.Contains(logs, "---") {
		t.Error("Expected logs to contain separator")
	}
}

func TestNewClaudeClient_DefaultTimeout(t *testing.T) {
	// This test just verifies client creation with defaults
	// We can't test actual execution without a real claude binary

	config := ClaudeConfig{
		ClaudePath: "/usr/bin/echo", // Use echo as a dummy executable
	}

	client, err := NewClaudeClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client.timeout == 0 {
		t.Error("Expected default timeout to be set")
	}
}
