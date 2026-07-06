package verify

import (
	"fmt"
	"strings"
)

// evidenceSchema is the shared {file, line, message} evidence array used by
// checks, ratings, and completeness. It is a Go primitive the schema.cel
// `evidence_schema()` helper returns.
func evidenceSchema() map[string]any {
	return map[string]any{
		"type": "array",
		"items": map[string]any{
			"type":                 "object",
			"required":             []string{"file", "line", "message"},
			"additionalProperties": false,
			"properties": map[string]any{
				"file":    map[string]any{"type": "string"},
				"line":    map[string]any{"type": "integer"},
				"message": map[string]any{"type": "string", "maxLength": 500},
			},
		},
	}
}

// checksObjectSchema builds the `checks` object schema: one boolean pass/fail
// entry (with failure evidence) per enabled check, keyed by check ID. CEL cannot
// build a map with runtime keys, so this is the Go primitive the schema.cel
// `check_schema(checks)` helper returns. Each check map carries "id"+"description".
func checksObjectSchema(checks []map[string]any) map[string]any {
	props := make(map[string]any, len(checks))
	ids := make([]string, 0, len(checks))
	for _, c := range checks {
		id, _ := c["id"].(string)
		desc, _ := c["description"].(string)
		props[id] = map[string]any{
			"type":                 "object",
			"description":          desc,
			"required":             []string{"pass", "evidence"},
			"additionalProperties": false,
			"properties": map[string]any{
				"pass":     map[string]any{"type": "boolean"},
				"evidence": evidenceSchema(),
			},
		}
		ids = append(ids, id)
	}
	return map[string]any{
		"type":                 "object",
		"description":          "Boolean pass/fail checks. Evaluate every check. Only include evidence for failures.",
		"required":             ids,
		"additionalProperties": false,
		"properties":           props,
	}
}

// ratingsObjectSchema builds the `ratings` object schema: a 0-100 score (with
// findings) per rated dimension, keyed by dimension name. It backs the
// schema.cel `rating_schema(rating_dimensions)` helper.
func ratingsObjectSchema(dimensions []string) map[string]any {
	props := make(map[string]any, len(dimensions))
	for _, dim := range dimensions {
		props[dim] = map[string]any{
			"type":                 "object",
			"required":             []string{"score", "findings"},
			"additionalProperties": false,
			"properties": map[string]any{
				"score":    map[string]any{"type": "integer", "minimum": 0, "maximum": 100},
				"findings": evidenceSchema(),
			},
		}
	}
	return map[string]any{
		"type":                 "object",
		"description":          "Rated dimensions (0-100). Include findings for scores below 80.",
		"required":             dimensions,
		"additionalProperties": false,
		"properties":           props,
	}
}

// criteriaArraySchema builds the `acceptance_criteria` array schema: exactly one
// scored verdict per criterion, in order, with the criteria enumerated in the
// description so the prompt need not repeat them. It backs the schema.cel
// `criteria_schema(criteria)` helper.
func criteriaArraySchema(criteria []string) map[string]any {
	var desc strings.Builder
	desc.WriteString("One verdict per acceptance criterion below, in this exact order. ")
	desc.WriteString("Echo the criterion text in `criteria`, set `pass` true only when the commits clearly satisfy it, and justify in `comments`:\n")
	for i, c := range criteria {
		fmt.Fprintf(&desc, "%d. %s\n", i+1, c)
	}
	return map[string]any{
		"type":        "array",
		"description": desc.String(),
		"minItems":    len(criteria),
		"maxItems":    len(criteria),
		"items": map[string]any{
			"type":                 "object",
			"required":             []string{"criteria", "pass", "comments"},
			"additionalProperties": false,
			"properties": map[string]any{
				"criteria": map[string]any{"type": "string"},
				"pass":     map[string]any{"type": "boolean"},
				"comments": map[string]any{"type": "string"},
			},
		},
	}
}

// BuildSchema returns the JSON output schema for a verification run by evaluating
// the default schema.cel expression. The schema is defined declaratively in CEL
// (see schema.cel / DefaultSchemaCEL); this is a convenience wrapper for callers
// that use the built-in schema with no per-fixture override.
func BuildSchema(checks []Check, issueAware bool, criteria []string) (string, error) {
	return EvalSchema("", checks, issueAware, criteria)
}
