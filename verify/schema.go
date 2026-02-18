package verify

import (
	"encoding/json"
	"os"
)

func evidenceSchema() map[string]any {
	return map[string]any{
		"type": "array",
		"items": map[string]any{
			"type":                 "object",
			"required":            []string{"file", "line", "message"},
			"additionalProperties": false,
			"properties": map[string]any{
				"file":    map[string]any{"type": "string"},
				"line":    map[string]any{"type": "integer"},
				"message": map[string]any{"type": "string", "maxLength": 500},
			},
		},
	}
}

func BuildSchema(checks []Check) (string, error) {
	checkProps := make(map[string]any, len(checks))
	for _, c := range checks {
		checkProps[c.ID] = map[string]any{
			"type":                 "object",
			"description":         c.Description,
			"required":            []string{"pass", "evidence"},
			"additionalProperties": false,
			"properties": map[string]any{
				"pass":     map[string]any{"type": "boolean"},
				"evidence": evidenceSchema(),
			},
		}
	}

	ratingProps := make(map[string]any, len(RatingDimensions))
	for _, dim := range RatingDimensions {
		ratingProps[dim] = map[string]any{
			"type":                 "object",
			"required":            []string{"score", "findings"},
			"additionalProperties": false,
			"properties": map[string]any{
				"score":    map[string]any{"type": "integer", "minimum": 0, "maximum": 100},
				"findings": evidenceSchema(),
			},
		}
	}

	schema := map[string]any{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"type":                 "object",
		"description":         "Code review result with boolean checks, rated dimensions, and completeness assessment. Rating rubric: 0-39 critical, 40-59 significant, 60-79 minor, 80-100 good.",
		"required":            []string{"checks", "ratings", "completeness"},
		"additionalProperties": false,
		"properties": map[string]any{
			"checks": map[string]any{
				"type":                 "object",
				"description":         "Boolean pass/fail checks. Evaluate every check. Only include evidence for failures.",
				"required":            checkIDs(checks),
				"additionalProperties": false,
				"properties":          checkProps,
			},
			"ratings": map[string]any{
				"type":                 "object",
				"description":         "Rated dimensions (0-100). Include findings for scores below 80.",
				"required":            RatingDimensions,
				"additionalProperties": false,
				"properties":          ratingProps,
			},
			"completeness": map[string]any{
				"type":                 "object",
				"description":         "Overall completeness assessment against the diff, issue, and extra instructions.",
				"required":            []string{"pass", "summary", "evidence"},
				"additionalProperties": false,
				"properties": map[string]any{
					"pass":     map[string]any{"type": "boolean"},
					"summary":  map[string]any{"type": "string"},
					"evidence": evidenceSchema(),
				},
			},
		},
	}

	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func checkIDs(checks []Check) []string {
	ids := make([]string, len(checks))
	for i, c := range checks {
		ids[i] = c.ID
	}
	return ids
}

func SchemaFile(cfg ChecksConfig) (string, error) {
	schema, err := BuildSchema(EnabledChecks(cfg))
	if err != nil {
		return "", err
	}

	f, err := os.CreateTemp("", "gavel-schema-*.json")
	if err != nil {
		return "", err
	}
	if _, err := f.WriteString(schema); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	f.Close()
	return f.Name(), nil
}
