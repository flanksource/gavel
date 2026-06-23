package main

import (
	"testing"

	"github.com/flanksource/gavel/todos"
)

func TestTodosRunCommitFlagRegistered(t *testing.T) {
	flag := todosRunCmd.Flags().Lookup("commit")
	if flag == nil {
		t.Fatal("expected todos run --commit flag to be registered")
	}
	if flag.DefValue != "true" {
		t.Fatalf("expected --commit default true, got %q", flag.DefValue)
	}
}

func TestShouldCommitAfter(t *testing.T) {
	old := commitAfter
	t.Cleanup(func() { commitAfter = old })

	cases := []struct {
		name    string
		enabled bool
		result  *todos.ExecutionResult
		want    bool
	}{
		{"flag disabled", false, &todos.ExecutionResult{Success: true}, false},
		{"nil result", true, nil, false},
		{"failed run", true, &todos.ExecutionResult{Success: false}, false},
		{"already committed", true, &todos.ExecutionResult{Success: true, CommitSHA: "abc1234"}, false},
		{"success uncommitted", true, &todos.ExecutionResult{Success: true}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			commitAfter = tc.enabled
			if got := shouldCommitAfter(tc.result); got != tc.want {
				t.Fatalf("shouldCommitAfter(%+v) = %v, want %v", tc.result, got, tc.want)
			}
		})
	}
}
