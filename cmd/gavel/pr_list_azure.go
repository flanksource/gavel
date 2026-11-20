// Lists the open pull requests of the Azure DevOps repositories a local dashboard's search config
// names by URL, each through the connection a local gavel reaches it with.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/azuredevops"
	"github.com/flanksource/gavel/forge"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/ui"
)

// splitAzureRepos separates a search config's repos into the GitHub ones and the Azure DevOps
// repository URLs, on Services or a Server, whose path holds /_git/.
func splitAzureRepos(repos []string) (githubRepos, azureRepos []string) {
	for _, repo := range repos {
		if parsed, err := url.Parse(repo); err == nil && parsed.Scheme != "" && strings.Contains(parsed.Path, "/_git/") {
			azureRepos = append(azureRepos, repo)
			continue
		}
		githubRepos = append(githubRepos, repo)
	}
	return githubRepos, azureRepos
}

// searchConfiguredPRs searches GitHub with so, then lists the pull requests of azureRepos. A
// search scoped to no GitHub repositories while it names Azure ones searches no GitHub.
func searchConfiguredPRs(ctx context.Context, ghOpts github.Options, so github.PRSearchOptions, azureRepos []string) (github.PRSearchResults, *github.RateLimit, error) {
	var results github.PRSearchResults
	var rateLimit *github.RateLimit
	if len(so.Repos) > 0 || len(azureRepos) == 0 {
		var err error
		if results, rateLimit, err = github.SearchPRs(ghOpts, so); err != nil {
			return results, rateLimit, err
		}
	}
	listCtx, cancel := context.WithTimeout(ctx, connectionPRsTimeout)
	defer cancel()
	azure, err := localAzurePRs(listCtx, azureRepos)
	return append(results, azure...), rateLimit, err
}

// localAzurePRs lists the open pull requests of repos, Azure DevOps repository URLs, each through
// azuredevops.LocalConnection. A repository that can't be listed is reported, joined, after the
// rest are.
func localAzurePRs(ctx context.Context, repos []string) (github.PRSearchResults, error) {
	var listed []forge.RepoPullRequest
	var errs []error
	for _, raw := range repos {
		repo, err := git.ParseRepoURL(raw, models.ConnectionTypeAzureDevops)
		if err != nil {
			errs = append(errs, fmt.Errorf("Azure DevOps repository %s: %w", raw, err))
			continue
		}
		connection, options, err := azuredevops.LocalConnection(ctx, repo)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		pulls, err := forge.RepoPullRequests(ctx, connection, options, raw)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		listed = append(listed, pulls...)
	}
	return ui.ConnectionPRItems(listed), errors.Join(errs...)
}

// localPRRemote reaches a listed Azure DevOps pull request's repository through
// azuredevops.LocalConnection.
func localPRRemote(ctx context.Context, pr github.PRListItem) (*git.Repository, error) {
	repoURL, err := forge.PullRequestRepository(pr.URL)
	if err != nil {
		return nil, err
	}
	return azuredevops.LocalRemote(ctx, repoURL)
}
