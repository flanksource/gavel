package schemas

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type Document struct {
	Schema     map[string]any
	Properties map[string]any
	Order      []string
	validator  *jsonschema.Schema
}

var documents = sync.OnceValues(loadDocuments)

func Load(name string) (*Document, error) {
	loaded, err := documents()
	if err != nil {
		return nil, err
	}
	document, ok := loaded[name]
	if !ok {
		return nil, fmt.Errorf("unsupported action schema %q", name)
	}
	return document, nil
}

func loadDocuments() (map[string]*Document, error) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("read embedded schemas: %w", err)
	}
	loaded := make(map[string]*Document, len(entries))
	for _, entry := range entries {
		data, err := FS.ReadFile(entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read schema %s: %w", entry.Name(), err)
		}
		document, err := decodeDocument(data)
		if err != nil {
			return nil, fmt.Errorf("schema %s: %w", entry.Name(), err)
		}
		loaded[strings.TrimSuffix(entry.Name(), ".schema.json")] = document
	}
	return loaded, nil
}

func decodeDocument(data []byte) (*Document, error) {
	var document Document
	if err := json.Unmarshal(data, &document.Schema); err != nil {
		return nil, err
	}
	var shape struct {
		Properties map[string]any `json:"properties"`
		Order      []string       `json:"x-order"`
	}
	if err := json.Unmarshal(data, &shape); err != nil {
		return nil, err
	}
	if err := validateOrder(shape.Properties, shape.Order); err != nil {
		return nil, err
	}
	compiler := jsonschema.NewCompiler()
	const location = "urn:gavel:embedded-action"
	if err := compiler.AddResource(location, document.Schema); err != nil {
		return nil, err
	}
	compiled, err := compiler.Compile(location)
	if err != nil {
		return nil, err
	}
	document.Properties, document.Order, document.validator = shape.Properties, shape.Order, compiled
	return &document, nil
}

func validateOrder(properties map[string]any, order []string) error {
	if len(properties) == 0 || len(order) != len(properties) {
		return fmt.Errorf("x-order must cover every property exactly once")
	}
	seen := make(map[string]bool, len(order))
	for _, name := range order {
		if _, ok := properties[name]; !ok || seen[name] {
			return fmt.Errorf("invalid or duplicate x-order property %q", name)
		}
		property, ok := properties[name].(map[string]any)
		if !ok {
			return fmt.Errorf("property %q must be an object schema", name)
		}
		if property["type"] == "array" {
			if _, ok := property["items"].(map[string]any); !ok {
				return fmt.Errorf("array property %q must declare its item schema", name)
			}
		}
		seen[name] = true
	}
	return nil
}

func (d *Document) Validate(value map[string]any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var normalized any
	if err := json.Unmarshal(data, &normalized); err != nil {
		return err
	}
	return d.validator.Validate(normalized)
}
