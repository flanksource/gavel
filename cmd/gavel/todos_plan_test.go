package main

import (
	"context"
	"testing"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// seedCLIReviewTodo creates a todo parked in review behind a stub plan-review
// provider, wired as the runtime the CLI opens.
func seedCLIReviewTodo(t *testing.T, workDir, title string) (*testPlanReviewProvider, *types.TODO) {
	t.Helper()
	provider := &testPlanReviewProvider{testTODOProvider: testProviderFor(workDir)}
	oldOpen := openRuntimeTodosProvider
	openRuntimeTodosProvider = func(context.Context, string) (todos.Provider, error) { return provider, nil }
	t.Cleanup(func() { openRuntimeTodosProvider = oldOpen })

	created, err := provider.Create(t.Context(), todos.CreateRequest{Title: title, Status: types.StatusPending})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	review := types.StatusReview
	planPath := "/plans/p.md"
	planNew := types.PlanNew
	if err := provider.UpdateState(t.Context(), created, todos.StateUpdate{
		Status: &review, PlanPath: &planPath, PlanStatus: &planNew,
	}); err != nil {
		t.Fatalf("seed review: %v", err)
	}

	oldWorkingDir := workingDir
	workingDir = workDir
	t.Cleanup(func() { workingDir = oldWorkingDir })
	return provider, created
}

// stubApprovedRun intercepts the implement run --run chains, so approving in a
// test records the dispatch instead of spawning an agent.
func stubApprovedRun(t *testing.T) *types.TODOS {
	t.Helper()
	var dispatched types.TODOS
	old := runApprovedTODO
	runApprovedTODO = func(_ string, todoList types.TODOS, _ *todos.UserInteraction, _ todos.Provider) error {
		dispatched = todoList
		return nil
	}
	t.Cleanup(func() { runApprovedTODO = old })
	return &dispatched
}

func TestTodosPlanCommandsRegistered(t *testing.T) {
	registered := map[string]bool{}
	for _, command := range todosPlanCmd.Commands() {
		registered[command.Name()] = true
	}
	for _, name := range []string{"approve", "reject", "revise", "recover"} {
		if !registered[name] {
			t.Errorf("expected todos plan %s to be registered", name)
		}
	}
	if todosPlanApproveCmd.Flags().Lookup("run") == nil {
		t.Error("expected plan approve --run flag")
	}
}

func TestTodosPlanRecoverCallsProviderOutsideReviewState(t *testing.T) {
	workDir := t.TempDir()
	provider, created := seedCLIReviewTodo(t, workDir, "Recoverable plan")
	failed := types.StatusFailed
	if err := provider.UpdateState(t.Context(), created, todos.StateUpdate{Status: &failed}); err != nil {
		t.Fatalf("seed failed state: %v", err)
	}

	if err := runTodosPlanRecover(todosPlanRecoverCmd, []string{"Recoverable plan"}); err != nil {
		t.Fatalf("recover: %v", err)
	}
	if provider.recoveredID != created.ID {
		t.Errorf("recovered ID = %q, want %q", provider.recoveredID, created.ID)
	}
}

// Approving without --run is a review transition only: the plan is approved and
// the user decides when to implement it.
func TestTodosPlanApproveDoesNotRunByDefault(t *testing.T) {
	provider, created := seedCLIReviewTodo(t, t.TempDir(), "Approvable plan")
	dispatched := stubApprovedRun(t)

	oldRun := planApproveRun
	planApproveRun = false
	t.Cleanup(func() { planApproveRun = oldRun })

	if err := runTodosPlanApprove(todosPlanApproveCmd, []string{"Approvable plan"}); err != nil {
		t.Fatalf("approve: %v", err)
	}
	reloaded, err := provider.Get(t.Context(), created.FilePath)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.Status == types.StatusReview {
		t.Error("approved todo is still awaiting review")
	}
	if len(*dispatched) != 0 {
		t.Errorf("approve dispatched %d runs without --run", len(*dispatched))
	}
}

// --run chains the implement run: the approved plan executes in run mode as a
// fresh turn, never resuming the planning conversation.
func TestTodosPlanApproveWithRunChainsAnImplementRun(t *testing.T) {
	_, created := seedCLIReviewTodo(t, t.TempDir(), "Runnable plan")
	dispatched := stubApprovedRun(t)

	oldRun, oldMode, oldResume := planApproveRun, todosRunMode, resumeSession
	planApproveRun = true
	todosRunMode = types.ModePlan // the mode a plan run left behind
	resumeSession = true
	t.Cleanup(func() {
		planApproveRun, todosRunMode, resumeSession = oldRun, oldMode, oldResume
	})

	if err := runTodosPlanApprove(todosPlanApproveCmd, []string{"Runnable plan"}); err != nil {
		t.Fatalf("approve --run: %v", err)
	}
	if len(*dispatched) != 1 || (*dispatched)[0].ID != created.ID {
		t.Fatalf("dispatched = %#v, want the approved todo", *dispatched)
	}
	if todosRunMode != types.ModeRun {
		t.Errorf("run mode = %q, want run — an approved plan implements, it does not re-plan", todosRunMode)
	}
	if resumeSession {
		t.Error("approved plan resumed the planning session instead of opening an implement turn")
	}
}

// A todo that is not in review has no plan to approve.
func TestTodosPlanApproveRejectsNonReviewTodo(t *testing.T) {
	workDir := t.TempDir()
	provider, created := seedCLIReviewTodo(t, workDir, "Not in review")
	stubApprovedRun(t)
	inProgress := types.StatusInProgress
	if err := provider.UpdateState(t.Context(), created, todos.StateUpdate{Status: &inProgress}); err != nil {
		t.Fatalf("seed status: %v", err)
	}

	err := runTodosPlanApprove(todosPlanApproveCmd, []string{"Not in review"})
	if err == nil {
		t.Fatal("expected approve to reject a todo that is not awaiting review")
	}
}
