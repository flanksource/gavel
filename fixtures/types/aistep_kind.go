package types

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/fixtures"
)

const aiKindName = "ai"

// AIStepFunc grades an acceptance-criteria checklist step.
type AIStepFunc func(fixtures.FixtureTest, fixtures.RunOptions) fixtures.FixtureResult

// AIKind is the `ai` kind. A checklist step (markdown `ai:` front matter, or a
// YAML step with criteria) is graded by its AIStepFunc; a YAML step with only a
// prompt is a plain prompt round-trip whose reply is the CEL root `output`.
// Either way the lazy runtime is resolved first, only when there is work to do.
type AIKind struct {
	grade AIStepFunc
}

// NewAIKind builds the ai kind around grade. The registered kind grades with
// RunAIStep; a test that must not reach a provider registers one around a fake
// with fixtures.DefaultKinds.Replace, keeping the runtime resolution real.
func NewAIKind(grade AIStepFunc) AIKind { return AIKind{grade: grade} }

func (AIKind) Name() string { return aiKindName }

func (AIKind) Schema() json.RawMessage { return aiSchema }

func (k AIKind) Execute(ctx context.Context, step *fixtures.StepContext) (fixtures.StepOutcome, error) {
	test := step.Test
	if test.Step != nil {
		var err error
		if test, err = applyAIConfig(step); err != nil {
			return fixtures.StepOutcome{}, err
		}
	}
	opts := step.Opts
	opts.Context = ctx
	plain := test.Step != nil && len(checklistItems(test)) == 0
	if (plain || len(checklistItems(test)) > 0) && opts.ResolveSpec != nil && opts.Spec == nil && (test.AI == nil || test.AI.Spec == nil) {
		runtime, err := opts.ResolveSpec(ctx)
		if err != nil {
			return fixtures.StepOutcome{}, fmt.Errorf("resolve AI fixture runtime: %w", err)
		}
		opts.Runtime = &runtime
	}
	if plain {
		return runAIPrompt(ctx, test, opts)
	}
	result := k.grade(test, opts)
	outputs := map[string]any{"passed": result.Status == task.StatusPASS}
	if checklist, ok := result.Metadata["checklist"]; ok {
		plain, err := plainValue(checklist)
		if err != nil {
			return fixtures.StepOutcome{Result: result}, err
		}
		outputs["checklist"] = plain
	}
	return fixtures.StepOutcome{Result: result, Outputs: outputs}, nil
}

// aiStepConfig is a YAML/JSON ai step: a model or full runtime spec, and a
// prompt, a criteria checklist, or both (the prompt then instructs the grader).
type aiStepConfig struct {
	Model       string    `json:"model,omitempty"`
	Spec        *api.Spec `json:"spec,omitempty"`
	Prompt      string    `json:"prompt,omitempty"`
	Description string    `json:"description,omitempty"`
	Criteria    []string  `json:"criteria,omitempty"`
	Checklist   []string  `json:"checklist,omitempty"`
}

func applyAIConfig(step *fixtures.StepContext) (fixtures.FixtureTest, error) {
	cfg, err := fixtures.KindConfig[aiStepConfig](step)
	if err != nil {
		return step.Test, err
	}
	var items []fixtures.ChecklistItem
	for _, text := range slices.Concat(cfg.Criteria, cfg.Checklist) {
		items = append(items, fixtures.ChecklistItem{Text: text})
	}
	if len(items) == 0 && strings.TrimSpace(cfg.Prompt) == "" {
		return step.Test, fmt.Errorf("ai step %q declares neither a prompt nor criteria", step.Spec.ID)
	}
	test := step.Test
	test.AI = &fixtures.FixtureAIConfig{Model: cfg.Model, Spec: cfg.Spec}
	test.AIStep = &fixtures.AIStepSpec{Prompt: cfg.Prompt, Description: cfg.Description, Criteria: items}
	return test, nil
}

// runAIPrompt sends a plain prompt and exposes the reply as `output.text`
// (and `output.data` for structured output). Its verdict is the step's `cel:`.
func runAIPrompt(ctx context.Context, test fixtures.FixtureTest, opts fixtures.RunOptions) (fixtures.StepOutcome, error) {
	now := time.Now()
	runtime, config, err := resolveAISpec(test, opts, "fixture.prompt", api.Prompt{User: test.AIStep.Prompt, Source: "fixtures.ai-prompt"})
	if err != nil {
		return fixtures.StepOutcome{}, fmt.Errorf("resolve AI fixture runtime: %w", err)
	}
	provider, err := ai.NewProvider(config)
	if err != nil {
		return fixtures.StepOutcome{}, fmt.Errorf("build ai provider: %w", err)
	}
	defer ai.CloseProvider(provider) //nolint:errcheck
	resp, err := provider.Execute(ctx, runtime.Spec)
	if err != nil {
		return fixtures.StepOutcome{}, fmt.Errorf("ai prompt: %w", err)
	}
	if resp == nil {
		return fixtures.StepOutcome{}, fmt.Errorf("ai prompt returned no response")
	}
	output := map[string]any{"text": resp.Text}
	if resp.StructuredData != nil {
		if output["data"], err = plainValue(resp.StructuredData); err != nil {
			return fixtures.StepOutcome{}, err
		}
	}
	result := fixtures.FixtureResult{Name: test.Name, Type: aiKindName, Status: task.StatusPASS, Start: &now,
		Duration: time.Since(now), Stdout: resp.Text, Test: test}
	return fixtures.StepOutcome{Result: result, Outputs: map[string]any{"output": output}}, nil
}

// plainValue converts a typed value to the maps and slices CEL navigates.
func plainValue(value any) (any, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode %T for CEL: %w", value, err)
	}
	var plain any
	if err := json.Unmarshal(raw, &plain); err != nil {
		return nil, fmt.Errorf("decode %T for CEL: %w", value, err)
	}
	return plain, nil
}

var aiSchema = json.RawMessage(`{
  "type": "object",
  "description": "Send a prompt to an AI agent (its reply is the CEL root output.text), or grade a criteria checklist.",
  "additionalProperties": false,
  "anyOf": [{"required": ["prompt"]}, {"required": ["criteria"]}, {"required": ["checklist"]}],
  "properties": {
    "model": {"type": "string", "description": "Model, e.g. cli:sonnet or api:gpt-5."},
    "spec": {"type": "object", "description": "Complete Captain runtime spec; replaces runner defaults."},
    "prompt": {"type": "string"},
    "description": {"type": "string"},
    "criteria": {"type": "array", "items": {"type": "string"}},
    "checklist": {"type": "array", "items": {"type": "string"}}
  }
}`)
