// Package portable implements explicit interchange between native PostgreSQL
// TODO issues and repository-local .todos Markdown files. It is deliberately
// not a runtime Provider: callers must invoke Import or Export explicitly.
package portable

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	DefaultDirectory = ".todos"
	metadataID       = "id"
	metadataLabels   = "labels"
)

var (
	portableNamespace = uuid.MustParse("797e56c7-47f4-51dc-93b7-d7ce9904af21")
	filenameInvalid   = regexp.MustCompile(`[^a-z0-9]+`)
)

// ImportResult reports the database changes made by one explicit file import.
type ImportResult struct {
	Directory string   `json:"directory"`
	Created   int      `json:"created"`
	Updated   int      `json:"updated"`
	Unchanged int      `json:"unchanged"`
	Files     []string `json:"files"`
}

// ExportResult reports the files written by one explicit database export.
type ExportResult struct {
	Directory string   `json:"directory"`
	Exported  int      `json:"exported"`
	Files     []string `json:"files"`
}

type importIssue struct {
	ID           uuid.UUID
	Title        string
	Body         string
	Verification string
	Labels       []string
	Priority     native.Priority
	Status       native.IssueStatus
	File         string
}

type exportFile struct {
	path    string
	content string
}

// Import reads .todos Markdown only from the explicitly supplied files or
// directory and applies the supported portable fields to the native workspace
// for workDir. It never selects a runtime file provider or falls back from DB.
func Import(ctx context.Context, db *gorm.DB, workDir, directory string, files []string) (*ImportResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if db == nil {
		return nil, fmt.Errorf("portable TODO import requires PostgreSQL: %w", native.ErrInvalidInput)
	}
	workDir, err := absolutePath("working directory", workDir, "")
	if err != nil {
		return nil, err
	}
	directory, err = portableDirectory("import directory", workDir, directory)
	if err != nil {
		return nil, err
	}

	provider, err := todoruntime.New(ctx, workDir, db)
	if err != nil {
		return nil, err
	}
	workspace := provider.Workspace()
	if workspace == nil {
		return nil, errors.New("portable TODO import resolved no native workspace")
	}

	parsed, err := loadImportIssues(workDir, directory, files, workspace.ID)
	if err != nil {
		return nil, err
	}
	result := &ImportResult{Directory: directory, Files: make([]string, 0, len(parsed))}
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repository, err := native.NewRepository(tx)
		if err != nil {
			return err
		}
		for _, input := range parsed {
			result.Files = append(result.Files, input.File)
			current, err := repository.GetIssue(ctx, input.ID)
			switch {
			case errors.Is(err, native.ErrNotFound):
				if _, err := repository.CreateIssue(ctx, native.CreateIssueInput{
					ID:             input.ID,
					WorkspaceID:    workspace.ID,
					Title:          input.Title,
					Body:           input.Body,
					Verification:   input.Verification,
					Labels:         input.Labels,
					Priority:       input.Priority,
					Status:         input.Status,
					ExecutionState: native.ExecutionIdle,
					Actor:          "todos-file-import",
				}); err != nil {
					return fmt.Errorf("import %s: %w", input.File, err)
				}
				result.Created++
			case err != nil:
				return fmt.Errorf("resolve imported TODO %s: %w", input.File, err)
			case current.WorkspaceID != workspace.ID:
				return fmt.Errorf("import %s: native TODO %s belongs to workspace %s, not %s", input.File, input.ID, current.WorkspaceID, workspace.ID)
			default:
				if portableIssueMatches(current, input) {
					result.Unchanged++
					continue
				}
				labels := append([]string(nil), input.Labels...)
				if _, err := repository.UpdateIssue(ctx, input.ID, current.Version, native.IssuePatch{
					Title:        &input.Title,
					Body:         &input.Body,
					Verification: &input.Verification,
					Labels:       &labels,
					Priority:     &input.Priority,
					Status:       &input.Status,
					Actor:        "todos-file-import",
				}); err != nil {
					return fmt.Errorf("update imported TODO %s: %w", input.File, err)
				}
				result.Updated++
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Export writes supported durable issue fields from the native workspace for
// workDir into an explicitly selected .todos directory. Existing files for the
// same native ID are updated; unrelated collisions require force.
func Export(ctx context.Context, db *gorm.DB, workDir, directory string, refs []string, force bool) (*ExportResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if db == nil {
		return nil, fmt.Errorf("portable TODO export requires PostgreSQL: %w", native.ErrInvalidInput)
	}
	workDir, err := absolutePath("working directory", workDir, "")
	if err != nil {
		return nil, err
	}
	directory, err = portableDirectory("export directory", workDir, directory)
	if err != nil {
		return nil, err
	}
	provider, err := todoruntime.New(ctx, workDir, db)
	if err != nil {
		return nil, err
	}
	workspace := provider.Workspace()
	if workspace == nil {
		return nil, errors.New("portable TODO export resolved no native workspace")
	}
	repository := provider.Repository()
	issues, err := exportIssues(ctx, repository, workspace.ID, refs)
	if err != nil {
		return nil, err
	}
	prepared := make([]exportFile, 0, len(issues))
	for _, issue := range issues {
		content, err := exportMarkdown(issue)
		if err != nil {
			return nil, err
		}
		path := filepath.Join(directory, portableFilename(issue))
		if err := refuseUnrelatedCollision(path, issue.ID, force); err != nil {
			return nil, err
		}
		prepared = append(prepared, exportFile{path: path, content: content})
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, fmt.Errorf("create portable TODO export directory: %w", err)
	}

	result := &ExportResult{Directory: directory, Files: make([]string, 0, len(prepared))}
	for _, file := range prepared {
		if err := atomicWrite(file.path, file.content); err != nil {
			return nil, err
		}
		result.Files = append(result.Files, file.path)
		result.Exported++
	}
	return result, nil
}

func loadImportIssues(workDir, directory string, files []string, workspaceID uuid.UUID) ([]importIssue, error) {
	var parsed types.TODOS
	if len(files) == 0 {
		var err error
		parsed, err = todos.DiscoverTODOs(directory, todos.DiscoveryFilters{})
		if err != nil {
			return nil, err
		}
	} else {
		parsed = make(types.TODOS, 0, len(files))
		for _, file := range files {
			path := strings.TrimSpace(file)
			if path == "" {
				return nil, errors.New("portable TODO import file is empty")
			}
			if !filepath.IsAbs(path) {
				path = filepath.Join(workDir, path)
			}
			todo, err := todos.ParseTODO(path)
			if err != nil {
				return nil, fmt.Errorf("parse portable TODO %s: %w", path, err)
			}
			parsed = append(parsed, todo)
		}
	}
	if len(parsed) == 0 {
		return nil, fmt.Errorf("no portable TODO Markdown files found in %s", directory)
	}

	result := make([]importIssue, 0, len(parsed))
	seen := map[uuid.UUID]string{}
	for _, todo := range parsed {
		input, err := importIssueFromTODO(todo, directory, workspaceID)
		if err != nil {
			return nil, err
		}
		if previous := seen[input.ID]; previous != "" {
			return nil, fmt.Errorf("portable TODO files %s and %s use the same id %s", previous, input.File, input.ID)
		}
		seen[input.ID] = input.File
		result = append(result, input)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].File < result[j].File })
	return result, nil
}

func importIssueFromTODO(todo *types.TODO, root string, workspaceID uuid.UUID) (importIssue, error) {
	if todo == nil {
		return importIssue{}, errors.New("portable TODO is nil")
	}
	title := strings.TrimSpace(todo.Title)
	if title == "" {
		return importIssue{}, fmt.Errorf("portable TODO %s has no title", todo.FilePath)
	}
	priority, err := importPriority(todo.Priority)
	if err != nil {
		return importIssue{}, fmt.Errorf("portable TODO %s: %w", todo.FilePath, err)
	}
	status, err := importStatus(todo.Status)
	if err != nil {
		return importIssue{}, fmt.Errorf("portable TODO %s: %w", todo.FilePath, err)
	}
	id, err := importedID(todo, root, workspaceID)
	if err != nil {
		return importIssue{}, err
	}
	body := strings.TrimSpace(todo.MarkdownBody)
	return importIssue{
		ID: id, Title: title, Body: body,
		Verification: todos.ExtractVerificationFixture(body),
		Labels:       importedLabels(todo.Metadata),
		Priority:     priority,
		Status:       status,
		File:         todo.FilePath,
	}, nil
}

func importedID(todo *types.TODO, root string, workspaceID uuid.UUID) (uuid.UUID, error) {
	if value := strings.TrimSpace(metadataString(todo.Metadata, metadataID)); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			return uuid.Nil, fmt.Errorf("portable TODO %s has invalid id %q: %w", todo.FilePath, value, err)
		}
		return id, nil
	}
	relative, err := filepath.Rel(root, todo.FilePath)
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve portable TODO path %s: %w", todo.FilePath, err)
	}
	identity := workspaceID.String() + "\x00" + filepath.ToSlash(relative)
	return uuid.NewSHA1(portableNamespace, []byte(identity)), nil
}

