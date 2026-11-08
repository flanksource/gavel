package form

import (
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/flanksource/clicky/schemaform"
)

// formSchema is the part of a form's JSON Schema the values are shaped by:
// its properties, and the `allOf` if/then/else branches that declare
// properties shown only while their condition holds. Which branch holds, and
// the defaults it fills, are decided by clicky schemaform, the same rules the
// terminal and web forms render with.
type formSchema struct {
	Properties map[string]map[string]any `json:"properties"`
	Required   []string                  `json:"required"`
	AllOf      []formBranch              `json:"allOf,omitempty"`
	parsed     *schemaform.Schema
}

// formBranch is one `allOf` member: unconditional properties, or a `then` and
// `else` chosen by its `if`.
type formBranch struct {
	Properties map[string]map[string]any `json:"properties"`
	Then       *formBranch               `json:"then,omitempty"`
	Else       *formBranch               `json:"else,omitempty"`
}

// declared is every property the schema or any of its branches declares.
func (s formSchema) declared() map[string]map[string]any {
	props := maps.Clone(s.Properties)
	if props == nil {
		props = map[string]map[string]any{}
	}
	for _, clause := range s.AllOf {
		for _, branch := range []*formBranch{&clause, clause.Then, clause.Else} {
			if branch != nil {
				maps.Copy(props, branch.Properties)
			}
		}
	}
	return props
}

// shown is the schema's effective properties for values and the keys they
// make required.
func (s formSchema) shown(values map[string]any) schemaform.Properties {
	return schemaform.Effective(s.parsed, values)
}

// prune drops the values of properties only an inactive branch declares: a
// hidden field is absent, not a stale answer.
func (s formSchema) prune(values map[string]any) map[string]any {
	shown := s.shown(values).Map
	for key := range s.declared() {
		if _, ok := shown[key]; !ok {
			delete(values, key)
		}
	}
	return values
}

func parseFormSchema(schema json.RawMessage) (formSchema, error) {
	var doc formSchema
	if err := json.Unmarshal(schema, &doc); err != nil {
		return doc, fmt.Errorf("schema: %w", err)
	}
	parsed, err := schemaform.Parse(schema)
	if err != nil {
		return doc, fmt.Errorf("schema: %w", err)
	}
	doc.parsed = parsed
	return doc, nil
}

// prefill builds the values a form starts from: supplied over the step's
// `default:`, strings typed by the schema (a `--param n=5` or a rendered
// `{{…}}` default is an integer when n is one), then the schema's own defaults
// for what is still unset — a branch's once the values selecting it are set.
func (s formSchema) prefill(defaults, supplied map[string]any) (map[string]any, error) {
	values := map[string]any{}
	props := s.declared()
	for _, layer := range []map[string]any{defaults, supplied} {
		for key, value := range layer {
			prop, declared := props[key]
			if !declared && len(props) > 0 {
				return nil, fmt.Errorf("unknown value %q (declared: %s)", key, strings.Join(slices.Sorted(maps.Keys(props)), ", "))
			}
			typed, err := coerce(prop, value)
			if err != nil {
				return nil, fmt.Errorf("value %q: %w", key, err)
			}
			values[key] = typed
		}
	}
	return s.prune(s.integers(schemaform.ApplyDefaults(s.parsed, values))), nil
}


// coerce types a string value by its property's type; any other value is
// already typed (a --params file, a UI start request).
func coerce(prop map[string]any, value any) (any, error) {
	text, ok := value.(string)
	if !ok {
		return value, nil
	}
	switch prop["type"] {
	case "integer":
		return strconv.ParseInt(strings.TrimSpace(text), 10, 64)
	case "number":
		return strconv.ParseFloat(strings.TrimSpace(text), 64)
	case "boolean":
		return strconv.ParseBool(strings.TrimSpace(text))
	case "array":
		items := strings.Split(text, ",")
		out := make([]any, len(items))
		for i, item := range items {
			typed, err := coerce(asMap(prop["items"]), strings.TrimSpace(item))
			if err != nil {
				return nil, fmt.Errorf("item %d: %w", i, err)
			}
			out[i] = typed
		}
		return out, nil
	case "object":
		var object map[string]any
		if err := json.Unmarshal([]byte(text), &object); err != nil {
			return nil, fmt.Errorf("want a JSON object: %w", err)
		}
		return object, nil
	}
	return text, nil
}

func asMap(value any) map[string]any {
	m, _ := value.(map[string]any)
	return m
}

// integers turns the whole float64 values JSON decodes into int64 for integer
// properties, so CEL compares them with the integers other steps produce.
func (s formSchema) integers(values map[string]any) map[string]any {
	for key, prop := range s.declared() {
		if f, ok := values[key].(float64); ok && prop["type"] == "integer" && f == math.Trunc(f) {
			values[key] = int64(f)
		}
	}
	return values
}

// missing lists the required keys values leaves unset, the active branches'
// included.
func (s formSchema) missing(values map[string]any) []string {
	var keys []string
	for _, key := range s.shown(values).Required {
		if _, ok := values[key]; !ok {
			keys = append(keys, key)
		}
	}
	return keys
}
