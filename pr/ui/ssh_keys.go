// Serves /api/ssh-keys: the signed-in user's SSH public keys, which identify them to the
// hosted push server. Keys belong to a user, not an organization.
package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/flanksource/gavel/sshkeys"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
)

// sshKeyUser returns the signed-in user's key store, or answers why there isn't one.
func (s *Server) sshKeyUser(w http.ResponseWriter, r *http.Request) (sshkeys.Store, uuid.UUID, bool) {
	scope := tenant.FromContext(r.Context())
	if scope.IsLocal() || scope.UserID == nil || s.orgDB == nil {
		writeJSONError(w, http.StatusNotFound, errors.New("SSH keys are only used by the hosted push server (gavel serve --auth clerk)"))
		return sshkeys.Store{}, uuid.Nil, false
	}
	return sshkeys.Store{DB: s.orgDB}, *scope.UserID, true
}

func (s *Server) handleSSHKeyList(w http.ResponseWriter, r *http.Request) {
	store, user, ok := s.sshKeyUser(w, r)
	if !ok {
		return
	}
	keys, err := store.List(r.Context(), user)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, keys)
}

func (s *Server) handleSSHKeyAdd(w http.ResponseWriter, r *http.Request) {
	store, user, ok := s.sshKeyUser(w, r)
	if !ok {
		return
	}
	var body struct {
		Name      string `json:"name"`
		PublicKey string `json:"publicKey"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
		return
	}
	key, err := store.Add(r.Context(), user, body.Name, body.PublicKey)
	switch {
	case errors.Is(err, sshkeys.ErrUnsupportedKey):
		writeJSONError(w, http.StatusBadRequest, err)
	case errors.Is(err, sshkeys.ErrDuplicate):
		writeJSONError(w, http.StatusConflict, err)
	case err != nil:
		writeJSONError(w, http.StatusInternalServerError, err)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(key) //nolint:errcheck
	}
}

func (s *Server) handleSSHKeyDelete(w http.ResponseWriter, r *http.Request) {
	store, user, ok := s.sshKeyUser(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("%q is not an SSH key ID", r.PathValue("id")))
		return
	}
	switch err := store.Delete(r.Context(), user, id); {
	case errors.Is(err, sshkeys.ErrNotFound):
		writeJSONError(w, http.StatusNotFound, err)
	case err != nil:
		writeJSONError(w, http.StatusInternalServerError, err)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
