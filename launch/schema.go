package launch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/flanksource/commons-db/shell"
	"github.com/invopop/jsonschema"
	validator "github.com/santhosh-tekuri/jsonschema/v6"
)

func object(properties map[string]any, required ...string) map[string]any {
	value := map[string]any{"type": "object", "properties": properties, "additionalProperties": false}
	if len(required) > 0 {
		value["required"] = required
	}
	return value
}

func reflected(value any) (map[string]any, error) {
	schema := (&jsonschema.Reflector{}).Reflect(value)
	raw, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	err = json.Unmarshal(raw, &result)
	return result, err
}

func Schema(kind string, provider SchemaProvider) (map[string]any, error) {
	if kind != "" && kind != "test" && kind != "lint" {
		return nil, errors.New("unknown launch kind")
	}
	workspace, err := reflected(shell.Setup{})
	if err != nil {
		return nil, err
	}
	sections := map[string]any{}
	if kind != "lint" {
		test, err := provider.Schema("test")
		if err != nil {
			return nil, err
		}
		sections["test"] = test
	}
	if kind != "test" {
		lint, err := lintSchema(provider)
		if err != nil {
			return nil, err
		}
		sections["lint"] = lint
	}
	config := object(sections)
	if kind != "" {
		config["required"] = []string{kind}
	} else {
		config["anyOf"] = []any{map[string]any{"required": []string{"test"}}, map[string]any{"required": []string{"lint"}}}
	}
	root := object(map[string]any{
		"workspace": workspace, "preparationId": map[string]any{"type": "string", "minLength": 1},
		"config": config, "labels": map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "string"}},
	}, "config")
	root["$schema"] = "https://json-schema.org/draft/2020-12/schema"
	root["oneOf"] = []any{map[string]any{"required": []string{"workspace"}}, map[string]any{"required": []string{"preparationId"}}}
	return root, nil
}

func lintSchema(provider SchemaProvider) (map[string]any, error) {
	native, err := provider.Schema("lint")
	if err != nil {
		return nil, err
	}
	prompt, err := reflected(PromptConfig{})
	if err != nil {
		return nil, err
	}
	entry := object(map[string]any{
		"id":      map[string]any{"type": "string", "pattern": "^[A-Za-z][A-Za-z0-9_-]*$"},
		"linter":  map[string]any{"type": "string", "minLength": 1},
		"options": map[string]any{"type": "object"},
	}, "id", "linter")
	entry["allOf"] = []any{map[string]any{
		"if":   map[string]any{"properties": map[string]any{"linter": map[string]any{"const": "prompt"}}},
		"then": map[string]any{"required": []string{"options"}, "properties": map[string]any{"options": prompt}},
		"else": map[string]any{"properties": map[string]any{"options": native}},
	}}
	return object(map[string]any{
		"files":   map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"changed": map[string]any{"type": "boolean"}, "since": map[string]any{"type": "string"},
		"linters": map[string]any{"type": "array", "items": entry},
	}), nil
}

func Decode(data []byte, kind string, provider SchemaProvider) (Request, error) {
	schema, err := Schema(kind, provider)
	if err != nil {
		return Request{}, err
	}
	if err := Validate(schema, data); err != nil {
		return Request{}, err
	}
	var request Request
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return request, err
	}
	if request.Config.Lint != nil {
		seen := map[string]bool{}
		for _, linter := range request.Config.Lint.Linters {
			if seen[linter.ID] {
				return request, fmt.Errorf("duplicate linter configuration id %q", linter.ID)
			}
			seen[linter.ID] = true
		}
	}
	return request, nil
}

func Validate(schema map[string]any, data []byte) error {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	encoded, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	var resource any
	if err := json.Unmarshal(encoded, &resource); err != nil {
		return err
	}
	compiler := validator.NewCompiler()
	if err := compiler.AddResource("urn:gavel:launch", resource); err != nil {
		return err
	}
	compiled, err := compiler.Compile("urn:gavel:launch")
	if err != nil {
		return err
	}
	return compiled.Validate(value)
}
