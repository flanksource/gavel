// Package aifix drives an AI provider (via Captain's streaming agent providers)
// to fix linter violations and re-lint
// until the result is clean, max-iterations is reached, or the cumulative
// cost cap is hit. Provider selection is determined by AIConfig.Backend.
package aifix

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/logger"
	gavelai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
)

//go:embed lint-ai-fix.prompt
var lintAIFixPrompt string

// Prompts returns the configurable lint.fix prompt descriptor.
func Prompts() []prompts.Prompt {
	return []prompts.Prompt{{
		ID:          prompts.LintFix,
		Title:       "Lint AI fix",
		Description: "Repairs violations for `gavel lint --ai-fix` and the `gavel commit` lint gate.",
		ConfigPath:  prompts.LintFix,
		Default:     lintAIFixPrompt,
		UsedBy:      []string{"gavel lint --ai-fix", "gavel commit (lint gate)"},
	}}
}

// Request is the public input to aifix.Run.
type Request struct {
	WorkDir       string
	Linters       []string
	Initial       []*linters.LinterResult
	MaxIterations int

	// AIConfig describes which provider + model + budget aifix should use.
	// Callers build it from captain's CLI flags + ~/.captain.yaml overlay
	// via captaincli.AIRuntimeOptions.ToConfig() so `gavel lint --ai-fix`
	// honours the same defaults as `captain ai prompt`. An empty Model
	// surfaces captain's "run captain configure" error.
	AIConfig captainai.Config

	// AIRequestProto is the per-iteration request template. aifix sets
	// Prompt.System and Prompt.User on a clone of this struct each turn; all
	// other fields (Permissions, Memory, Budget, Model knobs, presets, …)
	// flow through unchanged so saved captain defaults reach the provider.
	AIRequestProto captainai.Request

	// BaseAI and PromptSpec resolve lint.fix independently from commit.message.
	// The prompt is rendered again after every re-lint so each turn receives only
	// the violations that remain.
	BaseAI     api.Spec
	PromptSpec verify.PromptSpec

	// ReLint is invoked after each AI iteration to check whether
	// violations remain. It must run with the same scope (linters, files)
	// that the AI just attempted to fix.
	ReLint func(ctx context.Context) ([]*linters.LinterResult, error)

	// OnEvent (optional) is forwarded directly from the captain loop. Each
	// event carries the iteration index and the captain Event payload.
	OnEvent func(iter int, ev captainai.Event)
}

// Result summarises the outcome of an AI fix run.
type Result struct {
	FinalResults []*linters.LinterResult
	StopReason   string
	TotalCostUSD float64
	Iterations   int
}

