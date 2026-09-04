package lifecycle

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/api/registry"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

// HostKind is the entrypoint a run was started from. It decides one thing —
// the permission posture the host itself contributes — and it is a layer like
// any other rather than a flag threaded through the executor, so a host cannot
// quietly rewrite a posture the prompt already declared.
type HostKind string

const (
	// HostCLI runs the prompt exactly as its frontmatter declares. The terminal
	// answers nothing: an approval it raised would block until the timeout.
	HostCLI HostKind = "cli"
	// HostDashboard serves the approval endpoints, so it lowers the posture to
	// `default` and attaches the durable broker. Every tool call it cannot
	// pre-approve becomes a question a person can answer.
	HostDashboard HostKind = "dashboard"
)

// DefaultTimeout caps a run's wall-clock duration when neither .gavel.yaml nor
// the request supplies one. It is stamped onto the resolved spec rather than
// applied later, because captain's promptrun refuses a run nobody bounded.
const DefaultTimeout = 30 * time.Minute

// LayerInput is the material one run's spec is folded from. Everything here is
// already loaded and rendered: this type describes the STACK, not how to find
// its pieces, which is what makes the precedence assertable on its own.
type LayerInput struct {
	// Config is the merged .gavel.yaml: its `ai:` base and the `todos.*` section.
	Config verify.GavelConfig
	// Step names the lifecycle step being run and therefore which project block
	// contributes: `todos.run`, `todos.plan`, `todos.triage` and `todos.verify`
	// for the built-in steps, `todos.steps.<name>` for any other.
	Step string
	// Frontmatter are the .prompt documents contributing to this run, lowest
	// precedence first: the embedded default, then a `file:` override.
	Frontmatter []api.SpecLayer
	// StepSpec is the lifecycle step's own declaration — the setup, workflow and
	// permissions todos.yaml gives the step, templates expanded. It outranks the
	// prompt's frontmatter (the definition decides how a step runs, the prompt
	// only what it says) and yields to the project's `todos.<step>` block.
	StepSpec api.Spec
	// Todos supply the per-todo `llm:` layer. A group run folds each in order, so
	// the last todo naming a model wins — one agent session runs them all.
	Todos []*types.TODO
	// Prior are user-scope layers a continuation inherits from the run it
	// continues: the spec that run was dispatched with, and the runtime it
	// actually resolved. They sit below the host and the request, so continuing a
	// run inherits how it ran without outranking where it is being continued from
	// or what the caller now asks for.
	Prior []api.SpecLayer
	// Host is the entrypoint; see HostKind.
	Host HostKind
	// Request is what the caller explicitly asked for: parsed CLI flags or the
	// dashboard payload's spec. A knob the caller did not set must arrive zero or
	// it beats the frontmatter it claims to defer to.
	Request api.Spec
}

// Layers returns the ordered spec layers, lowest precedence first:
//
//	.gavel.yaml ai:  <  todos.timeout  <  prompt frontmatter  <  lifecycle step
//	<  .gavel.yaml todos.<step>  <  the todo's llm:  <  the host  <  the request
//
// todos.timeout is the odd one: it is a context CONSTRAINT rather than a value,
// so it can only ever lower a budget. A project cap that a prompt's own longer
// budget silently overrode was a cap in name only, and one that overrode a
// deliberately short prompt budget was worse.
//
// A layer that configures nothing is omitted rather than appended empty, so the
// trace a caller reports names only the sources that actually spoke.
func Layers(in LayerInput) []api.SpecLayer {
	layers := []api.SpecLayer{{
		Name:   ".gavel.yaml ai",
		Source: api.SpecLayerSourcePreset,
		Scope:  api.SpecLayerGlobal,
		Spec:   in.Config.AI,
	}}
	if timeout := strings.TrimSpace(in.Config.Todos.Timeout); timeout != "" {
		layers = append(layers, api.SpecLayer{
			Name:        ".gavel.yaml todos.timeout",
			Source:      api.SpecLayerSourceProfile,
			Scope:       api.SpecLayerContext,
			Constraints: api.RuntimeConstraints{Limits: api.RunLimits{Budget: api.Budget{Timeout: timeout}}},
		})
	}
	for _, layer := range in.Frontmatter {
		layer.Spec = withoutPromptBody(layer.Spec)
		layers = append(layers, layer)
	}
	if !api.IsEmpty(in.StepSpec) {
		layers = append(layers, api.SpecLayer{
			Name:   "lifecycle step " + in.Step,
			Source: api.SpecLayerSourceProfile,
			Scope:  api.SpecLayerSurface,
			Spec:   withoutPromptBody(in.StepSpec),
		})
	}
	if step := stepSpec(in.Config.Todos, in.Step); !api.IsEmpty(step) {
		layers = append(layers, api.PromptSpecLayer(".gavel.yaml todos."+in.Step, step))
	}
	for _, todo := range in.Todos {
		if layer, ok := todoLayer(todo); ok {
			layers = append(layers, layer)
		}
	}
	for _, prior := range in.Prior {
		if !api.IsEmpty(prior.Spec) {
			layers = append(layers, prior)
		}
	}
	if host, ok := hostLayer(in.Host); ok {
		layers = append(layers, host)
	}
	if !api.IsEmpty(in.Request) {
		layers = append(layers, api.RequestSpecLayer("request", in.Request))
	}
	return layers
}

