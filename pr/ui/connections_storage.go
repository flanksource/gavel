package ui

import (
	"fmt"
	"net/http"

	shared "github.com/flanksource/commons-db/cmd/query/connections"
	catalog "github.com/flanksource/gavel/connections"
	"github.com/flanksource/gavel/tenant"
)

func (s *Server) handleConnectionStorage(w http.ResponseWriter, r *http.Request) {
	scope := tenant.FromContext(r.Context())
	if scope.OrgID != s.orgID() {
		writeJSONError(w, http.StatusForbidden, fmt.Errorf("organization mismatch"))
		return
	}
	writeJSON(w, shared.StorageInfo{
		Layers:       []shared.StorageLayer{{Name: "db", Label: "Encrypted database", Writable: scope.IsLocal() || scope.Role == "org:admin"}},
		DefaultLayer: "db", KEK: catalog.Status(r.Context(), ""),
	})
}
