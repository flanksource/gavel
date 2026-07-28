package main

import (
	"path/filepath"

	"github.com/flanksource/gavel/todos/types"
)

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
