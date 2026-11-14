package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/projects"
	"github.com/flanksource/gavel/settings"
	"github.com/flanksource/gavel/verify"
)

// SettingsRoutes routes a new human-review thread to the comment agent its
// target's settings configure (comments.agents across the target's layers).
type SettingsRoutes struct {
	Settings settings.Service
	// Locate is the settings target t lies at in the host; ok is false for a
	// target the host's settings do not cover (it is not routed).
	Locate func(ctx context.Context, t comments.Target) (target settings.Target, ok bool, err error)
}

var _ comments.CommentAgents = SettingsRoutes{}

// AgentSettings is a target's settings layers and the comment agent they
// resolve to, each field traced to the layer that set it.
type AgentSettings struct {
	Settings settings.Resolved           `json:"settings"`
	Agent    verify.ResolvedCommentAgent `json:"agent"`
}

func (s SettingsRoutes) RouteFor(ctx context.Context, c comments.Comment) (comments.Route, bool, error) {
	if c.WorkflowID != comments.HumanReview().ID {
		return comments.Route{}, false, nil
	}
	target, ok, err := s.Locate(ctx, c.Target)
	if err != nil || !ok {
		return comments.Route{}, false, err
	}
	resolved, err := s.agentSettings(ctx, target)
	agent := resolved.Agent
	if err != nil || !agent.Found || !agent.Agent.IsEnabled() {
		return comments.Route{}, false, err
	}
	deliver, err := AgentTarget(agent.Agent)
	if err != nil {
		return comments.Route{}, false, err
	}
	return comments.Route{Source: agent.Provenance["enabled"].Layer, Target: deliver, Assign: agent.Agent.Assigns()}, true, nil
}

// Agent is t's settings and the comment agent they resolve to.
func (s SettingsRoutes) Agent(ctx context.Context, t comments.Target) (AgentSettings, error) {
	target, err := s.locate(ctx, t)
	if err != nil {
		return AgentSettings{}, err
	}
	return s.agentSettings(ctx, target)
}

// CommentAgent answers t's AgentSettings.
func (s SettingsRoutes) CommentAgent(ctx context.Context, t comments.Target) (any, error) {
	return s.Agent(ctx, t)
}

// SaveCommentAgent saves req.Agent (a verify.CommentAgent, or null to remove
// the entry) at req.Layer (a settings.Ref) of the target's chain and answers
// its AgentSettings.
func (s SettingsRoutes) SaveCommentAgent(ctx context.Context, req comments.SaveCommentAgentRequest) (any, error) {
	target, err := s.locate(ctx, req.Target)
	if err != nil {
		return nil, err
	}
	var layer settings.Ref
	if err := decodeStrict(req.Layer, &layer); err != nil {
		return nil, fmt.Errorf("%w: comment agent layer: %v", comments.ErrInvalid, err)
	}
	var agent *verify.CommentAgent
	if err := decodeStrict(req.Agent, &agent); err != nil {
		return nil, fmt.Errorf("%w: comment agent: %v", comments.ErrInvalid, err)
	}
	if err := settings.SaveCommentAgent(ctx, s.Settings, target, layer, req.Page, agent); err != nil {
		return nil, err
	}
	return s.agentSettings(ctx, target)
}

func (s SettingsRoutes) locate(ctx context.Context, t comments.Target) (settings.Target, error) {
	target, ok, err := s.Locate(ctx, t)
	if err == nil && !ok {
		err = fmt.Errorf("%w: this host's settings do not cover %s", comments.ErrInvalid, t.URI)
	}
	return target, err
}

func (s SettingsRoutes) agentSettings(ctx context.Context, target settings.Target) (AgentSettings, error) {
	resolved, err := s.Settings.Resolve(ctx, target)
	if err != nil {
		return AgentSettings{}, err
	}
	agent, err := resolved.CommentAgent(target)
	return AgentSettings{Settings: resolved, Agent: agent}, err
}

func decodeStrict(data json.RawMessage, out any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
}

// ProjectRoutes routes the threads of gavel://project/<name> scopes on its
// files (target URIs <scope>/files/<path>) and on its dashboard pages
// (<scope>/pages/<page key>); the scope's other targets, such as a pull
// request, are not routed.
func ProjectRoutes(catalog projects.Service, configured settings.Service) SettingsRoutes {
	return SettingsRoutes{Settings: configured, Locate: func(ctx context.Context, t comments.Target) (settings.Target, bool, error) {
		if !strings.HasPrefix(t.ScopeURI, "gavel://project/") {
			return settings.Target{}, false, nil
		}
		var target settings.Target
		var err error
		switch {
		case strings.HasPrefix(t.URI, t.ScopeURI+"/files/"):
			target, err = ProjectTarget(ctx, catalog, t.ScopeURI, t.URI)
		case strings.HasPrefix(t.URI, t.ScopeURI+"/pages/"):
			target, err = ProjectPageTarget(ctx, catalog, t.ScopeURI, t.URI)
		default:
			return settings.Target{}, false, nil
		}
		return target, err == nil, err
	}}
}

