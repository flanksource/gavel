package main

import (
	"context"
	"fmt"
	"os"

	"github.com/flanksource/gavel/ai/aifix"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/linters"
)

// runCommitAIFix is the lintActionAIFix branch of handleCommitLintFindings.
// It invokes Claude to repair the findings in result.Lint and re-runs the
// commit lint pass with the SAME gate configuration the original commit
// used. On clean it returns lintFindingsAIFixed; on residual violations it
// re-prompts the user (so they can pick AI Fix again, triage, bypass, or
// cancel).
func runCommitAIFix(workDir string, result *commitpkg.Result) lintFindingsOutcome {
	if result == nil || result.Lint == nil {
		fmt.Fprintln(os.Stderr, "ai-fix: no lint result to operate on")
		return lintFindingsBlocked
	}

	files := uniqueViolationFiles(result.Lint.Results)
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "ai-fix: no files reported violations")
		return lintFindingsBlocked
	}

	requested := commitGateRequest(result.Lint.Gates)
	ctx := context.Background()

	fixRes, err := aifix.Run(ctx, aifix.Request{
		WorkDir: workDir,
		Files:   files,
		Initial: result.Lint.Results,
		ReLint: func(rctx context.Context) ([]*linters.LinterResult, error) {
			return runCommitLint(rctx, workDir, requested, files)
		},
		OnEvent: aifix.NewStderrRenderer(os.Stderr),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ai-fix: %v\n", err)
		return lintFindingsBlocked
	}

	fmt.Fprintf(os.Stderr, "ai-fix: stop=%s iterations=%d cost=$%.4f\n",
		fixRes.StopReason, fixRes.Iterations, fixRes.TotalCostUSD)

	residual := countViolations(fixRes.FinalResults)
	if residual == 0 {
		return lintFindingsAIFixed
	}

	fmt.Fprintf(os.Stderr, "ai-fix: %d residual violation(s) remain\n", residual)
	for _, lr := range fixRes.FinalResults {
		if lr == nil || lr.Skipped {
			continue
		}
		for _, v := range lr.Violations {
			fmt.Fprintln(os.Stderr, formatCommitLintViolation(lr.Linter, v))
		}
	}

	stub := &commitpkg.Result{Lint: &commitpkg.LintGateResult{
		Results:    fixRes.FinalResults,
		Violations: residual,
		Gates:      result.Lint.Gates,
	}}
	return handleCommitLintFindings(workDir, stub)
}

func uniqueViolationFiles(results []*linters.LinterResult) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, r := range results {
		if r == nil || r.Skipped {
			continue
		}
		for _, v := range r.Violations {
			if v.File == "" {
				continue
			}
			if _, dup := seen[v.File]; dup {
				continue
			}
			seen[v.File] = struct{}{}
			out = append(out, v.File)
		}
	}
	return out
}

func countViolations(results []*linters.LinterResult) int {
	n := 0
	for _, r := range results {
		if r == nil || r.Skipped {
			continue
		}
		n += len(r.Violations)
	}
	return n
}

// commitGateRequest mirrors applyLintGate's mapping of gate booleans onto
// the runCommitLint linterNames argument. Keeping this in sync ensures the
// re-lint Claude triggered runs over the same linter set as the original
// commit gate.
func commitGateRequest(gates commitpkg.LintGates) []string {
	switch {
	case gates.FullLint && gates.Secrets:
		return nil
	case gates.FullLint:
		return []string{"!betterleaks"}
	case gates.Secrets:
		return []string{"betterleaks"}
	default:
		return nil
	}
}
