package todos

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

// fakeFeedbackExec implements Executor and FeedbackExecutor, returning a
// scripted result for the resume turn.
type fakeFeedbackExec struct {
	feedbacks int
	result    *ExecutionResult
}

func (f *fakeFeedbackExec) Name() string { return "fake-feedback" }
func (f *fakeFeedbackExec) Execute(*ExecutorContext, *types.TODO) (*ExecutionResult, error) {
	return &ExecutionResult{Success: true, ExecutorName: f.Name()}, nil
}
func (f *fakeFeedbackExec) SendFeedback(_ *ExecutorContext, _ []*types.TODO, _ string) (*ExecutionResult, error) {
	f.feedbacks++
	if f.result != nil {
		return f.result, nil
	}
	return &ExecutionResult{Success: true, ExecutorName: f.Name()}, nil
}

// plainExec implements only Executor (no SendFeedback).
type plainExec struct{}

func (plainExec) Name() string { return "plain" }
func (plainExec) Execute(*ExecutorContext, *types.TODO) (*ExecutionResult, error) {
	return &ExecutionResult{Success: true, ExecutorName: "plain"}, nil
}

func newOutcomeCtx() *ExecutorContext {
	return NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
}

func writeOutcomeTodo(t *testing.T, dir string) *types.TODO {
	t.Helper()
	path := filepath.Join(dir, "todo.md")
	content := "---\ntitle: outcome test\npriority: medium\nstatus: in_progress\n---\n\nbody\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	todo, err := ParseTODO(path)
	if err != nil {
		t.Fatal(err)
	}
	return todo
}

func writePlanFile(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "plan-for-outcome.md")
	if err := os.WriteFile(path, []byte("# Plan\n\n1. do it\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestApplyOutcomeRunTransitions(t *testing.T) {
	t.Run("successful run without verification remains pending", func(t *testing.T) {
		dir := t.TempDir()
		todo := writeOutcomeTodo(t, dir)
		e := NewTODOExecutor(dir, plainExec{}, "")
		result := &ExecutionResult{Success: true, ExecutorName: "plain", Summary: "did it", EndStatus: types.EndCompleted}
		if err := e.applyOutcome(context.Background(), todo, result); err != nil {
			t.Fatalf("applyOutcome: %v", err)
		}
		if todo.Status != types.StatusPending {
			t.Errorf("status = %s, want pending", todo.Status)
		}
		if todo.LastRunSummary != "did it" {
			t.Errorf("summary = %q", todo.LastRunSummary)
		}
		if todo.RunMode != types.ModeRun {
			t.Errorf("run mode = %q, want run", todo.RunMode)
		}
	})

	t.Run("ask persists questions", func(t *testing.T) {
		dir := t.TempDir()
		todo := writeOutcomeTodo(t, dir)
		e := NewTODOExecutor(dir, plainExec{}, "")
		result := &ExecutionResult{
			Success: true, ExecutorName: "plain", Summary: "blocked",
			EndStatus: types.EndAsk,
			Questions: []types.AgentQuestion{{Text: "which db?"}},
		}
		if err := e.applyOutcome(context.Background(), todo, result); err != nil {
			t.Fatalf("applyOutcome: %v", err)
		}
		if todo.Status != types.StatusAsk {
			t.Errorf("status = %s, want ask", todo.Status)
		}
		if len(todo.Questions) != 1 || todo.Questions[0].Text != "which db?" {
			t.Errorf("questions = %+v", todo.Questions)
		}
		// A later completed run clears the stale questions.
		done := &ExecutionResult{Success: true, ExecutorName: "plain", Summary: "resolved", EndStatus: types.EndCompleted}
		if err := e.applyOutcome(context.Background(), todo, done); err != nil {
			t.Fatalf("applyOutcome: %v", err)
		}
		if len(todo.Questions) != 0 {
			t.Errorf("questions not cleared: %+v", todo.Questions)
		}
	})

	t.Run("no envelope falls back to pending", func(t *testing.T) {
		dir := t.TempDir()
		todo := writeOutcomeTodo(t, dir)
		e := NewTODOExecutor(dir, plainExec{}, "")
		result := &ExecutionResult{Success: true, ExecutorName: "plain"}
		if err := e.applyOutcome(context.Background(), todo, result); err != nil {
			t.Fatalf("applyOutcome: %v", err)
		}
		if todo.Status != types.StatusPending {
			t.Errorf("status = %s, want pending", todo.Status)
		}
	})

	// The definition-of-done verdict lands the run verified or unverified; a run
	// with no DoD stays open/pending for human review.
	t.Run("DoD verdict maps to verified/unverified", func(t *testing.T) {
		cases := []struct {
			name string
			dod  *DoDOutcome
			want types.Status
		}{
			{"passed → verified", &DoDOutcome{Ran: true, Passed: true}, types.StatusVerified},
			{"failed → unverified", &DoDOutcome{Ran: true, Passed: false}, types.StatusUnverified},
			{"no DoD → pending", nil, types.StatusPending},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				dir := t.TempDir()
				todo := writeOutcomeTodo(t, dir)
				e := NewTODOExecutor(dir, plainExec{}, "")
				result := &ExecutionResult{Success: true, ExecutorName: "plain", EndStatus: types.EndCompleted, DoD: tc.dod}
				if err := e.applyOutcome(context.Background(), todo, result); err != nil {
					t.Fatalf("applyOutcome: %v", err)
				}
				if todo.Status != tc.want {
					t.Errorf("status = %s, want %s", todo.Status, tc.want)
				}
			})
		}
	})

	// An ask outcome wins even when a DoD verdict is present (the agent parked
	// before the checks could pass).
	t.Run("ask beats DoD", func(t *testing.T) {
		dir := t.TempDir()
		todo := writeOutcomeTodo(t, dir)
		e := NewTODOExecutor(dir, plainExec{}, "")
		result := &ExecutionResult{
			Success: true, ExecutorName: "plain", EndStatus: types.EndAsk,
			Questions: []types.AgentQuestion{{Text: "which?"}},
			DoD:       &DoDOutcome{Ran: true, Passed: false},
		}
		if err := e.applyOutcome(context.Background(), todo, result); err != nil {
			t.Fatalf("applyOutcome: %v", err)
		}
		if todo.Status != types.StatusAsk {
			t.Errorf("status = %s, want ask", todo.Status)
		}
	})
}

