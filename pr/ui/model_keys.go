// Serves /api/model-keys: the organization's model API keys, which its sandbox runs on substrate
// use. Every member sees which keys are set; only admins change them, and keys never come back out.
package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/flanksource/gavel/modelkeys"
	"github.com/flanksource/gavel/tenant"
)

// modelKeyOrg returns the request's organization and its key store, or answers why there isn't
// one. admin requires the user to be an organization admin.
func (s *Server) modelKeyOrg(w http.ResponseWriter, r *http.Request, admin bool) (tenant.Scope, modelkeys.Store, bool) {
	scope := tenant.FromContext(r.Context())
	if scope.IsLocal() || s.orgDB == nil {
		writeJSONError(w, http.StatusNotFound, errors.New("model keys are only used by hosted sandbox runs (gavel serve --auth clerk); local runs use serve's environment"))
		return scope, modelkeys.Store{}, false
	}
	if admin && !scope.IsAdmin() {
		writeJSONError(w, http.StatusForbidden, errors.New("only an organization admin can change its model keys"))
		return scope, modelkeys.Store{}, false
	}
	return scope, modelkeys.Store{DB: s.orgDB}, true
}

func (s *Server) handleModelKeyList(w http.ResponseWriter, r *http.Request) {
	scope, store, ok := s.modelKeyOrg(w, r, false)
	if !ok {
		return
	}
	keys, err := store.List(r.Context(), scope.OrgID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, map[string]any{"kinds": modelkeys.Kinds, "keys": keys})
}

func (s *Server) handleModelKeySet(w http.ResponseWriter, r *http.Request) {
	scope, store, ok := s.modelKeyOrg(w, r, true)
	if !ok {
		return
	}
	var body struct {
		Kind string `json:"kind"`
		Key  string `json:"key"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
		return
	}
	provider := r.PathValue("provider")
	if kind, known := modelkeys.KindNamed(body.Kind); known && kind.Provider != provider {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("kind %s is a key for %s, not %s", kind.Name, kind.Provider, provider))
		return
	}
	switch err := store.Set(r.Context(), scope.OrgID, scope.UserID, body.Kind, body.Key); {
	case errors.Is(err, modelkeys.ErrInvalidKey):
		writeJSONError(w, http.StatusBadRequest, err)
		return
	case err != nil:
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	keys, err := store.List(r.Context(), scope.OrgID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	for _, key := range keys {
		if key.Provider == provider {
			writeJSON(w, key)
			return
		}
	}
	writeJSONError(w, http.StatusInternalServerError, fmt.Errorf("organization %s's %s key was saved but isn't listed", scope.OrgID, provider))
}

func (s *Server) handleModelKeyDelete(w http.ResponseWriter, r *http.Request) {
	scope, store, ok := s.modelKeyOrg(w, r, true)
	if !ok {
		return
	}
	switch err := store.Delete(r.Context(), scope.OrgID, r.PathValue("provider")); {
	case errors.Is(err, modelkeys.ErrNotFound):
		writeJSONError(w, http.StatusNotFound, err)
	case err != nil:
		writeJSONError(w, http.StatusInternalServerError, err)
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
