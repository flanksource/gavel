// Package prompt runs Captain prompts and reports their structured lint findings.
package prompt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/git/changeset"
	"github.com/flanksource/gavel/lint"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
)

type Options struct {
	// AIRuntimeOptions configure the --ai-fix fixer. The reviewing prompt's
	// runtime lives in Prompt, parsed from the arguments after --.
	captaincli.AIRuntimeOptions
	Prompt        captaincli.AIPromptOptions `flag:"-" json:"-" yaml:"-"`
	DryRun        bool                       `flag:"dry-run" help:"Render the prompt and lint schema without calling a model"`
	Exclude       []string                   `flag:"exclude" help:"Exclude rule IDs or * patterns, optionally source:rule (repeat or comma-separate)"`
	Summary       bool                       `flag:"summary" help:"Group findings by rule"`
	SummaryLimit  int                        `flag:"summary-limit" default:"5" help:"Maximum example locations per rule"`
	Changed       bool                       `flag:"changed" help:"Review local staged/unstaged/untracked files, or the runner changeset in CI"`
	Since         string                     `flag:"since" help:"Select changes since a ref's merge-base plus local edits; HEAD = local edits, staged = staged files"`
	ChangedFiles  []string                   `flag:"changed-files" help:"Explicit changed files, unioned with --changed or --since when supplied"`
	AIFix         bool                       `flag:"ai-fix" help:"Use lint.fix from Gavel config to repair findings and re-run the prompt until clean"`
	AIFixMaxIters int                        `flag:"ai-fix-max-iterations" default:"3" help:"Max AI→re-lint cycles"`
	Stdin         io.Reader                  `flag:"-" json:"-" yaml:"-"`
	// Paths are the files and directories under review, passed to the prompt
	// as the files template variable.
	Paths     []string             `flag:"-" json:"-" yaml:"-"`
	Selection *changeset.Selection `flag:"-" json:"-" yaml:"-"`
}

// LintScope maps the prompt's paths and change selection onto lint.Options so
// the scope resolves exactly as it does for `gavel lint`.
func (opts Options) LintScope(ctx context.Context, workDir string) lint.Options {
	return lint.Options{WorkDir: workDir, Context: ctx, Changed: opts.Changed, Since: opts.Since, ChangedFiles: opts.ChangedFiles, Files: opts.Paths, Selection: opts.Selection}
}

// scoped injects the review scope as the files variable. skip reports a change
// selection with nothing in scope.
func (opts Options) scoped(ctx context.Context, workDir string) (Options, bool, error) {
	paths := opts.Paths
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			path = filepath.Join(workDir, path)
		}
		if _, err := os.Stat(path); err != nil {
			return opts, false, fmt.Errorf("prompt lint path: %w", err)
		}
	}
	if scope := opts.LintScope(ctx, workDir); scope.ChangeSelection().Active() {
		resolved, skip, err := scope.WithChanges()
		if err != nil || skip {
			return opts, skip, err
		}
		paths = make([]string, 0, len(resolved.Files))
		for _, file := range resolved.Files {
			rel, err := filepath.Rel(workDir, file)
			if err != nil {
				return opts, false, err
			}
			paths = append(paths, rel)
		}
	}
	if len(paths) == 0 {
		return opts, false, nil
	}
	if slices.ContainsFunc(opts.Prompt.Var, func(v string) bool { return strings.HasPrefix(v, "files=") }) {
		return opts, false, fmt.Errorf("--var files conflicts with the path arguments and change selection")
	}
	if opts.Prompt.Vars != "" {
		var vars map[string]any
		if err := json.Unmarshal([]byte(opts.Prompt.Vars), &vars); err != nil {
			return opts, false, fmt.Errorf("decode --vars: %w", err)
		}
		if _, ok := vars["files"]; ok {
			return opts, false, fmt.Errorf("--vars files conflicts with the path arguments and change selection")
		}
	}
	opts.Prompt.Var = append(slices.Clone(opts.Prompt.Var), "files="+strings.Join(paths, " "))
	return opts, false, nil
}

func Run(ctx context.Context, opts Options) ([]*linters.LinterResult, error) {
	exclusions, err := lint.ParseRuleExclusions(opts.Exclude)
	if err != nil {
		return nil, err
	}
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	opts, skip, err := opts.scoped(ctx, workDir)
	if err != nil || skip {
		return nil, err
	}
	start := time.Now()
	run, runErr := captaincli.RunPrompt(ctx, invocation(opts))
	results := resultsFromRun(resultInput{Name: opts.name(), Run: run, Error: runErr, WorkDir: workDir, Duration: time.Since(start)})
	exclusions.FilterResults(results)
	return results, nil
}

