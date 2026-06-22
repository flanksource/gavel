package main

import (
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
