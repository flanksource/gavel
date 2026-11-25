package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/flanksource/captain/pkg/captainconfig"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/ai/aifix"
	"github.com/flanksource/gavel/linters"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/verify"
)

// runAIFix resolves lint.fix from Gavel config (overlaid by opts.Runtime) to
// repair the violations in initial, then calls opts.ReLint over the same scope.
// It loops until clean, MaxIterations is reached, or the configured budget is
// hit.
//
// On stop reasons "max-iterations" / "max-cost" with residual violations,
// runAIFix prints a summary to stderr but does NOT itself set exitCode —
// the caller continues with the (still-non-empty) results, which the
// existing exit-code path turns into a non-zero exit.
type aiFixOptions struct {
	Context       context.Context
	WorkDir       string
	Linters       []string
	Runtime       captaincli.AIRuntimeOptions
	MaxIterations int
	ReLint        func(context.Context) ([]*linters.LinterResult, error)
}

// runScopedAIFix repairs only the open findings req names in current. Each
// fix→re-lint cycle re-lints the run's whole scope, so the returned results
// replace current without losing other findings, but the loop only sees, and
// stops on, the requested fingerprints: ignored, dismissed and unselected
// findings are never sent to the fixer.
func runScopedAIFix(ctx context.Context, opts LintOptions, req testui.LintFixRequest, output *testui.RerunOutputBuffer, current []*linters.LinterResult) ([]*linters.LinterResult, *testui.LintFixRecord, error) {
	selected := make(map[string]bool, len(req.Fingerprints))
	for _, fp := range req.Fingerprints {
		selected[fp] = true
	}
	initial := linters.SelectViolations(current, selected)
	if len(initial) == 0 {
		return nil, nil, fmt.Errorf("none of the %d requested findings is open in this run", len(req.Fingerprints))
	}
	var names []string
	for _, result := range initial {
		names = append(names, result.Linter)
	}
	runtime := opts.AIRuntimeOptions
	if req.Model != "" {
		runtime.Model = req.Model
	}
	progress := output.StdoutWriter()
	var latest []*linters.LinterResult
	cycle := 0
	outcome, err := runAIFix(aiFixOptions{
		Context: ctx, WorkDir: opts.WorkDir, Linters: names, Runtime: runtime,
		MaxIterations: cmp.Or(req.MaxIterations, opts.AIFixMaxIters),
		ReLint: func(rctx context.Context) ([]*linters.LinterResult, error) {
			cycle++
			rerun := opts
			rerun.Context = rctx
			rerun.OutputTee = progress
			full, err := executeLintRerun(rerun, testui.RerunRequest{})
			if err != nil {
				return nil, err
			}
			latest = full
			remaining := linters.SelectViolations(full, selected)
			fmt.Fprintf(progress, "ai-fix cycle %d: %d of %d requested finding(s) remain\n", cycle, violationCount(remaining), len(selected))
			return remaining, nil
		},
	}, initial)
	if err != nil {
		return nil, nil, err
	}
	if latest == nil {
		return nil, nil, fmt.Errorf("ai-fix stopped (%s) before re-linting, so its result is unknown", outcome.StopReason)
	}
	record := &testui.LintFixRecord{
		Fingerprints: req.Fingerprints, Model: outcome.Model, Iterations: outcome.Iterations,
		StopReason: outcome.StopReason, CostUSD: outcome.TotalCostUSD,
	}
	for _, result := range linters.SelectViolations(latest, selected) {
		for _, v := range result.Violations {
			record.Remaining = append(record.Remaining, v.Fingerprint)
		}
	}
	return latest, record, nil
}

func violationCount(results []*linters.LinterResult) int {
	count := 0
	for _, result := range results {
		count += len(result.Violations)
	}
	return count
}

// aiFixOutcome is a finished fix loop and the model that ran it.
type aiFixOutcome struct {
	*aifix.Result
	Model string
}

func runAIFix(opts aiFixOptions, initial []*linters.LinterResult) (aiFixOutcome, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}
	unchanged := aiFixOutcome{Result: &aifix.Result{FinalResults: initial}}

	gavelCfg, err := verify.LoadGavelConfig(opts.WorkDir)
	if err != nil {
		return unchanged, err
	}
	saved, _, err := captainconfig.Load()
	if err != nil {
		return unchanged, err
	}
	runtime := lintFixRuntime{Prompt: aifix.ResolveOptions{Base: gavelCfg.AI, Prompt: gavelCfg.Lint.Fix, Dir: opts.WorkDir, Linters: opts.Linters}, Runtime: opts.Runtime, Saved: saved}
	resolved, err := runtime.Resolve(initial)
	if err != nil {
		return unchanged, err
	}
	for _, warning := range resolved.Resolution.Warnings {
		logger.Warnf("lint ai-fix: %s", warning)
	}

	renderer := newAIFixRenderer()
	res, err := aifix.Run(ctx, aifix.Request{
		Initial:        initial,
		MaxIterations:  opts.MaxIterations,
		AIConfig:       resolved.Config,
		AIRequestProto: resolved.Request,
		BuildRequest:   runtime.BuildRequest,
		ReLint:         opts.ReLint,
		OnEvent:        renderer.Handle,
	})
	if runErr := errors.Join(err, renderer.Flush()); runErr != nil {
		return unchanged, runErr
	}

	logger.Infof("ai-fix: stop=%s iterations=%d cost=$%.4f",
		res.StopReason, res.Iterations, res.TotalCostUSD)

	if res.StopReason != "condition-met" {
		residual := 0
		for _, lr := range res.FinalResults {
			if lr == nil || lr.Skipped {
				continue
			}
			residual += len(lr.Violations)
		}
		if residual > 0 {
			fmt.Fprintf(os.Stderr,
				"ai-fix: stopped with %d residual violation(s) (reason=%s)\n",
				residual, res.StopReason)
		}
	}
	return aiFixOutcome{Result: res, Model: resolved.Config.Model.Name}, nil
}
