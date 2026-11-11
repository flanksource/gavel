package jsonlint

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// EditorSchema resolves the same declarations as the JSON linter and bundles
// resources so browser forms do not need filesystem access or cross-origin fetches.
func EditorSchema(ctx context.Context, data []byte, path, override string) (map[string]any, error) {
	reference := override
	if reference == "" {
		var err error
		reference, err = yamlHeader(data)
		if err != nil {
			return nil, err
		}
		if reference == "" {
			var value any
			if err := yaml.Unmarshal(data, &value); err != nil {
				return nil, err
			}
			reference, err = schemaDeclaration(value)
			if err != nil {
				return nil, err
			}
		}
		if reference == "" {
			if reference, err = PathSchema(path); err != nil {
				return nil, err
			}
		}
	}
	if reference == "" {
		return nil, nil
	}
	location, err := ResolveSchema(reference, filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	if _, err := newValidator(ctx).compile(location, filepath.Dir(path)); err != nil {
		return nil, err
	}
	resources := map[string]any{}
	names := map[string]string{}
	var load func(string) (string, error)
	var rewrite func(any, *url.URL) (any, error)
	load = func(location string) (string, error) {
		if name, ok := names[location]; ok {
			return name, nil
		}
		name := fmt.Sprintf("resource%d", len(names))
		names[location] = name
		raw, err := (schemaLoader{ctx: ctx}).Load(location)
		if err != nil {
			return "", err
		}
		base, err := url.Parse(location)
		if err != nil {
			return "", err
		}
		if object, ok := raw.(map[string]any); ok {
			if id, ok := object["$id"].(string); ok {
				parsed, err := url.Parse(id)
				if err != nil {
					return "", err
				}
				names[base.ResolveReference(parsed).String()] = name
			}
		}
		node, err := rewrite(raw, base)
		if err != nil {
			return "", err
		}
		resources[name] = node
		return name, nil
	}
	rewrite = func(value any, base *url.URL) (any, error) {
		switch node := value.(type) {
		case map[string]any:
			if id, ok := node["$id"].(string); ok {
				ref, err := url.Parse(id)
				if err != nil {
					return nil, err
				}
				base = base.ResolveReference(ref)
			}
			delete(node, "$id")
			for key, child := range node {
				if key == "$ref" {
					ref, ok := child.(string)
					if !ok {
						return nil, fmt.Errorf("$ref must be a string")
					}
					parsed, err := url.Parse(ref)
					if err != nil {
						return nil, err
					}
					resolved := base.ResolveReference(parsed)
					fragment := resolved.Fragment
					resolved.Fragment = ""
					name, err := load(resolved.String())
					if err != nil {
						return nil, err
					}
					if fragment != "" && fragment[0] != '/' {
						return nil, fmt.Errorf("named schema anchors are not supported by the editor: %s", ref)
					}
					node[key] = "#/$defs/" + name + fragment
				} else {
					next, err := rewrite(child, base)
					if err != nil {
						return nil, err
					}
					node[key] = next
				}
			}
		case []any:
			for i, child := range node {
				next, err := rewrite(child, base)
				if err != nil {
					return nil, err
				}
				node[i] = next
			}
		}
		return value, nil
	}
	parsed, err := url.Parse(location)
	if err != nil {
		return nil, err
	}
	fragment := parsed.Fragment
	parsed.Fragment = ""
	name, err := load(parsed.String())
	if err != nil {
		return nil, err
	}
	return map[string]any{"$ref": "#/$defs/" + name + fragment, "$defs": resources}, nil
}
