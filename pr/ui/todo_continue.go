package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
)

// answeredAttempt is the attempt an answer resumes: the one the answered
// session belongs to, whose link names the step it was dispatched for. The
// phase index cannot answer this — it lists a run step that reached
// verification under both run and verify, one run under two phases.
//
// The attempt must be the todo's active one. A session of an older attempt is
// a question that is no longer the todo's to answer; resuming it would continue
// a turn the todo has already moved past.
func answeredAttempt(ctx context.Context, provider todos.Provider, todo *types.TODO, sessionID string, active *captaindb.PromptRun) (run.SessionAttempt, int, error) {
	attempts, ok := provider.(run.SessionAttemptProvider)
	if !ok {
		return run.SessionAttempt{}, http.StatusNotImplemented, errors.New("answering a session requires native TODO storage")
	}
	attempt, err := attempts.SessionAttempt(ctx, todo, sessionID)
	switch {
	case errors.Is(err, native.ErrNotFound):
		return run.SessionAttempt{}, http.StatusNotFound, err
	case err != nil:
		return run.SessionAttempt{}, http.StatusInternalServerError, err
	}
	if active == nil || active.ID != attempt.PromptRunID {
		return run.SessionAttempt{}, http.StatusConflict, fmt.Errorf(
			"session %s is not the active attempt of todo %s; its question is no longer answerable", sessionID, todos.TODOReference(todo))
	}
	return attempt, http.StatusOK, nil
}
