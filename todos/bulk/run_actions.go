package bulk

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
)

// RunFlags are the parameters shared by the run-shaped bulk actions. They are
// the knobs a caller can vary per batch; everything else is resolved from
// .gavel.yaml and each TODO's own frontmatter, exactly as a single run is.
type RunFlags struct {
	Model  string `flag:"model" help:"Override the model for this batch, as the compact mode:model:effort form"`
	Effort string `flag:"effort" help:"Reasoning effort" enum:"low,medium,high"`
	Resume bool   `flag:"resume" help:"Resume each TODO's prior session instead of starting fresh"`
}

func (RunFlags) ClickyActionFlags() {}

// Spec projects the batch-level overrides onto the top resolution layer. It is
// deliberately partial: an empty field means "whatever the workspace and the
// TODO already say", not a default asserted here.
func (f RunFlags) Spec() api.Spec {
	spec := api.Spec{}
	if model := strings.TrimSpace(f.Model); model != "" {
		spec.Name = model
	}
	if effort := strings.ToLower(strings.TrimSpace(f.Effort)); effort != "" {
		spec.Effort = api.Effort(effort)
	}
	return spec
}

// RunRequest is what a resolver is asked to turn into run options.
type RunRequest struct {
	Dir string
	// Step is the lifecycle step this batch runs on every selected TODO.
	Step  string
	Todo  *types.TODO
	Flags RunFlags
}

// RunResolver produces the run options for one TODO.
//
// It is injected because resolving them is genuinely not uniform across
// entrypoints: the dashboard serves an approval endpoint, so a run of its can be
// admitted to ask for a tool approval where an unattended CLI batch cannot. A
// single hardcoded resolution would have to be wrong for one of them.
type RunResolver func(ctx context.Context, req RunRequest) (run.Options, error)

// DefaultRunResolver is the plain resolution: the batch's overrides as the
// request layer, and no approval endpoint to answer a prompt. Everything else
// — the prompt, the spec layers, the timeout — is the lifecycle's, folded by
// the host when the run resolves.
func DefaultRunResolver(_ context.Context, req RunRequest) (run.Options, error) {
	return run.Options{
		Step:    req.Step,
		Request: req.Flags.Spec(),
		Resume:  req.Flags.Resume,
		Host:    lifecycle.HostCLI,
	}, nil
}

// StartRun returns the item function for a named-step bulk action.
//
// run, plan and triage differ only in the step name — a step declares its own
// prompt, spec and outcomes — so no behaviour is asserted here and the
// lifecycle decides. That is what makes "triage these forty" and "plan these
// forty" one code path rather than three.
func StartRun(step string, flags RunFlags, registry *run.Registry, dir string, resolve RunResolver, broker todos.ApprovalBroker) (ItemFunc, error) {
	step = strings.TrimSpace(step)
	if step == "" {
		return nil, fmt.Errorf("lifecycle step name is required")
	}
	if registry == nil {
		return nil, fmt.Errorf("run registry is required")
	}
	if resolve == nil {
		resolve = DefaultRunResolver
	}

	return func(ctx context.Context, provider todos.Provider, todo *types.TODO) (ItemResult, error) {
		opts, err := resolve(ctx, RunRequest{Dir: dir, Step: step, Todo: todo, Flags: flags})
		if err != nil {
			return ItemResult{}, err
		}
		req := run.Request{
			Provider: provider,
			Registry: registry,
			Todo:     todo,
			Dir:      dir,
			Options:  opts,
			Broker:   broker,
		}
		// Resolving first turns a misconfigured run into a per-item error before
		// any agent session is admitted, so one bad TODO does not leave a
		// half-started batch behind it.
		if _, err := run.Resolve(ctx, req); err != nil {
			return ItemResult{}, err
		}
		started, err := run.Start(req)
		if err != nil {
			return ItemResult{}, err
		}
		return ItemResult{SessionID: started.SessionID, Status: started.Status}, nil
	}, nil
}
