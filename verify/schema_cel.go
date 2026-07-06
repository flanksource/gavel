package verify

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// DefaultSchemaCEL is the built-in schema.cel expression that produces the verify
// output schema. It is declarative — the schema shape lives here, not in Go — and
// is overridable per fixture (the AI-step `schema.cel` frontmatter). Go only
// supplies the map-building primitives CEL cannot express (the dynamic-keyed
// checks/ratings objects); everything else is assembled in the expression.
//
//go:embed schema.cel
var DefaultSchemaCEL string

// schemaEnv is the CEL environment for schema.cel: it exposes the parsed fixture
// outline (criteria, checks, rating_dimensions, issue_aware) plus helper
// functions that build the JSON-schema fragments CEL cannot express natively.
var schemaEnv = newSchemaEnv()

func newSchemaEnv() *cel.Env {
	mapType := cel.MapType(cel.StringType, cel.DynType)
	env, err := cel.NewEnv(
		cel.Variable("criteria", cel.ListType(cel.StringType)),
		cel.Variable("checks", cel.ListType(mapType)),
		cel.Variable("rating_dimensions", cel.ListType(cel.StringType)),
		cel.Variable("issue_aware", cel.BoolType),
		cel.Function("check_schema", cel.Overload("check_schema_list",
			[]*cel.Type{cel.ListType(mapType)}, mapType, cel.UnaryBinding(celCheckSchema))),
		cel.Function("rating_schema", cel.Overload("rating_schema_list",
			[]*cel.Type{cel.ListType(cel.StringType)}, mapType, cel.UnaryBinding(celRatingSchema))),
		cel.Function("criteria_schema", cel.Overload("criteria_schema_list",
			[]*cel.Type{cel.ListType(cel.StringType)}, mapType, cel.UnaryBinding(celCriteriaSchema))),
		cel.Function("evidence_schema", cel.Overload("evidence_schema_nullary",
			[]*cel.Type{}, mapType, cel.FunctionBinding(celEvidenceSchema))),
		cel.Function("merge", cel.Overload("merge_maps",
			[]*cel.Type{mapType, mapType}, mapType, cel.BinaryBinding(celMerge))),
	)
	if err != nil {
		panic(fmt.Sprintf("schema CEL env: %v", err))
	}
	return env
}

// EvalSchema evaluates a schema.cel expression against the review's parsed
// outline and returns the JSON output schema. An empty expr uses DefaultSchemaCEL.
// It replaces the old imperative schema construction: the shape is declared in
// CEL, and the dynamic-keyed map fragments come from the registered helpers.
func EvalSchema(expr string, checks []Check, issueAware bool, criteria []string) (string, error) {
	if strings.TrimSpace(expr) == "" {
		expr = DefaultSchemaCEL
	}
	ast, iss := schemaEnv.Compile(expr)
	if iss != nil && iss.Err() != nil {
		return "", fmt.Errorf("schema.cel compile: %w", iss.Err())
	}
	prg, err := schemaEnv.Program(ast)
	if err != nil {
		return "", fmt.Errorf("schema.cel program: %w", err)
	}
	out, _, err := prg.Eval(map[string]any{
		"criteria":          criteria,
		"checks":            checksToActivation(checks),
		"rating_dimensions": RatingDimensions,
		"issue_aware":       issueAware,
	})
	if err != nil {
		return "", fmt.Errorf("schema.cel eval: %w", err)
	}
	native, err := out.ConvertToNative(reflect.TypeOf(map[string]any{}))
	if err != nil {
		return "", fmt.Errorf("schema.cel result is not an object: %w", err)
	}
	b, err := json.MarshalIndent(normalizeJSON(native), "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// normalizeJSON makes a CEL-converted value json.Marshal-safe: nested dynamic
// maps come back as map[any]any (non-string keys are illegal for encoding/json),
// so recursively re-key them to map[string]any.
func normalizeJSON(v any) any {
	switch t := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, val := range t {
			m[k] = normalizeJSON(val)
		}
		return m
	case map[any]any:
		m := make(map[string]any, len(t))
		for k, val := range t {
			m[fmt.Sprint(k)] = normalizeJSON(val)
		}
		return m
	case []any:
		s := make([]any, len(t))
		for i, val := range t {
			s[i] = normalizeJSON(val)
		}
		return s
	default:
		return v
	}
}

func checksToActivation(checks []Check) []map[string]any {
	out := make([]map[string]any, len(checks))
	for i, c := range checks {
		out[i] = map[string]any{"id": c.ID, "description": c.Description}
	}
	return out
}

// --- CEL helper bindings ----------------------------------------------------

func celCheckSchema(arg ref.Val) ref.Val {
	checks, err := nativeMaps(arg)
	if err != nil {
		return types.NewErr("check_schema: %v", err)
	}
	return types.DefaultTypeAdapter.NativeToValue(checksObjectSchema(checks))
}

func celRatingSchema(arg ref.Val) ref.Val {
	dims, err := nativeStrings(arg)
	if err != nil {
		return types.NewErr("rating_schema: %v", err)
	}
	return types.DefaultTypeAdapter.NativeToValue(ratingsObjectSchema(dims))
}

func celCriteriaSchema(arg ref.Val) ref.Val {
	criteria, err := nativeStrings(arg)
	if err != nil {
		return types.NewErr("criteria_schema: %v", err)
	}
	return types.DefaultTypeAdapter.NativeToValue(criteriaArraySchema(criteria))
}

func celEvidenceSchema(...ref.Val) ref.Val {
	return types.DefaultTypeAdapter.NativeToValue(evidenceSchema())
}

func celMerge(a, b ref.Val) ref.Val {
	am, err := nativeStringMap(a)
	if err != nil {
		return types.NewErr("merge: %v", err)
	}
	bm, err := nativeStringMap(b)
	if err != nil {
		return types.NewErr("merge: %v", err)
	}
	out := make(map[string]any, len(am)+len(bm))
	for k, v := range am {
		out[k] = v
	}
	for k, v := range bm {
		out[k] = v
	}
	return types.DefaultTypeAdapter.NativeToValue(out)
}

func nativeMaps(arg ref.Val) ([]map[string]any, error) {
	v, err := arg.ConvertToNative(reflect.TypeOf([]map[string]any{}))
	if err != nil {
		return nil, err
	}
	return v.([]map[string]any), nil
}

func nativeStrings(arg ref.Val) ([]string, error) {
	v, err := arg.ConvertToNative(reflect.TypeOf([]string{}))
	if err != nil {
		return nil, err
	}
	return v.([]string), nil
}

func nativeStringMap(arg ref.Val) (map[string]any, error) {
	v, err := arg.ConvertToNative(reflect.TypeOf(map[string]any{}))
	if err != nil {
		return nil, err
	}
	return v.(map[string]any), nil
}
