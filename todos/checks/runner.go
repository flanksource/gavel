// Package checks runs the configured gavel test and lint suite after an agent
// reports done, returning a compact failure summary the caller feeds back to
// the agent. It is invoked by the todos executor's post-completion check loop.
//
// Tests run through testrunner.Run directly. Linting goes through a LintFunc
// registered by the CLI (SetLintRunner): the linter registry must be assembled
// above the linters package (it imports every adapter), which neither this
// package nor todos can do, so cmd/gavel wires it at startup.
package checks

import (
	"context"
	"fmt"
	"time"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/report"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/todos/types"
)

// LintFunc runs the configured linters in workDir and returns their results.
type LintFunc func(ctx context.Context, workDir string, cfg types.AgentLintConfig) ([]*linters.LinterResult, error)

var lintRunner LintFunc

// SetLintRunner registers the linter implementation used by Run. The CLI calls
// it once at startup (wrapping executeLinters); when unset, Run runs tests only.
func SetLintRunner(fn LintFunc) { lintRunner = fn }

// Result is the outcome of one check pass.
type Result struct {
	OK         bool   // true when nothing failed: no failing tests, no test error, no lint findings
	Summary    string // compact markdown digest of the failures, for feeding back to the agent
	Failed     int    // failing test count
	Violations int    // total lint violations
}

// Run executes the configured tests and linters in workDir and reports whether
// everything passed, plus a compact failure summary when it did not. A nil
// cfg.Test skips tests; a nil cfg.Lint (or an unregistered LintRunner) skips
// linting.
func Run(ctx context.Context, workDir string, cfg types.AgentChecksConfig) (Result, error) {
	var (
		tests   []parsers.Test
		summary parsers.TestSummary
		testErr error
	)
	if cfg.Test != nil {
		out, err := runTests(ctx, workDir, *cfg.Test, &summary)
		tests = out
		// A failed/compile-broken test run surfaces as err with the tree counts
		// in summary; that is feedback for the agent, not a hard stop.
		testErr = err
	}

	var lintResults []*linters.LinterResult
	if cfg.Lint != nil && lintRunner != nil {
		l, err := lintRunner(ctx, workDir, *cfg.Lint)
		if err != nil {
			return Result{}, fmt.Errorf("running linters: %w", err)
		}
		lintResults = l
	}

	violations, lintFailed := lintFindings(lintResults)
	res := Result{
		Failed:     summary.Failed,
		Violations: violations,
		OK:         summary.Failed == 0 && testErr == nil && !lintFailed,
	}
	if !res.OK {
		res.Summary = report.BuildCompact(tests, lintResults, report.DefaultBudget)
		if testErr != nil {
			res.Summary += fmt.Sprintf("\n### Test run error\n\n```\n%s\n```\n", testErr.Error())
		}
	}
	return res, nil
}

func runTests(ctx context.Context, workDir string, cfg types.AgentTestConfig, summary *parsers.TestSummary) ([]parsers.Test, error) {
	opts := testrunner.RunOptions{
		WorkDir:       workDir,
		StartingPaths: cfg.Paths,
		Changed:       cfg.Changed,
		Recursive:     true,
		PreBuild:      true,
		Context:       ctx,
		SummaryOut:    summary,
	}
	if cfg.Timeout != "" {
		d, err := time.ParseDuration(cfg.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid test timeout %q: %w", cfg.Timeout, err)
		}
		opts.Timeout = d
	}
	out, err := testrunner.Run(opts)
	tests, _ := out.([]parsers.Test)
	return tests, err
}

// lintFindings totals violations across results and reports whether any linter
// failed outright (errored or timed out) or surfaced violations.
func lintFindings(results []*linters.LinterResult) (violations int, failed bool) {
	for _, lr := range results {
		if lr == nil || lr.Skipped {
			continue
		}
		violations += len(lr.Violations)
		if lr.TimedOut || !lr.Success || lr.HasViolations() {
			failed = true
		}
	}
	return violations, failed
}
