package bulk

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/run"
	todospec "github.com/flanksource/gavel/todos/spec"
	"github.com/flanksource/gavel/todos/types"
)

// RunFlags are the parameters shared by the run-shaped bulk actions. They are
// the knobs a caller can vary per batch; everything else is resolved from
// .gavel.yaml and each TODO's own frontmatter, exactly as a single run is.
type RunFlags struct {
	Model  string `flag:"model" help:"Override the model for this batch"`
	Effort string `flag:"effort" help:"Reasoning effort" enum:"low,medium,high"`
	Driver string `flag:"driver" help:"Execution driver" enum:"api,agent,cli,cmux"`
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
	Dir    string
	Prompt string
	Todo   *types.TODO
	Flags  RunFlags
}

// RunResolver produces the run options for one TODO.
//
// It is injected because resolving them is genuinely not uniform across
// entrypoints. The dashboard applies a (driver, backend) catalog the CLI has no
// equivalent of, and it serves an approval endpoint, so a run of its can be
// admitted to ask for a Bash approval where an unattended CLI batch cannot. A
// single hardcoded resolution would have to be wrong for one of them.
type RunResolver func(ctx context.Context, req RunRequest) (run.Options, error)

// DefaultRunResolver is the plain resolution: .gavel.yaml plus each TODO's own
// frontmatter, with the batch's overrides on top and no approval endpoint to
// answer a prompt.
func DefaultRunResolver(_ context.Context, req RunRequest) (run.Options, error) {
	resolved, err := todospec.Resolve(todospec.Input{
		WorkDir:  req.Dir,
		Prompt:   req.Prompt,
		Todos:    []*types.TODO{req.Todo},
		Override: req.Flags.Spec(),
		Driver:   strings.TrimSpace(req.Flags.Driver),
	})
	if err != nil {
		return run.Options{}, err
	}
	return run.Options{
		Spec:      resolved.Spec,
		Driver:    string(resolved.Driver),
		RunMode:   resolved.Mode,
		Prompt:    resolved.Prompt,
		Envelope:  resolved.Envelope,
		Resume:    req.Flags.Resume,
		Template:  resolved.Template,
		Approvals: resolved.Approvals,
		Timeout:   resolved.Timeout,
	}, nil
}

// StartRun returns the item function for a named-prompt bulk action.
//
// run, plan and triage differ only in the prompt name — a prompt declares its
// own behaviour class, so no mode is asserted here and the catalog decides.
// That is what makes "triage these forty" and "plan these forty" one code path
// rather than three.
func StartRun(prompt string, flags RunFlags, registry *run.Registry, dir string, resolve RunResolver) (ItemFunc, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return nil, fmt.Errorf("prompt name is required")
	}
	if registry == nil {
		return nil, fmt.Errorf("run registry is required")
	}
	if resolve == nil {
		resolve = DefaultRunResolver
	}

	return func(ctx context.Context, provider todos.Provider, todo *types.TODO) (ItemResult, error) {
		opts, err := resolve(ctx, RunRequest{Dir: dir, Prompt: prompt, Todo: todo, Flags: flags})
		if err != nil {
			return ItemResult{}, err
		}
		todoList := []*types.TODO{todo}
		opts.Spec.SessionID = run.ResolveSessionID(opts, todoList)
		req := run.Request{
			Provider: provider,
			Registry: registry,
			Todos:    todoList,
			Dir:      dir,
			Backend:  todos.ProviderDB,
			Options:  opts,
		}
		// Constructing the executor first turns a misconfigured run into a
		// per-item error before any agent session is admitted, so one bad TODO
		// does not leave a half-started batch behind it.
		if _, _, err := run.NewExecutorContext(ctx, req); err != nil {
			return ItemResult{}, err
		}
		started, err := run.Start(req)
		if err != nil {
			return ItemResult{}, err
		}
		return ItemResult{SessionID: started.SessionID, Status: started.Status}, nil
	}, nil
}