// Run executes the AI-fix loop and returns the post-fix lint results.
func Run(ctx context.Context, req Request) (*Result, error) {
	if len(req.Initial) == 0 || !hasViolations(req.Initial) {
		return &Result{FinalResults: req.Initial, StopReason: "condition-met"}, nil
	}
	if req.ReLint == nil {
		return nil, fmt.Errorf("aifix.Run: ReLint is required")
	}

	gavelai.NormalizeEnv()

	p, err := captainai.NewProvider(req.AIConfig)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := gavelai.CloseProvider(p); err != nil {
			logger.Warnf("aifix: failed to close AI provider: %v", err)
		}
	}()
	streamer, ok := p.(captainai.StreamingProvider)
	if !ok {
		return nil, fmt.Errorf("aifix: backend %q is not streaming; choose a streaming agent backend", req.AIConfig.Model.Backend)
	}

	current := req.Initial

	// captainai.LoopOptions.BuildRequest cannot return an error, so park
	// any ReLint failure in this closure variable and surface it once the
	// loop unwinds. Returning continue=false ensures we stop fast instead
	// of asking the model to "fix" stale violations.
	var loopErr error
	loopStopReason := "loop-error"
	loopRes, err := captainai.RunUntil(ctx, captainai.LoopOptions{
		Provider:      streamer,
		MaxIterations: req.MaxIterations,
		MaxCostUSD:    req.AIConfig.Budget.Cost,
		SessionReuse:  true,
		BuildRequest: func(iter int, prev *captainai.LoopIteration) (captainai.Request, bool) {
			if iter > 0 {
				next, e := req.ReLint(ctx)
				if e != nil {
					loopErr = fmt.Errorf("re-lint between iterations failed: %w", e)
					loopStopReason = "relint-error"
					return captainai.Request{}, false
				}
				current = next
				if !hasViolations(current) {
					return captainai.Request{}, false
				}
			}
			resolved, e := ResolveSpec(req.BaseAI, req.PromptSpec, req.WorkDir, req.Linters, current)
			if e != nil {
				loopErr = fmt.Errorf("resolve lint.fix prompt: %w", e)
				loopStopReason = "prompt-error"
				return captainai.Request{}, false
			}
			turn := req.AIRequestProto
			turn.Prompt = resolved.Prompt
			return turn, true
		},
		OnEvent: req.OnEvent,
	})
	if err != nil {
		return &Result{
			FinalResults: current,
			StopReason:   loopReason(loopRes, "error"),
			TotalCostUSD: loopTotal(loopRes),
			Iterations:   loopIters(loopRes),
		}, err
	}
	if loopErr != nil {
		return &Result{
			FinalResults: current,
			StopReason:   loopStopReason,
			TotalCostUSD: loopTotal(loopRes),
			Iterations:   loopIters(loopRes),
		}, fmt.Errorf("aifix: %w", loopErr)
	}

	return &Result{
		FinalResults: current,
		StopReason:   loopRes.StopReason,
		TotalCostUSD: loopRes.TotalCost,
		Iterations:   len(loopRes.Iterations),
	}, nil
}

func loopReason(r *captainai.LoopResult, fallback string) string {
	if r != nil && r.StopReason != "" {
		return r.StopReason
	}
	return fallback
}

func loopTotal(r *captainai.LoopResult) float64 {
	if r == nil {
		return 0
	}
	return r.TotalCost
}

func loopIters(r *captainai.LoopResult) int {
	if r == nil {
		return 0
	}
	return len(r.Iterations)
}

func hasViolations(results []*linters.LinterResult) bool {
	for _, r := range results {
		if r == nil || r.Skipped {
			continue
		}
		if len(r.Violations) > 0 {
			return true
		}
	}
	return false
}

// ResolveSpec renders the configurable lint.fix operation for one iteration.
func ResolveSpec(base api.Spec, override verify.PromptSpec, workDir string, linterNames []string, results []*linters.LinterResult) (api.Spec, error) {
	return override.Resolve(base, lintAIFixPrompt, map[string]any{
		"workDir":    workDir,
		"linters":    strings.Join(linterNames, ", "),
		"violations": formatViolations(results),
	}, workDir)
}

func formatViolations(results []*linters.LinterResult) string {
	var s strings.Builder
	for _, r := range results {
		if r == nil || r.Skipped || len(r.Violations) == 0 {
			continue
		}
		for _, v := range r.Violations {
			s.WriteString(formatViolationLine(r.Linter, v))
			s.WriteString("\n")
		}
	}
	return strings.TrimSuffix(s.String(), "\n")
}

func formatViolationLine(linter string, v models.Violation) string {
	rule := ""
	if v.Rule != nil {
		rule = v.Rule.Method
	}
	msg := ""
	if v.Message != nil {
		msg = *v.Message
	}
	loc := v.File
	if v.Line > 0 {
		loc = fmt.Sprintf("%s:%d", loc, v.Line)
	}
	if rule != "" {
		return fmt.Sprintf("  %s [%s/%s] %s", loc, linter, rule, msg)
	}
	return fmt.Sprintf("  %s [%s] %s", loc, linter, msg)
}
