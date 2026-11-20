// The connection a local gavel reaches an Azure DevOps repository's API through, in the order its
// clients authenticate: the git connection owning it, AZURE_DEVOPS_EXT_PAT, then the az CLI login.
package azuredevops

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
)

// cliLogin caches the az CLI login's token for every local connection that carries it.
var cliLogin = &authenticator{fetch: cliAccessToken, now: time.Now}

// LocalConnection is the connection, with its client options, a local gavel reaches repo through:
// the git connection that owns it; else one for repo's organization carrying AZURE_DEVOPS_EXT_PAT;
// else one whose API calls carry the az CLI login's token.
func LocalConnection(ctx context.Context, repo git.RepoURL) (models.Connection, git.ClientOptions, error) {
	described := "Azure DevOps repository " + Repository{Organization: repo.Owner, Project: repo.Project, Name: repo.Name}.URL()
	owner, err := owningConnection(ctx, repo, described)
	if err != nil {
		return models.Connection{}, git.ClientOptions{}, err
	}
	if owner != nil {
		return *owner, git.ClientOptions{}, nil
	}
	organization := "https://dev.azure.com/" + repo.Owner
	if pat := strings.TrimSpace(os.Getenv("AZURE_DEVOPS_EXT_PAT")); pat != "" {
		return models.Connection{Name: "AZURE_DEVOPS_EXT_PAT", Type: models.ConnectionTypeAzureDevops, URL: organization, Password: pat},
			git.ClientOptions{}, nil
	}
	return models.Connection{Name: "az CLI", Type: models.ConnectionTypeAzureDevops, URL: organization},
		git.ClientOptions{TokenSource: cliLogin.cliToken}, nil
}
