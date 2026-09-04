// Package prompt owns gavel's todo prompts as captain dotprompt templates, and
// the Catalog that resolves a prompt NAME to the template, behaviour class, and
// result envelope it runs with.
//
// Name and class are separate axes. Template selection used to be keyed off
// types.RunMode, which made "which prompt runs" and "how the run behaves" the
// same thing: a third prompt required a fourth behaviour class, and a fourth
// class required widening a database CHECK constraint. A prompt now declares its
// class, so any number of prompts can share one — triage is plan-class, like the
// plan prompt, and neither commits nor verifies.
//
// Rendering goes through captain's engine end-to-end: a template's frontmatter
// (model, permissions.mode, budget, effort) is folded into the returned
// ai.Request, so the .prompt file — not Go code — declares how a prompt executes.
// The per-todo sections are assembled in Go and injected as {{{body}}}; the
// prompt's structured-output envelope schema rides on the request as a native
// SchemaJSON field (never in the prompt text) so .gavel.yaml overrides and the
// dashboard's editable prompt cannot break the output contract, and each driver
// delivers it the way its LLM supports.
package prompt

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	dotprompt "github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/todos/types"
)

//go:embed todos-run.prompt
var runTemplate string

//go:embed todos-plan.prompt
var planTemplate string

//go:embed todos-triage.prompt
var triageTemplate string

// Options configures Render.
type Options struct {
	WorkDir string
	// Prompt names the template to render. Empty means the default run prompt.
	// It is independent of Mode: several prompts share one behaviour class.
	Prompt string
	// Envelope is the structured result the prompt returns. Empty derives it from
	// Mode, so a caller that only knows the behaviour class still gets the right
	// schema for the two prompts that predate named selection.
	Envelope EnvelopeKind
	Mode     types.RunMode // ModeRun or ModePlan
	// Spec is the canonical Captain run configuration. Render consumes
	// Prompt.User as the TODO body override and merges every other field over the
	// template request without projecting it through a Gavel-specific adapter.
	Spec api.Spec
	// Template is the resolved .gavel.yaml override source (spec.Resolved.Template);
	// empty renders the embedded default for Mode.
	Template string
	// ExistingPlan is the current content of the todo's recorded plan file (plan
	// mode only); empty means the todo has no prior plan.
	ExistingPlan string
	// Backlog is a compact index of the other open TODOs in the workspace, so a
	// triage run can spot duplicates. Empty omits the section.
	Backlog string
	// Inputs are the lifecycle step's declared template variables, evaluated
	// from the todo. They are folded over the built-in variables above, so a
	// step that declares `existingPlan` decides what the template sees.
	Inputs map[string]any
}

// Default returns the embedded .prompt source for a built-in prompt name. It is
// the lowest template layer: todos/headless renders its frontmatter to seed the run
// spec, and Render falls back to it when no .gavel.yaml override supplies a
// template.
//
// It takes a name rather than a mode because several prompts share a mode; the
// mode no longer identifies a template.
func Default(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		name = DefaultName
	}
	for _, def := range builtins() {
		if def.Name == name {
			return def.Builtin, nil
		}
	}
	return "", fmt.Errorf("prompt %q has no built-in template", name)
}

// TemplateData is the variable set every todo prompt template renders against.
// It is exported because todos/headless renders the same templates' frontmatter with
// the same variables, and a template that computes its frontmatter would resolve
// differently against a divergent set.
func TemplateData(todoList []*types.TODO, opts Options) map[string]any {
	multiple := len(todoList) > 1
	var body strings.Builder
	for i, todo := range todoList {
		number := 0
		if multiple {
			number = i + 1
		}
		body.WriteString(buildTODOSection(todo, opts.WorkDir, true, number, opts.Envelope == EnvelopeTriage))
	}
	data := map[string]any{
		"multiple":     multiple,
		"count":        len(todoList),
		"body":         body.String(),
		"existingPlan": opts.ExistingPlan,
		"backlog":      opts.Backlog,
	}
	for name, value := range opts.Inputs {
		data[name] = value
	}
	return data
}

// promptName is the name Render reports in errors and stamps on the request.
//
// An unnamed prompt falls back to the one matching the behaviour class, not to
// the default: before prompts had names the mode WAS the selector, so a caller
// that supplies only Mode must keep getting the template it always got. Falling
// back to DefaultName instead would silently hand a plan-mode run the run prompt.
func (o Options) promptName() string {
	if name := strings.TrimSpace(o.Prompt); name != "" {
		return name
	}
	if o.Mode != "" {
		return string(o.Mode)
	}
	return DefaultName
}

