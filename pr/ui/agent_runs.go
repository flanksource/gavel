// Hosted agent runs: a hosted todo's step is admitted and recorded here while its prompt runs in
// a sandbox job of the todo's repository.
package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/remote"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

// AgentRuns runs hosted todos' steps on substrate, and their test and lint runs in sandbox jobs.
type AgentRuns struct {
	Dispatcher *sandbox.Dispatcher
	// Substrate is the substrate backend agent runs execute on through captain's substrate
	// sandbox, as its options spell it; nil refuses agent runs.
	Substrate map[string]any
	// Sources has the repositories' git directories, which a run starts from and its commits
	// come back to.
	Sources *sandbox.Sources
	// Transcripts keeps the transcripts agent runs send.
	Transcripts RunTranscripts
	// RunAPI is serve's run API, which answers what an agent run asks of its PR.
	RunAPI http.Handler
}

// SetAgentRuns runs the organization's todo steps in sandbox jobs. Call it before serving.
func (s *Server) SetAgentRuns(runs AgentRuns) {
	s.agentRuns = &runs
}

// requireTodoChangeable fails with native.ErrForbidden unless the request's user may change todo.
func requireTodoChangeable(ctx context.Context, todo *types.TODO) error {
	id, err := uuid.Parse(todo.ID)
	if err != nil {
		return fmt.Errorf("todo %q has no native ID: %w", todo.ID, err)
	}
	var createdBy *uuid.UUID
	if todo.CreatedBy != "" {
		creator, err := uuid.Parse(todo.CreatedBy)
		if err != nil {
			return fmt.Errorf("todo %s names creator %q, which isn't a user ID: %w", todo.ID, todo.CreatedBy, err)
		}
		createdBy = &creator
	}
	return native.RequireChangeable(ctx, id, createdBy)
}

var errAgentRunsNeedSubstrate = errors.New("agent runs need a substrate sandbox: start gavel serve with --sandbox substrate")

// attachRemote runs req in a sandbox when source is one of a hosted organization's repositories,
// and only for the todo's creator or an org admin. The run acts for the request's user, so its
// tool calls, answered on serve with its approvals, see and change only what that user may. A
// local source's run is left as it is.
func (s *Server) attachRemote(ctx context.Context, source todoSource, req *run.Request) error {
	if source.Hosted == nil {
		return nil
	}
	if err := requireTodoChangeable(ctx, req.Todo); err != nil {
		return err
	}
	if err := requireRepositoryRuntime(req.Options); err != nil {
		return err
	}
	runner, err := s.remoteRunner(ctx, *source.Hosted, req.Todo)
	if err != nil {
		return err
	}
	req.Remote, req.PRs, req.Scope = runner, s.organizationPRs, tenant.FromContext(ctx)
	return nil
}

// requireRepositoryRuntime refuses a hosted run that chooses its own runtime: each step runs
// with the runtime the repository's configuration resolves it to.
func requireRepositoryRuntime(options run.Options) error {
	model := options.Request.Model
	var chosen string
	switch {
	case model.Name != "":
		chosen = "model " + model.Name
	case model.ID != "":
		chosen = "model " + model.ID
	case model.Provider != nil:
		chosen = "provider " + model.Provider.Name
	case model.Mode != "":
		chosen = "mode " + string(model.Mode)
	case len(model.Fallbacks) > 0:
		chosen = "fallback models"
	case len(options.Presets) > 0:
		chosen = "presets " + strings.Join(options.Presets, ", ")
	default:
		return nil
	}
	return fmt.Errorf("a hosted run uses the repository's runtime for each step; this one asks for %s", chosen)
}

// organizationPRs is the client the Server's organization's PR at url is read and acted on
// through.
func (s *Server) organizationPRs(ctx context.Context, url string) (provider.Client, int, error) {
	return OrganizationPRs(s.ghOpts.TokenSource, s.cacheStore())(ctx, url)
}

// OrganizationPRs reads an organization's PRs with token, its GitHub App installation's token
// source, never the host's credentials, keeping GitHub's responses in responses, its own cache.
// Hosted todos work on GitHub PRs only.
func OrganizationPRs(token func() (string, error), responses *cache.Store) func(ctx context.Context, url string) (provider.Client, int, error) {
	return func(ctx context.Context, url string) (provider.Client, int, error) {
		if provider.IsAzure(url) {
			return nil, 0, fmt.Errorf("hosted todos work on GitHub PRs; %s isn't one", url)
		}
		if token == nil {
			return nil, 0, fmt.Errorf("this organization has no GitHub App installation to read %s with", url)
		}
		return provider.ResolveURL(provider.Options{Context: ctx, TokenSource: token, Cache: responses}, url)
	}
}

// remoteRunner runs todo's steps on substrate, from workspace's repository, as the request's user.
func (s *Server) remoteRunner(ctx context.Context, workspace HostedWorkspace, todo *types.TODO) (lifecycle.RemoteRunner, error) {
	if s.agentRuns == nil || s.agentRuns.Substrate == nil {
		return nil, errAgentRunsNeedSubstrate
	}
	id, err := uuid.Parse(todo.ID)
	if err != nil {
		return nil, fmt.Errorf("todo %q has no native ID to run under: %w", todo.ID, err)
	}
	return &remote.SubstrateRunner{
		Dispatcher: s.agentRuns.Dispatcher, Sources: s.agentRuns.Sources, Backend: s.agentRuns.Substrate,
		Scope: tenant.FromContext(ctx), Repo: workspace.Repo, SHA: workspace.SHA, TodoID: id,
		Transcripts: s.agentRuns.Transcripts.Keep, RunAPI: s.agentRuns.RunAPI,
	}, nil
}
