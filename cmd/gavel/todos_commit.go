package main

import (
	"context"
	"fmt"
	"path/filepath"

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

// maybeCommitAfter runs the `gavel commit` pipeline over the agent's changes
// when `todos run --commit` is set and the executor did not already commit them,
// then runs issue verification over the resulting commits when `--verify` is set.
func maybeCommitAfter(workDir string, provider todos.Provider, todo *types.TODO, result *todos.ExecutionResult) {
	var hashes []string
	if shouldCommitAfter(result) {
		committed, err := commitAfterAgent(workDir, todo)
		if err != nil {
			logger.Errorf("commit after agent failed: %v", err)
		}
		hashes = committed
	}
	if verifyAfter && result != nil && result.Success {
		maybeVerifyAfter(workDir, provider, todo, hashes)
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

// maybeVerifyAfter scores the just-committed work against the issue's spec and
// stored acceptance criteria, printing the verdict. It is advisory and logs
// (rather than fails) when there is nothing to verify.
func maybeVerifyAfter(workDir string, provider todos.Provider, todo *types.TODO, hashes []string) {
	result, err := todos.RunIssueVerification(context.Background(), provider, todo, todos.VerifyOptions{
		WorkDir: todoWorkDir(workDir, todo),
		Model:   verifyModel,
		Commits: hashes,
	})
	if err != nil {
		logger.Warnf("issue verification skipped: %v", err)
		return
	}
	fmt.Println(result.Pretty().ANSI())
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
