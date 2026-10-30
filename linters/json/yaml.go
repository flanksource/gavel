package jsonlint

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"gopkg.in/yaml.v3"
)

func yamlValue(node *yaml.Node, active map[*yaml.Node]bool) (any, error) {
	if active[node] {
		return nil, fmt.Errorf("line %d: cyclic YAML alias", node.Line)
	}
	active[node] = true
	defer delete(active, node)
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			return nil, nil
		}
		return yamlValue(node.Content[0], active)
	case yaml.AliasNode:
		return yamlValue(node.Alias, active)
	case yaml.MappingNode:
		return yamlObject(node, active)
	case yaml.SequenceNode:
		result := make([]any, len(node.Content))
		for i, child := range node.Content {
			value, err := yamlValue(child, active)
			if err != nil {
				return nil, err
			}
			result[i] = value
		}
		return result, nil
	default:
		return yamlScalar(node)
	}
}

func yamlObject(node *yaml.Node, active map[*yaml.Node]bool) (map[string]any, error) {
	result := map[string]any{}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Tag != "!!merge" {
			continue
		}
		value, err := yamlValue(node.Content[i+1], active)
		if err != nil {
			return nil, err
		}
		if err := mergeYAML(result, value); err != nil {
			return nil, fmt.Errorf("line %d: %w", node.Content[i].Line, err)
		}
	}
	seen := map[string]bool{}
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		if key.Tag == "!!merge" {
			continue
		}
		if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
			return nil, fmt.Errorf("line %d: JSON objects require string keys", key.Line)
		}
		if seen[key.Value] {
			return nil, fmt.Errorf("line %d: duplicate mapping key %q", key.Line, key.Value)
		}
		seen[key.Value] = true
		value, err := yamlValue(node.Content[i+1], active)
		if err != nil {
			return nil, err
		}
		result[key.Value] = value
	}
	return result, nil
}

func mergeYAML(result map[string]any, value any) error {
	switch value := value.(type) {
	case map[string]any:
		for key, entry := range value {
			if _, exists := result[key]; !exists {
				result[key] = entry
			}
		}
	case []any:
		for _, entry := range value {
			if _, ok := entry.(map[string]any); !ok {
				return fmt.Errorf("YAML merge requires mappings")
			}
			if err := mergeYAML(result, entry); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("YAML merge requires a mapping or sequence of mappings")
	}
	return nil
}

func yamlScalar(node *yaml.Node) (any, error) {
	switch node.Tag {
	case "!!str", "!!timestamp":
		return node.Value, nil
	case "!!null":
		return nil, nil
	case "!!int":
		text := strings.ReplaceAll(node.Value, "_", "")
		number, ok := new(big.Int).SetString(text, 0)
		if !ok {
			return nil, fmt.Errorf("line %d: invalid integer %q", node.Line, node.Value)
		}
		return json.Number(number.String()), nil
	case "!!float":
		return yamlFloat(node)
	default:
		var value any
		if err := node.Decode(&value); err != nil {
			return nil, err
		}
		return value, nil
	}
}

func yamlFloat(node *yaml.Node) (json.Number, error) {
	text, negative := strings.CutPrefix(strings.TrimPrefix(strings.ReplaceAll(node.Value, "_", ""), "+"), "-")
	mantissa, exponent, hasExponent := strings.Cut(strings.ToLower(text), "e")
	whole, fraction, hasFraction := strings.Cut(mantissa, ".")
	whole = strings.TrimLeft(whole, "0")
	if whole == "" {
		whole = "0"
	}
	text = whole
	if hasFraction {
		if fraction == "" {
			fraction = "0"
		}
		text += "." + fraction
	}
	if hasExponent {
		text += "e" + exponent
	}
	if negative {
		text = "-" + text
	}
	if !json.Valid([]byte(text)) {
		return "", fmt.Errorf("line %d: %q cannot be represented as a JSON number", node.Line, node.Value)
	}
	return json.Number(text), nil
}
