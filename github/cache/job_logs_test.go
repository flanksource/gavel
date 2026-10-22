// Tests that job logs live in the blob store with only their key in job_log_caches, that a
// legacy row holding the logs inline is moved to the store on first read, and that another
// organization can't read them.
package cache

import (
	"context"
	"io"
	"testing"

	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type jobLogRow struct {
	LogKey string
	LogsGz []byte
}

func readJobLogRow(t *testing.T, s *Store, jobID int64) jobLogRow {
	t.Helper()
	var row jobLogRow
	require.NoError(t, s.gorm().Raw(`SELECT log_key, logs_gz FROM job_log_caches WHERE org_id = ? AND job_id = ?`, s.org, jobID).Scan(&row).Error)
	return row
}

func blobExists(t *testing.T, store blob.Store, key string) bool {
	t.Helper()
	exists, err := store.Exists(context.Background(), key)
	require.NoError(t, err)
	return exists
}

func TestJobLogsLiveInTheBlobStore(t *testing.T) {
	local, other := openTenancyStores(t)
	blobs := blob.Local(t.TempDir())
	local.blobs, other.blobs = blobs, blobs

	local.PutJobLogs(41, "a/b", "step 1\nstep 2\n")

	key := "orgs/" + tenant.LocalOrgID.String() + "/joblogs/41.log.gz"
	assert.True(t, blobExists(t, blobs, key))
	row := readJobLogRow(t, local, 41)
	assert.Equal(t, "joblogs/41.log.gz", row.LogKey)
	assert.Nil(t, row.LogsGz, "the row keeps only the key")

	logs, ok := local.GetJobLogs(41)
	require.True(t, ok)
	assert.Equal(t, "step 1\nstep 2\n", logs)
	_, ok = other.GetJobLogs(41)
	assert.False(t, ok, "another organization can't read them")
}

func TestLegacyJobLogsMoveToTheBlobStoreOnRead(t *testing.T) {
	local, _ := openTenancyStores(t)
	blobs := blob.Local(t.TempDir())
	local.blobs = blobs
	gz, err := gzipString("legacy logs")
	require.NoError(t, err)
	require.NoError(t, local.gorm().Exec(`INSERT INTO job_log_caches (org_id, job_id, repo, logs_gz, fetched_at) VALUES (?, 7, 'a/b', ?, now())`, tenant.LocalOrgID, gz).Error)

	logs, ok := local.GetJobLogs(7)
	require.True(t, ok)
	assert.Equal(t, "legacy logs", logs)

	row := readJobLogRow(t, local, 7)
	assert.Equal(t, "joblogs/7.log.gz", row.LogKey)
	assert.Nil(t, row.LogsGz)
	reader, err := blobs.Read(context.Background(), "orgs/"+tenant.LocalOrgID.String()+"/joblogs/7.log.gz")
	require.NoError(t, err)
	defer reader.Close()
	stored, err := io.ReadAll(reader)
	require.NoError(t, err)
	unzipped, err := gunzipString(stored)
	require.NoError(t, err)
	assert.Equal(t, "legacy logs", unzipped)

	logs, ok = local.GetJobLogs(7)
	require.True(t, ok, "and reads from the store afterwards")
	assert.Equal(t, "legacy logs", logs)
}

func TestOrgBlobsIsWhereJobLogsLive(t *testing.T) {
	local, _ := openTenancyStores(t)
	local.blobs = blob.Local(t.TempDir())
	local.PutJobLogs(5, "a/b", "logs")

	assert.True(t, blobExists(t, local.OrgBlobs(tenant.LocalOrgID), "joblogs/5.log.gz"))
}
