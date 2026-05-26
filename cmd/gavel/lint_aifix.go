package main

import (
	"context"
	"fmt"
	"os"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/ai/aifix"
	"github.com/flanksource/gavel/linters"
)

// runAIFix invokes the AI configured by `captain configure` (overlaid by any
// `gavel lint --model=… --budget=…` flags) to repair the violations in
// allResults, then re-lints with the same scope. It loops until clean,
// MaxIterations is reached, or the configured BudgetUSD is hit.
//
// On stop reasons "max-iterations" / "max-cost" with residual violations,
// runAIFix prints a summary to stderr but does NOT itself set exitCode —
// the caller continues with the (still-non-empty) results, which the
// existing exit-code path turns into a non-zero exit.
func runAIFix(opts LintOptions, initial []*linters.LinterResult) ([]*linters.LinterResult, error) {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// Build ai.Config + ai.Request template from captain's flag/env/saved-
	// defaults overlay. ToRequest's prompt args are empty placeholders;
	// aifix fills SystemPrompt + Prompt per iteration.
	aiCfg := opts.ToConfig()
	aiProto := opts.ToRequest("", "", "")
	// Claude CLI streaming needs --verbose + --strict-mcp-config; nothing in
	// AIRuntimeOptions binds those, so seed them here. Non-claude backends
	// ignore both.
	aiProto.Verbose = true
	aiProto.StrictMCP = true

	res, err := aifix.Run(ctx, aifix.Request{
		WorkDir:        opts.WorkDir,
		Linters:        opts.Linters,
		Files:          opts.Files,
		Initial:        initial,
		MaxIterations:  opts.AIFixMaxIters,
		AIConfig:       aiCfg,
		AIRequestProto: aiProto,
		ReLint: func(rctx context.Context) ([]*linters.LinterResult, error) {
			rerunOpts := opts
			rerunOpts.Context = rctx
			rerunOpts.AIFix = false
			return executeLinters(rerunOpts)
		},
		OnEvent: aifix.NewStderrRenderer(os.Stderr),
	})
	if err != nil {
		return initial, err
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
