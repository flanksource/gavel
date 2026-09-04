package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos/run"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/google/uuid"
)

type todoRunStopRequest struct {
	Ref         string    `json:"ref"`
	PromptRunID uuid.UUID `json:"promptRunId"`
}

func (s *Server) handleTodoRunStop(w http.ResponseWriter, r *http.Request) {
	var payload todoRunStopRequest
	if err := decodeTodoRequest(r, &payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
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
	promptRun, err := detailProvider.Captain().GetPromptRun(r.Context(), payload.PromptRunID)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	if promptRun.State != captaindb.PromptRunStatePending && promptRun.State != captaindb.PromptRunStateRunning && promptRun.State != captaindb.PromptRunStateWaiting {
		writeTodoError(w, http.StatusConflict, fmt.Errorf("attempt is already %s", promptRun.State))
		return
	}
	// Cancel the run's outstanding tool approvals before cancelling the run. They
	// outlive the process that raised them, so a stopped run would otherwise leave
	// rows a dashboard still offers to answer — and answering one would unblock a
	// broker that is no longer there to hear it.
	if err := detailProvider.Captain().CancelPendingTurnRequests(
		r.Context(), promptRun.SessionID, payload.PromptRunID, "run stopped"); err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	status := "stopping"
	if err := todoRuns().Stop(payload.PromptRunID); err != nil {
		// A run this process does not own is either driven by another live
		// process — which this one cannot cancel — or abandoned by a dispatcher
		// that exited. Only the second is this request's to end, and only the
		// ownership claim can tell them apart.
		if !errors.Is(err, run.ErrNotOwned) {
			writeTodoError(w, http.StatusConflict, err)
			return
		}
		reclaimer, ok := provider.(*todoruntime.Provider)
		if !ok {
			writeTodoError(w, http.StatusConflict, err)
			return
		}
		reclaimed, reason, reclaimErr := reclaimer.ReclaimRun(r.Context(), payload.PromptRunID)
		if reclaimErr != nil {
			writeTodoError(w, http.StatusInternalServerError, reclaimErr)
			return
		}
		if !reclaimed {
			writeTodoError(w, http.StatusConflict, fmt.Errorf("%w: %s", err, reason))
			return
		}
		status = "reclaimed"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(map[string]any{"status": status, "promptRunId": payload.PromptRunID}); err != nil {
		panic(fmt.Errorf("encode TODO stop response: %w", err))
	}
}
