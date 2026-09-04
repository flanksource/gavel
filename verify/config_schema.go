package verify

import (
	"encoding/json"

	"github.com/flanksource/captain/pkg/api"
)

// ConfigSchemaID is the canonical URL editors fetch to validate .gavel.yaml.
// Reference it from a file with a leading comment:
//
//	# yaml-language-server: $schema=https://raw.githubusercontent.com/flanksource/gavel/main/gavel.schema.json
const ConfigSchemaID = "https://raw.githubusercontent.com/flanksource/gavel/main/gavel.schema.json"

// ConfigJSONSchema renders the documented JSON Schema for the .gavel.yaml
// (a.k.a. .gavel.yml) configuration file. It is the single source of truth for
// the committed gavel.schema.json artifact and SCHEMA.md; regenerate the JSON
// with `go generate .` after changing GavelConfig.
func ConfigJSONSchema() (string, error) {
	b, err := json.MarshalIndent(gavelConfigSchema(), "", "  ")
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil
}

func gavelConfigSchema() map[string]any {
	specSchema := captainSpecSchema()
	schema := object(
		"Root configuration for Gavel. Place .gavel.yaml (or .gavel.yml) in ~/, the git root, "+
			"or the target directory; layers merge in that order with later layers overriding earlier ones. "+
			"Run `gavel config [path]` to inspect the merged result.",
		map[string]any{
			"ai":       aiSchema(specSchema),
			"lint":     lintSchema(),
			"commit":   commitSchema(),
			"fixtures": fixturesSchema(),
			"ssh":      sshSchema(),
			"pre":      hookStepsSchema("Top-level hooks run before the main test/lint pipeline, in declaration order. Appended across layers."),
			"post":     hookStepsSchema("Top-level hooks run after the main pipeline as non-blocking cleanup/reporting. Appended across layers."),
			"secrets":  secretsSchema(),
			"procfile": procfileSchema(),
			"checks":   checksSchema(),
			"todos":    todosSchema(specSchema),
			"status":   statusSchema(),
			"test":     testSchema(),
			"pr":       prSchema(),
		},
	)
	schema["$schema"] = "https://json-schema.org/draft/2020-12/schema"
	schema["$id"] = ConfigSchemaID
	schema["title"] = "Gavel configuration (.gavel.yaml)"
	patchSetupDefs(specSchema)
	schema["$defs"] = specSchema["$defs"]
	return schema
}

// patchSetupDefs documents the commons-db setup types that api.Spec reflects
// into $defs. Reflection sees Go strings, so mode/uncommitted/ignored all arrive
// as bare {"type":"string"} — editor completion offers nothing and a typo is
// only caught at run time. This stamps on the values the Go constants already
// define, plus the defaults the runtime actually applies, so the schema, the
// docs and shell.Worktree.ApplyDefaults agree. It is the gavel-local stand-in
// until commons-db carries jsonschema: tags of its own, and stays idempotent if
// they land later.
func patchSetupDefs(specSchema map[string]any) {
	defs, ok := specSchema["$defs"].(map[string]any)
	if !ok {
		panic("captain spec schema has no $defs")
	}

	patchDefProp(defs, "Checkout", "mode", map[string]any{
		"description": "Checkout source. Inferred from url/path when unset: url means remote, path means local.",
		"enum":        []any{"none", "local", "remote"},
	})
	patchDefProp(defs, "Checkout", "since", map[string]any{
		"description": "Commit-ish whose merge-base diff against HEAD is folded into the reported changed " +
			"files. Informational only — it has no bearing on what the worktree contains.",
	})
	patchDefProp(defs, "Worktree", "mode", map[string]any{
		"description": "Worktree lifecycle: new creates a disposable worktree and removes it afterwards, " +
			"existing reuses the one at path, none runs in the checkout itself.",
		"enum": []any{"none", "new", "existing"},
	})
	patchDefProp(defs, "Worktree", "base", map[string]any{
		"description": "Commit-ish the new worktree branches from. Defaults to HEAD so the start commit is " +
			"the tree you are looking at, independent of checkout.ref.",
		"default": "HEAD",
	})
	// No default: for uncommitted — it is conditional on base, and a static
	// "clone" here would promise editor users something only sometimes true.
	patchDefProp(defs, "Worktree", "uncommitted", map[string]any{
		"description": "Whether staged, unstaged and untracked changes are carried into the new worktree. " +
			"Defaults to clone when base is HEAD, otherwise skip: uncommitted work is a diff against your " +
			"HEAD, so replaying it onto a worktree branched elsewhere applies to the wrong context. Nothing " +
			"is ever stashed — the source repository is never mutated.",
		"enum": []any{"clone", "skip"},
	})
	patchDefProp(defs, "Worktree", "ignored", map[string]any{
		"description": "Whether gitignored content is copied into the new worktree. `git worktree add` never " +
			"brings it, so skip leaves the tree without node_modules/, .env and build caches. Use skip in " +
			"repositories with a very large ignored tree.",
		"default": "clone",
		"enum":    []any{"clone", "skip"},
	})
}

// patchDefProp merges documentation onto one reflected property. A missing
// definition or property panics rather than being skipped: silently doing
// nothing would ship the bare strings this function exists to replace.
func patchDefProp(defs map[string]any, def, prop string, patch map[string]any) {
	definition, ok := defs[def].(map[string]any)
	if !ok {
		panic("captain spec schema has no $defs/" + def)
	}
	props, ok := definition["properties"].(map[string]any)
	if !ok {
		panic("captain spec schema $defs/" + def + " has no properties")
	}
	target, ok := props[prop].(map[string]any)
	if !ok {
		panic("captain spec schema $defs/" + def + " has no " + prop + " property")
	}
	for key, value := range patch {
		target[key] = value
	}
}

