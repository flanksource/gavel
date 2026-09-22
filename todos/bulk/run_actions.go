package bulk

import (
	"context"
	"fmt"
	"path/filepath"
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
	Presets   []string `flag:"preset" help:"Runtime preset name or ID for this batch; repeat to layer presets in order"`
	NoPresets bool     `flag:"no-presets" help:"Clear configured runtime presets for this batch"`
	// Deprecated: accepted by programmatic callers only so resolution can warn.
	RuntimeProfile string `json:"runtimeProfile,omitempty"`
	Model          string `flag:"model" help:"Override the model for this batch, as the compact mode:model:effort form"`
	Effort         string `flag:"effort" help:"Reasoning effort" enum:"low,medium,high,xhigh"`
	Resume         bool   `flag:"resume" help:"Resume each TODO's prior session instead of starting fresh"`
	// A batch is where a wrong duplicate call costs most: forty verdicts land
	// unattended, and the two that close TODOs cannot be undone by a later run.
	Preview bool `flag:"preview" help:"Run the agent, but report a triage verdict that would close a TODO instead of applying it"`
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
	Step string
	Todo *types.TODO
	// Batch are the refs of every TODO in the selection, this one included. A
	// triage render marks them so the agent knows which backlog entries are having
	// their verdicts decided alongside the one it is looking at.
	Batch []string
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
	presetsSet := req.Flags.NoPresets || req.Flags.Presets != nil
	presets := append([]string(nil), req.Flags.Presets...)
	if req.Flags.NoPresets {
		presets = []string{}
	}
	return run.Options{
		Presets:        presets,
		PresetsSet:     presetsSet,
		RuntimeProfile: req.Flags.RuntimeProfile,
		Step:           req.Step,
		Request:        req.Flags.Spec(),
		Resume:         req.Flags.Resume,
		Batch:          append([]string(nil), req.Batch...),
		Preview:        req.Flags.Preview,
		Host:           lifecycle.HostCLI,
	}, nil
}

// RunSpec is what StartRun needs to build a named-step item function.
type RunSpec struct {
	// Step is the lifecycle step to run on every selected TODO.
	Step  string
	Flags RunFlags
	// Batch are the refs of the whole selection, passed to each run so a triage
	// render can mark the backlog entries being decided alongside it.
	Batch    []string
	Registry *run.Registry
	// Dir is the fallback workspace for a TODO whose own CWD is not absolute.
	Dir     string
	Resolve RunResolver
	Broker  todos.ApprovalBroker
}

// StartRun returns the item function for a named-step bulk action.
//
// run, plan and triage differ only in the step name — a step declares its own
// prompt, spec and outcomes — so no behaviour is asserted here and the
// lifecycle decides. That is what makes "triage these forty" and "plan these
// forty" one code path rather than three.
func StartRun(spec RunSpec) (ItemFunc, error) {
	step := strings.TrimSpace(spec.Step)
	if step == "" {
		return nil, fmt.Errorf("lifecycle step name is required")
	}
	if spec.Registry == nil {
		return nil, fmt.Errorf("run registry is required")
	}
	resolve := spec.Resolve
	if resolve == nil {
		resolve = DefaultRunResolver
	}

	return func(ctx context.Context, provider todos.Provider, todo *types.TODO) (ItemResult, error) {
		if todo == nil {
			return ItemResult{}, fmt.Errorf("bulk run: no todo")
		}
		workDir := spec.Dir
		if cwd := strings.TrimSpace(todo.CWD); filepath.IsAbs(cwd) {
			workDir = filepath.Clean(cwd)
		}
		opts, err := resolve(ctx, RunRequest{
			Dir: workDir, Step: step, Todo: todo, Batch: spec.Batch, Flags: spec.Flags,
		})
		if err != nil {
			return ItemResult{}, err
		}
		req := run.Request{
			Provider: provider,
			Registry: spec.Registry,
			Todo:     todo,
			Dir:      workDir,
			Options:  opts,
			Broker:   spec.Broker,
		}
		// Resolving first turns a misconfigured run into a per-item error before
		// any agent session is admitted, so one bad TODO does not leave a
		// half-started batch behind it.
		prepared, err := run.Resolve(ctx, req)
		if err != nil {
			return ItemResult{}, err
		}
		req.Prepared = prepared
		started, err := run.Start(req)
		if err != nil {
			return ItemResult{}, err
		}
		return ItemResult{SessionID: started.SessionID, Status: started.Status}, nil
	}, nil
}
