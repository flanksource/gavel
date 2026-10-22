// Tests LoadGlobalGavelConfig: it reads the home .gavel.yaml alone, so settings only the serve
// host may choose, such as storage, can't come from a repository's config.
package verify

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/blob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadGlobalGavelConfigReadsOnlyTheHomeConfig(t *testing.T) {
	home, repo := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.Mkdir(filepath.Join(repo, ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("storage:\n  path: /var/lib/gavel/blobs\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".gavel.yaml"), []byte("storage:\n  path: /tmp/stolen\n"), 0o644))
	t.Chdir(repo)

	global, err := LoadGlobalGavelConfig()
	require.NoError(t, err)
	assert.Equal(t, blob.Config{Path: "/var/lib/gavel/blobs"}, global.Storage)

	layered, err := LoadGavelConfig(repo)
	require.NoError(t, err)
	assert.Equal(t, "/tmp/stolen", layered.Storage.Path, "the layered loader lets the repository override, which is why storage reads the global one")
}

func TestLoadGlobalGavelConfigWithoutAHomeConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	global, err := LoadGlobalGavelConfig()
	require.NoError(t, err)
	assert.True(t, global.Storage.IsZero())
}

func TestLoadGlobalGavelConfigReportsAnInvalidHomeConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("storage: [not, an, object]\n"), 0o644))
	_, err := LoadGlobalGavelConfig()
	require.Error(t, err)
	assert.Contains(t, err.Error(), filepath.Join(home, ".gavel.yaml"))
}
