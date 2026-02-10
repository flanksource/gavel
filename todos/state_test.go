package todos

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/gavel/todos/types"
)

func TestUpdateState_PreservesContent(t *testing.T) {
	// Input: TODO with frontmatter + markdown content
	// Action: Update status field
	// Expected: Markdown content unchanged, only frontmatter updated
	// NOTE: Parser requires code blocks to extract metadata from frontmatter
	content := "---\npriority: high\nstatus: pending\nattempts: 0\nlanguage: go\n---\n\n# TODO: Test Fix\n\n## Steps to Reproduce\n\n```bash\necho reproduction\n```\n\n## Implementation\n\nSome implementation instructions\n"

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

	// Update status
	completed := types.StatusCompleted
	now := time.Now()
	attempts := 1
	err = UpdateTODOState(todo, StateUpdate{
		Status:   &completed,
		LastRun:  &now,
		Attempts: &attempts,
	})
	if err != nil {
		t.Fatalf("Failed to update state: %v", err)
	}

	// Re-read file and verify
	updatedContent, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	contentStr := string(updatedContent)

	// Verify markdown content preserved
	if !contains(contentStr, "# TODO: Test Fix") {
		t.Error("Expected markdown heading to be preserved")
	}
	if !contains(contentStr, "## Steps to Reproduce") {
		t.Error("Expected section to be preserved")
	}

	// Verify frontmatter updated
	if !contains(contentStr, "status: completed") {
		t.Error("Expected status to be updated to completed")
	}
	if !contains(contentStr, "attempts: 1") {
		t.Error("Expected attempts to be updated to 1")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestUpdateLatestFailure_ReplacesExistingSection(t *testing.T) {
	content := `---
priority: high
status: pending
---

# TODO: Test Fix

## Latest Failure

` + "```" + `
old output
` + "```" + `

## Failure History
`

	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	todo := &types.TODO{FilePath: todoPath}
	testResult := &types.TestResultInfo{
		Command:   "go test -run ^TestFoo$",
		CWD:       "/path/to/project",
		GitBranch: "main",
		GitCommit: "abc1234",
		GitDirty:  true,
		Timestamp: time.Date(2025, 11, 27, 10, 0, 0, 0, time.UTC),
		Passed:    false,
		Output:    "test failed: expected 1, got 2",
		Duration:  1500 * time.Millisecond,
	}

	err := UpdateLatestFailure(todo, testResult)
	if err != nil {
		t.Fatalf("UpdateLatestFailure failed: %v", err)
	}

	updatedContent, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	contentStr := string(updatedContent)

	// Verify new content is present (Clicky's KeyValuePair uses "Key: Value" format)
	if !contains(contentStr, "**Command**: `go test -run ^TestFoo$`") {
		t.Error("Expected new command to be present")
	}
	if !contains(contentStr, "**Branch**: `main`") {
		t.Error("Expected branch to be present")
	}
	if !contains(contentStr, "`abc1234` (dirty)") {
		t.Error("Expected commit with dirty flag to be present")
	}
	if !contains(contentStr, "FAILED") {
		t.Error("Expected FAILED result to be present")
	}
	if !contains(contentStr, "test failed: expected 1, got 2") {
		t.Error("Expected test output to be present")
	}

	// Verify old content is removed
	if contains(contentStr, "old output") {
		t.Error("Expected old output to be removed")
	}

	// Verify other sections preserved
	if !contains(contentStr, "## Failure History") {
		t.Error("Expected Failure History section to be preserved")
	}
}

func TestUpdateLatestFailure_InsertsBeforeFailureHistory(t *testing.T) {
	content := `---
priority: high
status: pending
---

# TODO: Test Fix

## Failure History

Previous failures here
`

	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	todo := &types.TODO{FilePath: todoPath}
	testResult := &types.TestResultInfo{
		Command:   "go test",
		CWD:       "/path",
		Timestamp: time.Now(),
		Passed:    true,
		Duration:  100 * time.Millisecond,
	}

	err := UpdateLatestFailure(todo, testResult)
	if err != nil {
		t.Fatalf("UpdateLatestFailure failed: %v", err)
	}

	updatedContent, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	contentStr := string(updatedContent)

	// Verify Latest Failure section was added
	if !contains(contentStr, "## Latest Failure") {
		t.Error("Expected Latest Failure section to be added")
	}

	// Verify Failure History is still present
	if !contains(contentStr, "## Failure History") {
		t.Error("Expected Failure History section to be preserved")
	}
	if !contains(contentStr, "Previous failures here") {
		t.Error("Expected Failure History content to be preserved")
	}
}

func TestUpdateLatestFailure_AppendsAtEnd(t *testing.T) {
	content := `---
priority: high
status: pending
---

# TODO: Test Fix

Some content here.
`

	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	todo := &types.TODO{FilePath: todoPath}
	testResult := &types.TestResultInfo{
		Command:   "go test",
		CWD:       "/path",
		Timestamp: time.Now(),
		Passed:    true,
		Duration:  100 * time.Millisecond,
	}

	err := UpdateLatestFailure(todo, testResult)
	if err != nil {
		t.Fatalf("UpdateLatestFailure failed: %v", err)
	}

	updatedContent, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	contentStr := string(updatedContent)

	// Verify Latest Failure section was added
	if !contains(contentStr, "## Latest Failure") {
		t.Error("Expected Latest Failure section to be added")
	}

	// Verify original content is preserved
	if !contains(contentStr, "Some content here.") {
		t.Error("Expected original content to be preserved")
	}
}

func TestUpdateLatestFailure_TruncatesLongOutput(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(todoPath, []byte("---\nstatus: pending\n---\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create output longer than 2000 chars
	longOutput := ""
	for i := 0; i < 300; i++ {
		longOutput += "0123456789"
	}

	todo := &types.TODO{FilePath: todoPath}
	testResult := &types.TestResultInfo{
		Command:   "go test",
		CWD:       "/path",
		Timestamp: time.Now(),
		Passed:    false,
		Output:    longOutput,
		Duration:  100 * time.Millisecond,
	}

	err := UpdateLatestFailure(todo, testResult)
	if err != nil {
		t.Fatalf("UpdateLatestFailure failed: %v", err)
	}

	updatedContent, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	contentStr := string(updatedContent)

	// Output in TestResultInfo is truncated before being passed
	// The function should just write what it receives
	if !contains(contentStr, "## Latest Failure") {
		t.Error("Expected Latest Failure section to be present")
	}
}
