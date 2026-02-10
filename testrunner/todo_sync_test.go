package testrunner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/flanksource/gavel/todos/types"
	"github.com/goccy/go-yaml"
)

func TestTodoSyncGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		failure  TestFailure
		expected string
	}{
		{
			name: "simple test name",
			failure: TestFailure{
				Name: "TestUserLogin",
			},
			expected: "testuserlogin",
		},
		{
			name: "ginkgo spec",
			failure: TestFailure{
				Name: "should validate email",
			},
			expected: "should validate email",
		},
	}

	sync := NewTodoSync(t.TempDir(), "")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sync.generateTodoSlug(tt.failure)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTodoSyncGenerateRerunCommand(t *testing.T) {
	tests := []struct {
		name     string
		failure  TestFailure
		contains string
	}{
		{
			name: "go test command",
			failure: TestFailure{
				Name:      "TestUserLogin",
				File:      "pkg/auth/user_test.go",
				Framework: GoTest,
			},
			contains: "go test -run ^TestUserLogin$",
		},
		{
			name: "ginkgo command",
			failure: TestFailure{
				Name:      "should validate email",
				Framework: Ginkgo,
			},
			contains: `ginkgo --focus="should validate email"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.failure.RerunCommand()
			if !strings.Contains(result, tt.contains) {
				t.Errorf("got %q, wanted to contain %q", result, tt.contains)
			}
		})
	}
}

func TestTodoSyncCreateTodo(t *testing.T) {
	tmpDir := t.TempDir()

	failure := TestFailure{
		Name:      "TestUserLogin",
		Package:   "github.com/flanksource/gavel/auth",
		Message:   "expected nil, got error",
		File:      "pkg/auth/user_test.go",
		Line:      42,
		Framework: GoTest,
	}

	sync := NewTodoSync(tmpDir, "")
	todoPath, err := sync.createTodo(failure)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(todoPath); err != nil {
		t.Errorf("todo file not created: %v", err)
	}

	content, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("failed to read todo file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "TestUserLogin") {
		t.Errorf("todo content missing test name")
	}
	if !strings.Contains(contentStr, "go test -run") {
		t.Errorf("todo content missing re-run command")
	}
}

func TestTodoSyncFindExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a TODO file manually (using actual slug format)
	todoPath := filepath.Join(tmpDir, "testuserlogin-001.md")
	if err := os.WriteFile(todoPath, []byte("# TODO\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	failure := TestFailure{
		Name: "TestUserLogin",
	}

	sync := NewTodoSync(tmpDir, "")
	found, err := sync.findExistingTodo(failure)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if found != todoPath {
		t.Errorf("got %q, want %q", found, todoPath)
	}
}

func TestTodoSyncUpdateTodo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create initial TODO with attempts: 1
	initialContent := `---
priority: high
status: pending
attempts: 1
last_run: 2025-01-17T10:00:00Z
language: go
---

# TODO: Fix Test - TestUserLogin

## Test Information
`
	todoPath := filepath.Join(tmpDir, "test-testuserlogin-001.md")
	if err := os.WriteFile(todoPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	failure := TestFailure{
		Name:    "TestUserLogin",
		Message: "still failing",
	}

	sync := NewTodoSync(tmpDir, "")
	err := sync.updateTodo(todoPath, failure)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "attempts: 2") {
		t.Errorf("attempts not incremented, content: %s", contentStr)
	}
	if !strings.Contains(contentStr, "still failing") {
		t.Errorf("failure message not added to history")
	}
	if !strings.Contains(contentStr, "last_run:") {
		t.Errorf("last_run field not found in updated frontmatter")
	}
}

func TestTodoSyncGenerateContentWithTODOFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()

	failure := TestFailure{
		Name:      "TestUserLogin",
		Package:   "github.com/flanksource/gavel/auth",
		Message:   "expected nil, got error",
		File:      "pkg/auth/user_test.go",
		Line:      42,
		Framework: GoTest,
	}

	sync := NewTodoSync(tmpDir, "")
	content := sync.generateTodoContent(failure, "")

	// Check that frontmatter is properly generated
	if !strings.Contains(content, "---") {
		t.Errorf("frontmatter delimiters not found")
	}
	if !strings.Contains(content, "priority: high") {
		t.Errorf("priority field not found")
	}
	if !strings.Contains(content, "status: pending") {
		t.Errorf("status field not found")
	}
	if !strings.Contains(content, "language: go") {
		t.Errorf("language field not found")
	}
	if !strings.Contains(content, "attempts: 1") {
		t.Errorf("attempts field not found")
	}
	if !strings.Contains(content, "last_run:") {
		t.Errorf("last_run field not found")
	}

	// Extract and parse frontmatter to verify it's valid YAML
	lines := strings.Split(content, "\n")
	frontmatterStart := -1
	frontmatterEnd := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if frontmatterStart == -1 {
				frontmatterStart = i
			} else {
				frontmatterEnd = i
				break
			}
		}
	}

	if frontmatterStart == -1 || frontmatterEnd == -1 {
		t.Fatalf("could not find frontmatter boundaries")
	}

	frontmatterLines := lines[frontmatterStart+1 : frontmatterEnd]
	frontmatterYAML := strings.Join(frontmatterLines, "\n")

	var frontmatter TODOFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &frontmatter); err != nil {
		t.Fatalf("failed to parse generated frontmatter as TODOFrontmatter: %v", err)
	}

	// Verify the fields are correctly set
	if frontmatter.Priority != PriorityHigh {
		t.Errorf("expected priority %v, got %v", PriorityHigh, frontmatter.Priority)
	}
	if frontmatter.Status != StatusPending {
		t.Errorf("expected status %v, got %v", StatusPending, frontmatter.Status)
	}
	if frontmatter.Language != LanguageGo {
		t.Errorf("expected language %v, got %v", LanguageGo, frontmatter.Language)
	}
	if frontmatter.Attempts != 1 {
		t.Errorf("expected attempts 1, got %d", frontmatter.Attempts)
	}
	if frontmatter.LastRun == nil {
		t.Errorf("expected last_run to be set")
	}
}
