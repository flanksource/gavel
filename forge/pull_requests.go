// Lists the open pull requests of an organization's repositories that come from git connections
// whose pull requests the dashboard lists, each through the connection that brings it.
package forge

import (
	"context"
	"errors"
	"fmt"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/connections"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// pullRequestSources are the repository sources whose pull requests are listed through their
// connection. A GitHub repository's come from GitHub search instead.
var pullRequestSources = []string{models.ConnectionTypeAzureDevops}

// RepoPullRequest is one open pull request of a repository.
type RepoPullRequest struct {
	Repo        git.RepoURL
	PullRequest git.PullRequest
}

// PullRequests lists the open pull requests of scope's repositories that come from a connection
// listing them, through that connection as stored, never hydrated; options configures the provider
// clients, which a hosted server makes Defensive. A repository whose connection was deleted has
// none. Each repository that can't be listed is reported, joined, after the rest are.
func PullRequests(ctx context.Context, db *gorm.DB, scope tenant.Scope, options git.ClientOptions) ([]RepoPullRequest, error) {
	var repos []struct {
		Name         string
		RemoteURL    string
		ConnectionID uuid.UUID
	}
	if err := db.WithContext(ctx).Raw(`SELECT name, remote_url, connection_id FROM repos
		WHERE org_id = ? AND connection_id IS NOT NULL AND source IN ? ORDER BY name`, scope.OrgID, pullRequestSources).
		Scan(&repos).Error; err != nil {
		return nil, fmt.Errorf("list the organization's repositories: %w", err)
	}
	if len(repos) == 0 {
		return nil, nil
	}
	store, closeStore, err := connections.OpenStore(ctx, connections.Options{DB: db, Scope: scope})
	if err != nil {
		return nil, fmt.Errorf("open the organization's connection catalog: %w", err)
	}
	clients := make(map[uuid.UUID]namedClient)
	var listed []RepoPullRequest
	var errs []error
	for _, repo := range repos {
		client, ok := clients[repo.ConnectionID]
		if !ok {
			if client, err = connectionClient(ctx, store, repo.ConnectionID, options); err != nil {
				errs = append(errs, fmt.Errorf("read the connection of %s: %w", repo.Name, err))
				continue
			}
			clients[repo.ConnectionID] = client
		}
		pulls, err := openPullRequests(ctx, client, repo.Name, repo.RemoteURL)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		listed = append(listed, pulls...)
	}
	return listed, errors.Join(append(errs, closeStore())...)
}

// RepoPullRequests lists the open pull requests of the repository at remoteURL through
// connection, which a caller resolved; options configures its provider client.
func RepoPullRequests(ctx context.Context, connection models.Connection, options git.ClientOptions, remoteURL string) ([]RepoPullRequest, error) {
	client, err := clientOf(ctx, connection, options)
	if err != nil {
		return nil, err
	}
	return openPullRequests(ctx, client, remoteURL, remoteURL)
}

// openPullRequests lists the open pull requests of the repository at remoteURL, which name
// describes, through client.
func openPullRequests(ctx context.Context, client namedClient, name, remoteURL string) ([]RepoPullRequest, error) {
	remote, err := client.Repository(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("repository %s of connection %q: %w", name, client.Name, err)
	}
	pulls, err := client.Provider.ListPullRequests(ctx, remote.URL, git.ListPullRequestsOptions{State: git.PullRequestOpen})
	if err != nil {
		return nil, fmt.Errorf("list the pull requests of %s: %w", name, err)
	}
	listed := make([]RepoPullRequest, 0, len(pulls))
	for _, pull := range pulls {
		listed = append(listed, RepoPullRequest{Repo: remote.URL, PullRequest: pull})
	}
	return listed, nil
}
