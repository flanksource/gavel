// The git connection lookup a local gavel sets to read its connection files and catalog, which
// comes before AZURE_DEVOPS_EXT_PAT and the az CLI for an Azure DevOps repository.
package azuredevops

import (
	"context"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
)

// FindConnection, when set, is the git connection that owns an Azure DevOps repository, or nil
// when none does. A local gavel sets it to read its connection files and catalog.
var FindConnection func(context.Context, git.RepoURL) (*models.Connection, error)
