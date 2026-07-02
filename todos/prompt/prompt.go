// Package prompt owns the built-in todo agent prompts (run, plan, and the
// final-result resume turn) as captain dotprompt templates. Rendering goes
// through captain's engine end-to-end: a template's frontmatter (model,
// permissions.mode, budget, effort) is folded into the returned ai.Request, so
// the .prompt file — not Go code — declares how a mode executes. The per-todo
// sections are assembled in Go and injected as {{{body}}}; the mode's
// structured-output envelope schema is appended OUTSIDE the template so
// .gavel.yaml overrides and the dashboard's editable prompt cannot break the
// output contract.
package prompt

import (
	_ "embed"
	"fmt"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	dotprompt "github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

//go:embed todos-run.prompt
var runTemplate string

//go:embed todos-plan.prompt
var planTemplate string

//go:embed todos-final.prompt
var finalTemplate string

// Options configures Render.
type Options struct {
	WorkDir string
	Mode    types.RunMode // ModeRun or ModePlan (ModeVerify runs through the verify engine)
	Effort  string
	// Template is the resolved .gavel.yaml override source (see ResolveTemplate);
	// empty renders the embedded default for Mode.
	Template string
	// ExistingPlan is the current content of the todo's recorded plan file (plan
	// mode only); empty means the todo has no prior plan.
	ExistingPlan string
	// BodyOverride, when set, replaces the rendered prompt body verbatim — the
	// dashboard's editable prompt. The envelope schema instruction is still
	// appended so an override cannot break the structured-output contract.
	BodyOverride string
}

// ResolveTemplate reads the mode's .gavel.yaml prompt override for dir
// (todos.runPrompt / todos.planPrompt), returning the inline/file template
// source or "" when unset (the embedded default is then used). A
// configured-but-missing file is a hard error.
func ResolveTemplate(dir string, mode types.RunMode) (string, error) {
	cfg, err := verify.LoadGavelConfig(dir)
	if err != nil {
		return "", fmt.Errorf("load .gavel.yaml for todos prompts: %w", err)
	}
	var override verify.PromptOverride
	var key string
	switch mode {
	case types.ModeRun:
		override, key = cfg.Todos.RunPrompt, "todos.runPrompt"
	case types.ModePlan:
		override, key = cfg.Todos.PlanPrompt, "todos.planPrompt"
	default:
		return "", fmt.Errorf("mode %q has no todo prompt template", mode)
	}
	tmpl, err := override.Resolve(dir, "")
	if err != nil {
		return "", fmt.Errorf("resolve %s override: %w", key, err)
	}
	return tmpl, nil
}

// Render renders the mode's prompt for a group of todos and returns the full
// captain request: frontmatter-declared request options folded in by the
// engine, the effort directive leading the body, and the mode's envelope
// schema instruction trailing it. A single todo keeps plain framing; several
// are numbered so the agent can address each in turn.
func Render(todoList []*types.TODO, opts Options) (captainai.Request, captainai.Config, error) {
	if len(todoList) == 0 {
		return captainai.Request{}, captainai.Config{}, fmt.Errorf("no todos supplied")
	}
	template, err := templateSource(opts)
	if err != nil {
		return captainai.Request{}, captainai.Config{}, err
	}

	multiple := len(todoList) > 1
	var body strings.Builder
	for i, todo := range todoList {
		number := 0
		if multiple {
			number = i + 1
		}
		body.WriteString(buildTODOSection(todo, opts.WorkDir, true, number))
	}

	req, cfg, err := dotprompt.Load(template).Render(map[string]any{
		"multiple":     multiple,
		"count":        len(todoList),
		"body":         body.String(),
		"existingPlan": opts.ExistingPlan,
	}, nil)
	if err != nil {
		return captainai.Request{}, captainai.Config{}, fmt.Errorf("render todos %s prompt: %w", opts.Mode, err)
	}

	user := req.Prompt.User
	if opts.BodyOverride != "" {
		user = opts.BodyOverride
	}
	if directive := EffortDirective(opts.Effort); directive != "" {
		user = directive + "\n\n" + user
	}
	schema, err := EnvelopeSchemaJSON(opts.Mode)
	if err != nil {
		return captainai.Request{}, captainai.Config{}, err
	}
	req.Prompt.User = user + "\n\n" + ai.SchemaInstruction(schema)
	req.Prompt.Source = "todos." + string(opts.Mode)
	return req, cfg, nil
}

func templateSource(opts Options) (string, error) {
	if strings.TrimSpace(opts.Template) != "" {
		return opts.Template, nil
	}
	switch opts.Mode {
	case types.ModeRun:
		return runTemplate, nil
	case types.ModePlan:
		return planTemplate, nil
	default:
		return "", fmt.Errorf("mode %q has no todo prompt template (verify runs through the verify engine)", opts.Mode)
	}
}

// EnvelopeSchemaJSON is the JSON schema of the mode's structured final result,
// embedded in every prompt and in the final-result resume turn.
func EnvelopeSchemaJSON(mode types.RunMode) (string, error) {
	var v any
	switch mode {
	case types.ModePlan:
		v = &types.PlanEnvelope{}
	case types.ModeRun:
		v = &types.ResultEnvelope{}
	default:
		return "", fmt.Errorf("mode %q has no result envelope", mode)
	}
	raw, err := api.SchemaJSON(v)
	if err != nil {
		return "", fmt.Errorf("build %s envelope schema: %w", mode, err)
	}
	return string(raw), nil
}

// FinalResultRequest renders the resume turn that asks a finished (or timed
// out) session to emit ONLY the mode's final result JSON.
func FinalResultRequest(mode types.RunMode, sessionID string, timedOut bool) (captainai.Request, error) {
	if sessionID == "" {
		return captainai.Request{}, fmt.Errorf("final-result request needs a session to resume")
	}
	req, _, err := dotprompt.Load(finalTemplate).Render(map[string]any{"timedOut": timedOut}, nil)
	if err != nil {
		return captainai.Request{}, fmt.Errorf("render todos final prompt: %w", err)
	}
	schema, err := EnvelopeSchemaJSON(mode)
	if err != nil {
		return captainai.Request{}, err
	}
	req.Prompt.User += "\n\n" + ai.SchemaInstruction(schema)
	req.Prompt.Source = "todos.final"
	req.SessionID = sessionID
	return req, nil
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
// registry: run, plan, and the todo-verification override of the verify
// template.
func Prompts() []prompts.Prompt {
	return []prompts.Prompt{
		{
			ID:          prompts.TodosRun,
			Title:       "Todo run prompt",
			Description: "The agent prompt for `gavel todos run`: framing, the TODO items, and instructions.",
			ConfigPath:  "todos.runPrompt",
			Default:     runTemplate,
		},
		{
			ID:          prompts.TodosPlan,
			Title:       "Todo plan prompt",
			Description: "The agent prompt for plan-mode runs: read-only investigation that produces a reviewable implementation plan.",
			ConfigPath:  "todos.planPrompt",
			Default:     planTemplate,
		},
		{
			ID:          prompts.TodosVerify,
			Title:       "Todo verify prompt",
			Description: "Overrides the verify template for TODO verification only; `gavel verify` keeps verify.promptTemplate.",
			ConfigPath:  "todos.verifyPrompt",
			Default:     verify.DefaultPromptTemplate(),
		},
	}
}
