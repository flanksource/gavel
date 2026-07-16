package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

var (
	errTodoRunStopped      = todos.ErrExecutionCancelled
	errTodoRunAlreadyOwned = errors.New("todo already has a dashboard-owned run")
	errTodoRunNotOwned     = errors.New("todo run is not owned by this dashboard process")
	errTodoRunNotStoppable = errors.New("todo run driver cannot be stopped safely")
	errTodoRunStopping     = errors.New("todo run is already stopping")
)

type todoRunControlStatus struct {
	CanStop  bool
	Stopping bool
}

type activeTodoRun struct {
	cancel    context.CancelCauseFunc
	stoppable bool
	stopping  bool
}

type todoRunRegistry struct {
	mu      sync.Mutex
	byIssue map[uuid.UUID]*activeTodoRun
}

func newTodoRunRegistry() *todoRunRegistry {
	return &todoRunRegistry{byIssue: map[uuid.UUID]*activeTodoRun{}}
}

func (r *todoRunRegistry) register(issueIDs []uuid.UUID, stoppable bool, cancel context.CancelCauseFunc) (func(), error) {
	if cancel == nil {
		return nil, errors.New("todo run cancel function is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.byIssue == nil {
		r.byIssue = map[uuid.UUID]*activeTodoRun{}
	}
	for _, issueID := range issueIDs {
		if r.byIssue[issueID] != nil {
			return nil, fmt.Errorf("%w: %s", errTodoRunAlreadyOwned, issueID)
		}
	}
	active := &activeTodoRun{cancel: cancel, stoppable: stoppable}
	for _, issueID := range issueIDs {
		r.byIssue[issueID] = active
	}
	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		for _, issueID := range issueIDs {
			if r.byIssue[issueID] == active {
				delete(r.byIssue, issueID)
			}
		}
	}, nil
}

func (r *todoRunRegistry) status(issueID uuid.UUID) todoRunControlStatus {
	r.mu.Lock()
	defer r.mu.Unlock()
	active := r.byIssue[issueID]
	if active == nil || !active.stoppable {
		return todoRunControlStatus{}
	}
	return todoRunControlStatus{CanStop: !active.stopping, Stopping: active.stopping}
}

func (r *todoRunRegistry) stop(issueID uuid.UUID) error {
	r.mu.Lock()
	active := r.byIssue[issueID]
	switch {
	case active == nil:
		r.mu.Unlock()
		return errTodoRunNotOwned
	case !active.stoppable:
		r.mu.Unlock()
		return errTodoRunNotStoppable
	case active.stopping:
		r.mu.Unlock()
		return errTodoRunStopping
	}
	active.stopping = true
	cancel := active.cancel
	r.mu.Unlock()
	cancel(errTodoRunStopped)
	return nil
}

type todoRunStopRequest struct {
	Ref         string    `json:"ref"`
	PromptRunID uuid.UUID `json:"promptRunId"`
}

func (s *Server) handleTodoRunStop(w http.ResponseWriter, r *http.Request) {
	var payload todoRunStopRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid stop request: %w", err))
		return
	}
	if payload.Ref == "" || payload.PromptRunID == uuid.Nil {
		writeTodoError(w, http.StatusBadRequest, errors.New("ref and promptRunId are required"))
		return
	}
	dir := s.resolveTodoDir(r.URL.Query().Get("dir"))
	provider, err := openTodoProvider(r.Context(), dir)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	detailProvider, ok := provider.(sessionDetailProvider)
	if !ok || detailProvider.Captain() == nil || detailProvider.Repository() == nil {
		writeTodoError(w, http.StatusNotImplemented, errors.New("stopping attempts requires native TODO storage"))
		return
	}
	todo, err := provider.Get(r.Context(), payload.Ref)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	issueID, err := uuid.Parse(todo.ID)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, fmt.Errorf("native TODO has invalid ID %q: %w", todo.ID, err))
		return
	}
	issue, err := detailProvider.Repository().GetIssue(r.Context(), issueID)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	if issue.ActivePromptRunID == nil || *issue.ActivePromptRunID != payload.PromptRunID {
		writeTodoError(w, http.StatusConflict, errors.New("attempt is no longer the active prompt run"))
		return
	}
	run, err := detailProvider.Captain().GetPromptRun(r.Context(), payload.PromptRunID)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	if run.State != captaindb.PromptRunStatePending && run.State != captaindb.PromptRunStateRunning && run.State != captaindb.PromptRunStateWaiting {
		writeTodoError(w, http.StatusConflict, fmt.Errorf("attempt is already %s", run.State))
		return
	}
	if err := s.todoRuns.stop(issueID); err != nil {
		writeTodoError(w, http.StatusConflict, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(map[string]any{"status": "stopping", "promptRunId": payload.PromptRunID}); err != nil {
		panic(fmt.Errorf("encode TODO stop response: %w", err))
	}
}

func todoRunIssueIDs(todoList []*types.TODO) []uuid.UUID {
	issueIDs := make([]uuid.UUID, 0, len(todoList))
	for _, todo := range todoList {
		if todo == nil {
			continue
		}
		if issueID, err := uuid.Parse(todo.ID); err == nil {
			issueIDs = append(issueIDs, issueID)
		}
	}
	return issueIDs
}
