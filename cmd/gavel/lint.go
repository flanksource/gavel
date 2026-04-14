package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/linters/eslint"
	"github.com/flanksource/gavel/linters/golangci"
	"github.com/flanksource/gavel/linters/jscpd"
	"github.com/flanksource/gavel/linters/markdownlint"
	"github.com/flanksource/gavel/linters/pyright"
	"github.com/flanksource/gavel/linters/ruff"
	"github.com/flanksource/gavel/linters/vale"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/repomap"
)

type LintOptions struct {
	Linters   string   `flag:"linters" help:"Comma-separated linter names or * for all" default:"*"`
	Ignore    []string `flag:"ignore" help:"Glob patterns to exclude from linting"`
	Triage    bool     `flag:"triage" help:"Interactive mode to select violation types to ignore"`
	Fix       bool     `flag:"fix" help:"Enable auto-fixing"`
	NoCache   bool     `flag:"no-cache" help:"Disable caching/debounce"`
	Timeout   string   `flag:"timeout" help:"Timeout per linter (e.g. 5m, 30s)" default:"5m"`
	SyncTodos string   `flag:"sync-todos" help:"Sync violations to TODO files in directory (default: .todos/lint)"`
	GroupBy   string   `flag:"group-by" help:"Group synced TODOs by: file, package, message" default:"file"`
	WorkDir   string   `flag:"work-dir" help:"Working directory"`
	Changed   bool     `flag:"changed" help:"Only report new issues vs origin/main (or $GAVEL_CHANGED_BASE)"`
	Since     string   `flag:"since" help:"Only report new issues since <ref> (merge-base with HEAD)"`
	UI        bool     `flag:"ui" help:"Launch browser UI to view violations"`
	Files     []string `args:"true"`
}

func (o LintOptions) Pretty() api.Text {
	t := clicky.Text("")
	if o.WorkDir != "" {
		t = t.Append("WorkDir: ", "text-muted").Append(o.WorkDir, "text-blue-500").Space()
	}
	if o.Linters != "*" {
		t = t.Append("Linters: ", "text-muted").Append(o.Linters, "text-blue-500").Space()
	}
	if o.Fix {
		t = t.Append("Fix: on", "text-green-500").Space()
	}
	if len(o.Files) > 0 {
		t = t.Append("Files: ", "text-muted").Append(clicky.CompactList(o.Files), "text-blue-500")
	}
	return t
}

func (o LintOptions) Help() string {
	return `Run linters on the project.

Automatically detects which linters are available and runs them.
Supports: golangci-lint, ruff, eslint, pyright, markdownlint, vale, jscpd.

Examples:
  gavel lint
  gavel lint jscpd
  gavel lint jscpd eslint
  gavel lint --linters=golangci-lint
  gavel lint --linters=golangci-lint,ruff
  gavel lint --fix
  gavel lint --triage
  gavel lint ./pkg/...`
}

func init() {
	clicky.AddNamedCommand("lint", rootCmd, LintOptions{}, runLint)
}

