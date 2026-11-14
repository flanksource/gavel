package form

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// ConnectionLookupURL is the commons-db connections endpoint a `format:
// connection` property picks from; ConnectionLookupFilter names the lookup.
const (
	ConnectionLookupURL    = "/api/v1/connection"
	ConnectionLookupFilter = "connection"
)

// normaliseSchema rewrites every `format: connection` property (with an
// optional `x-connection-types: [sql_server, …]`) into the commons-db
// connection picker — the `x-clicky-lookup` commons-db's query profiles
// declare, scoped to those types — whose committed value is the connection
// name, and refuses an enum with no or repeated options (a `{$cel: …}` enum
// computed from an earlier query can come back empty). It also records the
// authored property order as `x-order`, which a re-encoded schema would
// otherwise lose.
func normaliseSchema(schema json.RawMessage) (json.RawMessage, error) {
	if len(bytes.TrimSpace(schema)) == 0 {
		return nil, fmt.Errorf("schema is required")
	}
	var doc map[string]any
	if err := json.Unmarshal(schema, &doc); err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	if doc["type"] != "object" {
		return nil, fmt.Errorf("schema must be an object schema (type: object), got type %v", doc["type"])
	}
	order, err := propertyOrder(schema)
	if err != nil {
		return nil, err
	}
	if _, declared := doc["x-order"]; !declared && len(order) > 0 {
		doc["x-order"] = order
	}
	if err := normaliseProperties(doc, ""); err != nil {
		return nil, err
	}
	return json.Marshal(doc)
}

func normaliseProperties(doc map[string]any, path string) error {
	props, _ := doc["properties"].(map[string]any)
	for name, raw := range props {
		prop, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("schema property %s%s must be an object", path, name)
		}
		if err := normaliseConnectionProp(prop, path+name); err != nil {
			return err
		}
		if err := checkEnum(prop, path+name); err != nil {
			return err
		}
		if err := normaliseProperties(prop, path+name+"."); err != nil {
			return err
		}
	}
	// A conditional branch declares properties too: a `format: connection`
	// under `then:` is the same picker.
	for _, key := range []string{"then", "else"} {
		if branch, ok := doc[key].(map[string]any); ok {
			if err := normaliseProperties(branch, path); err != nil {
				return err
			}
		}
	}
	for _, key := range []string{"allOf", "anyOf", "oneOf"} {
		items, _ := doc[key].([]any)
		for _, item := range items {
			if branch, ok := item.(map[string]any); ok {
				if err := normaliseProperties(branch, path); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// checkEnum refuses an enum a person could not pick from: not a list, empty,
// or with an option repeated.
func checkEnum(prop map[string]any, name string) error {
	raw, declared := prop["enum"]
	if !declared {
		return nil
	}
	options, ok := raw.([]any)
	if !ok {
		return fmt.Errorf("schema property %s: enum must be a list, got %T(%v)", name, raw, raw)
	}
	if len(options) == 0 {
		return fmt.Errorf("schema property %s: enum has no options", name)
	}
	seen := map[string]bool{}
	for i, option := range options {
		switch option.(type) {
		case string, float64, bool, nil:
		default:
			return fmt.Errorf("schema property %s: enum option %d must be a string, number or boolean, got %T(%v)", name, i, option, option)
		}
		key := fmt.Sprintf("%T:%v", option, option)
		if seen[key] {
			return fmt.Errorf("schema property %s: enum option %d (%v) is repeated", name, i, option)
		}
		seen[key] = true
	}
	return nil
}

func normaliseConnectionProp(prop map[string]any, name string) error {
	types, hasTypes := prop["x-connection-types"]
	if prop["format"] != "connection" {
		if hasTypes {
			return fmt.Errorf("schema property %s: x-connection-types needs format: connection", name)
		}
		return nil
	}
	if _, ok := prop["x-clicky-lookup"]; ok {
		return fmt.Errorf("schema property %s: format: connection declares its own lookup; drop x-clicky-lookup", name)
	}
	if t, ok := prop["type"]; ok && t != "string" {
		return fmt.Errorf("schema property %s: a connection is its name, so type must be string, got %v", name, t)
	}
	url := ConnectionLookupURL
	if hasTypes {
		list, err := connectionTypes(types)
		if err != nil {
			return fmt.Errorf("schema property %s: %w", name, err)
		}
		url += "?types=" + strings.Join(list, ",")
	}
	prop["type"] = "string"
	prop["x-clicky-lookup"] = map[string]any{
		"url": url, "filter": ConnectionLookupFilter, "searchParam": "__lookup_q", "multi": false,
	}
	return nil
}

func connectionTypes(raw any) ([]string, error) {
	items, ok := raw.([]any)
	if !ok || len(items) == 0 {
		return nil, fmt.Errorf("x-connection-types must be a non-empty list of connection types, got %v", raw)
	}
	types := make([]string, len(items))
	for i, item := range items {
		s, ok := item.(string)
		if !ok || strings.TrimSpace(s) == "" {
			return nil, fmt.Errorf("x-connection-types[%d] must be a connection type name, got %v", i, item)
		}
		types[i] = s
	}
	return types, nil
}

// propertyOrder lists the top-level `properties` keys in authored order.
func propertyOrder(schema json.RawMessage) ([]string, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(schema, &top); err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}
	props, ok := top["properties"]
	if !ok {
		return nil, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(props))
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return nil, fmt.Errorf("schema properties must be an object")
	}
	var keys []string
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("schema properties: %w", err)
		}
		keys = append(keys, token.(string))
		var skip json.RawMessage
		if err := decoder.Decode(&skip); err != nil {
			return nil, fmt.Errorf("schema property %s: %w", token, err)
		}
	}
	return keys, nil
}
