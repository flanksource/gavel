package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	shared "github.com/flanksource/commons-db/cmd/query/connections"
	connectioncli "github.com/flanksource/commons-db/cmd/query/connections/cli"
	catalog "github.com/flanksource/gavel/connections"
	"github.com/flanksource/gavel/forge"
	"github.com/flanksource/gavel/tenant"
	"gorm.io/gorm"
)

// catalogBase is where the Settings connections page reaches the connection catalog, in the
// shape clicky-ui's ConnectionsPage reads: {base}/connection and {base}/connections/dashboard.
const catalogBase = "/api/catalog"

// hostedConnectionPolicy is what a hosted organization's catalog offers and accepts: only the
// git connections its repositories come from, with literal values, so none names a secret,
// config map or account of the server's to hydrate.
var hostedConnectionPolicy = shared.Policy{AllowedTypes: forge.GitTypes,
	SecretSources: []string{shared.SourceValue}, URLSources: []string{shared.SourceValue}}

// openConnectionCatalog opens the catalog of the request's organization, for a member to read and
// an admin to change; ok is false once it has answered the request with why it can't.
func (s *Server) openConnectionCatalog(w http.ResponseWriter, r *http.Request) (*shared.Service, *gorm.DB, tenant.Scope, func() error, bool) {
	scope := tenant.FromContext(r.Context())
	if scope.OrgID != s.orgID() {
		writeJSONError(w, http.StatusForbidden, fmt.Errorf("organization mismatch"))
		return nil, nil, scope, nil, false
	}
	if r.Method != http.MethodGet && !scope.IsLocal() && scope.Role != "org:admin" {
		writeJSONError(w, http.StatusForbidden, fmt.Errorf("organization admin required"))
		return nil, nil, scope, nil, false
	}
	db := s.orgDB
	if db == nil {
		db = s.catalogDB
	}
	if db == nil {
		writeJSONError(w, http.StatusServiceUnavailable, fmt.Errorf("connections require a database"))
		return nil, nil, scope, nil, false
	}
	options := catalog.Options{DB: db, Scope: scope}
	if s.hosted {
		options.Policy = hostedConnectionPolicy
	}
	service, close, err := catalog.Open(r.Context(), options)
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, err)
		return nil, nil, scope, nil, false
	}
	return service, db, scope, close, true
}

// handleConnectionCatalog serves the connections page's form schema and dashboard from the
// catalog's service. Testing or resolving a connection, and checking the fleet's health, dial it,
// which a hosted organization's may not have the server do.
func (s *Server) handleConnectionCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && s.hosted {
		writeJSONError(w, http.StatusForbidden, fmt.Errorf("a hosted organization's connections aren't tested from the server: it reaches only public addresses, and only when it syncs or reads their repositories"))
		return
	}
	service, _, _, close, ok := s.openConnectionCatalog(w, r)
	if !ok {
		return
	}
	defer func() { _ = close() }()
	service.Handler(catalogBase, http.NotFoundHandler()).ServeHTTP(w, r)
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	service, db, scope, close, ok := s.openConnectionCatalog(w, r)
	if !ok {
		return
	}
	result, err := connectionResult(service, db, w, r, scope, s.hosted)
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
	if r.Method != http.MethodGet && s.connectionsChanged != nil {
		s.connectionsChanged()
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

func connectionResult(service *shared.Service, db *gorm.DB, w http.ResponseWriter, r *http.Request, scope tenant.Scope, hosted bool) (any, error) {
	ref := r.PathValue("reference")
	switch {
	case r.Method == http.MethodGet && ref == "":
		return service.Inventory(shared.ListOptions{Type: r.URL.Query().Get("type")})
	case r.Method == http.MethodGet:
		return service.ResolveReference(ref)
	case r.Method == http.MethodDelete && hosted:
		// A hosted organization's repositories come from its connections, so they go with it.
		return nil, forge.DeleteConnection(r.Context(), db, scope, ref)
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
	if strings.HasPrefix(r.URL.Path, catalogBase+"/") {
		// The connections page sends back the connection it loaded; the fields the server keeps
		// are the server's to set.
		for _, field := range []string{"id", "org_id", "created_by", "created_at", "updated_at"} {
			delete(body, field)
		}
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
	if hosted {
		if err := hostedConnectionType(body["type"], ref == ""); err != nil {
			return nil, err
		}
	}
	if ref == "" {
		body["created_by"] = scope.UserID
		return service.Create(r.Context(), body)
	}
	return service.Update(r.Context(), ref, body)
}

// hostedConnectionType refuses a type other than the git ones a hosted organization's
// repositories come from: any other would have serve connect to whatever the organization names.
// An update that leaves the type alone needs none.
func hostedConnectionType(value any, required bool) error {
	kind, _ := value.(string)
	switch {
	case kind == "" && !required:
		return nil
	case kind == "":
		return fmt.Errorf("a hosted organization's connection needs a type: one of %s", strings.Join(forge.GitTypes, ", "))
	case !slices.Contains(forge.GitTypes, kind):
		return fmt.Errorf("a hosted organization keeps only git connections (%s), not %s", strings.Join(forge.GitTypes, ", "), kind)
	}
	return nil
}
