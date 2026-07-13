package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

// criteriaCatalogItem is one standard acceptance-criterion check the dashboard
// offers in the "add criteria" combobox.
type criteriaCatalogItem struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// handleCriteriaCatalog returns the static verify.AllChecks catalog (id +
// category + description) so the dashboard can offer the standard checks as
// selectable options when adding acceptance criteria. The list is grouped by
// category in catalog order.
func (s *Server) handleCriteriaCatalog(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	items := make([]criteriaCatalogItem, 0, len(verify.AllChecks))
	for _, c := range verify.AllChecks {
		items = append(items, criteriaCatalogItem{ID: c.ID, Category: c.Category, Description: c.Description})
	}
	json.NewEncoder(w).Encode(items) //nolint:errcheck
}

// todoCriteriaPayload is the shared request body for the criteria/verify
// endpoints: a todo reference plus, for the save endpoint, the full criteria
// list, and an optional model override for the AI endpoints.
type todoCriteriaPayload struct {
	Dir      string                      `json:"dir,omitempty"`
	Ref      string                      `json:"ref,omitempty"`
	Model    string                      `json:"model,omitempty"`
	Backend  string                      `json:"backend,omitempty"`
	Criteria []types.AcceptanceCriterion `json:"criteria,omitempty"`
	// Prompt, when non-empty, overrides the rendered verify prompt verbatim — the
	// dashboard's editable prompt.
	Prompt string `json:"prompt,omitempty"`
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
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid json"))
		return
	}
	provider, _, todo, status, err := s.loadTodoForWrite(r, payload.Dir, payload.Ref)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	if err := s.saveCriteria(r, provider, todo, payload.Criteria); err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeRefreshedTodo(w, r, provider, todo)
}

// handleTodoCriteriaGenerate drafts acceptance criteria for a todo with an AI
// model (the same catalog-seeded generator the CLI uses) and saves them.
func (s *Server) handleTodoCriteriaGenerate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoCriteriaPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid json"))
		return
	}
	provider, _, todo, status, err := s.loadTodoForWrite(r, payload.Dir, payload.Ref)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	agent, err := commit.BuildAgent(commit.Options{}, payload.Model)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	criteria, err := todos.Generate(ctx, agent, todo.Title, todo.MarkdownBody)
	if err != nil {
		writeTodoError(w, http.StatusBadGateway, err)
		return
	}
	if err := s.saveCriteria(r, provider, todo, criteria); err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeRefreshedTodo(w, r, provider, todo)
}

// handleTodoVerify scores a todo's commits against its acceptance criteria and
// returns the structured verdict plus the refreshed todo (its status may have
// moved to verified/pending).
func (s *Server) handleTodoVerify(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoCriteriaPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid json"))
		return
	}
	provider, source, todo, status, err := s.loadTodoForWrite(r, payload.Dir, payload.Ref)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	result, err := todos.RunIssueVerification(r.Context(), provider, todo, todos.VerifyOptions{
		WorkDir: todoVerifyWorkDir(source.Dir, todo),
		Model:   payload.Model,
		Backend: payload.Backend,
		Prompt:  payload.Prompt,
	})
	if err != nil {
		writeTodoError(w, http.StatusBadGateway, err)
		return
	}
	refreshed := todo
	if rt, gerr := provider.Get(r.Context(), todos.TODOReference(todo)); gerr == nil {
		refreshed = rt
	}
	json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
		"result": result,
		"todo":   summarizeTodo(refreshed, true),
	})
}

// handleTodoVerifyPreview renders the verify prompt a verify run would send, so
// the dashboard's editable prompt can be seeded for Verify mode (mirrors
// handleTodoRunPreview for Run/Plan).
func (s *Server) handleTodoVerifyPreview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoCriteriaPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid json"))
		return
	}
	_, source, todo, status, err := s.loadTodoForWrite(r, payload.Dir, payload.Ref)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	prompt, err := todos.PreviewIssueVerification(todo, todos.VerifyOptions{
		WorkDir: todoVerifyWorkDir(source.Dir, todo),
		Model:   payload.Model,
		Backend: payload.Backend,
	})
	if err != nil {
		writeTodoError(w, http.StatusBadGateway, err)
		return
	}
	json.NewEncoder(w).Encode(todoRunPreviewResponse{Prompt: prompt, Mode: "verify", Count: 1}) //nolint:errcheck
}

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

// handleTodoVerificationFixture saves a todo's "## Verification" fixture
// markdown, rewriting the section in place, and returns the refreshed todo —
// mirroring handleTodoCriteria for the Verification tab's FixtureEditor.
func (s *Server) handleTodoVerificationFixture(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoVerificationFixturePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid json"))
		return
	}
	provider, _, todo, status, err := s.loadTodoForWrite(r, payload.Dir, payload.Ref)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	body := todos.UpsertVerificationFixture(todo.MarkdownBody, payload.Fixture)
	if err := provider.Edit(r.Context(), todo, todos.EditRequest{Body: &body}); err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeRefreshedTodo(w, r, provider, todo)
}

// writeRefreshedTodo re-reads the todo so the response reflects the provider's
// authoritative state (rewritten body, re-parsed criteria) and encodes it.
func (s *Server) writeRefreshedTodo(w http.ResponseWriter, r *http.Request, provider todos.Provider, todo *types.TODO) {
	if refreshed, err := provider.Get(r.Context(), todos.TODOReference(todo)); err == nil {
		todo = refreshed
	}
	json.NewEncoder(w).Encode(summarizeTodo(todo, true)) //nolint:errcheck
}

// todoVerifyWorkDir resolves the directory a todo's commits live in (the todo's
// cwd against the workspace dir); git resolves the repository root from there.
func todoVerifyWorkDir(baseDir string, todo *types.TODO) string {
	if todo == nil || todo.CWD == "" {
		return baseDir
	}
	if filepath.IsAbs(todo.CWD) {
		return todo.CWD
	}
	return filepath.Join(baseDir, todo.CWD)
}
