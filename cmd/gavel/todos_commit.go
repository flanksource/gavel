package main

import (
	"context"
	"path/filepath"

	"github.com/flanksource/commons/logger"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// shouldCommitAfter reports whether `gavel commit` should run after a TODO's
// agent completes, honoring the `--commit` flag. Plan runs change nothing, and
// an ask outcome has half-done work by design — committing it is a decision for
// the answer turn.
func shouldCommitAfter(result *todos.ExecutionResult) bool {
	if todosRunMode == types.ModePlan {
		return false
	}
	if result != nil && result.EndStatus == types.EndAsk {
		return false
	}
	return todos.ShouldCommitAfter(result, commitAfter)
}

// maybeCommitAfter runs the `gavel commit` pipeline over the agent's changes
// when `todos run --commit` is set and the executor did not already commit them.
// Verification is the run loop's definition-of-done (the fixture/CEL verdict),
// not a separate post-commit scoring pass — `gavel todos verify` re-runs it
// manually.
func maybeCommitAfter(workDir string, provider todos.Provider, todo *types.TODO, result *todos.ExecutionResult) {
	if shouldCommitAfter(result) {
		if _, err := commitAfterAgent(workDir, todo); err != nil {
			logger.Errorf("commit after agent failed: %v", err)
		}
	}
}

// commitAfterAgent drives the same commit pipeline as `gavel commit`, staging
// every change the agent made (Stage=all) in the TODO's working directory's git
// root, and returns the resulting commit hashes.
func commitAfterAgent(workDir string, todo *types.TODO) ([]string, error) {
	cwd := ""
	meta := commitpkg.AgentRunMetadata{}
	if todo != nil {
		cwd = todo.CWD
		meta.IssueID = todo.ID
		if todo.LLM != nil {
			meta.SessionID = todo.LLM.SessionId
		}
	}
	result, err := commitpkg.RunAfterAgent(context.Background(), workDir, cwd, meta)
	if err != nil {
		return nil, err
	}
	return commitHashes(result), nil
}

func commitHashes(result *commitpkg.Result) []string {
	if result == nil {
		return nil
	}
	var out []string
	for _, c := range result.Commits {
		if c.Hash != "" {
			out = append(out, c.Hash)
		}
	}
	return out
}

// todoWorkDir resolves the directory a TODO's agent worked in (workDir joined
// with the TODO's cwd); git resolves the repository root from there.
func todoWorkDir(workDir string, todo *types.TODO) string {
	if todo == nil || todo.CWD == "" {
		return workDir
	}
	if filepath.IsAbs(todo.CWD) {
		return todo.CWD
	}
	return filepath.Join(workDir, todo.CWD)
}