func runLint(opts LintOptions) (any, error) {
	if opts.WorkDir == "" {
		opts.WorkDir, _ = os.Getwd()
	}

	if opts.UI {
		uiServer = startTestUI()
	}

	groups := groupFilesByGitRoot(opts)
	opts.WorkDir = groups[0].gitRoot
	opts.Files = groups[0].files
	if uiServer != nil {
		uiServer.SetGitRoot(opts.WorkDir)
	}
	logger.Infof("Running linters %s", opts.Pretty().ANSI())

	var allResults []*linters.LinterResult
	for _, g := range groups {
		groupOpts := opts
		groupOpts.WorkDir = g.gitRoot
		groupOpts.Files = g.files
		results, err := executeLinters(groupOpts)
		if err != nil {
			return nil, err
		}
		allResults = append(allResults, results...)
	}

	gavelCfg, err := verify.LoadGavelConfig(opts.WorkDir)
	if err != nil {
		logger.Warnf("Failed to load .gavel.yaml: %v", err)
	}

	if filtered := linters.FilterIgnoredViolations(allResults, gavelCfg.Lint.Ignore); filtered > 0 {
		logger.Infof("Filtered %d ignored violations", filtered)
	}

	if uiServer != nil {
		uiServer.SetLintResults(allResults)
		uiServer.MarkDone()
		var violations int
		for _, lr := range allResults {
			if lr.Skipped {
				continue
			}
			violations += len(lr.Violations)
		}
		if violations > 0 {
			exitCode = 1
		}
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		return nil, nil
	}

	if opts.Triage {
		newRules, err := runTriage(allResults, opts.WorkDir)
		if err != nil {
			return nil, fmt.Errorf("triage failed: %w", err)
		}
		if len(newRules) > 0 {
			gavelCfg.Lint.Ignore = append(gavelCfg.Lint.Ignore, newRules...)
			gitRoot := repomap.FindGitRoot(opts.WorkDir)
			if gitRoot == "" {
				gitRoot = opts.WorkDir
			}
			if err := verify.SaveGavelConfig(gitRoot, gavelCfg); err != nil {
				return nil, fmt.Errorf("failed to save .gavel.yaml: %w", err)
			}
			logger.Infof("Saved %d new ignore rules to .gavel.yaml", len(newRules))
			linters.FilterIgnoredViolations(allResults, newRules)
		}
	}

	if opts.SyncTodos != "" {
		todosDir := filepath.Join(opts.SyncTodos, "lint")
		syncResult, err := linters.SyncLintTodos(allResults, linters.SyncOptions{
			TodosDir: todosDir,
			GroupBy:  opts.GroupBy,
			WorkDir:  opts.WorkDir,
		})
		if err != nil {
			return allResults, fmt.Errorf("failed to sync todos: %w", err)
		}
		logger.Infof("Synced TODOs: %d created, %d updated, %d completed",
			len(syncResult.Created), len(syncResult.Updated), len(syncResult.Completed))
	}

	return allResults, nil
}

// executeLinters runs all applicable linters and returns their results.
// Reusable by both the lint command and the test --lint flag.
func executeLinters(opts LintOptions) ([]*linters.LinterResult, error) {
	if opts.WorkDir == "" {
		opts.WorkDir, _ = os.Getwd()
	}

	timeout, err := time.ParseDuration(opts.Timeout)
	if err != nil {
		timeout = models.DefaultLinterTimeout
	}

	registry := linters.NewRegistry()
	registry.Register(golangci.NewGolangciLint(opts.WorkDir))
	registry.Register(ruff.NewRuff(opts.WorkDir))
	registry.Register(eslint.NewESLint(opts.WorkDir))
	registry.Register(pyright.NewPyright(opts.WorkDir))
	registry.Register(markdownlint.NewMarkdownlint(opts.WorkDir))
	registry.Register(vale.NewVale(opts.WorkDir))
	registry.Register(jscpd.NewJSCPD(opts.WorkDir))

	// Check if any positional args are linter names
	var linterFilter []string
	var actualFiles []string
	for _, f := range opts.Files {
		if registry.Has(f) {
			linterFilter = append(linterFilter, f)
		} else {
			actualFiles = append(actualFiles, f)
		}
	}
	opts.Files = actualFiles

	requestedLinters := registry.List()
	if len(linterFilter) > 0 {
		requestedLinters = linterFilter
	} else if opts.Linters != "*" {
		requestedLinters = strings.Split(opts.Linters, ",")
	}

	var allResults []*linters.LinterResult
	for _, name := range requestedLinters {
		linter, ok := registry.Get(strings.TrimSpace(name))
		if !ok {
			logger.Warnf("Unknown linter: %s (available: %s)", name, strings.Join(registry.List(), ", "))
			continue
		}

		if !hasMatchingFiles(opts.WorkDir, linter.DefaultIncludes()) {
			logger.V(2).Infof("Skipping %s: no matching files", linter.Name())
			continue
		}

		if _, err := exec.LookPath(linter.Name()); err != nil {
			logger.V(2).Infof("Skipping %s: not found on PATH", linter.Name())
			allResults = append(allResults, &linters.LinterResult{
				Linter:  linter.Name(),
				Skipped: true,
				Error:   "not found on PATH",
			})
			continue
		}

		runOpts := linters.RunOptions{
			WorkDir:   opts.WorkDir,
			Files:     opts.Files,
			Ignores:   opts.Ignore,
			Fix:       opts.Fix,
			NoCache:   opts.NoCache,
			Timeout:   timeout,
			ForceJSON: true,
		}

		if linter.Name() == "golangci-lint" {
			if ref := lintBaseRef(opts); ref != "" {
				base, err := resolveMergeBase(opts.WorkDir, ref)
				if err != nil {
					logger.Warnf("golangci-lint --new-from-rev: %v", err)
				} else {
					runOpts.ExtraArgs = append(runOpts.ExtraArgs, "--new-from-rev="+base)
				}
			}
		}

		if mixin, ok := linter.(linters.OptionsMixin); ok {
			mixin.SetOptions(runOpts)
		}

		result := linters.RunLinter(linter, runOpts)
		result.WorkDir = opts.WorkDir
		allResults = append(allResults, &result)
	}

	return allResults, nil
}

