// Tests OpenGlobalStorage and OpenWorkspaceStorage: no store without a storage section, the
// configured store with one, and a descriptive error for an incomplete one.
package verify

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/blob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenGlobalStorage(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeHomeConfig := func(content string) {
		t.Helper()
		require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte(content), 0o644))
	}

	store, err := OpenGlobalStorage(context.Background())
	require.NoError(t, err)
	assert.Nil(t, store, "no storage section keeps today's directories")

	bucket := t.TempDir()
	writeHomeConfig("storage:\n  path: " + bucket + "\n")
	store, err = OpenGlobalStorage(context.Background())
	require.NoError(t, err)
	require.NotNil(t, store)
	require.NoError(t, store.Write(context.Background(), "probe.txt", strings.NewReader("ok")))
	assert.FileExists(t, filepath.Join(bucket, "probe.txt"))

	writeHomeConfig("storage:\n  type: s3\n")
	_, err = OpenGlobalStorage(context.Background())
	require.EqualError(t, err, "open the storage in "+filepath.Join(home, ".gavel.yaml")+": storage.bucket is required for s3 storage")
}

func TestOpenWorkspaceStorage(t *testing.T) {
	home, bucket, workDir := t.TempDir(), t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	store, err := OpenWorkspaceStorage(context.Background(), workDir)
	require.NoError(t, err)
	assert.Nil(t, store)

	require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("storage:\n  path: "+bucket+"\n"), 0o644))
	store, err = OpenWorkspaceStorage(context.Background(), workDir)
	require.NoError(t, err)
	require.NoError(t, store.Write(context.Background(), "probe.txt", strings.NewReader("ok")))
	prefix, err := blob.WorkspacePrefix(workDir)
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(bucket, filepath.FromSlash(prefix), "probe.txt"))
}