// ProjectPageTarget is the settings target of a dashboard page of the project
// a gavel://project/<name> scope names (uri <scope>/pages/<page key>): the
// key pages/<page key> in the project's database layers, the way a web
// project's pages are kept, so a page entry is keyed pages/<page key> at the
// project layer. The project's .gavel.yaml files configure its files, not its
// pages.
func ProjectPageTarget(ctx context.Context, catalog projects.Service, scope, uri string) (settings.Target, error) {
	project, err := scopeProject(ctx, catalog, scope)
	if err != nil {
		return settings.Target{}, err
	}
	key := strings.TrimPrefix(uri, scope+"/")
	if !strings.HasPrefix(uri, scope+"/pages/") || !verify.IsCommentPageKey(key) {
		return settings.Target{}, fmt.Errorf("%w: %s is not a page under %s", comments.ErrInvalid, uri, scope)
	}
	return settings.Target{Key: key, Project: &settings.Project{ID: project.ID}}, nil
}

// scopeProject is the project a gavel://project/<name> scope names.
func scopeProject(ctx context.Context, catalog projects.Service, scope string) (projects.Project, error) {
	name, err := url.PathUnescape(strings.TrimPrefix(scope, "gavel://project/"))
	if err != nil || name == scope {
		return projects.Project{}, fmt.Errorf("%w: scope %s is not a gavel://project scope", comments.ErrInvalid, scope)
	}
	project, err := catalog.Get(ctx, name)
	if errors.Is(err, projects.ErrNotFound) {
		return project, fmt.Errorf("%w: scope %s: %v", comments.ErrInvalid, scope, err)
	}
	return project, err
}

// ProjectFile is the project a gavel://project/<name> scope names, its
// canonical root, and the absolute path of the file uri names under <scope>/files/.
func ProjectFile(ctx context.Context, catalog projects.Service, scope, uri string) (projects.Project, string, string, error) {
	rel, ok := strings.CutPrefix(uri, scope+"/files/")
	if !ok || !verify.IsCommentPageKey(rel) {
		return projects.Project{}, "", "", fmt.Errorf("%w: %s is not a file under %s", comments.ErrInvalid, uri, scope)
	}
	project, err := scopeProject(ctx, catalog, scope)
	if err != nil {
		return projects.Project{}, "", "", err
	}
	root, err := projects.CanonicalPath(project.ResolvedDir())
	if err != nil {
		return projects.Project{}, "", "", fmt.Errorf("resolve project %q directory: %w", project.Name, err)
	}
	return project, root, filepath.Join(root, filepath.FromSlash(rel)), nil
}

// ProjectTarget is the settings target of the project file uri names.
func ProjectTarget(ctx context.Context, catalog projects.Service, scope, uri string) (settings.Target, error) {
	project, root, path, err := ProjectFile(ctx, catalog, scope, uri)
	if err != nil {
		return settings.Target{}, err
	}
	return settings.Target{Path: path, Project: &settings.Project{ID: project.ID, Dir: root}}, nil
}

// WebRoutes routes the threads of a host whose scope is a web project: the
// project named by the scope URI (created on first use), with each target URI
// a slash-separated path beneath its scope.
func WebRoutes(catalog projects.Service, configured settings.Service) SettingsRoutes {
	return SettingsRoutes{Settings: configured, Locate: func(ctx context.Context, t comments.Target) (settings.Target, bool, error) {
		target, err := WebTarget(ctx, catalog, t.ScopeURI, t.URI)
		return target, err == nil, err
	}}
}

// WebTarget is the settings target of uri in the web project of scope.
func WebTarget(ctx context.Context, catalog projects.Service, scope, uri string) (settings.Target, error) {
	key, ok := strings.CutPrefix(uri, scope+"/")
	if !ok || !verify.IsCommentPageKey(key) {
		return settings.Target{}, fmt.Errorf("%w: target %s is not a path beneath scope %s", comments.ErrInvalid, uri, scope)
	}
	project, err := catalog.EnsureWeb(ctx, scope)
	if err != nil {
		return settings.Target{}, fmt.Errorf("the web project of scope %s: %w", scope, err)
	}
	return settings.Target{Key: key, Project: &settings.Project{ID: project.ID}}, nil
}

// AgentTarget is the captain run a configured comment agent starts.
func AgentTarget(agent verify.CommentAgent) (comments.DeliverTarget, error) {
	target := comments.DeliverTarget{Kind: "run", Prompt: agent.PromptOrDefault(), GroupKey: agent.Group, Edits: agent.Edits}
	options, err := json.Marshal(RunOptions{Presets: agent.Presets, Spec: agent.Spec, Variables: agent.Variables, Instructions: agent.Instructions})
	if err != nil {
		return target, fmt.Errorf("encode the comment agent's run options: %w", err)
	}
	if string(options) != "{}" {
		target.RunOptions = options
	}
	return target, nil
}
