package main

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/testrunner/outline"
	"github.com/flanksource/gavel/testrunner/parsers"
)

type testOutlineOptions struct {
	Paths       []string `json:"paths,omitempty" args:"true"`
	AISummary   bool     `json:"ai_summary,omitempty" flag:"ai-summary" help:"Generate one-line AI summaries of what each test verifies"`
	Frameworks  []string `json:"frameworks,omitempty" flag:"framework" help:"Limit to frameworks: gotest, ginkgo, vitest"`
	Duplication bool     `json:"duplication,omitempty" flag:"duplication" default:"true" help:"Compute per-test duplication via jscpd"`
	History     bool     `json:"history,omitempty" flag:"history" default:"true" help:"Join run history from .gavel snapshots"`
}

func (o testOutlineOptions) Help() string {
	return `Statically outline tests without running them.

Renders a directory/file tree with one row per test function, ginkgo spec, or
vitest test showing location, body size, cyclomatic complexity, duplication
(via jscpd), run history (from .gavel/run-*.json snapshots), and a one-line
description of what the test verifies. Dynamic specs (non-literal ginkgo or
t.Run names) are flagged instead of expanded.

Optionally pass package or file paths to limit the outline. Paths are matched
relative to --cwd.`
}

func runTestOutline(opts testOutlineOptions) (any, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return nil, err
	}
	frameworks := make([]parsers.Framework, 0, len(opts.Frameworks))
	for _, fw := range opts.Frameworks {
		frameworks = append(frameworks, normalizeOutlineFramework(fw))
	}
	return outline.Build(outline.Options{
		WorkDir:     workDir,
		Paths:       opts.Paths,
		Frameworks:  frameworks,
		AISummary:   opts.AISummary,
		Duplication: opts.Duplication,
		History:     opts.History,
	})
}

// normalizeOutlineFramework maps CLI spellings ("gotest", "go-test") onto
// parsers.GoTest ("go test"); unknown values pass through so outline.Build
// rejects them with a clear error.
func normalizeOutlineFramework(name string) parsers.Framework {
	switch name {
	case "gotest", "go-test", "go":
		return parsers.GoTest
	default:
		return parsers.Framework(name)
	}
}

func init() {
	cmd := clicky.AddNamedCommand("outline", testCmd, testOutlineOptions{}, runTestOutline)
	cmd.Short = "Static outline of tests with size, complexity, duplication, and run history"
	cmd.Flags().SetInterspersed(true)
}
