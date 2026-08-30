// Package spec resolves the single run configuration a TODO operation executes
// with. Every entrypoint — `gavel todos run`, `todos plan revise`,
// `todos check`, and the dashboard's run/approve/revise handlers — calls
// Resolve, so a `.gavel.yaml` setting means the same thing regardless of which
// binary read it.
//
// Resolution is a fold over an ordered list of Layers, lowest precedence first:
//
//	.gavel.yaml ai:  <  <mode>.prompt frontmatter  <  todos.<mode> file frontmatter
//	                 <  .gavel.yaml todos.<mode>   <  per-todo frontmatter
//	                 <  the request (CLI flags or dashboard payload)
//
// Expressing the layering as data rather than code is what makes provenance
// fall out of the fold: the accumulator is diffed after each layer, so the
// answer to "where did this value come from" is derived, not maintained
// alongside.
//
// The boundary with todos/prompt: this package resolves the run CONFIGURATION,
// todos/prompt renders the prompt TEXT. Resolved.Template carries the override
// template source across that boundary; the prompt body itself is deliberately
// stripped from every configuration layer so a `.gavel.yaml` inline prompt is
// used as a template exactly once, rather than also being injected raw as an
// unrendered body override.
package spec

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/api/registry"
	"github.com/flanksource/gavel/todos/drivers"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

// DefaultTimeout caps a run's wall-clock duration when neither .gavel.yaml nor
// the request supplies one.
const DefaultTimeout = 30 * time.Minute

// Layer is one contribution to the resolved spec, named for provenance.
type Layer struct {
	Name string
	Spec api.Spec
}

// Input is everything an entrypoint knows before resolution.
type Input struct {
	// WorkDir is the discovery root the .gavel.yaml layers are loaded from. It is
	// NOT joined with a todo's CWD — the executor owns that join.
	WorkDir string
	// Mode is the behaviour class: run, plan, or verify. ModeVerify has no prompt
	// template and therefore no per-prompt PromptSpec layer.
	//
	// When Prompt names a prompt, that prompt's declared class wins and Mode is
	// only checked for agreement — supplying both and disagreeing is an error, not
	// a silent preference for one.
	Mode types.RunMode
	// Prompt names the template to run: run, plan, triage, or a name declared in
	// .gavel.yaml todos.prompts. Empty defaults to the prompt matching Mode.
	Prompt string
	// Todos supplies the per-todo `llm:` frontmatter layer. A group run folds each
	// todo in order, so the last todo's explicit model wins — the same single
	// agent session runs them all.
	Todos []*types.TODO
	// Override is the request: parsed CLI flags or the dashboard payload's spec.
	// It is the highest layer, so a flag the user did not set must arrive zero
	// (gate on cmd.Flags().Changed) or it will beat the .prompt frontmatter.
	Override api.Spec
	// Driver selects the execution mechanism; empty resolves from .gavel.yaml and
	// then drivers.Default.
	Driver string
	// CanApprove reports whether this entrypoint can answer a tool-approval
	// request. The CLI cannot: its approval registry is never drained, so a run
	// that asks would block forever.
	CanApprove bool
}

// Resolved is the single answer every entrypoint runs with.
type Resolved struct {
	// Spec is the executable run configuration, validated except for the prompt
	// body — which todos/prompt.Render supplies and validates.
	Spec api.Spec
	// Mode is the resolved behaviour class, so callers stop re-deriving it.
	Mode types.RunMode
	// Prompt is the resolved prompt name.
	Prompt string
	// Envelope is the structured result the resolved prompt returns.
	Envelope todoprompt.EnvelopeKind
	// Template is the template source to render — an override's when it supplies
	// one, otherwise the embedded built-in.
	Template string
	Driver   drivers.Kind
	GroupBy  string
	// Approvals gates Bash behind human approval (see TodosConfig.Approvals).
	Approvals bool
	// Timeout is Spec.Budget.Timeout parsed, for callers that need a duration.
	Timeout time.Duration
	// Layers is the fold's input, lowest precedence first.
	Layers []Layer
	// Provenance maps a dotted spec path ("model.name", "budget.cost") to the
	// name of the layer that last set it.
	Provenance map[string]string
}

