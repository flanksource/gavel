package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
)

// isolatedTodosRun writes cfg as the workspace's .gavel.yaml and neutralises
// every other input the resolution seam reads: HOME, because LoadGavelConfig
// merges ~/.gavel.yaml and would otherwise make whoever runs the suite a hidden
// layer, and the run flags, which are the highest layer — a value another test
// left behind would outrank the config under test.
func isolatedTodosRun(t *testing.T, cfg string) string {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	resetTodosRunFlags(t)

	dir := t.TempDir()
	if strings.TrimSpace(cfg) != "" {
		if err := os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte(cfg), 0o644); err != nil {
			t.Fatalf("write .gavel.yaml: %v", err)
		}
	}
	return dir
}

// resetTodosRunFlags restores every `todos run` flag to its registered default
// for the duration of the test. The flags are the request layer, so one left
// behind by another test outranks the configuration under test.
func resetTodosRunFlags(t *testing.T) {
	t.Helper()
	saved := struct {
		step, model, effort, status string
		budget                      float64
		turns                       int
		commit                      bool
		dirtyRun, dry, resume       bool
		force, pick                 bool
	}{
		todosStep, todoModel, todoEffort, filterStatus,
		maxBudget, maxTurns,
		commitAfter,
		dirty, dryRun, resumeSession,
		forceRun, interactive,
	}
	t.Cleanup(func() {
		todosStep, todoModel, todoEffort, filterStatus = saved.step, saved.model, saved.effort, saved.status
		maxBudget, maxTurns = saved.budget, saved.turns
		commitAfter = saved.commit
		dirty, dryRun, resumeSession = saved.dirtyRun, saved.dry, saved.resume
		forceRun, interactive = saved.force, saved.pick
	})
	todosStep, todoModel, todoEffort, filterStatus = "", "", "", ""
	maxBudget, maxTurns = 0, 0
	commitAfter = true
	dirty, dryRun, resumeSession = false, false, false
	forceRun, interactive = false, false
}

// stubRunTodoProvider serves a fixed backlog; the run seams are stubbed above
// it, so nothing else on the provider is reached.
type stubRunTodoProvider struct {
	todos.Provider
	items types.TODOS
}

func (p *stubRunTodoProvider) List(context.Context, todos.DiscoveryFilters) (types.TODOS, error) {
	return p.items, nil
}

// stubTodoRunSeams replaces run.Resolve and run.Start with recorders returning
// resolution, so a dispatch can be asserted without standing up an agent.
func stubTodoRunSeams(t *testing.T, resolution *lifecycle.Resolution, step, reason string) *[]run.Request {
	t.Helper()
	var started []run.Request
	oldResolve, oldStart := run.Resolve, run.Start
	run.Resolve = func(_ context.Context, req run.Request) (*run.Prepared, error) {
		return &run.Prepared{
			Step:       lifecycle.Step{Name: step},
			Reason:     reason,
			Resolution: resolution,
			SessionID:  "session-" + req.Todo.ID,
		}, nil
	}
	run.Start = func(req run.Request) (run.StartResult, error) {
		started = append(started, req)
		return run.StartResult{Status: "started", SessionID: "session-" + req.Todo.ID}, nil
	}
	t.Cleanup(func() { run.Resolve, run.Start = oldResolve, oldStart })
	return &started
}

func stubTodoBacklog(t *testing.T, items ...*types.TODO) {
	t.Helper()
	provider := &stubRunTodoProvider{items: items}
	old := openRuntimeTodosProvider
	openRuntimeTodosProvider = func(context.Context, string) (todos.Provider, error) { return provider, nil }
	t.Cleanup(func() { openRuntimeTodosProvider = old })
}

func runTodo(id, title string) *types.TODO {
	return &types.TODO{
		ID:              id,
		Provider:        todos.ProviderDB,
		TODOFrontmatter: types.TODOFrontmatter{Title: title, Status: types.StatusPending},
	}
}

// The step axis replaced --mode/--prompt, so its help has to name where the
// vocabulary comes from: a project's lifecycle declares it, and `todos steps`
// lists it.
func TestTodosRunStepFlagDocumentsTheLifecycleVocabulary(t *testing.T) {
	step := todosRunCmd.Flags().Lookup("step")
	if step == nil {
		t.Fatal("expected todos run --step flag to be registered")
	}
	for _, want := range []string{"lifecycle", "gavel todos steps"} {
		if !strings.Contains(step.Usage, want) {
			t.Errorf("--step usage %q does not mention %q", step.Usage, want)
		}
	}
	for _, keep := range []string{"model", "effort", "max-budget", "max-turns", "resume", "force", "dirty", "dry-run", "commit", "interactive", "status"} {
		if todosRunCmd.Flags().Lookup(keep) == nil {
			t.Errorf("expected todos run --%s flag to survive the step cutover", keep)
		}
	}
}

