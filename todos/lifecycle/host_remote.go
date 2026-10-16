// Runs that execute elsewhere: a host with a RemoteRunner hands each step's prompt run to it — a
// sandbox job — and keeps admitting, recording and classifying the run itself.
package lifecycle

import (
	"context"
	"fmt"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
)

// RemoteRunner runs a step's prompt somewhere else and returns its result, as promptrun.Run
// would have.
type RemoteRunner interface {
	Run(ctx context.Context, run RemoteRun) (promptrun.Result, error)
}

// RemoteRun is one step's prompt run for a RemoteRunner.
type RemoteRun struct {
	// Input is the run as promptrun.Run would take it here. Its Hooks, Provider and Record
	// belong to this process: the runner rebuilds gavel's hooks where it runs from Hooks, and
	// files the run under Input.Record itself.
	Input promptrun.Input
	Hooks RunHooks
	// Admitted is called once Captain has admitted the run and before it is set up, as the hook
	// a local recorded run carries is. Nil for an unrecorded run.
	Admitted func() error
}

// remoteRunnable refuses a step a remote runner can't run: one that resumes an agent session,
// whose transcript the sandbox doesn't have.
func (h *Host) remoteRunnable(step Step, opts RunOptions) error {
	if h.Remote == nil {
		return nil
	}
	if opts.Resume {
		return fmt.Errorf("step %s resumes an agent session, which a run executing elsewhere can't do yet: start a fresh run", step.Name)
	}
	return nil
}

// remoteRuntime refuses a runtime a remote runner can't provide: cmux, whose agent runs in a
// terminal on this host.
func (h *Host) remoteRuntime(step Step, spec api.Spec) error {
	if h.Remote != nil && spec.Mode == api.ModeCmux {
		return fmt.Errorf("step %s runs in cmux, a terminal on this host, which a run executing elsewhere can't use: choose the agent or cli mode", step.Name)
	}
	return nil
}
