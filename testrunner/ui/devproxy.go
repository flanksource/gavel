package testui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// SetDevProxy serves the page from a Vite dev server instead of the embedded
// bundle: proxy is the reverse proxy to it (vitedev.Proxy), for
// `gavel test|fixtures --ui --dev`. The /api/* routes stay on this server.
func (s *Server) SetDevProxy(proxy http.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devProxy = proxy
}

func (s *Server) devProxyHandler() http.Handler {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.devProxy
}

// serveDev is the dev-mode catch-all. Export URLs still render here; an /api/
// path reaching it has no Go route — Vite would answer it 200 with index.html,
// which the UI can only report as invalid JSON — so it is a JSON 404; every
// other path (the page, /src modules, /@vite and /@fs files) goes to Vite.
func (s *Server) serveDev(w http.ResponseWriter, r *http.Request, proxy http.Handler) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("no API route for %s %s", r.Method, r.URL.Path)}) //nolint:errcheck
		return
	}
	if req, ok := parseRouteRequest(r); ok && req.IsExport && exportableTab(req.Tab) {
		s.handleExport(w, r, req)
		return
	}
	proxy.ServeHTTP(w, r)
}
