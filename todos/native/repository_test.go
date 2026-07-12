package native

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeStrings(t *testing.T) {
	assert.Equal(t, []string{"bug", "priority:high", "todos"}, normalizeStrings([]string{
		" Todos ", "BUG", "bug", "", "priority:HIGH",
	}))
}

func TestNormalizeAliases(t *testing.T) {
	aliases, err := normalizeAliases([]AliasInput{
		{Alias: " E2A3B8C2 ", Kind: " GRITE "},
		{Alias: "external-42", Kind: "External"},
		{Alias: "e2a3b8c2", Kind: "grite"},
	})
	require.NoError(t, err)
	assert.Equal(t, []AliasInput{
		{Alias: "e2a3b8c2", Kind: "grite"},
		{Alias: "external-42", Kind: "external"},
	}, aliases)

	_, err = normalizeAliases([]AliasInput{
		{Alias: "same", Kind: "grite"},
		{Alias: "same", Kind: "external"},
	})
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestCanonicalRelationship(t *testing.T) {
	lower := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	higher := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	issueID, targetID := canonicalRelationship(higher, lower, RelationshipRelatedTo)
	assert.Equal(t, lower, issueID)
	assert.Equal(t, higher, targetID)

	issueID, targetID = canonicalRelationship(higher, lower, RelationshipDependsOn)
	assert.Equal(t, higher, issueID)
	assert.Equal(t, lower, targetID)
}

func TestNormalizeWorkspacePath(t *testing.T) {
	assert.Equal(t, "/workspace/gavel", normalizeWorkspacePath(" /workspace/./gavel/../gavel "))
	assert.Empty(t, normalizeWorkspacePath("   "))
}

func TestRelationshipBlocksIsReadOnly(t *testing.T) {
	assert.False(t, RelationshipBlocks.valid())
}

func TestMarshalPayload(t *testing.T) {
	payload, err := marshalPayload(map[string]any{"ok": true})
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(payload))

	payload, err = marshalPayload(nil)
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(payload))

	_, err = marshalPayload(json.RawMessage(`{"broken"`))
	require.ErrorIs(t, err, ErrInvalidInput)
}

func TestNewRepositoryRejectsNilDatabase(t *testing.T) {
	_, err := NewRepository(nil)
	require.ErrorIs(t, err, ErrInvalidInput)
}
