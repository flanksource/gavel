// Tests that a provider built for an organization keeps its workspace and todos in
// that organization, while New keeps using the local organization.
package runtime_test

import (
	"testing"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewForOrgScopesTheWorkspace(t *testing.T) {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_runtime_tenancy"}).DSN())
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	migrated, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, migrated.Close()) })
	db := migrated.Gorm()
	otherOrg := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Other')`, otherOrg, otherOrg.String()).Error)

	localDir, otherDir := t.TempDir(), t.TempDir()
	// New and NewForOrg take the test's own handle: Open resolves a process-wide
	// shared database that another test in this package may already hold.
	_, err = todoruntime.New(t.Context(), db, todoruntime.WorkspaceOptions{Name: "local", RootPath: localDir})
	require.NoError(t, err)
	_, err = todoruntime.NewForOrg(t.Context(), db, todoruntime.WorkspaceOptions{Name: "other", RootPath: otherDir}, tenant.Scope{OrgID: otherOrg})
	require.NoError(t, err)

	for dir, want := range map[string]uuid.UUID{localDir: tenant.LocalOrgID, otherDir: otherOrg} {
		var orgs []string
		require.NoError(t, db.Raw(`
			SELECT workspace.org_id::text FROM todo_workspaces AS workspace
			JOIN todo_workspace_paths AS path ON path.workspace_id = workspace.id
			WHERE path.path = ?`, dir).Scan(&orgs).Error)
		assert.Equal(t, []string{want.String()}, orgs, dir)
	}
}
