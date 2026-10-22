package ui

import (
	"fmt"
	"net/http"
	"strings"
)

var embeddedDevAssets = assetsHandler()

// SetDevProxy switches the "/" catch-all to proxy, the reverse proxy to a
// running Vite dev server (vitedev.Proxy), so `pr list --ui --dev` serves the
// source modules instead of the embedded production bundle. /api/* stays on
// the Go server through more-specific mux patterns, while handleDevRoute keeps
// embedded assets local.
func (s *Server) SetDevProxy(proxy http.Handler) {
	s.devProxy = proxy
}

// handleDevRoute is the dev-mode "/" handler. Embedded production assets stay
// reachable so a tab loaded before a prod/dev restart can still fetch its lazy
// chunks instead of receiving Vite's HTML fallback. Server-side export URLs
// still render via handleExport; everything else is proxied to Vite.
//
// An /api/ path reaching this catch-all has no Go route — typically a server
// built before the route existed. Vite would answer it 200 with index.html,
// which the UI can only report as "invalid JSON", so it fails as a JSON 404.
func (s *Server) handleDevRoute(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		w.Header().Set("Content-Type", "application/json")
		writeJSONError(w, http.StatusNotFound, fmt.Errorf("no API route for %s %s", r.Method, r.URL.Path))
		return
	}
	if strings.HasPrefix(r.URL.Path, "/_assets/") {
		embeddedDevAssets.ServeHTTP(w, r)
		return
	}
	if req, ok := parseRouteRequest(r); ok && req.IsExport {
		s.handleExport(w, r, req)
		return
	}
	s.devProxy.ServeHTTP(w, r)
}
