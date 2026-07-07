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

func TestTodosProviderFlagRegistered(t *testing.T) {
	flag := todosCmd.PersistentFlags().Lookup("provider")
	if flag == nil {
		t.Fatal("expected todos --provider flag to be registered")
	}
	if flag.DefValue != todos.ProviderGrite {
		t.Fatalf("expected default provider %q, got %q", todos.ProviderGrite, flag.DefValue)
	}
}

func TestTodosRunFlagsRegistered(t *testing.T) {
	for _, name := range []string{"mode", "model", "effort"} {
		if flag := todosRunCmd.Flags().Lookup(name); flag == nil {
			t.Fatalf("expected todos run --%s flag to be registered", name)
		}
	}
}

func TestTodosCreateCommandRegistered(t *testing.T) {
	if !stringSliceContains(todosCmd.Aliases, "todo") {
		t.Fatalf("expected singular todo alias on todos command, got %#v", todosCmd.Aliases)
	}
	if !stringSliceContains(todosCreateCmd.Aliases, "new") {
		t.Fatalf("expected create command to have new alias, got %#v", todosCreateCmd.Aliases)
	}
	for _, name := range []string{"dir", "title", "body", "body-file", "priority", "status"} {
		if flag := todosCreateCmd.Flags().Lookup(name); flag == nil {
			t.Fatalf("expected todos create --%s flag to be registered", name)
		}
	}
}

func TestTodosSyncCommandRegistered(t *testing.T) {
	for _, name := range []string{"dir", "markers", "ignore", "dry-run"} {
		if flag := todosSyncCmd.Flags().Lookup(name); flag == nil {
			t.Fatalf("expected todos sync --%s flag to be registered", name)
		}
	}
}

