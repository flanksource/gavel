// Tests that a run's recordings, sidecars included, are copied to the configured blob store under
// recordings/, and that without one they stay only in .gavel/recordings.
package record

import (
	"context"
	"io"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/gavel/blob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersistCopiesTheRunsRecordings(t *testing.T) {
	workDir := t.TempDir()
	store := NewStore(workDir, time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC))
	file, result, err := store.Create("login flow", KindHTTP)
	require.NoError(t, err)
	_, err = file.WriteString(`{"log":{}}`)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	sidecar, sidecarPath, err := store.CreateSidecar("login flow", KindHTTP, "ca.pem")
	require.NoError(t, err)
	_, err = sidecar.WriteString("CERT")
	require.NoError(t, err)
	require.NoError(t, sidecar.Close())

	bucket := blob.Local(t.TempDir())
	require.NoError(t, store.Persist(context.Background(), bucket))

	read := func(key string) string {
		t.Helper()
		reader, err := bucket.Read(context.Background(), key)
		require.NoError(t, err, key)
		defer reader.Close()
		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		return string(data)
	}
	assert.Equal(t, `{"log":{}}`, read("recordings/"+filepath.Base(result.Path)))
	assert.Equal(t, "CERT", read("recordings/"+filepath.Base(sidecarPath)))
	assert.FileExists(t, result.Path, "the local copy stays for this run's viewers")
}

func TestPersistWithoutAStoreIsANoOp(t *testing.T) {
	store := NewStore(t.TempDir(), time.Now())
	file, _, err := store.Create("x", KindANSI)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	assert.NoError(t, store.Persist(context.Background(), nil))
}
