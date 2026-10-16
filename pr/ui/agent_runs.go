// Hosted agent runs: a hosted todo's step is admitted and recorded here while its prompt runs in
// a sandbox job of the todo's repository.
package ui

import (
	"context"
	"errors"
	"fmt"

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

// AgentRuns runs hosted todos' steps in sandbox jobs.
type AgentRuns struct {
	Dispatcher *sandbox.Dispatcher
	// Relay hands each job's messages to the run waiting for them; the run API delivers them.
	Relay *remote.Relay
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

var errAgentRunsNeedSandbox = errors.New("agent runs need a sandbox: start gavel serve with --sandbox")

// attachRemote runs req in a sandbox when source is one of a hosted organization's repositories,
// and only for the todo's creator or an org admin: the run writes as a system job, so this is
// where the request's user is checked. A sandboxed run's tool calls and approvals are answered
// on serve. A local source's run is left as it is.
func (s *Server) attachRemote(ctx context.Context, source todoSource, req *run.Request) error {
	if source.Hosted == nil {
		return nil
	}
	if err := requireTodoChangeable(ctx, req.Todo); err != nil {
		return err
	}
	runner, err := s.remoteRunner(ctx, *source.Hosted, req.Todo)
	if err != nil {
		return err
	}
	req.Remote, req.PRs = runner, s.organizationPRs
	return nil
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

// remoteRunner runs todo's steps in sandbox jobs of workspace's repository, as the request's user.
func (s *Server) remoteRunner(ctx context.Context, workspace HostedWorkspace, todo *types.TODO) (lifecycle.RemoteRunner, error) {
	if s.agentRuns == nil {
		return nil, errAgentRunsNeedSandbox
	}
	id, err := uuid.Parse(todo.ID)
	if err != nil {
		return nil, fmt.Errorf("todo %q has no native ID to run under: %w", todo.ID, err)
	}
	return &remote.Runner{
		Jobs: remote.SandboxJobs{
			Dispatcher: s.agentRuns.Dispatcher, Scope: tenant.FromContext(ctx),
			RepoID: workspace.Repo.ID, SHA: workspace.SHA, TodoID: id,
		},
		Relay: s.agentRuns.Relay,
	}, nil
}
