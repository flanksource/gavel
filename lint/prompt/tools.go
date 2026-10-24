package prompt

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/internal/fileread"
	"github.com/flanksource/gavel/lint"
	"github.com/flanksource/gavel/models"
)

const (
	readFileTool        = "read_file"
	lintTool            = "lint"
	reportViolationTool = "report_violation"
)

// reviewRun is one path's prompt run: the tools its agent calls are bound to
// its work dir, and the violations it reports through them collect here.
type reviewRun struct {
	WorkDir string

	mu       sync.Mutex
	reported []models.Violation
	onReport func([]models.Violation)
}

// setOnReport registers fn to see the reported violations after each report.
func (r *reviewRun) setOnReport(fn func([]models.Violation)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onReport = fn
}

// Reported returns the violations reported through report_violation so far.
func (r *reviewRun) Reported() []models.Violation {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.reported)
}

func (r *reviewRun) tools() []api.ToolDefinition {
	readOnly, writes := true, false
	return []api.ToolDefinition{
		{
			Name:        readFileTool,
			Description: "Read a text file under the review's work dir. offset is the 1-based first line and limit the number of lines; omit both for the whole file.",
			InputSchema: objectSchema(map[string]any{
				"path":   map[string]any{"type": "string", "description": "Path relative to the work dir"},
				"offset": map[string]any{"type": "integer", "minimum": 1},
				"limit":  map[string]any{"type": "integer", "minimum": 1},
			}, "path"),
			ReadOnlyHint: &readOnly, DefaultPermission: api.ToolPolicyAllow,
			Handler: r.readFile,
		},
		{
			Name:        lintTool,
			Description: "Run the project's configured linters (as gavel lint does) on files or directories and return their violations.",
			InputSchema: objectSchema(map[string]any{
				"paths":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "minItems": 1, "description": "Paths relative to the work dir"},
				"linters": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Limit to these linters"},
			}, "paths"),
			ReadOnlyHint: &readOnly, DefaultPermission: api.ToolPolicyAllow,
			Handler: r.lint,
		},
		{
			Name:        reportViolationTool,
			Description: "Report one violation as soon as you find it. Reported violations are part of the review's result even if the final answer omits them.",
			InputSchema: objectSchema(map[string]any{
				"file":       map[string]any{"type": "string", "description": "Path relative to the work dir"},
				"line":       map[string]any{"type": "integer", "minimum": 0},
				"column":     map[string]any{"type": "integer", "minimum": 0},
				"end_line":   map[string]any{"type": "integer", "minimum": 0},
				"end_column": map[string]any{"type": "integer", "minimum": 0},
				"severity":   map[string]any{"type": "string", "enum": []string{"error", "warning", "info"}},
				"message":    map[string]any{"type": "string", "minLength": 1},
				"rule": objectSchema(map[string]any{
					"method":  map[string]any{"type": "string"},
					"package": map[string]any{"type": "string"},
				}),
			}, "message"),
			ReadOnlyHint: &writes, DefaultPermission: api.ToolPolicyAllow,
			Handler: r.reportViolation,
		},
	}
}

type readFileInput struct {
	Path   string `json:"path"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

type readFileOutput struct {
	Path       string `json:"path"`
	StartLine  int    `json:"start_line"`
	EndLine    int    `json:"end_line"`
	TotalLines int    `json:"total_lines"`
	Content    string `json:"content"`
}

func (r *reviewRun) readFile(_ context.Context, raw map[string]any) (any, error) {
	input, err := decodeToolInput[readFileInput](raw)
	if err != nil {
		return nil, err
	}
	data, err := fileread.Read(r.WorkDir, input.Path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", input.Path, err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	start := max(input.Offset, 1)
	if start > len(lines) {
		return nil, fmt.Errorf("offset %d is past the end of %s (%d lines)", start, input.Path, len(lines))
	}
	end := len(lines)
	if input.Limit > 0 {
		end = min(end, start+input.Limit-1)
	}
	return readFileOutput{
		Path: input.Path, StartLine: start, EndLine: end, TotalLines: len(lines),
		Content: strings.Join(lines[start-1:end], "\n"),
	}, nil
}

type lintInput struct {
	Paths   []string `json:"paths"`
	Linters []string `json:"linters"`
}

func (r *reviewRun) lint(ctx context.Context, raw map[string]any) (any, error) {
	input, err := decodeToolInput[lintInput](raw)
	if err != nil {
		return nil, err
	}
	if len(input.Paths) == 0 {
		return nil, fmt.Errorf("lint requires one or more paths")
	}
	for _, path := range input.Paths {
		if !filepath.IsLocal(path) {
			return nil, fmt.Errorf("lint path %q must be relative and stay inside the work dir", path)
		}
	}
	results, err := lint.Execute(lint.Options{WorkDir: r.WorkDir, Context: ctx, Files: input.Paths, Linters: input.Linters})
	if err != nil {
		return nil, err
	}
	var violations []models.Violation
	var failures []string
	for _, result := range results {
		if result.Error != "" {
			failures = append(failures, result.Linter+": "+result.Error)
		}
		for _, v := range result.Violations {
			if rel, err := filepath.Rel(r.WorkDir, v.File); err == nil && filepath.IsAbs(v.File) {
				v.File = rel
			}
			violations = append(violations, v)
		}
	}
	return map[string]any{"violations": violations, "linter_errors": failures}, nil
}

func (r *reviewRun) reportViolation(_ context.Context, raw map[string]any) (any, error) {
	v, err := decodeToolInput[models.Violation](raw)
	if err != nil {
		return nil, err
	}
	if err := validateViolation(&v, r.WorkDir); err != nil {
		return nil, err
	}
	r.mu.Lock()
	r.reported = append(r.reported, v)
	reported, onReport := slices.Clone(r.reported), r.onReport
	r.mu.Unlock()
	if onReport != nil {
		onReport(reported)
	}
	return map[string]any{"recorded": true, "reported": len(reported)}, nil
}

// mergeViolations is the run's final finding list: the structured answer, plus
// any violation reported through the tool that the answer left out.
func mergeViolations(structured, reported []models.Violation) []models.Violation {
	merged := slices.Clone(structured)
	key := func(v models.Violation) string {
		var rule, message string
		if v.Rule != nil {
			rule = v.Rule.Package + "/" + v.Rule.Method
		}
		if v.Message != nil {
			message = *v.Message
		}
		return fmt.Sprintf("%s:%d:%s:%s", v.File, v.Line, rule, message)
	}
	seen := map[string]bool{}
	for _, v := range merged {
		seen[key(v)] = true
	}
	for _, v := range reported {
		if !seen[key(v)] {
			seen[key(v)] = true
			merged = append(merged, v)
		}
	}
	return merged
}

func decodeToolInput[T any](raw map[string]any) (T, error) {
	var out T
	data, err := json.Marshal(raw)
	if err != nil {
		return out, fmt.Errorf("encode tool input: %w", err)
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return out, fmt.Errorf("decode tool input: %w", err)
	}
	return out, nil
}

func objectSchema(properties map[string]any, required ...string) map[string]any {
	schema := map[string]any{"type": "object", "properties": properties, "additionalProperties": false}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}
