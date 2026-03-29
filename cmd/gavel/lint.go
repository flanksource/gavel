package main

import (
	"os"
	"os/exec"
	"strings"

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
)

type LintOptions struct {
	Linters string   `flag:"linters" help:"Comma-separated linter names or * for all" default:"*"`
	Ignore  []string `flag:"ignore" help:"Glob patterns to exclude from linting"`
	Fix     bool     `flag:"fix" help:"Enable auto-fixing"`
	NoCache bool     `flag:"no-cache" help:"Disable caching/debounce"`
	WorkDir string   `flag:"work-dir" help:"Working directory"`
	Files   []string `args:"true"`
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
  gavel lint ./pkg/...`
}

func init() {
	clicky.AddNamedCommand("lint", rootCmd, LintOptions{}, runLint)
}

func runLint(opts LintOptions) (any, error) {
	if opts.WorkDir == "" {
		opts.WorkDir, _ = os.Getwd()
	}
	logger.Infof("Running linters %s", opts.Pretty().ANSI())

	// Register linters with the working directory
	linters.DefaultRegistry.Register(golangci.NewGolangciLint(opts.WorkDir))
	linters.DefaultRegistry.Register(ruff.NewRuff(opts.WorkDir))
	linters.DefaultRegistry.Register(eslint.NewESLint(opts.WorkDir))
	linters.DefaultRegistry.Register(pyright.NewPyright(opts.WorkDir))
	linters.DefaultRegistry.Register(markdownlint.NewMarkdownlint(opts.WorkDir))
	linters.DefaultRegistry.Register(vale.NewVale(opts.WorkDir))
	linters.DefaultRegistry.Register(jscpd.NewJSCPD(opts.WorkDir))

	// Check if any positional args are linter names
	var linterFilter []string
	var actualFiles []string
	for _, f := range opts.Files {
		if linters.DefaultRegistry.Has(f) {
			linterFilter = append(linterFilter, f)
		} else {
			actualFiles = append(actualFiles, f)
		}
	}
	opts.Files = actualFiles

	// Filter linters based on --linters flag or positional linter names
	requestedLinters := linters.DefaultRegistry.List()
	if len(linterFilter) > 0 {
		requestedLinters = linterFilter
	} else if opts.Linters != "*" {
		requestedLinters = strings.Split(opts.Linters, ",")
	}

	var allResults []*linters.LinterResult
	for _, name := range requestedLinters {
		linter, ok := linters.DefaultRegistry.Get(strings.TrimSpace(name))
		if !ok {
			logger.Warnf("Unknown linter: %s (available: %s)", name, strings.Join(linters.DefaultRegistry.List(), ", "))
			continue
		}

		// Skip linters with no matching files in the project
		if !hasMatchingFiles(opts.WorkDir, linter.DefaultIncludes()) {
			logger.V(2).Infof("Skipping %s: no matching files", linter.Name())
			allResults = append(allResults, &linters.LinterResult{
				Linter:  linter.Name(),
				Skipped: true,
				Error:   "no matching files",
			})
			continue
		}

		// Skip linters whose binary is not on PATH
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
			ForceJSON: true,
		}

		if mixin, ok := linter.(linters.OptionsMixin); ok {
			mixin.SetOptions(runOpts)
		}

		result := linters.RunLinter(linter, runOpts)
		allResults = append(allResults, &result)
	}

	return allResults, nil
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
