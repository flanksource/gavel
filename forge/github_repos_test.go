// Tests how an organization reaches its GitHub repositories: through the App installation that
// grants one, else the token of the github.com connection that brings it.
package forge_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/forge"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func githubRepo(t *testing.T, db *gorm.DB, scope tenant.Scope, name string, id int64, connection *uuid.UUID, installation *int64) {
	t.Helper()
	require.NoError(t, db.Exec(`INSERT INTO repos (org_id, name, source, github_repo_id, connection_id, installation_id, provider_repo_id, remote_url)
		VALUES (?, ?, 'github', ?, ?, ?, ?, ?)`, scope.OrgID, name, id, connection, installation, strconv.FormatInt(id, 10), "https://github.com/"+name+".git").Error)
}

func TestAnOrgReachesAGitHubRepositoryOnlyAConnectionBringsWithItsToken(t *testing.T) {
	db := catalog(t)
	scope, other := hostedOrg(t, db), hostedOrg(t, db)
	require.NoError(t, db.Exec(`INSERT INTO github_installations (id, org_id, account_login, account_type) VALUES (42, ?, 'acme', 'Organization')`, scope.OrgID).Error)
	acme := saveFor(t, db, scope, models.Connection{Name: "acme-gh", Type: models.ConnectionTypeGithub, Password: "ghp_acme",
		Properties: map[string]string{"organization": "acme"}})
	enterprise := saveFor(t, db, scope, models.Connection{Name: "corp-ghe", Type: models.ConnectionTypeGithub, URL: "https://ghe.example.com/api/v3",
		Password: "ghp_corp", Properties: map[string]string{"organization": "corp"}})
	installation := int64(42)
	githubRepo(t, db, scope, "acme/api", 101, &acme, nil)
	githubRepo(t, db, scope, "acme/cli", 103, &acme, nil)
	githubRepo(t, db, scope, "acme/web", 102, &acme, &installation)
	githubRepo(t, db, scope, "corp/tool", 201, &enterprise, nil)

	groups, err := forge.GitHubConnectionRepos(context.Background(), db, scope)
	require.NoError(t, err)
	assert.Equal(t, []forge.ConnectionRepos{{Connection: "acme-gh", Token: "ghp_acme", Repos: []string{"acme/api", "acme/cli"}}}, groups,
		"the App reaches web; a GitHub Enterprise connection's repositories aren't on github.com")

	access, err := forge.GitHubRepo(context.Background(), db, scope, "Acme/API")
	require.NoError(t, err)
	assert.Equal(t, forge.GitHubRepoAccess{ConnectionToken: "ghp_acme"}, access, "named case-insensitively")
	access, err = forge.GitHubRepo(context.Background(), db, scope, "acme/web")
	require.NoError(t, err)
	assert.Equal(t, forge.GitHubRepoAccess{InstallationID: 42}, access, "the App's installation wins")
	access, err = forge.GitHubRepo(context.Background(), db, scope, "acme/unknown")
	require.NoError(t, err)
	assert.Equal(t, forge.GitHubRepoAccess{}, access)

	groups, err = forge.GitHubConnectionRepos(context.Background(), db, other)
	require.NoError(t, err)
	assert.Empty(t, groups)
	access, err = forge.GitHubRepo(context.Background(), db, other, "acme/api")
	require.NoError(t, err)
	assert.Equal(t, forge.GitHubRepoAccess{}, access, "another organization's repository isn't reached")
}