// ResolveLayers folds Layers through captain's resolver, which is where the
// precedence, the constraint intersection and the trace all come from. Gavel
// keeps no private fold: two implementations of "which layer wins" is one more
// than the number of answers that can be right.
func ResolveLayers(in LayerInput) (api.ResolvedSpec, error) {
	return api.ResolveSpecLayers(Layers(in)...)
}

// PromptLayers renders the frontmatter of each .prompt document that contributes
// to a run — the embedded default, then a `file:` override — and returns them as
// layers alongside the template source Render should use.
//
// Frontmatter is rendered with the same variables Render uses so a template that
// computes its frontmatter resolves; the todo body is empty here because only
// the spec half is being extracted.
func PromptLayers(workDir string, todoList []*types.TODO, definition todoprompt.Definition) ([]api.SpecLayer, string, error) {
	data := todoprompt.TemplateData(todoList, todoprompt.Options{
		WorkDir:  workDir,
		Prompt:   definition.Name,
		Envelope: definition.Envelope,
		Mode:     definition.Class,
	})
	data["body"] = ""

	var layers []api.SpecLayer
	if strings.TrimSpace(definition.Builtin) != "" {
		spec, err := verify.RenderPromptSpec(definition.Builtin, data, verify.PromptSpecOptions{Declared: true})
		if err != nil {
			return nil, "", fmt.Errorf("render built-in %s prompt frontmatter: %w", definition.Name, err)
		}
		layers = append(layers, api.PromptSpecLayer("todos-"+definition.Name+".prompt", spec))
	}
	template, err := definition.Template(workDir)
	if err != nil {
		return nil, "", err
	}
	if definition.Override.File != "" {
		spec, err := verify.RenderPromptSpec(template, data, verify.PromptSpecOptions{Declared: true})
		if err != nil {
			return nil, "", fmt.Errorf("render todos.%s file frontmatter: %w", definition.Name, err)
		}
		layers = append(layers, api.PromptSpecLayer("todos."+definition.Name+" file", spec))
	}
	return layers, template, nil
}

