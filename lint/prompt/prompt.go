// Package prompt runs Captain prompts and reports their structured lint findings.
package prompt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/git/changeset"
	"github.com/flanksource/gavel/lint"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/workspace"
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
	UI            bool                       `flag:"ui" help:"Launch browser UI to view violations"`
	Addr          string                     `flag:"addr" help:"Interface to bind --ui HTTP server. Defaults to 0.0.0.0 (all interfaces); set localhost to restrict to this machine." default:"0.0.0.0"`
	Stdin         io.Reader                  `flag:"-" json:"-" yaml:"-"`
	// Paths are the files and directories under review, all in one prompt run,
	// passed to the prompt as the files list.
	Paths     []string             `flag:"-" json:"-" yaml:"-"`
	Selection *changeset.Selection `flag:"-" json:"-" yaml:"-"`
	// Dirs are the invocation's resolved directories; paths are relative to
	// Dirs.Invocation.
	Dirs workspace.Dirs `flag:"-" json:"-" yaml:"-"`
	// Publisher, when set, sees every run's result as it changes: pending before
	// any prompt starts, running once launched, then final.
	Publisher Publisher            `flag:"-" json:"-" yaml:"-"`
	OnEvent   func(int, api.Event) `flag:"-" json:"-" yaml:"-"`
}

// Publisher receives in-flight lint results. testui.Server implements it.
type Publisher interface {
	UpsertLintResult(*linters.LinterResult)
	RemoveLintResult(linter, workDir string)
}

func (opts Options) workDir() (string, error) {
	if opts.Dirs.Invocation == "" {
		return "", errors.New("lint prompt: the invocation's workspace directories were not resolved")
	}
	return opts.Dirs.Invocation, nil
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
	if slices.ContainsFunc(opts.Prompt.Var, func(v string) bool { return varKey(v) == "files" }) {
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

// varKey is the template variable a --var pair sets: key=, key:= and key[]=
// all set key.
func varKey(pair string) string {
	key, _, _ := strings.Cut(pair, "=")
	key, _, _ = strings.Cut(strings.TrimSuffix(key, ":"), "[")
	return key
}

// review is one prompt run over every path in scope, or the bare prompt when
// there is no path scope.
type review struct {
	Name     string
	Prompt   captaincli.AIPromptOptions
	Scope    fileScope
	Prepared *captaincli.PreparedPrompt
	Run      *reviewRun
}

// review passes every path to the prompt as the files list, a typed --var.
func (opts Options) review(workDir string, paths []string) (review, error) {
	scope, err := newFileScope(workDir, paths)
	if err != nil {
		return review{}, err
	}
	prompt := opts.Prompt
	if len(paths) > 0 {
		files, err := json.Marshal(paths)
		if err != nil {
			return review{}, fmt.Errorf("encode files: %w", err)
		}
		prompt.Var = append(slices.Clone(opts.Prompt.Var), "files:="+string(files))
	}
	return review{Name: opts.name(), Prompt: prompt, Scope: scope}, nil
}

// Run reviews every path in scope in one Captain prompt run, published as
// pending before it launches.
func Run(ctx context.Context, opts Options) ([]*linters.LinterResult, error) {
	exclusions, err := lint.ParseRuleExclusions(opts.Exclude)
	if err != nil {
		return nil, err
	}
	workDir, err := opts.workDir()
	if err != nil {
		return nil, err
	}
	paths, skip, err := opts.scope(ctx, workDir)
	if err != nil || skip {
		return nil, err
	}
	scoped, err := opts.review(workDir, paths)
	if err != nil {
		return nil, err
	}
	var stdin []byte
	if opts.Stdin != nil {
		// Each model variant renders the prompt, so piped input is read once and replayed.
		if stdin, err = io.ReadAll(opts.Stdin); err != nil {
			return nil, fmt.Errorf("read prompt stdin: %w", err)
		}
	}
	results := runReviews(ctx, reviewsInput{
		Reviews: []review{scoped}, WorkDir: workDir, Stdin: stdin, File: opts.Prompt.File,
		Publisher: opts.Publisher, OnEvent: opts.OnEvent,
	})
	exclusions.FilterResults(results)
	if err := lint.FilterConfiguredViolations(results, workDir); err != nil {
		return nil, fmt.Errorf("apply lint.ignore to prompt findings: %w", err)
	}
	return results, nil
}

func Render(ctx context.Context, opts Options) ([]captaincli.PromptRenderResult, error) {
	if _, err := lint.ParseRuleExclusions(opts.Exclude); err != nil {
		return nil, err
	}
	workDir, err := opts.workDir()
	if err != nil {
		return nil, err
	}
	paths, skip, err := opts.scope(ctx, workDir)
	if err != nil || skip {
		return nil, err
	}
	scoped, err := opts.review(workDir, paths)
	if err != nil {
		return nil, err
	}
	run := &reviewRun{WorkDir: workDir, Scope: scoped.Scope}
	result, err := captaincli.RenderPrompt(ctx, invocation(invocationInput{File: opts.Prompt.File, Prompt: scoped.Prompt, Stdin: opts.Stdin, Tools: run.tools()}))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", scoped.Name, err)
	}
	return []captaincli.PromptRenderResult{result}, nil
}

type invocationInput struct {
	File   string
	Prompt captaincli.AIPromptOptions
	Stdin  io.Reader
	Tools  []api.ToolDefinition
	// Group and Name place a launched run's task in the caller's group.
	Group   *clickytask.Group
	Name    string
	OnEvent func(int, api.Event)
}

func invocation(input invocationInput) captaincli.PromptInvocation {
	prompt := input.Prompt
	prompt.RequiredPresets = append(append([]string(nil), prompt.RequiredPresets...), "lint")
	return captaincli.PromptInvocation{
		Reference: input.File, Options: prompt, Stdin: input.Stdin, Tools: input.Tools,
		Group: input.Group, Name: input.Name, OnEvent: input.OnEvent,
	}
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
