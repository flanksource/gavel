// Tests listing the open pull requests of an organization's repositories that come from git
// connections whose pull requests the dashboard lists: each through its own connection.
package forge_test

import (
	"context"
	"testing"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/forge"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAnOrgsAzureRepositoriesListTheirOpenPullRequests(t *testing.T) {
	db := catalog(t)
	scope := hostedOrg(t, db)
	azure := fakeListing(t, "/acme/Platform/_apis/git/repositories/rules/pullrequests")
	azure.set(`{"value":[{"pullRequestId":9,"status":"active","title":"Tighten the rules","creationDate":"2026-11-02T10:00:00Z",
		"sourceRefName":"refs/heads/tighten","targetRefName":"refs/heads/main","isDraft":true,"mergeStatus":"succeeded",
		"createdBy":{"displayName":"Ada Lovelace","imageUrl":"https://avatars.example/ada"},
		"lastMergeSourceCommit":{"commitId":"1111111111111111111111111111111111111111"},
		"repository":{"name":"rules","project":{"name":"Platform"}}}]}`)
	conn := saveFor(t, db, scope, models.Connection{Name: "acme-azure", Type: models.ConnectionTypeAzureDevops, URL: azure.URL + "/acme",
		Password: "azure-pat", InsecureTLS: true})
	connectionRepo(t, db, scope, "Platform/rules", "azure_devops", &conn, azure.URL+"/acme/Platform/_git/rules")
	gitlab := saveFor(t, db, scope, models.Connection{Name: "acme-gitlab", Type: models.ConnectionTypeGitlab, URL: "https://gitlab.example.com",
		Password: "glpat", Properties: map[string]string{"group": "acme"}})
	connectionRepo(t, db, scope, "acme/api", "gitlab", &gitlab, "https://gitlab.example.com/acme/api.git")
	connectionRepo(t, db, scope, "Platform/released", "azure_devops", nil, azure.URL+"/acme/Platform/_git/released")

	listed, err := forge.PullRequests(context.Background(), db, scope, git.ClientOptions{})
	require.NoError(t, err)
	require.Len(t, listed, 1, "only a repository whose connection lists its pull requests")
	assert.Equal(t, "acme", listed[0].Repo.Owner)
	assert.Equal(t, "Platform", listed[0].Repo.Project)
	assert.Equal(t, "rules", listed[0].Repo.Name)
	pr := listed[0].PullRequest
	assert.Equal(t, 9, pr.Number)
	assert.Equal(t, "Tighten the rules", pr.Title)
	assert.Equal(t, git.PullRequestOpen, pr.State)
	assert.Equal(t, azure.URL+"/acme/Platform/_git/rules/pullrequest/9", pr.URL)
	require.NotNil(t, pr.Metadata)
	assert.True(t, pr.Metadata.Draft)

	other := hostedOrg(t, db)
	listed, err = forge.PullRequests(context.Background(), db, other, git.ClientOptions{})
	require.NoError(t, err)
	assert.Empty(t, listed, "another organization lists none of them")
}

func TestARepositoryWhosePullRequestsCantBeListedIsReported(t *testing.T) {
	db := catalog(t)
	scope := hostedOrg(t, db)
	azure := fakeListing(t, "/acme/Platform/_apis/git/repositories/rules/pullrequests")
	azure.set(`{"value":[]}`)
	conn := saveFor(t, db, scope, models.Connection{Name: "acme-azure", Type: models.ConnectionTypeAzureDevops, URL: azure.URL + "/acme",
		Password: "azure-pat", InsecureTLS: true})
	connectionRepo(t, db, scope, "Platform/rules", "azure_devops", &conn, azure.URL+"/acme/Platform/_git/rules")
	connectionRepo(t, db, scope, "Platform/gone", "azure_devops", &conn, azure.URL+"/acme/Platform/_git/gone")

	listed, err := forge.PullRequests(context.Background(), db, scope, git.ClientOptions{})
	require.ErrorContains(t, err, "list the pull requests of Platform/gone")
	assert.Empty(t, listed)
}

// connectionRepo records a repository of scope's that connection brings from remoteURL; a nil
// connection is one a deleted connection left behind.
func connectionRepo(t *testing.T, db *gorm.DB, scope tenant.Scope, name, source string, connection *uuid.UUID, remoteURL string) {
	t.Helper()
	require.NoError(t, db.Exec(`INSERT INTO repos (org_id, name, source, connection_id, remote_url, provider_repo_id, default_branch)
		VALUES (?, ?, ?, ?, ?, ?, 'main')`, scope.OrgID, name, source, connection, remoteURL, uuid.NewString()).Error)
}

func TestAPullRequestIsReachedThroughTheConnectionOfItsOrgsRepository(t *testing.T) {
	db := catalog(t)
	scope, other := hostedOrg(t, db), hostedOrg(t, db)
	azure := fakeListing(t, "/acme/Platform/_apis/git/repositories/rules/pullrequests/9")
	conn := saveFor(t, db, scope, models.Connection{Name: "acme-azure", Type: models.ConnectionTypeAzureDevops, URL: azure.URL + "/acme",
		Password: "azure-pat", InsecureTLS: true})
	connectionRepo(t, db, scope, "Platform/rules", "azure_devops", &conn, azure.URL+"/acme/Platform/_git/rules")
	pr := azure.URL + "/acme/Platform/_git/rules/pullrequest/9"

	remote, err := forge.PullRequestRemote(context.Background(), db, scope, pr, git.ClientOptions{})
	require.NoError(t, err)
	assert.Equal(t, "rules", remote.URL.Name)
	assert.Equal(t, "Platform", remote.URL.Project)
	require.NotNil(t, remote.Provider)
	held, err := forge.PullRequestRepo(context.Background(), db, scope, pr, git.ClientOptions{})
	require.NoError(t, err)
	assert.Equal(t, "main", held.DefaultBranch, "the default branch its sync recorded")
	assert.Equal(t, "rules", held.Remote.URL.Name)

	_, err = forge.PullRequestRemote(context.Background(), db, other, pr, git.ClientOptions{})
	require.EqualError(t, err, "no repository of the organization's git connections holds pull request "+pr)
	_, err = forge.PullRequestRemote(context.Background(), db, scope, azure.URL+"/acme/Platform/_git/other/pullrequest/9", git.ClientOptions{})
	require.ErrorContains(t, err, "no repository of the organization's git connections holds")
	_, err = forge.PullRequestRemote(context.Background(), db, scope, "https://github.com/acme/api/pull/7", git.ClientOptions{})
	require.EqualError(t, err, `pull request URL "https://github.com/acme/api/pull/7" has no /pullrequest/ path`)
}

func TestAPullRequestURLNamesItsRepository(t *testing.T) {
	repo, err := forge.PullRequestRepository("https://dev.azure.com/acme/Platform/_git/rules/pullrequest/9")
	require.NoError(t, err)
	assert.Equal(t, "https://dev.azure.com/acme/Platform/_git/rules", repo)
}
