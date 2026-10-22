package ui

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/gavel/git/branchmerge"
	prcreate "github.com/flanksource/gavel/pr/create"
	"github.com/flanksource/gavel/sandbox"
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
	// Title and LeaveOut shape a hosted landing's PR.
	Title    string   `json:"title,omitempty"`
	LeaveOut []string `json:"leaveOut,omitempty"`
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
	opts := land.Options{Via: payload.Via, Base: payload.Base, Draft: payload.Draft, Title: payload.Title, LeaveOut: payload.LeaveOut}
	if err := opts.Validate(); err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	if opts.Via == native.LandingPR {
		opts.Deps = prcreate.DefaultDeps()
	}
	provider, source, todo, err := s.resolveTodoReference(r.Context(), todoSource{Dir: payload.Dir}, payload.Ref)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}
	backed, ok := provider.(land.Provider)
	if !ok {
		writeTodoError(w, http.StatusNotImplemented, land.ErrNotNative)
		return
	}
	var landing *native.RunLanding
	if source.Hosted != nil {
		landing, err = s.landHostedRun(r.Context(), backed, source, todo, opts)
	} else if strings.TrimSpace(opts.Title) != "" || len(opts.LeaveOut) > 0 {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("%w: a PR title and files to leave out apply to hosted landings only", land.ErrInvalidOptions))
		return
	} else {
		landing, err = landTodoRun(r.Context(), backed, todo, opts)
	}
	// Even a failed landing may have moved the checkout or removed the worktree.
	s.touchGit(source.Dir)
	if landing != nil {
		s.notify()
	}
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
	var conflict *branchmerge.ConflictError
	var prConflict *prcreate.ConflictError
	var hostedRefusal hostedLandingRefusal
	switch {
	case errors.Is(err, land.ErrInvalidOptions), errors.As(err, &hostedRefusal), errors.Is(err, sandbox.ErrInvalidLeaveOut):
		return http.StatusBadRequest
	case errors.Is(err, land.ErrNotNative):
		return http.StatusNotImplemented
	case errors.As(err, &conflict), errors.As(err, &prConflict),
		errors.Is(err, native.ErrAlreadyLanded), errors.Is(err, land.ErrNoRunWorkspace),
		errors.Is(err, land.ErrNoCommits), errors.Is(err, land.ErrWorktreeDirty),
		errors.Is(err, branchmerge.ErrCheckoutNotReady):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
