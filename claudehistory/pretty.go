package claudehistory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/commons/logger"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// ToolUseResult represents a formatted tool use for display with an index for ordering.
type ToolUseResult struct {
	ToolUse ToolUse
	Index   int
}

// Pretty formats a ToolUse for display using the Clicky Text API.
// It provides specialized formatting for different tool types:
//   - Bash: shows command with syntax highlighting
//   - Edit: shows unified diff with context lines
//   - Write: shows file preview with language detection
//   - Read: shows file path with offset/limit range
//   - Grep/Glob: shows pattern and path
//   - WebFetch: shows URL
//
// All tool uses include optional metadata like timestamp and working directory.
func (t ToolUse) Pretty() api.Text {
	cwd, _ := os.Getwd()

	text := clicky.Text("").
		Append(t.Tool, "text-blue-300 wrap-space")

	// Add metadata on next line if available
	if t.Timestamp != nil || t.CWD != "" {
		text = text.NewLine()
	}

	if t.Timestamp != nil {
		if time.Since(*t.Timestamp).Hours() < 24 {
			text = text.Append(t.Timestamp.Format("15:04:05")+"  ", "text-gray-500")
		} else {
			text = text.Append(t.Timestamp.Format("2006-01-02")+"  ", "text-gray-500")
		}
	}

	if logger.IsDebugEnabled() && t.CWD != "" {
		text = text.Add(icons.Folder).
			Append(" "+getRelativePath(t.CWD, cwd), "text-gray-400 text-xs")
	}

	filepath, _ := t.Input["file_path"].(string)
	data := t.Input

	if desc, ok := data["description"].(string); ok && desc != "" {
		text = text.Append(": ", "text-gray-400").Append(desc, "text-gray-700")
		delete(data, "description")
	}

	if timeout, ok := data["timeout"].(float64); ok && timeout > 0 {
		data["timeout"] = time.Duration(timeout) * time.Millisecond
	}

	if filepath != "" {
		delete(data, "file_path")
		filepath = getRelativePath(filepath, cwd)
	}

	switch t.Tool {
	case "Bash":
		text = text.Add(clicky.CodeBlock(t.Input["command"].(string), "bash"))
		delete(data, "command")
	case "Edit":

		oldStr, _ := t.Input["old_string"].(string)
		newStr, _ := t.Input["new_string"].(string)
		text = text.Add(formatEditDiff(filepath, oldStr, newStr, cwd))
		delete(data, "old_string")
		delete(data, "new_string")

	case "Write":
		content, _ := t.Input["content"].(string)
		delete(data, "content")
		text = text.Add(formatWritePreview(filepath, content, cwd))

	case "Read":
		limit, _ := t.Input["limit"].(float64)
		offset, _ := t.Input["offset"].(float64)
		delete(data, "limit")
		delete(data, "offset")
		text = text.Append(filepath)
		if offset > 0 || limit > 0 {
			text = text.Append(fmt.Sprintf("[%d:%d]", int(offset), int(limit)), "text-gray-500")
		}

	case "Grep":
		pattern, _ := t.Input["pattern"].(string)
		path, _ := t.Input["path"].(string)
		delete(data, "pattern")
		delete(data, "path")
		text = text.Append(strings.TrimSpace(fmt.Sprintf("%s %s", pattern, path)))
	case "Glob":
		if pattern, ok := t.Input["pattern"].(string); ok {
			text = text.Append(pattern)
		}
		delete(data, "path")
		delete(data, "pattern")

	case "WebFetch":
		if url, ok := t.Input["url"].(string); ok {
			text = text.Append(url)
		}
		delete(data, "url")

	case "Task":
		desc, _ := t.Input["description"].(string)
		if desc == "" {
			if prompt, ok := t.Input["prompt"].(string); ok && len(prompt) > 80 {
				desc = prompt[:80] + "..."
			} else {
				desc, _ = t.Input["prompt"].(string)
			}
		}
		text = text.Add(api.Text{}.Add(icons.ArrowRight)).
			Append(" Task", "text-blue-600")
		if desc != "" {
			text = text.Append(": ", "text-gray-400").Append(desc, "text-gray-700")
		}
		data = nil

	case "TodoWrite":
		text = text.Add(api.Text{}.Add(icons.ArrowRight)).
			Append(" TodoWrite", "text-blue-600")
		data = nil

	default:
		text = text.Add(api.Text{}.Add(icons.ArrowRight)).
			Append(fmt.Sprintf(" %s", t.Tool), "text-blue-600")

	}

	if len(data) > 0 {
		text = text.Add(clicky.Map(data, "max-w-[100ch]"))
	}

	return text
}

// ToolUseSummary provides a summary of tool uses
type ToolUseSummary struct {
	TotalCount   int
	ToolFilter   string
	LimitApplied int
}

func (s ToolUseSummary) Pretty() api.Text {
	text := clicky.Text("•").
		Append(fmt.Sprintf(" Found %d commands", s.TotalCount), "font-bold text-blue-600")

	if s.ToolFilter != "" {
		text = text.Append(fmt.Sprintf(" (filtered by %s)", s.ToolFilter), "text-gray-500")
	}

	if s.LimitApplied > 0 && s.TotalCount > s.LimitApplied {
		text = text.Append(fmt.Sprintf("\n  Showing first %d results", s.LimitApplied), "text-yellow-600")
	}

	return text
}

// NoResultsError represents a diagnostic error when no tools are found
type NoResultsError struct {
	Filter          Filter
	CurrentDir      string
	SearchedAll     bool
	SessionsFound   int
	SessionsScanned int
}

