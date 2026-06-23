package main

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons/logger"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/repomap"
)

// shouldCommitAfter reports whether `gavel commit` should run after a TODO's
// agent completes. It only fires for a successful run that the executor did not
// already commit itself: the inline claude executor sets CommitSHA after its own
// commit, so committing again would either duplicate that change set or sweep up
// the user's restored working-tree changes. The cmux executor leaves CommitSHA
// empty — that is the path `--commit` is meant to cover.
func shouldCommitAfter(result *todos.ExecutionResult) bool {
	if !commitAfter || result == nil || !result.Success {
		return false
	}
	return result.CommitSHA == ""
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
// every change the agent made (Stage=all) and committing it in the TODO's
// working directory's git root.
func commitAfterAgent(workDir string, todo *types.TODO) error {
	commitDir := resolveTodoCommitDir(workDir, todo)
	if root := repomap.FindGitRoot(commitDir); root != "" {
		commitDir = root
	}

	cfg, err := verify.LoadGavelConfig(commitDir)
	if err != nil {
		logger.Warnf("Failed to load .gavel.yaml: %v", err)
	}

	result, err := commitpkg.Run(context.Background(), commitpkg.Options{
		WorkDir: commitDir,
		Stage:   commitpkg.StageAll,
		Config:  cfg.Commit,
	})
	if err != nil {
		if errors.Is(err, commitpkg.ErrNothingStaged) {
			logger.Infof("commit: no changes to commit")
			return nil
		}
		return err
	}
	for _, c := range result.Commits {
		logger.Infof("Committed %s: %s", c.Hash, firstLine(c.Message))
	}
	return nil
}

// resolveTodoCommitDir resolves the directory the agent worked in, mirroring how
// the executors derive their working directory from the TODO's CWD.
func resolveTodoCommitDir(workDir string, todo *types.TODO) string {
	if todo != nil && todo.CWD != "" {
		if filepath.IsAbs(todo.CWD) {
			return filepath.Clean(todo.CWD)
		}
		if workDir != "" {
			return filepath.Clean(filepath.Join(workDir, todo.CWD))
		}
		return filepath.Clean(todo.CWD)
	}
	return workDir
}

func firstLine(s string) string {
	return strings.SplitN(s, "\n", 2)[0]
}
