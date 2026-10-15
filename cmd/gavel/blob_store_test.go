// Tests configuredBlobStore: no store without a storage section, the configured store with one,
// and a descriptive error for an incomplete one.
package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfiguredBlobStore(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeHomeConfig := func(content string) {
		t.Helper()
		require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte(content), 0o644))
	}

	store, err := configuredBlobStore(context.Background())
	require.NoError(t, err)
	assert.Nil(t, store, "no storage section keeps today's directories")

	bucket := t.TempDir()
	writeHomeConfig("storage:\n  path: " + bucket + "\n")
	store, err = configuredBlobStore(context.Background())
	require.NoError(t, err)
	require.NotNil(t, store)
	require.NoError(t, store.Write(context.Background(), "probe.txt", strings.NewReader("ok")))
	assert.FileExists(t, filepath.Join(bucket, "probe.txt"))

	writeHomeConfig("storage:\n  type: s3\n")
	_, err = configuredBlobStore(context.Background())
	require.EqualError(t, err, "open the storage in "+filepath.Join(home, ".gavel.yaml")+": storage.bucket is required for s3 storage")
}
