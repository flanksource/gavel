package fixtures

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/goccy/go-yaml"

	"github.com/flanksource/commons-db/tracing"

	"github.com/flanksource/gavel/traces"
)

// TraceScope is how long one armed trace runs: around one fixture (test), the
// fixture file's steps (file), or the whole run (run).
type TraceScope string

const (
	TraceScopeTest TraceScope = "test"
	TraceScopeFile TraceScope = "file"
	TraceScopeRun  TraceScope = "run"
)

// TraceSpec is one `traces:` entry: a commons-db trace kind the run's trace
// host serves, its params, how long it runs, and an optional CEL requirement
// over the kind's summary of the fixture's window.
type TraceSpec struct {
	Kind    string          `json:"kind" yaml:"kind"`
	Params  json.RawMessage `json:"params,omitempty" yaml:"params,omitempty"`
	Scope   TraceScope      `json:"scope,omitempty" yaml:"scope,omitempty"`
	Require string          `json:"require,omitempty" yaml:"require,omitempty"`
}

// templated reports whether the params hold a `{{…}}` template, rendered
// against the fixture's scope when a test-scope capture starts.
func (spec TraceSpec) templated() bool {
	return bytes.Contains(spec.Params, []byte("{{"))
}

// rendered is spec with every string of its params rendered as a Go template
// over vars, the way a step's config is (renderConfig): `{{.env.X}}`,
// `{{.steps.<id>.…}}`. Params without a template are returned as they are.
func (spec TraceSpec) rendered(vars map[string]any) (TraceSpec, error) {
	if !spec.templated() {
		return spec, nil
	}
	params, err := decodeOrderedJSON(spec.Params)
	if err != nil {
		return spec, fmt.Errorf("trace %s params: %w", spec.Kind, err)
	}
	if params, err = renderValue(params, vars); err != nil {
		return spec, fmt.Errorf("trace %s params: %w", spec.Kind, err)
	}
	if spec.Params, err = encodeOrderedJSON(params); err != nil {
		return spec, fmt.Errorf("trace %s params: %w", spec.Kind, err)
	}
	return spec, nil
}

// TraceSpecs is a fixture's `traces:`. A nil *TraceSpecs declares nothing and
// leaves the fixture to the run-wide --traces default; a non-nil empty one
// (`traces: []`, or `none` in the compact form) opts out of it.
type TraceSpecs []TraceSpec

// traceCatalog is the kinds every gavel trace host serves, which is what a
// declaration is validated against before anything runs.
var traceCatalog = sync.OnceValues(traces.Kinds)

// ParseTraces reads the compact form a table column and --traces take: kinds
// separated by commas or spaces, each with its default params and test scope.
// An empty value declares nothing (nil); "none" opts out (empty, non-nil).
func ParseTraces(value string) (*TraceSpecs, error) {
	fields := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' })
	if len(fields) == 0 {
		return nil, nil
	}
	if len(fields) == 1 && strings.EqualFold(fields[0], "none") {
		return &TraceSpecs{}, nil
	}
	specs := make(TraceSpecs, 0, len(fields))
	for _, field := range fields {
		specs = append(specs, TraceSpec{Kind: field})
	}
	return &specs, specs.normalize()
}

// UnmarshalYAML accepts the compact string, a list of kinds, a list of full
// entries, or any mix of the last two.
func (s *TraceSpecs) UnmarshalYAML(data []byte) error {
	var raw any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("traces: %w", err)
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("traces: %w", err)
	}
	return s.UnmarshalJSON(encoded)
}

// UnmarshalJSON accepts what UnmarshalYAML does, as JSON.
func (s *TraceSpecs) UnmarshalJSON(data []byte) error {
	var compact string
	if err := json.Unmarshal(data, &compact); err == nil {
		parsed, err := ParseTraces(compact)
		if err != nil {
			return err
		}
		if parsed == nil {
			return fmt.Errorf("traces: an empty string declares no trace; use [] to opt out")
		}
		*s = *parsed
		return nil
	}
	var entries []json.RawMessage
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("traces: want a kind list or a list of {kind, params, scope, require}, got %s", data)
	}
	specs := make(TraceSpecs, 0, len(entries))
	for _, entry := range entries {
		spec, err := decodeTraceEntry(entry)
		if err != nil {
			return err
		}
		specs = append(specs, spec)
	}
	*s = specs
	return specs.normalize()
}

func decodeTraceEntry(entry json.RawMessage) (TraceSpec, error) {
	var kind string
	if err := json.Unmarshal(entry, &kind); err == nil {
		return TraceSpec{Kind: kind}, nil
	}
	var spec TraceSpec
	strict := json.NewDecoder(bytes.NewReader(entry))
	strict.DisallowUnknownFields()
	if err := strict.Decode(&spec); err != nil {
		return spec, fmt.Errorf("traces: %w", err)
	}
	return spec, nil
}