// Resolve folds every configuration layer into one spec. It fails loud on a
// malformed model, timeout, effort conflict, or permission value rather than
// dropping the offending value and running with a silent default.
func Resolve(in Input) (Resolved, error) {
	cfg, err := verify.LoadGavelConfig(in.WorkDir)
	if err != nil {
		return Resolved{}, fmt.Errorf("load .gavel.yaml: %w", err)
	}
	definition, err := resolvePrompt(&in, cfg.Todos)
	if err != nil {
		return Resolved{}, err
	}

	byMode, template, err := modeLayers(in, cfg, definition)
	if err != nil {
		return Resolved{}, err
	}
	layers := append([]Layer{{Name: ".gavel.yaml ai", Spec: cfg.AI}}, byMode...)

	// todos.timeout sits above the operation spec so it stays the one place a
	// project caps every todo run, whichever mode it is in.
	if t := strings.TrimSpace(cfg.Todos.Timeout); t != "" {
		layers = append(layers, Layer{
			Name: ".gavel.yaml todos.timeout",
			Spec: api.Spec{Budget: api.Budget{Timeout: t}},
		})
	}
	for _, todo := range in.Todos {
		if layer, ok := todoLayer(todo); ok {
			layers = append(layers, layer)
		}
	}
	layers = append(layers, Layer{Name: "request", Spec: in.Override})

	folded, provenance := fold(layers)
	driver, err := resolveDriver(in.Driver, cfg.Todos.Driver)
	if err != nil {
		return Resolved{}, err
	}
	if folded.Mode == "" {
		mode, _ := driver.RuntimeMode()
		folded.Mode = mode
	}

	if err := applyModel(&folded, in.Override.Effort); err != nil {
		return Resolved{}, err
	}
	timeout, err := applyTimeout(&folded)
	if err != nil {
		return Resolved{}, err
	}
	applyModeInvariants(&folded, in.Mode)
	if err := validate(folded); err != nil {
		return Resolved{}, err
	}

	driver, err = drivers.Parse(string(folded.Mode))
	if err != nil {
		return Resolved{}, fmt.Errorf("resolved model backend: %w", err)
	}
	approvals, err := resolveApprovals(cfg.Todos.Approvals, in.CanApprove)
	if err != nil {
		return Resolved{}, err
	}

	return Resolved{
		Spec:       folded,
		Mode:       in.Mode,
		Prompt:     definition.Name,
		Envelope:   definition.Envelope,
		Template:   template,
		Driver:     driver,
		GroupBy:    cfg.Todos.GroupBy,
		Approvals:  approvals,
		Timeout:    timeout,
		Layers:     layers,
		Provenance: provenance,
	}, nil
}

// applyModeInvariants enforces what a mode means regardless of configuration.
// Only run mode produces work to commit: a plan writes a document for review and
// a verify run grades what already exists, so a `commits:` block inherited from
// the ai: base or todos config must not turn either into a committing run.
func applyModeInvariants(s *api.Spec, mode types.RunMode) {
	if mode == types.ModeRun || s.Workflow == nil {
		return
	}
	s.Workflow.Commits = nil
}

