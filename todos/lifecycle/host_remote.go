// Runs that execute elsewhere: a host with a RemoteRunner hands each step's prompt run to it — a
// sandbox job — and keeps admitting, recording and classifying the run itself.
package lifecycle

import (
	"context"
	"fmt"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
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

// remoteRunnable refuses a step a remote runner can't run yet: one whose tools read todos
// through the provider while the agent works, and one that works on a linked PR.
func (h *Host) remoteRunnable(step Step, todo *types.TODO, definition todoprompt.Definition) error {
	if h.Remote == nil {
		return nil
	}
	if definition.Envelope == todoprompt.EnvelopeTriageNew {
		return fmt.Errorf("step %s reads todos while it runs, which a run executing elsewhere can't do yet", step.Name)
	}
	if todo.PR != nil && definition.Class != types.ModePlan {
		return fmt.Errorf("step %s works on linked PR %s, which can't run elsewhere yet", step.Name, todo.PR.URL)
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
