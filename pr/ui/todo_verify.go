package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// todoCriteriaPayload is the shared request body for acceptance-criteria
// endpoints: a todo reference, the full criteria list, and an optional model
// override for AI drafting.
type todoCriteriaPayload struct {
	Dir      string                      `json:"dir,omitempty"`
	Ref      string                      `json:"ref,omitempty"`
	Model    string                      `json:"model,omitempty"`
	Criteria []types.AcceptanceCriterion `json:"criteria,omitempty"`
}

// loadTodoForWrite resolves the provider + source for a todo mutation and loads
// the todo, mirroring handleTodoPatch's resolution. The returned status is the
// HTTP code to use when err is non-nil.
func (s *Server) loadTodoForWrite(r *http.Request, dir, ref string) (todos.Provider, todoSource, *types.TODO, int, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		ref = strings.TrimSpace(r.URL.Query().Get("ref"))
	}
	if ref == "" {
		return nil, todoSource{}, nil, http.StatusBadRequest, fmt.Errorf("ref is required")
	}
	source := todoSourceFromRequest(r)
	if dir != "" {
		source.Dir = dir
	}
	provider, source, todo, err := s.resolveTodoReference(r.Context(), source, ref)
	if err != nil {
		return nil, todoSource{}, nil, http.StatusNotFound, err
	}
	return provider, source, todo, 0, nil
}

// handleTodoCriteria saves a todo's acceptance criteria, rewriting its
// "## Acceptance Criteria" section, and returns the refreshed todo.
func (s *Server) handleTodoCriteria(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoCriteriaPayload
	if err := decodeTodoRequest(r, &payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	provider, source, todo, status, err := s.loadTodoForWrite(r, payload.Dir, payload.Ref)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	if err := s.saveCriteria(r, provider, todo, payload.Criteria); err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeRefreshedTodo(w, r, provider, source.Dir, todo)
}

// Running a todo's definition of done is not a handler of its own: the
// dashboard posts `step: verify` to /api/todos/run, and the lifecycle's verify
// step — the ONLY verifier — runs it and records the report on the attempt.

// handleTodoVerificationSchema exposes the CLI-owned `gavel fixtures --schema`
// document to the dashboard's FixtureEditor.
func (s *Server) handleTodoVerificationSchema(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if s.fixtureSchemaProvider == nil {
		writeTodoError(w, http.StatusServiceUnavailable, fmt.Errorf("fixture schema provider is not configured"))
		return
	}
	doc, err := s.fixtureSchemaProvider()
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(doc) //nolint:errcheck
}

// saveCriteria rewrites the todo body's acceptance-criteria section in place.
func (s *Server) saveCriteria(r *http.Request, provider todos.Provider, todo *types.TODO, criteria []types.AcceptanceCriterion) error {
	body := todos.UpsertCriteriaSection(todo.MarkdownBody, criteria)
	return provider.Edit(r.Context(), todo, todos.EditRequest{Body: &body})
}

// todoVerificationFixturePayload is the request body for saving the
// Verification tab's fixture markdown (edited via the dashboard's FixtureEditor).
type todoVerificationFixturePayload struct {
	Dir     string `json:"dir,omitempty"`
	Ref     string `json:"ref,omitempty"`
	Fixture string `json:"fixture"`
}

// handleTodoVerificationFixture saves a todo's dedicated verification markdown
// and returns the refreshed todo, mirroring handleTodoCriteria.
func (s *Server) handleTodoVerificationFixture(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoVerificationFixturePayload
	if err := decodeTodoRequest(r, &payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	provider, source, todo, status, err := s.loadTodoForWrite(r, payload.Dir, payload.Ref)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	if err := provider.Edit(r.Context(), todo, todos.EditRequest{Verification: &payload.Fixture}); err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeRefreshedTodo(w, r, provider, source.Dir, todo)
}

// writeRefreshedTodo re-reads the todo so the response reflects the provider's
// authoritative state (rewritten body, re-parsed criteria) and encodes it. A
// re-read that fails is reported, never papered over with the pre-edit todo:
// the edit is committed, and a 200 carrying the old body would show the user
// their save silently reverted.
func (s *Server) writeRefreshedTodo(w http.ResponseWriter, r *http.Request, provider todos.Provider, dir string, todo *types.TODO) {
	refreshed, err := provider.Get(r.Context(), todos.TODOReference(todo))
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError,
			fmt.Errorf("the edit was saved, but re-reading todo %s failed: %w", todos.TODOReference(todo), err))
		return
	}
	sum, err := todoDetail(r.Context(), provider, dir, refreshed)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(sum) //nolint:errcheck
}