// The request is the TOP spec layer, so an untouched flag must contribute
// nothing: a non-zero default here silently beats the configuration it claims
// to defer to.
func TestTodosRequestSpecCarriesOnlyFlagsThatWereSet(t *testing.T) {
	resetTodosRunFlags(t)
	if spec := todosRequestSpec(); !api.IsEmpty(spec) {
		t.Fatalf("default run flags contributed %+v, want an empty request layer", spec)
	}

	todoModel, todoEffort = "cli:opus:high", "high"
	maxBudget, maxTurns = 2.5, 7
	dirty = true
	spec := todosRequestSpec()
	if spec.Name != "cli:opus:high" || spec.Effort != api.Effort("high") {
		t.Errorf("model/effort = %q/%q", spec.Name, spec.Effort)
	}
	if spec.Budget.Cost != 2.5 || spec.Budget.MaxTurns != 7 {
		t.Errorf("budget = %+v", spec.Budget)
	}
	if spec.Setup == nil || spec.Setup.Checkout == nil || spec.Setup.Checkout.Worktree == nil ||
		spec.Setup.Checkout.Worktree.Uncommitted != shell.CloneClone {
		t.Errorf("--dirty setup = %+v, want a worktree carrying uncommitted content", spec.Setup)
	}
	// No flag speaks for the workflow: the lifecycle step declares it, and a
	// request-layer verify stub would only ever shadow the step's own.
	if spec.Workflow != nil {
		t.Errorf("request workflow = %+v, want none", spec.Workflow)
	}
}

// One issue, one lifecycle, one run: every discovered todo gets its own request
// naming the step the caller asked for, dispatched from the CLI host.
func TestRunTodosRunStartsOneRunPerTodoNamingTheStep(t *testing.T) {
	workingDirFor(t, isolatedTodosRun(t, ""))
	first, second := runTodo("aaaaaaaa-0000-4000-8000-000000000001", "First"), runTodo("aaaaaaaa-0000-4000-8000-000000000002", "Second")
	stubTodoBacklog(t, first, second)
	started := stubTodoRunSeams(t, &lifecycle.Resolution{Prompt: "prompt body"}, "plan", "applies: subject.status == 'pending'")
	todosStep = "plan"
	resumeSession = true

	out := captureStdout(t, func() {
		if err := runTodosRun(todosRunCmd, nil); err != nil {
			t.Fatalf("runTodosRun: %v", err)
		}
	})

	if len(*started) != 2 {
		t.Fatalf("started %d runs, want one per todo", len(*started))
	}
	for i, req := range *started {
		if req.Options.Step != "plan" {
			t.Errorf("run %d step = %q, want plan", i, req.Options.Step)
		}
		if req.Options.Host != lifecycle.HostCLI {
			t.Errorf("run %d host = %q, want the CLI host", i, req.Options.Host)
		}
		if !req.Options.Resume {
			t.Errorf("run %d dropped --resume", i)
		}
		if req.Options.Concurrent {
			t.Errorf("run %d was dispatched concurrently without --force", i)
		}
		if req.Registry == nil || req.Provider == nil {
			t.Errorf("run %d = %+v, want the process registry and provider", i, req)
		}
	}
	if (*started)[0].Todo != first || (*started)[1].Todo != second {
		t.Errorf("dispatched todos = %v/%v, want the discovered pair", (*started)[0].Todo, (*started)[1].Todo)
	}
	// The chosen step and why it was chosen are reported before the dispatch: a
	// lifecycle picking the step means the caller cannot know it otherwise.
	if !strings.Contains(out, "plan") || !strings.Contains(out, "applies: subject.status == 'pending'") {
		t.Errorf("run output did not report the chosen step and reason:\n%s", out)
	}
}

