// Package form registers the `form` fixture step kind: a JSON-schema form a
// person fills while the run waits — the run's parameters (a fixture's
// `params:` lowers to a form step with id `params`) or an approval gate that
// later steps read through `if:`.
package form

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/flanksource/clicky/prompt"
	"github.com/flanksource/clicky/task"

	"github.com/flanksource/gavel/fixtures"
)

func init() { fixtures.RegisterKind(formKind{}) }

// AnsweredBySupplied is answeredBy when supplied params satisfied the form.
const AnsweredBySupplied = "supplied"

type formKind struct{}

type config struct {
	Title       string          `json:"title,omitempty"`
	Description string          `json:"description,omitempty"`
	Schema      json.RawMessage `json:"schema"`
	Default     map[string]any  `json:"default,omitempty"`
}

const configSchema = `{
  "type": "object", "additionalProperties": false, "required": ["schema"],
  "properties": {
    "title": {"type": "string", "description": "Form title (templated)."},
    "description": {"type": "string", "description": "Form description (templated)."},
    "schema": {"type": "object", "description": "JSON Schema (type: object) of the answer; format: connection picks a saved connection; like any step config it is templated, and {$cel: <expr>} computes a typed value, e.g. enum: {$cel: \"databases.rows.map(r, r.name)\"} offers an earlier query's rows."},
    "default": {"type": "object", "description": "Prefilled values (templated)."}
  }
}`

func (formKind) Name() string            { return fixtures.FormKind }
func (formKind) Schema() json.RawMessage { return json.RawMessage(configSchema) }

func (formKind) Execute(ctx context.Context, step *fixtures.StepContext) (fixtures.StepOutcome, error) {
	start := time.Now()
	cfg, err := fixtures.KindConfig[config](step)
	if err != nil {
		return failed(step, start, err), nil
	}
	f, err := newForm(step, cfg)
	if err != nil {
		return failed(step, start, err), nil
	}
	if step.Spec.ID == fixtures.ParamsStepID && len(f.shape.missing(f.values)) == 0 && prompt.Validate(f.schema, f.values) == nil {
		return f.outcome(start, prompt.Answer{Values: f.values, At: time.Now(), By: AnsweredBySupplied}), nil
	}
	if step.Opts.Prompter == nil {
		return f.unanswered(failed(step, start, f.noPrompter())), nil
	}
	answer, err := f.ask(ctx)
	if err != nil {
		return f.unanswered(failed(step, start, err)), nil
	}
	return f.outcome(start, answer), nil
}

// Form states a step's form record names.
const (
	FormAnswered   = "answered"
	FormRejected   = "rejected"
	FormSupplied   = "supplied"
	FormUnanswered = "unanswered"
)

// record is the form a step showed and how it was answered, published under
// fixtures.FormMetadataKey for the UI to render read-only with the schema.
func (f *form) record(state string, values map[string]any, answer prompt.Answer) map[string]any {
	record := map[string]any{"title": firstNonEmpty(f.cfg.Title, f.id), "schema": f.schema, "values": values, "state": state}
	if f.cfg.Description != "" {
		record["description"] = f.cfg.Description
	}
	if answer.By != "" {
		record["answeredBy"] = answer.By
	}
	if !answer.At.IsZero() {
		record["answeredAt"] = answer.At.UTC().Format(time.RFC3339)
	}
	return record
}

// unanswered stamps a failed step with its form and the values it was prefilled with.
func (f *form) unanswered(outcome fixtures.StepOutcome) fixtures.StepOutcome {
	outcome.Result.Metadata = map[string]any{fixtures.FormMetadataKey: f.record(FormUnanswered, f.values, prompt.Answer{})}
	return outcome
}

// form is one step's resolved form: its rendered config, normalised schema
// and prefilled values.
type form struct {
	step   *fixtures.StepContext
	cfg    config
	id     string
	schema json.RawMessage
	shape  formSchema
	values map[string]any
}

func newForm(step *fixtures.StepContext, cfg config) (*form, error) {
	id := firstNonEmpty(step.Spec.ID, step.Test.Name)
	schema, err := normaliseSchema(cfg.Schema)
	if err != nil {
		return nil, fmt.Errorf("form %q: %w", id, err)
	}
	shape, err := parseFormSchema(schema)
	if err != nil {
		return nil, fmt.Errorf("form %q: %w", id, err)
	}
	var supplied map[string]any
	if id == fixtures.ParamsStepID {
		supplied = step.Opts.Params
	}
	values, err := shape.prefill(cfg.Default, supplied)
	if err != nil {
		return nil, fmt.Errorf("form %q: %w", id, err)
	}
	return &form{step: step, cfg: cfg, id: id, schema: schema, shape: shape, values: values}, nil
}

