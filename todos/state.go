package todos

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/todos/types"
)

// StateUpdate represents a partial update to a TODO's frontmatter.
// Only non-nil fields will be updated in the TODO file.
type StateUpdate struct {
	SessionID *string
	Status    *types.Status
	Attempts  *int
	LastRun   *time.Time
}

// UpdateTODOState atomically updates the frontmatter of a TODO file with the provided updates.
// The file is updated using a temp file + rename pattern to ensure atomicity.
// The in-memory TODO object is also updated to reflect the changes.
// Only non-nil fields in StateUpdate will be applied to the TODO.
func UpdateTODOState(todo *types.TODO, updates StateUpdate) error {
	result, err := ParseFrontmatterFromFile(todo.FilePath)
	if err != nil {
		return err
	}

	frontmatter := result.Frontmatter

	// Apply updates
	if updates.Status != nil {
		frontmatter.Status = *updates.Status
	}
	if updates.Attempts != nil {
		frontmatter.Attempts = *updates.Attempts
	}
	if updates.LastRun != nil {
		frontmatter.LastRun = updates.LastRun
	}
	if updates.SessionID != nil {
		if frontmatter.LLM == nil {
			frontmatter.LLM = &types.LLM{}
		}
		frontmatter.LLM.SessionId = *updates.SessionID
	}

	newContent, err := WriteFrontmatter(&frontmatter, result.MarkdownContent)
	if err != nil {
		return err
	}

	// Write atomically (temp file + rename)
	tmpFile := todo.FilePath + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpFile, todo.FilePath); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	// Update in-memory TODO object
	todo.TODOFrontmatter = frontmatter

	return nil
}

// UpdateTODOStateFromFile is a convenience wrapper that parses the TODO file at filePath
// and then updates its state. Equivalent to ParseTODO followed by UpdateTODOState.
func UpdateTODOStateFromFile(filePath string, updates StateUpdate) error {
	todo, err := ParseTODO(filePath)
	if err != nil {
		return err
	}
	return UpdateTODOState(todo, updates)
}

// WriteTODOFile writes a new TODO file with the given content.
// The parent directory is created if it doesn't exist.
// The TODO is serialized as YAML frontmatter followed by markdown content.
func WriteTODOFile(filePath string, todo *types.TODO) error {

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	yamlData, err := todo.AsYaml()
	if err != nil {
		return fmt.Errorf("failed to marshal TODO to YAML: %w", err)
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(yamlData), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ReadTODOState reads only the frontmatter state from a TODO file without parsing
// the entire markdown structure. This is more efficient than ParseTODO when only
// the frontmatter metadata is needed.
func ReadTODOState(filePath string) (*types.TODOFrontmatter, error) {
	result, err := ParseFrontmatterFromFile(filePath)
	if err != nil {
		return nil, err
	}
	return &result.Frontmatter, nil
}

// UpdateLatestFailure updates the "## Latest Failure" section in a TODO file with test result info.
// If the section doesn't exist, it is created before "## Failure History" or at the end.
func UpdateLatestFailure(todo *types.TODO, result *types.TestResultInfo) error {
	if result == nil {
		return nil
	}

	content, err := os.ReadFile(todo.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read TODO file: %w", err)
	}

	contentStr := string(content)

	// Build the new Latest Failure section content
	newSection := formatLatestFailureSection(result)

	// Find and replace the "## Latest Failure" section
	updatedContent := replaceLatestFailureSection(contentStr, newSection)

	// Write atomically
	tmpFile := todo.FilePath + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpFile, todo.FilePath); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// formatLatestFailureSection formats the test result info as a markdown section using clicky.Format.
func formatLatestFailureSection(result *types.TestResultInfo) string {
	output, _ := clicky.Format(result, clicky.FormatOptions{Markdown: true})
	return output + "\n"
}

// replaceLatestFailureSection finds and replaces the "## Latest Failure" section in content.
// If the section doesn't exist, it inserts before "## Failure History" or at the end.
func replaceLatestFailureSection(content, newSection string) string {
	lines := strings.Split(content, "\n")

	// Find the start and end of "## Latest Failure" section
	sectionStart := -1
	sectionEnd := -1
	failureHistoryIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## Latest Failure") {
			sectionStart = i
		} else if sectionStart >= 0 && sectionEnd < 0 && strings.HasPrefix(trimmed, "## ") {
			// Found start of next section
			sectionEnd = i
		} else if strings.HasPrefix(trimmed, "## Failure History") {
			failureHistoryIdx = i
			if sectionStart >= 0 && sectionEnd < 0 {
				sectionEnd = i
			}
		}
	}

	// If we found the section but not the end, it goes to end of file
	if sectionStart >= 0 && sectionEnd < 0 {
		sectionEnd = len(lines)
	}

	var result strings.Builder

	if sectionStart >= 0 {
		// Replace existing section
		for i := 0; i < sectionStart; i++ {
			result.WriteString(lines[i])
			result.WriteString("\n")
		}
		result.WriteString(newSection)
		for i := sectionEnd; i < len(lines); i++ {
			result.WriteString(lines[i])
			if i < len(lines)-1 {
				result.WriteString("\n")
			}
		}
	} else if failureHistoryIdx >= 0 {
		// Insert before "## Failure History"
		for i := 0; i < failureHistoryIdx; i++ {
			result.WriteString(lines[i])
			result.WriteString("\n")
		}
		result.WriteString(newSection)
		result.WriteString("\n")
		for i := failureHistoryIdx; i < len(lines); i++ {
			result.WriteString(lines[i])
			if i < len(lines)-1 {
				result.WriteString("\n")
			}
		}
	} else {
		// Append at end
		result.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			result.WriteString("\n")
		}
		result.WriteString("\n")
		result.WriteString(newSection)
	}

	return result.String()
}
