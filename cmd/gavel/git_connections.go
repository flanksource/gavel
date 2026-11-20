// The git connections a local gavel reads repository tokens from: the working directory's and
// the user's connection files, then a local dashboard's catalog. A hosted server never reads them.
package main

import (
	"context"
	"errors"
	"sync/atomic"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/azuredevops"
	"github.com/flanksource/gavel/forge"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/gitlab"
	"gorm.io/gorm"
)

func init() { useLocalConnections(true) }

// connectionCatalog is the database whose local catalog the lookup reads after the connection
// files: the migrated one a local dashboard serves from. A command has none, so a token lookup
// never opens a database.
var connectionCatalog atomic.Pointer[gorm.DB]

// useConnectionCatalog makes the lookup read db's local catalog after the connection files; nil
// reads the files alone.
func useConnectionCatalog(db *gorm.DB) { connectionCatalog.Store(db) }

// useLocalConnections makes GitHub, Azure DevOps and GitLab clients read a repository's token
// from this host's git connections, or, when off, never.
func useLocalConnections(on bool) {
	find := localConnection
	if !on {
		find = nil
	}
	github.FindConnection, azuredevops.FindConnection, gitlab.FindConnection = find, find, find
}

// localConnection is the git connection that owns repo among the working directory's connection
// files and the process database's local catalog, or nil when none does.
func localConnection(ctx context.Context, repo git.RepoURL) (*models.Connection, error) {
	dir, err := getWorkingDir()
	if err != nil {
		return nil, err
	}
	sources, closeSources, err := forge.Local(ctx, dir, connectionCatalog.Load())
	if err != nil {
		return nil, err
	}
	found, err := forge.Find(dbcontext.NewContext(ctx), repo, sources...)
	return found, errors.Join(err, closeSources())
}
