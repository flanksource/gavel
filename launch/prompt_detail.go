package launch

import (
	"encoding/json"
	"maps"
	"strings"

	captaincli "github.com/flanksource/captain/pkg/cli"
)

func promptFormDetail(detail captaincli.PromptDetail) (map[string]any, error) {
	schema, err := captaincli.PromptFormSchema(detail)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(detail.InputDefault)
	if err != nil {
		return nil, err
	}
	var values any
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, err
	}
	defaults, _ := publicInput(values, schema, schema)
	stripPrivateDefaults(schema, schema)
	runtime := map[string]any{}
	if detail.Run.Spec != nil {
		data, err := json.Marshal(detail.Run.Spec)
		if err != nil {
			return nil, err
		}
		var native map[string]any
		if err := json.Unmarshal(data, &native); err != nil {
			return nil, err
		}
		for _, key := range []string{"model", "mode", "effort", "budget"} {
			if value, exists := native[key]; exists {
				runtime[key] = value
			}
		}
	}
	prompt := map[string]any{"id": detail.ID, "name": detail.Name, "description": detail.Description, "version": detail.Version, "presets": detail.Presets, "variables": detail.Variables, "inputSchema": schema, "inputDefault": defaults, "run": map[string]any{"presets": detail.Run.Presets, "spec": runtime}}
	return map[string]any{"prompt": prompt, "variablesSchema": schema}, nil
}

func stripPrivateDefaults(value any, root map[string]any) {
	switch node := value.(type) {
	case map[string]any:
		if initial, exists := node["default"]; exists {
			if safe, include := publicInput(initial, node, root); include {
				node["default"] = safe
			} else {
				delete(node, "default")
			}
		}
		for _, child := range node {
			stripPrivateDefaults(child, root)
		}
	case []any:
		for _, child := range node {
			stripPrivateDefaults(child, root)
		}
	}
}

func publicInput(value any, schema, root map[string]any) (any, bool) {
	return publicInputRefs(value, schema, root, map[string]bool{})
}

func publicInputRefs(value any, schema, root map[string]any, visited map[string]bool) (any, bool) {
	if schema == nil {
		return nil, false
	}
	if schema["writeOnly"] == true {
		return nil, false
	}
	projected := false
	if reference, ok := schema["$ref"].(string); ok {
		if visited[reference] {
			return nil, false
		}
		visited[reference] = true
		if !strings.HasPrefix(reference, "#/") {
			return nil, false
		}
		resolved := any(root)
		for _, key := range strings.Split(strings.TrimPrefix(reference, "#/"), "/") {
			node, ok := resolved.(map[string]any)
			if !ok {
				return nil, false
			}
			resolved = node[strings.ReplaceAll(strings.ReplaceAll(key, "~1", "/"), "~0", "~")]
		}
		node, ok := resolved.(map[string]any)
		if !ok {
			return nil, false
		}
		var include bool
		value, include = publicInputRefs(value, node, root, visited)
		if !include {
			return nil, false
		}
		projected = true
	}
	for _, keyword := range []string{"allOf", "anyOf", "oneOf"} {
		branches, _ := schema[keyword].([]any)
		for _, branch := range branches {
			constraint, ok := branch.(map[string]any)
			if !ok {
				return nil, false
			}
			var include bool
			value, include = publicInputRefs(value, constraint, root, maps.Clone(visited))
			if !include {
				return nil, false
			}
			projected = true
		}
	}
	switch node := value.(type) {
	case map[string]any:
		result := map[string]any{}
		properties, _ := schema["properties"].(map[string]any)
		if projected && properties == nil {
			return value, true
		}
		for key, item := range node {
			property, _ := properties[key].(map[string]any)
			if safe, include := publicInput(item, property, root); include {
				result[key] = safe
			}
		}
		return result, true
	case []any:
		items, _ := schema["items"].(map[string]any)
		if projected && items == nil {
			return value, true
		}
		result := make([]any, len(node))
		for index, item := range node {
			safe, include := publicInput(item, items, root)
			if !include {
				return nil, false
			}
			result[index] = safe
		}
		return result, true
	}
	return value, true
}
