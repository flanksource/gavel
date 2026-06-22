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
