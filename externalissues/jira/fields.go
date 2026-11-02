package jira

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/externalissues"
)

type FieldDefinition struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Schema        json.RawMessage `json:"schema,omitempty"`
	Required      *bool           `json:"required,omitempty"`
	Operations    []string        `json:"operations,omitempty"`
	AllowedValues json.RawMessage `json:"allowedValues,omitempty"`
}

func MapFields(fields map[string]json.RawMessage, catalog map[string]FieldDefinition) (map[string]externalissues.Field, error) {
	result := make(map[string]externalissues.Field, len(fields))
	for id, value := range fields {
		if strings.TrimSpace(id) == "" || !json.Valid(value) {
			return nil, fmt.Errorf("field %q requires an ID and valid JSON", id)
		}
		definition := catalog[id]
		normalized, format, err := normalizeFieldValue(value, "field "+id+" $")
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", id, err)
		}
		result[id] = externalissues.Field{ID: id, Name: strings.TrimSpace(definition.Name), Schema: definition.Schema, Format: format, Value: normalized}
	}
	return result, nil
}

func normalizeFieldValue(raw json.RawMessage, path string) (json.RawMessage, string, error) {
	switch strings.TrimSpace(string(raw))[0] {
	case '{':
		var object map[string]json.RawMessage
		if err := json.Unmarshal(raw, &object); err != nil {
			return nil, "", err
		}
		var nodeType string
		if json.Unmarshal(object["type"], &nodeType) == nil && nodeType == "doc" {
			node, err := renderADF(raw, adfRenderOptions{Path: path})
			if err != nil {
				return nil, "", err
			}
			value, err := json.Marshal(strings.TrimSpace(node.Markdown()))
			return value, "markdown", err
		}
		for key, value := range object {
			normalized, _, err := normalizeFieldValue(value, path+"."+key)
			if err != nil {
				return nil, "", fmt.Errorf("property %s: %w", key, err)
			}
			object[key] = normalized
		}
		value, err := json.Marshal(object)
		return value, "", err
	case '[':
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return nil, "", err
		}
		for index, item := range items {
			normalized, _, err := normalizeFieldValue(item, fmt.Sprintf("%s[%d]", path, index))
			if err != nil {
				return nil, "", fmt.Errorf("item %d: %w", index, err)
			}
			items[index] = normalized
		}
		value, err := json.Marshal(items)
		return value, "", err
	default:
		return raw, "", nil
	}
}
