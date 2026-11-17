// The Azure DevOps PAT a git connection holds: a local gavel reads it from the connection that
// owns the repository, ahead of AZURE_DEVOPS_EXT_PAT and the az CLI.
package azuredevops

import (
	"context"
	"fmt"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
)

// FindConnection, when set, is the git connection that owns an Azure DevOps repository, or nil
// when none does. A local gavel sets it to read its connection files and catalog.
var FindConnection func(context.Context, git.RepoURL) (*models.Connection, error)

// connectionPAT is the PAT of the git connection that owns repo, or "" when no connection does.
func connectionPAT(ctx context.Context, repo Repository) (string, error) {
	if FindConnection == nil {
		return "", nil
	}
	described := "Azure DevOps repository " + repo.URL()
	target, err := git.ParseRepoURL(repo.URL(), models.ConnectionTypeAzureDevops)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", described, err)
	}
	connection, err := FindConnection(ctx, target)
	if err != nil {
		return "", fmt.Errorf("find the git connection for %s: %w", described, err)
	}
	if connection == nil {
		return "", nil
	}
	if connection.Password == "" {
		return "", fmt.Errorf("git connection %q owns %s but holds no personal access token", connection.Name, described)
	}
	return connection.Password, nil
}
