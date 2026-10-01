package ui

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/gavel/github/prcreate"
	"github.com/flanksource/gavel/todos/land"
	"github.com/flanksource/gavel/todos/native"
)

// landTodoRun is the landing seam; specs replace it to drive the handler
// without git or a database.
var landTodoRun = land.Land

type todoLandPayload struct {
	Ref   string            `json:"ref"`
	Dir   string            `json:"dir,omitempty"`
	Via   native.LandingVia `json:"via"`
	Base  string            `json:"base,omitempty"`
	Draft bool              `json:"draft,omitempty"`
}

// todoLandResponse carries the recorded landing. Error is set only alongside a
// landing whose worktree cleanup failed after it was recorded.
type todoLandResponse struct {
	Landing *native.RunLanding `json:"landing"`
	Error   string             `json:"error,omitempty"`
}

// handleTodoLand lands a TODO's newest run step worktree by merge or PR.
func (s *Server) handleTodoLand(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var payload todoLandPayload
	if err := decodeTodoRequest(r, &payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(payload.Ref) == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	opts := land.Options{Via: payload.Via, Base: payload.Base, Draft: payload.Draft}
	if err := opts.Validate(); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	if opts.Via == native.LandingPR {
		opts.Deps = prcreate.DefaultDeps()
	}
	provider, _, todo, err := s.resolveTodoReference(r.Context(), todoSource{Dir: payload.Dir}, payload.Ref)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	backed, ok := provider.(land.Provider)
	if !ok {
		writeTodoError(w, http.StatusNotImplemented, land.ErrNotNative)
		return
	}
	landing, err := landTodoRun(r.Context(), backed, todo, opts)
	if err != nil && landing != nil {
		writeTodoJSON(w, http.StatusInternalServerError, todoLandResponse{Landing: landing, Error: err.Error()})
		return
	}
	if err != nil {
		writeTodoError(w, todoLandStatus(err), err)
		return
	}
	writeTodoJSON(w, http.StatusOK, todoLandResponse{Landing: landing})
}

func todoLandStatus(err error) int {
	var conflict *land.ConflictError
	var prConflict *prcreate.ConflictError
	switch {
	case errors.Is(err, land.ErrInvalidOptions):
		return http.StatusBadRequest
	case errors.Is(err, land.ErrNotNative):
		return http.StatusNotImplemented
	case errors.As(err, &conflict), errors.As(err, &prConflict),
		errors.Is(err, native.ErrAlreadyLanded), errors.Is(err, land.ErrNoRunWorkspace),
		errors.Is(err, land.ErrNoCommits), errors.Is(err, land.ErrWorktreeDirty),
		errors.Is(err, land.ErrCheckoutNotReady):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
