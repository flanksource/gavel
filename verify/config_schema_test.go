package verify

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// specBoundaryTypes are the captain-spec-shaped config types the schema documents
// with a curated, hand-built shape (aiSchema / promptSpecSchema) rather than a
// 1:1 reflection of every nested api.Spec field. The parity walk asserts the node
// exists at the parent but does not descend into the spec internals.
var specBoundaryTypes = map[reflect.Type]bool{
	reflect.TypeOf(api.Spec{}):   true,
	reflect.TypeOf(PromptSpec{}): true,
}

func parsedConfigSchema(t *testing.T) map[string]any {
	t.Helper()
	raw, err := ConfigJSONSchema()
	require.NoError(t, err)

	var schema map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &schema), "schema must be valid JSON")
	return schema
}

func TestConfigJSONSchema_TopLevelShape(t *testing.T) {
	schema := parsedConfigSchema(t)

	assert.Equal(t, "https://json-schema.org/draft/2020-12/schema", schema["$schema"])
	assert.Equal(t, ConfigSchemaID, schema["$id"])
	assert.Equal(t, "object", schema["type"])
	assert.Equal(t, false, schema["additionalProperties"])
	assert.NotEmpty(t, schema["description"])
}

// TestConfigJSONSchema_CoversStruct walks GavelConfig with reflection and
// asserts every yaml-tagged field at every depth has a matching schema node.
// This fails loudly whenever a field is added to the config without being
// documented in the schema.
func TestConfigJSONSchema_CoversStruct(t *testing.T) {
	schema := parsedConfigSchema(t)
	assertSchemaCoversType(t, reflect.TypeOf(GavelConfig{}), schema, "$")
}

func assertSchemaCoversType(t *testing.T, typ reflect.Type, node map[string]any, path string) {
	t.Helper()
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	// api.Spec / PromptSpec are curated schema boundaries: the parent already
	// asserted the node exists; the spec's internal shape is hand-built.
	if specBoundaryTypes[typ] {
		return
	}

	switch typ.Kind() {
	case reflect.Struct:
		props, ok := node["properties"].(map[string]any)
		require.Truef(t, ok, "%s: schema node is missing object properties", path)
		for i := 0; i < typ.NumField(); i++ {
			name := yamlFieldName(typ.Field(i))
			if name == "" {
				continue
			}
			child, ok := props[name].(map[string]any)
			require.Truef(t, ok, "%s.%s: field is not documented in the schema", path, name)
			assertSchemaCoversType(t, typ.Field(i).Type, child, path+"."+name)
		}
	case reflect.Slice, reflect.Array:
		elem := typ.Elem()
		if deref(elem).Kind() == reflect.Struct {
			items, ok := node["items"].(map[string]any)
			require.Truef(t, ok, "%s[]: schema array is missing items", path)
			assertSchemaCoversType(t, elem, items, path+"[]")
		}
	case reflect.Map:
		elem := typ.Elem()
		if deref(elem).Kind() == reflect.Struct {
			value, ok := node["additionalProperties"].(map[string]any)
			require.Truef(t, ok, "%s.*: schema map is missing additionalProperties value", path)
			assertSchemaCoversType(t, elem, value, path+".*")
		}
	default:
		// scalar leaf — nothing to descend into
	}
}

func deref(typ reflect.Type) reflect.Type {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ
}

func yamlFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("yaml")
	if tag == "" || tag == "-" {
		return ""
	}
	if comma := indexByte(tag, ','); comma >= 0 {
		tag = tag[:comma]
	}
	return tag
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func TestConfigJSONSchema_DefaultsAndEnums(t *testing.T) {
	schema := parsedConfigSchema(t)

	// No built-in default model: the key must be absent rather than advertising
	// "" as a value, so the settings UI shows an unset field instead of a model
	// nobody chose.
	_, hasDefault := nodeAt(t, schema, "ai", "model")["default"]
	assert.False(t, hasDefault, "ai.model must advertise no built-in default model")
	assert.Equal(t, "gavel test --lint", nodeAt(t, schema, "ssh", "cmd")["default"],
		"ssh.cmd default should be the fallback command")

	mode := nodeAt(t, schema, "commit", "precommit", "mode")
	assert.Equal(t, "prompt", mode["default"], "precommit.mode default should be prompt")

	oneOf, ok := mode["oneOf"].([]any)
	require.True(t, ok, "mode should be a oneOf")
	stringBranch := oneOf[0].(map[string]any)
	assert.ElementsMatch(t, []any{"prompt", "fail", "skip"}, stringBranch["enum"])
}

