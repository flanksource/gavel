package linters

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/models"
)

// Linter represents a generic linter that can analyze files
type Linter interface {
	// Name returns the linter name (e.g., "golangci-lint", "eslint")
	Name() string

	// Run executes the linter and returns violations
	Run(ctx commonsContext.Context, task *clicky.Task) ([]models.Violation, error)

	// DefaultIncludes returns default file patterns this linter should process
	DefaultIncludes() []string

	// DefaultExcludes returns patterns this linter should ignore by default
	DefaultExcludes() []string

	// SupportsJSON returns true if linter supports JSON output
	SupportsJSON() bool

	// JSONArgs returns additional args needed for JSON output
	JSONArgs() []string

	// SupportsFix returns true if linter supports auto-fixing violations
	SupportsFix() bool

	// FixArgs returns additional args needed for fix mode
	FixArgs() []string

	// ValidateConfig validates linter-specific configuration
	ValidateConfig(config *models.LinterConfig) error
}

// OptionsMixin provides a way to set options on linters that support it
type OptionsMixin interface {
	SetOptions(opts RunOptions)
}

// MetadataProvider provides file and rule count information from linters
type MetadataProvider interface {
	GetFileCount() int
	GetRuleCount() int
}

// LinterWithLanguageSupport extends Linter to provide language-aware file filtering
type LinterWithLanguageSupport interface {
	Linter

	// GetSupportedLanguages returns the languages this linter can process
	GetSupportedLanguages() []string

	// GetEffectiveExcludes returns the complete list of exclusion patterns
	// using the all_language_excludes macro for the given language and config
	GetEffectiveExcludes(language string, config *models.Config) []string

	// GetEffectiveIncludes returns the complete list of inclusion patterns
	// for the given language and config
	GetEffectiveIncludes(language string, config *models.Config) []string
}

// RunOptions provides configuration for linter execution
type RunOptions struct {
	WorkDir    string
	Files      []string
	Config     *models.LinterConfig
	ArchConfig *models.Config // Full arch-unit config for all_language_excludes macro
	ForceJSON  bool
	Fix        bool // Enable auto-fixing mode
	NoCache    bool // Disable caching
	Timeout    time.Duration
	Ignores    []string
	ExtraArgs  []string
}

// Registry manages available linters
type Registry struct {
	linters map[string]Linter
}

// NewRegistry creates a new linter registry
func NewRegistry() *Registry {
	return &Registry{
		linters: make(map[string]Linter),
	}
}

// Register adds a linter to the registry
func (r *Registry) Register(linter Linter) {
	r.linters[linter.Name()] = linter
}

// Get retrieves a linter by name
func (r *Registry) Get(name string) (Linter, bool) {
	l, ok := r.linters[name]
	return l, ok
}

// List returns all registered linter names
func (r *Registry) List() []string {
	var names []string
	for name := range r.linters {
		names = append(names, name)
	}
	return names
}

// Has checks if a linter is registered
func (r *Registry) Has(name string) bool {
	_, ok := r.linters[name]
	return ok
}

// Count returns the number of registered linters
func (r *Registry) Count() int {
	return len(r.linters)
}

// Global registry instance
var DefaultRegistry = NewRegistry()

// LinterResult represents the result of running a linter
type LinterResult struct {
	Linter       string             `json:"linter"`
	Success      bool               `json:"success"`
	Skipped      bool               `json:"skipped,omitempty"`
	TimedOut     bool               `json:"timed_out,omitempty"`
	Duration     time.Duration      `json:"duration"`
	Violations   []models.Violation `json:"violations"`
	RawOutput    string             `json:"raw_output,omitempty"`
	Error        string             `json:"error,omitempty"`
	Debounced    bool               `json:"debounced,omitempty"`
	DebounceUsed time.Duration      `json:"debounce_used,omitempty"`
	FileCount    int                `json:"file_count,omitempty"`
	RuleCount    int                `json:"rule_count,omitempty"`
}

// GetViolationCount returns the number of violations found
func (lr *LinterResult) GetViolationCount() int {
	return len(lr.Violations)
}

// HasViolations returns true if violations were found
func (lr *LinterResult) HasViolations() bool {
	return len(lr.Violations) > 0
}

// IsSuccessWithViolations returns true if the linter ran successfully but found violations
func (lr *LinterResult) IsSuccessWithViolations() bool {
	return lr.Success && lr.HasViolations()
}

// Pretty returns a formatted text representation of the linter result
func (lr *LinterResult) Pretty() api.Text {
	var status string
	var style string

	if lr.Skipped {
		status = "⊘"
		style = "text-muted"
	} else if lr.TimedOut {
		status = "⏱"
		style = "text-red-600"
	} else if lr.Success {
		if lr.HasViolations() {
			status = "⚠️"
			style = "text-yellow-600"
		} else {
			status = "✅"
			style = "text-green-600"
		}
	} else {
		status = "❌"
		style = "text-red-600"
	}

	text := fmt.Sprintf("%s %s", status, lr.Linter)
	if lr.Skipped {
		text += " (skipped: " + lr.Error + ")"
	} else if lr.TimedOut {
		text += fmt.Sprintf(" (timed out after %v)", lr.Duration)
	} else if lr.Debounced {
		text += fmt.Sprintf(" (cached, %v)", lr.DebounceUsed)
	} else {
		if lr.FileCount > 0 {
			text += fmt.Sprintf(" (%d violations, %d files, %v)", len(lr.Violations), lr.FileCount, lr.Duration)
		} else {
			text += fmt.Sprintf(" (%d violations, %v)", len(lr.Violations), lr.Duration)
		}
	}

	if lr.Error != "" && !lr.Skipped {
		text += fmt.Sprintf(" - Error: %s", lr.Error)
	}

	return api.Text{Content: text, Style: style}
}

