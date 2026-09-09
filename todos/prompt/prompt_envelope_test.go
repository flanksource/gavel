package prompt

import (
	"encoding/json"
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

func TestPlanEnvelopeSchemaUsesFlatScalarFields(t *testing.T) {
	raw, err := EnvelopeSchemaJSON(EnvelopePlan)
	if err != nil {
		t.Fatalf("EnvelopeSchemaJSON: %v", err)
	}
	got, err := captainai.SchemaJSONForRuntime(api.Anthropic, captainai.ModeAgent, api.Prompt{SchemaJSON: raw})
	if err != nil {
		t.Fatalf("SchemaJSONForRuntime: %v", err)
	}

	var root map[string]any
	if err := json.Unmarshal(got, &root); err != nil {
		t.Fatalf("decode transformed schema: %v", err)
	}
	defs, ok := root["$defs"].(map[string]any)
	if !ok {
		t.Fatalf("transformed schema has no $defs: %s", got)
	}
	planEnvelope, ok := defs["PlanEnvelope"].(map[string]any)
	if !ok {
		t.Fatalf("transformed schema has no PlanEnvelope definition: %s", got)
	}
	properties, ok := planEnvelope["properties"].(map[string]any)
	if !ok {
		t.Fatalf("PlanEnvelope has no properties: %#v", planEnvelope)
	}
	for _, field := range []string{"summary", "endStatus", "planStatus", "planPath", "planContent"} {
		property, ok := properties[field].(map[string]any)
		if !ok {
			t.Fatalf("PlanEnvelope.%s is %T, want schema object", field, properties[field])
		}
		if property["$ref"] != nil {
			t.Fatalf("PlanEnvelope.%s = %#v, want a top-level scalar", field, property)
		}
	}
	if properties["plan"] != nil {
		t.Fatalf("PlanEnvelope.plan = %#v, want no nested plan object", properties["plan"])
	}
}

// TestTriageEnvelopeSchemaUsesFlatScalarFields mirrors the plan envelope's
// guard: a nested object becomes a $ref node, which some backends refuse to
// emit, so every triage field must be a top-level scalar or scalar array.
func TestTriageEnvelopeSchemaUsesFlatScalarFields(t *testing.T) {
	raw, err := EnvelopeSchemaJSON(EnvelopeTriage)
	if err != nil {
		t.Fatalf("EnvelopeSchemaJSON: %v", err)
	}
	got, err := captainai.SchemaJSONForRuntime(api.Anthropic, captainai.ModeAgent, api.Prompt{SchemaJSON: raw})
	if err != nil {
		t.Fatalf("SchemaJSONForRuntime: %v", err)
	}

	var root map[string]any
	if err := json.Unmarshal(got, &root); err != nil {
		t.Fatalf("decode transformed schema: %v", err)
	}
	defs, ok := root["$defs"].(map[string]any)
	if !ok {
		t.Fatalf("transformed schema has no $defs: %s", got)
	}
	triage, ok := defs["TriageEnvelope"].(map[string]any)
	if !ok {
		t.Fatalf("transformed schema has no TriageEnvelope definition: %s", got)
	}
	properties, ok := triage["properties"].(map[string]any)
	if !ok {
		t.Fatalf("TriageEnvelope has no properties: %#v", triage)
	}
	for _, field := range []string{
		"summary", "endStatus", "verdict", "title", "body",
		"verification", "priority", "status", "duplicateOf", "comment",
	} {
		property, ok := properties[field].(map[string]any)
		if !ok {
			t.Fatalf("TriageEnvelope.%s is %T, want schema object", field, properties[field])
		}
		if property["$ref"] != nil {
			t.Fatalf("TriageEnvelope.%s = %#v, want a top-level scalar", field, property)
		}
	}
	related, ok := properties["related"].(map[string]any)
	if !ok {
		t.Fatalf("TriageEnvelope.related is %T, want schema object", properties["related"])
	}
	if related["type"] != "array" {
		t.Fatalf("TriageEnvelope.related = %#v, want an array of scalars", related)
	}
}

// TestEnvelopeSchemaBytesStable pins the schema identity shared by initial,
// retry, and feedback turns in a claude-agent session.
func TestEnvelopeSchemaBytesStable(t *testing.T) {
	for _, name := range []string{"run", "plan", "triage"} {
		req, _, err := renderResolvedForTest([]*types.TODO{newTestTODO("solo", "task")}, Options{Prompt: name, Envelope: envelopeForPrompt(t, name)})
		if err != nil {
			t.Fatalf("Render(%s): %v", name, err)
		}
		initial := string(req.Prompt.SchemaJSON)
		if initial == "" {
			t.Fatalf("Render(%s) produced no envelope schema", name)
		}

		recomputed, err := EnvelopeSchemaJSON(envelopeForPrompt(t, name))
		if err != nil {
			t.Fatalf("EnvelopeSchemaJSON(%s): %v", name, err)
		}
		if string(recomputed) != initial {
			t.Errorf("%s: recomputed schema differs from Render's:\n initial=%s\n recomputed=%s", name, initial, recomputed)
		}
	}
}

// envelopeForPrompt resolves a built-in prompt's envelope through the catalog, so
// the test asserts against the same wiring production uses rather than a literal.
func envelopeForPrompt(t *testing.T, name string) EnvelopeKind {
	t.Helper()
	catalog, err := NewCatalog(verify.TodosConfig{})
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	def, err := catalog.Lookup(name)
	if err != nil {
		t.Fatalf("Lookup(%s): %v", name, err)
	}
	return def.Envelope
}
