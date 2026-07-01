package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

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

// BuildSchema builds the JSON output schema for a verification run. When
// issueAware is set the schema requires an overall `implemented` verdict; when
// criteria is additionally non-empty it requires an `acceptance_criteria` array
// with one scored entry per stored criterion.
func BuildSchema(checks []Check, issueAware bool, criteria []string) (string, error) {
	checkProps := make(map[string]any, len(checks))
	for _, c := range checks {
		checkProps[c.ID] = map[string]any{
			"type":                 "object",
			"description":          c.Description,
			"required":             []string{"pass", "evidence"},
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
			"required":             []string{"score", "findings"},
			"additionalProperties": false,
			"properties": map[string]any{
				"score":    map[string]any{"type": "integer", "minimum": 0, "maximum": 100},
				"findings": evidenceSchema(),
			},
		}
	}

	properties := map[string]any{
		"checks": map[string]any{
			"type":                 "object",
			"description":          "Boolean pass/fail checks. Evaluate every check. Only include evidence for failures.",
			"required":             checkIDs(checks),
			"additionalProperties": false,
			"properties":           checkProps,
		},
		"ratings": map[string]any{
			"type":                 "object",
			"description":          "Rated dimensions (0-100). Include findings for scores below 80.",
			"required":             RatingDimensions,
			"additionalProperties": false,
			"properties":           ratingProps,
		},
		"completeness": map[string]any{
			"type":                 "object",
			"description":          "Overall completeness assessment against the diff, issue, and extra instructions.",
			"required":             []string{"pass", "summary", "evidence"},
			"additionalProperties": false,
			"properties": map[string]any{
				"pass":     map[string]any{"type": "boolean"},
				"summary":  map[string]any{"type": "string"},
				"evidence": evidenceSchema(),
			},
		},
	}
	required := []string{"checks", "ratings", "completeness"}

	if issueAware {
		properties["implemented"] = map[string]any{
			"type":        "boolean",
			"description": "True only if the commits fully and correctly implement the issue and every required acceptance criterion is met.",
		}
		required = append(required, "implemented")
	}
	if len(criteria) > 0 {
		// Generate the acceptance-criteria schema FROM the criteria so the prompt
		// doesn't have to enumerate them: one array entry per criterion, in this
		// exact order, each {criteria, pass, comments}.
		var desc strings.Builder
		desc.WriteString("One verdict per acceptance criterion below, in this exact order. ")
		desc.WriteString("Echo the criterion text in `criteria`, set `pass` true only when the commits clearly satisfy it, and justify in `comments`:\n")
		for i, c := range criteria {
			fmt.Fprintf(&desc, "%d. %s\n", i+1, c)
		}
		properties["acceptance_criteria"] = map[string]any{
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
		required = append(required, "acceptance_criteria")
	}

	schema := map[string]any{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"type":                 "object",
		"description":          "Code review result with boolean checks, rated dimensions, and completeness assessment. Rating rubric: 0-39 critical, 40-59 significant, 60-79 minor, 80-100 good.",
		"required":             required,
		"additionalProperties": false,
		"properties":           properties,
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

func SchemaFile(cfg ChecksConfig, issueAware bool, criteria []string) (string, error) {
	schema, err := BuildSchema(EnabledChecks(cfg), issueAware, criteria)
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