func exportIssues(ctx context.Context, repository *native.Repository, workspaceID uuid.UUID, refs []string) ([]native.Issue, error) {
	if len(refs) == 0 {
		issues, err := repository.ListIssues(ctx, workspaceID)
		if err != nil {
			return nil, err
		}
		sort.Slice(issues, func(i, j int) bool { return issues[i].ID.String() < issues[j].ID.String() })
		return issues, nil
	}
	issues := make([]native.Issue, 0, len(refs))
	seen := map[uuid.UUID]bool{}
	for _, ref := range refs {
		issue, err := repository.GetIssueByRef(ctx, workspaceID, ref)
		if err != nil {
			return nil, fmt.Errorf("resolve native TODO %q for export: %w", ref, err)
		}
		if !seen[issue.ID] {
			issues = append(issues, *issue)
			seen[issue.ID] = true
		}
	}
	return issues, nil
}

func exportMarkdown(issue native.Issue) (string, error) {
	priority, err := exportPriority(issue.Priority)
	if err != nil {
		return "", fmt.Errorf("export native TODO %s: %w", issue.ID, err)
	}
	status, err := exportStatus(issue.Status)
	if err != nil {
		return "", fmt.Errorf("export native TODO %s: %w", issue.ID, err)
	}
	body := strings.TrimSpace(issue.Body)
	if verification := strings.TrimSpace(issue.Verification); verification != "" {
		body = todos.UpsertVerificationFixture(body, verification)
	}
	frontmatter := types.TODOFrontmatter{
		Title:    issue.Title,
		Priority: priority,
		Status:   status,
	}
	frontmatter.Metadata = map[string]any{
		metadataID:     issue.ID.String(),
		metadataLabels: append([]string(nil), issue.Labels...),
	}
	return todos.WriteFrontmatter(&frontmatter, "\n"+body+"\n")
}