func TestApplyOutcomePlanTransitions(t *testing.T) {
	newPlanResult := func(status types.PlanStatus, path string) *ExecutionResult {
		return &ExecutionResult{
			Success: true, ExecutorName: "plain", Summary: "planned",
			EndStatus: types.EndCompleted,
			Plan:      &types.PlanResult{Status: status, Path: path},
		}
	}

	t.Run("new plan → review with recorded path", func(t *testing.T) {
		dir := t.TempDir()
		todo := writeOutcomeTodo(t, dir)
		plan := writePlanFile(t, dir)
		e := NewTODOExecutor(dir, plainExec{}, "")
		e.SetMode(types.ModePlan)
		if err := e.applyOutcome(context.Background(), todo, newPlanResult(types.PlanNew, plan)); err != nil {
			t.Fatalf("applyOutcome: %v", err)
		}
		if todo.Status != types.StatusReview {
			t.Errorf("status = %s, want review", todo.Status)
		}
		if todo.PlanPath != plan || todo.PlanStatus != types.PlanNew {
			t.Errorf("plan bookkeeping = %q/%q", todo.PlanPath, todo.PlanStatus)
		}
		if todo.RunMode != types.ModePlan {
			t.Errorf("run mode = %q, want plan", todo.RunMode)
		}
	})

	t.Run("inline plan content → review without path", func(t *testing.T) {
		dir := t.TempDir()
		todo := writeOutcomeTodo(t, dir)
		e := NewTODOExecutor(dir, plainExec{}, "")
		e.SetMode(types.ModePlan)
		result := newPlanResult(types.PlanNew, "")
		result.Plan.Content = "- [x] inspect\n- [ ] implement"

		if err := e.applyOutcome(context.Background(), todo, result); err != nil {
			t.Fatalf("applyOutcome: %v", err)
		}
		if todo.Status != types.StatusReview {
			t.Errorf("status = %s, want review", todo.Status)
		}
		if todo.PlanPath != "" || todo.PlanStatus != types.PlanNew {
			t.Errorf("plan bookkeeping = %q/%q", todo.PlanPath, todo.PlanStatus)
		}
	})

	t.Run("unchanged plan → pending, path kept", func(t *testing.T) {
		dir := t.TempDir()
		todo := writeOutcomeTodo(t, dir)
		prior := writePlanFile(t, dir)
		// The prior plan run persisted the path; an unchanged re-plan keeps it.
		if err := UpdateTODOState(todo, StateUpdate{PlanPath: &prior}); err != nil {
			t.Fatal(err)
		}
		e := NewTODOExecutor(dir, plainExec{}, "")
		e.SetMode(types.ModePlan)
		if err := e.applyOutcome(context.Background(), todo, newPlanResult(types.PlanUnchanged, "")); err != nil {
			t.Fatalf("applyOutcome: %v", err)
		}
		if todo.Status != types.StatusPending {
			t.Errorf("status = %s, want pending", todo.Status)
		}
		if todo.PlanPath != prior {
			t.Errorf("recorded plan path lost: %q", todo.PlanPath)
		}
	})

	t.Run("missing plan file is a hard error", func(t *testing.T) {
		dir := t.TempDir()
		todo := writeOutcomeTodo(t, dir)
		e := NewTODOExecutor(dir, plainExec{}, "")
		e.SetMode(types.ModePlan)
		err := e.applyOutcome(context.Background(), todo, newPlanResult(types.PlanNew, filepath.Join(dir, "missing.md")))
		if err == nil || !strings.Contains(err.Error(), "plan file") {
			t.Fatalf("expected invalid plan file error, got %v", err)
		}
	})

	t.Run("plan ask → ask", func(t *testing.T) {
		dir := t.TempDir()
		todo := writeOutcomeTodo(t, dir)
		e := NewTODOExecutor(dir, plainExec{}, "")
		e.SetMode(types.ModePlan)
		result := &ExecutionResult{
			Success: true, ExecutorName: "plain", Summary: "blocked",
			EndStatus: types.EndAsk,
			Questions: []types.AgentQuestion{{Text: "monorepo?"}},
		}
		if err := e.applyOutcome(context.Background(), todo, result); err != nil {
			t.Fatalf("applyOutcome: %v", err)
		}
		if todo.Status != types.StatusAsk {
			t.Errorf("status = %s, want ask", todo.Status)
		}
	})
}

