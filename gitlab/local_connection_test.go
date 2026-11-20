// Tests the connection a local gavel reaches a GitLab project through: the git connection owning
// it, else GITLAB_TOKEN on gitlab.com; a self-managed host needs a connection.
package gitlab

import (
	"context"
	"errors"
	"testing"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// owning makes connection, or err, the lookup's answer, with GITLAB_TOKEN set to token.
func owning(t *testing.T, connection *models.Connection, err error, token string) {
	t.Helper()
	previous := FindConnection
	FindConnection = func(context.Context, git.RepoURL) (*models.Connection, error) { return connection, err }
	t.Cleanup(func() { FindConnection = previous })
	t.Setenv("GITLAB_TOKEN", token)
}

func project(t *testing.T, raw string) git.RepoURL {
	t.Helper()
	repo, err := git.ParseRepoURL(raw, models.ConnectionTypeGitlab)
	require.NoError(t, err)
	return repo
}

func TestALocalGitLabProjectIsReachedThroughTheConnectionThatOwnsIt(t *testing.T) {
	owner := models.Connection{Name: "acme", Type: models.ConnectionTypeGitlab, URL: "https://gitlab.example.com", Password: "connection-token"}
	owning(t, &owner, nil, "env-token")
	connection, err := LocalConnection(context.Background(), project(t, "https://gitlab.example.com/acme/api.git"))
	require.NoError(t, err)
	assert.Equal(t, owner, connection)
}

func TestALocalGitLabDotComProjectCarriesGITLAB_TOKENWhenNoConnectionOwnsIt(t *testing.T) {
	owning(t, nil, nil, "env-token")
	connection, err := LocalConnection(context.Background(), project(t, "https://gitlab.com/acme/platform/api.git"))
	require.NoError(t, err)
	assert.Equal(t, models.Connection{Name: "GITLAB_TOKEN", Type: models.ConnectionTypeGitlab, URL: "https://gitlab.com", Password: "env-token"}, connection)

}

func TestALocalGitLabProjectWithoutACredentialIsRefusedWithWhatWouldReachIt(t *testing.T) {
	owning(t, nil, nil, "env-token")
	_, err := LocalConnection(context.Background(), project(t, "https://gitlab.example.com/acme/api.git"))
	require.EqualError(t, err, "no git connection owns GitLab project https://gitlab.example.com/acme/api: a self-managed GitLab is reached only through a gitlab connection for gitlab.example.com")

	owning(t, nil, nil, "")
	_, err = LocalConnection(context.Background(), project(t, "https://gitlab.com/acme/api.git"))
	require.EqualError(t, err, "no git connection owns GitLab project https://gitlab.com/acme/api and GITLAB_TOKEN is not set")
}

func TestALocalGitLabProjectsConnectionMustHoldAToken(t *testing.T) {
	owning(t, &models.Connection{Name: "acme", Type: models.ConnectionTypeGitlab}, nil, "env-token")
	_, err := LocalConnection(context.Background(), project(t, "https://gitlab.com/acme/api.git"))
	require.EqualError(t, err, `git connection "acme" owns GitLab project https://gitlab.com/acme/api but holds no personal access token`)

	owning(t, nil, errors.New("the catalog's key is unavailable"), "env-token")
	_, err = LocalConnection(context.Background(), project(t, "https://gitlab.com/acme/api.git"))
	require.EqualError(t, err, "find the git connection for GitLab project https://gitlab.com/acme/api: the catalog's key is unavailable")
}