func portableIssueMatches(issue *native.Issue, input importIssue) bool {
	return issue.Title == input.Title && issue.Body == input.Body && issue.Verification == input.Verification &&
		issue.Priority == input.Priority && issue.Status == input.Status && sameStrings(issue.Labels, input.Labels)
}

func sameStrings(left, right []string) bool {
	left = append([]string(nil), left...)
	right = append([]string(nil), right...)
	sort.Strings(left)
	sort.Strings(right)
	return len(left) == len(right) && strings.Join(left, "\x00") == strings.Join(right, "\x00")
}

func importPriority(priority types.Priority) (native.Priority, error) {
	switch priority {
	case types.PriorityLow:
		return native.PriorityLow, nil
	case types.PriorityMedium:
		return native.PriorityMedium, nil
	case types.PriorityHigh:
		return native.PriorityHigh, nil
	default:
		return "", fmt.Errorf("priority %q is not portable; supported values are low, medium, and high", priority)
	}
}

func exportPriority(priority native.Priority) (types.Priority, error) {
	switch priority {
	case native.PriorityLow:
		return types.PriorityLow, nil
	case native.PriorityMedium:
		return types.PriorityMedium, nil
	case native.PriorityHigh:
		return types.PriorityHigh, nil
	default:
		return "", fmt.Errorf("priority %q is not representable in .todos", priority)
	}
}

