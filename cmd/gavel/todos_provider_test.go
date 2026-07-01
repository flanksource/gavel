package main

import (
	"context"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos"
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

func TestValidateTodosRunOptions(t *testing.T) {
	oldMode, oldEffort := todosMode, todoEffort
	defer func() {
		todosMode = oldMode
		todoEffort = oldEffort
	}()

	todosMode = "cmux"
	todoEffort = "high"
	if err := validateTodosRunOptions(); err != nil {
		t.Fatalf("expected cmux/high to validate: %v", err)
	}

	todosMode = "bad"
	if err := validateTodosRunOptions(); err == nil || !strings.Contains(err.Error(), "--mode") {
		t.Fatalf("expected mode validation error, got %v", err)
	}

	todosMode = "inline"
	todoEffort = "too-much"
	if err := validateTodosRunOptions(); err == nil || !strings.Contains(err.Error(), "--effort") {
		t.Fatalf("expected effort validation error, got %v", err)
	}
}

func TestNewClaudeConfigModelOverride(t *testing.T) {
	oldModel := todoModel
	defer func() { todoModel = oldModel }()

	todoModel = "opus"
	todo := &types.TODO{TODOFrontmatter: types.TODOFrontmatter{LLM: &types.LLM{Model: "sonnet"}}}

	cfg := newClaudeConfig("/repo", todo)

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
