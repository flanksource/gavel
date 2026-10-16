// SandboxJobs starts a remote run's job as an agent run of the sandbox dispatcher: the todo's
// repository checked out at the commit serve resolved the run from.
package remote

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flanksource/gavel/sandbox"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
)

// SandboxJobs starts the jobs of one todo's runs.
type SandboxJobs struct {
	Dispatcher *sandbox.Dispatcher
	Scope      tenant.Scope
	RepoID     uuid.UUID
	// SHA is the commit of the repository's default branch the run was resolved from.
	SHA    string
	TodoID uuid.UUID
}

// Start starts job as agent run id.
func (s SandboxJobs) Start(ctx context.Context, id uuid.UUID, job Job) (Started, error) {
	encoded, err := json.Marshal(job)
	if err != nil {
		return nil, fmt.Errorf("encode the job of run %s: %w", id, err)
	}
	todo := s.TodoID
	run, err := s.Dispatcher.StartRun(ctx, id, s.Scope, sandbox.RunSpec{
		Kind: "agent", RepoID: s.RepoID, Ref: "HEAD", SHA: s.SHA, TodoID: &todo, Agent: encoded,
	}, nil)
	if err != nil {
		return nil, err
	}
	return sandboxRun{run: run}, nil
}

type sandboxRun struct{ run *sandbox.Run }

func (r sandboxRun) Wait(ctx context.Context) (int, error) {
	result, err := r.run.Wait(ctx)
	return result.ExitCode, err
}

func (r sandboxRun) Cancel(ctx context.Context) error { return r.run.Cancel(ctx) }
