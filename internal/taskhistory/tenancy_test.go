// Tests that archived task runs are isolated per organization: another org can
// neither list, read, nor overwrite a run, even when it imports the same run ID.
package taskhistory_test

import (
	"testing"
	"time"

	clickytask "github.com/flanksource/clicky/task"
	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/internal/taskhistory"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskHistoryIsolatesOrganizations(t *testing.T) {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_task_history_tenancy"}).DSN())
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	db, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	otherOrg := uuid.New()
	require.NoError(t, db.Gorm().Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Other')`, otherOrg, otherOrg.String()).Error)

	local, err := taskhistory.NewStore(db.Gorm())
	require.NoError(t, err)
	other := local.ForOrg(otherOrg)
	ctx := t.Context()
	now := time.Now().UTC()
	record := func(id, name string) taskhistory.Record {
		return taskhistory.Record{
			Run:        clickytask.RunMeta{ID: id, Name: name, StartedAt: now.Add(-time.Minute).Format(time.RFC3339Nano)},
			Snapshots:  []clickytask.TaskSnapshot{{ID: id, GroupID: id, Name: name, Type: "group", Status: string(clickytask.StatusSuccess)}},
			ArchivedAt: now.Add(-time.Minute),
		}
	}

	_, err = local.Import(ctx, []taskhistory.Record{record("run-local", "Local")})
	require.NoError(t, err)
	runs, err := other.ListRuns(ctx, now)
	require.NoError(t, err)
	assert.Empty(t, runs)
	snapshots, err := other.Snapshot(ctx, "run-local")
	require.NoError(t, err)
	assert.Nil(t, snapshots)

	_, err = other.Import(ctx, []taskhistory.Record{record("run-local", "Hijacked"), record("run-other", "Other")})
	require.NoError(t, err)
	runs, err = local.ListRuns(ctx, now)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	assert.Equal(t, "Local", runs[0].Run.Name, "another org's import must not overwrite this org's run")
	snapshots, err = local.Snapshot(ctx, "run-local")
	require.NoError(t, err)
	require.Len(t, snapshots, 1)
	assert.Equal(t, "Local", snapshots[0].Name)

	runs, err = other.ListRuns(ctx, now)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	assert.Equal(t, "run-other", runs[0].Run.ID)
}
