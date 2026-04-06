package linters

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

const (
	GroupByFile    = "file"
	GroupByPackage = "package"
	GroupByMessage = "message"
)

type SyncOptions struct {
	TodosDir string
	GroupBy  string
	WorkDir  string
}

type SyncResult struct {
	Created   []string
	Updated   []string
	Completed []string
}

type violationGroup struct {
	Key        string
	Linters    map[string]bool
	Violations []models.Violation
	Files      map[string]bool
	Priority   types.Priority
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 80 {
		s = s[:80]
		s = strings.TrimRight(s, "-")
	}
	return s
}

func SyncLintTodos(results []*LinterResult, opts SyncOptions) (*SyncResult, error) {
	groups := groupViolations(results, opts.GroupBy, opts.WorkDir)

	if err := os.MkdirAll(opts.TodosDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create todos dir: %w", err)
	}

	existing, err := discoverExistingTodos(opts.TodosDir)
	if err != nil {
		return nil, fmt.Errorf("failed to discover existing todos: %w", err)
	}

	syncResult := &SyncResult{}
	seen := map[string]bool{}

	for slug, group := range groups {
		filename := slug + ".md"
		path := filepath.Join(opts.TodosDir, filename)
		seen[filename] = true

		if _, exists := existing[filename]; exists {
			if err := updateLintTodo(path, group); err != nil {
				return nil, fmt.Errorf("update todo %s: %w", path, err)
			}
			syncResult.Updated = append(syncResult.Updated, path)
		} else {
			if err := createLintTodo(path, group); err != nil {
				return nil, fmt.Errorf("create todo %s: %w", path, err)
			}
			syncResult.Created = append(syncResult.Created, path)
		}
	}

	for filename := range existing {
		if seen[filename] {
			continue
		}
		path := filepath.Join(opts.TodosDir, filename)
		if err := completeLintTodo(path); err != nil {
			logger.Warnf("failed to complete todo %s: %v", path, err)
			continue
		}
		syncResult.Completed = append(syncResult.Completed, path)
	}

	return syncResult, nil
}

func groupViolations(results []*LinterResult, groupBy, workDir string) map[string]violationGroup {
	groups := map[string]violationGroup{}

	for _, result := range results {
		if result == nil || !result.Success {
			continue
		}
		for _, v := range result.Violations {
			file := v.File
			if workDir != "" {
				if rel, err := filepath.Rel(workDir, file); err == nil {
					file = rel
				}
			}

			keys := violationGroupKeys(file, v, result.Linter, groupBy)
			for _, key := range keys {
				slug := slugify(key)
				g, ok := groups[slug]
				if !ok {
					g = violationGroup{
						Key:     key,
						Linters: map[string]bool{},
						Files:   map[string]bool{},
					}
				}
				g.Linters[result.Linter] = true
				g.Files[file] = true
				g.Violations = append(g.Violations, v)
				g.Priority = higherPriority(g.Priority, severityToPriority(v.Severity))
				groups[slug] = g
			}
		}
	}
	return groups
}

func violationGroupKeys(file string, v models.Violation, linter, groupBy string) []string {
	switch groupBy {
	case GroupByPackage:
		return []string{filepath.Dir(file)}
	case GroupByMessage:
		if v.Rule != nil && v.Rule.Method != "" {
			return []string{linter + "-" + v.Rule.Method}
		}
		if v.Message != nil {
			return []string{linter + "-" + *v.Message}
		}
		return []string{linter + "-unknown"}
	default: // file
		return []string{file}
	}
}

func severityToPriority(sev models.ViolationSeverity) types.Priority {
	switch sev {
	case models.SeverityError:
		return types.PriorityHigh
	case models.SeverityWarning:
		return types.PriorityMedium
	default:
		return types.PriorityLow
	}
}

func higherPriority(a, b types.Priority) types.Priority {
	if priorityRank(a) <= priorityRank(b) {
		return a
	}
	return b
}

func priorityRank(p types.Priority) int {
	switch p {
	case types.PriorityHigh:
		return 0
	case types.PriorityMedium:
		return 1
	case types.PriorityLow:
		return 2
	default:
		return 999
	}
}

func createLintTodo(path string, group violationGroup) error {
	now := time.Now()
	fm := types.TODOFrontmatter{
		Priority: group.Priority,
		Status:   types.StatusPending,
		LastRun:  &now,
		Path:     sortedKeys(group.Files),
	}

	body := formatViolationsBody(group)
	content, err := todos.WriteFrontmatter(&fm, body)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func updateLintTodo(path string, group violationGroup) error {
	parsed, err := todos.ParseFrontmatterFromFile(path)
	if err != nil {
		return err
	}

	parsed.Frontmatter.Priority = group.Priority
	parsed.Frontmatter.Path = sortedKeys(group.Files)

	if parsed.Frontmatter.Status == types.StatusCompleted {
		parsed.Frontmatter.Status = types.StatusPending
	}

	now := time.Now()
	parsed.Frontmatter.LastRun = &now

	body := formatViolationsBody(group)
	content, err := todos.WriteFrontmatter(&parsed.Frontmatter, body)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func completeLintTodo(path string) error {
	parsed, err := todos.ParseFrontmatterFromFile(path)
	if err != nil {
		return err
	}
	if parsed.Frontmatter.Status == types.StatusCompleted {
		return nil
	}
	parsed.Frontmatter.Status = types.StatusCompleted
	now := time.Now()
	parsed.Frontmatter.LastRun = &now

	content, err := todos.WriteFrontmatter(&parsed.Frontmatter, parsed.MarkdownContent)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func formatViolationsBody(group violationGroup) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "\n# Lint: %s\n\n", group.Key)

	linters := sortedKeys(group.Linters)
	fmt.Fprintf(&sb, "**Linters:** %s\n\n", strings.Join(linters, ", "))
	sb.WriteString("## Violations\n\n")

	maxViolations := 100
	for i, v := range group.Violations {
		if i >= maxViolations {
			fmt.Fprintf(&sb, "\n... and %d more violations\n", len(group.Violations)-maxViolations)
			break
		}
		loc := v.File
		if v.Line > 0 {
			loc = fmt.Sprintf("%s:%d", v.File, v.Line)
		}
		source := v.Source
		if v.Rule != nil && v.Rule.Method != "" {
			source = fmt.Sprintf("%s/%s", v.Source, v.Rule.Method)
		}
		msg := ""
		if v.Message != nil {
			msg = *v.Message
		}
		fmt.Fprintf(&sb, "- `%s` [%s] %s\n", loc, source, msg)
	}
	return sb.String()
}

func discoverExistingTodos(dir string) (map[string]bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]bool{}, nil
		}
		return nil, err
	}
	result := map[string]bool{}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			result[e.Name()] = true
		}
	}
	return result, nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