// normalize defaults every scope to test and validates the entries.
func (s TraceSpecs) normalize() error {
	for i := range s {
		if s[i].Scope == "" {
			s[i].Scope = TraceScopeTest
		}
	}
	return s.Validate()
}

// Validate refuses an unknown kind (naming the registered ones), params the
// kind refuses, an unknown scope, a terminal not scoped to its test (only one
// PTY feeds it), and one kind declared twice (its CEL root would be ambiguous).
func (s TraceSpecs) Validate() error {
	kinds, err := traceCatalog()
	if err != nil {
		return fmt.Errorf("traces: %w", err)
	}
	seen := map[string]bool{}
	for _, spec := range s {
		if err := spec.validate(kinds); err != nil {
			return err
		}
		if seen[spec.Kind] {
			return fmt.Errorf("traces: %s is declared twice; a fixture arms each kind once", spec.Kind)
		}
		seen[spec.Kind] = true
	}
	return nil
}

func (spec TraceSpec) validate(kinds *tracing.Kinds) error {
	switch spec.Scope {
	case TraceScopeTest, TraceScopeFile, TraceScopeRun:
	default:
		return fmt.Errorf("traces: %s scope %q is not one of test, file, run", spec.Kind, spec.Scope)
	}
	if spec.templated() {
		// Templated params are validated once rendered, when the capture starts.
		if _, err := traces.Lookup(kinds, spec.Kind); err != nil {
			return fmt.Errorf("traces: %w", err)
		}
		if spec.Scope != TraceScopeTest {
			return fmt.Errorf("traces: %s params are templated, so its scope must be test: a %s-scope capture starts before any step has run", spec.Kind, spec.Scope)
		}
	} else if err := traces.Validate(kinds, spec.Kind, spec.Params); err != nil {
		return fmt.Errorf("traces: %w", err)
	}
	if spec.Kind == terminalTraceKind && spec.Scope != TraceScopeTest {
		return fmt.Errorf("traces: terminal records one fixture's PTY, so its scope must be test, not %s", spec.Scope)
	}
	return nil
}

// TracesSchema is the JSON schema of a `traces:` declaration: the compact
// string, or a list of kinds and {kind, params, scope, require} entries whose
// params are validated against the kind's own params schema.
func TracesSchema() (map[string]any, error) {
	kinds, err := traceCatalog()
	if err != nil {
		return nil, err
	}
	infos, err := kinds.List()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(infos))
	byKind := make([]any, 0, len(infos))
	for _, info := range infos {
		names = append(names, info.Name)
		byKind = append(byKind, map[string]any{
			"if":   map[string]any{"properties": map[string]any{"kind": map[string]any{"const": info.Name}}, "required": []string{"kind"}},
			"then": map[string]any{"properties": map[string]any{"params": info.Params}},
		})
	}
	entry := map[string]any{
		"type": "object", "additionalProperties": false, "required": []string{"kind"}, "allOf": byKind,
		"properties": map[string]any{
			"kind":    map[string]any{"enum": names, "description": "A trace kind the run's trace host serves."},
			"params":  map[string]any{"type": "object", "description": "The kind's params; omitted keys take the kind's defaults."},
			"scope":   map[string]any{"enum": []string{string(TraceScopeTest), string(TraceScopeFile), string(TraceScopeRun)}, "default": string(TraceScopeTest), "description": "test: around this fixture; file: the file's steps share one capture, windowed per fixture; run: the whole run shares one."},
			"require": map[string]any{"type": "string", "description": "CEL over the kind's summary of the fixture's window (its keys, and the root `trace`); false fails a fixture that otherwise passed.", "examples": []string{"count > 0"}},
		},
	}
	return map[string]any{
		"description": "Traces armed around the fixture; each kind's summary of the fixture's window is the CEL root named by the kind. A comma-separated list of kinds is the compact form, and [] (or none) opts out of a --traces default.",
		"anyOf": []any{
			map[string]any{"type": "string", "examples": []string{"http_proxy", "http_proxy,terminal", "none"}},
			map[string]any{"type": "array", "items": map[string]any{"anyOf": []any{map[string]any{"enum": names}, entry}}},
		},
	}, nil
}

// key identifies the capture a file or run scope shares between fixtures: the
// same kind with the same params.
func (spec TraceSpec) key() string {
	return string(spec.Scope) + "/" + spec.Kind + "/" + string(spec.Params)
}
