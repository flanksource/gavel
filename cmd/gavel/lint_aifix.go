package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/flanksource/captain/pkg/captainconfig"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/ai/aifix"
	"github.com/flanksource/gavel/linters"
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

func runAIFix(opts aiFixOptions, initial []*linters.LinterResult) ([]*linters.LinterResult, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	gavelCfg, err := verify.LoadGavelConfig(opts.WorkDir)
	if err != nil {
		return initial, err
	}
	saved, _, err := captainconfig.Load()
	if err != nil {
		return initial, err
	}
	runtime := lintFixRuntime{Prompt: aifix.ResolveOptions{Base: gavelCfg.AI, Prompt: gavelCfg.Lint.Fix, Dir: opts.WorkDir, Linters: opts.Linters}, Runtime: opts.Runtime, Saved: saved}
	resolved, err := runtime.Resolve(initial)
	if err != nil {
		return initial, err
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
		return initial, runErr
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
	return res.FinalResults, nil
}
