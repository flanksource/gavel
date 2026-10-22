// Tests that a local store stays inside its root: a symlink planted under it can't redirect a
// read, write, stat or delete to a file outside.
package blob

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalRefusesSymlinksThatEscapeItsRoot(t *testing.T) {
	ctx := context.Background()
	root, outside := t.TempDir(), t.TempDir()
	secret := filepath.Join(outside, "secret.json")
	require.NoError(t, os.WriteFile(secret, []byte(`{"secret":true}`), 0o644))
	require.NoError(t, os.Symlink(secret, filepath.Join(root, "run-link.json")))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "dir-link")))
	store := Local(root)

	_, err := store.Read(ctx, "run-link.json")
	require.Error(t, err, "read through a file symlink")
	_, err = store.Read(ctx, "dir-link/secret.json")
	require.Error(t, err, "read through a directory symlink")
	exists, err := store.Exists(ctx, "dir-link/secret.json")
	assert.False(t, exists && err == nil, "stat through a directory symlink")
	require.Error(t, store.Write(ctx, "dir-link/planted.json", strings.NewReader("x")), "write through a directory symlink")
	assert.NoFileExists(t, filepath.Join(outside, "planted.json"))
	_ = store.Delete(ctx, "dir-link/secret.json")
	assert.FileExists(t, secret, "delete through a directory symlink")
}
