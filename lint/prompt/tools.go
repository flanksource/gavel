package prompt

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/internal/fileread"
	"github.com/flanksource/gavel/lint"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
)

const (
	readFileTool        = "read_file"
	lintTool            = "lint"
	reportViolationTool = "report_violation"
)

// reviewRun is one review's prompt run: the tools its agent calls are bound to
// its work dir and file scope, and the violations it reports collect here.
type reviewRun struct {
	WorkDir string
	Scope   fileScope
	// Name prefixes the "<review>/<linter>" results the lint tool publishes.
	Name      string
	Publisher Publisher

	mu          sync.Mutex
	reported    []models.Violation
	onReport    func([]models.Violation)
	identityOf  runIdentity
	lintResults []*linters.LinterResult
}

type runIdentity struct{ RunID, SessionID string }

// setRun records the launched run, which the lint tool's results name. It is
// set once launchPrompt returns, before the agent can call a tool.
func (r *reviewRun) setRun(runID, sessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.identityOf = runIdentity{RunID: runID, SessionID: sessionID}
}

func (r *reviewRun) identity() runIdentity {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.identityOf
}

// linted returns the final results of the linters the agent ran, each naming
// sessionID when it was published before the session was known.
func (r *reviewRun) linted(sessionID string) []*linters.LinterResult {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*linters.LinterResult, len(r.lintResults))
	for i, result := range r.lintResults {
		copied := *result
		copied.SessionID = firstNonEmpty(copied.SessionID, sessionID)
		out[i] = &copied
	}
	return out
}

// record publishes a linter result and keeps it, replacing an earlier result
// for the same linter and work dir, as the publisher does.
func (r *reviewRun) record(result *linters.LinterResult) {
	r.mu.Lock()
	r.lintResults = slices.DeleteFunc(r.lintResults, func(old *linters.LinterResult) bool {
		return old.Linter == result.Linter && old.WorkDir == result.WorkDir
	})
	if !result.InFlight() {
		r.lintResults = append(r.lintResults, result)
	}
	r.mu.Unlock()
	if r.Publisher != nil {
		r.Publisher.UpsertLintResult(result)
	}
}

