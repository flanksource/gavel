package service

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAuthConfig_MissingFileReturnsZero(t *testing.T) {
	withTempHome(t)
	cfg, err := LoadAuthConfig()
	require.NoError(t, err)
	assert.Empty(t, cfg.Token)
}

func TestSaveLoadAuthConfig_Roundtrip(t *testing.T) {
	withTempHome(t)
	require.NoError(t, SaveAuthConfig(AuthConfig{Token: "ghp_example"}))

	out, err := LoadAuthConfig()
	require.NoError(t, err)
	assert.Equal(t, "ghp_example", out.Token)
}

func TestSaveAuthConfig_WritesUserOnlyPerms(t *testing.T) {
	withTempHome(t)
	require.NoError(t, SaveAuthConfig(AuthConfig{Token: "ghp_example"}))
	path, err := AuthConfigPath()
	require.NoError(t, err)
	info, err := os.Stat(path)
	require.NoError(t, err)
	// The whole point of this file is to hold a credential — 0600 keeps
	// other users on a shared machine from reading it.
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

func TestSaveAuthConfig_RejectsEmptyToken(t *testing.T) {
	withTempHome(t)
	assert.Error(t, SaveAuthConfig(AuthConfig{}))
}

func TestDiscoverGitHubToken_PrefersGitHubTokenEnv(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "from-github-token")
	t.Setenv("GH_TOKEN", "from-gh-token")
	// Force PATH empty so `gh auth token` can never win.
	t.Setenv("PATH", "")

	tok, src, err := DiscoverGitHubToken()
	require.NoError(t, err)
	assert.Equal(t, "from-github-token", tok)
	assert.Equal(t, TokenSourceEnvGitHub, src)
}

func TestDiscoverGitHubToken_FallsBackToGHToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "from-gh-token")
	t.Setenv("PATH", "")

	tok, src, err := DiscoverGitHubToken()
	require.NoError(t, err)
	assert.Equal(t, "from-gh-token", tok)
	assert.Equal(t, TokenSourceEnvGH, src)
}

func TestDiscoverGitHubToken_NoneReturnsEmpty(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	t.Setenv("PATH", "") // no gh CLI reachable

	tok, src, err := DiscoverGitHubToken()
	require.NoError(t, err)
	assert.Empty(t, tok)
	assert.Empty(t, string(src))
}
