package fixtures

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/flanksource/commons-db/connection"
	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/shell"
	flanksourceContext "github.com/flanksource/commons/context"
)

// StepConnectionCatalog is the connection catalog a step's `connection:` and a
// `format: connection` picker read: host first (when set), then gavel's layered
// store for workDir. kinds/dbconn registers it. A params-reading setup resolves
// its references through it, so a picked connection resolves to the one the
// picker offered.
var StepConnectionCatalog func(ctx context.Context, workDir string, host dbcontext.ConnectionResolver) (dbcontext.Context, error)

// paramsTemplate matches a Go template action reading the `params` root.
var paramsTemplate = regexp.MustCompile(`\{\{[^}]*\.params\b`)

// deferredSetup is a file's `setup:` that reads params or computes a value
// with `$cel`: it can only be prepared once the file's params form answered.
type deferredSetup struct {
	spec      *SetupSpec
	sourceDir string
}

// setupReadsParams reports whether a setup was kept as authored because it
// reads `{{.params.…}}` or computes a value, e.g.
// `connections: {kubernetes: {$cel: "params.cluster"}}`.
func setupReadsParams(spec *SetupSpec) bool {
	return spec.raw != nil
}

// isDeferredSetup reports whether an authored setup must wait for the params
// form: it templates the params root or holds a `{$cel: …}` value.
func isDeferredSetup(tree any, encoded []byte) bool {
	return paramsTemplate.Match(encoded) || containsCEL(tree)
}

func containsCEL(value any) bool {
	switch v := value.(type) {
	case orderedObject:
		for _, member := range v {
			if member.Key == CELValueKey || containsCEL(member.Value) {
				return true
			}
		}
	case []any:
		for _, item := range v {
			if containsCEL(item) {
				return true
			}
		}
	}
	return false
}

// withoutDeferredValues is the setup's shape with every value only the params
// form can supply — a `{$cel: …}` object or a templated string — as null, so
// the authored setup is still checked against shell.Setup when it is parsed.
func withoutDeferredValues(value any) any {
	switch v := value.(type) {
	case string:
		if strings.Contains(v, "{{") {
			return nil
		}
		return v
	case orderedObject:
		shape := orderedObject{}
		for _, member := range v {
			if member.Key == CELValueKey {
				return nil
			}
			shape = append(shape, orderedMember{Key: member.Key, Value: withoutDeferredValues(member.Value)})
		}
		return shape
	case []any:
		shape := make([]any, len(v))
		for i, item := range v {
			shape[i] = withoutDeferredValues(item)
		}
		return shape
	default:
		return value
	}
}

// connectionKinds are the json names of connection.ExecConnections' object
// connections (kubernetes, aws, gcp, azure, opensearch): the kinds a
// connection reference can name.
var connectionKinds = sync.OnceValue(func() map[string]bool {
	kinds := map[string]bool{}
	typ := reflect.TypeFor[connection.ExecConnections]()
	for i := range typ.NumField() {
		field := typ.Field(i)
		if field.Type.Kind() != reflect.Pointer || field.Type.Elem().Kind() != reflect.Struct {
			continue
		}
		if name, _, _ := strings.Cut(field.Tag.Get("json"), ","); name != "" && name != "-" {
			kinds[name] = true
		}
	}
	return kinds
})

// expandConnectionShorthand reads a string under `connections.<kind>` as that
// connection's reference, `{connection: <ref>}`: `kubernetes:
// connection://acme`, or a `{$cel: "params.cluster"}` that rendered one. A
// null leaves the connection unconfigured.
func expandConnectionShorthand(setup any) any {
	root, ok := setup.(orderedObject)
	if !ok {
		return setup
	}
	for i, member := range root {
		connections, ok := member.Value.(orderedObject)
		if member.Key != "connections" || !ok {
			continue
		}
		expanded := make(orderedObject, len(connections))
		for j, entry := range connections {
			expanded[j] = entry
			if ref, isString := entry.Value.(string); isString && connectionKinds()[entry.Key] {
				expanded[j].Value = orderedObject{{Key: "connection", Value: ref}}
			}
		}
		root[i].Value = expanded
	}
	return root
}

// decodeSetup strictly decodes an ordered setup tree into shell.Setup.
func decodeSetup(tree any) (shell.Setup, error) {
	var setup shell.Setup
	encoded, err := encodeOrderedJSON(tree)
	if err != nil {
		return setup, err
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	return setup, decoder.Decode(&setup)
}

// prepareParamsSetup renders the setup as authored — templates and `$cel`
// values — over the scope's step outputs (`params` among them) and prepares
// the result, resolving its connection references through host first, then
// the catalog.
func prepareParamsSetup(ctx flanksourceContext.Context, mdPath string, deferred deferredSetup, scope *StepScope, host dbcontext.ConnectionResolver) (*PreparedSetup, error) {
	tree, err := decodeOrderedJSON(deferred.spec.raw)
	if err != nil {
		return nil, fmt.Errorf("%s: setup: decode: %w", mdPath, err)
	}
	vars := stepIDRoots(scope)
	vars["env"] = scope.Env.Map()
	vars["steps"] = scope.StepOutputs()
	rendered, err := renderValue(tree, vars)
	if err != nil {
		return nil, fmt.Errorf("%s: setup: %w", mdPath, err)
	}
	setup, err := decodeSetup(expandConnectionShorthand(rendered))
	if err != nil {
		return nil, fmt.Errorf("%s: setup: decode rendered: %w", mdPath, err)
	}
	spec := SetupSpec{Setup: setup}
	// The host answers references ahead of gavel's layered store, as it does
	// for every step's connection, whenever the catalog is registered.
	if StepConnectionCatalog == nil {
		if setupReferencesConnection(&spec.Setup) {
			return nil, fmt.Errorf("%s: setup: no step connection catalog is registered (import fixtures/kinds/dbconn)", mdPath)
		}
		return prepareSetup(ctx, mdPath, deferred.sourceDir, &spec, nil)
	}
	catalog, err := StepConnectionCatalog(ctx, deferred.sourceDir, host)
	if err != nil {
		return nil, fmt.Errorf("%s: setup: open connection catalog: %w", mdPath, err)
	}
	return prepareSetup(ctx, mdPath, deferred.sourceDir, &spec, &catalog)
}
