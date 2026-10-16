// Hosted agent runs: a hosted todo's step is admitted and recorded here while its prompt runs in
// a sandbox job of the todo's repository.
package ui

import (
	"context"
	"errors"
	"fmt"

	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/remote"
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

var errAgentRunsNeedSandbox = errors.New("agent runs need a sandbox: start gavel serve with --sandbox")

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
