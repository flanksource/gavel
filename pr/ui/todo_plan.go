package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/captain/pkg/claude"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/internal/database"
)

// todoPlanResponse is the exit-plan-mode plan for a TODO's agent session: the plan
// markdown plus where it lives. Found is false when the session produced no plan
// (yet) — the normal "nothing to show" state, mirroring the session-stats endpoint.
type todoPlanResponse struct {
	Found   bool   `json:"found"`
	Path    string `json:"path,omitempty"`
	Content string `json:"content,omitempty"`
	OnDisk  bool   `json:"onDisk,omitempty"`
	Slug    string `json:"slug,omitempty"`
}

// handleTodoSessionPlan returns the plan a plan-mode run produced, recovered from
// the Claude session by id via captain's canonical plan resolver (which prefers the
// on-disk ~/.claude/plans/<slug>.md over the inline transcript copy and also handles
// codex). A session with no plan is reported as found=false, not an error, so the
// dashboard simply shows an empty Plan tab.
func (s *Server) handleTodoSessionPlan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sessionID := strings.TrimSpace(r.URL.Query().Get("sessionId"))
	if sessionID == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("sessionId is required"))
		return
	}
	if _, err := database.Shared(r.Context()); err != nil {
		writeTodoError(w, http.StatusServiceUnavailable, fmt.Errorf("prepare Captain database: %w", err))
		return
	}
	res, err := captaincli.RunPlan(captaincli.PlanOptions{SessionID: sessionID})
	if err != nil {
		// No session, or a session that never planned: nothing to show, not a failure.
		json.NewEncoder(w).Encode(todoPlanResponse{Found: false}) //nolint:errcheck
		return
	}
	json.NewEncoder(w).Encode(todoPlanResponse{ //nolint:errcheck
		Found:   true,
		Path:    res.Path,
		Content: res.Content,
		OnDisk:  res.OnDisk,
		Slug:    res.Slug,
	})
}

// handleTodoSessionPlanSave rewrites the plan file so a human can refine a plan
// before approving it. It only ever writes within Claude's plans directory (a path
// outside it is rejected), resolving the path from the session id when the client
// does not supply one.
func (s *Server) handleTodoSessionPlanSave(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload struct {
		SessionID string `json:"sessionId"`
		Path      string `json:"path"`
		Content   string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}

	path := strings.TrimSpace(payload.Path)
	if path == "" {
		sessionID := strings.TrimSpace(payload.SessionID)
		if sessionID == "" {
			writeTodoError(w, http.StatusBadRequest, fmt.Errorf("path or sessionId is required"))
			return
		}
		if _, err := database.Shared(r.Context()); err != nil {
			writeTodoError(w, http.StatusServiceUnavailable, fmt.Errorf("prepare Captain database: %w", err))
			return
		}
		res, err := captaincli.RunPlan(captaincli.PlanOptions{SessionID: sessionID})
		if err != nil || strings.TrimSpace(res.Path) == "" {
			writeTodoError(w, http.StatusNotFound, fmt.Errorf("no plan file to save for session %q", sessionID))
			return
		}
		path = res.Path
	}

	if !plansDirContains(path) {
		writeTodoError(w, http.StatusForbidden, fmt.Errorf("plan path %q is outside the plans directory", path))
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	if err := os.WriteFile(path, []byte(payload.Content), 0o644); err != nil {
		writeTodoError(w, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(todoPlanResponse{ //nolint:errcheck
		Found:   true,
		Path:    path,
		Content: payload.Content,
		OnDisk:  true,
	})
}

// plansDirContains reports whether path resolves inside Claude's plans directory, so
// the save endpoint can only rewrite plan files and never an arbitrary location
// (including via a "../" escape).
func plansDirContains(path string) bool {
	dir, err := filepath.Abs(claude.GetPlansDir())
	if err != nil {
		return false
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(dir, abs)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
