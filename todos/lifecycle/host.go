package lifecycle

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

// Host adapts a todo to the lifecycle: it projects the todo onto the subject
// the predicates read, derives the identity a step run is admitted under, folds
// the step's spec layers, runs the step through captain, and — OnOutcome — is
// the one place a todo's status is written from a run.
type Host struct {
	Provider todos.Provider
	Def      *Engine
	Config   verify.GavelConfig
	WorkDir  string
	Kind     HostKind
}

// NewHost loads the project's configuration and lifecycle for a work dir.
func NewHost(provider todos.Provider, workDir string, kind HostKind) (*Host, error) {
	cfg, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return nil, fmt.Errorf("load .gavel.yaml: %w", err)
	}
	def, err := LoadWith(cfg.Todos.Lifecycle, workDir)
	if err != nil {
		return nil, err
	}
	engine, err := New(def)
	if err != nil {
		return nil, err
	}
	return &Host{Provider: provider, Def: engine, Config: cfg, WorkDir: workDir, Kind: kind}, nil
}

// Subject projects the todo onto the variables the lifecycle declares. Every
// declared field is present so a predicate never reads a missing key; the
// engine checks the projection against the declarations before evaluating.
func (h *Host) Subject(ctx context.Context, todo *types.TODO) (map[string]any, error) {
	if todo == nil {
		return nil, fmt.Errorf("lifecycle subject: todo is nil")
	}
	plan, err := h.planState(ctx, todo)
	if err != nil {
		return nil, err
	}
	dod, err := h.VerifyDocument(todo)
	if err != nil {
		return nil, err
	}
	labels := todo.Labels
	if labels == nil {
		labels = []string{}
	}
	return map[string]any{
		"id":       todo.ID,
		"status":   string(todo.Status),
		"priority": string(todo.Priority),
		"labels":   labels,
		"body":     todo.MarkdownBody,
		"attempts": todo.Attempts,
		"execution": map[string]any{
			"state": todo.ExecutionState,
		},
		"plan": map[string]any{
			"exists":   plan.Exists,
			"approved": plan.Approved,
			"content":  plan.Content,
			"path":     plan.Path,
			"revision": plan.Revision,
		},
		"verification": map[string]any{
			"exists":   dod.Declared(),
			"document": dod.Fixture,
		},
	}, nil
}

func (h *Host) planState(ctx context.Context, todo *types.TODO) (todos.PlanState, error) {
	if h.Provider == nil {
		return todos.PlanState{}, fmt.Errorf("lifecycle subject: host has no provider")
	}
	plans, ok := h.Provider.(todos.PlanStateProvider)
	if !ok {
		return todos.PlanState{}, fmt.Errorf("lifecycle subject: provider %T does not expose plan state", h.Provider)
	}
	state, err := plans.PlanState(ctx, todo)
	if err != nil {
		return todos.PlanState{}, fmt.Errorf("lifecycle subject: plan state: %w", err)
	}
	return state, nil
}

// Runs is the todo's run history as the predicates see it: every recorded step
// run, oldest first, under the step name it was dispatched as — a custom step
// included — so `size(runs)` counts them and `last.<step>` is the newest of
// each. The provider must expose the history: a lifecycle evaluated against
// runs it cannot see would re-verify work it never saw land.
func (h *Host) Runs(ctx context.Context, todo *types.TODO) ([]StepRun, error) {
	if h.Provider == nil {
		return nil, fmt.Errorf("lifecycle runs: host has no provider")
	}
	history, ok := h.Provider.(todos.RunHistoryProvider)
	if !ok {
		return nil, fmt.Errorf("lifecycle runs: provider %T does not expose run history", h.Provider)
	}
	records, err := history.RunHistory(ctx, todo)
	if err != nil {
		return nil, fmt.Errorf("lifecycle runs: %w", err)
	}
	runs := make([]StepRun, 0, len(records))
	for _, record := range records {
		runs = append(runs, StepRun{
			Step: record.Step, State: record.State, Outcome: record.Outcome, PromptRunID: record.PromptRunID,
			StartedAt: record.StartedAt, FinishedAt: record.FinishedAt,
		})
	}
	return runs, nil
}

// Context is the todo as the engine evaluates it.
func (h *Host) Context(ctx context.Context, todo *types.TODO) (Context, error) {
	subject, err := h.Subject(ctx, todo)
	if err != nil {
		return Context{}, err
	}
	runs, err := h.Runs(ctx, todo)
	if err != nil {
		return Context{}, err
	}
	return Context{Subject: subject, Runs: runs}, nil
}

// Next is the step the lifecycle would run for this todo now.
func (h *Host) Next(ctx context.Context, todo *types.TODO) (Step, bool, error) {
	lc, err := h.Context(ctx, todo)
	if err != nil {
		return Step{}, false, err
	}
	return h.Def.Next(lc)
}