func (r *reviewRun) remove(linter, workDir string) {
	r.mu.Lock()
	r.lintResults = slices.DeleteFunc(r.lintResults, func(old *linters.LinterResult) bool {
		return old.Linter == linter && old.WorkDir == workDir
	})
	r.mu.Unlock()
	if r.Publisher != nil {
		r.Publisher.RemoveLintResult(linter, workDir)
	}
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
	required := []string{"message"}
	if r.Scope.Exact {
		required = append(required, "file")
	}
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
			Name: reportViolationTool,
			Description: "Report one or more violations as soon as you find them; batch the findings you have instead of one call each. " +
				"Give code (the offending text, verbatim) whenever the finding quotes the file: it is checked against file and line, and a quote that is not there is rejected with where it actually appears. " +
				"Each rejected entry is returned with its index and reason; the others are recorded. Reported violations are part of the review's result even if the final answer omits them.",
			InputSchema: objectSchema(map[string]any{
				"violations": map[string]any{
					"type":     "array",
					"minItems": 1,
					"items": objectSchema(map[string]any{
						"file":       r.Scope.fileProperty(),
						"line":       map[string]any{"type": "integer", "minimum": 0, "description": "1-based line of the offending text"},
						"column":     map[string]any{"type": "integer", "minimum": 0, "description": "1-based column; derived from code when omitted"},
						"end_line":   map[string]any{"type": "integer", "minimum": 0, "description": "Last line when code spans several lines"},
						"end_column": map[string]any{"type": "integer", "minimum": 0, "description": "Column just past the offending text; derived from code when omitted"},
						"severity":   map[string]any{"type": "string", "enum": []string{"error", "warning", "info"}},
						"message":    map[string]any{"type": "string", "minLength": 1},
						"code":       map[string]any{"type": "string", "minLength": 1, "description": "The offending text exactly as it appears in the file; requires file and line"},
						"rule": objectSchema(map[string]any{
							"method":  map[string]any{"type": "string"},
							"package": map[string]any{"type": "string"},
						}),
					}, required...),
				},
			}, "violations"),
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
	results, err := r.executeLint(ctx, input)
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

// executeLint runs the agent's lint call and publishes it as "<review>/lint"
// while running, then as one final "<review>/<linter>" result per linter run.
// A failed call stays published as a failed "<review>/lint" result.
func (r *reviewRun) executeLint(ctx context.Context, input lintInput) ([]*linters.LinterResult, error) {
	id := r.identity()
	row := func(linter string) *linters.LinterResult {
		return &linters.LinterResult{Linter: r.Name + "/" + linter, WorkDir: r.WorkDir, RunID: id.RunID, SessionID: id.SessionID, Violations: []models.Violation{}}
	}
	running := row("lint")
	running.Status = linters.LinterStatusRunning
	r.record(running)
	start := time.Now()
	results, err := lint.Execute(lint.Options{WorkDir: r.WorkDir, Context: ctx, Files: input.Paths, Linters: input.Linters})
	if err != nil {
		failed := row("lint")
		failed.Duration, failed.Error = time.Since(start), err.Error()
		r.record(failed)
		return nil, err
	}
	r.remove(running.Linter, running.WorkDir)
	for _, result := range results {
		published := *result
		published.Linter = r.Name + "/" + result.Linter
		published.WorkDir = firstNonEmpty(result.WorkDir, r.WorkDir)
		published.RunID, published.SessionID = id.RunID, id.SessionID
		if published.Violations == nil {
			published.Violations = []models.Violation{}
		}
		r.record(&published)
	}
	return results, nil
}

type reportInput struct {
	Violations []models.Violation `json:"violations"`
}

// rejectedViolation names an entry of a report_violation call that was not
// recorded, so the agent can correct and resend just that one.
type rejectedViolation struct {
	Index int    `json:"index"`
	Error string `json:"error"`
}

type reportOutput struct {
	Recorded int                 `json:"recorded"`
	Reported int                 `json:"reported"`
	Rejected []rejectedViolation `json:"rejected,omitempty"`
}

func (r *reviewRun) reportViolation(_ context.Context, raw map[string]any) (any, error) {
	input, err := decodeToolInput[reportInput](raw)
	if err != nil {
		return nil, err
	}
	if len(input.Violations) == 0 {
		return nil, fmt.Errorf("report_violation requires at least one violation")
	}
	var accepted []models.Violation
	var rejected []rejectedViolation
	for i, v := range input.Violations {
		err := locateCode(r.WorkDir, &v)
		if err == nil {
			err = validateViolation(&v, r.WorkDir)
		}
		if err == nil {
			err = r.Scope.check(r.WorkDir, v.File)
		}
		if err != nil {
			rejected = append(rejected, rejectedViolation{Index: i, Error: err.Error()})
			continue
		}
		accepted = append(accepted, v)
	}
	if len(accepted) == 0 {
		reasons := make([]string, len(rejected))
		for i, entry := range rejected {
			reasons[i] = fmt.Sprintf("violation %d: %s", entry.Index, entry.Error)
		}
		return nil, fmt.Errorf("no violation recorded: %s", strings.Join(reasons, "; "))
	}
	r.mu.Lock()
	r.reported = append(r.reported, accepted...)
	reported, onReport := slices.Clone(r.reported), r.onReport
	r.mu.Unlock()
	if onReport != nil {
		onReport(reported)
	}
	return reportOutput{Recorded: len(accepted), Reported: len(reported), Rejected: rejected}, nil
}

// locateCode checks a violation's quoted code against its file under dir: the
// text must sit within line..end_line. Missing columns are filled from the match.
func locateCode(dir string, v *models.Violation) error {
	if v.Code == nil || *v.Code == "" {
		return nil
	}
	code := *v.Code
	if v.File == "" || v.Line <= 0 {
		return fmt.Errorf("code requires file and line")
	}
	path := v.File
	if filepath.IsAbs(path) {
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return fmt.Errorf("file %s is not under the work dir: %w", v.File, err)
		}
		path = rel
	}
	data, err := fileread.Read(dir, path)
	if err != nil {
		return fmt.Errorf("read %s: %w", v.File, err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	last := max(v.EndLine, v.Line+strings.Count(code, "\n"))
	if v.Line <= len(lines) {
		window := strings.Join(lines[v.Line-1:min(last, len(lines))], "\n")
		if index := strings.Index(window, code); index >= 0 {
			if v.Column == 0 && !strings.Contains(window[:index], "\n") {
				v.Column = utf8.RuneCountInString(window[:index]) + 1
				if v.EndColumn == 0 && !strings.Contains(code, "\n") {
					v.EndColumn = v.Column + utf8.RuneCountInString(code)
				}
			}
			return nil
		}
	}
	for i := range lines {
		if strings.HasPrefix(strings.Join(lines[i:], "\n"), code) || strings.Contains(lines[i], code) {
			return fmt.Errorf("code %q is not on line %d of %s; it appears on line %d", code, v.Line, v.File, i+1)
		}
	}
	return fmt.Errorf("code %q does not appear in %s", code, v.File)
}

// mergeViolations is the run's final finding list: the structured answer, plus
// any violation reported through the tool that the answer left out. Where both
// carry the same finding, the reported copy wins: report_violation checked it.
func mergeViolations(structured, reported []models.Violation) []models.Violation {
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
	byKey := map[string]models.Violation{}
	for _, v := range reported {
		if _, dup := byKey[key(v)]; !dup {
			byKey[key(v)] = v
		}
	}
	seen := map[string]bool{}
	merged := make([]models.Violation, 0)
	for _, v := range structured {
		if checked, ok := byKey[key(v)]; ok {
			v = checked
		}
		if !seen[key(v)] {
			seen[key(v)] = true
			merged = append(merged, v)
		}
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
