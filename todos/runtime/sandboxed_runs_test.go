// Tests that a workspace whose runs execute in sandboxes gives its provider that knowledge, which
// keeps the provider from reading plans off this host's disk for those runs.
package runtime

import (
	"testing"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSandboxedRunsReachTheProvider(t *testing.T) {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_runtime_sandboxed_runs"}).DSN())
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	migrated, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, migrated.Close()) })
	db := migrated.Gorm()
	org := uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Acme')`, org, org.String()).Error)
	scope := tenant.Scope{OrgID: org}

	sandboxed, err := NewForOrg(t.Context(), db, WorkspaceOptions{Name: "app", RootPath: t.TempDir(), SandboxedRuns: true}, scope)
	require.NoError(t, err)
	assert.True(t, sandboxed.sandboxedRuns)

	local, err := NewForOrg(t.Context(), db, WorkspaceOptions{Name: "local", RootPath: t.TempDir()}, scope)
	require.NoError(t, err)
	assert.False(t, local.sandboxedRuns)
}
