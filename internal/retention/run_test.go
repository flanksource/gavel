// Tests retention.Run against a migrated database and local blob stores: across two
// organizations, exactly the rows and blobs older than their window are deleted.
package retention

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func openRetentionDB(t *testing.T) *gorm.DB {
	t.Helper()
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_retention"}).DSN())
	t.Setenv(database.EnvDisable, "")
	opened, err := database.Open(t.Context(), database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, opened.Close()) })
	return opened.Gorm()
}

func putBlob(t *testing.T, store blob.Store, key string) {
	t.Helper()
	require.NoError(t, store.Write(context.Background(), key, strings.NewReader(key)))
}

func hasBlob(t *testing.T, store blob.Store, key string) bool {
	t.Helper()
	exists, err := store.Exists(context.Background(), key)
	require.NoError(t, err)
	return exists
}

func countRows(t *testing.T, db *gorm.DB, query string, args ...any) int {
	t.Helper()
	var n int
	require.NoError(t, db.Raw(query, args...).Scan(&n).Error)
	return n
}

func TestRunDeletesWhatOutlivedItsWindow(t *testing.T) {
	db := openRetentionDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	old, fresh := now.Add(-90*24*time.Hour), now.Add(-24*time.Hour)
	local, other := tenant.LocalOrgID, uuid.New()
	require.NoError(t, db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Other')`, other, other.String()).Error)

	logs := blob.Local(t.TempDir())
	orgStore := func(org uuid.UUID) blob.Store { return blob.WithPrefix(logs, "orgs/"+org.String()+"/") }
	workDir, otherWorkDir := t.TempDir(), t.TempDir()
	workspace := func(dir string) (blob.Store, error) { return blob.Local(filepath.Join(dir, ".gavel")), nil }
	recordingsRoot := t.TempDir()
	recordings := blob.Local(recordingsRoot)

	// Job logs: two old ones in different orgs, a fresh one, and an old legacy row with no blob.
	for _, row := range []struct {
		org     uuid.UUID
		job     int64
		fetched time.Time
		key     string
	}{{local, 1, old, "joblogs/1.log.gz"}, {other, 1, old, "joblogs/1.log.gz"}, {local, 2, fresh, "joblogs/2.log.gz"}, {local, 3, old, ""}} {
		require.NoError(t, db.Exec(`INSERT INTO job_log_caches (org_id, job_id, repo, log_key, fetched_at) VALUES (?, ?, 'a/b', ?, ?)`, row.org, row.job, row.key, row.fetched).Error)
		if row.key != "" {
			putBlob(t, orgStore(row.org), row.key)
		}
	}

	// Test and lint runs. A sha-* row names the snapshot last.json points at, so only its row goes.
	localRuns, _ := workspace(workDir)
	otherRuns, _ := workspace(otherWorkDir)
	for _, row := range []struct {
		org     uuid.UUID
		dir     string
		store   blob.Store
		runID   string
		kind    string
		started time.Time
	}{
		{local, workDir, localRuns, "run-old", "test", old},
		{local, workDir, localRuns, "run-fresh", "test", fresh},
		{local, workDir, localRuns, "run-mixed-old", "test+lint", old},
		{local, workDir, localRuns, "sha-abc", "test", old},
		{other, otherWorkDir, otherRuns, "run-lint-old", "lint", old},
		{other, otherWorkDir, otherRuns, "run-lint-fresh", "lint", fresh},
	} {
		require.NoError(t, db.Exec(`INSERT INTO test_run_caches (org_id, workspace_dir, run_id, path, kind, started_ts, synced_at) VALUES (?, ?, ?, ?, ?, ?, now())`,
			row.org, row.dir, row.runID, filepath.Join(row.dir, ".gavel", row.runID+".json"), row.kind, row.started.UnixNano()).Error)
		putBlob(t, row.store, row.runID+".json")
	}

	// Linter executions.
	for _, executed := range []time.Time{old, fresh} {
		require.NoError(t, db.Exec(`INSERT INTO linter_executions (linter_name, work_dir, executed_at, duration_ms, violation_count, success) VALUES ('golangci-lint', '/w', ?, 1, 0, true)`, executed).Error)
	}

	// Sandbox runs: an old finished one with a log, a fresh one, and an unfinished old one.
	oldRun, freshRun, unfinished := uuid.New(), uuid.New(), uuid.New()
	for _, row := range []struct {
		id       uuid.UUID
		finished *time.Time
		key      string
	}{{oldRun, &old, "sandbox/" + oldRun.String() + ".log"}, {freshRun, &fresh, "sandbox/" + freshRun.String() + ".log"}, {unfinished, nil, ""}} {
		require.NoError(t, db.Exec(`INSERT INTO sandbox_runs (id, org_id, kind, status, backend, log_key, created_at, finished_at) VALUES (?, ?, 'test', 'succeeded', 'container', ?, ?, ?)`,
			row.id, other, row.key, old, row.finished).Error)
		if row.key != "" {
			putBlob(t, orgStore(other), row.key)
		}
	}

	// Recordings, aged by modification time.
	const oldRecording, freshRecording = "workspaces/abc/recordings/rec-old.har", "workspaces/abc/recordings/rec-fresh.har"
	putBlob(t, recordings, oldRecording)
	putBlob(t, recordings, freshRecording)
	putBlob(t, recordings, "workspaces/abc/run-old.json")
	require.NoError(t, os.Chtimes(filepath.Join(recordingsRoot, filepath.FromSlash(oldRecording)), old, old))
	require.NoError(t, os.Chtimes(filepath.Join(recordingsRoot, "workspaces", "abc", "run-old.json"), old, old))

	report, err := Run(ctx, db, Stores{Org: orgStore, Workspace: workspace, Recordings: recordings}, now)
	require.NoError(t, err)

	assert.Equal(t, Report{JobLogs: 3, TestRuns: 3, LintRuns: 1, Recordings: 1, LinterExecutions: 1, SandboxRuns: 1}, report)

	assert.Equal(t, 1, countRows(t, db, `SELECT count(*) FROM job_log_caches`))
	assert.False(t, hasBlob(t, orgStore(local), "joblogs/1.log.gz"))
	assert.False(t, hasBlob(t, orgStore(other), "joblogs/1.log.gz"))
	assert.True(t, hasBlob(t, orgStore(local), "joblogs/2.log.gz"))

	var runs []string
	require.NoError(t, db.Raw(`SELECT run_id FROM test_run_caches ORDER BY run_id`).Scan(&runs).Error)
	assert.Equal(t, []string{"run-fresh", "run-lint-fresh"}, runs)
	assert.False(t, hasBlob(t, localRuns, "run-old.json"))
	assert.False(t, hasBlob(t, localRuns, "run-mixed-old.json"))
	assert.True(t, hasBlob(t, localRuns, "sha-abc.json"), "the snapshot last.json may name stays")
	assert.True(t, hasBlob(t, localRuns, "run-fresh.json"))
	assert.False(t, hasBlob(t, otherRuns, "run-lint-old.json"))
	assert.True(t, hasBlob(t, otherRuns, "run-lint-fresh.json"))

	assert.Equal(t, 1, countRows(t, db, `SELECT count(*) FROM linter_executions`))

	var sandbox []uuid.UUID
	require.NoError(t, db.Raw(`SELECT id FROM sandbox_runs ORDER BY created_at, id`).Scan(&sandbox).Error)
	assert.ElementsMatch(t, []uuid.UUID{freshRun, unfinished}, sandbox, "an unfinished run stays")
	assert.False(t, hasBlob(t, orgStore(other), "sandbox/"+oldRun.String()+".log"))
	assert.True(t, hasBlob(t, orgStore(other), "sandbox/"+freshRun.String()+".log"))

	assert.False(t, hasBlob(t, recordings, oldRecording))
	assert.True(t, hasBlob(t, recordings, freshRecording))
	assert.True(t, hasBlob(t, recordings, "workspaces/abc/run-old.json"), "only recordings are aged by modification time")
}

func TestRunReportsAnInvalidWindow(t *testing.T) {
	db := openRetentionDB(t)
	setProperty(t, "retention.job_logs", "two months")
	_, err := Run(context.Background(), db, Stores{}, time.Now())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "retention.job_logs=two months")
}

