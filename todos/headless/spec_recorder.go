package headless

import (
	"github.com/flanksource/captain/pkg/ai/agent"
	todopkg "github.com/flanksource/gavel/todos"
)

// specRecorder reports the run's request once setup has transformed it — the
// checkout consumed, Cwd pointing at the tree the agent works in. It trails the
// setup plugin in the hook list and the runner dispatches PreRun in hook order,
// so what it observes is the spec the provider is dispatched with rather than
// the one the run was asked for. That distinction is the whole point: replaying
// a persisted spec that still asked for a checkout would clone a second tree.
type specRecorder struct {
	meta   todopkg.RunStartMetadata
	report func(todopkg.RunStartMetadata)
}

func (r *specRecorder) Name() string { return "gavel-spec-recorder" }

func (r *specRecorder) PreRun(hc *agent.HookContext) error {
	if hc.Request == nil {
		return nil
	}
	meta := r.meta
	spec := *hc.Request
	meta.Spec = &spec
	r.report(meta)
	return nil
}
