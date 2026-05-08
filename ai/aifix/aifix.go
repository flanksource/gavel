// Package aifix drives Claude (via captain's claude_cli provider) to fix
// linter violations and re-lint until the result is clean, max-iterations is
// reached, or the cumulative cost cap is hit.
package aifix

import (
	"context"
	"fmt"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/provider"
	gavelai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
)

// Request is the public input to aifix.Run.
type Request struct {
	WorkDir       string
	Linters       []string
	Files         []string
	Initial       []*linters.LinterResult
	Model         string
	MaxIterations int
	MaxCostUSD    float64

	// ReLint is invoked after each Claude iteration to check whether
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

const defaultModel = "claude-code-opus"

// Run executes the AI-fix loop and returns the post-fix lint results.
func Run(ctx context.Context, req Request) (*Result, error) {
	if len(req.Initial) == 0 || !hasViolations(req.Initial) {
		return &Result{FinalResults: req.Initial, StopReason: "condition-met"}, nil
	}
	if req.ReLint == nil {
		return nil, fmt.Errorf("aifix.Run: ReLint is required")
	}

	gavelai.NormalizeEnv()

	model := req.Model
	if model == "" {
		model = defaultModel
	}

	cli := provider.NewClaudeCLI(model)

	current := req.Initial
	systemPrompt := buildSystemPrompt(req.WorkDir, req.Linters)

	loopRes, err := captainai.RunUntil(ctx, captainai.LoopOptions{
		Provider:      cli,
		MaxIterations: req.MaxIterations,
		MaxCostUSD:    req.MaxCostUSD,
		SessionReuse:  true,
		BuildRequest: func(iter int, prev *captainai.LoopIteration) (captainai.Request, bool) {
			if iter > 0 {
				next, relErr := req.ReLint(ctx)
				if relErr != nil {
					// Surface as a continue=false (condition-met) so the
					// loop returns; the caller compares len(violations)
					// in current to detect partial success vs failure.
					return captainai.Request{}, false
				}
				current = next
				if !hasViolations(current) {
					return captainai.Request{}, false
				}
			}
			return captainai.Request{
				SystemPrompt:   systemPrompt,
				Prompt:         buildPrompt(req.WorkDir, current),
				PermissionMode: "acceptEdits",
				StrictMCP:      true,
				Verbose:        true,
			}, true
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

// buildSystemPrompt sets the framing for every iteration. We keep it short
// because the per-turn prompt already contains the violation list — the
// system prompt only carries durable context.
func buildSystemPrompt(workDir string, linterNames []string) string {
	var s strings.Builder
	s.WriteString("You are running inside a developer's git working tree at ")
	s.WriteString(workDir)
	s.WriteString(". Fix the linter violations the user lists, then stop. ")
	s.WriteString("Edit files in place. Do not run any commit-related commands. ")
	s.WriteString("Prefer minimal, targeted edits over rewrites. ")
	if len(linterNames) > 0 {
		s.WriteString("The active linters are: ")
		s.WriteString(strings.Join(linterNames, ", "))
		s.WriteString(". ")
	}
	s.WriteString("After your edits, the user will re-run the linters to verify.")
	return s.String()
}

func buildPrompt(workDir string, results []*linters.LinterResult) string {
	var s strings.Builder
	s.WriteString("Fix the following linter violations:\n\n")
	for _, r := range results {
		if r == nil || r.Skipped || len(r.Violations) == 0 {
			continue
		}
		for _, v := range r.Violations {
			s.WriteString(formatViolationLine(r.Linter, v))
			s.WriteString("\n")
		}
	}
	return s.String()
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