func TestRunHonoursConfiguredWindows(t *testing.T) {
	db := openRetentionDB(t)
	setProperty(t, "retention.linter_executions", "1h")
	now := time.Now().UTC()
	require.NoError(t, db.Exec(`INSERT INTO linter_executions (linter_name, work_dir, executed_at, duration_ms, violation_count, success) VALUES ('x', '/w', ?, 1, 0, true)`, now.Add(-2*time.Hour)).Error)

	report, err := Run(context.Background(), db, Stores{}, now)
	require.NoError(t, err)
	assert.Equal(t, 1, report.LinterExecutions, "two hours is past a one-hour window")
}

func TestRunReportsAMissingBlobStore(t *testing.T) {
	db := openRetentionDB(t)
	require.NoError(t, db.Exec(`INSERT INTO job_log_caches (org_id, job_id, repo, log_key, fetched_at) VALUES (?, 1, 'a/b', 'joblogs/1.log.gz', ?)`,
		tenant.LocalOrgID, time.Now().Add(-90*24*time.Hour)).Error)

	_, err := Run(context.Background(), db, Stores{Org: func(uuid.UUID) blob.Store { return nil }}, time.Now())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no blob store for organization "+tenant.LocalOrgID.String())
	assert.Equal(t, 1, countRows(t, db, `SELECT count(*) FROM job_log_caches`), "the row stays with its blob")
}
