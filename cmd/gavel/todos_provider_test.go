package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/drivers"
	"github.com/flanksource/gavel/todos/types"
)

func TestTodosHasNoRuntimeProviderSelection(t *testing.T) {
	if flag := todosCmd.PersistentFlags().Lookup("provider"); flag != nil {
		t.Fatalf("unexpected retired --provider flag: %+v", flag)
	}
}

func TestTodosRunFlagsRegistered(t *testing.T) {
	for _, name := range []string{"mode", "model", "effort"} {
		if flag := todosRunCmd.Flags().Lookup(name); flag == nil {
			t.Fatalf("expected todos run --%s flag to be registered", name)
		}
	}
}

func TestTodosVerifyCommandRemoved(t *testing.T) {
	for _, command := range todosCmd.Commands() {
		if command.Name() == "verify" {
			t.Fatalf("unexpected removed todos verify command: %+v", command)
		}
	}
	if err := todosCmd.Args(todosCmd, []string{"verify"}); err == nil {
		t.Fatal("expected removed todos verify invocation to be rejected")
	}
}

func TestTodosCreateCommandRegistered(t *testing.T) {
	if !stringSliceContains(todosCmd.Aliases, "todo") {
		t.Fatalf("expected singular todo alias on todos command, got %#v", todosCmd.Aliases)
	}
	if !stringSliceContains(todosCreateCmd.Aliases, "new") {
		t.Fatalf("expected create command to have new alias, got %#v", todosCreateCmd.Aliases)
	}
	for _, name := range []string{"title", "body", "plan", "verification", "priority", "status"} {
		if flag := todosCreateCmd.Flags().Lookup(name); flag == nil {
			t.Fatalf("expected todos create --%s flag to be registered", name)
		}
	}
	if flag := todosCreateCmd.Flags().Lookup("body-file"); flag != nil {
		t.Fatalf("unexpected retired todos create --body-file flag: %+v", flag)
	}
}

func TestTodosSyncCommandRegistered(t *testing.T) {
	for _, name := range []string{"markers", "ignore", "dry-run"} {
		if flag := todosSyncCmd.Flags().Lookup(name); flag == nil {
			t.Fatalf("expected todos sync --%s flag to be registered", name)
		}
	}
}

