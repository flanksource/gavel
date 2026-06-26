package ui

import (
	"testing"
	"time"

	"github.com/flanksource/gavel/todos/cmux"
	"github.com/flanksource/gavel/todos/types"
)

// failedTodoWithSession builds a failed todo carrying the given session id, the
// state reconcileSessionStatus reconciles against a live agent session.
func failedTodoWithSession(sessionID string) *types.TODO {
	return &types.TODO{
		TODOFrontmatter: types.TODOFrontmatter{
			Status: types.StatusFailed,
			LLM:    &types.LLM{SessionId: sessionID},
		},
	}
}

func TestReconcileSessionStatusFlipsFailedWhileSessionExecuting(t *testing.T) {
	dir := t.TempDir()
	const sessionID = "reconcile-live-0001"
	// A live tailer feeding the session marks it in progress, so a failed todo
	// whose session is executing again must read as in-progress.
	acc := cmux.GlobalSessionStats().Begin(sessionID, "claude", "claude-opus-4-8", "medium", time.Now())
	defer acc.Finish()

	todo := failedTodoWithSession(sessionID)
	reconcileSessionStatus(todo, dir)

	if todo.Status != types.StatusInProgress {
		t.Fatalf("Status = %q, want %q while the session is executing", todo.Status, types.StatusInProgress)
	}
}

func TestReconcileSessionStatusLeavesFailedWhenNoLiveSession(t *testing.T) {
	dir := t.TempDir()
	// A recorded session id with no live tailer and no on-disk log is the normal
	// "session is over" case: the failed status must stand.
	todo := failedTodoWithSession("reconcile-missing-0002")
	reconcileSessionStatus(todo, dir)

	if todo.Status != types.StatusFailed {
		t.Fatalf("Status = %q, want %q when no session is executing", todo.Status, types.StatusFailed)
	}
}

func TestReconcileSessionStatusOnlyTouchesFailedTodos(t *testing.T) {
	dir := t.TempDir()
	const sessionID = "reconcile-nonfailed-0003"
	acc := cmux.GlobalSessionStats().Begin(sessionID, "claude", "claude-opus-4-8", "medium", time.Now())
	defer acc.Finish()

	// A completed todo whose session is live must not be dragged back to running;
	// the reconcile is scoped to failed todos restarting their session.
	todo := &types.TODO{TODOFrontmatter: types.TODOFrontmatter{
		Status: types.StatusCompleted,
		LLM:    &types.LLM{SessionId: sessionID},
	}}
	reconcileSessionStatus(todo, dir)

	if todo.Status != types.StatusCompleted {
		t.Fatalf("Status = %q, want %q (only failed todos reconcile)", todo.Status, types.StatusCompleted)
	}
}

func TestReconcileSessionStatusNoSessionIsNoop(t *testing.T) {
	todo := &types.TODO{TODOFrontmatter: types.TODOFrontmatter{Status: types.StatusFailed}}
	reconcileSessionStatus(todo, t.TempDir())

	if todo.Status != types.StatusFailed {
		t.Fatalf("Status = %q, want %q when the todo has no session", todo.Status, types.StatusFailed)
	}
}