// VerifyDocument renders the todo's own definition of done: its `## Verification`
// fixture, its acceptance criteria, and the checks the project enables — the
// document a step's `workflow.verify.fixture` placeholder expands to. The grader
// it declares for acceptance criteria is resolved from the verify chain —
// `.gavel.yaml todos.verify` over `ai:` — never from the step being run, so an
// implementer's model and session never mark their own work.
//
// No run spec is consulted: the document describes the todo, and which step
// runs it is the lifecycle's decision. Forcing the configured checks on for one
// run is therefore not a request-level toggle here; `.gavel.yaml checks.enabled`
// or the todo's own `checks:` front matter is what enables them.
func (h *Host) VerifyDocument(todo *types.TODO) (todos.DefinitionOfDone, error) {
	grader, err := h.graderSpec(todo)
	if err != nil {
		return todos.DefinitionOfDone{}, err
	}
	return todos.BuildDefinitionOfDone(todos.DefinitionOfDoneOptions{
		WorkDir: h.WorkDir,
		Todos:   []*types.TODO{todo},
		Grader:  grader,
	})
}

// graderSpec is the verification chain resolved on its own, the way `gavel
// todos check` resolves it: no prompt frontmatter and no step declaration, only
// the project's configuration and the host.
//
// The chain has to name a runnable model only when there is something for it to
// grade: the document declares its `ai:` block for acceptance criteria alone, so
// a project that configures no model still verifies todos whose definition of
// done is fixture steps.
func (h *Host) graderSpec(todo *types.TODO) (api.Spec, error) {
	resolved, err := ResolveLayers(LayerInput{
		Config: h.Config,
		Step:   StepVerify,
		Todos:  []*types.TODO{todo},
		Host:   h.Kind,
	})
	if err != nil {
		return api.Spec{}, fmt.Errorf("resolve verification spec: %w", err)
	}
	spec := resolved.Spec
	if len(todo.AcceptanceCriteria) == 0 {
		return spec, nil
	}
	if err := ApplyModel(&spec, ""); err != nil {
		return api.Spec{}, err
	}
	if _, err := ApplyTimeout(&spec); err != nil {
		return api.Spec{}, err
	}
	ApplyClassInvariants(&spec, types.ModeVerify)
	if err := RequireModel(spec); err != nil {
		return api.Spec{}, fmt.Errorf("verification spec for the acceptance-criteria grader: %w", err)
	}
	if err := ValidateSpec(spec); err != nil {
		return api.Spec{}, fmt.Errorf("verification spec for the acceptance-criteria grader: %w", err)
	}
	return spec, nil
}

// classOf is the behaviour class a step runs as — what captain's mode axis and
// the run's commit/verify invariants key on. The verify step grades without an
// agent turn; a plan or triage envelope is a read-only pass; everything else
// implements.
func classOf(step Step) types.RunMode {
	if step.Name == StepVerify || step.Prompt == promptRefPrefix+StepVerify {
		return types.ModeVerify
	}
	switch step.EnvelopeOrDefault() {
	case EnvelopePlan, EnvelopeTriage:
		return types.ModePlan
	}
	return types.ModeRun
}

const (
	promptRefPrefix  = "todos."
	promptFilePrefix = "file:"
)

// promptFor resolves a step's prompt reference: `todos.<name>` names a built-in
// prompt (overridable through `.gavel.yaml todos.<name>`), `file:<path>` a
// project-owned template rendered with the step's envelope.
//
// A verify-class step has no prompt to resolve — it runs the definition of
// done — and Lifecycle.Validate has already refused a `verify` step that
// references anything but todos.verify.
func (h *Host) promptFor(step Step) (todoprompt.Definition, error) {
	ref := strings.TrimSpace(step.Prompt)
	envelope := todoprompt.EnvelopeKind(step.EnvelopeOrDefault())
	switch {
	case classOf(step) == types.ModeVerify:
		return todoprompt.Definition{Name: StepVerify, Class: types.ModeVerify}, nil
	case strings.HasPrefix(ref, promptRefPrefix):
		catalog, err := todoprompt.NewCatalog(h.Config.Todos)
		if err != nil {
			return todoprompt.Definition{}, err
		}
		definition, err := catalog.Lookup(strings.TrimPrefix(ref, promptRefPrefix))
		if err != nil {
			return todoprompt.Definition{}, fmt.Errorf("step %s: %w", step.Name, err)
		}
		if step.Envelope != "" && definition.Envelope != envelope {
			return todoprompt.Definition{}, fmt.Errorf("step %s: prompt %s returns a %s envelope, the step declares %s",
				step.Name, ref, definition.Envelope, envelope)
		}
		return definition, nil
	case strings.HasPrefix(ref, promptFilePrefix):
		return todoprompt.Definition{
			Name:     step.Name,
			Class:    classOf(step),
			Envelope: envelope,
			Title:    step.Name,
			Override: verify.PromptSpec{File: strings.TrimPrefix(ref, promptFilePrefix)},
			Origin:   ref,
		}, nil
	}
	return todoprompt.Definition{}, fmt.Errorf("step %s: prompt %q is neither todos.<name> nor file:<path>", step.Name, step.Prompt)
}

// stepWorkDir is the directory a todo's step runs in: the todo's own cwd when it
// names one, resolved against the host's work dir.
func (h *Host) stepWorkDir(todo *types.TODO) string {
	if cwd := strings.TrimSpace(todo.CWD); cwd != "" {
		if filepath.IsAbs(cwd) {
			return filepath.Clean(cwd)
		}
		if h.WorkDir != "" {
			return filepath.Clean(filepath.Join(h.WorkDir, cwd))
		}
		return filepath.Clean(cwd)
	}
	if h.WorkDir != "" {
		return filepath.Clean(h.WorkDir)
	}
	return "."
}