func TestResumeAppliesOutcome(t *testing.T) {
	dir := t.TempDir()
	todo := writeOutcomeTodo(t, dir)
	todo.Status = types.StatusAsk
	exec := &fakeFeedbackExec{result: &ExecutionResult{
		Success: true, ExecutorName: "fake-feedback", Summary: "answered and done",
		EndStatus: types.EndCompleted,
	}}
	e := NewTODOExecutor(dir, exec, "")
	results, err := e.Resume(newOutcomeCtx(), []*types.TODO{todo}, "use postgres")
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if exec.feedbacks != 1 {
		t.Errorf("feedbacks = %d, want 1", exec.feedbacks)
	}
	if len(results) != 1 || results[0].Summary != "answered and done" {
		t.Errorf("results = %+v", results)
	}
	if todo.Status != types.StatusPending {
		t.Errorf("status = %s, want pending", todo.Status)
	}
}

func TestResumeRequiresFeedbackExecutor(t *testing.T) {
	dir := t.TempDir()
	todo := writeOutcomeTodo(t, dir)
	e := NewTODOExecutor(dir, plainExec{}, "")
	if _, err := e.Resume(newOutcomeCtx(), []*types.TODO{todo}, "answer"); err == nil {
		t.Fatal("Resume must error for executors without session resumption")
	}
}