func (e NoResultsError) Pretty() api.Text {
	text := clicky.Text("").
		Add(icons.Error).
		AddText(" No commands found matching criteria", "font-bold text-red-600").
		NewLine().NewLine().
		AddText("Diagnostics:", "font-bold text-yellow-600").
		NewLine()

	// Show what was searched
	if e.SearchedAll {
		text = text.AddText("  • Searched all sessions across all directories", "text-gray-600")
	} else {
		text = text.AddText(fmt.Sprintf("  • Searched current directory: %s", e.CurrentDir), "text-gray-600")
	}

	text = text.NewLine().
		AddText(fmt.Sprintf("  • Sessions found: %d", e.SessionsFound), "text-gray-600").
		NewLine().
		AddText(fmt.Sprintf("  • Sessions scanned: %d", e.SessionsScanned), "text-gray-600")

	// Show filters applied
	if e.Filter.Tool != "" {
		text = text.NewLine().
			AddText(fmt.Sprintf("  • Tool filter: %s", e.Filter.Tool), "text-gray-600")
	}

	if e.Filter.Limit > 0 {
		text = text.NewLine().
			AddText(fmt.Sprintf("  • Limit: %d", e.Filter.Limit), "text-gray-600")
	}

	// Suggestions
	text = text.NewLine().NewLine().
		AddText("Suggestions:", "font-bold text-cyan-600").
		NewLine().
		AddText("  • Try removing filters (e.g., --tool)", "text-cyan-500").
		NewLine().
		AddText("  • Use --all to search all sessions", "text-cyan-500").
		NewLine().
		AddText("  • Increase --limit value", "text-cyan-500").
		NewLine().
		AddText("  • Check if Claude Code has been used recently", "text-cyan-500")

	return text
}

func (e NoResultsError) String() string {
	return e.Pretty().String()
}

func (e NoResultsError) ANSI() string {
	return e.Pretty().ANSI()
}

func (e NoResultsError) HTML() string {
	return e.Pretty().HTML()
}

func (e NoResultsError) Markdown() string {
	return e.Pretty().Markdown()
}

func (e NoResultsError) Error() string {
	if e.Filter.Tool != "" {
		return fmt.Sprintf("no %s commands found in history", e.Filter.Tool)
	}
	return "no commands found in history"
}

func getRelativePath(filePath string, workDir string) string {
	if rel, err := filepath.Rel(workDir, filePath); err == nil {
		return rel
	}
	return filePath
}

// createDiffPatch generates a unified diff patch between old and new strings.
func createDiffPatch(oldStr, newStr string) api.Text {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldStr, newStr, false)

	var result = clicky.Text("")

	// Create unified diff format
	for _, diff := range diffs {
		text := diff.Text
		switch diff.Type {
		case diffmatchpatch.DiffDelete:
			// Add red/removed lines with - prefix
			for _, line := range strings.Split(text, "\n") {
				if line != "" || diff.Text == "\n" {
					result = result.Append("-", "text-red-700").Append(line, "text-red-500").NewLine()
				}
			}
		case diffmatchpatch.DiffInsert:
			// Add green/added lines with + prefix
			for _, line := range strings.Split(text, "\n") {
				if line != "" || diff.Text == "\n" {
					result = result.Append("+", "text-green-700").Append(line, "text-green-500").NewLine()

				}
			}
		case diffmatchpatch.DiffEqual:
			// Add context lines with space prefix (show up to 3 lines of context)
			lines := strings.Split(text, "\n")
			contextLines := 3

			// Show last N lines before change
			startIdx := len(lines) - contextLines
			if startIdx < 0 {
				startIdx = 0
			}

			for i := startIdx; i < len(lines); i++ {
				if lines[i] != "" || i < len(lines)-1 {
					result = result.Append(lines[i], "text-gray-300").NewLine()
				}
			}
		}
	}

	return result
}

// formatEditDiff creates a visually formatted diff for Edit tool operations.
func formatEditDiff(filePath string, oldStr, newStr string, workDir string) api.Text {
	relPath := getRelativePath(filePath, workDir)

	result := api.Text{}.Add(icons.File).
		Append(" Editing ", "text-blue-600 font-bold").
		Append(relPath, "text-cyan-600 font-medium").
		NewLine().Add(createDiffPatch(oldStr, newStr))

	return result
}

// formatWritePreview creates a formatted preview for Write tool operations.
func formatWritePreview(filePath string, content string, workDir string) api.Text {
	relPath := getRelativePath(filePath, workDir)

	result := api.Text{}.Add(icons.File).
		Append(" Writing ", "text-blue-600 font-bold").
		Append(relPath, "text-cyan-600 font-medium").
		NewLine()

	// Truncate long content
	preview := content
	maxLines := 20
	lines := strings.Split(content, "\n")
	if len(lines) > maxLines {
		preview = strings.Join(lines[:maxLines], "\n") + "\n... (truncated)"
	}

	// Code with language detection - chroma will syntax highlight
	code := api.NewCode(preview, detectLanguage(relPath))

	return result.Add(code)
}

// detectLanguage detects the programming language from file extension.
func detectLanguage(filePath string) string {
	ext := filepath.Ext(filePath)
	langMap := map[string]string{
		".go":   "go",
		".py":   "python",
		".js":   "javascript",
		".ts":   "typescript",
		".tsx":  "typescript",
		".jsx":  "javascript",
		".md":   "markdown",
		".yaml": "yaml",
		".yml":  "yaml",
		".json": "json",
		".sh":   "bash",
		".bash": "bash",
		".sql":  "sql",
		".html": "html",
		".css":  "css",
	}
	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return ""
}
