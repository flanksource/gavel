package ui

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/flanksource/commons/logger"
)

// SetDevProxy switches the "/" catch-all to a reverse proxy targeting a running
// Vite dev server, so `pr list --ui --dev` serves hot-reloaded modules instead
// of the embedded production bundle. /api/* and the asset routes stay on the Go
// server because they are registered as more-specific mux patterns. Returns an
// error on a malformed target rather than silently disabling dev mode.
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

// handleDevRoute is the dev-mode "/" handler. Server-side export URLs
// (e.g. /prs?format=json) still render via handleExport; everything else —
// the SPA shell, /@vite/*, /src/*, client routes — is proxied to Vite.
func (s *Server) handleDevRoute(w http.ResponseWriter, r *http.Request) {
	if req, ok := parseRouteRequest(r); ok && req.IsExport {
		s.handleExport(w, r, req)
		return
	}
	s.devProxy.ServeHTTP(w, r)
}
