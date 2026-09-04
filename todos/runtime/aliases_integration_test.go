package runtime

import (
	"os"
	"path/filepath"
	"testing"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderAliasesIntegration(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres native runtime tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir: filepath.Join(t.TempDir(), "postgres"), Database: "gavel_todo_aliases",
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
		Name: "Aliases", RootPath: root, Repositories: []string{"example/aliases"},
	})
	require.NoError(t, err)

	todo, err := provider.Create(t.Context(), todos.CreateRequest{Title: "Push me", Body: "body"})
	require.NoError(t, err)

	require.NoError(t, provider.AddAlias(t.Context(), todo, todos.TodoAlias{Alias: "legacy-42", Kind: "import"}))
	firstVersion := todo.Version

	// A second alias must not drop the first: native storage replaces the whole
	// alias set, so AddAlias reads before it writes.
	require.NoError(t, provider.AddAlias(t.Context(), todo, todos.TodoAlias{Alias: "example/aliases#7", Kind: "github"}))

	aliases, err := provider.Aliases(t.Context(), todo)
	require.NoError(t, err)
	assert.ElementsMatch(t, []todos.TodoAlias{
		{Alias: "legacy-42", Kind: "import"},
		{Alias: "example/aliases#7", Kind: "github"},
	}, aliases)

	// The caller's TODO is refreshed, so a follow-up mutation passes its
	// optimistic-concurrency check against the bumped version.
	assert.Greater(t, todo.Version, firstVersion)
	require.NoError(t, provider.Comment(t.Context(), todo, "pushed"))

	// Both aliases resolve back to the same issue.
	for _, ref := range []string{"legacy-42", "example/aliases#7"} {
		resolved, err := provider.Get(t.Context(), ref)
		require.NoError(t, err, "resolve %q", ref)
		assert.Equal(t, todo.ID, resolved.ID)
	}

	// A stale version fails loudly rather than clobbering a concurrent write.
	stale := *todo
	stale.Version = firstVersion - 1
	require.Error(t, provider.AddAlias(t.Context(), &stale, todos.TodoAlias{Alias: "stale", Kind: "github"}))
}

// The list and detail reads both decorate a TODO with the GitHub issue it is
// linked to, so the dashboard can filter linked from unlinked work. Only the
// `github` alias counts: an imported reference is not an issue link.
func TestProviderExternalIssueIntegration(t *testing.T) {
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		t.Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres native runtime tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir: filepath.Join(t.TempDir(), "postgres"), Database: "gavel_todo_external",
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
		Name: "External", RootPath: root, Repositories: []string{"example/external"},
	})
	require.NoError(t, err)

	pushed, err := provider.Create(t.Context(), todos.CreateRequest{Title: "Pushed", Body: "body"})
	require.NoError(t, err)
	require.NoError(t, provider.AddAlias(t.Context(), pushed,
		todos.TodoAlias{Alias: "example/external#42", Kind: "github"}))

	imported, err := provider.Create(t.Context(), todos.CreateRequest{Title: "Imported", Body: "body"})
	require.NoError(t, err)
	require.NoError(t, provider.AddAlias(t.Context(), imported,
		todos.TodoAlias{Alias: "legacy-9", Kind: "import"}))

	bare, err := provider.Create(t.Context(), todos.CreateRequest{Title: "Bare", Body: "body"})
	require.NoError(t, err)

	want := &types.ExternalIssue{
		Kind: "github", Repo: "example/external", Number: 42,
		URL: "https://github.com/example/external/issues/42",
	}

	detail, err := provider.Get(t.Context(), pushed.ID)
	require.NoError(t, err)
	assert.Equal(t, want, detail.ExternalIssue)

	for _, ref := range []string{imported.ID, bare.ID} {
		unlinked, err := provider.Get(t.Context(), ref)
		require.NoError(t, err)
		assert.Nil(t, unlinked.ExternalIssue, "todo %s has no GitHub issue", ref)
	}

	listed, err := provider.List(t.Context(), todos.DiscoveryFilters{})
	require.NoError(t, err)
	byID := map[string]*types.ExternalIssue{}
	for _, item := range listed {
		byID[item.ID] = item.ExternalIssue
	}
	assert.Equal(t, want, byID[pushed.ID])
	assert.Nil(t, byID[imported.ID])
	assert.Nil(t, byID[bare.ID])
}
