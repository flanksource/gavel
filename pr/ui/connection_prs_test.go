// Tests the dashboard's list items for pull requests of repositories git connections bring: they
// carry their provider and the GitHub spelling of state and mergeability the UI reads.
package ui

import (
	"testing"
	"time"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/forge"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
)

func TestAnAzurePullRequestIsListedAsADashboardItem(t *testing.T) {
	opened := time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)
	rules := git.RepoURL{Provider: models.ConnectionTypeAzureDevops, Owner: "acme", Project: "Platform", Name: "rules"}
	items := ConnectionPRItems([]forge.RepoPullRequest{
		{Repo: rules, PullRequest: git.PullRequest{Number: 9, URL: "https://dev.azure.com/acme/Platform/_git/rules/pullrequest/9",
			State: git.PullRequestOpen, Head: "tighten", Base: "main", Title: "Tighten the rules", Author: "Ada Lovelace", CreatedAt: opened,
			Metadata: &git.PullRequestMetadata{AuthorAvatarURL: "https://avatars.example/ada", Mergeability: "conflicts", Draft: true}}},
		{Repo: rules, PullRequest: git.PullRequest{Number: 10, URL: "https://dev.azure.com/acme/Platform/_git/rules/pullrequest/10",
			State: git.PullRequestMerged, Head: "loosen", Base: "main", Title: "Loosen", Author: "Grace Hopper", CreatedAt: opened}},
	})

	assert.Equal(t, github.PRSearchResults{
		{Provider: types.PRProviderAzure, Number: 9, Title: "Tighten the rules", Author: "Ada Lovelace", AuthorAvatarURL: "https://avatars.example/ada",
			Repo: "acme/Platform/rules", Source: "tighten", Target: "main", State: "OPEN", IsDraft: true, Mergeable: "CONFLICTING",
			URL: "https://dev.azure.com/acme/Platform/_git/rules/pullrequest/9", UpdatedAt: opened, ShowAuthor: true},
		{Provider: types.PRProviderAzure, Number: 10, Title: "Loosen", Author: "Grace Hopper", Repo: "acme/Platform/rules", Source: "loosen",
			Target: "main", State: "MERGED", Mergeable: "UNKNOWN", URL: "https://dev.azure.com/acme/Platform/_git/rules/pullrequest/10",
			UpdatedAt: opened, ShowAuthor: true},
	}, items, "Azure keeps no update time, so an item is as recent as its creation")

	row := prIndexRow(items[0])
	assert.Equal(t, types.PRProviderAzure, row.Provider)
	assert.Equal(t, "acme/platform/rules", row.Key().Repo, "indexed under the key a todo's link to the PR has")
	assert.Equal(t, types.PRProviderGitHub, prIndexRow(github.PRListItem{Repo: "acme/api", Number: 7}).Provider, "a GitHub search result")
}
