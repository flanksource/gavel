package lifecycle

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// Resolution is one step resolved down to the request captain would be given,
// without dispatching it: the prompt rendered, every spec layer folded, and the
// provenance of the fold.
//
// It exists so `--dry-run`, the dashboard's run preview and the run itself all
// read one answer. A preview that resolved separately from the run it previews
// is a preview of a different run.
type Resolution struct {
	RuntimeProfile *runtimeprofiles.Resolution
	Step           Step
	// Class is the behaviour class the step runs as: what the commit and verify
	// invariants key on.
	Class types.RunMode
	// Prompt is the rendered user prompt. Empty for a verify step, which runs the
	// definition of done rather than an agent turn.
	Prompt string
	// Spec is the validated, dispatchable request.
	Spec api.Spec
	// Timeout is Spec.Budget.Timeout parsed, for callers that need a duration.
	Timeout time.Duration
	// WorkDir is the directory the step runs in: the todo's own cwd when it names
	// one, resolved against the host's.
	WorkDir string
	// Trace is captain's provenance for the fold, lowest precedence first.
	Trace      []api.SpecLayer
	Provenance map[string]api.FieldProvenance
	Warnings   []string

	// lc and prepared are the fold itself, carried so Dispatch runs exactly what
	// Resolve reported. Re-folding at dispatch time would let the run and the
	// preview of it differ whenever the todo changed in between — which is
	// precisely the window a person is looking at the preview.
	lc       Context
	prepared *preparedStep
}

// UseSession pins the agent session id the run will be dispatched with, so a
// caller that must follow the session log can know it before the agent starts.
// It writes through to the request Dispatch actually sends: a caller that set
// it on the reported copy alone would be watching a log nothing writes to.
func (r *Resolution) UseSession(sessionID string) {
	if r == nil || sessionID == "" {
		return
	}
	r.Spec.SessionID = sessionID
	fields := runtimeFields{Name: "lifecycle session", Paths: []string{"/sessionId"}, Replace: true}
	r.Provenance = recordRuntimeFields(r.Provenance, fields)
	if r.prepared != nil {
		r.prepared.request.SessionID = sessionID
		r.prepared.provenance = recordRuntimeFields(r.prepared.provenance, fields)
	}
}

// Resolve prepares a step without running it: the prompt rendered, every spec
// layer folded, nothing dispatched and nothing persisted.
func (h *Host) Resolve(ctx context.Context, todo *types.TODO, step Step, opts RunOptions) (*Resolution, error) {
	lc, prepared, err := h.resolveStep(ctx, todo, step, opts)
	if err != nil {
		return nil, err
	}
	exec := opts.Exec
	if exec == nil {
		exec = todos.NewExecutorContext(ctx, logger.StandardLogger(), nil)
	}
	input := h.runInput(exec, todo, prepared, opts)
	warnings, err := promptrun.Preflight(input.input)
	if err != nil {
		return nil, fmt.Errorf("step %s preflight: %w", step.Name, err)
	}
	return &Resolution{
		RuntimeProfile: prepared.runtimeProfile,
		Step:           step,
		Class:          prepared.class,
		Prompt:         prepared.request.Prompt.User,
		Spec:           prepared.request,
		Timeout:        prepared.timeout,
		WorkDir:        prepared.workDir,
		Trace:          prepared.trace,
		Provenance:     prepared.provenance,
		Warnings:       mergeRuntimeWarnings(prepared.warnings, warnings),
		lc:             lc,
		prepared:       prepared,
	}, nil
}

func mergeRuntimeWarnings(groups ...[]string) []string {
	var warnings []string
	for _, group := range groups {
		for _, warning := range group {
			if !slices.Contains(warnings, warning) {
				warnings = append(warnings, warning)
			}
		}
	}
	return warnings
}
