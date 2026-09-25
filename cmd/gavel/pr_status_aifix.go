package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	capsetup "github.com/flanksource/captain/pkg/ai/agent/setup"
	capverify "github.com/flanksource/captain/pkg/ai/agent/verify"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/clicky"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/ai/prfix"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/prwatch"
	"github.com/flanksource/gavel/utils"
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
//
// With --worktree the run's spec gains a setup layer and captain's setup plugin
// moves every turn into a fresh worktree at the PR head. A PR that conflicts
// with its base has that merge started in the run's tree before the first turn,
// so the agent resolves markers rather than deciding how to merge.
func runPRStatusAIFix(ctx context.Context, opts PRStatusOptions, ghOpts github.Options, result *prwatch.PRWatchResult) error {
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
	// Repo is the repo root even when invoked from a subdirectory, while Cwd
	// stays where the user ran the command: the runner records the agent's edits
	// relative to Repo, and they only line up with the dirty set the commit hooks
	// read out of git — which git anchors on the root however deep it is
	// invoked — when Repo is that same root.
	repoRoot := utils.GitRoot(workDir)

	gavelCfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return fmt.Errorf("load .gavel.yaml: %w", err)
	}
	layerOpts := prFixLayerOptions{
		Resolve:    prfix.ResolveOptions{Base: gavelCfg.AI, Prompt: gavelCfg.PR.Fix, Dir: workDir, PR: prContextOf(result, statusText)},
		Iterations: opts.AIFixMaxIters,
	}
	if opts.Worktree {
		if err := github.EnsurePRCommits(github.Options{WorkDir: repoRoot, Repo: ghOpts.Repo, Token: ghOpts.Token}, result.PR); err != nil {
			return fmt.Errorf("--worktree: %w", err)
		}
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return fmt.Errorf("--worktree: resolve the user cache directory: %w", err)
		}
		layerOpts.Setup = prFixWorktreeSetup(prFixWorktree{RepoRoot: repoRoot, PR: result.PR, CacheDir: cacheDir})
	}
	layers, err := prFixLayers(layerOpts)
	if err != nil {
		return fmt.Errorf("resolve pr.fix prompt: %w", err)
	}
	saved, _, err := captainconfig.Load()
	if err != nil {
		return err
	}
	resolved, err := buildAIFixRequest(aiFixRequestOptions{Runtime: opts.AIRuntimeOptions, Layers: layers, Saved: saved, Dir: workDir})
	if err != nil {
		return err
	}
	for _, warning := range resolved.Resolution.Warnings {
		logger.Warnf("pr ai-fix: %s", warning)
	}
	aiCfg, req := resolved.Config, resolved.Request
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
		return fmt.Errorf("runtime %q is not streaming; choose a cli, agent or cmux runtime", p.GetRuntime())
	}

	maxIters := capverify.MaxIterationsForWorkflow(req.Workflow)
	logger.Infof("pr ai-fix: invoking %s (%s), max-iter=%d, budget=$%.2f",
		aiCfg.Model.Name, p.GetRuntime(), maxIters, aiCfg.Budget.Cost)

	tee := newAIFixVerifyTee()
	defer func() {
		if err := tee.Close(); err != nil {
			logger.Warnf("pr ai-fix: verify output tee: %v", err)
		}
	}()

	hooks, err := prFixHooks(ctx, prFixRun{Workflow: req.Workflow, Provider: p, Tee: tee, RepoRoot: repoRoot, Result: result})
	if err != nil {
		return err
	}

	runStart := time.Now()
	renderer := newAIFixRenderer()
	runner := &agent.Runner[string]{
		Provider:      streamer,
		Request:       req,
		Hooks:         hooks,
		MaxIterations: maxIters,
		Repo:          repoRoot,
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

// prFixRun is what prFixHooks wires the run's hooks from.
type prFixRun struct {
	Workflow *api.Workflow
	Provider captainai.Provider
	Tee      io.Writer
	RepoRoot string
	Result   *prwatch.PRWatchResult
}

// prFixHooks assembles the run's hooks, in an order each later hook relies on:
//  1. captain's setup plugin, which realises the spec's `setup:` (--worktree, or
//     one declared in the prompt/override) and moves the run into it;
//  2. the conflict merge, started in whatever tree setup left the run in;
//  3. the commit pipeline, which concludes that merge once it is resolved and
//     pushes each turn to the PR branch — pushing per turn is what makes the
//     verify command meaningful, since CI only re-runs on what the remote sees;
//  4. the workflow's checks, teed to the terminal so a multi-minute check is
//     visibly moving.
func prFixHooks(ctx context.Context, run prFixRun) ([]any, error) {
	hooks := []any{&capsetup.Plugin{BaseDir: run.RepoRoot}}
	if report := prConflicts(run.Result); report != nil {
		hooks = append(hooks, &prFixMergeHook{report: report})
	}
	hooks = append(hooks, commitpkg.AgentHooks(commitpkg.AgentHooksOptions{
		Commits:    run.Workflow.Commits,
		Push:       true,
		PushBranch: run.Result.PR.HeadRefName,
	})...)
	verifyHooks, err := capverify.HooksFor(ctx, run.Workflow, capverify.Options{Provider: run.Provider, Output: run.Tee})
	if err != nil {
		return nil, err
	}
	return append(hooks, verifyHooks...), nil
}

type prFixLayerOptions struct {
	Resolve prfix.ResolveOptions
	// Iterations is --ai-fix-max-iterations; 0 keeps the prompt's cap.
	Iterations int
	// Setup is --worktree's isolation, layered over whatever the prompt declares.
	Setup *shell.Setup
}

func prFixLayers(options prFixLayerOptions) ([]api.SpecLayer, error) {
	layers, err := prfix.Layers(options.Resolve)
	if err != nil {
		return nil, err
	}
	if options.Setup != nil {
		layers = append(layers, api.RequestSpecLayer("--worktree", api.Spec{Setup: options.Setup}))
	}
	if options.Iterations <= 0 {
		return layers, nil
	}
	composed, err := api.ComposeSpecLayers(api.ResolveSpecOptions{Layers: layers})
	if err != nil {
		return nil, err
	}
	if err := applyMaxIterations(&composed.Spec, options.Iterations); err != nil {
		return nil, err
	}
	return append(layers, api.RequestSpecLayer("--ai-fix-max-iterations", api.Spec{Workflow: &api.Workflow{Verify: &api.Verify{MaxIterations: options.Iterations}}})), nil
}

type prFixWorktree struct {
	RepoRoot string
	PR       *github.PRInfo
	// CacheDir is the user cache directory the worktree is created under.
	CacheDir string
}

// prFixWorktreeSetup describes the --worktree isolation as captain setup: a new
// worktree of the local checkout, branched from the PR head commit rather than
// from whatever the caller has checked out, and placed under the user cache
// directory so the repository never gains a nested checkout. Uncommitted work
// stays behind — it belongs to the caller's branch, not the PR — while ignored
// content (installed dependencies, build output) is copied so the agent can
// build and test.
func prFixWorktreeSetup(w prFixWorktree) *shell.Setup {
	name := fmt.Sprintf("pr-%d-%d", w.PR.Number, time.Now().Unix())
	return &shell.Setup{Checkout: &shell.Checkout{
		Path: w.RepoRoot,
		Worktree: &shell.Worktree{
			Mode:        shell.WorktreeNew,
			Prefix:      fmt.Sprintf("gavel/pr-%d-fix", w.PR.Number),
			Base:        w.PR.HeadRefOID,
			Path:        filepath.Join(w.CacheDir, "gavel", "worktrees", filepath.Base(w.RepoRoot), name),
			Uncommitted: shell.CloneSkip,
			Ignored:     shell.CloneClone,
		},
	}}
}

const prFixMergeHookName = "pr-fix-merge"

// prFixMergeHook starts merging the PR's base into the run's tree before the
// first turn, leaving the conflicted paths for the agent. It is a PreRun hook,
// registered after captain's setup plugin, so under --worktree the merge happens
// in the worktree and the caller's checkout is never touched.
type prFixMergeHook struct {
	report *github.MergeConflictReport
}

func (h *prFixMergeHook) Name() string { return prFixMergeHookName }

func (h *prFixMergeHook) PreRun(hc *agent.HookContext) error {
	dir := hc.Workspace().Cwd
	files, err := github.BeginConflictMerge(dir, h.report)
	if err != nil {
		return err
	}
	paths := make([]string, 0, len(files))
	for _, f := range files {
		paths = append(paths, f.Path)
	}
	hc.Notify("[pre-run] merging origin/%s into %s: %d conflicted — %s",
		h.report.BaseRefName, h.report.HeadRefName, len(files), strings.Join(paths, ", "))
	return nil
}

// prConflicts returns the watch result's conflict report when it lists files to
// resolve. A CONFLICTING verdict git could not reproduce has none, so there is
// no merge to start.
func prConflicts(result *prwatch.PRWatchResult) *github.MergeConflictReport {
	if result == nil || result.Conflicts == nil || len(result.Conflicts.Files) == 0 {
		return nil
	}
	return result.Conflicts
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
	pr := prfix.PRContext{
		Number:             result.PR.Number,
		Title:              result.PR.Title,
		URL:                result.PR.URL,
		Branch:             result.PR.HeadRefName,
		BaseBranch:         result.PR.BaseRefName,
		StatusText:         statusText,
		UnresolvedComments: result.UnresolvedComments(),
	}
	if report := prConflicts(result); report != nil {
		for _, f := range report.Files {
			pr.Conflicts = append(pr.Conflicts, prfix.Conflict{Path: f.Path, Kind: f.Kind})
		}
	}
	return pr
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
