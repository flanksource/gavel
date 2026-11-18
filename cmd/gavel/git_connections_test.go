// Tests that a local gavel reads repository tokens from its working directory's git connections,
// and that a hosted server never reads the host's.
package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// connectionFiles makes the working directory hold a github connection for acme's repositories,
// with no other GitHub credential on the host.
func connectionFiles(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".connections"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".connections", "acme.yaml"), []byte("apiVersion: mission-control.flanksource.com/v1\nkind: Connection\n"+
		"metadata: {name: acme}\nspec: {properties: {organization: acme}, github: {personalAccessToken: {value: gh_file}}}\n"), 0o600))
	previous := workingDir
	workingDir = dir
	t.Cleanup(func() { workingDir = previous })
}

func TestALocalRunReadsRepositoryTokensFromItsConnections(t *testing.T) {
	connectionFiles(t)
	t.Setenv("GAVEL_DB", "off")
	client, err := github.GitClient(context.Background(), github.Options{Repo: "acme/api"})
	require.NoError(t, err)
	assert.Equal(t, "gh_file", client.Auth.Token)
}

func TestAHostedServerNeverReadsTheHostsConnections(t *testing.T) {
	connectionFiles(t)
	newHostedFixture(t)
	_, err := github.GitClient(context.Background(), github.Options{Repo: "acme/api"})
	require.ErrorIs(t, err, github.ErrNoToken)
}
