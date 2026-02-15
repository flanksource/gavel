package todos

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpsertAttemptsSection_New(t *testing.T) {
	content := "---\nstatus: pending\n---\n\n# My TODO\n\nSome content.\n"
	row := "| 1 | completed | 2026-02-13 15:00 | claude-code | 2m30s | $0.0500 | 12345 | `abc1234` | [transcript](foo.attempts/attempt-1.md) |"

	result := upsertAttemptsSection(content, row)

	assert.Contains(t, result, "## Attempts")
	assert.Contains(t, result, "| # | Status |")
	assert.Contains(t, result, row)
	// Table header should appear before the row
	headerIdx := strings.Index(result, "## Attempts")
	rowIdx := strings.Index(result, row)
	assert.Greater(t, rowIdx, headerIdx)
}

func TestUpsertAttemptsSection_Append(t *testing.T) {
	content := `---
status: pending
---

# My TODO

## Attempts

| # | Status | Date | Model | Duration | Cost | Tokens | Commit | Transcript |
|---|--------|------|-------|----------|------|--------|--------|------------|
| 1 | failed | 2026-02-13 15:00 | claude-code | 2m30s | $0.0500 | 12345 |  | [transcript](foo.attempts/attempt-1.md) |
`

	row := "| 2 | completed | 2026-02-13 15:30 | claude-code | 1m45s | $0.0300 | 9876 | `abc1234` | [transcript](foo.attempts/attempt-2.md) |"
	result := upsertAttemptsSection(content, row)

	assert.Contains(t, result, row)
	// Should have both rows
	assert.Equal(t, 1, strings.Count(result, "## Attempts"))
	assert.Contains(t, result, "| 1 | failed |")
	assert.Contains(t, result, "| 2 | completed |")
}

func TestUpsertAttemptsSection_BeforeNextSection(t *testing.T) {
	content := `---
status: pending
---

# My TODO

## Attempts

| # | Status | Date | Model | Duration | Cost | Tokens | Commit | Transcript |
|---|--------|------|-------|----------|------|--------|--------|------------|
| 1 | failed | 2026-02-13 15:00 | claude-code | 2m30s | $0.0500 | 12345 |  | [transcript](foo.attempts/attempt-1.md) |

## Latest Failure

Some failure info.
`

	row := "| 2 | completed | 2026-02-13 15:30 | claude-code | 1m45s | $0.0300 | 9876 | `abc1234` | [transcript](foo.attempts/attempt-2.md) |"
	result := upsertAttemptsSection(content, row)

	// New row should appear before ## Latest Failure
	rowIdx := strings.Index(result, row)
	failureIdx := strings.Index(result, "## Latest Failure")
	assert.Greater(t, failureIdx, rowIdx)
	assert.Contains(t, result, "Some failure info.")
}

func TestFormatAttemptRow(t *testing.T) {
	attempt := types.Attempt{
		Status:     types.StatusCompleted,
		Timestamp:  time.Date(2026, 2, 13, 15, 30, 0, 0, time.UTC),
		Duration:   105 * time.Second,
		Cost:       0.03,
		Tokens:     9876,
		Model:      "claude-code",
		Commit:     "abc1234",
		Transcript: "foo.attempts/attempt-2.md",
	}

	row := formatAttemptRow(2, attempt)

	assert.Contains(t, row, "| 2 |")
	assert.Contains(t, row, "completed")
	assert.Contains(t, row, "2026-02-13 15:30")
	assert.Contains(t, row, "claude-code")
	assert.Contains(t, row, "$0.0300")
	assert.Contains(t, row, "9876")
	assert.Contains(t, row, "`abc1234`")
	assert.Contains(t, row, "[transcript](foo.attempts/attempt-2.md)")
}

func TestSaveAttempt_WritesTranscriptAndTable(t *testing.T) {
	dir := t.TempDir()
	todoPath := filepath.Join(dir, "my-todo.md")
	require.NoError(t, os.WriteFile(todoPath, []byte("---\npriority: high\nstatus: pending\n---\n\n# My TODO\n\nSome content.\n"), 0644))

	todo := &types.TODO{FilePath: todoPath}
	todo.Attempts = 1
	todo.Status = types.StatusCompleted

	result := &ExecutionResult{
		Success:      true,
		ExecutorName: "claude-code",
		TokensUsed:   12345,
		CostUSD:      0.05,
		Duration:     150 * time.Second,
		CommitSHA:    "abc1234",
		Transcript:   NewExecutionTranscript(),
	}

	err := saveAttempt(todo, result)
	require.NoError(t, err)

	// Check transcript file exists
	transcriptPath := filepath.Join(dir, "my-todo.attempts", "attempt-1.md")
	_, err = os.Stat(transcriptPath)
	assert.NoError(t, err)

	transcriptContent, err := os.ReadFile(transcriptPath)
	require.NoError(t, err)
	assert.Contains(t, string(transcriptContent), "# Attempt 1")
	assert.Contains(t, string(transcriptContent), "claude-code")

	// Check the TODO file has the attempts table
	todoContent, err := os.ReadFile(todoPath)
	require.NoError(t, err)
	assert.Contains(t, string(todoContent), "## Attempts")
	assert.Contains(t, string(todoContent), "| 1 |")
	assert.Contains(t, string(todoContent), "[transcript](my-todo.attempts/attempt-1.md)")
}