func Render(ctx context.Context, opts Options) (captaincli.PromptRenderResult, error) {
	if _, err := lint.ParseRuleExclusions(opts.Exclude); err != nil {
		return captaincli.PromptRenderResult{}, err
	}
	workDir, err := os.Getwd()
	if err != nil {
		return captaincli.PromptRenderResult{}, err
	}
	opts, skip, err := opts.scoped(ctx, workDir)
	if err != nil || skip {
		return captaincli.PromptRenderResult{}, err
	}
	return captaincli.RenderPrompt(ctx, invocation(opts))
}

func invocation(opts Options) captaincli.PromptInvocation {
	opts.Prompt.RequiredPresets = append(append([]string(nil), opts.Prompt.RequiredPresets...), "lint")
	return captaincli.PromptInvocation{Reference: opts.Prompt.File, Options: opts.Prompt, Stdin: opts.Stdin}
}

func (opts Options) name() string {
	ref := opts.Prompt.File
	if ref == "" && strings.HasPrefix(opts.Prompt.PromptRef, "@") {
		ref = strings.TrimPrefix(opts.Prompt.PromptRef, "@")
	}
	if ref == "" || ref == "-" {
		return "prompt"
	}
	return strings.TrimSuffix(filepath.Base(ref), ".prompt")
}

type resultInput struct {
	Name     string
	Run      captaincli.PromptRunResult
	Error    error
	WorkDir  string
	Duration time.Duration
}

func resultsFromRun(input resultInput) []*linters.LinterResult {
	name := input.Name
	if name == "" {
		name = "prompt"
	}
	if len(input.Run.Runs) == 0 {
		output := runOutput{Name: name, WorkDir: input.WorkDir, Dir: input.Run.Dir, Data: input.Run.StructuredOutput, Duration: input.Run.Duration, Elapsed: input.Duration}
		if input.Error != nil {
			output.Error = input.Error.Error()
		}
		return []*linters.LinterResult{output.result()}
	}
	results := make([]*linters.LinterResult, 0, len(input.Run.Runs))
	for _, run := range input.Run.Runs {
		output := runOutput{Name: name + ":" + run.Selector, WorkDir: input.WorkDir, Dir: run.Dir, Data: run.StructuredOutput, Duration: run.Duration, Error: run.Error}
		if run.Status == "failed" && output.Error == "" {
			output.Error = "prompt execution failed"
		}
		results = append(results, output.result())
	}
	if input.Error != nil && !slices.ContainsFunc(results, func(result *linters.LinterResult) bool { return !result.Success }) {
		results = append(results, (runOutput{Name: name, WorkDir: input.WorkDir, Error: input.Error.Error()}).result())
	}
	return results
}

type runOutput struct {
	Name, WorkDir, Dir, Duration, Error string
	Data                                map[string]any
	Elapsed                             time.Duration
}

func (output runOutput) result() *linters.LinterResult {
	dir := output.WorkDir
	if output.Dir != "" {
		dir = output.Dir
	}
	result := &linters.LinterResult{Linter: output.Name, WorkDir: dir, Duration: output.Elapsed, Error: output.Error}
	if output.Duration != "" {
		var err error
		result.Duration, err = time.ParseDuration(output.Duration)
		if err != nil {
			result.Error = fmt.Sprintf("invalid prompt duration %q: %v", output.Duration, err)
		}
	}
	if result.Error != "" {
		return result
	}
	violations, err := decodeViolations(output.Data, dir)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Success, result.Violations = true, violations
	return result
}

func decodeViolations(data map[string]any, dir string) ([]models.Violation, error) {
	var output struct {
		Violations []models.Violation `json:"violations"`
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("encode prompt lint output: %w", err)
	}
	if err := json.Unmarshal(raw, &output); err != nil {
		return nil, fmt.Errorf("decode prompt lint output: %w", err)
	}
	if output.Violations == nil {
		return nil, fmt.Errorf("prompt lint output requires a violations array")
	}
	for i := range output.Violations {
		v := &output.Violations[i]
		if v.Message == nil || strings.TrimSpace(*v.Message) == "" {
			return nil, fmt.Errorf("prompt lint violation %d requires a nonempty message", i+1)
		}
		if v.Severity != "" && v.Severity != models.SeverityError && v.Severity != models.SeverityWarning && v.Severity != models.SeverityInfo {
			return nil, fmt.Errorf("prompt lint violation %d has invalid severity %q", i+1, v.Severity)
		}
		if v.Line < 0 || v.Column < 0 {
			return nil, fmt.Errorf("prompt lint violation %d has a negative location", i+1)
		}
		if v.File != "" && !filepath.IsAbs(v.File) {
			v.File = filepath.Join(dir, v.File)
		}
		if v.Source == "" {
			v.Source = "prompt"
		}
	}
	return output.Violations, nil
}
