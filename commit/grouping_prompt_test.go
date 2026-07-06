package commit

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/stretchr/testify/require"
)

// updateGrouping regenerates the committed golden file instead of asserting
// against it: `go test ./commit -run TestRenderGroupingPrompt -update-grouping-golden`.
var updateGrouping = flag.Bool("update-grouping-golden", false, "update grouping prompt golden files")

func assertGroupingGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *updateGrouping {
		require.NoError(t, os.MkdirAll("testdata", 0o755))
		require.NoError(t, os.WriteFile(path, []byte(got), 0o644))
		return
	}
	want, err := os.ReadFile(path)
	require.NoErrorf(t, err, "missing golden %s; regenerate with: go test ./commit -run TestRenderGroupingPrompt -update-grouping-golden", path)
	require.Equal(t, string(want), got)
}

// groupingTable is a markdown status table whose single row deliberately carries
// a rename arrow and & / < so the golden asserts dotprompt's NoEscape rendering:
// a plain-ASCII fixture would stay green even if HTML-escaping regressed.
const groupingTable = "| Scope | File | Status |\n" +
	"| ----- | ---- | ------ |\n" +
	"| api | old & <legacy>.go → new.go | renamed |\n"

func TestRenderGroupingPromptBase(t *testing.T) {
	got, schema, strictness, err := renderGroupingPrompt(groupingPromptTemplate, groupingTable, 3, true)
	require.NoError(t, err)
	require.NotEmpty(t, schema, "frontmatter output.schema must be carried through")
	require.Equal(t, api.SchemaStrictnessRetry, strictness, "frontmatter declares schemaStrictness: retry")

	require.NotContains(t, got, "&amp;", "table must not be HTML-escaped")
	require.NotContains(t, got, "&lt;", "table must not be HTML-escaped")
	require.Contains(t, got, "old & <legacy>.go → new.go", "raw table content must be preserved")
	require.Contains(t, got, "Treat each scope as the primary boundary", "groupByScope selects the scope branch")
	require.Contains(t, got, "Produce at most 3 commits", "maxCommits renders the cap rule")

	groups := groupsSchemaNode(t, schema)
	require.EqualValues(t, 3, groups["maxItems"], "maxCommits caps the groups array as maxItems")
	require.EqualValues(t, 1, groups["minItems"])
	assertGroupingGolden(t, "grouping-base.golden", got)
}

func TestRenderGroupingPromptFlatNoCap(t *testing.T) {
	got, schema, _, err := renderGroupingPrompt(groupingPromptTemplate, groupingTable, 0, false)
	require.NoError(t, err)
	require.Contains(t, got, "The Scope column is a hint", "flat grouping selects the logical-change branch")
	require.NotContains(t, got, "Produce at most", "zero maxCommits omits the cap rule")

	groups := groupsSchemaNode(t, schema)
	_, hasMax := groups["maxItems"]
	require.False(t, hasMax, "zero maxCommits omits maxItems from the schema")
	require.EqualValues(t, 1, groups["minItems"], "minItems is always declared")
}

// groupsSchemaNode unmarshals the grouping output schema and returns its
// properties.groups node (the array carrying the minItems/maxItems cap).
func groupsSchemaNode(t *testing.T, schemaJSON json.RawMessage) map[string]any {
	t.Helper()
	var schema map[string]any
	require.NoError(t, json.Unmarshal(schemaJSON, &schema))
	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok, "schema has a properties object")
	groups, ok := props["groups"].(map[string]any)
	require.True(t, ok, "schema has a groups array")
	return groups
}
