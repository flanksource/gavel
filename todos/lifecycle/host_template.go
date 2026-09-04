package lifecycle

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/flanksource/captain/pkg/api"
)

// placeholder is a `{{subject.verification.document}}` reference inside a step
// spec: a dotted path into the variables the applicability predicates see.
var placeholder = regexp.MustCompile(`\{\{\s*([A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z_][A-Za-z0-9_]*)*)\s*\}\}`)

// expandSpec substitutes every placeholder in a step's spec from vars. A string
// that is exactly one placeholder takes the variable's value with its type — a
// document stays a document, a number a number — and a string embedding one is
// rendered with the value's text form. An unknown path is an error: a spec that
// referenced a fact the subject does not carry would otherwise run with the
// placeholder as a literal.
func expandSpec(spec *api.Spec, vars map[string]any) (api.Spec, error) {
	if spec == nil {
		return api.Spec{}, nil
	}
	raw, err := json.Marshal(spec)
	if err != nil {
		return api.Spec{}, fmt.Errorf("expand step spec: %w", err)
	}
	var tree any
	if err := json.Unmarshal(raw, &tree); err != nil {
		return api.Spec{}, fmt.Errorf("expand step spec: %w", err)
	}
	expanded, err := expandValue(tree, vars)
	if err != nil {
		return api.Spec{}, fmt.Errorf("expand step spec: %w", err)
	}
	data, err := json.Marshal(expanded)
	if err != nil {
		return api.Spec{}, fmt.Errorf("expand step spec: %w", err)
	}
	var out api.Spec
	if err := json.Unmarshal(data, &out); err != nil {
		return api.Spec{}, fmt.Errorf("expand step spec: %w", err)
	}
	return out, nil
}

func expandValue(value any, vars map[string]any) (any, error) {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			expanded, err := expandValue(item, vars)
			if err != nil {
				return nil, err
			}
			out[key] = expanded
		}
		return out, nil
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			expanded, err := expandValue(item, vars)
			if err != nil {
				return nil, err
			}
			out[i] = expanded
		}
		return out, nil
	case string:
		return expandString(typed, vars)
	default:
		return value, nil
	}
}

func expandString(s string, vars map[string]any) (any, error) {
	matches := placeholder.FindAllStringSubmatchIndex(s, -1)
	if len(matches) == 0 {
		return s, nil
	}
	if len(matches) == 1 && matches[0][0] == 0 && matches[0][1] == len(s) {
		return lookupPath(s[matches[0][2]:matches[0][3]], vars)
	}
	var out strings.Builder
	last := 0
	for _, match := range matches {
		out.WriteString(s[last:match[0]])
		value, err := lookupPath(s[match[2]:match[3]], vars)
		if err != nil {
			return nil, err
		}
		fmt.Fprint(&out, value)
		last = match[1]
	}
	out.WriteString(s[last:])
	return out.String(), nil
}

func lookupPath(path string, vars map[string]any) (any, error) {
	var current any = vars
	for _, segment := range strings.Split(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("placeholder {{%s}}: %q is not an object", path, segment)
		}
		next, ok := object[segment]
		if !ok {
			return nil, fmt.Errorf("placeholder {{%s}}: %q is not a known field", path, segment)
		}
		current = next
	}
	return current, nil
}
