package editor

import (
	"context"
	"net/http"

	"github.com/flanksource/gavel/comments/agents"
	"github.com/flanksource/gavel/projects"
	"github.com/flanksource/gavel/settings"
	"github.com/flanksource/gavel/verify"
)

// settingsTarget is the settings target of a thread's file.
func settingsTarget(ctx context.Context, scope, uri string) (settings.Target, error) {
	catalog, err := projects.Shared(ctx)
	if err != nil {
		return settings.Target{}, err
	}
	return agents.ProjectTarget(ctx, catalog, scope, uri)
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

func (s *Server) fileSettings(ctx context.Context, id string) (fileSettings, error) {
	doc, err := s.Store.Read(id)
	if err != nil {
		return fileSettings{}, err
	}
	file, err := targetOf(ctx, doc.Path)
	if err != nil {
		return fileSettings{}, err
	}
	target, err := settingsTarget(ctx, file.Scope, file.URI)
	if err != nil {
		return fileSettings{}, err
	}
	service, err := settings.Shared(ctx)
	if err != nil {
		return fileSettings{}, err
	}
	resolved, err := service.Resolve(ctx, target)
	if err != nil {
		return fileSettings{}, err
	}
	agent, err := resolved.CommentAgent(target)
	if err != nil {
		return fileSettings{}, err
	}
	return fileSettings{Scope: file.Scope, Target: file.URI, Resolved: resolved, Agent: agent}, nil
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
		Layer settings.Ref         `json:"layer"`
		Page  bool                 `json:"page"`
		Agent *verify.CommentAgent `json:"agent"`
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
	doc, err := s.Store.Read(r.PathValue("id"))
	if err != nil {
		respond(w, nil, err)
		return
	}
	file, err := targetOf(r.Context(), doc.Path)
	if err != nil {
		respond(w, nil, err)
		return
	}
	target, err := settingsTarget(r.Context(), file.Scope, file.URI)
	if err != nil {
		respond(w, nil, err)
		return
	}
	service, err := settings.Shared(r.Context())
	if err != nil {
		respond(w, nil, err)
		return
	}
	if err := settings.SaveCommentAgent(r.Context(), service, target, input.Layer, input.Page, input.Agent); err != nil {
		respond(w, nil, err)
		return
	}
	out, err := s.fileSettings(r.Context(), doc.ID)
	respond(w, out, err)
}
