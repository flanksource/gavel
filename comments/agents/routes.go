package agents

import (
	"context"
	"encoding/json"
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
	// Locate is the settings target c's target lies at in the host; ok is
	// false for a target the host's settings do not cover (it is not routed).
	Locate func(ctx context.Context, c comments.Comment) (target settings.Target, ok bool, err error)
}

var _ comments.RouteResolver = SettingsRoutes{}

func (s SettingsRoutes) RouteFor(ctx context.Context, c comments.Comment) (comments.Route, bool, error) {
	if c.WorkflowID != comments.HumanReview().ID {
		return comments.Route{}, false, nil
	}
	target, ok, err := s.Locate(ctx, c)
	if err != nil || !ok {
		return comments.Route{}, false, err
	}
	resolved, err := s.Settings.Resolve(ctx, target)
	if err != nil {
		return comments.Route{}, false, err
	}
	agent, err := resolved.CommentAgent(target)
	if err != nil || !agent.Found || !agent.Agent.IsEnabled() {
		return comments.Route{}, false, err
	}
	deliver, err := AgentTarget(agent.Agent)
	if err != nil {
		return comments.Route{}, false, err
	}
	return comments.Route{Source: agent.Provenance["enabled"].Layer, Target: deliver, Assign: agent.Agent.Assigns()}, true, nil
}

// ProjectRoutes routes the threads on files of gavel://project/<name> scopes
// (target URIs <scope>/files/<path>); the scope's other targets, such as a
// pull request, are not routed.
func ProjectRoutes(catalog projects.Service, configured settings.Service) SettingsRoutes {
	return SettingsRoutes{Settings: configured, Locate: func(ctx context.Context, c comments.Comment) (settings.Target, bool, error) {
		if !strings.HasPrefix(c.Target.ScopeURI, "gavel://project/") || !strings.HasPrefix(c.Target.URI, c.Target.ScopeURI+"/files/") {
			return settings.Target{}, false, nil
		}
		target, err := ProjectTarget(ctx, catalog, c.Target.ScopeURI, c.Target.URI)
		return target, err == nil, err
	}}
}

// ProjectFile is the project a gavel://project/<name> scope names, its
// canonical root, and the absolute path of the file uri names under <scope>/files/.
func ProjectFile(ctx context.Context, catalog projects.Service, scope, uri string) (projects.Project, string, string, error) {
	name, err := url.PathUnescape(strings.TrimPrefix(scope, "gavel://project/"))
	if err != nil || name == scope {
		return projects.Project{}, "", "", fmt.Errorf("%w: scope %s is not a gavel://project scope", comments.ErrInvalid, scope)
	}
	rel, ok := strings.CutPrefix(uri, scope+"/files/")
	if !ok || !verify.IsCommentPageKey(rel) {
		return projects.Project{}, "", "", fmt.Errorf("%w: %s is not a file under %s", comments.ErrInvalid, uri, scope)
	}
	project, err := catalog.Get(ctx, name)
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
	return SettingsRoutes{Settings: configured, Locate: func(ctx context.Context, c comments.Comment) (settings.Target, bool, error) {
		target, err := WebTarget(ctx, catalog, c.Target.ScopeURI, c.Target.URI)
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
	target := comments.DeliverTarget{Kind: "run", Prompt: agent.PromptOrDefault()}
	options, err := json.Marshal(RunOptions{Presets: agent.Presets, Spec: agent.Spec, Variables: agent.Variables, Instructions: agent.Instructions})
	if err != nil {
		return target, fmt.Errorf("encode the comment agent's run options: %w", err)
	}
	if string(options) != "{}" {
		target.RunOptions = options
	}
	return target, nil
}