func TestBuildCheckVerifiers(t *testing.T) {
	t.Run("disabled yields none", func(t *testing.T) {
		plugins, _, err := BuildCheckVerifiers(t.TempDir(), []*types.TODO{{}}, nil)
		if err != nil {
			t.Fatalf("BuildCheckVerifiers: %v", err)
		}
		if len(plugins) != 0 {
			t.Errorf("plugins = %d, want 0 when checks are disabled", len(plugins))
		}
	})

	t.Run("frontmatter lint config yields a fixture verifier", func(t *testing.T) {
		enabled := true
		todo := &types.TODO{}
		todo.Checks = &types.AgentChecksConfig{
			Enabled:       &enabled,
			MaxIterations: 2,
			Lint:          &types.AgentLintConfig{},
		}
		plugins, maxIter, err := BuildCheckVerifiers(t.TempDir(), []*types.TODO{todo}, nil)
		if err != nil {
			t.Fatalf("BuildCheckVerifiers: %v", err)
		}
		if len(plugins) != 1 {
			t.Fatalf("plugins = %d, want 1 aggregate DoD verifier", len(plugins))
		}
		if plugins[0].Name() != "definition-of-done" {
			t.Errorf("plugin name = %q", plugins[0].Name())
		}
		if maxIter != 3 {
			t.Errorf("maxIter = %d, want initial run + 2 feedback rounds", maxIter)
		}
	})

	t.Run("verification fixture auto-enables the DoD verifier", func(t *testing.T) {
		// No checks config at all, but the todo carries a `## Verification`
		// fixture — its definition of done gates the loop regardless.
		todo := &types.TODO{}
		todo.Verification = []*fixtures.FixtureNode{{Test: &fixtures.FixtureTest{Name: "it works"}}}
		plugins, maxIter, err := BuildCheckVerifiers(t.TempDir(), []*types.TODO{todo}, nil)
		if err != nil {
			t.Fatalf("BuildCheckVerifiers: %v", err)
		}
		if len(plugins) != 1 || plugins[0].Name() != "definition-of-done" {
			t.Fatalf("plugins = %d, want one 'definition-of-done' verifier", len(plugins))
		}
		if maxIter != types.DefaultMaxCheckIterations+1 {
			t.Errorf("maxIter = %d, want default budget %d", maxIter, types.DefaultMaxCheckIterations+1)
		}
	})

	t.Run("acceptance criteria auto-enable the DoD verifier as a checklist", func(t *testing.T) {
		// No checks and no `## Verification`, just criteria — they become the
		// checklist ai step feeding results.checklist.
		todo := &types.TODO{}
		todo.AcceptanceCriteria = []types.AcceptanceCriterion{{Text: "retries on 5xx"}}
		plugins, _, err := BuildCheckVerifiers(t.TempDir(), []*types.TODO{todo}, nil)
		if err != nil {
			t.Fatalf("BuildCheckVerifiers: %v", err)
		}
		if len(plugins) != 1 || plugins[0].Name() != "definition-of-done" {
			t.Fatalf("plugins = %d, want one 'definition-of-done' verifier from criteria", len(plugins))
		}
		if v, ok := plugins[0].(*celVerifier); !ok || v.aiStep == nil {
			t.Fatal("criteria should synthesize the checklist ai step on the verifier")
		}
	})

	t.Run("workflow verify maxIterations overrides the loop cap", func(t *testing.T) {
		// A run carrying a Workflow.Verify.maxIterations overrides the default
		// budget (initial run + N feedback rounds).
		todo := &types.TODO{}
		todo.Verification = []*fixtures.FixtureNode{{Test: &fixtures.FixtureTest{Name: "it works"}}}
		plugins, maxIter, err := BuildCheckVerifiers(t.TempDir(), []*types.TODO{todo}, &api.Verify{MaxIterations: 5})
		if err != nil {
			t.Fatalf("BuildCheckVerifiers: %v", err)
		}
		if len(plugins) != 1 {
			t.Fatalf("plugins = %d, want 1 aggregate DoD verifier", len(plugins))
		}
		if maxIter != 6 {
			t.Errorf("maxIter = %d, want 5 override + 1 initial run", maxIter)
		}
	})
}
