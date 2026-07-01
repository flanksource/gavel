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
	Priority  *types.Priority
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
	if updates.Priority != nil {
		frontmatter.Priority = *updates.Priority
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

// EditTODOContent updates a file-backed TODO's title (frontmatter) and/or body
// (markdown content), preserving everything not being edited. Only the non-nil
// fields in EditRequest are applied. The in-memory TODO is updated to match.
func EditTODOContent(todo *types.TODO, edit EditRequest) error {
	if edit.IsEmpty() {
		return fmt.Errorf("nothing to edit: title or body is required")
	}
	result, err := ParseFrontmatterFromFile(todo.FilePath)
	if err != nil {
		return err
	}
	frontmatter := result.Frontmatter
	markdown := result.MarkdownContent

	if edit.Title != nil {
		title := strings.TrimSpace(*edit.Title)
		if title == "" {
			return fmt.Errorf("title cannot be empty")
		}
		frontmatter.Title = title
	}
	if edit.Body != nil {
		markdown = normalizeMarkdownBody(*edit.Body)
	}

	newContent, err := WriteFrontmatter(&frontmatter, markdown)
	if err != nil {
		return err
	}
	if err := atomicWriteFile(todo.FilePath, newContent); err != nil {
		return err
	}

	todo.TODOFrontmatter = frontmatter
	todo.MarkdownBody = markdown
	return nil
}

// AppendComment records a comment in a file-backed TODO's "## Comments" section,
// creating the section when absent. File TODOs have no event log, so comments
// live inline in the markdown the way attempts and failures already do.
func AppendComment(todo *types.TODO, body string) error {
	body = strings.TrimSpace(body)
	if body == "" {
		return fmt.Errorf("comment body is required")
	}
	content, err := os.ReadFile(todo.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read TODO file: %w", err)
	}
	entry := fmt.Sprintf("### %s\n\n%s", time.Now().Format("2006-01-02 15:04"), body)
	updated := appendCommentSection(string(content), entry)
	if err := atomicWriteFile(todo.FilePath, updated); err != nil {
		return err
	}
	if parsed, perr := ParseFrontmatter(updated); perr == nil {
		todo.MarkdownBody = parsed.MarkdownContent
	}
	return nil
}

// normalizeMarkdownBody ensures the edited body starts on its own line so it
// sits cleanly after the frontmatter's closing delimiter.
func normalizeMarkdownBody(body string) string {
	body = strings.TrimRight(body, "\n")
	if body == "" {
		return ""
	}
	return "\n" + strings.TrimLeft(body, "\n") + "\n"
}

const commentsSectionHeader = "## Comments"

// appendCommentSection inserts entry at the end of the "## Comments" section,
// creating that section at the end of the file when it does not yet exist.
// Mirrors upsertAttemptsSection so comments group under a single heading.
func appendCommentSection(content, entry string) string {
	idx := strings.Index(content, commentsSectionHeader)
	if idx < 0 {
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		return content + "\n" + commentsSectionHeader + "\n\n" + entry + "\n"
	}

	rest := content[idx:]
	lines := strings.Split(rest, "\n")
	sectionEnd := len(lines)
	for i := 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "## ") {
			sectionEnd = i
			break
		}
	}
	insertAt := sectionEnd
	for insertAt > 0 && strings.TrimSpace(lines[insertAt-1]) == "" {
		insertAt--
	}

	result := content[:idx]
	result += strings.Join(lines[:insertAt], "\n") + "\n\n"
	result += entry + "\n"
	if sectionEnd < len(lines) {
		result += "\n" + strings.Join(lines[sectionEnd:], "\n")
	}
	return result
}

// atomicWriteFile writes content to path via a temp file + rename so a reader
// never observes a half-written TODO file.
func atomicWriteFile(path, content string) error {
	tmpFile := path + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := os.Rename(tmpFile, path); err != nil {
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
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