// aiSchema exposes the complete Captain spec because every field is a valid
// inherited default, including prompt.system, workspace, permissions, and env.
// x-prompt-picker tells the settings UI to replace the generic object form with
// the shared rich PromptPicker editor.
func aiSchema(schema map[string]any) map[string]any {
	node := specNodeSchema(schema,
		"Base AI spec inherited by every AI operation. Configure model, prompt, workspace, permissions, environment, and runtime defaults here.",
		"",
		"Default catalog model slug for all AI operations (e.g. claude-sonnet-4-5). "+
			"There is no built-in default: set it here, or globally as ai.defaultModel "+
			"in ~/.captain.yaml. A run that finds neither stops and tells you to run "+
			"`gavel configure`.")
	node["x-prompt-picker"] = true
	return node
}

// specNodeSchema clones the captain Spec definition into a config node carrying
// its own description and model default. A config field typed api.Spec accepts
// every spec field, so the shape is the spec's — only the documentation and the
// default differ per operation.
func specNodeSchema(schema map[string]any, description, modelDefault, modelDescription string) map[string]any {
	defs, ok := schema["$defs"].(map[string]any)
	if !ok {
		panic("captain spec schema has no $defs")
	}
	spec, ok := defs["Spec"].(map[string]any)
	if !ok {
		panic("captain spec schema has no Spec definition")
	}
	node := cloneSchemaMap(spec)
	node["description"] = description

	props, ok := spec["properties"].(map[string]any)
	if !ok {
		panic("captain Spec definition has no properties")
	}
	props = cloneSchemaMap(props)
	model, ok := props["model"].(map[string]any)
	if !ok {
		panic("captain Spec definition has no model property")
	}
	model = cloneSchemaMap(model)
	// An empty modelDefault means "no built-in default" — omit the key rather
	// than advertising "" to the settings UI as if it were a value.
	if modelDefault != "" {
		model["default"] = modelDefault
	}
	model["description"] = modelDescription
	props["model"] = model
	node["properties"] = props
	return node
}

func cloneSchemaMap(source map[string]any) map[string]any {
	clone := make(map[string]any, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

func captainSpecSchema() map[string]any {
	raw, err := api.SchemaJSON(&api.Spec{})
	if err != nil {
		panic("generate Captain spec schema: " + err.Error())
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		panic("decode Captain spec schema: " + err.Error())
	}
	return schema
}

// --- leaf builders -------------------------------------------------------

func object(desc string, props map[string]any) map[string]any {
	return map[string]any{
		"type":                 "object",
		"description":          desc,
		"additionalProperties": false,
		"properties":           props,
	}
}

func mapObject(desc string, value map[string]any) map[string]any {
	return map[string]any{
		"type":                 "object",
		"description":          desc,
		"additionalProperties": value,
	}
}

func arrayOf(desc string, item map[string]any) map[string]any {
	return map[string]any{
		"type":        "array",
		"description": desc,
		"items":       item,
	}
}

func stringProp(desc string) map[string]any {
	return map[string]any{"type": "string", "description": desc}
}

func stringWithDefault(desc, def string) map[string]any {
	m := stringProp(desc)
	m["default"] = def
	return m
}

func boolProp(desc string) map[string]any {
	return map[string]any{"type": "boolean", "description": desc}
}

func intProp(desc string) map[string]any {
	return map[string]any{"type": "integer", "description": desc}
}

func numberProp(desc string) map[string]any {
	return map[string]any{"type": "number", "description": desc}
}

func intWithDefault(desc string, def int) map[string]any {
	m := intProp(desc)
	m["default"] = def
	return m
}

func boolWithDefault(desc string, def bool) map[string]any {
	m := boolProp(desc)
	m["default"] = def
	return m
}

func stringArray(desc string) map[string]any {
	return map[string]any{
		"type":        "array",
		"description": desc,
		"items":       map[string]any{"type": "string"},
	}
}

// enumStringArray models a string list whose entries are restricted to a known
// vocabulary, so an editor completes them and a typo is caught in .gavel.yaml
// rather than at generation time.
func enumStringArray(desc string, values []string) map[string]any {
	allowed := make([]any, len(values))
	for i, v := range values {
		allowed[i] = v
	}
	m := stringArray(desc)
	m["items"] = map[string]any{"type": "string", "enum": allowed}
	return m
}

func stringArrayWithDefault(desc string, def []string) map[string]any {
	m := stringArray(desc)
	defAny := make([]any, len(def))
	for i, s := range def {
		defAny[i] = s
	}
	m["default"] = defAny
	return m
}

// checkModeObject models a CheckMode field: the string values prompt/fail/skip,
// or the boolean false (an alias for skip).
func checkModeObject(desc, def string) map[string]any {
	return object(desc, map[string]any{
		"mode": map[string]any{
			"description": "Gate behavior. Use false as an alias for skip.",
			"default":     def,
			"oneOf": []any{
				map[string]any{"type": "string", "enum": []any{"prompt", "fail", "skip"}},
				map[string]any{"type": "boolean", "enum": []any{false}},
			},
		},
	})
}
