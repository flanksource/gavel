package ui

import (
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	"net/http"
)

type todoNewTriageResponse struct {
	Status      string `json:"status"`
	SessionID   string `json:"sessionId,omitempty"`
	PromptRunID string `json:"promptRunId,omitempty"`
	Error       string `json:"error,omitempty"`
}

func (s *Server) startNewTodoTriage(r *http.Request, provider todos.Provider, source todoSource, todo *types.TODO) *todoNewTriageResponse {
	req := run.Request{Provider: provider, Registry: todoRuns(), Todo: todo, Dir: source.Dir, Options: run.Options{Step: "triage.new", Host: lifecycle.HostDashboard}, Approvals: true}
	todo.MarkdownBody = todos.AbsolutizeAttachmentURLs(todo.MarkdownBody, requestOrigin(r))
	prepared, err := run.Resolve(r.Context(), req)
	if err == nil {
		err = validateTodoRunRuntime(prepared.Resolution.Spec)
	}
	if err != nil {
		return &todoNewTriageResponse{Status: "failed", Error: err.Error()}
	}
	req.Prepared = prepared
	started, err := run.Start(req)
	if err != nil {
		return &todoNewTriageResponse{Status: "failed", Error: err.Error()}
	}
	response := &todoNewTriageResponse{Status: started.Status, SessionID: started.SessionID}
	if started.PromptRunID != uuid.Nil {
		response.PromptRunID = started.PromptRunID.String()
	}
	return response
}
