package fixtures

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// validateKindConfig checks a step's config against its kind's JSON schema.
func validateKindConfig(kind Kind, config json.RawMessage) error {
	schema := kind.Schema()
	if len(schema) == 0 {
		return fmt.Errorf("kind %q declares no config schema", kind.Name())
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schema))
	if err != nil {
		return fmt.Errorf("kind %q: parse schema: %w", kind.Name(), err)
	}
	url := "kind://" + kind.Name() + ".json"
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(url, doc); err != nil {
		return fmt.Errorf("kind %q: load schema: %w", kind.Name(), err)
	}
	compiled, err := compiler.Compile(url)
	if err != nil {
		return fmt.Errorf("kind %q: compile schema: %w", kind.Name(), err)
	}
	if len(config) == 0 {
		config = json.RawMessage(`{}`)
	}
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(config))
	if err != nil {
		return fmt.Errorf("decode config: %w", err)
	}
	if err := compiled.Validate(instance); err != nil {
		return fmt.Errorf("config does not match the %s schema: %w", kind.Name(), err)
	}
	return nil
}

var durationSchema = map[string]any{"type": "string", "format": "duration", "examples": []string{"500ms", "2s", "3m"}}

var backoffSchema = map[string]any{
	"type": "object", "additionalProperties": false,
	"properties": map[string]any{"factor": map[string]any{"type": "number", "minimum": 1}, "max": durationSchema},
}

var jitterSchema = map[string]any{"type": "string", "description": `A percentage ("20%") or an absolute duration ("500ms").`}

// WaitPolicySchema is the JSON schema of a `wait:` block.
func WaitPolicySchema() map[string]any {
	return map[string]any{
		"type": "object", "additionalProperties": false,
		"description": "Poll the step's observable state until a CEL verdict.",
		"properties": map[string]any{
			"until":    map[string]any{"type": "string", "description": "CEL over the step's outputs (+ env, steps); success when true."},
			"failWhen": map[string]any{"type": "string", "description": "CEL that ends the wait early, loudly."},
			"interval": durationSchema, "backoff": backoffSchema, "jitter": jitterSchema, "timeout": durationSchema,
		},
	}
}

// RetryPolicySchema is the JSON schema of a `retry:` block.
func RetryPolicySchema() map[string]any {
	return map[string]any{
		"type": "object", "additionalProperties": false,
		"description": "Re-run Execute while `when` holds over the failed result.",
		"properties": map[string]any{
			"attempts": map[string]any{"type": "integer", "minimum": 1},
			"when":     map[string]any{"type": "string", "description": "CEL over error, exitCode, outputs, status (and the outputs as roots)."},
			"delay":    durationSchema, "backoff": backoffSchema, "jitter": jitterSchema,
		},
	}
}

// stepEnvelopeProperties are the schema properties every step accepts.
func stepEnvelopeProperties() map[string]any {
	return map[string]any{
		"id":         map[string]any{"type": "string", "description": "Step id; later steps read its outputs as steps.<id>."},
		"name":       map[string]any{"type": "string"},
		"cel":        map[string]any{"type": "string", "description": "CEL verdict over the step's outputs, env and steps."},
		"export":     map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "string"}, "description": `Env vars for later steps: NAME: "{{.outputs.x}}"; "*" exports a map; "" suppresses a kind default.`},
		"override":   map[string]any{"type": "boolean", "description": "Allow re-exporting a name an earlier step exported."},
		"background": map[string]any{"type": "boolean", "description": "Run for the rest of the fixture; the verdict is collected at teardown."},
		"wait":       WaitPolicySchema(),
		"retry":      RetryPolicySchema(),
		"timeout":    durationSchema,
	}
}

// DocumentSchema is the JSON schema of a YAML/JSON fixture document, with a
// step oneOf over every registered kind. overrides replace a kind's config
// schema (the CLI derives the test/lint option sets from its own flags).
func DocumentSchema(overrides map[string]json.RawMessage) (map[string]any, error) {
	var branches []any
	for _, name := range DefaultKinds.Names() {
		kind, err := DefaultKinds.Lookup(name)
		if err != nil {
			return nil, err
		}
		raw := kind.Schema()
		if override, ok := overrides[name]; ok {
			raw = override
		}
		branch, err := stepBranchSchema(name, raw)
		if err != nil {
			return nil, err
		}
		branches = append(branches, branch)
	}
	defaults := map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{
		"wait": WaitPolicySchema(), "retry": RetryPolicySchema(), "timeout": durationSchema}}
	return map[string]any{
		"$schema": "https://json-schema.org/draft/2020-12/schema",
		"title":   "gavel fixture document",
		"type":    "object", "additionalProperties": false, "required": []string{"kind", "steps"},
		"properties": map[string]any{
			"kind":     map[string]any{"const": "Fixture"},
			"name":     map[string]any{"type": "string"},
			"defaults": defaults,
			"setup":    map[string]any{"type": "object", "description": "Same as the markdown front-matter setup:."},
			"env":      map[string]any{"type": "object", "additionalProperties": map[string]any{"type": []string{"string", "number", "boolean"}}},
			"steps":    map[string]any{"type": "array", "minItems": 1, "items": map[string]any{"oneOf": branches}},
		},
	}, nil
}

// stepBranchSchema merges the envelope with one kind's config schema: the
// kind's properties next to the envelope's, `kind` pinned to its name, and the
// kind's required/anyOf/additionalProperties carried over.
func stepBranchSchema(name string, raw json.RawMessage) (map[string]any, error) {
	var config map[string]any
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, fmt.Errorf("kind %q: schema is not a JSON object: %w", name, err)
	}
	properties := stepEnvelopeProperties()
	if kindProps, ok := config["properties"].(map[string]any); ok {
		maps.Copy(properties, kindProps)
	}
	properties["kind"] = map[string]any{"const": name}
	branch := map[string]any{"type": "object", "title": name, "properties": properties}
	required := []any{"kind"}
	if kindRequired, ok := config["required"].([]any); ok {
		required = append(required, kindRequired...)
	}
	branch["required"] = required
	for _, key := range []string{"description", "anyOf", "additionalProperties"} {
		if value, ok := config[key]; ok {
			branch[key] = value
		}
	}
	return branch, nil
}
