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
