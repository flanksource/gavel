package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/stretchr/testify/require"
)

// newTestProvider opens a workspace provider over a freshly migrated,
// isolated database on the shared test server.
func newTestProvider(t *testing.T, databaseName string) *Provider {
	t.Helper()
	dsn := dbtest.ForT(t, dbtest.Options{Name: databaseName}).DSN()

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
		Name: "Test", RootPath: root, Repositories: []string{"example/test"},
	})
	require.NoError(t, err)
	return provider
}
