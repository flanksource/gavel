package main

import (
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
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

func TestResolveTodoCommitDir(t *testing.T) {
	abs := filepath.Join(string(filepath.Separator), "abs", "path")
	cases := []struct {
		name    string
		workDir string
		todo    *types.TODO
		want    string
	}{
		{"nil todo falls back to workDir", "/repo", nil, "/repo"},
		{"empty cwd falls back to workDir", "/repo", &types.TODO{}, "/repo"},
		{
			name:    "relative cwd joins workDir",
			workDir: "/repo",
			todo:    &types.TODO{TODOFrontmatter: types.TODOFrontmatter{CWD: "sub/dir"}},
			want:    filepath.Clean("/repo/sub/dir"),
		},
		{
			name:    "absolute cwd is used verbatim",
			workDir: "/repo",
			todo:    &types.TODO{TODOFrontmatter: types.TODOFrontmatter{CWD: abs}},
			want:    abs,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveTodoCommitDir(tc.workDir, tc.todo); got != tc.want {
				t.Fatalf("resolveTodoCommitDir(%q, %+v) = %q, want %q", tc.workDir, tc.todo, got, tc.want)
			}
		})
	}
}
