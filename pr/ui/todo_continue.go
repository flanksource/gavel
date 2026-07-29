package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos/drivers"
	"github.com/flanksource/gavel/todos/types"
)

// A continuation is a run derived from one that already happened: approving a
// reviewed plan into its implementing run, revising that plan, or answering the
// questions an ask turn left behind. All three used to build their run options
// independently, which is why two of them forgot to inherit the runtime the
// previous turn resolved and one reported a session id it never dispatched.
//
// Prior is the prompt run being continued — the durable record of the spec that
// turn was dispatched with and the runtime it actually resolved. It is nil when
// the provider keeps no run history, in which case a continuation resolves
// exactly like a fresh run.
type continuation struct {
	Dir   string
	Todos []*types.TODO
	Prior *captaindb.PromptRun
	// Override is what the caller explicitly asked for; it outranks everything
	// inherited from Prior.
	Override todoRunPayload
	Mode     types.RunMode
	// Resume continues Prior's conversation rather than opening a new one. It is
	// an explicit decision per call site, never a leftover payload field: a
	// continuation that does not resume must not inherit a session.
	Resume bool
}

// continueRun resolves the run options a continuation executes with.
//
// What is inherited depends on whether the mode changes. A continuation that
// stays in the same mode continues that run's configuration: its dispatched
// spec is the base, concretised by the runtime the turn resolved. A mode change
// is not a continuation of configuration — a plan's read-only posture and its
// investigation budget belong to planning — so across one only the runtime
// selection carries, and everything else resolves from `.gavel.yaml` for the
// new mode. Either way the caller's own options stay the highest layer.
func continueRun(c continuation) (todoRunOptions, error) {
	// A todo's recorded mode is empty until its first run; parsing it here means
	// the mode compared against the prior run is the one that will execute.
	mode, err := types.ParseRunMode(string(c.Mode))
	if err != nil {
		return todoRunOptions{}, err
	}
	c.Mode = mode

	prior, err := priorRunSpec(c.Prior, c.Mode)
	if err != nil {
		return todoRunOptions{}, err
	}

	payload := c.Override
	payload.Spec = prior.Merge(priorRunRuntime(c.Prior)).Merge(payload.Spec)
	if !c.Resume {
		payload.Spec = payload.Spec.WithoutSession()
	}
	payload.Driver = continuationDriver(payload, c.Prior)
	payload.RunMode = string(c.Mode)
	payload.Plan = false
	payload.Resume = c.Resume

	opts, err := normalizeTodoRunOptions(c.Dir, c.Todos, payload)
	if err != nil {
		return todoRunOptions{}, err
	}
	// The session id the run will actually use, decided the same way a fresh run
	// from /api/todos/run decides it — so a continuation can be reported and
	// attached to instead of answering with whatever the client happened to send.
	opts.Spec.SessionID = resolveRunSessionID(opts, c.Todos)
	return opts, nil
}

// priorRunSpec returns the spec the prior run was dispatched with, as the layer
// a same-mode continuation builds on. A mode change inherits nothing from it.
//
// Two things are stripped even then. The prompt is the previous turn's user
// message — todoprompt.Render treats a non-empty Prompt.User as the request, so
// replaying it would re-send instructions the agent has already acted on. Setup
// is post-transform: the captain setup hook consumes Checkout and rewrites Cwd
// to the tree it produced, so replaying it would pin the continuation to a
// workspace the previous run owned instead of letting configuration materialise
// one. (Setup.Env is `json:"-"` and never round-trips regardless.)
func priorRunSpec(prior *captaindb.PromptRun, mode types.RunMode) (api.Spec, error) {
	if prior == nil || len(prior.RenderedSpec) == 0 || prior.Runtime.Mode != string(mode) {
		return api.Spec{}, nil
	}
	data, err := json.Marshal(prior.RenderedSpec)
	if err != nil {
		return api.Spec{}, fmt.Errorf("encode prior run spec: %w", err)
	}
	var spec api.Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return api.Spec{}, fmt.Errorf("decode prior run spec: %w", err)
	}
	spec.Prompt = api.Prompt{}
	spec.Setup = nil
	// Verify.Fixture is a persistence stamp, not run configuration: nothing in
	// the run path reads it, and the next run re-stamps it from the issue.
	if spec.Workflow != nil && spec.Workflow.Verify != nil {
		spec.Workflow.Verify.Fixture = ""
	}
	return spec, nil
}

// priorRunRuntime is what a continuation inherits across any mode change: the
// model, backend and effort the prior turn actually resolved. It is the concrete
// selection behind a family alias, so a codex session can never be continued by
// claude.
func priorRunRuntime(prior *captaindb.PromptRun) api.Spec {
	if prior == nil {
		return api.Spec{}
	}
	resolved := prior.Runtime.Resolved
	return api.Spec{Model: api.Model{
		Name:    strings.TrimSpace(resolved.Model),
		Backend: api.Backend(strings.TrimSpace(resolved.Backend)),
		Effort:  api.Effort(strings.TrimSpace(resolved.Effort)),
	}}
}

// continuationDriver returns the driver mechanism a continuation runs on: the
// one the caller asked for, otherwise the prior run's. A run's recorded driver
// also carries non-driver runner labels (a verify step's "gavel-fixtures"),
// which fall through to `.gavel.yaml` todos.driver rather than erroring.
func continuationDriver(payload todoRunPayload, prior *captaindb.PromptRun) string {
	if strings.TrimSpace(payload.Driver) != "" || strings.TrimSpace(payload.Mode) != "" {
		return payload.Driver
	}
	if prior == nil {
		return ""
	}
	kind, err := drivers.Parse(prior.Runtime.Driver)
	if err != nil {
		return ""
	}
	return string(kind)
}
