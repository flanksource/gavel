package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/testrunner/outline"
	"github.com/flanksource/gavel/testrunner/parsers"
)

type testOutlineOptions struct {
	Paths        []string `json:"paths,omitempty" args:"true"`
	AISummary    bool     `json:"ai_summary,omitempty" flag:"ai-summary" help:"Generate one-line AI summaries of what each test verifies"`
	Frameworks   []string `json:"frameworks,omitempty" flag:"framework" help:"Limit to frameworks: go, ginkgo, jest, vitest, playwright, fixture"`
	FixtureFiles []string `json:"fixture_files,omitempty" flag:"fixture-files" help:"Markdown fixture globs; overrides .gavel.yaml fixtures.files"`
	Duplication  bool     `json:"duplication,omitempty" flag:"duplication" default:"true" help:"Compute per-test duplication via jscpd"`
	History      bool     `json:"history,omitempty" flag:"history" default:"true" help:"Join run history from .gavel snapshots"`
}

func (o testOutlineOptions) Help() api.Textable {
	return clicky.Text(`Statically outline tests without running them.

Renders a directory/file tree with one row per Go test, Ginkgo spec, Jest test,
Vitest test, Playwright test, or Markdown fixture. Rows show source location,
available static metrics, run history, and a description of what the test
verifies. Dynamic declarations are flagged instead of expanded.

Optionally pass package or file paths to limit the outline. Paths are matched
relative to --cwd.`)
}

func runTestOutline(opts testOutlineOptions) (any, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return nil, err
	}
	frameworks := make([]parsers.Framework, 0, len(opts.Frameworks))
	for _, fw := range opts.Frameworks {
		framework, err := normalizeOutlineFramework(fw)
		if err != nil {
			return nil, err
		}
		frameworks = append(frameworks, framework)
	}
	return outline.Build(outline.Options{
		WorkDir:      workDir,
		Paths:        opts.Paths,
		Frameworks:   frameworks,
		FixtureFiles: opts.FixtureFiles,
		AISummary:    opts.AISummary,
		Duplication:  opts.Duplication,
		History:      opts.History,
	})
}

func normalizeOutlineFramework(name string) (parsers.Framework, error) {
	return outline.ParseFramework(name)
}

func init() {
	cmd := clicky.AddNamedCommand("outline", testCmd, testOutlineOptions{}, runTestOutline)
	cmd.Short = "Static outline of tests with size, complexity, duplication, and run history"
	cmd.Flags().SetInterspersed(true)
}