func (f *form) noPrompter() error {
	err := fmt.Errorf("form %q needs an interactive prompter (--ui or a TTY)", f.id)
	if missing := f.shape.missing(f.values); len(missing) > 0 {
		return fmt.Errorf("%w; missing required: %s", err, strings.Join(missing, ", "))
	}
	if invalid := prompt.Validate(f.schema, f.values); invalid != nil {
		return fmt.Errorf("%w; %v", err, invalid)
	}
	return err
}

// ask blocks on the run's prompter with the fixture node marked waiting.
func (f *form) ask(ctx context.Context) (prompt.Answer, error) {
	p := prompt.Prompt{
		Kind: fixtures.FormKind, Title: firstNonEmpty(f.cfg.Title, f.id), Description: f.cfg.Description,
		Schema: f.schema, Default: f.values, Owner: f.step.Opts.RunID, CreatedAt: time.Now(),
		// node is the step's name in the run's progress tree, so the run view can
		// show the pending form inline on that step.
		Labels: map[string]string{"fixture": f.step.Test.SourceDir, "step": f.id, "node": f.step.Test.Name},
	}
	if err := fixtures.ReportWaiting(ctx, true); err != nil {
		return prompt.Answer{}, fmt.Errorf("form %q: report waiting: %w", f.id, err)
	}
	answer, err := f.step.Opts.Prompter.Ask(ctx, p)
	err = errors.Join(err, context.Cause(ctx))
	if resumeErr := fixtures.ReportWaiting(ctx, false); resumeErr != nil {
		err = errors.Join(err, fmt.Errorf("report running: %w", resumeErr))
	}
	if err != nil {
		return prompt.Answer{}, fmt.Errorf("form %q: %w", f.id, err)
	}
	if answer.Cancelled {
		return answer, nil
	}
	if err := prompt.Validate(f.schema, answer.Values); err != nil {
		return prompt.Answer{}, fmt.Errorf("form %q: answer does not match its schema: %w", f.id, err)
	}
	if missing := f.shape.missing(answer.Values); len(missing) > 0 {
		return prompt.Answer{}, fmt.Errorf("form %q: answer does not match its schema: missing required %s", f.id, strings.Join(missing, ", "))
	}
	answer.Values = f.shape.prune(f.shape.integers(maps.Clone(answer.Values)))
	return answer, nil
}

// outcome maps an answer onto the step's outputs: a cancelled form is
// approved=false and passes; `if:` on a later step decides what that means.
func (f *form) outcome(start time.Time, answer prompt.Answer) fixtures.StepOutcome {
	values := answer.Values
	if answer.Cancelled || values == nil {
		values = map[string]any{}
	}
	at := answer.At
	if at.IsZero() {
		at = time.Now()
	}
	outputs := maps.Clone(values)
	outputs["approved"] = !answer.Cancelled
	outputs["values"] = values
	outputs["answeredBy"] = answer.By
	outputs["answeredAt"] = at.UTC().Format(time.RFC3339)
	state := FormAnswered
	switch {
	case answer.Cancelled:
		state = FormRejected
	case answer.By == AnsweredBySupplied:
		state = FormSupplied
	}
	answer.At = at
	res := result(f.step, start, task.StatusPASS, "")
	res.Metadata = map[string]any{fixtures.FormMetadataKey: f.record(state, values, answer)}
	return fixtures.StepOutcome{Result: res, Outputs: outputs}
}

func failed(step *fixtures.StepContext, start time.Time, err error) fixtures.StepOutcome {
	return fixtures.StepOutcome{Result: result(step, start, task.StatusFAIL, err.Error())}
}

func result(step *fixtures.StepContext, start time.Time, status task.Status, message string) fixtures.FixtureResult {
	return fixtures.FixtureResult{Name: firstNonEmpty(step.Spec.Name, step.Spec.ID, step.Test.Name), Type: step.Spec.Kind,
		Start: &start, Duration: time.Since(start), Status: status, Error: message}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
