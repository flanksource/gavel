package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/labels"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/ops"
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
	// Parent makes the TODO a child of the named top-level TODO. A nil pointer
	// leaves the hierarchy unchanged; an explicit empty string detaches.
	Parent *string `json:"parent,omitempty"`
}

// todoUpdateChanges validates a PATCH and splits it by the provider write that
// applies each part. A PATCH may move the TODO in the hierarchy (parent), edit
// content (title/body/labels), change state (status/priority), add a comment,
// or any combination; at least one operation is required.
func todoUpdateChanges(payload todoUpdatePayload, attachments []todoAttachmentSummary) (ops.EditChanges, string, error) {
	changes := ops.EditChanges{Parent: payload.Parent}
	if payload.Status != "" {
		if err := types.ValidateAssignableStatus(payload.Status); err != nil {
			return changes, "", err
		}
		changes.State.Status = &payload.Status
	}
	if payload.Priority != "" {
		if err := types.ValidatePriority(payload.Priority); err != nil {
			return changes, "", err
		}
		changes.State.Priority = &payload.Priority
	}
	if payload.Title != nil {
		title := strings.TrimSpace(*payload.Title)
		if title == "" {
			return changes, "", fmt.Errorf("title cannot be empty")
		}
		changes.Content.Title = &title
	}
	changes.Content.Body = payload.Body
	if payload.Labels != nil {
		normalized := make([]string, 0, len(*payload.Labels))
		for _, label := range *payload.Labels {
			if label = labels.Normalize(label); label != "" {
				normalized = append(normalized, label)
			}
		}
		changes.Content.Labels = &normalized
	}
	comment := strings.TrimSpace(payload.Comment)
	if len(attachments) > 0 {
		comment = todoBodyWithAttachments(comment, attachments)
	}
	if changes.Parent == nil && changes.State.Status == nil && changes.State.Priority == nil && changes.Content.IsEmpty() && comment == "" {
		return changes, "", fmt.Errorf("status, priority, title, body, labels, parent, or comment is required")
	}
	return changes, comment, nil
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
	changes, comment, err := todoUpdateChanges(payload, attachments)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
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
	if _, ok := provider.(todos.ParentProvider); changes.Parent != nil && !ok {
		writeTodoError(w, http.StatusNotImplemented, fmt.Errorf("setting a parent requires native TODO storage"))
		return
	}
	if todo, err = applyTodoPatch(r.Context(), provider, todo, changes, comment); err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	sum, err := todoDetail(r.Context(), provider, source.Dir, todo)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(sum) //nolint:errcheck
}

// applyTodoPatch performs a PATCH's writes and returns the TODO as storage now
// holds it. Order: parent, content, reopen/close, comment. The parent is the one
// write the hierarchy can refuse, so it goes before anything is changed; the
// comment goes last so a reopen-with-comment posts against the now-open TODO
// and lands last in the timeline.
func applyTodoPatch(ctx context.Context, provider todos.Provider, todo *types.TODO, changes ops.EditChanges, comment string) (*types.TODO, error) {
	todo, err := ops.ApplyEdit(ctx, provider, todo, changes)
	if err != nil {
		return nil, err
	}
	if comment != "" {
		if err := provider.Comment(ctx, todo, comment); err != nil {
			return nil, err
		}
	}
	if changes.Parent == nil && changes.Content.IsEmpty() && comment == "" {
		return todo, nil
	}
	// A parent change, an edit and a comment each mutate the hierarchy, body or
	// event history, so the response is re-read; a failed re-read is reported
	// rather than answered with the pre-write TODO.
	refreshed, err := provider.Get(ctx, todo.ID)
	if err != nil {
		return nil, fmt.Errorf("re-read TODO %s after update: %w", todo.ID, err)
	}
	return refreshed, nil
}

// handleTodoDelete closes a TODO. `children=archive|detach` says what happens to
// its open children; a TODO that has some and no choice is refused with 409.
func (s *Server) handleTodoDelete(w http.ResponseWriter, r *http.Request) {
	ref := strings.TrimSpace(r.URL.Query().Get("ref"))
	if ref == "" {
		writeTodoDeleteError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	children, err := todos.ParseChildrenDisposition(r.URL.Query().Get("children"))
	if err != nil {
		writeTodoDeleteError(w, http.StatusBadRequest, err)
		return
	}
	provider, _, todo, err := s.resolveTodoReference(r.Context(), todoSourceFromRequest(r), ref)
	if err != nil {
		writeTodoDeleteError(w, http.StatusNotFound, err)
		return
	}
	if _, err := todos.Archive(r.Context(), provider, todo, todos.ArchiveOptions{Children: children}); err != nil {
		writeTodoDeleteError(w, http.StatusInternalServerError, err)
		return
	}
	fmt.Fprint(w, `{"status":"ok"}`)
}

// writeTodoDeleteError keeps 409 for the open-children refusal alone. The
// dashboard reads any 409 on a delete as "ask what to do with the open
// children" and shows the message as the question, so every other conflict — a
// stale version, an ambiguous reference — answers 412 here.
func writeTodoDeleteError(w http.ResponseWriter, fallback int, err error) {
	status := todoErrorStatus(fallback, err)
	if status == http.StatusConflict && !errors.Is(err, native.ErrOpenChildren) {
		status = http.StatusPreconditionFailed
	}
	writeTodoErrorStatus(w, status, err)
}
