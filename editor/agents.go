package editor

import (
	"net/http"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/gavel/comments/agents"
)

// managedPrompt is the library prompt a new editor agent session runs by default.
const managedPrompt = "editor-comments"

// captainSessions serves captain's session routes once gavel's pool is captain's.
func captainSessions() http.Handler {
	sessions := http.StripPrefix("/api/captain/sessions", captaincli.SessionHandler())
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := agents.OpenCaptain(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		sessions.ServeHTTP(w, r)
	})
}

// captainPromptRuns serves captain's prompt-run routes (snapshot, stream, stop)
// for the runs the editor launched, once gavel's pool is captain's.
func captainPromptRuns() http.Handler {
	runs := http.StripPrefix("/api/captain/prompt/runs", captaincli.PromptRunHandler())
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := agents.OpenCaptain(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		runs.ServeHTTP(w, r)
	})
}

// permissionCatalog serves GET /api/file/{id}/agent-run/permissions?provider&mode:
// captain's permission catalog (tools, MCP servers, plugins, skills) for that
// runtime, rooted at the project holding the file, wherever the editor was started.
func (s *Server) permissionCatalog(w http.ResponseWriter, r *http.Request) {
	doc, err := s.Store.Read(r.PathValue("id"))
	if err != nil {
		respond(w, nil, err)
		return
	}
	target, err := targetOf(r.Context(), doc.Path)
	if err != nil {
		respond(w, nil, err)
		return
	}
	if err := agents.OpenCaptain(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	captaincli.PermissionCatalogHandler(target.Dir).ServeHTTP(w, r)
}
