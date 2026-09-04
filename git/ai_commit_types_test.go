package git

import (
	"encoding/json"
	"testing"

	"github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/require"
)

// schemaTypeProperty pulls the `type` property out of the rendered prompt's
// output schema, which is what the provider enforces against the model.
func schemaTypeProperty(t *testing.T, schemaJSON json.RawMessage) map[string]any {
	t.Helper()
	require.NotEmpty(t, schemaJSON, "prompt frontmatter must declare an output schema")

	var schema struct {
		Properties map[string]map[string]any `json:"properties"`
	}
	require.NoError(t, json.Unmarshal(schemaJSON, &schema))
	require.Contains(t, schema.Properties, "type")
	return schema.Properties["type"]
}

func schemaEnum(t *testing.T, prop map[string]any) []string {
	t.Helper()
	raw, ok := prop["enum"].([]any)
	require.Truef(t, ok, "type property must declare an enum, got %#v", prop)
	out := make([]string, len(raw))
	for i, v := range raw {
		out[i], ok = v.(string)
		require.True(t, ok, "enum entries must be strings")
	}
	return out
}

// TestCommitMessageSchema_TypeIsEnumerated locks in the fix for a message
// generated as "feat|fix|perf|...: <subject>": the model was handed `type` as a
// free-form string whose *description* listed the choices, and echoed that list
// verbatim. The choices must reach the model as a schema enum it cannot invent
// around, never as prose.
func TestCommitMessageSchema_TypeIsEnumerated(t *testing.T) {
	_, schemaJSON, err := renderCommitPrompt(sampleCommit(), commitMessagePrompt, 0, defaultTypes())
	require.NoError(t, err)

	prop := schemaTypeProperty(t, schemaJSON)
	require.Equal(t, defaultTypes(), schemaEnum(t, prop))

	description, _ := prop["description"].(string)
	require.NotContains(t, description, "|",
		"the description must not restate the choices as a pipe list — that is the string the model echoed")
}

// TestCommitMessageSchema_EnumFollowsConfiguredTypes proves .gavel.yaml
// commit.types reaches the schema, so a project that allows only feat and fix
// cannot be handed a chore.
func TestCommitMessageSchema_EnumFollowsConfiguredTypes(t *testing.T) {
	configured := []string{"feat", "fix"}
	allowed, err := allowedCommitTypes(configured)
	require.NoError(t, err)

	promptText, schemaJSON, err := renderCommitPrompt(sampleCommit(), commitMessagePrompt, 0, allowed)
	require.NoError(t, err)

	require.Equal(t, configured, schemaEnum(t, schemaTypeProperty(t, schemaJSON)))
	require.Contains(t, promptText, "exactly one of feat|fix",
		"the prose requirement must list the same vocabulary as the enum")
	require.NotContains(t, promptText, "chore", "a type the project excluded must not be offered")
}

// TestCommitMessagePrompt_FrontmatterParsesBeforeTemplating pins that the enum
// lives in Go, not in the frontmatter. Frontmatter is parsed as YAML before any
// templating, so a Handlebars block inside it broke every reader that parses
// the document without rendering it — the settings editor 400ed on the default
// commit prompt, and the prompt catalog rendered it as a zero-valued spec.
func TestCommitMessagePrompt_FrontmatterParsesBeforeTemplating(t *testing.T) {
	doc, err := prompt.Parse(commitMessagePrompt)
	require.NoError(t, err, "the default commit prompt must parse as a document without rendering")
	output, ok := doc.Frontmatter["output"].(map[string]any)
	require.True(t, ok, "the output schema still lives in the frontmatter: %#v", doc.Frontmatter["output"])

	declared, err := json.Marshal(output["schema"])
	require.NoError(t, err)
	var schema struct {
		Properties map[string]map[string]any `json:"properties"`
	}
	require.NoError(t, json.Unmarshal(declared, &schema))
	require.Contains(t, schema.Properties, "type")
	require.NotContains(t, schema.Properties["type"], "enum",
		"the unrendered frontmatter declares no vocabulary; the schema builder enumerates it per project")
}

// A template that declares an output schema without a `type` property has
// nothing to enumerate and is handed over untouched, not rewritten.
func TestEnumerateCommitType_LeavesSchemasWithoutATypeProperty(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"subject":{"type":"string"}}}`)
	got, err := enumerateCommitType(schema, defaultTypes())
	require.NoError(t, err)
	require.JSONEq(t, string(schema), string(got))

	got, err = enumerateCommitType(nil, defaultTypes())
	require.NoError(t, err)
	require.Empty(t, got)
}

// TestCommitMessageSchema_BoundsFreeTextFields guards the other half of the bad
// message, where the body's prose landed in scope and subject.
func TestCommitMessageSchema_BoundsFreeTextFields(t *testing.T) {
	_, schemaJSON, err := renderCommitPrompt(sampleCommit(), commitMessagePrompt, 0, defaultTypes())
	require.NoError(t, err)

	var schema struct {
		Properties map[string]map[string]any `json:"properties"`
	}
	require.NoError(t, json.Unmarshal(schemaJSON, &schema))

	for field, want := range map[string]float64{"scope": 20, "subject": 100} {
		require.Contains(t, schema.Properties, field)
		require.Equalf(t, want, schema.Properties[field]["maxLength"],
			"%s must be length-bounded so prose cannot be dumped into it", field)
	}
}

func TestAllowedCommitTypes(t *testing.T) {
	t.Run("no configuration uses gavel's defaults", func(t *testing.T) {
		got, err := allowedCommitTypes(nil)
		require.NoError(t, err)
		require.Equal(t, commitTypeNames(models.SelectableCommitTypes()), got)
		require.Contains(t, got, string(models.CommitTypeFeat))
		require.NotContains(t, got, string(models.CommitTypeOther),
			"the parser's fallback bucket must never be offered as a choice")
	})

	t.Run("configured types replace the defaults", func(t *testing.T) {
		got, err := allowedCommitTypes([]string{"feat", "fix", "chore"})
		require.NoError(t, err)
		require.Equal(t, []string{"feat", "fix", "chore"}, got)
	})

	t.Run("surrounding whitespace is tolerated", func(t *testing.T) {
		got, err := allowedCommitTypes([]string{" feat ", "fix"})
		require.NoError(t, err)
		require.Equal(t, []string{"feat", "fix"}, got)
	})

	t.Run("an unknown configured type fails loudly", func(t *testing.T) {
		_, err := allowedCommitTypes([]string{"feat", "shipit"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "shipit", "the error must name the offending value")
		require.Contains(t, err.Error(), "feat", "the error must list the known types")
	})
}

// TestCommitTypeIsValid covers the shared vocabulary the enum and the response
// check both read from.
func TestCommitTypeIsValid(t *testing.T) {
	for _, ct := range models.SelectableCommitTypes() {
		require.Truef(t, ct.IsValid(), "%s is selectable so it must be valid", ct)
	}
	require.True(t, models.CommitTypeOther.IsValid(), "other stays valid on input")
	require.False(t, models.CommitTypeUnknown.IsValid(), "the empty type is not valid")
	require.False(t, models.CommitType("feat|fix|perf").IsValid(), "an echoed choice list is not a type")
}
