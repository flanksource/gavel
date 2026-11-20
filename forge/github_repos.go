// How an organization reaches its GitHub repositories: through the App installation that grants
// one, else the token of the github.com connection that brings it, as stored and never hydrated.
package forge

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/gavel/connections"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConnectionRepos are the GitHub repositories (owner/name) only one connection brings, with its
// token.
type ConnectionRepos struct {
	Connection string
	Token      string
	Repos      []string
}

// GitHubRepoAccess is how an organization reaches one of its GitHub repositories: the App
// installation that grants it, else the token of the connection that brings it; neither when the
// organization holds no such repository.
type GitHubRepoAccess struct {
	InstallationID  int64
	ConnectionToken string
}

// GitHubConnectionRepos groups scope's GitHub repositories that no App installation grants by the
// github.com connection that brings them. A GitHub Enterprise connection's repositories aren't on
// github.com, so they are left out.
func GitHubConnectionRepos(ctx context.Context, db *gorm.DB, scope tenant.Scope) ([]ConnectionRepos, error) {
	var repos []struct {
		Name         string
		ConnectionID uuid.UUID
	}
	if err := db.WithContext(ctx).Raw(`SELECT name, connection_id FROM repos
		WHERE org_id = ? AND source = 'github' AND installation_id IS NULL AND connection_id IS NOT NULL
		ORDER BY connection_id, name`, scope.OrgID).Scan(&repos).Error; err != nil {
		return nil, fmt.Errorf("list the organization's GitHub repositories: %w", err)
	}
	if len(repos) == 0 {
		return nil, nil
	}
	tokens, err := githubTokens(ctx, db, scope)
	if err != nil {
		return nil, err
	}
	var groups []ConnectionRepos
	byConnection := map[uuid.UUID]int{}
	for _, repo := range repos {
		token, ok := tokens[repo.ConnectionID]
		if !ok {
			continue
		}
		index, seen := byConnection[repo.ConnectionID]
		if !seen {
			index = len(groups)
			byConnection[repo.ConnectionID] = index
			groups = append(groups, ConnectionRepos{Connection: token.name, Token: token.token})
		}
		groups[index].Repos = append(groups[index].Repos, repo.Name)
	}
	return groups, nil
}

// GitHubRepo is how scope reaches its GitHub repository repo (owner/name, compared
// case-insensitively).
func GitHubRepo(ctx context.Context, db *gorm.DB, scope tenant.Scope, repo string) (GitHubRepoAccess, error) {
	var row struct {
		InstallationID *int64
		ConnectionID   *uuid.UUID
	}
	result := db.WithContext(ctx).Raw(`SELECT installation_id, connection_id FROM repos
		WHERE org_id = ? AND source = 'github' AND lower(name) = lower(?) LIMIT 1`, scope.OrgID, repo).Scan(&row)
	if result.Error != nil {
		return GitHubRepoAccess{}, fmt.Errorf("find the organization's GitHub repository %s: %w", repo, result.Error)
	}
	switch {
	case result.RowsAffected == 0:
		return GitHubRepoAccess{}, nil
	case row.InstallationID != nil:
		return GitHubRepoAccess{InstallationID: *row.InstallationID}, nil
	case row.ConnectionID == nil:
		return GitHubRepoAccess{}, nil
	}
	tokens, err := githubTokens(ctx, db, scope)
	if err != nil {
		return GitHubRepoAccess{}, err
	}
	return GitHubRepoAccess{ConnectionToken: tokens[*row.ConnectionID].token}, nil
}

type githubToken struct{ name, token string }

// githubTokens are the tokens of scope's github.com connections, by connection ID.
func githubTokens(ctx context.Context, db *gorm.DB, scope tenant.Scope) (map[uuid.UUID]githubToken, error) {
	var ids []uuid.UUID
	if err := db.WithContext(ctx).Raw(`SELECT id FROM connections WHERE org_id = ? AND type = 'github'`, scope.OrgID).Scan(&ids).Error; err != nil {
		return nil, fmt.Errorf("list the organization's GitHub connections: %w", err)
	}
	store, closeStore, err := connections.OpenStore(ctx, connections.Options{DB: db, Scope: scope})
	if err != nil {
		return nil, fmt.Errorf("open the organization's connection catalog: %w", err)
	}
	tokens := make(map[uuid.UUID]githubToken, len(ids))
	var errs []error
	for _, id := range ids {
		connection, err := store.ResolveConnection(dbcontext.NewContext(ctx), id.String())
		if err != nil {
			errs = append(errs, fmt.Errorf("read GitHub connection %s: %w", id, err))
			continue
		}
		if onGitHubDotCom(connection.URL) {
			tokens[id] = githubToken{name: connection.Name, token: connection.Password}
		}
	}
	return tokens, errors.Join(append(errs, closeStore())...)
}

// onGitHubDotCom reports whether a github connection's API URL, empty for the default, is
// github.com's.
func onGitHubDotCom(api string) bool {
	if api == "" {
		return true
	}
	u, err := url.Parse(api)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return host == "github.com" || host == "api.github.com"
}
