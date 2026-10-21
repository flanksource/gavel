package ui

import (
	"net/http"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

// todoNewRunResponse is the run a create request started on the saved todo.
type todoNewRunResponse struct {
	Status      string `json:"status"`
	SessionID   string `json:"sessionId,omitempty"`
	PromptRunID string `json:"promptRunId,omitempty"`
	Error       string `json:"error,omitempty"`
}

// startNewTodoStep starts the named lifecycle step on a todo the request just
// saved. A failure is reported in the response rather than failing the create:
// the todo exists, and the caller retries the step on it.
func (s *Server) startNewTodoStep(r *http.Request, provider todos.Provider, source todoSource, todo *types.TODO, step string) *todoNewRunResponse {
	req := run.Request{Provider: provider, Registry: todoRuns(), Todo: todo, Dir: source.Dir, Options: run.Options{Step: step, Host: lifecycle.HostDashboard}, Approvals: true}
	if err := s.attachRemote(r.Context(), source, &req); err != nil {
		return &todoNewRunResponse{Status: "failed", Error: err.Error()}
	}
	todo.MarkdownBody = todos.AbsolutizeAttachmentURLs(todo.MarkdownBody, requestOrigin(r))
	prepared, err := run.Resolve(r.Context(), req)
	if err == nil {
		err = validateTodoRunRuntime(prepared.Resolution.Spec)
	}
	if err != nil {
		return &todoNewRunResponse{Status: "failed", Error: err.Error()}
	}
	req.Prepared = prepared
	started, err := run.Start(req)
	if err != nil {
		return &todoNewRunResponse{Status: "failed", Error: err.Error()}
	}
	response := &todoNewRunResponse{Status: started.Status, SessionID: started.SessionID}
	if started.PromptRunID != uuid.Nil {
		response.PromptRunID = started.PromptRunID.String()
	}
	return response
}
