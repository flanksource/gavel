package form

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/kinds/dbconn"
	"github.com/flanksource/gavel/fixtures/kinds/lookup"
)

// QueryLookupFilter names an `x-query` field's lookup.
const QueryLookupFilter = "query"

// lowering is one form step's schema lowering: the fixture connections it
// resolves `x-query` and connection pickers through (opened on first use), the
// lookup bindings it registered for the step's lifetime, and the connection
// defaults it prefills the prompt with.
type lowering struct {
	ctx     context.Context
	step    *fixtures.StepContext
	conns   *dbconn.Connections
	tokens  []string
	prefill []prefill
}

// prefill is a value the prompt starts from at path when nothing set it.
type prefill struct {
	path  []string
	value string
}

func (l *lowering) connections() (*dbconn.Connections, error) {
	if l.conns != nil {
		return l.conns, nil
	}
	conns, err := dbconn.StepConnections(l.ctx, l.step, dbconn.Options{})
	if err != nil {
		return nil, fmt.Errorf("fixture connections: %w", err)
	}
	l.conns = conns
	return conns, nil
}

// lookupSpec registers binding for the field at path and is its
// `x-clicky-lookup`.
func (l *lowering) lookupSpec(path []string, filter string, binding lookup.Binding) (map[string]any, error) {
	pointer := lookup.Pointer(path)
	token, err := lookup.Register(pointer, binding)
	if err != nil {
		return nil, err
	}
	l.tokens = append(l.tokens, token)
	return map[string]any{"url": lookup.URL(token, pointer), "filter": filter, "searchParam": "__lookup_q", "multi": false}, nil
}

// release drops every binding the step registered.
func (l *lowering) release() {
	for _, token := range l.tokens {
		lookup.Release(token)
	}
	l.tokens = nil
}

// connectionLookup is a `format: connection` picker's lookup in a fixture
// that declares `connections:` — its named connections listed beside the
// catalog's — and, for a picker of one type, the type's declared default as
// its prefill. Without a declaration it is nil: the global catalog lookup.
func (l *lowering) connectionLookup(prop map[string]any, path, types []string) (map[string]any, error) {
	if l.step.Test.Connections == nil {
		return nil, nil
	}
	conns, err := l.connections()
	if err != nil {
		return nil, err
	}
	if _, explicit := prop["default"]; !explicit && len(types) == 1 {
		if resolved, ok := conns.DefaultFor(types[0]); ok {
			l.prefill = append(l.prefill, prefill{path: path, value: resolved.Ref})
		}
	}
	return l.lookupSpec(path, ConnectionLookupFilter, lookup.ConnectionBinding{Types: types, Connections: conns, WorkDir: l.step.WorkDir})
}

// lowerQuery rewrites an `x-query: {connection?, sql, columns?}` property into
// a lookup bound to its SQL. connection names a sibling `format: connection`
// property (the lookup is scoped to its value), else a named or catalog
// connection resolved now; omitted, the fixture's single default.
func (l *lowering) lowerQuery(prop map[string]any, path []string, siblings map[string]map[string]any) error {
	raw, ok := prop["x-query"]
	if !ok {
		return nil
	}
	delete(prop, "x-query")
	name := strings.Join(path, ".")
	binding, err := queryBinding(raw, prop)
	if err != nil {
		return fmt.Errorf("schema property %s: %w", name, err)
	}
	conns, err := l.connections()
	if err != nil {
		return fmt.Errorf("schema property %s: x-query: %w", name, err)
	}
	from, err := l.bindConnection(&binding, conns, raw.(map[string]any)["connection"], siblings)
	if err != nil {
		return fmt.Errorf("schema property %s: x-query.connection: %w", name, err)
	}
	if from != "" {
		binding.From = append(slices.Clone(path[:len(path)-1]), from)
	}
	spec, err := l.lookupSpec(path, QueryLookupFilter, binding)
	if err != nil {
		return err
	}
	if from != "" {
		spec["scope"] = map[string]any{"param": "connection", "from": from}
	}
	prop["type"] = "string"
	prop["x-clicky-lookup"] = spec
	return nil
}

