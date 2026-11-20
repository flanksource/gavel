// The connection, and the repository through it, a local gavel reaches an Azure DevOps or GitLab
// repository with: each provider's local connection, which its own package resolves.
package provider

import (
	"context"
	"fmt"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/azuredevops"
	"github.com/flanksource/gavel/gitlab"
)

// localConnections resolve each provider's local connection with its client options.
var localConnections = map[string]func(context.Context, git.RepoURL) (models.Connection, git.ClientOptions, error){
	models.ConnectionTypeAzureDevops: azuredevops.LocalConnection,
	models.ConnectionTypeGitlab: func(ctx context.Context, repo git.RepoURL) (models.Connection, git.ClientOptions, error) {
		connection, err := gitlab.LocalConnection(ctx, repo)
		return connection, git.ClientOptions{}, err
	},
}

// LocalConnection is the connection, with its client options, a local gavel reaches repo of
// provider type kind through.
func LocalConnection(ctx context.Context, kind string, repo git.RepoURL) (models.Connection, git.ClientOptions, error) {
	local, ok := localConnections[kind]
	if !ok {
		return models.Connection{}, git.ClientOptions{}, fmt.Errorf("a local gavel reaches no %s repository like %s through a git connection", kind, repo.Redacted())
	}
	return local(ctx, repo)
}

// LocalRemote is the repository at repoURL, of provider type kind, as a local gavel reaches it
// through LocalConnection.
func LocalRemote(ctx context.Context, kind, repoURL string) (*git.Repository, error) {
	repo, err := git.ParseRepoURL(repoURL, kind)
	if err != nil {
		return nil, fmt.Errorf("%s repository %s: %w", kind, repoURL, err)
	}
	connection, options, err := LocalConnection(ctx, kind, repo)
	if err != nil {
		return nil, err
	}
	client, err := git.FromConnection(dbcontext.NewContext(ctx), connection, options)
	if err != nil {
		return nil, fmt.Errorf("connection %q: %w", connection.Name, err)
	}
	remote, err := client.Repository(repoURL)
	if err != nil {
		return nil, fmt.Errorf("connection %q: %w", connection.Name, err)
	}
	return remote, nil
}
