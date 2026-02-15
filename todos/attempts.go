package todos

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/gavel/todos/types"
)

func saveAttempt(todo *types.TODO, result *ExecutionResult) error {
	attempt := types.Attempt{
		Timestamp: time.Now(),
		Duration:  result.Duration,
		Cost:      result.CostUSD,
		Tokens:    result.TokensUsed,
		Model:     result.ExecutorName,
		Commit:    result.CommitSHA,
	}
	if result.Success {
		attempt.Status = types.StatusCompleted
	} else {
		attempt.Status = types.StatusFailed
	}

	transcriptPath, err := writeTranscript(todo, result)
	if err != nil {
		return fmt.Errorf("writing transcript: %w", err)
	}
	attempt.Transcript = transcriptPath

	return appendAttemptRow(todo, attempt)
}

func writeTranscript(todo *types.TODO, result *ExecutionResult) (string, error) {
	base := strings.TrimSuffix(filepath.Base(todo.FilePath), ".md")
	attemptsDir := filepath.Join(filepath.Dir(todo.FilePath), base+".attempts")
	if err := os.MkdirAll(attemptsDir, 0755); err != nil {
		return "", fmt.Errorf("creating attempts dir: %w", err)
	}

	filename := fmt.Sprintf("attempt-%d.md", todo.Attempts)
	fullPath := filepath.Join(attemptsDir, filename)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Attempt %d\n\n", todo.Attempts))
	sb.WriteString(fmt.Sprintf("- **Status:** %s\n", result.statusString()))
	sb.WriteString(fmt.Sprintf("- **Date:** %s\n", time.Now().Format("2006-01-02 15:04")))
	sb.WriteString(fmt.Sprintf("- **Model:** %s\n", result.ExecutorName))
	sb.WriteString(fmt.Sprintf("- **Duration:** %s\n", result.Duration.Round(time.Second)))
	sb.WriteString(fmt.Sprintf("- **Cost:** $%.4f\n", result.CostUSD))
	sb.WriteString(fmt.Sprintf("- **Tokens:** %d\n", result.TokensUsed))
	if result.CommitSHA != "" {
		sb.WriteString(fmt.Sprintf("- **Commit:** `%s`\n", result.CommitSHA))
	}

	if result.Transcript != nil && len(result.Transcript.Entries) > 0 {
		sb.WriteString("\n## Transcript\n\n")
		for _, entry := range result.Transcript.Entries {
			sb.WriteString(fmt.Sprintf("**[%s] %s** (%s)\n", entry.Timestamp.Format("15:04:05"), entry.Type, entry.Role))
			if entry.Content != "" {
				sb.WriteString(entry.Content + "\n")
			}
			sb.WriteString("\n")
		}
	}

	if err := os.WriteFile(fullPath, []byte(sb.String()), 0644); err != nil {
		return "", fmt.Errorf("writing transcript file: %w", err)
	}

	// Return relative path from TODO file's directory
	return filepath.Join(base+".attempts", filename), nil
}

func appendAttemptRow(todo *types.TODO, attempt types.Attempt) error {
	content, err := os.ReadFile(todo.FilePath)
	if err != nil {
		return fmt.Errorf("reading TODO file: %w", err)
	}

	row := formatAttemptRow(todo.Attempts, attempt)
	updated := upsertAttemptsSection(string(content), row)

	tmpFile := todo.FilePath + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(updated), 0644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := os.Rename(tmpFile, todo.FilePath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}

func formatAttemptRow(n int, a types.Attempt) string {
	commit := ""
	if a.Commit != "" {
		commit = "`" + a.Commit + "`"
	}
	transcript := ""
	if a.Transcript != "" {
		transcript = fmt.Sprintf("[transcript](%s)", a.Transcript)
	}
	return fmt.Sprintf("| %d | %s | %s | %s | %s | $%.4f | %d | %s | %s |",
		n, a.Status, a.Timestamp.Format("2006-01-02 15:04"),
		a.Model, a.Duration.Round(time.Second),
		a.Cost, a.Tokens, commit, transcript)
}

const attemptsTableHeader = `## Attempts

| # | Status | Date | Model | Duration | Cost | Tokens | Commit | Transcript |
|---|--------|------|-------|----------|------|--------|--------|------------|`

func upsertAttemptsSection(content, newRow string) string {
	idx := strings.Index(content, "## Attempts")
	if idx < 0 {
		// Append new section
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		return content + "\n" + attemptsTableHeader + "\n" + newRow + "\n"
	}

	// Find the end of the table (next ## heading or EOF)
	rest := content[idx:]
	lines := strings.Split(rest, "\n")

	tableEnd := len(lines)
	for i := 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "## ") && trimmed != "## Attempts" {
			tableEnd = i
			break
		}
	}

	// Insert new row before the end of the table section
	insertAt := tableEnd
	for insertAt > 0 && strings.TrimSpace(lines[insertAt-1]) == "" {
		insertAt--
	}

	result := content[:idx]
	result += strings.Join(lines[:insertAt], "\n") + "\n"
	result += newRow + "\n"
	if tableEnd < len(lines) {
		result += "\n" + strings.Join(lines[tableEnd:], "\n")
	}
	return result
}

func (r *ExecutionResult) statusString() string {
	if r.Success {
		return "completed"
	}
	if r.Skipped {
		return "skipped"
	}
	return "failed"
}
