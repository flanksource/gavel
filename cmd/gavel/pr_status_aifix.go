package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/prwatch"
)

// runPRStatusAIFix feeds the rendered status into the AI configured by
// `captain configure` (overlaid by any --model/--budget/--backend flags).
// Unlike `lint --ai-fix`, this is a single-shot prompt by default — the
// status snapshot is captured once and handed to the model, which is
// expected to edit files in the local git working tree. We don't re-poll
// GitHub between iterations because new check results take minutes to
// produce; users wanting that loop should pass --ai-fix-max-iterations > 1
// and accept the wall-clock cost.
func runPRStatusAIFix(ctx context.Context, opts PRStatusOptions, result *prwatch.PRWatchResult) error {
	if result == nil || result.PR == nil {
		return fmt.Errorf("no PR result available")
	}

	statusText := result.Pretty().ANSI()
	if strings.TrimSpace(statusText) == "" {
		return fmt.Errorf("rendered status was empty")
	}

	workDir, err := getWorkingDir()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}
	aiCfg, aiProto, err := buildAIFixRequest(opts.AIRuntimeOptions, api.Spec{}, workDir)
	if err != nil {
		return err
	}
	if aiCfg.Model.Name == "" {
		return fmt.Errorf("no model configured: pass --model or run `captain configure`")
	}

	prompt := buildPRStatusPrompt(result, statusText)
	systemPrompt := buildPRStatusSystemPrompt(result)

	p, err := captainai.NewProvider(aiCfg)
	if err != nil {
		return err
	}
	defer func() {
		if err := ai.CloseProvider(p); err != nil {
			logger.Warnf("pr ai-fix: failed to close AI provider: %v", err)
		}
	}()
	streamer, ok := p.(captainai.StreamingProvider)
	if !ok {
		return fmt.Errorf("backend %q is not streaming; choose a streaming backend (claude-cli, codex-cli, gemini-cli)", aiCfg.Model.Backend)
	}

	maxIters := opts.AIFixMaxIters
	if maxIters <= 0 {
		maxIters = 1
	}

	logger.Infof("pr ai-fix: invoking %s (%s), max-iter=%d, budget=$%.2f",
		aiCfg.Model.Name, aiCfg.Model.Backend, maxIters, aiCfg.Budget.Cost)

	runStart := time.Now()
	loopRes, err := captainai.RunUntil(ctx, captainai.LoopOptions{
		Provider:      streamer,
		MaxIterations: maxIters,
		MaxCostUSD:    aiCfg.Budget.Cost,
		SessionReuse:  true,
		BuildRequest: func(iter int, prev *captainai.LoopIteration) (captainai.Request, bool) {
			if iter > 0 {
				return captainai.Request{}, false
			}
			turn := aiProto
			turn.Prompt.System = systemPrompt
			turn.Prompt.User = prompt
			return turn, true
		},
		OnEvent: newAIFixRenderer(),
	})
	if err != nil {
		return err
	}

	logger.Infof("pr ai-fix: stop=%s iterations=%d cost=$%.4f",
		loopRes.StopReason, len(loopRes.Iterations), loopRes.TotalCost)

	if err := renderCaptainHistory(runStart, loopSessionID(loopRes)); err != nil {
		logger.Warnf("pr ai-fix: failed to render captain history: %v", err)
	}
	return nil
}

// loopSessionID returns the session the run actually used — the last iteration
// that reported one. Iterations before the provider emits its session-init
// event carry an empty id, so the last non-empty wins rather than the last.
func loopSessionID(res *captainai.LoopResult) string {
	if res == nil {
		return ""
	}
	sessionID := ""
	for _, iter := range res.Iterations {
		if iter.SessionID != "" {
			sessionID = iter.SessionID
		}
	}
	return sessionID
}

// historyOptionsForRun builds the captain history query for a finished ai-fix
// run. Identifying the run by its own session id is the whole point: --last
// resolves via a trailing-run-of-tools heuristic, which happily selects an
// unrelated session another agent wrote to more recently (the reported case
// rendered session 929f3d1b for a run whose session was 24eec7df). SessionID
// already narrows to exactly one session, so Last would only add clipping on
// model/effort changes within it.
//
// A backend that reports no session id falls back to the old --last behaviour;
// history rendering is best-effort trailing output, not part of the fix.
func historyOptionsForRun(runStart time.Time, sessionID string) captaincli.HistoryOptions {
	opts := captaincli.HistoryOptions{
		Since: runStart.Add(-2 * time.Second),
		Limit: 0,
	}
	if sessionID == "" {
		opts.Last = true
		return opts
	}
	opts.SessionID = sessionID
	return opts
}

// renderCaptainHistory invokes captain's RunHistory and writes the result to
// stdout the same way `captain history --last` does — captain prints
// line-by-line when stdout is a TTY and emits a structured table otherwise.
func renderCaptainHistory(runStart time.Time, sessionID string) error {
	if _, err := database.Shared(context.Background()); err != nil {
		return fmt.Errorf("prepare Captain database: %w", err)
	}
	result, err := captaincli.RunHistory(historyOptionsForRun(runStart, sessionID))
	if err != nil {
		return err
	}
	if result == nil {
		// renderLineByLine already wrote to stdout.
		return nil
	}
	clicky.MustPrint(result, clicky.FormatOptions{})
	return nil
}

func buildPRStatusSystemPrompt(result *prwatch.PRWatchResult) string {
	var s strings.Builder
	s.WriteString("You are running inside a developer's git working tree.")
	if result.PR != nil {
		fmt.Fprintf(&s, " The current PR is #%d (%q).", result.PR.Number, result.PR.Title)
		if result.PR.HeadRefName != "" {
			fmt.Fprintf(&s, " HEAD branch: %s.", result.PR.HeadRefName)
		}
	}
	s.WriteString(" The user will paste the rendered output of `gavel pr status` below.")
	s.WriteString(" Read it, identify failing GitHub Actions jobs and unresolved review comments,")
	s.WriteString(" and edit files in place to fix them. Prefer minimal, targeted edits over rewrites.")
	s.WriteString(" Do not run any commit-related commands; the user will commit after verification.")
	return s.String()
}

func buildPRStatusPrompt(result *prwatch.PRWatchResult, statusText string) string {
	var s strings.Builder
	s.WriteString("Fix the failures and unresolved comments visible in this PR status snapshot:\n\n")
	s.WriteString("```\n")
	s.WriteString(statusText)
	if !strings.HasSuffix(statusText, "\n") {
		s.WriteString("\n")
	}
	s.WriteString("```\n")
	if result.PR != nil && result.PR.URL != "" {
		fmt.Fprintf(&s, "\nFor reference, the PR URL is %s.\n", result.PR.URL)
	}
	return s.String()
}