func TestRunTodosSyncFileProviderDryRun(t *testing.T) {
	workDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(workDir, "main.go"), []byte("// TODO: dry run only\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldProvider := todosProvider
	oldWorkingDir := workingDir
	oldTodosDir := todosDir
	oldMarkers := todosSyncMarkers
	oldIgnore := todosSyncIgnore
	oldDryRun := todosSyncDryRun
	t.Cleanup(func() {
		todosProvider = oldProvider
		workingDir = oldWorkingDir
		todosDir = oldTodosDir
		todosSyncMarkers = oldMarkers
		todosSyncIgnore = oldIgnore
		todosSyncDryRun = oldDryRun
	})

	todosProvider = todos.ProviderFiles
	workingDir = workDir
	todosDir = ""
	todosSyncMarkers = []string{"TODO", "FIXME"}
	todosSyncIgnore = nil
	todosSyncDryRun = true

	if err := runTodosSync(todosSyncCmd, nil); err != nil {
		t.Fatalf("runTodosSync: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workDir, ".todos")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create .todos, stat err=%v", err)
	}
}

func TestRunTodosCreateFileProvider(t *testing.T) {
	workDir := t.TempDir()

	oldProvider := todosProvider
	oldWorkingDir := workingDir
	oldTodosDir := todosDir
	oldTitle := todoCreateTitle
	oldBody := todoCreateBody
	oldBodyFile := todoCreateBodyFile
	oldPriority := todoCreatePriority
	oldStatus := todoCreateStatus
	t.Cleanup(func() {
		todosProvider = oldProvider
		workingDir = oldWorkingDir
		todosDir = oldTodosDir
		todoCreateTitle = oldTitle
		todoCreateBody = oldBody
		todoCreateBodyFile = oldBodyFile
		todoCreatePriority = oldPriority
		todoCreateStatus = oldStatus
	})

	todosProvider = todos.ProviderFiles
	workingDir = workDir
	todosDir = ""
	todoCreateTitle = ""
	todoCreateBody = "Created from the CLI."
	todoCreateBodyFile = ""
	todoCreatePriority = string(types.PriorityHigh)
	todoCreateStatus = string(types.StatusDraft)

	if err := runTodosCreate(todosCreateCmd, []string{"CLI", "todo"}); err != nil {
		t.Fatalf("runTodosCreate: %v", err)
	}

	items, err := todos.NewFileProvider(workDir, "").List(context.Background(), todos.DiscoveryFilters{})
	if err != nil {
		t.Fatalf("list created todos: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("created todo count = %d, want 1", len(items))
	}
	if items[0].Title != "CLI todo" || items[0].Priority != types.PriorityHigh || items[0].Status != types.StatusDraft {
		t.Fatalf("unexpected created todo: %+v", items[0])
	}
	detail, err := todos.NewFileProvider(workDir, "").Get(context.Background(), items[0].FilePath)
	if err != nil {
		t.Fatalf("get created todo: %v", err)
	}
	if !strings.Contains(detail.MarkdownBody, "Created from the CLI.") {
		t.Fatalf("created body missing content: %+v", detail)
	}
}

func TestTodosPlanCommandsRegistered(t *testing.T) {
	if todosPlanRejectCmd.Flags().Lookup("dir") == nil {
		t.Error("expected plan reject --dir flag")
	}
	if todosPlanReviseCmd.Flags().Lookup("feedback") == nil {
		t.Error("expected plan revise --feedback flag")
	}
}

func TestTodosPlanRejectReturnsToPending(t *testing.T) {
	workDir := t.TempDir()
	provider := todos.NewFileProvider(workDir, "")
	created, err := provider.Create(context.Background(), todos.CreateRequest{Title: "Reviewable plan", Status: types.StatusPending})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	review := types.StatusReview
	planPath := "/plans/p.md"
	planNew := types.PlanNew
	if err := todos.UpdateTODOState(created, todos.StateUpdate{Status: &review, PlanPath: &planPath, PlanStatus: &planNew}); err != nil {
		t.Fatalf("seed review: %v", err)
	}

	oldProvider, oldWorkingDir, oldDir := todosProvider, workingDir, todosDir
	todosProvider, workingDir, todosDir = todos.ProviderFiles, workDir, ""
	t.Cleanup(func() { todosProvider, workingDir, todosDir = oldProvider, oldWorkingDir, oldDir })

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

func TestValidateTodosRunOptions(t *testing.T) {
	oldMode, oldEffort := todosMode, todoEffort
	defer func() {
		todosMode = oldMode
		todoEffort = oldEffort
	}()

	for _, mode := range []string{"", "run", "plan", "verify"} {
		todosMode = mode
		todoEffort = "xhigh"
		if err := validateTodosRunOptions(); err != nil {
			t.Fatalf("expected mode %q to validate: %v", mode, err)
		}
	}

	// The legacy mechanism values are rejected — --mode is the todo operation
	// now; the mechanism is --driver.
	for _, mode := range []string{"cmux", "inline", "bad"} {
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

// TestNewDriverConfigModelOverride pins the CLI --model flag beating the
// todo's recorded model (the sdk-specific config builder it replaced is gone).
func TestNewDriverConfigModelOverride(t *testing.T) {
	oldModel := todoModel
	defer func() { todoModel = oldModel }()

	todoModel = "opus"
	todo := &types.TODO{TODOFrontmatter: types.TODOFrontmatter{LLM: &types.LLM{Model: "sonnet"}}}

	cfg := newDriverConfig(drivers.ClaudeHeadless, "/repo", todo)

	if cfg.Model != "opus" {
		t.Fatalf("expected CLI model override, got %q", cfg.Model)
	}
}

func TestEffortDirective(t *testing.T) {
	if got := effortDirective("high"); !strings.Contains(got, "edge cases") {
		t.Fatalf("unexpected high effort directive: %q", got)
	}
	if got := effortDirective(""); !strings.Contains(got, "Think carefully") {
		t.Fatalf("unexpected default effort directive: %q", got)
	}
}

func TestNewTodosProviderRejectsDirWithGrite(t *testing.T) {
	old := todosProvider
	todosProvider = todos.ProviderGrite
	defer func() { todosProvider = old }()

	_, err := newTodosProvider("/repo", ".todos")
	if err == nil || !strings.Contains(err.Error(), "--dir is only supported") {
		t.Fatalf("expected --dir validation error, got %v", err)
	}
}

func TestTodoMatchesArgMatchesGriteIDPrefix(t *testing.T) {
	todo := &types.TODO{ID: "962e67fe4556b8370666f0304281d554", Provider: todos.ProviderGrite}
	if !todoMatchesArg(todo, "962e67fe", "/repo") {
		t.Fatal("expected short grite ID to match")
	}
}

func TestRunTodosListHidesCompletedByDefault(t *testing.T) {
	workDir := seedListTodos(t)

	got := runListWithGlobals(t, workDir, TodosListOptions{})

	if len(got) != 2 {
		t.Fatalf("default list length = %d, want 2: %+v", len(got), got)
	}
	for _, todo := range got {
		if todo.Status == types.StatusCompleted {
			t.Fatalf("default list included completed todo: %+v", todo)
		}
	}
}

func TestRunTodosListAllIncludesCompleted(t *testing.T) {
	workDir := seedListTodos(t)

	got := runListWithGlobals(t, workDir, TodosListOptions{All: true})

	if len(got) != 3 {
		t.Fatalf("--all list length = %d, want 3: %+v", len(got), got)
	}
	if !hasStatus(got, types.StatusCompleted) {
		t.Fatalf("--all list did not include completed todo: %+v", got)
	}
}

func TestRunTodosListStatusCompletedOverridesDefaultHide(t *testing.T) {
	workDir := seedListTodos(t)

	got := runListWithGlobals(t, workDir, TodosListOptions{Status: string(types.StatusCompleted)})

	if len(got) != 1 {
		t.Fatalf("completed list length = %d, want 1: %+v", len(got), got)
	}
	if got[0].Status != types.StatusCompleted {
		t.Fatalf("expected completed todo, got %+v", got[0])
	}
}

func seedListTodos(t *testing.T) string {
	t.Helper()
	workDir := t.TempDir()
	provider := todos.NewFileProvider(workDir, "")
	for _, item := range []struct {
		title  string
		status types.Status
	}{
		{"Pending item", types.StatusPending},
		{"Running item", types.StatusInProgress},
		{"Completed item", types.StatusCompleted},
	} {
		if _, err := provider.Create(t.Context(), todos.CreateRequest{
			Title:  item.title,
			Status: item.status,
		}); err != nil {
			t.Fatalf("seed %q: %v", item.title, err)
		}
	}
	return workDir
}

func runListWithGlobals(t *testing.T, workDir string, opts TodosListOptions) types.TODOS {
	t.Helper()
	oldProvider := todosProvider
	oldWorkingDir := workingDir
	todosProvider = todos.ProviderFiles
	workingDir = workDir
	t.Cleanup(func() {
		todosProvider = oldProvider
		workingDir = oldWorkingDir
	})

	out, err := runTodosList(opts)
	if err != nil {
		t.Fatalf("runTodosList: %v", err)
	}
	got, ok := out.(types.TODOS)
	if !ok {
		t.Fatalf("runTodosList returned %T, want types.TODOS", out)
	}
	return got
}

func hasStatus(todoList types.TODOS, status types.Status) bool {
	for _, todo := range todoList {
		if todo.Status == status {
			return true
		}
	}
	return false
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
