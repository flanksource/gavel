// Serves /api/commit-identity: the name and email the signed-in user's hosted runs commit as,
// over their GitHub identity. A commit identity belongs to a user, not an organization.
package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/flanksource/gavel/commitauthor"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
)

// commitIdentityUser returns the signed-in user's commit identity store, or answers why there
// isn't one.
func (s *Server) commitIdentityUser(w http.ResponseWriter, r *http.Request) (commitauthor.Store, uuid.UUID, bool) {
	scope := tenant.FromContext(r.Context())
	if scope.IsLocal() || scope.UserID == nil || s.orgDB == nil {
		writeJSONError(w, http.StatusNotFound, errors.New("a commit identity is only used by hosted runs (gavel serve --auth clerk); local commits use git's own"))
		return commitauthor.Store{}, uuid.Nil, false
	}
	return commitauthor.Store{DB: s.orgDB}, *scope.UserID, true
}

func (s *Server) handleCommitIdentityGet(w http.ResponseWriter, r *http.Request) {
	store, user, ok := s.commitIdentityUser(w, r)
	if !ok {
		return
	}
	identity, err := store.Get(r.Context(), user)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, identity)
}

func (s *Server) handleCommitIdentitySet(w http.ResponseWriter, r *http.Request) {
	store, user, ok := s.commitIdentityUser(w, r)
	if !ok {
		return
	}
	var body struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
		return
	}
	identity, err := store.Set(r.Context(), user, body.Name, body.Email)
	switch {
	case errors.Is(err, commitauthor.ErrInvalid):
		writeJSONError(w, http.StatusBadRequest, err)
	case err != nil:
		writeJSONError(w, http.StatusInternalServerError, err)
	default:
		writeJSON(w, identity)
	}
}