// Render renders the named prompt for a group of todos and returns the full
// captain request: frontmatter-declared request options folded in by the
// engine, the effort directive leading the body, and the prompt's envelope
// schema riding as a native field. A single todo keeps plain framing; several
// are numbered so the agent can address each in turn.
func Render(todoList []*types.TODO, opts Options) (captainai.Request, captainai.Config, error) {
	if len(todoList) == 0 {
		return captainai.Request{}, captainai.Config{}, fmt.Errorf("no todos supplied")
	}
	template, err := templateSource(opts)
	if err != nil {
		return captainai.Request{}, captainai.Config{}, err
	}

	req, cfg, err := dotprompt.Load(template).Render(TemplateData(todoList, opts), nil)
	if err != nil {
		return captainai.Request{}, captainai.Config{}, fmt.Errorf("render todos %s prompt: %w", opts.promptName(), err)
	}

	user := req.Prompt.User
	if opts.Spec.Prompt.User != "" {
		user = opts.Spec.Prompt.User
	}
	renderedPrompt := req.Prompt
	override := opts.Spec
	override.Prompt.User = ""
	override.Prompt.Source = ""
	override.Prompt.Schema = nil
	override.Prompt.SchemaJSON = nil
	req = req.Merge(override)
	// Resolve last, after the merge: the template's model is resolved when it is
	// rendered, and an override that names a different one leaves the request
	// carrying a name from the caller with nothing else filled in. Callers are
	// handed a driver-ready model, which is the contract the executor relies on.
	if strings.TrimSpace(req.Name) != "" {
		resolved, err := captainai.Resolve(req.Model)
		if err != nil {
			return captainai.Request{}, captainai.Config{}, fmt.Errorf("resolve todos %s runtime: %w", opts.promptName(), err)
		}
		req.Model = resolved
	}

	effort := string(req.Effort)
	if directive := EffortDirective(effort); directive != "" {
		user = directive + "\n\n" + user
	}
	schema, err := EnvelopeSchemaJSON(opts.envelope())
	if err != nil {
		return captainai.Request{}, captainai.Config{}, err
	}
	// The envelope schema rides as a native field, never appended to the prompt
	// text: the driver delivers it the way its LLM supports, and a hostile
	// template/body override cannot touch a separate field. It is set identically
	// on every turn of the run session (see the executor), which is what the
	// claude-agent per-turn byte-equality guard requires.
	req.Prompt.User = user
	req.Prompt.Schema = renderedPrompt.Schema
	req.Prompt.SchemaJSON = schema
	req.Prompt.Source = "todos." + opts.promptName()
	if err := req.Validate(); err != nil {
		return captainai.Request{}, captainai.Config{}, fmt.Errorf("validate todos %s spec: %w", opts.promptName(), err)
	}
	return req, cfg, nil
}

func templateSource(opts Options) (string, error) {
	if strings.TrimSpace(opts.Template) != "" {
		return opts.Template, nil
	}
	return Default(opts.promptName())
}

// envelope resolves which structured result this run expects, falling back to
// the behaviour class for callers that predate named prompts.
func (o Options) envelope() EnvelopeKind {
	if o.Envelope != "" {
		return o.Envelope
	}
	return envelopeFor(o.Mode)
}

// EnvelopeSchemaJSON is the JSON schema of a prompt's structured final result.
// The executor computes it once per run session and threads the same bytes to
// every turn (initial, retry, and feedback) so the claude-agent
// provider's per-turn byte-equality guard holds by identity.
func EnvelopeSchemaJSON(kind EnvelopeKind) (json.RawMessage, error) {
	var v any
	switch kind {
	case EnvelopePlan:
		v = &types.PlanEnvelope{}
	case EnvelopeResult:
		v = &types.ResultEnvelope{}
	case EnvelopeTriage:
		v = &types.TriageEnvelope{}
	default:
		return nil, fmt.Errorf("envelope %q is not one of result, plan, triage", kind)
	}
	raw, err := api.SchemaJSON(v)
	if err != nil {
		return nil, fmt.Errorf("build %s envelope schema: %w", kind, err)
	}
	return raw, nil
}

// EffortDirective renders the leading reasoning-effort instruction for a run,
// or "" for an unknown tier.
func EffortDirective(effort string) string {
	switch effort {
	case "low":
		return "Be concise."
	case "medium", "":
		return "Think carefully before implementing."
	case "high":
		return "Think hard and reason thoroughly; consider edge cases before implementing."
	case "xhigh":
		return "Think very hard, reason exhaustively, and validate edge cases before implementing."
	default:
		return ""
	}
}

// Prompts returns the overridable todo prompt descriptors for the settings
// registry. It is derived from the built-in catalog rather than restated, so a
// new built-in prompt appears in the settings UI without a second edit.
// builtinUsedBy names the commands that run each built-in todo prompt.
var builtinUsedBy = map[string][]string{
	"run":    {"gavel todos run"},
	"plan":   {"gavel todos run --mode plan", "gavel todos plan"},
	"triage": {"gavel todos run --prompt triage"},
}

func Prompts() []prompts.Prompt {
	defs := builtins()
	registry := make([]prompts.Prompt, 0, len(defs))
	for _, def := range defs {
		registry = append(registry, prompts.Prompt{
			ID:          "todos." + def.Name,
			Title:       def.Title,
			Description: def.Description,
			ConfigPath:  "todos." + def.Name,
			Default:     def.Builtin,
			UsedBy:      builtinUsedBy[def.Name],
		})
	}
	return registry
}
