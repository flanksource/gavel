// Lists the open pull requests of the Azure DevOps and GitLab repositories a local dashboard's search
// config names by URL, each through the connection a local gavel reaches it with.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/forge"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/gitlab"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/pr/ui"
)

// connectionRepo is a repository a search config names by URL whose pull requests are listed
// through a git connection: an Azure DevOps or GitLab one, Provider its connection type.
type connectionRepo struct {
	URL      string
	Provider string
}

// splitConnectionRepos separates a search config's repos into the GitHub ones and the Azure DevOps
// and GitLab repository URLs: an Azure DevOps one's path holds /_git/, on Services or a Server; a
// GitLab one is on gitlab.com, or on a host whose gitlab connection owns it.
func splitConnectionRepos(ctx context.Context, repos []string) (githubRepos []string, connectionRepos []connectionRepo, err error) {
	for _, repo := range repos {
		provider, err := connectionRepoProvider(ctx, repo)
		if err != nil {
			return nil, nil, err
		}
		if provider == "" {
			githubRepos = append(githubRepos, repo)
			continue
		}
		connectionRepos = append(connectionRepos, connectionRepo{URL: repo, Provider: provider})
	}
	return githubRepos, connectionRepos, nil
}

// connectionRepoProvider is the connection type of the repository at repo when a git connection
// lists its pull requests, or "" for a GitHub one.
func connectionRepoProvider(ctx context.Context, repo string) (string, error) {
	parsed, err := url.Parse(repo)
	if err != nil || parsed.Scheme == "" {
		return "", nil
	}
	if strings.Contains(parsed.Path, "/_git/") {
		return models.ConnectionTypeAzureDevops, nil
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "github.com" {
		return "", nil
	}
	if host == "gitlab.com" {
		return models.ConnectionTypeGitlab, nil
	}
	target, err := git.ParseRepoURL(repo, models.ConnectionTypeGitlab)
	if err != nil {
		return "", nil
	}
	owner, err := forge.OwningConnection(ctx, gitlab.FindConnection, target, "repository "+repo)
	if err != nil {
		return "", err
	}
	if owner != nil && owner.Type == models.ConnectionTypeGitlab {
		return models.ConnectionTypeGitlab, nil
	}
	return "", nil
}

// searchConfiguredPRs searches GitHub with so, then lists the pull requests of connectionRepos. A
// search scoped to no GitHub repositories while it names connection ones searches no GitHub.
func searchConfiguredPRs(ctx context.Context, ghOpts github.Options, so github.PRSearchOptions, connectionRepos []connectionRepo) (github.PRSearchResults, *github.RateLimit, error) {
	var results github.PRSearchResults
	var rateLimit *github.RateLimit
	if len(so.Repos) > 0 || len(connectionRepos) == 0 {
		var err error
		if results, rateLimit, err = github.SearchPRs(ghOpts, so); err != nil {
			return results, rateLimit, err
		}
	}
	listCtx, cancel := context.WithTimeout(ctx, connectionPRsTimeout)
	defer cancel()
	listed, err := localConnectionPRs(listCtx, connectionRepos)
	return append(results, listed...), rateLimit, err
}

// localConnectionPRs lists the open pull requests of repos, each through its provider's local
// connection. A repository that can't be listed is reported, joined, after the rest are.
func localConnectionPRs(ctx context.Context, repos []connectionRepo) (github.PRSearchResults, error) {
	var listed []forge.RepoPullRequest
	var errs []error
	for _, repo := range repos {
		parsed, err := git.ParseRepoURL(repo.URL, repo.Provider)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s repository %s: %w", repo.Provider, repo.URL, err))
			continue
		}
		connection, options, err := provider.LocalConnection(ctx, repo.Provider, parsed)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		pulls, err := forge.RepoPullRequests(ctx, connection, options, repo.URL)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		listed = append(listed, pulls...)
	}
	return ui.ConnectionPRItems(listed), errors.Join(errs...)
}

// localPRRemote reaches a listed Azure DevOps or GitLab pull request's repository through its
// provider's local connection.
func localPRRemote(ctx context.Context, pr github.PRListItem) (*git.Repository, error) {
	parsed, err := forge.ParsePullRequestURL(pr.URL)
	if err != nil {
		return nil, err
	}
	return provider.LocalRemote(ctx, parsed.Provider, parsed.Repo)
}
