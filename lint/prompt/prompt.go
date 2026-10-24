// Package prompt runs Captain prompts and reports their structured lint findings.
package prompt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/git/changeset"
	"github.com/flanksource/gavel/lint"
	"github.com/flanksource/gavel/linters"
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
	Concurrency   int                        `flag:"concurrency" default:"4" help:"Maximum prompt runs in flight; each file or directory is its own run"`
	AIFix         bool                       `flag:"ai-fix" help:"Use lint.fix from Gavel config to repair findings and re-run the prompt until clean"`
	AIFixMaxIters int                        `flag:"ai-fix-max-iterations" default:"3" help:"Max AI→re-lint cycles"`
	UI            bool                       `flag:"ui" help:"Launch browser UI to view violations"`
	Addr          string                     `flag:"addr" help:"Interface to bind --ui HTTP server. Defaults to 0.0.0.0 (all interfaces); set localhost to restrict to this machine." default:"0.0.0.0"`
	Stdin         io.Reader                  `flag:"-" json:"-" yaml:"-"`
	// Paths are the files and directories under review; each is its own prompt
	// run, passed to the prompt as the files template variable.
	Paths     []string             `flag:"-" json:"-" yaml:"-"`
	Selection *changeset.Selection `flag:"-" json:"-" yaml:"-"`
	// Publisher, when set, sees every run's result as it changes: pending before
	// any prompt starts, running once launched, then final.
	Publisher Publisher `flag:"-" json:"-" yaml:"-"`
}

// Publisher receives in-flight lint results. testui.Server implements it.
type Publisher interface {
	UpsertLintResult(*linters.LinterResult)
	RemoveLintResult(linter, workDir string)
}

// LintScope maps the prompt's paths and change selection onto lint.Options so
// the scope resolves exactly as it does for `gavel lint`.
func (opts Options) LintScope(ctx context.Context, workDir string) lint.Options {
	return lint.Options{WorkDir: workDir, Context: ctx, Changed: opts.Changed, Since: opts.Since, ChangedFiles: opts.ChangedFiles, Files: opts.Paths, Selection: opts.Selection}
}

// scope resolves the paths under review, relative to workDir. skip reports a
// change selection with nothing in scope.
func (opts Options) scope(ctx context.Context, workDir string) ([]string, bool, error) {
	paths := opts.Paths
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			path = filepath.Join(workDir, path)
		}
		if _, err := os.Stat(path); err != nil {
			return nil, false, fmt.Errorf("prompt lint path: %w", err)
		}
	}
	if scope := opts.LintScope(ctx, workDir); scope.ChangeSelection().Active() {
		resolved, skip, err := scope.WithChanges()
		if err != nil || skip {
			return nil, skip, err
		}
		paths = make([]string, 0, len(resolved.Files))
		for _, file := range resolved.Files {
			rel, err := filepath.Rel(workDir, file)
			if err != nil {
				return nil, false, err
			}
			paths = append(paths, rel)
		}
	}
	if len(paths) == 0 {
		return nil, false, nil
	}
	if slices.ContainsFunc(opts.Prompt.Var, func(v string) bool { return strings.HasPrefix(v, "files=") }) {
		return nil, false, fmt.Errorf("--var files conflicts with the path arguments and change selection")
	}
	if opts.Prompt.Vars != "" {
		var vars map[string]any
		if err := json.Unmarshal([]byte(opts.Prompt.Vars), &vars); err != nil {
			return nil, false, fmt.Errorf("decode --vars: %w", err)
		}
		if _, ok := vars["files"]; ok {
			return nil, false, fmt.Errorf("--vars files conflicts with the path arguments and change selection")
		}
	}
	return paths, false, nil
}

// review is one prompt run: one path under review, or the whole prompt when
// there is no path scope.
type review struct {
	Name   string
	Prompt captaincli.AIPromptOptions
}

func (opts Options) reviews(paths []string) []review {
	if len(paths) == 0 {
		return []review{{Name: opts.name(), Prompt: opts.Prompt}}
	}
	reviews := make([]review, 0, len(paths))
	for _, path := range paths {
		prompt := opts.Prompt
		prompt.Var = append(slices.Clone(opts.Prompt.Var), "files="+path)
		reviews = append(reviews, review{Name: opts.name() + ":" + path, Prompt: prompt})
	}
	return reviews
}

// Run reviews each path in scope as its own Captain prompt run. Every run is
// published as pending before the first one launches.
func Run(ctx context.Context, opts Options) ([]*linters.LinterResult, error) {
	exclusions, err := lint.ParseRuleExclusions(opts.Exclude)
	if err != nil {
		return nil, err
	}
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	paths, skip, err := opts.scope(ctx, workDir)
	if err != nil || skip {
		return nil, err
	}
	var stdin []byte
	if opts.Stdin != nil {
		// Every run renders the prompt, so piped input is read once and replayed.
		if stdin, err = io.ReadAll(opts.Stdin); err != nil {
			return nil, fmt.Errorf("read prompt stdin: %w", err)
		}
	}
	results := runReviews(ctx, reviewsInput{
		Reviews: opts.reviews(paths), WorkDir: workDir, Stdin: stdin, File: opts.Prompt.File,
		Concurrency: opts.Concurrency, Publisher: opts.Publisher,
	})
	exclusions.FilterResults(results)
	return results, nil
}

func Render(ctx context.Context, opts Options) ([]captaincli.PromptRenderResult, error) {
	if _, err := lint.ParseRuleExclusions(opts.Exclude); err != nil {
		return nil, err
	}
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	paths, skip, err := opts.scope(ctx, workDir)
	if err != nil || skip {
		return nil, err
	}
	var stdin []byte
	if opts.Stdin != nil {
		if stdin, err = io.ReadAll(opts.Stdin); err != nil {
			return nil, fmt.Errorf("read prompt stdin: %w", err)
		}
	}
	var rendered []captaincli.PromptRenderResult
	for _, review := range opts.reviews(paths) {
		run := &reviewRun{WorkDir: workDir}
		result, err := captaincli.RenderPrompt(ctx, invocation(opts.Prompt.File, review.Prompt, stdinReader(stdin), run.tools()))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", review.Name, err)
		}
		rendered = append(rendered, result)
	}
	return rendered, nil
}

func invocation(file string, prompt captaincli.AIPromptOptions, stdin io.Reader, tools []api.ToolDefinition) captaincli.PromptInvocation {
	prompt.RequiredPresets = append(append([]string(nil), prompt.RequiredPresets...), "lint")
	return captaincli.PromptInvocation{Reference: file, Options: prompt, Stdin: stdin, Tools: tools}
}

func stdinReader(stdin []byte) io.Reader {
	if stdin == nil {
		return nil
	}
	return bytes.NewReader(stdin)
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
