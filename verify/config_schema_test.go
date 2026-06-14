package verify

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	assert.Equal(t, "claude", nodeAt(t, schema, "verify", "model")["default"],
		"verify.model default should be claude")
	assert.Equal(t, "gavel test --lint", nodeAt(t, schema, "ssh", "cmd")["default"],
		"ssh.cmd default should be the fallback command")

	mode := nodeAt(t, schema, "commit", "precommit", "mode")
	assert.Equal(t, "prompt", mode["default"], "precommit.mode default should be prompt")

	compatMode := nodeAt(t, schema, "commit", "compatibility", "mode")
	assert.Equal(t, "skip", compatMode["default"], "compatibility.mode default should be skip")

	oneOf, ok := mode["oneOf"].([]any)
	require.True(t, ok, "mode should be a oneOf")
	stringBranch := oneOf[0].(map[string]any)
	assert.ElementsMatch(t, []any{"prompt", "fail", "skip"}, stringBranch["enum"])
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
