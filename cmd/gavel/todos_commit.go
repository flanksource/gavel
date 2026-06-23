package main

import (
	"context"

	"github.com/flanksource/commons/logger"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// shouldCommitAfter reports whether `gavel commit` should run after a TODO's
// agent completes, honoring the `--commit` flag.
func shouldCommitAfter(result *todos.ExecutionResult) bool {
	return todos.ShouldCommitAfter(result, commitAfter)
}

// maybeCommitAfter runs the full `gavel commit` pipeline over the agent's changes
// when `todos run --commit` is set and the executor did not already commit them.
func maybeCommitAfter(workDir string, todo *types.TODO, result *todos.ExecutionResult) {
	if !shouldCommitAfter(result) {
		return
	}
	if err := commitAfterAgent(workDir, todo); err != nil {
		logger.Errorf("commit after agent failed: %v", err)
	}
}

// commitAfterAgent drives the same commit pipeline as `gavel commit`, staging
// every change the agent made (Stage=all) in the TODO's working directory's git
// root.
func commitAfterAgent(workDir string, todo *types.TODO) error {
	cwd := ""
	if todo != nil {
		cwd = todo.CWD
	}
	return commitpkg.RunAfterAgent(context.Background(), workDir, cwd)
}
