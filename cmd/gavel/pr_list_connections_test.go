// Tests a local dashboard's Azure DevOps pull requests and GitLab merge requests: the repositories
// its search config names by URL, each listed through the connection a local gavel reaches it with.
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
	"github.com/flanksource/gavel/gitlab"
	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestASearchConfigNamesAzureAndGitLabRepositoriesByURL(t *testing.T) {
	previous := gitlab.FindConnection
	gitlab.FindConnection = func(_ context.Context, repo git.RepoURL) (*models.Connection, error) {
		if repo.Host == "gitlab.example.com" {
			return &models.Connection{Name: "lab", Type: models.ConnectionTypeGitlab, URL: "https://gitlab.example.com", Password: "glpat"}, nil
		}
		return nil, nil
	}
	t.Cleanup(func() { gitlab.FindConnection = previous })

	githubRepos, connectionRepos, err := splitConnectionRepos(context.Background(), []string{"acme/api", "https://dev.azure.com/acme/Platform/_git/rules",
		"https://github.com/acme/web", "https://tfs.example.com/DefaultCollection/Platform/_git/site", "https://gitlab.com/acme/platform/api",
		"https://gitlab.example.com/team/svc", "https://git.example.com/team/tool"})
	require.NoError(t, err)
	assert.Equal(t, []string{"acme/api", "https://github.com/acme/web", "https://git.example.com/team/tool"}, githubRepos,
		"a host no gitlab connection owns isn't GitLab")
	assert.Equal(t, []connectionRepo{
		{URL: "https://dev.azure.com/acme/Platform/_git/rules", Provider: models.ConnectionTypeAzureDevops},
		{URL: "https://tfs.example.com/DefaultCollection/Platform/_git/site", Provider: models.ConnectionTypeAzureDevops},
		{URL: "https://gitlab.com/acme/platform/api", Provider: models.ConnectionTypeGitlab},
		{URL: "https://gitlab.example.com/team/svc", Provider: models.ConnectionTypeGitlab},
	}, connectionRepos)

	gitlab.FindConnection = func(context.Context, git.RepoURL) (*models.Connection, error) {
		return nil, errors.New("the catalog's key is unavailable")
	}
	_, _, err = splitConnectionRepos(context.Background(), []string{"https://gitlab.example.com/team/svc"})
	require.EqualError(t, err, "find the git connection for repository https://gitlab.example.com/team/svc: the catalog's key is unavailable")
}

func TestARepoArgumentMayNameAnAzureRepositoryByURL(t *testing.T) {
	repo, err := resolveRepoArg("https://dev.azure.com/acme/Platform/_git/rules")
	require.NoError(t, err)
	assert.Equal(t, "https://dev.azure.com/acme/Platform/_git/rules", repo, "kept as the URL the search config names it by")
	repo, err = resolveRepoArg("https://gitlab.com/acme/platform/api")
	require.NoError(t, err)
	assert.Equal(t, "https://gitlab.com/acme/platform/api", repo, "a GitLab project too")
	repo, err = resolveRepoArg("https://github.com/acme/api")
	require.NoError(t, err)
	assert.Equal(t, "acme/api", repo)
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

	items, err := localConnectionPRs(context.Background(), []connectionRepo{{URL: azure.URL + "/acme/Platform/_git/rules", Provider: models.ConnectionTypeAzureDevops}})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, types.PRProviderAzure, items[0].Provider)
	assert.Equal(t, "acme/Platform/rules", items[0].Repo)
	assert.Equal(t, 9, items[0].Number)
	assert.Equal(t, "Tighten the rules", items[0].Title)

	_, err = localConnectionPRs(context.Background(), []connectionRepo{{URL: azure.URL + "/acme/Platform/_git/gone", Provider: models.ConnectionTypeAzureDevops}})
	require.ErrorContains(t, err, "list the pull requests of "+azure.URL+"/acme/Platform/_git/gone")

	searched := false
	ghOpts := github.Options{TokenSource: func() (string, error) {
		searched = true
		return "", errors.New("GitHub was searched")
	}}
	items, _, err = searchConfiguredPRs(context.Background(), ghOpts, github.PRSearchOptions{},
		[]connectionRepo{{URL: azure.URL + "/acme/Platform/_git/rules", Provider: models.ConnectionTypeAzureDevops}})
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

func TestALocalDashboardReachesAListedGitLabMergeRequestThroughItsConnection(t *testing.T) {
	previous := gitlab.FindConnection
	gitlab.FindConnection = func(context.Context, git.RepoURL) (*models.Connection, error) {
		return &models.Connection{Name: "lab", Type: models.ConnectionTypeGitlab, URL: "https://gitlab.example.com", Password: "glpat"}, nil
	}
	t.Cleanup(func() { gitlab.FindConnection = previous })

	remote, err := localPRRemote(context.Background(), github.PRListItem{Provider: types.PRProviderGitLab, Repo: "team/svc", Number: 5,
		URL: "https://gitlab.example.com/team/svc/-/merge_requests/5"})
	require.NoError(t, err)
	assert.Equal(t, "gitlab.example.com", remote.URL.Host)
	assert.Equal(t, "svc", remote.URL.Name)
	assert.Equal(t, models.ConnectionTypeGitlab, remote.Provider.Type())
}

func TestALocalDashboardListsTheMergeRequestsOfItsConfiguredGitLabProjects(t *testing.T) {
	var lab *httptest.Server
	lab = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v4/projects/team/svc/merge_requests" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[{"iid":5,"state":"opened","title":"Speed up","web_url":"` + lab.URL + `/team/svc/-/merge_requests/5",
			"source_branch":"fast","target_branch":"main","source_project_id":1,"target_project_id":1,"author":{"name":"Grace Hopper"}}]`))
	}))
	t.Cleanup(lab.Close)
	previous := gitlab.FindConnection
	gitlab.FindConnection = func(context.Context, git.RepoURL) (*models.Connection, error) {
		return &models.Connection{Name: "lab", Type: models.ConnectionTypeGitlab, URL: lab.URL, Password: "glpat", InsecureTLS: true}, nil
	}
	t.Cleanup(func() { gitlab.FindConnection = previous })

	items, err := localConnectionPRs(context.Background(), []connectionRepo{{URL: lab.URL + "/team/svc", Provider: models.ConnectionTypeGitlab}})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, types.PRProviderGitLab, items[0].Provider)
	assert.Equal(t, "team/svc", items[0].Repo)
	assert.Equal(t, 5, items[0].Number)
}
