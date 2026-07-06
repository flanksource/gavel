package verify

import (
	"encoding/json"
	"testing"
)

// A custom schema.cel expression can read the parsed outline (criteria, checks)
// and the helper functions, overriding the default schema entirely.
func TestEvalSchema_Override(t *testing.T) {
	checks := []Check{{ID: "definition-of-done", Description: "DoD met"}}
	criteria := []string{"endpoint streams NDJSON", "errors are surfaced"}

	expr := `{
		"type": "object",
		"required": ["acceptance_criteria"],
		"properties": {"acceptance_criteria": criteria_schema(criteria)}
	}`
	out, err := EvalSchema(expr, checks, true, criteria)
	if err != nil {
		t.Fatalf("EvalSchema override: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("override schema is not valid JSON: %v", err)
	}
	props := decoded["properties"].(map[string]any)
	ac := props["acceptance_criteria"].(map[string]any)
	if ac["type"] != "array" {
		t.Errorf("acceptance_criteria type = %v, want array", ac["type"])
	}
	// The array length is pinned to the number of criteria (minItems==maxItems==N).
	if got := ac["minItems"]; got != float64(len(criteria)) {
		t.Errorf("minItems = %v, want %d", got, len(criteria))
	}
	if got := ac["maxItems"]; got != float64(len(criteria)) {
		t.Errorf("maxItems = %v, want %d", got, len(criteria))
	}
}

// A syntactically invalid expression fails loudly rather than yielding an empty
// or partial schema.
func TestEvalSchema_InvalidExpr(t *testing.T) {
	if _, err := EvalSchema("this is not ( cel", nil, false, nil); err == nil {
		t.Fatal("expected a compile error for an invalid schema.cel expression")
	}
}