// resolvePrompt picks the prompt this run executes and reconciles it with the
// requested behaviour class, mutating in.Mode to the class that actually applies.
//
// The two inputs can disagree, and the disagreement is always the caller's error
// rather than a preference to resolve silently: `--prompt triage --mode run`
// means the caller asked for a read-only pass AND for a committing one.
// Verification is the exception — it runs no prompt at all.
func resolvePrompt(in *Input, cfg verify.TodosConfig) (todoprompt.Definition, error) {
	if in.Mode == types.ModeVerify {
		if strings.TrimSpace(in.Prompt) != "" {
			return todoprompt.Definition{}, fmt.Errorf(
				"prompt %q cannot be used with verify, which runs the fixture rather than an agent prompt", in.Prompt)
		}
		return todoprompt.Definition{Name: string(types.ModeVerify), Class: types.ModeVerify}, nil
	}

	catalog, err := todoprompt.NewCatalog(cfg)
	if err != nil {
		return todoprompt.Definition{}, err
	}
	name := strings.TrimSpace(in.Prompt)
	if name == "" {
		// No prompt named: the mode still selects one, which is what keeps every
		// pre-existing `--mode plan` caller working unchanged.
		name = string(in.Mode)
		if in.Mode == "" {
			name = todoprompt.DefaultName
		}
	}
	definition, err := catalog.Lookup(name)
	if err != nil {
		return todoprompt.Definition{}, err
	}
	if in.Mode != "" && in.Mode != definition.Class {
		return todoprompt.Definition{}, fmt.Errorf(
			"prompt %q runs as %s, but mode %s was requested; drop one",
			definition.Name, definition.Class, in.Mode)
	}
	in.Mode = definition.Class
	in.Prompt = definition.Name
	return definition, nil
}

// modeLayers returns the configuration layers the prompt itself contributes,
// plus the template source Render should use.
//
// Run, plan, and triage are prompt-driven: the prompt's .prompt frontmatter, an
// optional `file:` override's frontmatter, and the .gavel.yaml PromptSpec.
// Verification has no prompt — its checklist is generated from the todo's
// acceptance criteria — but it is still a run, and which model grades a
// definition of done is a configuration decision like any other, so todos.verify
// contributes a spec.
func modeLayers(in Input, cfg verify.GavelConfig, definition todoprompt.Definition) ([]Layer, string, error) {
	if in.Mode == types.ModeVerify {
		return []Layer{{Name: ".gavel.yaml todos.verify", Spec: cfg.Todos.Verify}}, "", nil
	}
	return templateLayers(in, definition)
}

// templateLayers renders the frontmatter of each template that contributes to
// the run — the embedded default, then a `file:` override — and returns them as
// layers alongside the template source Render should use.
//
// Frontmatter is rendered with the same variables Render uses so a template that
// computes its frontmatter resolves; the todo body is empty here because only
// the spec half is being extracted.
func templateLayers(in Input, definition todoprompt.Definition) ([]Layer, string, error) {
	data := todoprompt.TemplateData(in.Todos, todoprompt.Options{
		WorkDir:  in.WorkDir,
		Prompt:   definition.Name,
		Envelope: definition.Envelope,
		Mode:     definition.Class,
	})
	data["body"] = ""

	key := "todos." + definition.Name
	var layers []Layer
	if strings.TrimSpace(definition.Builtin) != "" {
		builtinSpec, err := verify.RenderPromptSpec(definition.Builtin, data)
		if err != nil {
			return nil, "", fmt.Errorf("render built-in %s prompt frontmatter: %w", definition.Name, err)
		}
		layers = append(layers, Layer{Name: "todos-" + definition.Name + ".prompt", Spec: withoutPromptBody(builtinSpec)})
	}

	template, err := definition.Template(in.WorkDir)
	if err != nil {
		return nil, "", err
	}
	if definition.Override.File != "" {
		fileSpec, err := verify.RenderPromptSpec(template, data)
		if err != nil {
			return nil, "", fmt.Errorf("render %s file frontmatter: %w", key, err)
		}
		layers = append(layers, Layer{Name: key + " file", Spec: withoutPromptBody(fileSpec)})
	}
	layers = append(layers, Layer{Name: ".gavel.yaml " + key, Spec: withoutPromptBody(definition.Override.Spec)})
	return layers, template, nil
}

// withoutPromptBody strips the parts of a configuration layer that describe the
// prompt TEXT rather than the run. The body reaches the agent as a rendered
// template (Resolved.Template), so carrying it here too would inject the raw
// unrendered source — `{{{body}}}` and all — as a body override.
//
// Prompt.System is kept: it is run configuration, not the templated body.
func withoutPromptBody(s api.Spec) api.Spec {
	s.Prompt.User = ""
	s.Prompt.Source = ""
	s.Prompt.Schema = nil
	s.Prompt.SchemaJSON = nil
	return s
}

