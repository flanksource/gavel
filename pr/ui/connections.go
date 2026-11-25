package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	shared "github.com/flanksource/commons-db/cmd/query/connections"
	connectioncli "github.com/flanksource/commons-db/cmd/query/connections/cli"
	catalog "github.com/flanksource/gavel/connections"
	"github.com/flanksource/gavel/tenant"
	"gorm.io/gorm"
)

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	scope := tenant.FromContext(r.Context())
	if scope.OrgID != s.orgID() {
		writeJSONError(w, http.StatusForbidden, fmt.Errorf("organization mismatch"))
		return
	}
	if r.Method != http.MethodGet && !scope.IsLocal() && scope.Role != "org:admin" {
		writeJSONError(w, http.StatusForbidden, fmt.Errorf("organization admin required"))
		return
	}
	if s.orgDB == nil {
		writeJSONError(w, http.StatusServiceUnavailable, fmt.Errorf("connections require a database"))
		return
	}
	service, close, err := catalog.Open(r.Context(), catalog.Options{DB: s.orgDB, Scope: scope})
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, err)
		return
	}
	result, err := connectionResult(service, w, r, scope)
	if closeErr := close(); closeErr != nil {
		writeJSONError(w, http.StatusServiceUnavailable, errors.Join(err, closeErr))
		return
	}
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}
		writeJSONError(w, status, err)
		return
	}
	if result == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	data, err := json.Marshal(result)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	var safe any
	if err = json.Unmarshal(data, &safe); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err)
		return
	}
	connectioncli.Redact(safe)
	writeJSON(w, safe)
}

func connectionResult(service *shared.Service, w http.ResponseWriter, r *http.Request, scope tenant.Scope) (any, error) {
	ref := r.PathValue("reference")
	switch {
	case r.Method == http.MethodGet && ref == "":
		return service.Inventory(shared.ListOptions{Type: r.URL.Query().Get("type")})
	case r.Method == http.MethodGet:
		return service.ResolveReference(ref)
	case r.Method == http.MethodDelete:
		return nil, service.Delete(ref)
	case r.Method == http.MethodPost && ref != "":
		return service.Validate(r.Context(), shared.ValidateOptions{Reference: ref})
	}
	var body map[string]any
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		return nil, err
	}
	if body == nil {
		return nil, fmt.Errorf("connection body must be an object")
	}
	var extra any
	if trailing := decoder.Decode(&extra); trailing != io.EOF {
		return nil, fmt.Errorf("connection body must contain one JSON object")
	}
	if _, supplied := body["org_id"]; supplied {
		return nil, fmt.Errorf("organization comes from the authenticated scope")
	}
	if _, supplied := body["created_by"]; supplied {
		return nil, fmt.Errorf("creator comes from the authenticated scope")
	}
	if ref == "" {
		body["created_by"] = scope.UserID
		return service.Create(r.Context(), body)
	}
	return service.Update(r.Context(), ref, body)
}
