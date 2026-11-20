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

func TestAnOrgsAzureAndGitLabRepositoriesListTheirOpenPullRequests(t *testing.T) {
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
	gitlabAPI := fakeListing(t, "/api/v4/projects/acme/api/merge_requests")
	gitlabAPI.set(`[{"iid":5,"state":"opened","title":"Speed up the API","web_url":"` + gitlabAPI.URL + `/acme/api/-/merge_requests/5",
		"source_branch":"fast","target_branch":"main","source_project_id":1,"target_project_id":1,"sha":"2222222222222222222222222222222222222222",
		"author":{"name":"Grace Hopper"},"created_at":"2026-11-03T10:00:00Z"}]`)
	gitlab := saveFor(t, db, scope, models.Connection{Name: "acme-gitlab", Type: models.ConnectionTypeGitlab, URL: gitlabAPI.URL,
		Password: "glpat", InsecureTLS: true, Properties: map[string]string{"group": "acme"}})
	connectionRepo(t, db, scope, "acme/api", "gitlab", &gitlab, gitlabAPI.URL+"/acme/api.git")
	connectionRepo(t, db, scope, "Platform/released", "azure_devops", nil, azure.URL+"/acme/Platform/_git/released")
	connectionRepo(t, db, scope, "acme/plain", "git", &gitlab, gitlabAPI.URL+"/acme/plain.git")

	listed, err := forge.PullRequests(context.Background(), db, scope, git.ClientOptions{})
	require.NoError(t, err)
	byProvider := map[string]forge.RepoPullRequest{}
	for _, entry := range listed {
		byProvider[entry.Repo.Provider] = entry
	}
	require.Len(t, listed, 2, "each repository whose connection lists its pull requests: %v", listed)
	mr := byProvider[models.ConnectionTypeGitlab]
	assert.Equal(t, "acme", mr.Repo.Owner)
	assert.Equal(t, "api", mr.Repo.Name)
	assert.Equal(t, 5, mr.PullRequest.Number)
	assert.Equal(t, "Speed up the API", mr.PullRequest.Title)
	assert.Equal(t, gitlabAPI.URL+"/acme/api/-/merge_requests/5", mr.PullRequest.URL)
	azurePR := byProvider[models.ConnectionTypeAzureDevops]
	assert.Equal(t, "acme", azurePR.Repo.Owner)
	assert.Equal(t, "Platform", azurePR.Repo.Project)
	assert.Equal(t, "rules", azurePR.Repo.Name)
	pr := azurePR.PullRequest
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
	require.ErrorIs(t, err, forge.ErrNotConnectionPullRequest)
}

func TestAMergeRequestIsReachedThroughTheConnectionOfItsOrgsGitLabRepository(t *testing.T) {
	db := catalog(t)
	scope := hostedOrg(t, db)
	gitlabAPI := fakeListing(t, "/api/v4/projects/acme/platform/api/merge_requests/5")
	conn := saveFor(t, db, scope, models.Connection{Name: "acme-gitlab", Type: models.ConnectionTypeGitlab, URL: gitlabAPI.URL,
		Password: "glpat", InsecureTLS: true})
	connectionRepo(t, db, scope, "acme/platform/api", "gitlab", &conn, gitlabAPI.URL+"/acme/platform/api.git")

	held, err := forge.PullRequestRepo(context.Background(), db, scope, gitlabAPI.URL+"/acme/platform/api/-/merge_requests/5", git.ClientOptions{})
	require.NoError(t, err)
	assert.Equal(t, "acme/platform", held.Remote.URL.Owner, "a nested group")
	assert.Equal(t, "api", held.Remote.URL.Name)
	assert.Equal(t, models.ConnectionTypeGitlab, held.Remote.URL.Provider)
	assert.Equal(t, "main", held.DefaultBranch)

	_, err = forge.PullRequestRepo(context.Background(), db, scope, gitlabAPI.URL+"/acme/platform/web/-/merge_requests/5", git.ClientOptions{})
	require.ErrorContains(t, err, "no repository of the organization's git connections holds")
}

func TestAPullRequestURLNamesItsRepositoryProviderAndNumber(t *testing.T) {
	pr, err := forge.ParsePullRequestURL("https://dev.azure.com/acme/Platform/_git/rules/pullrequest/9")
	require.NoError(t, err)
	assert.Equal(t, forge.PullRequestURL{Repo: "https://dev.azure.com/acme/Platform/_git/rules", Provider: models.ConnectionTypeAzureDevops, Number: 9}, pr)

	pr, err = forge.ParsePullRequestURL("https://gitlab.com/acme/platform/api/-/merge_requests/5/diffs")
	require.NoError(t, err)
	assert.Equal(t, forge.PullRequestURL{Repo: "https://gitlab.com/acme/platform/api", Provider: models.ConnectionTypeGitlab, Number: 5}, pr)

	_, err = forge.ParsePullRequestURL("https://github.com/acme/api/pull/7")
	require.ErrorIs(t, err, forge.ErrNotConnectionPullRequest)
	assert.EqualError(t, err, `pull request URL "https://github.com/acme/api/pull/7" is neither an Azure DevOps .../pullrequest/N nor a GitLab .../-/merge_requests/N URL: `+
		forge.ErrNotConnectionPullRequest.Error())

	_, err = forge.ParsePullRequestURL("https://gitlab.com/acme/api/-/merge_requests/new")
	require.EqualError(t, err, `pull request URL "https://gitlab.com/acme/api/-/merge_requests/new" has no positive number`)
}
