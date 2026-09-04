package run

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
)

// A Continuation is a run derived from one that already happened: approving a
// reviewed plan into its implementing run, revising that plan, or answering the
// questions an ask turn left behind. Every entrypoint that continues a run —
// the dashboard's review actions and `gavel todos plan` — builds its options
// here. They used to build them independently, which is why two of them forgot
// to inherit the runtime the previous turn resolved and handed a codex session
// to claude.
//
// Prior is the prompt run being continued — the durable record of the spec that
// turn was dispatched with and the runtime it actually resolved. It is nil when
// the provider keeps no run history, in which case a continuation resolves
// exactly like a fresh run.
type Continuation struct {
	Dir      string
	Provider todos.Provider
	Todo     *types.TODO
	Prior    *captaindb.PromptRun
	// Override is what the caller explicitly asked for — validated wire options
	// or parsed flags — and it outranks everything inherited from Prior. Its Host
	// names the entrypoint; its Step, if it names one, must be the
	// continuation's.
	Override Options
	// Step is the lifecycle step the continuation runs.
	Step string
	// Resume continues Prior's conversation rather than opening a new one. It is
	// an explicit decision per call site, never a leftover option: a continuation
	// that does not resume must not inherit a session.
	Resume bool
	// Message is the user turn a resumed session continues with.
	Message string
}

// Continue resolves the run options a continuation executes with.
//
// What is inherited depends on whether the behaviour class changes. A
// continuation that stays in the same class continues that run's
// configuration: its dispatched spec is one layer, concretised by the runtime
// the turn resolved. A class change is not a continuation of configuration — a
// plan's read-only posture and its investigation budget belong to planning — so
// across one only the runtime selection carries, and everything else resolves
// from `.gavel.yaml` for the new step.
//
// The inherited spec is a LAYER, not a merge performed here. Folding it in by
// hand meant a second definition of precedence that could — and did — disagree
// with the one every other entrypoint resolves through; as layers, the prior run
// sits below the host and the caller exactly like any other authored default.
func Continue(c Continuation) (Options, error) {
	step := strings.TrimSpace(c.Step)
	if step == "" {
		return Options{}, fmt.Errorf("continuation names no lifecycle step")
	}
	if requested := strings.TrimSpace(c.Override.Step); requested != "" && requested != step {
		return Options{}, fmt.Errorf("this action runs the %s step; options.step %q cannot change it", step, requested)
	}
	if c.Override.Host == "" {
		return Options{}, fmt.Errorf("continuation of the %s step names no host", step)
	}
	class, err := stepClass(c.Provider, c.Dir, c.Override.Host, step)
	if err != nil {
		return Options{}, err
	}
	prior, err := PriorLayers(c.Prior, class, c.Resume)
	if err != nil {
		return Options{}, err
	}
	opts := c.Override
	opts.Step, opts.Prior, opts.Resume, opts.Message = step, prior, c.Resume, c.Message
	return opts, nil
}

// stepClass is the behaviour class of a named step in the workspace's
// lifecycle, which is what decides whether a prior run's configuration is
// continued or only its runtime selection.
func stepClass(provider todos.Provider, dir string, host lifecycle.HostKind, name string) (types.RunMode, error) {
	h, err := lifecycle.NewHost(provider, dir, host)
	if err != nil {
		return "", err
	}
	step, ok := h.Def.Definition().Step(name)
	if !ok {
		return "", fmt.Errorf("step %q is not part of lifecycle %s; steps: %s",
			name, h.Def.Definition().Name, strings.Join(h.Def.Definition().StepNames(), ", "))
	}
	return lifecycle.Class(step), nil
}

// PriorLayers projects the run being continued onto the user-scope layers below
// the caller's own request. See lifecycle.LayerInput.Prior.
func PriorLayers(prior *captaindb.PromptRun, class types.RunMode, resume bool) ([]api.SpecLayer, error) {
	spec, err := priorRunSpec(prior, class)
	if err != nil {
		return nil, err
	}
	if !resume {
		spec = spec.WithoutSession()
	}
	return []api.SpecLayer{
		api.RequestSpecLayer("prior run spec", spec),
		api.RequestSpecLayer("prior run runtime", priorRunRuntime(prior)),
	}, nil
}

// priorRunSpec returns the spec the prior run was dispatched with, as the layer
// a same-class continuation builds on. A class change inherits nothing from it.
//
// Two things are stripped even then. The prompt is the previous turn's user
// message — the prompt renderer treats a non-empty Prompt.User as the request,
// so replaying it would re-send instructions the agent has already acted on.
// Setup is post-transform: the captain setup hook consumes Checkout and
// rewrites Cwd to the tree it produced, so replaying it would pin the
// continuation to a workspace the previous run owned instead of letting
// configuration materialise one. (Setup.Env is `json:"-"` and never
// round-trips regardless.)
func priorRunSpec(prior *captaindb.PromptRun, class types.RunMode) (api.Spec, error) {
	if prior == nil || len(prior.RenderedSpec) == 0 || prior.Runtime.Mode != string(class) {
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

// priorRunRuntime is what a continuation inherits across any class change: the
// model, mode and effort the prior turn actually resolved. It is the concrete
// selection behind a family alias, so a codex session can never be continued by
// claude.
func priorRunRuntime(prior *captaindb.PromptRun) api.Spec {
	if prior == nil {
		return api.Spec{}
	}
	resolved := prior.Runtime.Resolved
	return api.Spec{Model: api.Model{
		Name:   strings.TrimSpace(resolved.Model),
		Mode:   api.RuntimeMode(strings.TrimSpace(resolved.Mode)),
		Effort: api.Effort(strings.TrimSpace(resolved.Effort)),
	}}
}

// ActivePromptRunProvider exposes the prompt run backing a todo's current
// attempt. The native PostgreSQL runtime implements it; a provider that keeps
// no run history simply has none to report.
type ActivePromptRunProvider interface {
	ActivePromptRun(context.Context, *types.TODO) (*captaindb.PromptRun, error)
}

// PriorRun returns the prompt run currently attached to the todo — the record a
// continuation inherits from — or nil when there is none; every other failure
// is real.
func PriorRun(ctx context.Context, provider todos.Provider, todo *types.TODO) (*captaindb.PromptRun, error) {
	runs, ok := provider.(ActivePromptRunProvider)
	if !ok {
		return nil, nil
	}
	return runs.ActivePromptRun(ctx, todo)
}