// --dry-run is a preview of exactly the run that would follow: the rendered
// prompt, the layer stack that produced the spec, and the spec itself.
func TestRunTodosRunDryRunPrintsPromptAndTraceWithoutDispatching(t *testing.T) {
	workingDirFor(t, isolatedTodosRun(t, ""))
	stubTodoBacklog(t, runTodo("aaaaaaaa-0000-4000-8000-000000000003", "Previewable"))
	resolution := &lifecycle.Resolution{
		Prompt: "IMPLEMENT-THIS-TODO",
		Spec:   api.Spec{Model: api.Model{Name: "claude-sonnet-5"}, Budget: api.Budget{MaxTurns: 9}},
		Trace: []api.SpecLayer{
			{Name: ".gavel.yaml ai", Scope: api.SpecLayerGlobal},
			{Name: "todos-run.prompt", Scope: api.SpecLayerSurface},
			{Name: "request", Scope: api.SpecLayerUser},
		},
	}
	started := stubTodoRunSeams(t, resolution, "run", "always applies (no `when` predicate)")
	dryRun = true

	out := captureStdout(t, func() {
		if err := runTodosRun(todosRunCmd, nil); err != nil {
			t.Fatalf("runTodosRun --dry-run: %v", err)
		}
	})

	if len(*started) != 0 {
		t.Fatalf("--dry-run dispatched %d runs", len(*started))
	}
	if !strings.Contains(out, "IMPLEMENT-THIS-TODO") {
		t.Errorf("--dry-run did not print the rendered prompt:\n%s", out)
	}
	ai := strings.Index(out, ".gavel.yaml ai")
	prompt := strings.Index(out, "todos-run.prompt")
	request := strings.Index(out, "request")
	if ai < 0 || prompt < 0 || request < 0 || !(ai < prompt && prompt < request) {
		t.Errorf("--dry-run trace is not lowest-precedence-first:\n%s", out)
	}
	if !strings.Contains(out, "maxTurns: 9") {
		t.Errorf("--dry-run did not print the resolved spec as YAML:\n%s", out)
	}
}

// --force answers the "already running elsewhere" question up front, so the
// first dispatch is already the concurrent one.
func TestRunTodosRunForceDispatchesConcurrently(t *testing.T) {
	workingDirFor(t, isolatedTodosRun(t, ""))
	stubTodoBacklog(t, runTodo("aaaaaaaa-0000-4000-8000-000000000004", "Forced"))
	started := stubTodoRunSeams(t, &lifecycle.Resolution{}, "run", "always applies (no `when` predicate)")
	forceRun = true

	captureStdout(t, func() {
		if err := runTodosRun(todosRunCmd, nil); err != nil {
			t.Fatalf("runTodosRun --force: %v", err)
		}
	})
	if len(*started) != 1 || !(*started)[0].Options.Concurrent {
		t.Fatalf("started = %+v, want one concurrent dispatch", *started)
	}
}

// --commit=false has to actually stop the run committing. Captain's spec merge
// reads an empty slice as "not stated", so a request layer cannot clear a
// lifecycle step's `commits:` — a dispatch that committed anyway would ignore
// an explicit instruction, so the resolution is refused instead.
func TestAssertRunCommitPolicyRefusesASilentlyCommittingRun(t *testing.T) {
	resetTodosRunFlags(t)
	committing := &run.Prepared{
		Step: lifecycle.Step{Name: "run"},
		Resolution: &lifecycle.Resolution{Spec: api.Spec{Workflow: &api.Workflow{
			Commits: []api.Commit{{On: api.CommitOnRun, Gates: api.CommitGatesFull}},
		}}},
	}

	commitAfter = true
	if err := assertRunCommitPolicy(committing); err != nil {
		t.Fatalf("--commit=true rejected the lifecycle's own commit policy: %v", err)
	}

	commitAfter = false
	err := assertRunCommitPolicy(committing)
	if err == nil || !strings.Contains(err.Error(), "--commit=false") {
		t.Fatalf("--commit=false accepted a committing run: %v", err)
	}

	if err := assertRunCommitPolicy(&run.Prepared{
		Step:       lifecycle.Step{Name: "plan"},
		Resolution: &lifecycle.Resolution{},
	}); err != nil {
		t.Fatalf("--commit=false rejected a run that commits nothing: %v", err)
	}
}

func TestValidateTodosRunOptions(t *testing.T) {
	resetTodosRunFlags(t)
	for _, effort := range []string{"", "low", "medium", "high", "xhigh"} {
		todoEffort = effort
		if err := validateTodosRunOptions(); err != nil {
			t.Fatalf("expected effort %q to validate: %v", effort, err)
		}
	}
	todoEffort = "too-much"
	if err := validateTodosRunOptions(); err == nil || !strings.Contains(err.Error(), "--effort") {
		t.Fatalf("expected effort validation error, got %v", err)
	}
}

// workingDirFor points the CLI at dir for the test's duration.
func workingDirFor(t *testing.T, dir string) {
	t.Helper()
	old := workingDir
	workingDir = dir
	t.Cleanup(func() { workingDir = old })
}