// ApplyModel resolves the canonical compact model grammar once in Captain.
//
// An explicit effort that the expanded model overrides is an error, not a
// silent drop: `--effort high` against a model pinned to `:low` means the user
// asked for two different things and only one can happen.
func ApplyModel(s *api.Spec, requestedEffort api.Effort) error {
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

// ApplyTimeout stamps the run's deadline onto the spec. captain's promptrun has
// no compiled-in ceiling and refuses a run that declares none, so the default
// belongs here, where it is one value the preview and the run both read.
func ApplyTimeout(s *api.Spec) (time.Duration, error) {
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

// ApplyClassInvariants enforces what a behaviour class means regardless of
// configuration. Only run-class steps produce work to commit: a plan writes a
// document for review and a verify run grades what already exists, so a
// `commits:` block inherited from the ai: base or todos config must not turn
// either into a committing run.
func ApplyClassInvariants(s *api.Spec, class types.RunMode) {
	if class == types.ModeRun || s.Workflow == nil {
		return
	}
	s.Workflow.Commits = nil
}

// ValidateSpec checks every section the resolved spec owns EXCEPT the model.
// Spec.Validate is not used because it requires a prompt body, which
// todos/prompt.Render supplies after this point — it validates the complete
// request there.
//
// The model is RequireModel's, separately, because whether a run needs one is a
// property of the step rather than of the spec: a verify step runs the
// definition of done, and a fixture-only definition of done never calls a model.
func ValidateSpec(s api.Spec) error {
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

// RequireModel asserts the spec names a runnable model. Callers that dispatch
// an agent turn call it; a verify-only step calls it only when the todo has
// acceptance criteria for a grader to judge.
func RequireModel(s api.Spec) error {
	if err := s.Model.Validate(); err != nil {
		return fmt.Errorf("model: %w", err)
	}
	return nil
}

// stepSpec is the project block's run configuration for one step: the typed
// `todos.<step>` block for a built-in step, `todos.steps.<name>` for any other.
// Verification has no prompt template — its checklist comes from the todo's
// acceptance criteria — but which model grades a definition of done is a
// configuration decision like any other, so it contributes a spec too; a prompt
// body written there is stripped like every other layer's.
func stepSpec(cfg verify.TodosConfig, step string) api.Spec {
	switch step {
	case "run":
		return withoutPromptBody(cfg.Run.Spec)
	case "plan":
		return withoutPromptBody(cfg.Plan.Spec)
	case "triage":
		return withoutPromptBody(cfg.Triage.Spec)
	case StepVerify:
		return withoutPromptBody(cfg.Verify)
	}
	return withoutPromptBody(cfg.Steps[step])
}

// hostLayer is the entrypoint's own contribution. Only the dashboard has one:
// it can answer a tool approval, so it lowers the posture to `default` and
// brokers what the mode then asks about.
//
// It is emphatically NOT a per-tool policy. `permissions.tools: {Bash: ask}` is
// unenforceable on every runtime captain speaks — RequireToolPolicySupport
// rejects it before the first model call — so a host that wrote one turned an
// approval-gated run into a boundary error.
func hostLayer(host HostKind) (api.SpecLayer, bool) {
	if host != HostDashboard {
		return api.SpecLayer{}, false
	}
	return api.SpecLayer{
		Name:   "host " + string(host),
		Source: api.SpecLayerSourceRequest,
		Scope:  api.SpecLayerUser,
		Spec:   api.Spec{Permissions: api.Permissions{Mode: api.PermissionDefault}},
	}, true
}

// todoLayer projects a todo's `llm:` frontmatter onto a layer. Zero values
// contribute nothing, so a todo naming only a model keeps the config's budget.
func todoLayer(todo *types.TODO) (api.SpecLayer, bool) {
	if todo == nil || todo.LLM == nil {
		return api.SpecLayer{}, false
	}
	var spec api.Spec
	spec.Name = todo.LLM.Model
	if todo.LLM.MaxCost > 0 {
		spec.Budget.Cost = todo.LLM.MaxCost
	}
	if todo.LLM.MaxTurns > 0 {
		spec.Budget.MaxTurns = todo.LLM.MaxTurns
	}
	if api.IsEmpty(spec) {
		return api.SpecLayer{}, false
	}
	return api.SpecLayer{
		Name:   "todo " + todoName(todo),
		Source: api.SpecLayerSourceProfile,
		Scope:  api.SpecLayerSurface,
		Spec:   spec,
	}, true
}

func todoName(todo *types.TODO) string {
	if title := strings.TrimSpace(todo.Title); title != "" {
		return title
	}
	return todo.ID
}

// withoutPromptBody strips the parts of a layer that describe the prompt TEXT
// rather than the run. The body reaches the agent as a rendered template, so
// carrying it here too would inject the raw unrendered source — `{{{body}}}`
// and all — as a body override.
//
// Prompt.System is kept: it is run configuration, not the templated body.
func withoutPromptBody(spec api.Spec) api.Spec {
	spec.Prompt.User = ""
	spec.Prompt.Source = ""
	spec.Prompt.Schema = nil
	spec.Prompt.SchemaJSON = nil
	return spec
}
