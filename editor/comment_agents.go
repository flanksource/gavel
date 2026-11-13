package editor

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/agents"
	"github.com/flanksource/gavel/projects"
	"github.com/flanksource/gavel/settings"
	"github.com/flanksource/gavel/verify"
)

// commentRoutes reads and saves the comment agents of project files, as the
// shared /api/v1/comments/agent route does.
func commentRoutes(ctx context.Context) (agents.SettingsRoutes, error) {
	catalog, err := projects.Shared(ctx)
	if err != nil {
		return agents.SettingsRoutes{}, err
	}
	configured, err := settings.Open(ctx, catalog)
	if err != nil {
		return agents.SettingsRoutes{}, err
	}
	return agents.ProjectRoutes(catalog, configured), nil
}

// fileSettings is GET /api/file/{id}/settings: the file's settings layers, each
// position of its chain and what can be saved there, and the comment agent
// that applies to it with each field traced to the layer that set it.
type fileSettings struct {
	Scope  string `json:"scope"`
	Target string `json:"target"`
	settings.Resolved
	Agent verify.ResolvedCommentAgent `json:"agent"`
}

// fileTargetOf is the comment target of the edited file id.
func (s *Server) fileTargetOf(ctx context.Context, id string) (comments.Target, error) {
	doc, err := s.Store.Read(id)
	if err != nil {
		return comments.Target{}, err
	}
	file, err := targetOf(ctx, doc.Path)
	if err != nil {
		return comments.Target{}, err
	}
	return comments.Target{ScopeURI: file.Scope, URI: file.URI}, nil
}

func (s *Server) fileSettings(ctx context.Context, id string) (fileSettings, error) {
	target, err := s.fileTargetOf(ctx, id)
	if err != nil {
		return fileSettings{}, err
	}
	routes, err := commentRoutes(ctx)
	if err != nil {
		return fileSettings{}, err
	}
	out, err := routes.Agent(ctx, target)
	if err != nil {
		return fileSettings{}, err
	}
	return fileSettings{Scope: target.ScopeURI, Target: target.URI, Resolved: out.Settings, Agent: out.Agent}, nil
}

func (s *Server) fileSettingsOf(w http.ResponseWriter, r *http.Request) {
	out, err := s.fileSettings(r.Context(), r.PathValue("id"))
	respond(w, out, err)
}

// saveFileAgent is PUT /api/file/{id}/settings/agent: it makes agent the
// comment agent at layer (its default, or with page its entry for the file),
// or removes that entry when agent is null, and answers the file's settings.
func (s *Server) saveFileAgent(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Layer json.RawMessage `json:"layer"`
		Page  bool            `json:"page"`
		Agent json.RawMessage `json:"agent"`
	}
	if !decode(w, r, &input) {
		return
	}
	principal, err := commentHost{store: s.Store}.Principal(r.Context())
	if err != nil {
		respond(w, nil, err)
		return
	}
	if principal.Actor.Kind != "user" {
		http.Error(w, "only a human configures comment agents", http.StatusForbidden)
		return
	}
	target, err := s.fileTargetOf(r.Context(), r.PathValue("id"))
	if err != nil {
		respond(w, nil, err)
		return
	}
	routes, err := commentRoutes(r.Context())
	if err != nil {
		respond(w, nil, err)
		return
	}
	if _, err := routes.SaveCommentAgent(r.Context(), comments.SaveCommentAgentRequest{Target: target, Layer: input.Layer, Page: input.Page, Agent: input.Agent}); err != nil {
		respond(w, nil, err)
		return
	}
	out, err := s.fileSettings(r.Context(), r.PathValue("id"))
	respond(w, out, err)
}
