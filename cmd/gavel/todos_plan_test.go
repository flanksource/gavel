package main

import (
	"context"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
)

// seedCLIReviewTodo creates a todo parked in review behind a stub plan-review
// provider, wired as the runtime the CLI opens. The plan actions place their
// continuation in the workspace's lifecycle, which reads the home layer, so the
// home is isolated like every other config-loading CLI test's.
func seedCLIReviewTodo(t *testing.T, workDir, title string) (*testPlanReviewProvider, *types.TODO) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
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

// approvedDispatch is one implement run the approve action chained.
type approvedDispatch struct {
	todo *types.TODO
	opts run.Options
}

// stubApprovedRun intercepts the implement run --run chains, so approving in a
// test records the dispatch instead of spawning an agent.
func stubApprovedRun(t *testing.T) *[]approvedDispatch {
	t.Helper()
	var dispatched []approvedDispatch
	old := runApprovedTODO
	runApprovedTODO = func(_ string, todo *types.TODO, _ todos.Provider, opts run.Options) error {
		dispatched = append(dispatched, approvedDispatch{todo: todo, opts: opts})
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

// --run chains the implement run: the approved plan executes as the run step in
// a fresh turn, never resuming the planning conversation — whatever the run
// flags a plan run may have left behind.
func TestTodosPlanApproveWithRunChainsAnImplementRun(t *testing.T) {
	_, created := seedCLIReviewTodo(t, t.TempDir(), "Runnable plan")
	dispatched := stubApprovedRun(t)

	oldRun, oldStep, oldResume := planApproveRun, todosStep, resumeSession
	planApproveRun = true
	todosStep = "plan" // the step a plan run left behind
	resumeSession = true
	t.Cleanup(func() {
		planApproveRun, todosStep, resumeSession = oldRun, oldStep, oldResume
	})

	if err := runTodosPlanApprove(todosPlanApproveCmd, []string{"Runnable plan"}); err != nil {
		t.Fatalf("approve --run: %v", err)
	}
	if len(*dispatched) != 1 || (*dispatched)[0].todo.ID != created.ID {
		t.Fatalf("dispatched = %#v, want the approved todo", *dispatched)
	}
	opts := (*dispatched)[0].opts
	if opts.Step != "run" {
		t.Errorf("step = %q, want run — an approved plan implements, it does not re-plan", opts.Step)
	}
	if opts.Resume || opts.Message != "" {
		t.Error("approved plan resumed the planning session instead of opening an implement turn")
	}
}

// codexPlanRun is the plan turn as the native runtime recorded it: requested
// as the codex family alias, resolved to a concrete codex model.
func codexPlanRun() *captaindb.PromptRun {
	return &captaindb.PromptRun{
		State: captaindb.PromptRunStateWaiting,
		Runtime: captaindb.PromptRunRuntime{
			Mode: string(types.ModePlan),
			Resolved: captaindb.PromptRunRuntimeSelection{
				Provider: "openai", Mode: "agent", Model: "gpt-5.6-sol", Effort: "high",
			},
		},
	}
}

// priorRuntimeLayer is the runtime a continuation inherited from the run it
// continues, as the layer the implement run folds below its own request.
func priorRuntimeLayer(t *testing.T, opts run.Options) api.Model {
	t.Helper()
	for _, layer := range opts.Prior {
		if layer.Name == "prior run runtime" {
			return layer.Spec.Model
		}
	}
	t.Fatalf("dispatch carried no prior run runtime layer: %+v", opts.Prior)
	return api.Model{}
}

// The implement run --run chains inherits the runtime the plan run resolved, so
// a plan drawn up by codex is implemented by codex — the same continuation seam
// the dashboard's approve action dispatches through.
func TestTodosPlanApproveWithRunInheritsThePlanRunRuntime(t *testing.T) {
	provider, _ := seedCLIReviewTodo(t, t.TempDir(), "Codex plan")
	provider.activeRun = codexPlanRun()
	dispatched := stubApprovedRun(t)

	oldRun := planApproveRun
	planApproveRun = true
	t.Cleanup(func() { planApproveRun = oldRun })

	if err := runTodosPlanApprove(todosPlanApproveCmd, []string{"Codex plan"}); err != nil {
		t.Fatalf("approve --run: %v", err)
	}
	if len(*dispatched) != 1 {
		t.Fatalf("dispatched %d runs, want 1", len(*dispatched))
	}
	runtime := priorRuntimeLayer(t, (*dispatched)[0].opts)
	if runtime.Name != "gpt-5.6-sol" || runtime.Mode != api.ModeAgent || string(runtime.Effort) != "high" {
		t.Errorf("inherited runtime = %s/%s/%s, want gpt-5.6-sol/agent/high — a codex plan must not be implemented by claude",
			runtime.Name, runtime.Mode, runtime.Effort)
	}
}

// Revising resumes the plan session on the runtime that session belongs to,
// through the same continuation seam; the run itself is stubbed at the
// dispatcher so the revision's options are what is asserted.
func TestTodosPlanReviseResumesOnThePlanRunRuntime(t *testing.T) {
	provider, created := seedCLIReviewTodo(t, t.TempDir(), "Codex revision")
	provider.activeRun = codexPlanRun()
	session := "sess-codex-plan"
	if err := provider.UpdateState(t.Context(), created, todos.StateUpdate{SessionID: &session}); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	var got run.Request
	oldResolve, oldStart := run.Resolve, run.Start
	run.Resolve = func(_ context.Context, req run.Request) (*run.Prepared, error) {
		got = req
		return &run.Prepared{Step: lifecycle.Step{Name: req.Options.Step}, Resolution: &lifecycle.Resolution{}}, nil
	}
	run.Start = func(run.Request) (run.StartResult, error) { return run.StartResult{Status: "started"}, nil }
	t.Cleanup(func() { run.Resolve, run.Start = oldResolve, oldStart })
	oldFeedback := planReviseFeedback
	planReviseFeedback = "bound the queue"
	t.Cleanup(func() { planReviseFeedback = oldFeedback })

	if err := runTodosPlanRevise(todosPlanReviseCmd, []string{"Codex revision"}); err != nil {
		t.Fatalf("revise: %v", err)
	}
	if got.Options.Step != "plan" || !got.Options.Resume || got.Options.Message != "bound the queue" {
		t.Errorf("revise dispatched %+v, want the plan step resumed with the feedback", got.Options)
	}
	if runtime := priorRuntimeLayer(t, got.Options); runtime.Name != "gpt-5.6-sol" || runtime.Mode != api.ModeAgent {
		t.Errorf("inherited runtime = %s/%s, want gpt-5.6-sol/agent", runtime.Name, runtime.Mode)
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