// todoLayer projects a todo's `llm:` frontmatter onto a spec layer. Zero values
// contribute nothing, so a todo naming only a model keeps the config's budget.
func todoLayer(todo *types.TODO) (Layer, bool) {
	if todo == nil || todo.LLM == nil {
		return Layer{}, false
	}
	var s api.Spec
	s.Name = todo.LLM.Model
	if todo.LLM.MaxCost > 0 {
		s.Budget.Cost = todo.LLM.MaxCost
	}
	if todo.LLM.MaxTurns > 0 {
		s.Budget.MaxTurns = todo.LLM.MaxTurns
	}
	if api.IsEmpty(s) {
		return Layer{}, false
	}
	return Layer{Name: "todo " + todoName(todo), Spec: s}, true
}

func todoName(todo *types.TODO) string {
	if strings.TrimSpace(todo.Title) != "" {
		return todo.Title
	}
	return todo.ID
}

// applyModel resolves the canonical compact model grammar once in Captain. The
// resolved adapter remains internal; the portable backend stays api, agent,
// cli, or cmux.
//
// An explicit effort that the expanded model overrides is an error, not a
// silent drop: `--effort high` against a model pinned to `:low` means the user
// asked for two different things and only one can happen.
func applyModel(s *api.Spec, requestedEffort api.Effort) error {
	expanded, err := registry.ResolveModel(s.Model)
	if err != nil {
		return fmt.Errorf("invalid todos model %q: %w", s.Name, err)
	}
	if requestedEffort != "" && expanded.Effort != requestedEffort {
		return fmt.Errorf("effort %q conflicts with model %q, which pins effort %q; drop one",
			requestedEffort, s.Name, expanded.Effort)
	}
	if expanded.Effort == "" {
		expanded.Effort = api.EffortMedium
	}
	s.Model = expanded
	return nil
}

func applyTimeout(s *api.Spec) (time.Duration, error) {
	timeout := DefaultTimeout
	if raw := strings.TrimSpace(s.Budget.Timeout); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil {
			return 0, fmt.Errorf("invalid todos timeout %q: %w", raw, err)
		}
		if parsed <= 0 {
			return 0, fmt.Errorf("todos timeout %q must be greater than zero", raw)
		}
		timeout = parsed
	}
	s.Budget.Timeout = timeout.String()
	return timeout, nil
}

// validate checks every section the resolved spec owns. Spec.Validate is not
// used because it requires a prompt body, which todos/prompt.Render supplies
// after this point — it validates the complete request there.
func validate(s api.Spec) error {
	if err := s.Model.Validate(); err != nil {
		return fmt.Errorf("model: %w", err)
	}
	if err := s.Budget.Validate(); err != nil {
		return fmt.Errorf("budget: %w", err)
	}
	if err := s.Permissions.Validate(); err != nil {
		return fmt.Errorf("permissions: %w", err)
	}
	if err := s.Workflow.Validate(); err != nil {
		return fmt.Errorf("workflow: %w", err)
	}
	return nil
}

// resolveDriver selects a default only when the prompt itself did not select a
// backend. Captain's compact model prefix then outranks both the sibling backend
// field and this Gavel default during ResolveModel.
func resolveDriver(request, configured string) (drivers.Kind, error) {
	if s := strings.TrimSpace(request); s != "" {
		return drivers.Parse(s)
	}
	if s := strings.TrimSpace(configured); s != "" {
		return drivers.Parse(s)
	}
	return drivers.Default, nil
}

// resolveApprovals decides the Bash-approval posture. It defaults to whatever
// the entrypoint can actually service, so the same config is safe everywhere,
// and refuses to enable approvals where nothing can answer them — a run that
// asked would block until its timeout with no indication why.
func resolveApprovals(configured *bool, canApprove bool) (bool, error) {
	approvals := canApprove
	if configured != nil {
		approvals = *configured
	}
	if approvals && !canApprove {
		return false, fmt.Errorf("todos.approvals is enabled but this entrypoint cannot answer approval requests; " +
			"a run would block forever waiting for one (set todos.approvals: false, or start the run from the dashboard)")
	}
	return approvals, nil
}
