package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/labels"
)

// todoLabelPayload is the write body for defining or updating one label's
// presentation. An empty Color takes the label's deterministic hashed hue, so
// "define this label" never silently repaints it.
type todoLabelPayload struct {
	Dir         string `json:"dir,omitempty"`
	Name        string `json:"name"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Description string `json:"description,omitempty"`
	// Global writes the definition once for every workspace instead of scoping
	// it to this one.
	Global bool `json:"global,omitempty"`
}

// todoLabelsResponse is the effective taxonomy plus per-label usage counts. The
// dashboard caches this separately from the todo list and joins them client
// side: definitions change rarely, the list changes constantly, and embedding
// the dictionary in every row would repeat it once per todo.
type todoLabelsResponse struct {
	Definitions labels.Definitions `json:"definitions"`
	Counts      map[string]int     `json:"counts,omitempty"`
	Palette     []string           `json:"palette,omitempty"`
	// Removed is set on a DELETE so the dashboard can report what the removal
	// actually did — a project removal also strips the label from TODOs, and
	// the count is the only way the client knows how many changed underneath it.
	Removed *labels.Removal `json:"removed,omitempty"`
}

func (s *Server) handleTodoLabels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		s.handleTodoLabelsList(w, r)
	case http.MethodPost:
		s.handleTodoLabelSet(w, r)
	case http.MethodDelete:
		s.handleTodoLabelDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTodoLabelsList(w http.ResponseWriter, r *http.Request) {
	store, err := s.todoLabelStore(r.Context(), todoSourceFromRequest(r))
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}

	definitions, err := store.LabelDefinitions(r.Context())
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	counts, err := store.LabelCounts(r.Context())
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}

	json.NewEncoder(w).Encode(todoLabelsResponse{ //nolint:errcheck
		Definitions: definitions,
		Counts:      counts,
		Palette:     labels.PaletteStrings(),
	})
}

func (s *Server) handleTodoLabelSet(w http.ResponseWriter, r *http.Request) {
	var payload todoLabelPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid label payload: %w", err))
		return
	}

	source := todoSourceFromRequest(r)
	if payload.Dir != "" {
		source.Dir = payload.Dir
	}
	store, err := s.todoLabelStore(r.Context(), source)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}

	definition := labels.Definition{
		Name:        labels.Normalize(payload.Name),
		Color:       labels.Color(labels.Normalize(payload.Color)),
		Icon:        labels.Normalize(payload.Icon),
		Description: strings.TrimSpace(payload.Description),
	}
	if definition.Name == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("name is required"))
		return
	}
	if definition.Color == "" {
		definition.Color = labels.Hash(definition.Name)
	}
	if err := definition.Validate(); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}

	saved, err := store.SetLabelDefinition(r.Context(), definition, payload.Global)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(saved) //nolint:errcheck
}

func (s *Server) handleTodoLabelDelete(w http.ResponseWriter, r *http.Request) {
	name := labels.Normalize(r.URL.Query().Get("name"))
	if name == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("name is required"))
		return
	}

	store, err := s.todoLabelStore(r.Context(), todoSourceFromRequest(r))
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}

	global := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("scope")), string(labels.ScopeGlobal))
	removal, err := store.DeleteLabelDefinition(r.Context(), name, global)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}

	definitions, err := store.LabelDefinitions(r.Context())
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	// The counts go back with the definitions because a project removal has just
	// changed them: the client's cached facet counts still show the label it
	// removed, and every TODO it was stripped from.
	counts, err := store.LabelCounts(r.Context())
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(todoLabelsResponse{ //nolint:errcheck
		Definitions: definitions,
		Counts:      counts,
		Removed:     &removal,
	})
}

// todoLabelStore resolves the provider's label-definition capability. Label
// presentation is stored in PostgreSQL, so a non-native provider reports that
// plainly rather than returning an empty taxonomy that would look like "no
// labels are defined".
func (s *Server) todoLabelStore(ctx context.Context, source todoSource) (todos.LabelDefinitionProvider, error) {
	provider, _, err := s.todoProviderContext(ctx, source)
	if err != nil {
		return nil, err
	}
	store, ok := provider.(todos.LabelDefinitionProvider)
	if !ok {
		return nil, fmt.Errorf("TODO provider does not support label definitions; native PostgreSQL storage is required")
	}
	return store, nil
}
