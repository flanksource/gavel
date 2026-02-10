// Package claudehistory provides functionality for parsing and analyzing
// Claude Code session history from JSONL files stored in ~/.claude/projects/.
package claudehistory

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/formatters"
	"github.com/flanksource/commons/collections"
	"github.com/flanksource/commons/logger"
)

// NormalizePath converts a filesystem path into a normalized format
// by replacing "/" and "." with "-" for use as a directory name.
func NormalizePath(path string) string {
	normalized := strings.ReplaceAll(path, "/", "-")
	normalized = strings.ReplaceAll(normalized, ".", "-")
	return normalized
}

// FindSessionFiles discovers Claude Code session JSONL files in the projects directory.
// If searchAll is false, it only searches for sessions matching the currentDir path.
// Returns a list of absolute paths to session files.
func FindSessionFiles(projectsDir, currentDir string, searchAll bool) ([]string, error) {
	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		logger.Debugf("Projects directory does not exist: %s", projectsDir)
		return nil, nil
	}

	var sessionFiles []string

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	logger.Debugf("Found %d project directories in %s", len(entries), projectsDir)

	var normalized string
	if !searchAll && currentDir != "" {
		normalized = NormalizePath(currentDir)
		logger.Debugf("Looking for directories matching: %s", normalized)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectPath := filepath.Join(projectsDir, entry.Name())

		if !searchAll && currentDir != "" {
			if !strings.HasSuffix(entry.Name(), normalized) {
				logger.Debugf("Skipping directory %s (doesn't match suffix %s)", entry.Name(), normalized)
				continue
			}
			logger.Debugf("Matched directory: %s", entry.Name())
		}

		matches, err := filepath.Glob(filepath.Join(projectPath, "*.jsonl"))
		if err != nil {
			logger.Warnf("Error globbing session files in %s: %v", projectPath, err)
			continue
		}

		logger.Debugf("Found %d session files in %s", len(matches), projectPath)
		sessionFiles = append(sessionFiles, matches...)
	}

	logger.Debugf("Total session files found: %d", len(sessionFiles))
	return sessionFiles, nil
}

// ExtractToolUses parses a Claude Code session JSONL file and extracts all tool use entries.
// Each line in the JSONL file represents a SessionEntry containing messages with tool uses.
// Returns a list of ToolUse objects with metadata like timestamp, CWD, and session ID.
func ExtractToolUses(sessionFile string) ([]ToolUse, error) {
	file, err := os.Open(sessionFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var toolUses []ToolUse
	scanner := bufio.NewScanner(file)

	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var entry SessionEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			logger.Debugf("Error parsing line in %s: %v", sessionFile, err)
			continue
		}

		for _, content := range entry.Message.Content {
			if content.Type != "tool_use" {
				os.Stderr.WriteString(clicky.MustFormat(content, formatters.FormatOptions{Pretty: true}) + "\n")
				continue
			}

			var timestamp *time.Time
			if entry.Timestamp != "" {
				t, err := time.Parse(time.RFC3339, entry.Timestamp)
				if err == nil {
					timestamp = &t
				}
			}

			toolUse := ToolUse{
				Tool:      content.Name,
				Input:     content.Input,
				Timestamp: timestamp,
				CWD:       entry.CWD,
				SessionID: entry.SessionID,
				ToolUseID: content.ID,
			}
			toolUses = append(toolUses, toolUse)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return toolUses, nil
}

// FormatCommand extracts a human-readable command string from a ToolUse.
// For Bash tools, it returns the command. For file operations, it returns the file path.
// For search tools, it returns the pattern and path. Falls back to JSON representation.
func FormatCommand(toolUse ToolUse) string {
	switch toolUse.Tool {
	case "Bash":
		if cmd, ok := toolUse.Input["command"].(string); ok {
			return cmd
		}
	case "Read", "Write", "Edit":
		if path, ok := toolUse.Input["file_path"].(string); ok {
			return path
		}
	case "Grep":
		pattern, _ := toolUse.Input["pattern"].(string)
		path, _ := toolUse.Input["path"].(string)
		return strings.TrimSpace(fmt.Sprintf("%s %s", pattern, path))
	case "Glob":
		if pattern, ok := toolUse.Input["pattern"].(string); ok {
			return pattern
		}
	case "WebFetch":
		if url, ok := toolUse.Input["url"].(string); ok {
			return url
		}
	}

	b, _ := json.Marshal(toolUse.Input)
	return string(b)
}

// FilterToolUses applies filter criteria to a list of tool uses and returns the filtered results.
// Supports filtering by tool name (with comma-separated patterns and negation), time range (Since/Before),
// and result limiting. Results are sorted by timestamp in descending order (newest first).
func FilterToolUses(toolUses []ToolUse, filter Filter) []ToolUse {
	var filtered []ToolUse

	for _, tu := range toolUses {
		// Use collections.MatchAny for tool filter to support patterns like "Bash,Read" or "!Write"
		if filter.Tool != "" {
			// Split comma-separated patterns and pass as variadic args
			patterns := strings.Split(filter.Tool, ",")
			for i, p := range patterns {
				patterns[i] = strings.TrimSpace(p)
			}

			matched, negated := collections.MatchAny([]string{tu.Tool}, patterns...)
			if negated {
				continue
			}
			if !matched {
				continue
			}
		}

		if filter.Since != nil && tu.Timestamp != nil && tu.Timestamp.Before(*filter.Since) {
			continue
		}

		if filter.Before != nil && tu.Timestamp != nil && tu.Timestamp.After(*filter.Before) {
			continue
		}

		filtered = append(filtered, tu)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Timestamp == nil {
			return false
		}
		if filtered[j].Timestamp == nil {
			return true
		}
		return filtered[i].Timestamp.After(*filtered[j].Timestamp)
	})

	if filter.Limit > 0 && len(filtered) > filter.Limit {
		filtered = filtered[:filter.Limit]
	}

	return filtered
}

// GetClaudeHome returns the absolute path to the Claude Code home directory (~/.claude).
func GetClaudeHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude")
}

// GetProjectsDir returns the absolute path to the Claude Code projects directory (~/.claude/projects).
func GetProjectsDir() string {
	return filepath.Join(GetClaudeHome(), "projects")
}

// ParseResult contains the results of parsing Claude Code session history.
type ParseResult struct {
	ToolUses        []ToolUse
	SessionsFound   int
	SessionsScanned int
}

// ParseHistory is the main entry point for parsing Claude Code session history.
// It discovers session files, extracts tool uses, applies filters, and returns aggregated results.
// If searchAll is false, only sessions matching the currentDir are parsed.
func ParseHistory(currentDir string, searchAll bool, filter Filter) (*ParseResult, error) {
	projectsDir := GetProjectsDir()

	sessionFiles, err := FindSessionFiles(projectsDir, currentDir, searchAll)
	if err != nil {
		return nil, err
	}

	result := &ParseResult{
		SessionsFound: len(sessionFiles),
	}

	if len(sessionFiles) == 0 {
		return result, nil
	}

	var allToolUses []ToolUse
	for _, sessionFile := range sessionFiles {
		toolUses, err := ExtractToolUses(sessionFile)
		if err != nil {
			logger.Warnf("Error extracting tool uses from %s: %v", sessionFile, err)
			continue
		}
		if len(toolUses) > 0 {
			result.SessionsScanned++
			allToolUses = append(allToolUses, toolUses...)
		}
	}

	result.ToolUses = FilterToolUses(allToolUses, filter)
	return result, nil
}
