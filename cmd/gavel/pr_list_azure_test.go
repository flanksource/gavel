// Tests a local dashboard's Azure DevOps pull requests: the repositories its search config names
// by URL, each listed through the connection a local gavel reaches it with.
package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/azuredevops"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestASearchConfigNamesAzureRepositoriesByURL(t *testing.T) {
	githubRepos, azureRepos := splitAzureRepos([]string{"acme/api", "https://dev.azure.com/acme/Platform/_git/rules",
		"https://github.com/acme/web", "https://tfs.example.com/DefaultCollection/Platform/_git/site"})
	assert.Equal(t, []string{"acme/api", "https://github.com/acme/web"}, githubRepos)
	assert.Equal(t, []string{"https://dev.azure.com/acme/Platform/_git/rules", "https://tfs.example.com/DefaultCollection/Platform/_git/site"}, azureRepos)
}

func TestALocalDashboardListsTheAzurePullRequestsOfItsConfiguredRepositories(t *testing.T) {
	azure := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/acme/Platform/_apis/git/repositories/rules/pullrequests" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"value":[{"pullRequestId":9,"status":"active","title":"Tighten the rules","creationDate":"2026-11-02T10:00:00Z",
			"sourceRefName":"refs/heads/tighten","targetRefName":"refs/heads/main","createdBy":{"displayName":"Ada Lovelace"},
			"repository":{"name":"rules","project":{"name":"Platform"}}}]}`))
	}))
	t.Cleanup(azure.Close)
	previous := azuredevops.FindConnection
	azuredevops.FindConnection = func(context.Context, git.RepoURL) (*models.Connection, error) {
		return &models.Connection{Name: "acme", Type: models.ConnectionTypeAzureDevops, URL: azure.URL + "/acme", Password: "azure-pat", InsecureTLS: true}, nil
	}
	t.Cleanup(func() { azuredevops.FindConnection = previous })

	items, err := localAzurePRs(context.Background(), []string{azure.URL + "/acme/Platform/_git/rules"})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, types.PRProviderAzure, items[0].Provider)
	assert.Equal(t, "acme/Platform/rules", items[0].Repo)
	assert.Equal(t, 9, items[0].Number)
	assert.Equal(t, "Tighten the rules", items[0].Title)

	_, err = localAzurePRs(context.Background(), []string{azure.URL + "/acme/Platform/_git/gone"})
	require.ErrorContains(t, err, "list the pull requests of "+azure.URL+"/acme/Platform/_git/gone")

	searched := false
	ghOpts := github.Options{TokenSource: func() (string, error) {
		searched = true
		return "", errors.New("GitHub was searched")
	}}
	items, _, err = searchConfiguredPRs(context.Background(), ghOpts, github.PRSearchOptions{}, []string{azure.URL + "/acme/Platform/_git/rules"})
	require.NoError(t, err)
	assert.False(t, searched, "a config naming only Azure DevOps repositories searches no GitHub")
	require.Len(t, items, 1)
	assert.Equal(t, types.PRProviderAzure, items[0].Provider)
}

func TestALocalDashboardReachesAListedAzurePullRequestThroughItsConnection(t *testing.T) {
	previous := azuredevops.FindConnection
	azuredevops.FindConnection = func(context.Context, git.RepoURL) (*models.Connection, error) {
		return &models.Connection{Name: "acme", Type: models.ConnectionTypeAzureDevops, URL: "https://tfs.example.com/acme", Password: "azure-pat"}, nil
	}
	t.Cleanup(func() { azuredevops.FindConnection = previous })

	remote, err := localPRRemote(context.Background(), github.PRListItem{Provider: types.PRProviderAzure, Repo: "acme/Platform/rules", Number: 9,
		URL: "https://tfs.example.com/acme/Platform/_git/rules/pullrequest/9"})
	require.NoError(t, err)
	assert.Equal(t, "tfs.example.com", remote.URL.Host)
	assert.Equal(t, "rules", remote.URL.Name)
	require.NotNil(t, remote.Provider)
}