func importStatus(status types.Status) (native.IssueStatus, error) {
	switch status {
	case types.StatusDraft:
		return native.StatusDraft, nil
	case types.StatusPending:
		return native.StatusOpen, nil
	case types.StatusVerified:
		return native.StatusVerified, nil
	case types.StatusCompleted:
		return native.StatusClosed, nil
	default:
		return "", fmt.Errorf("status %q is not portable; supported values are draft, pending, verified, and completed", status)
	}
}

func exportStatus(status native.IssueStatus) (types.Status, error) {
	switch status {
	case native.StatusDraft:
		return types.StatusDraft, nil
	case native.StatusOpen:
		return types.StatusPending, nil
	case native.StatusVerified:
		return types.StatusVerified, nil
	case native.StatusClosed:
		return types.StatusCompleted, nil
	default:
		return "", fmt.Errorf("status %q is not representable in .todos", status)
	}
}

func importedLabels(metadata map[string]any) []string {
	if metadata == nil {
		return nil
	}
	var labels []string
	switch value := metadata[metadataLabels].(type) {
	case string:
		labels = append(labels, value)
	case []string:
		labels = append(labels, value...)
	case []any:
		for _, item := range value {
			if label, ok := item.(string); ok {
				labels = append(labels, label)
			}
		}
	}
	return labels
}

func metadataString(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	value, _ := metadata[key].(string)
	return value
}

func portableFilename(issue native.Issue) string {
	slug := strings.Trim(filenameInvalid.ReplaceAllString(strings.ToLower(issue.Title), "-"), "-")
	if slug == "" {
		slug = "todo"
	}
	return slug + "-" + strings.ReplaceAll(issue.ID.String(), "-", "")[:8] + ".md"
}

func refuseUnrelatedCollision(path string, expected uuid.UUID, force bool) error {
	if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	if force {
		return nil
	}
	parsed, err := todos.ParseFrontmatterFromFile(path)
	if err == nil {
		if id, parseErr := uuid.Parse(metadataString(parsed.Frontmatter.Metadata, metadataID)); parseErr == nil && id == expected {
			return nil
		}
	}
	return fmt.Errorf("refuse to overwrite unrelated portable TODO %s; use --force to replace it", path)
}

func atomicWrite(path, content string) (returnErr error) {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".todo-export-*")
	if err != nil {
		return fmt.Errorf("create portable TODO temporary file: %w", err)
	}
	name := temporary.Name()
	defer func() {
		if err := os.Remove(name); err != nil && !errors.Is(err, os.ErrNotExist) {
			returnErr = errors.Join(returnErr, err)
		}
	}()
	if err := temporary.Chmod(0o644); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.WriteString(content); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(name, path); err != nil {
		return fmt.Errorf("publish portable TODO %s: %w", path, err)
	}
	return nil
}

func absolutePath(label, value, fallback string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = fallback
	}
	if value == "" {
		return "", fmt.Errorf("%s is required", label)
	}
	absolute, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("resolve %s %q: %w", label, value, err)
	}
	return filepath.Clean(absolute), nil
}

func portableDirectory(label, workDir, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = DefaultDirectory
	}
	if !filepath.IsAbs(value) {
		value = filepath.Join(workDir, value)
	}
	return absolutePath(label, value, "")
}
