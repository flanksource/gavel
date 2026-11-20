// The git connection that owns a repository a local gavel reaches, found by a lookup a local gavel
// sets to read its connection files and catalog; one that owns it must hold a token.
package forge

import (
	"context"
	"fmt"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
)

// OwningConnection is the git connection find says owns target, which described names for
// errors, or nil when none does or find is nil. One owning it must hold a personal access token.
func OwningConnection(ctx context.Context, find func(context.Context, git.RepoURL) (*models.Connection, error), target git.RepoURL, described string) (*models.Connection, error) {
	if find == nil {
		return nil, nil
	}
	connection, err := find(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("find the git connection for %s: %w", described, err)
	}
	if connection != nil && connection.Password == "" {
		return nil, fmt.Errorf("git connection %q owns %s but holds no personal access token", connection.Name, described)
	}
	return connection, nil
}