func TestRunTodosSyncRuntimeProviderDryRun(t *testing.T) {
	workDir := t.TempDir()
	stubRuntimeWithTestProvider(t)
	if err := os.WriteFile(filepath.Join(workDir, "main.go"), []byte("// TODO: dry run only\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldWorkingDir := workingDir
	oldMarkers := todosSyncMarkers
	oldIgnore := todosSyncIgnore
	oldDryRun := todosSyncDryRun
	t.Cleanup(func() {
		workingDir = oldWorkingDir
		todosSyncMarkers = oldMarkers
		todosSyncIgnore = oldIgnore
		todosSyncDryRun = oldDryRun
	})

	workingDir = workDir
	todosSyncMarkers = []string{"TODO", "FIXME"}
	todosSyncIgnore = nil
	todosSyncDryRun = true

	if err := runTodosSync(todosSyncCmd, nil); err != nil {
		t.Fatalf("runTodosSync: %v", err)
	}
	if items, err := testProviderFor(workDir).List(t.Context(), todos.DiscoveryFilters{}); err != nil || len(items) != 0 {
		t.Fatalf("dry-run created native TODOs: items=%v err=%v", items, err)
	}
}

func TestRunTodosCreateRuntimeProvider(t *testing.T) {
	workDir := t.TempDir()
	stubRuntimeWithTestProvider(t)

	oldWorkingDir := workingDir
	oldTitle := todoCreateTitle
	oldBody := todoCreateBody
	oldPlan := todoCreatePlan
	oldVerification := todoCreateVerification
	oldPriority := todoCreatePriority
	oldStatus := todoCreateStatus
	t.Cleanup(func() {
		workingDir = oldWorkingDir
		todoCreateTitle = oldTitle
		todoCreateBody = oldBody
		todoCreatePlan = oldPlan
		todoCreateVerification = oldVerification
		todoCreatePriority = oldPriority
		todoCreateStatus = oldStatus
	})

	workingDir = workDir
	todoCreateTitle = ""
	todoCreateBody = "Created from the CLI."
	todoCreatePlan = ""
	todoCreateVerification = ""
	todoCreatePriority = string(types.PriorityHigh)
	todoCreateStatus = string(types.StatusDraft)

	if err := runTodosCreate(todosCreateCmd, []string{"CLI", "todo"}); err != nil {
		t.Fatalf("runTodosCreate: %v", err)
	}

	items, err := testProviderFor(workDir).List(context.Background(), todos.DiscoveryFilters{})
	if err != nil {
		t.Fatalf("list created todos: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("created todo count = %d, want 1", len(items))
	}
	if items[0].Title != "CLI todo" || items[0].Priority != types.PriorityHigh || items[0].Status != types.StatusDraft {
		t.Fatalf("unexpected created todo: %+v", items[0])
	}
	detail, err := testProviderFor(workDir).Get(context.Background(), items[0].ID)
	if err != nil {
		t.Fatalf("get created todo: %v", err)
	}
	if !strings.Contains(detail.MarkdownBody, "Created from the CLI.") {
		t.Fatalf("created body missing content: %+v", detail)
	}
}

func TestTodosPlanCommandsRegistered(t *testing.T) {
	if todosPlanRejectCmd.Flags().Lookup("dir") != nil {
		t.Error("unexpected retired plan reject --dir flag")
	}
	if todosPlanReviseCmd.Flags().Lookup("feedback") == nil {
		t.Error("expected plan revise --feedback flag")
	}
}

func TestTodosPlanRejectReturnsToPending(t *testing.T) {
	workDir := t.TempDir()
	provider := &testPlanReviewProvider{testTODOProvider: testProviderFor(workDir)}
	oldOpen := openRuntimeTodosProvider
	openRuntimeTodosProvider = func(context.Context, string) (todos.Provider, error) { return provider, nil }
	t.Cleanup(func() { openRuntimeTodosProvider = oldOpen })
	created, err := provider.Create(context.Background(), todos.CreateRequest{Title: "Reviewable plan", Status: types.StatusPending})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	review := types.StatusReview
	planPath := "/plans/p.md"
	planNew := types.PlanNew
	if err := provider.UpdateState(t.Context(), created, todos.StateUpdate{Status: &review, PlanPath: &planPath, PlanStatus: &planNew}); err != nil {
		t.Fatalf("seed review: %v", err)
	}

	oldWorkingDir := workingDir
	workingDir = workDir
	t.Cleanup(func() { workingDir = oldWorkingDir })

	if err := runTodosPlanReject(todosPlanRejectCmd, []string{"Reviewable plan"}); err != nil {
		t.Fatalf("reject: %v", err)
	}
	reloaded, err := provider.Get(context.Background(), created.FilePath)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.Status != types.StatusPending {
		t.Errorf("status = %s, want pending", reloaded.Status)
	}
	if reloaded.PlanPath != "" {
		t.Errorf("plan path not cleared: %q", reloaded.PlanPath)
	}
}

type testPlanReviewProvider struct {
	*testTODOProvider
}

func (p *testPlanReviewProvider) ApprovePlan(ctx context.Context, todo *types.TODO, _, _ string) (*types.TODO, error) {
	return p.Get(ctx, todo.FilePath)
}

func (p *testPlanReviewProvider) RejectPlan(ctx context.Context, todo *types.TODO, _, _ string) (*types.TODO, error) {
	pending := types.StatusPending
	clearedPath := ""
	clearedStatus := types.PlanStatus("")
	if err := p.UpdateState(ctx, todo, todos.StateUpdate{Status: &pending, PlanPath: &clearedPath, PlanStatus: &clearedStatus}); err != nil {
		return nil, err
	}
	return p.Get(ctx, todo.FilePath)
}

func (p *testPlanReviewProvider) RequestPlanRevision(ctx context.Context, todo *types.TODO, _, _ string) (*types.TODO, error) {
	return p.Get(ctx, todo.FilePath)
}

func TestValidateTodosRunOptions(t *testing.T) {
	oldMode, oldEffort := todosMode, todoEffort
	defer func() {
		todosMode = oldMode
		todoEffort = oldEffort
	}()

	for _, mode := range []string{"", "run", "plan"} {
		todosMode = mode
		todoEffort = "xhigh"
		if err := validateTodosRunOptions(); err != nil {
			t.Fatalf("expected mode %q to validate: %v", mode, err)
		}
	}

	// The legacy mechanism values are rejected — --mode is the todo operation
	// now; the mechanism is --driver.
	for _, mode := range []string{"verify", "cmux", "inline", "bad"} {
		todosMode = mode
		if err := validateTodosRunOptions(); err == nil || !strings.Contains(err.Error(), "run mode") {
			t.Fatalf("expected mode %q to be rejected, got %v", mode, err)
		}
	}

	todosMode = "run"
	todoEffort = "too-much"
	if err := validateTodosRunOptions(); err == nil || !strings.Contains(err.Error(), "--effort") {
		t.Fatalf("expected effort validation error, got %v", err)
	}
}

// Review/ask todos must survive the CLI's post-run status cleanup — only
// in_progress gets reconciled.
func TestCleanupTODOStatusKeepsReviewAndAsk(t *testing.T) {
	for _, status := range []types.Status{types.StatusReview, types.StatusAsk} {
		todo := &types.TODO{TODOFrontmatter: types.TODOFrontmatter{Status: status}}
		cleanupTODOStatus(todo, &todos.ExecutionResult{Success: true})
		if todo.Status != status {
			t.Errorf("cleanup rewrote %s to %s", status, todo.Status)
		}
	}
}

// TestNewAgentRunConfigModelOverride pins the CLI --model flag beating the
// todo's recorded model in the canonical Spec.
func TestNewAgentRunConfigModelOverride(t *testing.T) {
	oldModel := todoModel
	defer func() { todoModel = oldModel }()

	todoModel = "opus"
	todo := &types.TODO{TODOFrontmatter: types.TODOFrontmatter{LLM: &types.LLM{Model: "sonnet"}}}

	cfg, err := newAgentRunConfig(context.Background(), drivers.Cli, "/repo", todo, nil)
	if err != nil {
		t.Fatalf("newAgentRunConfig: %v", err)
	}

	if cfg.Name != "opus" {
		t.Fatalf("expected CLI model override, got %q", cfg.Name)
	}
}

func TestNewAgentRunConfigLoadsPlanThroughActiveDBProvider(t *testing.T) {
	oldMode := todosRunMode
	todosRunMode = types.ModeRun
	t.Cleanup(func() { todosRunMode = oldMode })

	provider := &planContentSpy{content: "# Approved durable plan"}
	todo := &types.TODO{
		ID: "962e67fe-4556-b837-0666-f0304281d554",
		TODOFrontmatter: types.TODOFrontmatter{
			PlanPath: "/definitely/not/a/runtime/plan.md",
		},
	}
	cfg, err := newAgentRunConfig(context.Background(), drivers.Cli, "/repo", todo, provider)
	if err != nil {
		t.Fatalf("newAgentRunConfig: %v", err)
	}
	if cfg.ExistingPlan != provider.content {
		t.Fatalf("ExistingPlan = %q, want durable provider content", cfg.ExistingPlan)
	}
	if provider.calls != 1 || provider.mode != types.ModeRun || provider.todo != todo {
		t.Fatalf("PlanMarkdown call = calls:%d mode:%s todo:%p, want one run-mode call for %p", provider.calls, provider.mode, provider.todo, todo)
	}
}

type planContentSpy struct {
	todos.Provider
	content string
	calls   int
	mode    types.RunMode
	todo    *types.TODO
}

func (p *planContentSpy) PlanMarkdown(_ context.Context, todo *types.TODO, mode types.RunMode) (string, error) {
	p.calls++
	p.todo = todo
	p.mode = mode
	return p.content, nil
}

func TestEffortDirective(t *testing.T) {
	if got := effortDirective("high"); !strings.Contains(got, "edge cases") {
		t.Fatalf("unexpected high effort directive: %q", got)
	}
	if got := effortDirective(""); !strings.Contains(got, "Think carefully") {
		t.Fatalf("unexpected default effort directive: %q", got)
	}
}

func TestNewTodosProviderOpensNativeRuntime(t *testing.T) {
	oldOpen := openRuntimeTodosProvider
	t.Cleanup(func() {
		openRuntimeTodosProvider = oldOpen
	})

	opened := 0
	stub := testProviderFor("/repo")
	openRuntimeTodosProvider = func(_ context.Context, workDir string) (todos.Provider, error) {
		opened++
		if workDir != "/repo" {
			t.Fatalf("runtime opener workDir = %q, want /repo", workDir)
		}
		return stub, nil
	}
	provider, err := newTodosProvider("/repo")
	if err != nil {
		t.Fatalf("newTodosProvider: %v", err)
	}
	if provider != stub {
		t.Fatalf("newTodosProvider returned %T, want runtime stub", provider)
	}
	if opened != 1 {
		t.Fatalf("runtime opener calls = %d, want 1", opened)
	}
}

func TestResolveRequestedTODOsUsesDirectGetForImportedAlias(t *testing.T) {
	want := &types.TODO{
		ID:       "962e67fe-4556-b837-0666-f0304281d554",
		Provider: todos.ProviderDB,
		TODOFrontmatter: types.TODOFrontmatter{
			Title:  "Imported issue",
			Status: types.StatusPending,
		},
	}
	provider := &referenceSpyProvider{todo: want}
	got, err := resolveRequestedTODOs(context.Background(), provider, []string{"imported-alias"}, todos.DiscoveryFilters{})
	if err != nil {
		t.Fatalf("resolveRequestedTODOs: %v", err)
	}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("resolved TODOs = %#v, want direct alias target", got)
	}
	if len(provider.getRefs) != 1 || provider.getRefs[0] != "imported-alias" {
		t.Fatalf("Get refs = %#v, want imported alias", provider.getRefs)
	}
	if provider.listCalls != 0 {
		t.Fatalf("List calls = %d, want 0 for a directly resolvable alias", provider.listCalls)
	}
}

func TestRunTodosRunRejectsGroupedNativeExecution(t *testing.T) {
	oldOpen := openRuntimeTodosProvider
	oldWorkingDir := workingDir
	oldGroupBy := groupBy
	oldMode := todosMode
	oldRunMode := todosRunMode
	t.Cleanup(func() {
		openRuntimeTodosProvider = oldOpen
		workingDir = oldWorkingDir
		groupBy = oldGroupBy
		todosMode = oldMode
		todosRunMode = oldRunMode
	})

	provider := &singleIssueRuntimeStub{todo: &types.TODO{
		ID:       "962e67fe-4556-b837-0666-f0304281d554",
		Provider: todos.ProviderDB,
		TODOFrontmatter: types.TODOFrontmatter{
			Title:  "One native issue",
			Status: types.StatusPending,
		},
	}}
	openRuntimeTodosProvider = func(context.Context, string) (todos.Provider, error) { return provider, nil }
	workingDir = t.TempDir()
	groupBy = todos.GroupByRepo
	todosMode = string(types.ModeRun)

	err := runTodosRun(todosRunCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "--group-by is not supported") || !strings.Contains(err.Error(), "one issue at a time") {
		t.Fatalf("runTodosRun error = %v, want native grouped-execution rejection", err)
	}
}

type referenceSpyProvider struct {
	todos.Provider
	todo      *types.TODO
	getRefs   []string
	listCalls int
}

type singleIssueRuntimeStub struct {
	todos.Provider
	todo *types.TODO
}

func (p *singleIssueRuntimeStub) List(context.Context, todos.DiscoveryFilters) (types.TODOS, error) {
	return types.TODOS{p.todo}, nil
}

func (p *singleIssueRuntimeStub) SupportsGroupedExecution() bool { return false }

func (p *referenceSpyProvider) Get(_ context.Context, ref string) (*types.TODO, error) {
	p.getRefs = append(p.getRefs, ref)
	if ref != "imported-alias" {
		return nil, os.ErrNotExist
	}
	return p.todo, nil
}

func (p *referenceSpyProvider) List(context.Context, todos.DiscoveryFilters) (types.TODOS, error) {
	p.listCalls++
	return types.TODOS{p.todo}, nil
}

func stubRuntimeWithTestProvider(t *testing.T) {
	t.Helper()
	old := openRuntimeTodosProvider
	openRuntimeTodosProvider = func(_ context.Context, workDir string) (todos.Provider, error) {
		return testProviderFor(workDir), nil
	}
	t.Cleanup(func() { openRuntimeTodosProvider = old })
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
