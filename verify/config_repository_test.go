// Tests LoadRepositoryGavelConfig: a hosted repository's settings come from its own .gavel.yaml
// alone, never from the serve host's home config.
package verify

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRepositoryGavelConfigReadsOnlyTheRepositorysConfig(t *testing.T) {
	home, repo := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("ssh:\n  cmd: make deploy-from-the-host\nstorage:\n  path: /var/lib/gavel/blobs\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, ".gavel.yaml"), []byte("ssh:\n  cmd: make check\n"), 0o644))

	cfg, err := LoadRepositoryGavelConfig(repo)
	require.NoError(t, err)
	assert.Equal(t, "make check", cfg.SSH.Cmd)
	assert.True(t, cfg.Storage.IsZero(), "nothing comes from the host's home config")
	assert.Equal(t, DefaultGavelConfig().AI, cfg.AI, "the defaults still apply")
}

func TestLoadRepositoryGavelConfigWithoutOne(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg, err := LoadRepositoryGavelConfig(t.TempDir())
	require.NoError(t, err)
	assert.Equal(t, DefaultGavelConfig(), cfg)
}
