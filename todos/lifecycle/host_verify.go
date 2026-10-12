package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// VerifyStep runs the lifecycle's `verify` step for one todo and applies its
// outcome — the whole of what `gavel todos check` and the dashboard's verify
// action do.
//
// It is a named method rather than a caller assembling RunStep + OnOutcome for
// itself because "check this todo" has to mean the same thing from the terminal
// and from the dashboard, and because the verify step is the ONLY verifier: a
// second path that ran the fixture directly could reach a verdict the lifecycle
// never recorded.
func (h *Host) VerifyStep(ctx context.Context, todo *types.TODO, request api.Spec) (*types.CheckResult, error) {
	if h.Def == nil {
		return nil, fmt.Errorf("lifecycle host: no lifecycle loaded")
	}
	step, ok := h.Def.Definition().Step(StepVerify)
	if !ok {
		return nil, fmt.Errorf("lifecycle %s has no %q step", h.Def.Definition().Name, StepVerify)
	}
	start := time.Now()
	outcome, err := h.RunStep(ctx, todo, step, RunOptions{Request: request})
	if outcome == nil {
		if err == nil {
			err = fmt.Errorf("verify step produced no outcome")
		}
		return nil, err
	}
	status := outcome.Status
	if err != nil && status == "" {
		// The run happened; it just could not be classified. Record the attempt
		// under the status it already had rather than losing it.
		status = OutcomeKeep
	}
	if applyErr := h.OnOutcome(ctx, todo, step, outcome, status); applyErr != nil {
		err = errors.Join(err, applyErr)
	}
	if err != nil {
		return nil, err
	}
	report := outcome.Result.Verify
	if report == nil {
		return nil, fmt.Errorf("definition of done produced no report")
	}
	return h.checkResult(ctx, todo, outcome, *report, time.Since(start))
}

// checkResult renders the verdict for display and persists the "Latest Failure"
// record a failing check leaves on the todo. The commit it reports is the one
// the verify step checked — a run's worktree head is not the main checkout's.
func (h *Host) checkResult(ctx context.Context, todo *types.TODO, outcome *StepOutcome, report api.VerifyReport, elapsed time.Duration) (*types.CheckResult, error) {
	var workspace *api.WorkspaceRecord
	if outcome.Execution != nil {
		workspace = outcome.Execution.Workspace
	}
	target, info, err := verifiedTarget(outcome.VerifyTarget, workspace, h.stepWorkDir(todo))
	if err != nil {
		return nil, err
	}
	info.Timestamp, info.Passed, info.Duration = time.Now().Add(-elapsed), report.Passed, elapsed
	testResult := todos.BuildTestResultInfo(report, info)
	result := &types.CheckResult{
		TODO: todo, AllPassed: report.Passed, Duration: elapsed,
		Report: &report, TestResult: testResult, Target: &target,
	}
	if report.Passed || h.Provider == nil {
		return result, nil
	}
	persistCtx, cancel := todos.PersistenceContext(ctx)
	defer cancel()
	if err := h.Provider.UpdateLatestFailure(persistCtx, todo, testResult); err != nil {
		// The verdict itself is already recorded; losing its rendered failure
		// section is not worth turning a reported result into an error.
		fmt.Fprintf(os.Stderr, "failed to update Latest Failure section for %s: %v\n", todos.TODOReference(todo), err)
	}
	return result, nil
}
