package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// todoBulkTarget names one TODO to edit. Dir is its owning workspace; an empty
// Dir falls back to the same global reference lookup the single-item PATCH uses,
// so a selection spanning workspaces (severity/age grouping) needs no client-side
// workspace bookkeeping.
type todoBulkTarget struct {
	Dir string `json:"dir,omitempty"`
	Ref string `json:"ref"`
}

// todoBulkRequest applies one status, priority, and/or comment to many TODOs.
// The operations mirror the single-item PATCH minus title/body: bulk-editing
// content would mean writing the same prose over every selected TODO.
type todoBulkRequest struct {
	Items    []todoBulkTarget `json:"items"`
	Status   types.Status     `json:"status,omitempty"`
	Priority types.Priority   `json:"priority,omitempty"`
	Comment  string           `json:"comment,omitempty"`
}

type todoBulkItemResult struct {
	Dir   string       `json:"dir"`
	Ref   string       `json:"ref"`
	Todo  *todoSummary `json:"todo,omitempty"`
	Error string       `json:"error,omitempty"`
}

type todoBulkResponse struct {
	Updated int                  `json:"updated"`
	Failed  int                  `json:"failed"`
	Results []todoBulkItemResult `json:"results"`
}

func (s *Server) handleTodoBulk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	payload, update, err := parseTodoBulkRequest(r)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	json.NewEncoder(w).Encode(s.applyTodoBulk(r.Context(), payload, update)) //nolint:errcheck
}

// parseTodoBulkRequest validates the whole request before a single write lands:
// a malformed target or an unassignable status rejects the batch outright rather
// than half-applying it.
func parseTodoBulkRequest(r *http.Request) (todoBulkRequest, todos.StateUpdate, error) {
	var payload todoBulkRequest
	var update todos.StateUpdate
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return payload, update, fmt.Errorf("invalid request body: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return payload, update, fmt.Errorf("invalid request body: expected one JSON object")
	}
	if len(payload.Items) == 0 {
		return payload, update, fmt.Errorf("items is required and must name at least one TODO")
	}
	seen := make(map[todoBulkTarget]int, len(payload.Items))
	for i := range payload.Items {
		payload.Items[i].Dir = strings.TrimSpace(payload.Items[i].Dir)
		payload.Items[i].Ref = strings.TrimSpace(payload.Items[i].Ref)
		if payload.Items[i].Ref == "" {
			return payload, update, fmt.Errorf("items[%d].ref is required", i)
		}
		if first, ok := seen[payload.Items[i]]; ok {
			return payload, update, fmt.Errorf("items[%d] duplicates items[%d]: %q", i, first, payload.Items[i].Ref)
		}
		seen[payload.Items[i]] = i
	}
	if payload.Status != "" {
		if err := types.ValidateAssignableStatus(payload.Status); err != nil {
			return payload, update, err
		}
		update.Status = &payload.Status
	}
	if payload.Priority != "" {
		if err := types.ValidatePriority(payload.Priority); err != nil {
			return payload, update, err
		}
		update.Priority = &payload.Priority
	}
	payload.Comment = strings.TrimSpace(payload.Comment)
	if update.Status == nil && update.Priority == nil && payload.Comment == "" {
		return payload, update, fmt.Errorf("status, priority, or comment is required")
	}
	return payload, update, nil
}

// applyTodoBulk edits each target in request order and reports per-item outcomes
// instead of aborting on the first failure: a re-prioritization of forty TODOs
// must not be lost because one of them was archived in another tab. The loop is
// serial so the provider's optimistic version checks never race each other.
func (s *Server) applyTodoBulk(ctx context.Context, payload todoBulkRequest, update todos.StateUpdate) todoBulkResponse {
	response := todoBulkResponse{Results: make([]todoBulkItemResult, 0, len(payload.Items))}
	for _, item := range payload.Items {
		result := todoBulkItemResult{Dir: item.Dir, Ref: item.Ref}
		todo, err := s.applyTodoBulkItem(ctx, item, update, payload.Comment)
		if err != nil {
			result.Error = err.Error()
			response.Failed++
		} else {
			summary := summarizeTodo(todo, false)
			result.Todo = &summary
			response.Updated++
		}
		response.Results = append(response.Results, result)
	}
	return response
}

func (s *Server) applyTodoBulkItem(ctx context.Context, item todoBulkTarget, update todos.StateUpdate, comment string) (*types.TODO, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	provider, _, todo, err := s.resolveTodoReference(ctx, todoSource{Dir: item.Dir}, item.Ref)
	if err != nil {
		return nil, err
	}
	if update.Status != nil || update.Priority != nil {
		if err := provider.UpdateState(ctx, todo, update); err != nil {
			return nil, err
		}
	}
	if comment == "" {
		return todo, nil
	}
	if err := provider.Comment(ctx, todo, comment); err != nil {
		return nil, err
	}
	// The comment rewrites the body and adds an event; re-read so the response
	// carries the provider's authoritative state, as the single-item PATCH does.
	if refreshed, gerr := provider.Get(ctx, todo.ID); gerr == nil {
		todo = refreshed
	}
	return todo, nil
}
