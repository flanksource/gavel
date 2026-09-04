package labels

import "sort"

// Resolver resolves raw label strings against one fixed definition set.
//
// Construct exactly one per request and reuse it for every todo in the
// response. Resolving a label by querying the database is the N+1 that made
// /api/projects take 46 seconds; the whole definition set is a single indexed
// read, so it is loaded once and applied in memory.
type Resolver struct {
	workspace map[string]Definition
	global    map[string]Definition
}

// NewResolver indexes workspace definitions over global ones. Precedence is
// fixed by tier, not by argument order or slice order.
func NewResolver(workspace, global Definitions) *Resolver {
	return &Resolver{
		workspace: index(workspace, ScopeWorkspace),
		global:    index(global, ScopeGlobal),
	}
}

func index(defs Definitions, scope Scope) map[string]Definition {
	out := make(map[string]Definition, len(defs))
	for _, def := range defs {
		def.Name = Normalize(def.Name)
		if def.Name == "" {
			continue
		}
		def.Scope = scope
		out[def.Name] = def
	}
	return out
}

// Resolve returns a presentation for any label and never returns a zero value.
// The chain, first hit wins:
//
//  1. workspace row, exact name
//  2. global row, exact name
//  3. built-in default, exact name
//  4. workspace row, namespace key
//  5. global row, namespace key
//  6. built-in default, namespace key
//  7. Hash(label) -> palette hue, no icon, ScopeDerived
//
// Tiers 4-6 keep the full label as Name and record the key in MatchedKey, so a
// caller can explain why "source:todo" is coloured like "source".
func (r *Resolver) Resolve(label string) Definition {
	name := Normalize(label)

	if def, ok := r.lookup(name); ok {
		def.Name = name
		return def
	}

	if key := Key(name); key != "" {
		if def, ok := r.lookup(key); ok {
			def.Name = name
			def.MatchedKey = key
			return def
		}
	}

	return Definition{Name: name, Color: Hash(name), Scope: ScopeDerived}
}

// lookup walks the three stored tiers for one exact name.
func (r *Resolver) lookup(name string) (Definition, bool) {
	if name == "" {
		return Definition{}, false
	}
	if def, ok := r.workspace[name]; ok {
		return def, true
	}
	if def, ok := r.global[name]; ok {
		return def, true
	}
	if def, ok := builtins[name]; ok {
		return def, true
	}
	return Definition{}, false
}

// ResolveAll resolves a todo's labels, preserving their order.
func (r *Resolver) ResolveAll(labels []string) Definitions {
	if len(labels) == 0 {
		return nil
	}
	out := make(Definitions, 0, len(labels))
	for _, label := range labels {
		if Normalize(label) == "" {
			continue
		}
		out = append(out, r.Resolve(label))
	}
	return out
}

// All returns every stored and built-in definition after shadowing, sorted by
// name. It is the payload of the definitions endpoint and the rows of
// `gavel todos labels list`. Derived presentations are absent by construction —
// nothing is stored for them, so there is nothing to list or edit.
func (r *Resolver) All() Definitions {
	merged := make(map[string]Definition, len(builtins)+len(r.global)+len(r.workspace))
	for name, def := range builtins {
		merged[name] = def
	}
	for name, def := range r.global {
		merged[name] = def
	}
	for name, def := range r.workspace {
		merged[name] = def
	}

	out := make(Definitions, 0, len(merged))
	for _, def := range merged {
		out = append(out, def)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Derive resolves labels with no stored definitions at all — the fallback used
// by a TODO from a source that has no definition store, so every surface still
// renders a coloured chip.
func Derive(labels []string) Definitions {
	return (&Resolver{}).ResolveAll(labels)
}
