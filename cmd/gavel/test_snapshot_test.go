package main

import (
	"testing"

	"github.com/flanksource/gavel/testrunner"
)

func TestSnapshotArgsOmitRetiredTODOSyncFields(t *testing.T) {
	args := snapshotArgs(testrunner.RunOptions{})
	for _, retired := range []string{"sync_todos", "todos_dir", "todo_template"} {
		if _, exists := args[retired]; exists {
			t.Errorf("snapshot args still contain retired %q field", retired)
		}
	}
}
