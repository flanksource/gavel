package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// todoLinkPayload is the write body for creating or removing a link. Relation
// accepts either spelling ("depends-on" or "depends_on") and defaults to
// related_to.
type todoLinkPayload struct {
	Dir      string `json:"dir"`
	Ref      string `json:"ref"`
	Target   string `json:"target"`
	Relation string `json:"relation"`
}

type todoLinksResponse struct {
	Links []todos.Link `json:"links"`
}

func (s *Server) handleTodoLinks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		s.handleTodoLinksList(w, r)
	case http.MethodPost:
		s.handleTodoLinkCreate(w, r)
	case http.MethodDelete:
		s.handleTodoLinkDelete(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleTodoLinksList(w http.ResponseWriter, r *http.Request) {
	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	if ref == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	linker, todo, err := s.todoLinkTarget(r.Context(), todoSourceFromRequest(r), ref)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	links, err := linker.Links(r.Context(), todo)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	json.NewEncoder(w).Encode(todoLinksResponse{Links: links}) //nolint:errcheck
}

func (s *Server) handleTodoLinkCreate(w http.ResponseWriter, r *http.Request) {
	payload, relation, linker, todo, ok := s.decodeTodoLinkRequest(w, r)
	if !ok {
		return
	}
	link, err := linker.Link(r.Context(), todo, payload.Target, relation)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(link) //nolint:errcheck
}

func (s *Server) handleTodoLinkDelete(w http.ResponseWriter, r *http.Request) {
	payload, relation, linker, todo, ok := s.decodeTodoLinkRequest(w, r)
	if !ok {
		return
	}
	if err := linker.Unlink(r.Context(), todo, payload.Target, relation); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	links, err := linker.Links(r.Context(), todo)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	json.NewEncoder(w).Encode(todoLinksResponse{Links: links}) //nolint:errcheck
}

// decodeTodoLinkRequest resolves the shared write inputs. It writes the error
// response itself; ok is false when the caller must stop.
func (s *Server) decodeTodoLinkRequest(w http.ResponseWriter, r *http.Request) (
	todoLinkPayload, types.RelationKind, todos.RelationshipProvider, *types.TODO, bool,
) {
	var payload todoLinkPayload
	if err := decodeTodoRequest(r, &payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return payload, "", nil, nil, false
	}
	payload.Ref = strings.TrimSpace(payload.Ref)
	payload.Target = strings.TrimSpace(payload.Target)
	if payload.Ref == "" || payload.Target == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref and target are required"))
		return payload, "", nil, nil, false
	}
	relation, err := types.ParseRelationKind(payload.Relation)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return payload, "", nil, nil, false
	}
	source := todoSourceFromRequest(r)
	if payload.Dir != "" {
		source.Dir = payload.Dir
	}
	linker, todo, err := s.todoLinkTarget(r.Context(), source, payload.Ref)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return payload, "", nil, nil, false
	}
	return payload, relation, linker, todo, true
}

func (s *Server) todoLinkTarget(ctx context.Context, source todoSource, ref string) (
	todos.RelationshipProvider, *types.TODO, error,
) {
	provider, _, err := s.todoProviderContext(ctx, source)
	if err != nil {
		return nil, nil, err
	}
	linker, ok := provider.(todos.RelationshipProvider)
	if !ok {
		return nil, nil, fmt.Errorf("TODO provider does not support links; native PostgreSQL storage is required")
	}
	todo, err := provider.Get(ctx, ref)
	if err != nil {
		return nil, nil, err
	}
	return linker, todo, nil
}
