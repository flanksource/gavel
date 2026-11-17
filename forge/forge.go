// Package forge finds the git connection that holds a repository's credentials among the
// connections a gavel process can read: GitHub, GitLab, Azure DevOps and plain git ones.
package forge

import (
	"fmt"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
)

// Source is a set of connections, such as the file layers or an organization's catalog.
type Source interface {
	ListConnections() ([]*models.Connection, error)
}

// Find is the connection of sources that owns repo, hydrated through ctx, or nil when none does.
// The most specific owner wins: a git connection naming the repository, then a provider
// connection limited to its namespace, then one for its whole host; of equals, the earlier
// source's. A repo with no Name asks for its owner's connection, and one with no Owner either
// for the host's.
func Find(ctx dbcontext.Context, repo git.RepoURL, sources ...Source) (*models.Connection, error) {
	var best *models.Connection
	bestRank := 0
	for _, source := range sources {
		listed, err := source.ListConnections()
		if err != nil {
			return nil, fmt.Errorf("list connections: %w", err)
		}
		for _, connection := range listed {
			if rank := ownership(*connection, repo); rank > bestRank {
				best, bestRank = connection, rank
			}
		}
	}
	if best == nil {
		return nil, nil
	}
	hydrated, err := ctx.HydrateConnection(best)
	if err != nil {
		return nil, fmt.Errorf("read connection %q: %w", best.Name, err)
	}
	return hydrated, nil
}

// ownership ranks how specifically connection owns repo: 0 when it doesn't, 1 for every
// repository on its host, 2 for its namespace's, 3 for the one repository it names.
func ownership(connection models.Connection, repo git.RepoURL) int {
	scope, err := git.ScopeOf(connection)
	if err != nil {
		// Not a git connection, or one whose URL can't be read before hydration: it owns nothing
		// gavel can match.
		return 0
	}
	if scope.Owns(repo) != nil {
		return 0
	}
	switch {
	case scope.Repository != nil:
		if repo.Name == "" || !scope.Repository.SameRepository(repo) {
			return 0
		}
		return 3
	case scope.Namespace != "":
		return 2
	}
	return 1
}