// TestConfigJSONSchema_SetupDefsAreDocumented covers patchSetupDefs: the checkout
// and worktree enums are reflected out of commons-db as bare strings, so without
// the patch an editor offers no completion at all for the three fields that decide
// which tree a run happens in.
func TestConfigJSONSchema_SetupDefsAreDocumented(t *testing.T) {
	schema := parsedConfigSchema(t)
	defs, ok := schema["$defs"].(map[string]any)
	require.True(t, ok, "schema should carry the reflected captain $defs")

	def := func(name, prop string) map[string]any {
		t.Helper()
		definition, ok := defs[name].(map[string]any)
		require.True(t, ok, "$defs/%s should exist", name)
		props, ok := definition["properties"].(map[string]any)
		require.True(t, ok, "$defs/%s should document properties", name)
		node, ok := props[prop].(map[string]any)
		require.True(t, ok, "$defs/%s should document %q", name, prop)
		assert.NotEmpty(t, node["description"], "$defs/%s.%s should be described", name, prop)
		return node
	}

	assert.ElementsMatch(t, []any{"none", "local", "remote"}, def("Checkout", "mode")["enum"])

	worktreeMode := def("Worktree", "mode")
	assert.ElementsMatch(t, []any{"none", "new", "existing"}, worktreeMode["enum"])

	assert.Equal(t, "HEAD", def("Worktree", "base")["default"],
		"worktree.base defaults to HEAD so the start commit is deterministic")

	uncommitted := def("Worktree", "uncommitted")
	assert.ElementsMatch(t, []any{"clone", "skip"}, uncommitted["enum"])
	assert.NotContains(t, uncommitted, "default",
		"uncommitted's default is conditional on base, so a static default would misinform completion")

	ignored := def("Worktree", "ignored")
	assert.ElementsMatch(t, []any{"clone", "skip"}, ignored["enum"])
	assert.Equal(t, "clone", ignored["default"])
}

// TestConfigJSONSchema_PromptSpecShape asserts a PromptSpec node renders as the
// string|object union carrying x-prompt-id and the spec sub-fields the settings
// UI edits (model/prompt/budget), so the schema stays in step with promptSpecSchema.
func TestConfigJSONSchema_PromptSpecShape(t *testing.T) {
	schema := parsedConfigSchema(t)
	msg := nodeAt(t, schema, "commit", "message")
	fix := nodeAt(t, schema, "lint", "fix")

	assert.Equal(t, prompts.CommitMessage, msg["x-prompt-id"],
		"commit.message x-prompt-id must match the prompts registry ID")
	assert.ElementsMatch(t, []any{"string", "object"}, msg["type"],
		"a PromptSpec accepts a bare string or an object")

	props, ok := msg["properties"].(map[string]any)
	require.True(t, ok, "commit.message should document object properties")
	for _, key := range []string{"model", "fallbacks", "effort", "budget", "prompt", "file"} {
		assert.Contains(t, props, key, "commit.message should document %q", key)
	}
	promptProps, ok := props["prompt"].(map[string]any)["properties"].(map[string]any)
	require.True(t, ok, "prompt should document user/system")
	assert.Contains(t, promptProps, "user")
	assert.Contains(t, promptProps, "system")
	assert.Equal(t, prompts.LintFix, fix["x-prompt-id"],
		"lint.fix must be documented as a separate AI operation")
}

func TestConfigJSONSchema_AIDefaultsUsePromptPickerSpec(t *testing.T) {
	schema := parsedConfigSchema(t)
	ai := nodeAt(t, schema, "ai")

	assert.Equal(t, true, ai["x-prompt-picker"],
		"ai defaults should render through the shared PromptPicker field")
	props, ok := ai["properties"].(map[string]any)
	require.True(t, ok, "ai defaults should expose the complete Captain spec")
	for _, key := range []string{"model", "prompt", "budget", "memory", "permissions", "setup", "workflow", "cliArgs"} {
		assert.Contains(t, props, key, "ai defaults should document %q", key)
	}
	prompt := props["prompt"].(map[string]any)
	assert.Equal(t, "#/$defs/Prompt", prompt["$ref"])
	defs, ok := schema["$defs"].(map[string]any)
	require.True(t, ok, "Captain definitions should be embedded")
	promptDef, ok := defs["Prompt"].(map[string]any)
	require.True(t, ok, "Captain Prompt definition should be embedded")
	promptProps, ok := promptDef["properties"].(map[string]any)
	require.True(t, ok, "Captain Prompt definition should document its fields")
	assert.Contains(t, promptProps, "system", "AI defaults must support a system prompt")
}

// nodeAt walks properties[...] nodes by key, returning the leaf schema map.
func nodeAt(t *testing.T, schema map[string]any, keys ...string) map[string]any {
	t.Helper()
	node := schema
	for _, key := range keys {
		props, ok := node["properties"].(map[string]any)
		require.Truef(t, ok, "node missing properties before key %q", key)
		next, ok := props[key].(map[string]any)
		require.Truef(t, ok, "missing property %q", key)
		node = next
	}
	return node
}

// TestConfigJSONSchema_ExampleValidates loads the bundled annotated example and
// the repo's own .gavel.yaml and asserts every key they use is declared in the
// schema (additionalProperties is false everywhere, so an undocumented key would
// make the file invalid against the schema).
func TestConfigJSONSchema_ExampleParsesIntoConfig(t *testing.T) {
	for _, path := range []string{"../gavel.yaml.example", "../.gavel.yaml"} {
		data, err := os.ReadFile(path)
		require.NoErrorf(t, err, "read %s", path)
		var cfg GavelConfig
		require.NoErrorf(t, yaml.Unmarshal(data, &cfg), "%s should unmarshal into GavelConfig", path)
	}
}

// TestConfigSchema_GoldenMatchesCommitted guards the committed artifact: if
// GavelConfig changes, regenerate with `go generate .`.
func TestConfigSchema_GoldenMatchesCommitted(t *testing.T) {
	want, err := ConfigJSONSchema()
	require.NoError(t, err)

	committed, err := os.ReadFile(filepath.Join("..", "gavel.schema.json"))
	require.NoError(t, err, "gavel.schema.json should exist; run `go generate .`")

	assert.Equal(t, want, string(committed),
		"gavel.schema.json is stale; regenerate with `go generate .`")
}
