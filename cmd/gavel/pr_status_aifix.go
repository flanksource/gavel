package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	capverify "github.com/flanksource/captain/pkg/ai/agent/verify"
	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/ai/prfix"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/prwatch"
	"github.com/flanksource/gavel/verify"
)

// runPRStatusAIFix drives the pr.fix prompt through captain's agent.Runner: the
// agent edits the working tree, the prompt's workflow.commits policy commits and
// pushes each turn through gavel's commit pipeline, and its
// workflow.verify.commands re-poll `gavel pr status` — a non-zero exit feeds the
// output tail back as the next iteration's feedback, exit 0 stops the loop.
//
// Everything about the loop (model, budget, verify commands, iteration cap,
// commit policy) is declared in ai/prfix/pr-status-fix.prompt and overridable
// from .gavel.yaml `pr.fix`; CLI flags win over both.
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

	gavelCfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return fmt.Errorf("load .gavel.yaml: %w", err)
	}
	spec, err := prfix.ResolveSpec(gavelCfg.AI, gavelCfg.PR.Fix, workDir, prContextOf(result, statusText))
	if err != nil {
		return fmt.Errorf("resolve pr.fix prompt: %w", err)
	}
	if err := applyMaxIterations(&spec, opts.AIFixMaxIters); err != nil {
		return err
	}

	aiCfg, req, err := buildAIFixRequest(opts.AIRuntimeOptions, spec, workDir)
	if err != nil {
		return err
	}
	if aiCfg.Model.Name == "" {
		return fmt.Errorf("no model configured: pass --model or run `captain configure`")
	}
	if req.Workflow == nil || req.Workflow.Verify == nil || len(req.Workflow.Verify.Commands) == 0 {
		return fmt.Errorf("resolved pr.fix prompt declares no workflow.verify.commands: --ai-fix has no definition of done")
	}

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

	maxIters := capverify.MaxIterationsForWorkflow(req.Workflow)
	logger.Infof("pr ai-fix: invoking %s (%s), max-iter=%d, budget=$%.2f",
		aiCfg.Model.Name, aiCfg.Model.Backend, maxIters, aiCfg.Budget.Cost)

	// Commit hooks lead the list so that at PhaseRun they cut their commits
	// before any later hook acts on the result. Pushing per turn is what makes the
	// verify command meaningful: CI only re-runs on what the remote can see.
	hooks := commitpkg.AgentHooks(commitpkg.AgentHooksOptions{
		Commits: req.Workflow.Commits,
		Push:    true,
	})
	hooks = append(hooks, capverify.HooksForWorkflow(req.Workflow)...)

	runStart := time.Now()
	renderer := newAIFixRenderer()
	runner := &agent.Runner[string]{
		Provider:      streamer,
		Request:       req,
		Hooks:         hooks,
		MaxIterations: maxIters,
		Repo:          workDir,
		Cwd:           workDir,
		Scope:         capverify.ScopeForWorkflow(req.Workflow),
		OnEvent:       renderer.Handle,
	}
	res, runErr := runner.Run(ctx)
	renderErr := renderer.Flush()

	logger.Infof("pr ai-fix: stop=%s iterations=%d cost=$%.4f verified=%t",
		loopReason(res.Loop), loopIterations(res.Loop), loopCost(res.Loop), prFixVerified(res))
	if sha := lastCommitSHA(res.Response); sha != "" {
		logger.Infof("pr ai-fix: pushed %s", sha)
	}

	if err := renderCaptainHistory(runStart, loopSessionID(res.Loop)); err != nil {
		logger.Warnf("pr ai-fix: failed to render captain history: %v", err)
	}
	return errors.Join(runErr, renderErr)
}

// applyMaxIterations lets --ai-fix-max-iterations have the last word on the
// verify loop's cap; 0 keeps whatever the prompt or .gavel.yaml declared. A spec
// with no verify section cannot honour the flag at all, so that is an error
// rather than a silently ignored flag.
func applyMaxIterations(spec *api.Spec, iterations int) error {
	if iterations <= 0 {
		return nil
	}
	if spec.Workflow == nil || spec.Workflow.Verify == nil {
		return fmt.Errorf("--ai-fix-max-iterations set but the resolved pr.fix prompt declares no workflow.verify")
	}
	spec.Workflow.Verify.MaxIterations = iterations
	return nil
}

// prContextOf projects the watch result onto the prompt's template data.
func prContextOf(result *prwatch.PRWatchResult, statusText string) prfix.PRContext {
	unresolved := 0
	for _, c := range result.Comments {
		if !c.IsResolved && !c.IsOutdated {
			unresolved++
		}
	}
	return prfix.PRContext{
		Number:             result.PR.Number,
		Title:              result.PR.Title,
		URL:                result.PR.URL,
		Branch:             result.PR.HeadRefName,
		StatusText:         statusText,
		UnresolvedComments: unresolved,
	}
}

// prFixVerified reports whether the verify commands passed. Any stop reason other
// than condition-met (max iterations, budget, cost) means a check is still red.
func prFixVerified(res agent.Result[string]) bool {
	return res.Loop != nil && res.Loop.StopReason == "condition-met"
}

func loopReason(res *captainai.LoopResult) string {
	if res == nil || res.StopReason == "" {
		return "error"
	}
	return res.StopReason
}

func loopIterations(res *captainai.LoopResult) int {
	if res == nil {
		return 0
	}
	return len(res.Iterations)
}

func loopCost(res *captainai.LoopResult) float64 {
	if res == nil {
		return 0
	}
	return res.TotalCost
}

// lastCommitSHA returns the final commit the run's hooks recorded — what was
// pushed to the PR branch. Empty when no turn produced stageable changes.
func lastCommitSHA(resp *captainai.Response) string {
	if resp == nil || resp.Workspace == nil || len(resp.Workspace.Commits) == 0 {
		return ""
	}
	return resp.Workspace.Commits[len(resp.Workspace.Commits)-1].SHA
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
