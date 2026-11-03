// Tests that migrating moves job logs still inline in job_log_caches.logs_gz to the blob store,
// under the key the GitHub cache reads them by, before the column is dropped.
package database_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	commonsdbtest "github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationMovesInlineJobLogsToTheBlobStore(t *testing.T) {
	t.Setenv(database.EnvDSN, commonsdbtest.ForT(t, commonsdbtest.Options{Name: "gavel_inline_job_logs"}).DSN())
	t.Setenv(database.EnvDisable, "")
	t.Setenv(database.LegacyEnvDSN, "")
	t.Setenv(database.LegacyEnvDisable, "")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	// macOS ignores XDG_CACHE_HOME and uses $HOME/Library/Caches.
	cacheDir, err := os.UserCacheDir()
	require.NoError(t, err)
	ctx := t.Context()

	migrated, err := database.Open(ctx, database.WithMigrations())
	require.NoError(t, err)
	// The table as a release that kept job logs inline left it.
	var gz bytes.Buffer
	writer := gzip.NewWriter(&gz)
	_, err = writer.Write([]byte("inline logs"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	require.NoError(t, migrated.Gorm().Exec(`ALTER TABLE job_log_caches ADD COLUMN IF NOT EXISTS logs_gz bytea`).Error)
	require.NoError(t, migrated.Gorm().Exec(`INSERT INTO job_log_caches (org_id, job_id, repo, logs_gz, fetched_at) VALUES (?, 7, 'a/b', ?, now())`,
		tenant.LocalOrgID, gz.Bytes()).Error)
	require.NoError(t, migrated.Close())

	reopened, err := database.Open(ctx, database.WithMigrations())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, reopened.Close()) })
	db := reopened.Gorm()

	var key string
	require.NoError(t, db.Raw(`SELECT log_key FROM job_log_caches WHERE org_id = ? AND job_id = 7`, tenant.LocalOrgID).Scan(&key).Error)
	assert.Equal(t, "joblogs/7.log.gz", key, "the row names its logs in the blob store")
	var columns int64
	require.NoError(t, db.Raw(`SELECT count(*) FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'job_log_caches' AND column_name = 'logs_gz'`).Scan(&columns).Error)
	assert.Zero(t, columns, "logs_gz is dropped")

	stored, err := os.Open(filepath.Join(cacheDir, "gavel", "blobs", "orgs", tenant.LocalOrgID.String(), "joblogs", "7.log.gz"))
	require.NoError(t, err)
	defer stored.Close()
	reader, err := gzip.NewReader(stored)
	require.NoError(t, err)
	logs, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "inline logs", string(logs))
}
