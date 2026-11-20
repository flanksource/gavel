// The git connection that owns an Azure DevOps repository: a local gavel reads it from its connection
// files and catalog, ahead of AZURE_DEVOPS_EXT_PAT and the az CLI.
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

// owningConnection is the git connection that owns target, which described names for errors, or
// nil when none does or no lookup is set; one owning it must hold a PAT.
func owningConnection(ctx context.Context, target git.RepoURL, described string) (*models.Connection, error) {
	if FindConnection == nil {
		return nil, nil
	}
	connection, err := FindConnection(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("find the git connection for %s: %w", described, err)
	}
	if connection != nil && connection.Password == "" {
		return nil, fmt.Errorf("git connection %q owns %s but holds no personal access token", connection.Name, described)
	}
	return connection, nil
}
