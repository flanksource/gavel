// The git connections a local gavel reads repository tokens from: the working directory's and
// the user's connection files, then the local catalog. A hosted server never reads them.
package main

import (
	"context"
	"errors"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/azuredevops"
	"github.com/flanksource/gavel/forge"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/internal/database"
	"gorm.io/gorm"
)

func init() { useLocalConnections(true) }

// useLocalConnections makes GitHub and Azure DevOps clients read a repository's token from this
// host's git connections, or, when off, never.
func useLocalConnections(on bool) {
	find := localConnection
	if !on {
		find = nil
	}
	github.FindConnection, azuredevops.FindConnection = find, find
}

// localConnection is the git connection that owns repo among the working directory's connection
// files and the process database's local catalog, or nil when none does.
func localConnection(ctx context.Context, repo git.RepoURL) (*models.Connection, error) {
	dir, err := getWorkingDir()
	if err != nil {
		return nil, err
	}
	sources, closeSources, err := forge.Local(ctx, dir, localCatalogDB(ctx))
	if err != nil {
		return nil, err
	}
	found, err := forge.Find(dbcontext.NewContext(ctx), repo, sources...)
	return found, errors.Join(err, closeSources())
}

// localCatalogDB is the process database, or nil when this run has none: its catalog is then
// unavailable, as the GitHub cache is.
func localCatalogDB(ctx context.Context) *gorm.DB {
	db, err := database.Shared(ctx)
	if err != nil || db.Disabled() {
		return nil
	}
	return db.Gorm()
}
