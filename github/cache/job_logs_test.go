// Tests that job logs live in the blob store with only their key in job_log_caches, and that
// another organization can't read them.
package cache

import (
	"context"
	"testing"

	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readJobLogKey(t *testing.T, s *Store, jobID int64) string {
	t.Helper()
	var key string
	require.NoError(t, s.gorm().Raw(`SELECT log_key FROM job_log_caches WHERE org_id = ? AND job_id = ?`, s.org, jobID).Scan(&key).Error)
	return key
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
	assert.Equal(t, "joblogs/41.log.gz", readJobLogKey(t, local, 41), "the row keeps only the key")

	logs, ok := local.GetJobLogs(41)
	require.True(t, ok)
	assert.Equal(t, "step 1\nstep 2\n", logs)
	_, ok = other.GetJobLogs(41)
	assert.False(t, ok, "another organization can't read them")
}

func TestOrgBlobsIsWhereJobLogsLive(t *testing.T) {
	local, _ := openTenancyStores(t)
	local.blobs = blob.Local(t.TempDir())
	local.PutJobLogs(5, "a/b", "logs")

	assert.True(t, blobExists(t, local.OrgBlobs(tenant.LocalOrgID), "joblogs/5.log.gz"))
}
