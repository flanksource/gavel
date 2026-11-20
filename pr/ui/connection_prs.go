// Lists the pull requests of repositories git connections bring as the dashboard's PR list items,
// carrying their provider and the GitHub spelling of state and mergeability the UI reads.
package ui

import (
	"strings"

	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/forge"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/todos/types"
)

// prProviders names each connection type's PR provider as todo links and the PR index key it.
var prProviders = map[string]string{models.ConnectionTypeAzureDevops: types.PRProviderAzure, models.ConnectionTypeGitlab: types.PRProviderGitLab}

// ConnectionPRItems are listed's pull requests as PR list items. A provider keeping no update
// time, as Azure DevOps doesn't, has each item as recent as its creation.
func ConnectionPRItems(listed []forge.RepoPullRequest) github.PRSearchResults {
	items := make(github.PRSearchResults, 0, len(listed))
	for _, entry := range listed {
		pr, repo := entry.PullRequest, entry.Repo
		item := github.PRListItem{
			Provider: prProviders[repo.Provider], Number: pr.Number, Title: pr.Title, Author: pr.Author,
			Repo: strings.Join(nonEmpty(repo.Owner, repo.Project, repo.Name), "/"), Source: pr.Head, Target: pr.Base,
			State: provider.PRState(pr.State), Mergeable: "UNKNOWN", URL: pr.URL, UpdatedAt: pr.CreatedAt, ShowAuthor: true,
		}
		if metadata := pr.Metadata; metadata != nil {
			item.AuthorAvatarURL, item.IsDraft = metadata.AuthorAvatarURL, metadata.Draft
			item.Mergeable = provider.PRMergeable(metadata.Mergeability)
		}
		items = append(items, item)
	}
	return items
}

func nonEmpty(parts ...string) []string {
	kept := parts[:0]
	for _, part := range parts {
		if part != "" {
			kept = append(kept, part)
		}
	}
	return kept
}
