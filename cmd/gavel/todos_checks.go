package main

import (
	"context"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/todos/checks"
	"github.com/flanksource/gavel/todos/types"
)

// init registers the linter implementation the post-completion check loop
// (`gavel todos run --check`) uses. The loop lives in the todos package, which
// can't assemble the linter registry (it imports every adapter), so the CLI —
// always linked into the gavel binary for both the CLI and the PR dashboard —
// wires executeLinters here at startup.
func init() {
	checks.SetLintRunner(runChecksLint)
}

// runChecksLint adapts the check loop's lint config to the shared executeLinters
// pipeline so the loop reuses the exact linter detection, execution, and
// .gavel.yaml ignore filtering as `gavel lint`.
func runChecksLint(ctx context.Context, workDir string, cfg types.AgentLintConfig) ([]*linters.LinterResult, error) {
	opts := LintOptions{
		Linters: cfg.Linters,
		Changed: cfg.Changed,
		WorkDir: workDir,
		Timeout: cfg.Timeout,
		Context: ctx,
	}
	if opts.Timeout == "" {
		opts.Timeout = "5m"
	}
	return executeLinters(opts)
}
