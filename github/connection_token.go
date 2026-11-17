// The GitHub token a git connection holds: a local gavel reads it from the connection that owns
// the repository a call is for, ahead of the environment's token.
package github

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
)

// FindConnection, when set, is the git connection that owns a GitHub repository (owner and
// name), an owner's repositories (no name) or every repository on the host (neither), or nil
// when none does. A local gavel sets it to read its connection files and catalog.
var FindConnection func(context.Context, git.RepoURL) (*models.Connection, error)

// connectionToken is the token of the git connection that owns o's repository, or "" when no
// connection does. An organization's search asks for that organization's connection.
func (o Options) connectionToken() (string, error) {
	if FindConnection == nil {
		return "", nil
	}
	target := git.RepoURL{Scheme: "https", Host: githubWebHost(), Provider: models.ConnectionTypeGithub, Owner: o.owner}
	described := "GitHub host " + target.Host
	if o.owner != "" {
		described = "GitHub organization " + o.owner
	} else if repo := o.connectionRepo(); repo != "" {
		target.Owner, target.Name, _ = strings.Cut(repo, "/")
		target.Path = repo
		described = "GitHub repository " + repo
	}
	connection, err := FindConnection(context.Background(), target)
	if err != nil {
		return "", fmt.Errorf("find the git connection for %s: %w", described, err)
	}
	if connection == nil {
		return "", nil
	}
	if connection.Password == "" {
		return "", fmt.Errorf("%w: git connection %q owns %s but holds no token", ErrNoToken, connection.Name, described)
	}
	return connection.Password, nil
}

// connectionRepo is the owner/name o acts on: its Repo, else its checkout's GitHub origin. A
// checkout with no GitHub origin is no repository, so the host's connection applies.
func (o Options) connectionRepo() string {
	if o.Repo != "" || o.WorkDir == "" {
		return o.Repo
	}
	repo, err := o.resolveRepo()
	if err != nil {
		return ""
	}
	return repo
}

// githubWebHost is the host GitHub repositories live on: github.com for the public API, the
// API's own host for GitHub Enterprise.
func githubWebHost() string {
	api, err := url.Parse(githubAPIBase())
	if err != nil || api.Host == "" || strings.EqualFold(api.Host, "api.github.com") {
		return "github.com"
	}
	return api.Host
}
