package todos

import (
	"context"
	"testing"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/todos/checks"
	"github.com/flanksource/gavel/todos/types"
)

// checkState drives a fake lint runner: it reports a violation until failsLeft
// reaches zero. SendFeedback decrements it, modelling the agent fixing one issue
// per round.
type checkState struct{ failsLeft int }

func (s *checkState) lintFunc() checks.LintFunc {
	return func(ctx context.Context, workDir string, cfg types.AgentLintConfig) ([]*linters.LinterResult, error) {
		if s.failsLeft > 0 {
			return []*linters.LinterResult{{
				Linter:     "fake",
				Success:    true,
				Violations: []models.Violation{{File: "f.go", Line: 1, Message: models.StringPtr("boom")}},
			}}, nil
		}
		return []*linters.LinterResult{{Linter: "fake", Success: true}}, nil
	}
}

// fakeFeedbackExec implements Executor and FeedbackExecutor; each SendFeedback
// "fixes" one failure.
type fakeFeedbackExec struct {
	state     *checkState
	feedbacks int
}

func (f *fakeFeedbackExec) Name() string { return "fake-feedback" }
func (f *fakeFeedbackExec) Execute(*ExecutorContext, *types.TODO) (*ExecutionResult, error) {
	return &ExecutionResult{Success: true, ExecutorName: f.Name()}, nil
}
func (f *fakeFeedbackExec) SendFeedback(_ *ExecutorContext, _ []*types.TODO, _ string) (*ExecutionResult, error) {
	f.feedbacks++
	if f.state.failsLeft > 0 {
		f.state.failsLeft--
	}
	return &ExecutionResult{Success: true, ExecutorName: f.Name()}, nil
}

// plainExec implements only Executor (no SendFeedback).
type plainExec struct{}

func (plainExec) Name() string { return "plain" }
func (plainExec) Execute(*ExecutorContext, *types.TODO) (*ExecutionResult, error) {
	return &ExecutionResult{Success: true, ExecutorName: "plain"}, nil
}

func lintOnlyTodo(maxIters int) *types.TODO {
	enabled := true
	todo := &types.TODO{}
	todo.Checks = &types.AgentChecksConfig{
		Enabled:       &enabled,
		MaxIterations: maxIters,
		Lint:          &types.AgentLintConfig{}, // lint only — no real test subprocess
	}
	return todo
}

func newLoopCtx() *ExecutorContext {
	return NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
}

func withFakeLint(t *testing.T, fn checks.LintFunc) {
	t.Helper()
	checks.SetLintRunner(fn)
	t.Cleanup(func() { checks.SetLintRunner(nil) })
}

func TestRunCheckLoop_ConvergesAfterFeedback(t *testing.T) {
	state := &checkState{failsLeft: 2}
	withFakeLint(t, state.lintFunc())
	exec := &fakeFeedbackExec{state: state}

	e := NewTODOExecutor(t.TempDir(), exec, "")
	result := &ExecutionResult{Success: true, ExecutorName: exec.Name()}
	e.runCheckLoop(newLoopCtx(), []*types.TODO{lintOnlyTodo(3)}, result)

	if !result.Success {
		t.Errorf("expected success after convergence, got failure: %s", result.ErrorMessage)
	}
	if exec.feedbacks != 2 {
		t.Errorf("expected exactly 2 feedback rounds, got %d", exec.feedbacks)
	}
}

func TestRunCheckLoop_StopsAtMaxIterations(t *testing.T) {
	state := &checkState{failsLeft: 10} // never fixed
	withFakeLint(t, state.lintFunc())
	exec := &fakeFeedbackExec{state: state}

	e := NewTODOExecutor(t.TempDir(), exec, "")
	result := &ExecutionResult{Success: true, ExecutorName: exec.Name()}
	e.runCheckLoop(newLoopCtx(), []*types.TODO{lintOnlyTodo(2)}, result)

	if result.Success {
		t.Errorf("expected failure when checks never pass")
	}
	if exec.feedbacks != 2 {
		t.Errorf("expected exactly maxIterations=2 feedback rounds, got %d", exec.feedbacks)
	}
}

func TestRunCheckLoop_DegradesWithoutFeedbackExecutor(t *testing.T) {
	state := &checkState{failsLeft: 1}
	withFakeLint(t, state.lintFunc())

	e := NewTODOExecutor(t.TempDir(), plainExec{}, "")
	result := &ExecutionResult{Success: true, ExecutorName: "plain"}
	e.runCheckLoop(newLoopCtx(), []*types.TODO{lintOnlyTodo(3)}, result)

	if result.Success {
		t.Errorf("expected failure reported when executor cannot feed back")
	}
}

func TestRunCheckLoop_DisabledIsNoop(t *testing.T) {
	withFakeLint(t, func(context.Context, string, types.AgentLintConfig) ([]*linters.LinterResult, error) {
		t.Fatal("lint runner must not be called when checks are disabled")
		return nil, nil
	})
	exec := &fakeFeedbackExec{state: &checkState{}}
	e := NewTODOExecutor(t.TempDir(), exec, "") // no EnableChecks, no frontmatter checks

	result := &ExecutionResult{Success: true, ExecutorName: exec.Name()}
	e.runCheckLoop(newLoopCtx(), []*types.TODO{{}}, result)

	if !result.Success {
		t.Errorf("disabled check loop must leave result untouched")
	}
}
