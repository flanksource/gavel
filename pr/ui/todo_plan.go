package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// todoPlanResponse is the exit-plan-mode plan for a TODO's agent session: the plan
// markdown plus where it lives. Found is false when the session produced no plan
// (yet) — the normal "nothing to show" state, mirroring the session-stats endpoint.
type todoPlanResponse struct {
	Found   bool         `json:"found"`
	Path    string       `json:"path,omitempty"`
	Content string       `json:"content,omitempty"`
	OnDisk  bool         `json:"onDisk,omitempty"`
	Slug    string       `json:"slug,omitempty"`
	Ref     string       `json:"ref,omitempty"`
	Version int64        `json:"version,omitempty"`
	Todo    *todoSummary `json:"todo,omitempty"`
}

// handleTodoSessionPlan returns the latest immutable Captain revision selected
// on a native issue. The historical route name stays stable for the UI, but no
// session log or agent-owned plan file participates in the read.
func (s *Server) handleTodoSessionPlan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	if ref == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	provider, _, todo, err := s.resolveTodoReference(r.Context(), todoSourceFromRequest(r), ref)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	plans, ok := provider.(todos.PlanContentProvider)
	if !ok {
		writeTodoError(w, http.StatusNotImplemented, fmt.Errorf("PostgreSQL TODO provider cannot read Captain plan revisions"))
		return
	}
	content, err := plans.PlanMarkdown(r.Context(), todo, types.ModePlan)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(todoPlanResponse{ //nolint:errcheck
		Found: strings.TrimSpace(content) != "", Content: content,
		Ref: todo.ID, Version: todo.Version,
	})
}

// handleTodoSessionPlanSave appends a human-edited immutable Captain revision
// and keeps it selected on the issue. It never rewrites an agent plan file.
func (s *Server) handleTodoSessionPlanSave(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload struct {
		Dir     string `json:"dir,omitempty"`
		Ref     string `json:"ref"`
		Version int64  `json:"version,omitempty"`
		Content string `json:"content"`
		Actor   string `json:"actor,omitempty"`
	}
	if err := decodeTodoRequest(r, &payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(payload.Ref) == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	if strings.TrimSpace(payload.Content) == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("plan content is required"))
		return
	}
	source := todoSourceFromRequest(r)
	if payload.Dir != "" {
		source.Dir = payload.Dir
	}
	provider, source, todo, err := s.resolveTodoReference(r.Context(), source, payload.Ref)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	if payload.Version > 0 && payload.Version != todo.Version {
		writeTodoError(w, http.StatusConflict, fmt.Errorf("native TODO version conflict: issue %s expected version %d, current version %d", todo.ID, payload.Version, todo.Version))
		return
	}
	revisions, ok := provider.(todos.PlanRevisionProvider)
	if !ok {
		writeTodoError(w, http.StatusNotImplemented, fmt.Errorf("PostgreSQL TODO provider cannot append Captain plan revisions"))
		return
	}
	updated, err := revisions.SavePlanRevision(r.Context(), todo, payload.Content, payload.Actor)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	sum, err := todoDetail(r.Context(), provider, source.Dir, updated)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(todoPlanResponse{ //nolint:errcheck
		Found: true, Content: payload.Content,
		Ref: updated.ID, Version: updated.Version,
		Todo: todoSummaryPointer(sum),
	})
}

func todoSummaryPointer(summary todoSummary) *todoSummary {
	return &summary
}
