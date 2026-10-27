package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/ai/aifix"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
)

// runCommitAIFix is the lintActionAIFix branch of handleCommitLintFindings.
// It resolves commit.fix independently from commit.message and checks the
// original scope and gates after each repair. Residual findings block the commit.
func runCommitAIFix(workDir string, result *commitpkg.Result) lintFindingsOutcome {
	if result == nil || result.Lint == nil {
		fmt.Fprintln(os.Stderr, "ai-fix: no lint result to operate on")
		return lintFindingsBlocked
	}

	if len(uniqueViolationFiles(result.Lint.Results)) == 0 {
		fmt.Fprintln(os.Stderr, "ai-fix: no files reported violations")
		return lintFindingsBlocked
	}
	files := result.Lint.Files
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "ai-fix: original commit lint scope is missing")
		return lintFindingsBlocked
	}

	requested := commitGateRequest(result.Lint.Gates)
	ctx := context.Background()

	configDir := result.Lint.ConfigDir
	if configDir == "" {
		configDir = workDir
	}
	gavelCfg, err := verify.LoadGavelConfig(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ai-fix: %v\n", err)
		return lintFindingsBlocked
	}
	saved, _, err := captainconfig.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ai-fix: %v\n", err)
		return lintFindingsBlocked
	}
	runtime := lintFixRuntime{Prompt: aifix.ResolveOptions{Name: prompts.CommitFix, Base: gavelCfg.AI, Prompt: gavelCfg.CommitFixSpec(), Dir: workDir, Linters: requested}, Runtime: defaultAIRuntimeOptions(), Saved: saved}
	resolved, err := runtime.Resolve(result.Lint.Results)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ai-fix: %v\n", err)
		return lintFindingsBlocked
	}
	for _, warning := range resolved.Resolution.Warnings {
		logger.Warnf("commit lint ai-fix: %s", warning)
	}

	renderer := newAIFixRenderer()
	maxIterations := 3
	if result.Lint.Repair.AIFixMaxIterations != nil {
		maxIterations = *result.Lint.Repair.AIFixMaxIterations
	}
	fixRes, err := aifix.Run(ctx, aifix.Request{
		MaxIterations:  maxIterations,
		Initial:        result.Lint.Results,
		AIConfig:       resolved.Config,
		AIRequestProto: resolved.Request,
		BuildRequest:   runtime.BuildRequest,
		ReLint: func(rctx context.Context) ([]*linters.LinterResult, error) {
			return runCommitLint(rctx, workDir, requested, files)
		},
		OnEvent: renderer.Handle,
	})
	if runErr := errors.Join(err, renderer.Flush()); runErr != nil {
		fmt.Fprintf(os.Stderr, "ai-fix: %v\n", runErr)
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

	return lintFindingsBlocked
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
// AI-triggered re-lint runs over the same linter set as the original
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
