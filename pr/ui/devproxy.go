package ui

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/flanksource/commons/logger"
)

var embeddedDevAssets = assetsHandler()

// SetDevProxy switches the "/" catch-all to a reverse proxy targeting a running
// Vite dev server, so `pr list --ui --dev` serves hot-reloaded modules instead
// of the embedded production bundle. /api/* stays on the Go server through
// more-specific mux patterns, while handleDevRoute keeps embedded assets local.
// Returns an error on a malformed target rather than silently disabling dev mode.
func (s *Server) SetDevProxy(target string) error {
	u, err := url.Parse(target)
	if err != nil {
		return fmt.Errorf("parse vite dev url %q: %w", target, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("vite dev url %q must be absolute (scheme://host)", target)
	}

	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Warnf("dev proxy %s%s: %v", target, r.URL.Path, err)
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "vite dev server unreachable at %s — is `pnpm dev` running in pr/ui?\n", target)
	}
	s.devProxy = proxy
	return nil
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
