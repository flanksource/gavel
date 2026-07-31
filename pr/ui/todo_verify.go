package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	todospec "github.com/flanksource/gavel/todos/spec"
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

// handleTodoVerificationRun executes the persisted issue Verification fixture
// through the same verify-only Captain lifecycle used by `gavel todos check`.
// The response contains fixture/checklist evidence rather than a parallel
// score-oriented AI verdict.
func (s *Server) handleTodoVerificationRun(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoVerificationRunPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid json"))
		return
	}
	// The wire value is guarded before anything is loaded, so a malformed or
	// committing spec is a 400 rather than a failed run.
	if _, err := validateTodoVerificationSpec(payload.Spec); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	provider, source, todo, status, err := s.loadTodoForWrite(r, payload.Dir, payload.Ref)
	if err != nil {
		writeTodoError(w, status, err)
		return
	}
	workDir := todoVerifyWorkDir(source.Dir, todo)
	// The request is the top layer of the verification chain, not the whole of
	// it: .gavel.yaml todos.verify and ai: are what a dashboard check falls back
	// to, exactly as `gavel todos check` does.
	var override api.Spec
	if payload.Spec != nil {
		override = *payload.Spec
	}
	resolved, err := todospec.Resolve(todospec.Input{
		WorkDir:    workDir,
		Mode:       types.ModeVerify,
		Todos:      []*types.TODO{todo},
		Override:   override,
		CanApprove: true,
	})
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	// Re-validated after folding so the timeout comes from the layer that set it
	// and the 10m floor still applies when no layer names one.
	timeout, err := validateTodoVerificationSpec(&resolved.Spec)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()
	result := todos.CheckTODO(ctx, todo, todos.CheckOptions{
		WorkDir:  workDir,
		Timeout:  timeout,
		Logger:   logger.StandardLogger(),
		Provider: provider,
		Spec:     &resolved.Spec,
	})
	refreshed := todo
	if rt, gerr := provider.Get(r.Context(), todos.TODOReference(todo)); gerr == nil {
		refreshed = rt
	}
	json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
		"verification": result,
		"todo":         summarizeTodo(refreshed, true),
	})
}

type todoVerificationRunPayload struct {
	Dir  string    `json:"dir,omitempty"`
	Ref  string    `json:"ref,omitempty"`
	Spec *api.Spec `json:"spec,omitempty"`
}

func validateTodoVerificationSpec(spec *api.Spec) (time.Duration, error) {
	const defaultTimeout = 10 * time.Minute
	if spec == nil {
		return defaultTimeout, nil
	}
	if !reflect.ValueOf(spec.Model).IsZero() {
		if err := spec.Model.Validate(); err != nil {
			return 0, fmt.Errorf("model: %w", err)
		}
	}
	if err := spec.Budget.Validate(); err != nil {
		return 0, fmt.Errorf("budget: %w", err)
	}
	if err := spec.Permissions.Validate(); err != nil {
		return 0, fmt.Errorf("permissions: %w", err)
	}
	if err := spec.ToolPreferences.Validate(); err != nil {
		return 0, err
	}
	if err := spec.Workflow.Validate(); err != nil {
		return 0, fmt.Errorf("workflow: %w", err)
	}
	if spec.Workflow != nil && len(spec.Workflow.Commits) > 0 {
		return 0, fmt.Errorf("verification runs cannot commit")
	}
	if strings.TrimSpace(spec.Budget.Timeout) == "" {
		return defaultTimeout, nil
	}
	timeout, err := time.ParseDuration(spec.Budget.Timeout)
	if err != nil {
		return 0, fmt.Errorf("budget.timeout: %w", err)
	}
	if timeout <= 0 {
		return 0, fmt.Errorf("budget.timeout must be greater than zero")
	}
	return timeout, nil
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

// handleTodoVerificationFixture saves a todo's dedicated verification markdown
// and returns the refreshed todo, mirroring handleTodoCriteria.
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
	if err := provider.Edit(r.Context(), todo, todos.EditRequest{Verification: &payload.Fixture}); err != nil {
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
