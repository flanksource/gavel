package todos

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/todos/types"
)

func TestParseTODO_ValidFrontmatter(t *testing.T) {
	// Input: TODO with valid frontmatter (priority: high, language: go)
	content := `---
priority: high
status: pending
attempts: 0
language: go
---

# TODO: Test Fix

## Steps to Reproduce

### command: test-command

` + "```" + `bash
go test ./pkg
` + "```" + `

## Implementation

Fix the test

## Verification

### command: verify-test

` + "```" + `bash
go test ./pkg
` + "```" + ``

	// Create temp file
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse TODO
	todo, err := ParseTODO(todoPath)
	if err != nil {
		t.Fatalf("Failed to parse TODO: %v", err)
	}

	// Verify fields
	if todo.Priority != types.PriorityHigh {
		t.Errorf("Expected priority high, got %v", todo.Priority)
	}
	if todo.Status != types.StatusPending {
		t.Errorf("Expected status pending, got %v", todo.Status)
	}
	if todo.Language != types.LanguageGo {
		t.Errorf("Expected language go, got %v", todo.Language)
	}
	if todo.Attempts != 0 {
		t.Errorf("Expected attempts 0, got %v", todo.Attempts)
	}
}

func TestParseTODO_MissingPriority(t *testing.T) {
	// Input: TODO without priority field
	content := `---
status: pending
attempts: 0
language: go
---

# TODO: Test`

	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseTODO(todoPath)
	if err == nil {
		t.Error("Expected error for missing priority, got nil")
	}
	if err != nil && err.Error() != "missing required field: priority" {
		t.Errorf("Expected 'missing required field: priority', got %v", err)
	}
}

func TestParseTODO_InvalidStatus(t *testing.T) {
	// Input: TODO with status: "invalid"
	content := `---
priority: high
status: invalid
attempts: 0
language: go
---

# TODO: Test`

	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := ParseTODO(todoPath)
	if err == nil {
		t.Error("Expected error for invalid status, got nil")
	}
}

func TestParseTODO_ExtractSections(t *testing.T) {
	// Input: TODO with "Steps to Reproduce", "Implementation", "Verification" sections
	// NOTE: Parser requires code blocks to extract metadata from frontmatter
	content := "---\npriority: high\nstatus: pending\nattempts: 0\nlanguage: go\n---\n\n# TODO: Test\n\n## Steps to Reproduce\n\n```bash\necho reproduction\n```\n\n## Implementation\n\nSome implementation instructions\n\n## Verification\n\n```bash\necho verification\n```\n"

	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	todo, err := ParseTODO(todoPath)
	if err != nil {
		t.Fatalf("Failed to parse TODO: %v", err)
	}

	// Verify sections exist
	if len(todo.StepsToReproduce) == 0 {
		t.Error("Expected StepsToReproduce section to be extracted")
	}
	if len(todo.Verification) == 0 {
		t.Error("Expected Verification section to be extracted")
	}
}
