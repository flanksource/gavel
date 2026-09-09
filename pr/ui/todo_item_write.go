package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/labels"
	"github.com/flanksource/gavel/todos/types"
)

type todoUpdatePayload struct {
	Dir      string         `json:"dir,omitempty"`
	Ref      string         `json:"ref,omitempty"`
	Status   types.Status   `json:"status,omitempty"`
	Priority types.Priority `json:"priority,omitempty"`
	// Title/Body edit the TODO's content; a nil pointer leaves the field
	// unchanged (an explicit empty body is allowed, an empty title is not).
	Title *string `json:"title,omitempty"`
	Body  *string `json:"body,omitempty"`
	// Comment, when set, appends a comment. Combined with status it reopens (or
	// closes) the TODO with a comment in one request.
	Comment string `json:"comment,omitempty"`
	// Labels replaces the TODO's whole label set. A nil pointer leaves them
	// unchanged; a non-nil empty array clears every label.
	Labels *[]string `json:"labels,omitempty"`
}

func (s *Server) handleTodoPatch(w http.ResponseWriter, r *http.Request) {
	payload, attachments, err := parseTodoUpdatePayload(r)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	ref := strings.TrimSpace(payload.Ref)
	if ref == "" {
		ref = strings.TrimSpace(r.URL.Query().Get("ref"))
	}
	if ref == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	// A PATCH may edit content (title/body), change state (status/priority),
	// add a comment, or any combination; at least one operation is required.
	var update todos.StateUpdate
	if payload.Status != "" {
		if err := types.ValidateAssignableStatus(payload.Status); err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
		update.Status = &payload.Status
	}
	if payload.Priority != "" {
		if err := types.ValidatePriority(payload.Priority); err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
		update.Priority = &payload.Priority
	}

	var edit todos.EditRequest
	if payload.Title != nil {
		title := strings.TrimSpace(*payload.Title)
		if title == "" {
			writeTodoError(w, http.StatusBadRequest, fmt.Errorf("title cannot be empty"))
			return
		}
		edit.Title = &title
	}
	if payload.Body != nil {
		edit.Body = payload.Body
	}
	if payload.Labels != nil {
		normalized := make([]string, 0, len(*payload.Labels))
		for _, label := range *payload.Labels {
			if label = labels.Normalize(label); label != "" {
				normalized = append(normalized, label)
			}
		}
		edit.Labels = &normalized
	}
	comment := strings.TrimSpace(payload.Comment)
	if len(attachments) > 0 {
		comment = todoBodyWithAttachments(comment, attachments)
	}

	hasState := update.Status != nil || update.Priority != nil
	if !hasState && edit.IsEmpty() && comment == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("status, priority, title, body, labels, or comment is required"))
		return
	}

	source := todoSourceFromRequest(r)
	if payload.Dir != "" {
		source.Dir = payload.Dir
	}
	provider, source, todo, err := s.resolveTodoReference(r.Context(), source, ref)
	if err != nil {
		writeTodoError(w, http.StatusNotFound, err)
		return
	}
	// Order: edit content, then reopen/close, then comment, so a reopen-with-comment
	// posts the comment against the now-open TODO and it lands last in the timeline.
	if !edit.IsEmpty() {
		if err := provider.Edit(r.Context(), todo, edit); err != nil {
			writeTodoError(w, http.StatusInternalServerError, err)
			return
		}
	}
	if hasState {
		if err := provider.UpdateState(r.Context(), todo, update); err != nil {
			writeTodoError(w, http.StatusInternalServerError, err)
			return
		}
	}
	if comment != "" {
		if err := provider.Comment(r.Context(), todo, comment); err != nil {
			writeTodoError(w, http.StatusInternalServerError, err)
			return
		}
	}
	// Edits and comments mutate the body/event history; re-read so the response
	// reflects the provider's authoritative state (new event, rewritten body).
	if !edit.IsEmpty() || comment != "" {
		if refreshed, gerr := provider.Get(r.Context(), todo.ID); gerr == nil {
			todo = refreshed
		}
	}
	sum, err := todoDetail(r.Context(), provider, source.Dir, todo)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(sum) //nolint:errcheck
}

func (s *Server) handleTodoDelete(w http.ResponseWriter, r *http.Request) {
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
	if err := provider.Delete(r.Context(), todo); err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	fmt.Fprint(w, `{"status":"ok"}`)
}
