package runtime

import (
	"os"
	"path/filepath"
	"testing"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderRelationshipsIntegration(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres native runtime tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir: filepath.Join(t.TempDir(), "postgres"), Database: "gavel_todo_links",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, stop()) })

	t.Setenv(database.EnvDSN, dsn)
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })

	root := filepath.Join(t.TempDir(), "workspace")
	require.NoError(t, os.MkdirAll(root, 0o755))
	provider, err := New(t.Context(), opened.Gorm(), WorkspaceOptions{
		Name: "Links", RootPath: root, Repositories: []string{"example/links"},
	})
	require.NoError(t, err)

	blocked, err := provider.Create(t.Context(), todos.CreateRequest{Title: "Blocked work", Body: "needs the blocker"})
	require.NoError(t, err)
	blocker, err := provider.Create(t.Context(), todos.CreateRequest{Title: "Blocking work", Body: "must land first"})
	require.NoError(t, err)
	sibling, err := provider.Create(t.Context(), todos.CreateRequest{Title: "Duplicate report", Body: "same bug"})
	require.NoError(t, err)

	link, err := provider.Link(t.Context(), blocked, blocker.ID, types.RelationDependsOn)
	require.NoError(t, err)
	assert.Equal(t, types.RelationDependsOn, link.Relation)
	assert.Equal(t, blocker.ID, link.TargetID)
	assert.Equal(t, "Blocking work", link.TargetTitle)

	_, err = provider.Link(t.Context(), blocked, sibling.ID, types.RelationRelatedTo)
	require.NoError(t, err)

	links, err := provider.Links(t.Context(), blocked)
	require.NoError(t, err)
	require.Len(t, links, 2)
	byRelation := map[types.RelationKind]todos.Link{}
	for _, l := range links {
		byRelation[l.Relation] = l
	}
	assert.Equal(t, blocker.ID, byRelation[types.RelationDependsOn].TargetID)
	assert.Equal(t, sibling.ID, byRelation[types.RelationRelatedTo].TargetID)

	// The blocker sees the same dependency as the derived read-only relation.
	reverse, err := provider.Links(t.Context(), blocker)
	require.NoError(t, err)
	require.Len(t, reverse, 1)
	assert.Equal(t, types.RelationBlocks, reverse[0].Relation)
	assert.Equal(t, blocked.ID, reverse[0].TargetID)

	// blocks is derived; writing it must fail rather than create a second edge.
	_, err = provider.Link(t.Context(), blocker, blocked.ID, types.RelationBlocks)
	require.Error(t, err)

	// A dependency cycle must be rejected by the repository.
	_, err = provider.Link(t.Context(), blocker, blocked.ID, types.RelationDependsOn)
	require.ErrorIs(t, err, native.ErrRelationshipCycle)

	// A duplicate edge must be rejected rather than silently deduplicated.
	_, err = provider.Link(t.Context(), blocked, blocker.ID, types.RelationDependsOn)
	require.ErrorIs(t, err, native.ErrRelationshipExists)

	require.NoError(t, provider.Unlink(t.Context(), blocked, blocker.ID, types.RelationDependsOn))
	links, err = provider.Links(t.Context(), blocked)
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, types.RelationRelatedTo, links[0].Relation)

	require.ErrorIs(t,
		provider.Unlink(t.Context(), blocked, blocker.ID, types.RelationDependsOn),
		native.ErrRelationshipNotFound)
}