// lintBaseRef returns the git ref to use for --new-from-rev computation, or
// "" when neither --since nor --changed was set. --since wins if both are
// present.
func lintBaseRef(opts LintOptions) string {
	if opts.Since != "" {
		return opts.Since
	}
	if opts.Changed {
		if v := os.Getenv("GAVEL_CHANGED_BASE"); v != "" {
			return v
		}
		return "origin/main"
	}
	return ""
}

// resolveMergeBase returns the merge-base commit between HEAD and ref.
// Mirrors golangci-lint's own --new-from-rev semantics: issues present at
// merge-base are ignored, only regressions relative to it are reported.
func resolveMergeBase(workDir, ref string) (string, error) {
	cmd := exec.Command("git", "merge-base", "HEAD", ref)
	cmd.Dir = workDir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git merge-base HEAD %s: %w", ref, err)
	}
	return strings.TrimSpace(string(out)), nil
}

type lintGroup struct {
	gitRoot string
	files   []string
}

func groupFilesByGitRoot(opts LintOptions) []lintGroup {
	if len(opts.Files) == 0 {
		gitRoot := utils.FindGitRoot(opts.WorkDir)
		if gitRoot == "" {
			gitRoot = opts.WorkDir
		}
		return []lintGroup{{gitRoot: gitRoot}}
	}

	groups := make(map[string][]string)
	var order []string
	for _, f := range opts.Files {
		abs, _ := filepath.Abs(f)

		isDir := false
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			isDir = true
		}

		dir := abs
		if !isDir {
			dir = filepath.Dir(abs)
		}

		gitRoot := utils.FindGitRoot(dir)
		if gitRoot == "" {
			gitRoot = dir
		}

		if _, ok := groups[gitRoot]; !ok {
			order = append(order, gitRoot)
		}
		if !isDir {
			rel, _ := filepath.Rel(gitRoot, abs)
			groups[gitRoot] = append(groups[gitRoot], rel)
		}
	}

	result := make([]lintGroup, 0, len(order))
	for _, root := range order {
		result = append(result, lintGroup{gitRoot: root, files: groups[root]})
	}
	return result
}

// hasMatchingFiles checks if any files in workDir match at least one of the glob patterns.
// Uses GlobWalk to bail early on first match instead of scanning the entire tree.
func hasMatchingFiles(workDir string, patterns []string) bool {
	fsys := os.DirFS(workDir)
	for _, pattern := range patterns {
		found := false
		_ = doublestar.GlobWalk(fsys, pattern, func(path string, d os.DirEntry) error {
			found = true
			return doublestar.SkipDir
		})
		if found {
			return true
		}
	}
	return false
}