// GetChildren groups violations by file for tree rendering.
func (lr *LinterResult) GetChildren() []api.TreeNode {
	if len(lr.Violations) == 0 {
		return nil
	}

	// Group violations by file
	fileMap := make(map[string][]models.Violation)
	var fileOrder []string
	for _, v := range lr.Violations {
		if _, seen := fileMap[v.File]; !seen {
			fileOrder = append(fileOrder, v.File)
		}
		fileMap[v.File] = append(fileMap[v.File], v)
	}
	sort.Strings(fileOrder)

	wd, _ := os.Getwd()
	var children []api.TreeNode
	for _, file := range fileOrder {
		children = append(children, &fileViolationNode{
			file:       file,
			workDir:    wd,
			violations: fileMap[file],
		})
	}
	return children
}

// fileViolationNode groups violations under a file path in the tree.
type fileViolationNode struct {
	file       string
	workDir    string
	violations []models.Violation
}

func (f *fileViolationNode) Pretty() api.Text {
	rel := f.file
	if f.workDir != "" {
		if r, err := filepath.Rel(f.workDir, f.file); err == nil {
			rel = r
		}
	}
	return api.Text{}.Append("📄 ", "").Append(rel, "text-blue-500")
}

func (f *fileViolationNode) GetChildren() []api.TreeNode {
	var children []api.TreeNode
	for i := range f.violations {
		children = append(children, &violationNode{v: &f.violations[i]})
	}
	return children
}

// violationNode renders a single violation as a tree leaf.
type violationNode struct {
	v *models.Violation
}

func (vn *violationNode) Pretty() api.Text {
	t := api.Text{}
	verbose := clicky.Flags.LevelCount >= 1

	// Severity icon
	switch vn.v.Severity {
	case models.SeverityError:
		t = t.Append("✗ ", "text-red-500")
	default:
		t = t.Append("⚠ ", "text-yellow-500")
	}

	// In default mode, show line:col on the message line
	if !verbose && vn.v.Line > 0 {
		loc := fmt.Sprintf(":%d", vn.v.Line)
		if vn.v.Column > 0 {
			loc = fmt.Sprintf(":%d:%d", vn.v.Line, vn.v.Column)
		}
		t = t.Append(loc, "text-muted")
	}

	// Message
	if vn.v.Message != nil {
		t = t.Append(" "+*vn.v.Message, "")
	}

	// Rule name in parens
	if vn.v.Rule != nil && vn.v.Rule.Method != "" {
		t = t.Append(" ("+vn.v.Rule.Method+")", "text-muted")
	}

	// At -v, show source line with line number gutter and column pointer
	if verbose && vn.v.Line > 0 {
		if srcLine := readFileLine(vn.v.File, vn.v.Line); srcLine != "" {
			lineNo := fmt.Sprintf("%d", vn.v.Line)
			gutter := strings.Repeat(" ", len(lineNo))
			t = t.NewLine().Append(fmt.Sprintf("  %s │ ", lineNo), "text-muted").Append(srcLine, "max-w-[tw-20ch] truncate-suffix")
			if vn.v.Column > 0 {
				pointer := strings.Repeat(" ", vn.v.Column-1) + "^"
				t = t.NewLine().Append(fmt.Sprintf("  %s │ ", gutter), "text-muted").Append(pointer, "text-red-500")
			}
		}
	}

	// At -vv, show code fragment if available (e.g. jscpd duplicate code)
	if clicky.Flags.LevelCount >= 2 && vn.v.Code != nil && *vn.v.Code != "" {
		lang := ""
		if vn.v.Rule != nil {
			lang = strings.TrimPrefix(vn.v.Rule.Method, "duplicate-")
		}
		t = t.NewLine().Append(api.NewCode(*vn.v.Code, lang).ANSI(), "")
	}

	return t
}

func (vn *violationNode) GetChildren() []api.TreeNode { return nil }

// readFileLine reads a single line from a file (1-based line number).
func readFileLine(path string, lineNum int) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for i := 1; scanner.Scan(); i++ {
		if i == lineNum {
			return strings.TrimRight(scanner.Text(), " \t")
		}
	}
	return ""
}

// RunLinter executes a single linter and returns the result.
func RunLinter(linter Linter, opts RunOptions) LinterResult {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = models.DefaultLinterTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()

	task := clicky.StartTask[[]models.Violation](fmt.Sprintf("Running %s", linter.Name()), func(_ commonsContext.Context, t *clicky.Task) ([]models.Violation, error) {
		return linter.Run(commonsContext.NewContext(ctx), t)
	})
	violations, _ := task.GetResult()

	timedOut := ctx.Err() == context.DeadlineExceeded
	return LinterResult{
		Linter:     linter.Name(),
		Success:    task.IsOk() && !timedOut,
		TimedOut:   timedOut,
		Duration:   time.Since(start),
		Violations: violations,
		Error:      formatErr(task.Error()),
	}
}

func formatErr(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
