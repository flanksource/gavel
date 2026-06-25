package todos

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/todos/checks"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/repomap"
)

// runCheckLoop is the post-completion check loop: once the agent reports done
// (and any fixture Verification passed), it runs the configured gavel test/lint
// suite and, while that suite fails, feeds a compact failure summary back to the
// agent and re-runs the suite — up to MaxIterations feedback rounds. It mutates
// result: a clean suite leaves it untouched; a still-failing suite (or a feedback
// round that errors) flips result.Success to false with an explanatory message.
//
// The loop is opt-in: it runs only when the resolved AgentChecksConfig is
// enabled (via .gavel.yaml `checks`, a TODO's frontmatter `checks`, or the
// --check flag / dashboard toggle that sets forceChecks).
func (e *TODOExecutor) runCheckLoop(ctx *ExecutorContext, todosInGroup []*types.TODO, result *ExecutionResult) {
	if result == nil || !result.Success {
		return // only gate a successful agent run
	}

	gitRoot := e.checksWorkDir(todosInGroup)
	project, err := verify.LoadGavelConfig(gitRoot)
	if err != nil {
		ctx.Logger.Warnf("checks: failed to load .gavel.yaml: %v", err)
	}
	cfg := types.ResolveAgentChecks(project.Checks, firstChecksConfig(todosInGroup), e.forceChecks)
	if !cfg.IsEnabled() {
		return
	}

	for attempt := 0; attempt <= cfg.MaxIterations; attempt++ {
		ctx.Notify(Notification{Type: NotifyProgress, Message: "Running post-completion checks"})
		cr, err := checks.Run(ctx, gitRoot, cfg)
		if err != nil {
			ctx.Logger.Errorf("checks: run errored: %v", err)
			result.Success = false
			result.ErrorMessage = fmt.Sprintf("checks errored: %v", err)
			return
		}
		if cr.OK {
			ctx.Logger.Infof("checks: passed (%d test failures, %d violations)", cr.Failed, cr.Violations)
			return
		}
		ctx.Logger.Warnf("checks: %d failing tests, %d lint violations (attempt %d/%d)", cr.Failed, cr.Violations, attempt+1, cfg.MaxIterations+1)

		if attempt == cfg.MaxIterations {
			result.Success = false
			result.ErrorMessage = fmt.Sprintf("checks still failing after %d feedback iterations (%d test failures, %d violations)", cfg.MaxIterations, cr.Failed, cr.Violations)
			return
		}

		fb, ok := e.executor.(FeedbackExecutor)
		if !ok {
			ctx.Logger.Warnf("checks: executor %s cannot feed back; reporting failures", e.executor.Name())
			result.Success = false
			result.ErrorMessage = "checks failing and executor cannot feed results back to the agent"
			return
		}

		ctx.Notify(Notification{Type: NotifyProgress, Message: fmt.Sprintf("Re-running agent with check failures (iteration %d/%d)", attempt+1, cfg.MaxIterations)})
		fbResult, err := fb.SendFeedback(ctx, todosInGroup, cr.Summary)
		if err != nil {
			ctx.Logger.Errorf("checks: feedback round failed: %v", err)
			result.Success = false
			result.ErrorMessage = fmt.Sprintf("feedback round failed: %v", err)
			return
		}
		// An inline executor commits each feedback round; carry the latest SHA so
		// the caller's post-run commit step doesn't double-commit (see ShouldCommitAfter).
		if fbResult != nil && fbResult.CommitSHA != "" {
			result.CommitSHA = fbResult.CommitSHA
		}
	}
}

// checksWorkDir resolves the directory the checks run in: the git root of the
// agent's working directory (workDir joined with the group's cwd), mirroring how
// commit.RunAfterAgent derives its commit directory.
func (e *TODOExecutor) checksWorkDir(todosInGroup []*types.TODO) string {
	dir := e.workDir
	for _, todo := range todosInGroup {
		if todo == nil || todo.CWD == "" {
			continue
		}
		if filepath.IsAbs(todo.CWD) {
			dir = filepath.Clean(todo.CWD)
		} else if e.workDir != "" {
			dir = filepath.Clean(filepath.Join(e.workDir, todo.CWD))
		} else {
			dir = filepath.Clean(todo.CWD)
		}
		break
	}
	if root := repomap.FindGitRoot(dir); root != "" {
		return root
	}
	return dir
}

// allGroupResultsOK reports whether every todo that needed work has a
// successful result — the precondition for running the group check loop.
func (e *TODOExecutor) allGroupResultsOK(todosInGroup []*types.TODO, results map[string]*ExecutionResult) bool {
	if len(todosInGroup) == 0 {
		return false
	}
	for _, todo := range todosInGroup {
		r, ok := results[todo.FilePath]
		if !ok || r == nil || !r.Success {
			return false
		}
	}
	return true
}

// markGroupCheckFailure flips every todo in the group to failed when the shared
// check loop could not get the suite green, recording the reason on each result.
func (e *TODOExecutor) markGroupCheckFailure(ctx *ExecutorContext, todosInGroup []*types.TODO, results map[string]*ExecutionResult, checkResult *ExecutionResult) {
	for _, todo := range todosInGroup {
		todo.Status = types.StatusFailed
		todo.Attempts++
		if r := results[todo.FilePath]; r != nil {
			r.Success = false
			r.ErrorMessage = checkResult.ErrorMessage
			if checkResult.CommitSHA != "" {
				r.CommitSHA = checkResult.CommitSHA
			}
			if saveErr := e.saveAttempt(ctx, todo, r); saveErr != nil {
				fmt.Fprintf(os.Stderr, "failed to save attempt: %v\n", saveErr)
			}
		}
		e.updateProviderState(ctx, todo, StateUpdate{Status: &todo.Status, Attempts: &todo.Attempts})
	}
}

// firstChecksConfig returns the first TODO frontmatter `checks` block in the
// group, or nil when none set one. Frontmatter overrides the project default.
func firstChecksConfig(todosInGroup []*types.TODO) *types.AgentChecksConfig {
	for _, todo := range todosInGroup {
		if todo != nil && todo.Checks != nil {
			return todo.Checks
		}
	}
	return nil
}
