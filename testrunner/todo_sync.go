package testrunner

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/goccy/go-yaml"
)

// TodoSync manages syncing test failures to TODO files.
type TodoSync struct {
	todosDir     string
	templatePath string
}

// NewTodoSync creates a new TODO sync handler.
func NewTodoSync(todosDir, templatePath string) *TodoSync {
	return &TodoSync{
		todosDir:     todosDir,
		templatePath: templatePath,
	}
}

// SyncFailure creates or updates a TODO for a test failure.
func (ts *TodoSync) SyncFailure(failure TestFailure) (string, error) {
	if err := os.MkdirAll(ts.todosDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create todos directory: %w", err)
	}

	todoPath, err := ts.findExistingTodo(failure)
	if err != nil {
		return "", err
	}

	if todoPath != "" {
		return todoPath, ts.updateTodo(todoPath, failure)
	}

	return ts.createTodo(failure)
}

func (ts *TodoSync) findExistingTodo(failure TestFailure) (string, error) {
	entries, err := os.ReadDir(ts.todosDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	slug := ts.generateTodoSlug(failure)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), slug) {
			return filepath.Join(ts.todosDir, entry.Name()), nil
		}
	}

	return "", nil
}

func (ts *TodoSync) createTodo(failure TestFailure) (string, error) {
	slug := ts.generateTodoSlug(failure)

	// Find next number
	entries, _ := os.ReadDir(ts.todosDir)
	maxNum := 0
	pattern := regexp.MustCompile(regexp.QuoteMeta(slug) + `-(\d+)\.md$`)

	for _, entry := range entries {
		matches := pattern.FindStringSubmatch(entry.Name())
		if len(matches) > 1 {
			if num, err := strconv.Atoi(matches[1]); err == nil && num > maxNum {
				maxNum = num
			}
		}
	}

	todoPath := filepath.Join(ts.todosDir, fmt.Sprintf("%s-%03d.md", slug, maxNum+1))

	templateContent := ""
	if ts.templatePath != "" {
		if content, err := os.ReadFile(ts.templatePath); err == nil {
			templateContent = string(content)
		}
	}

	content := ts.generateTodoContent(failure, templateContent)
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write todo file: %w", err)
	}

	return todoPath, nil
}

func (ts *TodoSync) updateTodo(todoPath string, failure TestFailure) error {
	result, err := todos.ParseFrontmatterFromFile(todoPath)
	if err != nil {
		return err
	}

	frontmatter := result.Frontmatter
	frontmatter.Attempts++
	now := time.Now()
	frontmatter.LastRun = &now

	// Add to failure history
	historyEntry := fmt.Sprintf("\n### Attempt %d - %s\n%s\n", frontmatter.Attempts, now.Format(time.RFC3339), failure.Message)
	markdownContent := result.MarkdownContent + historyEntry

	updatedContent, err := todos.WriteFrontmatter(&frontmatter, markdownContent)
	if err != nil {
		return err
	}

	return os.WriteFile(todoPath, []byte(updatedContent), 0644)
}

func (ts *TodoSync) generateTodoSlug(failure TestFailure) string {
	slug := strings.ToLower(failure.Name)
	slug = regexp.MustCompile(`[:,./]`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

func (ts *TodoSync) generateTodoContent(failure TestFailure, template string) string {
	now := time.Now()
	frontmatter := types.TODOFrontmatter{
		Priority: types.PriorityHigh,
		Status:   types.StatusPending,
		Language: types.LanguageGo,
		Attempts: 1,
		LastRun:  &now,
	}

	frontmatterBytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		frontmatterBytes = []byte(fmt.Sprintf(`priority: high
status: pending
language: go
attempts: 1
last_run: %s
`, now.Format(time.RFC3339)))
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(frontmatterBytes)
	sb.WriteString("---\n\n")

	// Use PrettyTODO() for the markdown body
	body, _ := clicky.Format(failure.PrettyTODO(), clicky.FormatOptions{Markdown: true})
	sb.WriteString(body)

	if template != "" {
		sb.WriteString("\n## Fix Instructions\n\n")
		sb.WriteString(template)
		sb.WriteString("\n")
	}

	return sb.String()
}