// bindConnection points binding at its connection, returning the sibling
// property name when it is linked to one.
func (l *lowering) bindConnection(binding *lookup.QueryBinding, conns *dbconn.Connections, ref any, siblings map[string]map[string]any) (string, error) {
	if ref == nil {
		resolved, err := conns.SingleDefault()
		binding.Connection = resolved
		return "", err
	}
	name, ok := ref.(string)
	if !ok || strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("must name a sibling connection property or a connection, got %v", ref)
	}
	if sibling, linked := siblings[name]; linked {
		if sibling["format"] != "connection" {
			return "", fmt.Errorf("sibling property %s is not format: connection", name)
		}
		binding.Connections = conns
		return name, nil
	}
	resolved, err := conns.Resolve(name)
	binding.Connection = resolved
	return "", err
}

// queryBinding reads an x-query declaration on prop, which must otherwise be
// a plain string property.
func queryBinding(raw any, prop map[string]any) (lookup.QueryBinding, error) {
	spec, ok := raw.(map[string]any)
	if !ok {
		return lookup.QueryBinding{}, fmt.Errorf("x-query must be an object {connection?, sql, columns?}, got %T", raw)
	}
	for key := range spec {
		if key != "connection" && key != "sql" && key != "columns" {
			return lookup.QueryBinding{}, fmt.Errorf("x-query: unknown key %q (connection, sql, columns)", key)
		}
	}
	for _, conflict := range []string{"x-clicky-lookup", "enum", "format"} {
		if _, declared := prop[conflict]; declared {
			return lookup.QueryBinding{}, fmt.Errorf("x-query lists the options itself; drop %s", conflict)
		}
	}
	if t, ok := prop["type"]; ok && t != "string" {
		return lookup.QueryBinding{}, fmt.Errorf("an x-query value is a string, got type %v", t)
	}
	sql, _ := spec["sql"].(string)
	if strings.TrimSpace(sql) == "" {
		return lookup.QueryBinding{}, fmt.Errorf("x-query.sql is required, got %v", spec["sql"])
	}
	columns, err := queryColumns(spec["columns"])
	return lookup.QueryBinding{SQL: sql, Columns: columns}, err
}

func queryColumns(raw any) (map[string]string, error) {
	if raw == nil {
		return nil, nil
	}
	object, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("x-query.columns must map option fields to columns, got %T", raw)
	}
	columns := make(map[string]string, len(object))
	for role, column := range object {
		name, ok := column.(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("x-query.columns.%s must be a column name, got %v", role, column)
		}
		columns[role] = name
	}
	return columns, nil
}

// withPrefill is values with each connection default set where nothing is.
func (l *lowering) withPrefill(values map[string]any) map[string]any {
	out := maps.Clone(values)
	if out == nil {
		out = map[string]any{}
	}
	for _, p := range l.prefill {
		object := out
		for _, key := range p.path[:len(p.path)-1] {
			next, ok := object[key].(map[string]any)
			if !ok {
				next = map[string]any{}
				object[key] = next
			}
			object = next
		}
		if _, set := object[p.path[len(p.path)-1]]; !set {
			object[p.path[len(p.path)-1]] = p.value
		}
	}
	return out
}

// scopeProps is every property an object schema declares, its conditional
// and combinator branches' included: the siblings an x-query may link to.
func scopeProps(doc map[string]any) map[string]map[string]any {
	scope := map[string]map[string]any{}
	props, _ := doc["properties"].(map[string]any)
	for name, raw := range props {
		if prop, ok := raw.(map[string]any); ok {
			scope[name] = prop
		}
	}
	for _, branch := range branches(doc) {
		maps.Copy(scope, scopeProps(branch))
	}
	return scope
}

// branches are doc's then/else and allOf/anyOf/oneOf members.
func branches(doc map[string]any) []map[string]any {
	var out []map[string]any
	for _, key := range []string{"then", "else"} {
		if branch, ok := doc[key].(map[string]any); ok {
			out = append(out, branch)
		}
	}
	for _, key := range []string{"allOf", "anyOf", "oneOf"} {
		items, _ := doc[key].([]any)
		for _, item := range items {
			if branch, ok := item.(map[string]any); ok {
				out = append(out, branch)
			}
		}
	}
	return out
}
