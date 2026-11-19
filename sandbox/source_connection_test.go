// Tests the mirrors of repositories git connections bring: each is fetched with its connection's
// remote, kept under the repository's id, follows a renamed remote, and replaces the mirrors
// once named by GitHub repository id.
package sandbox

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	cdbgit "github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAConnectionsRepositoryIsMirroredThroughItsConnection(t *testing.T) {
	dataDir, remotes := t.TempDir(), t.TempDir()
	work := newWork(t)
	commitFile(t, work, "README.md", "first")
	git(t, work, "clone", "-q", "--bare", work, filepath.Join(remotes, "api.git"))
	git(t, work, "clone", "-q", "--bare", work, filepath.Join(remotes, "api-v2.git"))
	connection := uuid.New()
	var asked []string
	sources := &Sources{DataDir: dataDir, Connection: func(_ context.Context, repo Repo) (*cdbgit.Repository, error) {
		asked = append(asked, repo.RemoteURL)
		return cdbgit.NewRepository(cdbgit.RepositoryOptions{URL: repo.RemoteURL, ProviderType: models.ConnectionTypeGitlab})
	}}
	gitlab := Repo{ID: uuid.New(), Org: uuid.New(), Source: "gitlab", Name: "acme/api", ConnectionID: &connection, RemoteURL: "file://" + filepath.Join(remotes, "api.git")}

	var bundle bytes.Buffer
	_, err := sources.Bundle(context.Background(), gitlab, "refs/heads/main", &bundle)
	require.NoError(t, err)
	assert.Equal(t, "first", unbundle(t, bundle.Bytes(), "README.md"))
	mirror := filepath.Join(dataDir, "mirrors", gitlab.Org.String(), gitlab.ID.String()+".git")
	assert.DirExists(t, mirror, "kept under the repository's id")

	// The project is renamed: the mirror fetches from where the connection now lists it.
	second := commitFile(t, work, "README.md", "second")
	git(t, work, "push", "-q", filepath.Join(remotes, "api-v2.git"), "main")
	gitlab.RemoteURL = "file://" + filepath.Join(remotes, "api-v2.git")
	bundle.Reset()
	got, err := sources.Bundle(context.Background(), gitlab, "refs/heads/main", &bundle)
	require.NoError(t, err)
	assert.Equal(t, second, got)
	assert.Equal(t, []string{"file://" + filepath.Join(remotes, "api.git"), gitlab.RemoteURL}, asked)
}

func TestARepositoryWithoutItsConnectionIsNotMirrored(t *testing.T) {
	sources := &Sources{DataDir: t.TempDir()}
	orphan := Repo{ID: uuid.New(), Org: uuid.New(), Source: "gitlab", Name: "acme/api"}
	_, err := sources.Bundle(context.Background(), orphan, "refs/heads/main", &bytes.Buffer{})
	require.EqualError(t, err, "repository acme/api comes from no git connection any more; add one that brings it")

	connection := uuid.New()
	orphan.ConnectionID = &connection
	_, err = sources.Bundle(context.Background(), orphan, "refs/heads/main", &bytes.Buffer{})
	require.EqualError(t, err, "repository acme/api comes from a git connection, which this server doesn't serve")
}

func TestMirrorsNamedByGitHubRepositoryIDAreRemoved(t *testing.T) {
	dataDir := t.TempDir()
	org := uuid.New().String()
	byGitHubID := filepath.Join(dataDir, "mirrors", org, "42.git")
	byRepo := filepath.Join(dataDir, "mirrors", org, uuid.NewString()+".git")
	for _, dir := range []string{byGitHubID, byRepo} {
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "objects"), 0o755))
	}

	require.NoError(t, (&Sources{DataDir: dataDir}).RemoveGitHubIDMirrors())
	assert.NoDirExists(t, byGitHubID)
	assert.DirExists(t, byRepo)
	require.NoError(t, (&Sources{DataDir: t.TempDir()}).RemoveGitHubIDMirrors(), "a server with no mirrors yet")
}
