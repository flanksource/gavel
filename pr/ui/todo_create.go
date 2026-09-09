package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/gavel/prwatch"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

type todoPRVerificationPayload struct {
	PRNumber   int      `json:"prNumber,omitempty"`
	Repo       string   `json:"repo,omitempty"`
	CommentIDs []int64  `json:"commentIds,omitempty"`
	Actions    []string `json:"actions,omitempty"`
}

type todoCreatePayload struct {
	Dir      string         `json:"dir,omitempty"`
	Title    string         `json:"title"`
	Body     string         `json:"body,omitempty"`
	Priority types.Priority `json:"priority,omitempty"`
	Status   types.Status   `json:"status,omitempty"`
	// Criteria, when set, are folded into the body as a "## Acceptance Criteria"
	// checklist on create — used by the dashboard's "create todo from PR" flow to
	// seed a todo with the PR's failing tests and lint violations as criteria.
	Criteria []types.AcceptanceCriterion `json:"criteria,omitempty"`
	// PRVerification, when set, is converted server-side into a generated exec
	// fixture in "## Verification" so GitHub comments/actions are checked by the
	// fixture runner rather than duplicated as AI-scored acceptance criteria.
	PRVerification *todoPRVerificationPayload `json:"prVerification,omitempty"`
	// Labels tags the new todo. Presentation for each label is resolved from the
	// label definitions, not stored here.
	Labels []string `json:"labels,omitempty"`
}

type todoNewPayload struct {
	todoCreatePayload
	AutoSave *bool `json:"autoSave,omitempty"`
}

type todoAttachmentSummary struct {
	Field       string `json:"field,omitempty"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType,omitempty"`
	Size        int64  `json:"size"`
	ID          string `json:"id,omitempty"`
	URL         string `json:"url,omitempty"`
	IsImage     bool   `json:"isImage,omitempty"`
}

type todoNewResponse struct {
	Todo        todoSummary             `json:"todo"`
	AutoSave    bool                    `json:"autoSave"`
	Attachments []todoAttachmentSummary `json:"attachments,omitempty"`
}

// todoTransferPayload moves the todo at Ref from one native workspace to another.
type todoTransferPayload struct {
	Ref     string `json:"ref"`
	FromDir string `json:"fromDir,omitempty"`
	ToDir   string `json:"toDir"`
}

type todoTransferResponse struct {
	Dir  string      `json:"dir"`
	Todo todoSummary `json:"todo"`
}

func (s *Server) handleTodoCreate(w http.ResponseWriter, r *http.Request) {
	var payload todoCreatePayload
	if err := decodeTodoRequest(r, &payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	if payload.Status != "" {
		if err := types.ValidateAssignableStatus(payload.Status); err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
	}
	if payload.Priority != "" {
		if err := types.ValidatePriority(payload.Priority); err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
	}
	source := todoSourceFromRequest(r)
	if payload.Dir != "" {
		source.Dir = payload.Dir
	}
	provider, source, err := s.todoProviderContext(r.Context(), source)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	todo, err := provider.Create(r.Context(), todos.CreateRequest{
		Title:    payload.Title,
		Body:     bodyWithCreateSections(payload.Body, payload.Criteria, payload.PRVerification),
		Priority: payload.Priority,
		Status:   payload.Status,
		Labels:   payload.Labels,
	})
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	sum, err := todoDetail(r.Context(), provider, source.Dir, todo)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sum) //nolint:errcheck
}

// bodyWithCriteria folds an "## Acceptance Criteria" checklist into the todo body
// when criteria are supplied (e.g. a todo created from a PR's failing tests and
// lint violations), so they round-trip through the provider's parse as the todo's
// acceptance criteria. An empty list leaves the body untouched.
func bodyWithCriteria(body string, criteria []types.AcceptanceCriterion) string {
	if len(criteria) == 0 {
		return body
	}
	return todos.UpsertCriteriaSection(body, criteria)
}

func bodyWithCreateSections(body string, criteria []types.AcceptanceCriterion, verification *todoPRVerificationPayload) string {
	body = bodyWithCriteria(body, criteria)
	if verification == nil {
		return body
	}
	return prwatch.UpsertPRStatusVerification(body, prwatch.PRStatusVerification{
		PRNumber:   verification.PRNumber,
		Repo:       verification.Repo,
		CommentIDs: verification.CommentIDs,
		Actions:    verification.Actions,
	})
}

func (s *Server) handleTodoNew(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	payload, attachments, err := parseTodoNewPayload(r)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	autoSave := false
	if payload.AutoSave != nil {
		autoSave = *payload.AutoSave
	}
	if payload.Status == "" {
		if autoSave {
			payload.Status = types.StatusPending
		} else {
			payload.Status = types.StatusDraft
		}
	}
	if payload.Status != "" {
		if err := types.ValidateAssignableStatus(payload.Status); err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
	}
	if payload.Priority != "" {
		if err := types.ValidatePriority(payload.Priority); err != nil {
			writeTodoError(w, http.StatusBadRequest, err)
			return
		}
	}

	source := todoSourceFromRequest(r)
	if payload.Dir != "" {
		source.Dir = payload.Dir
	}
	provider, source, err := s.todoProviderContext(r.Context(), source)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	todo, err := provider.Create(r.Context(), todos.CreateRequest{
		Title:    payload.Title,
		Body:     bodyWithCreateSections(todoBodyWithAttachments(payload.Body, attachments), payload.Criteria, payload.PRVerification),
		Priority: payload.Priority,
		Status:   payload.Status,
	})
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	sum, err := todoDetail(r.Context(), provider, source.Dir, todo)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todoNewResponse{ //nolint:errcheck
		Todo:        sum,
		AutoSave:    autoSave,
		Attachments: attachments,
	})
}

func (s *Server) handleTodoTransfer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload todoTransferPayload
	if err := decodeTodoRequest(r, &payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(payload.Ref) == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	if strings.TrimSpace(payload.ToDir) == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("toDir is required"))
		return
	}
	requestedSource := todoSource{Dir: payload.FromDir}
	source, src, resolvedTodo, err := s.resolveTodoReference(r.Context(), requestedSource, payload.Ref)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	// Transfer through the canonical native id after a legacy alias lookup.
	payload.Ref = resolvedTodo.ID
	target, dst, err := s.todoProviderContext(r.Context(), todoSource{Dir: payload.ToDir})
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	// Refuse a no-op self-transfer, which would create a duplicate and then
	// delete the original in the same native workspace.
	if src.Dir == dst.Dir {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("source and target are the same workspace"))
		return
	}
	created, err := todos.Transfer(r.Context(), source, target, payload.Ref)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	sum, err := todoDetail(r.Context(), target, dst.Dir, created)
	if err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(todoTransferResponse{ //nolint:errcheck
		Dir:  dst.Dir,
		Todo: sum,
	})
}
