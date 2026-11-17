// Tests that the git connection owning a repository supplies its GitHub token: after an explicit
// token, before the environment, and never for a client with a TokenSource.
package github

import (
	"context"
	"encoding/json"
	"errors"
	nethttp "net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// connectionsOwning makes FindConnection answer with the token tokens holds for a lookup's
// owner/name, owner, or "" for the whole host, and records each lookup.
func connectionsOwning(t *testing.T, tokens map[string]string) *[]string {
	t.Helper()
	var lookups []string
	previous := FindConnection
	FindConnection = func(_ context.Context, repo git.RepoURL) (*models.Connection, error) {
		key := strings.Trim(repo.Owner+"/"+repo.Name, "/")
		lookups = append(lookups, key)
		if token, ok := tokens[key]; ok {
			return &models.Connection{Name: "conn-" + key, Type: models.ConnectionTypeGithub, Password: token}, nil
		}
		return nil, nil
	}
	t.Cleanup(func() { FindConnection = previous })
	return &lookups
}

func hostCredentials(t *testing.T) {
	t.Helper()
	t.Setenv("GITHUB_TOKEN", "host-token")
	t.Setenv("GH_TOKEN", "")
	t.Setenv("HOME", t.TempDir())
}

func TestTheConnectionOwningARepositorySuppliesItsToken(t *testing.T) {
	hostCredentials(t)
	lookups := connectionsOwning(t, map[string]string{"acme/api": "gh_api", "": "gh_all"})

	token, err := Options{Repo: "acme/api"}.token()
	require.NoError(t, err)
	assert.Equal(t, "gh_api", token)

	token, err = Options{}.token()
	require.NoError(t, err)
	assert.Equal(t, "gh_all", token, "a call for no repository uses the host's connection")

	token, err = Options{Repo: "acme/api", Token: "flag-token"}.token()
	require.NoError(t, err)
	assert.Equal(t, "flag-token", token, "an explicit token comes before any connection")

	token, err = Options{Repo: "acme/api", TokenSource: func() (string, error) { return "org-token", nil }}.token()
	require.NoError(t, err)
	assert.Equal(t, "org-token", token)
	assert.Equal(t, []string{"acme/api", ""}, *lookups, "neither an explicit token nor a TokenSource looks for a connection")
}

func TestWithoutAnOwningConnectionTheEnvironmentApplies(t *testing.T) {
	hostCredentials(t)
	connectionsOwning(t, map[string]string{})
	token, err := Options{Repo: "acme/api"}.token()
	require.NoError(t, err)
	assert.Equal(t, "host-token", token)
}

func TestAConnectionLookupFailureIsTheTokensError(t *testing.T) {
	hostCredentials(t)
	previous := FindConnection
	FindConnection = func(context.Context, git.RepoURL) (*models.Connection, error) {
		return nil, errors.New("the catalog's key is unavailable")
	}
	t.Cleanup(func() { FindConnection = previous })

	_, err := Options{Repo: "acme/api"}.token()
	require.EqualError(t, err, "find the git connection for GitHub repository acme/api: the catalog's key is unavailable")
}

func TestAConnectionWithoutATokenIsRefused(t *testing.T) {
	hostCredentials(t)
	connectionsOwning(t, map[string]string{"acme/api": ""})
	_, err := Options{Repo: "acme/api"}.token()
	require.ErrorIs(t, err, ErrNoToken)
	assert.EqualError(t, err, `no GitHub token: git connection "conn-acme/api" owns GitHub repository acme/api but holds no token`)
}

func TestACheckoutsRepositoryPicksTheConnection(t *testing.T) {
	hostCredentials(t)
	connectionsOwning(t, map[string]string{"acme/web": "gh_web"})
	dir := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"remote", "add", "origin", "git@github.com:acme/web.git"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}
	token, err := Options{WorkDir: dir}.token()
	require.NoError(t, err)
	assert.Equal(t, "gh_web", token)
}

func TestAnOrganizationsSearchUsesItsConnection(t *testing.T) {
	hostCredentials(t)
	connectionsOwning(t, map[string]string{"acme": "gh_acme", "": "gh_all"})
	var used []string
	api := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		used = append(used, strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"search": map[string]any{"issueCount": 0, "nodes": []any{}}}})
	}))
	t.Cleanup(api.Close)
	t.Setenv("GITHUB_API_URL", api.URL)

	_, _, err := SearchPRs(Options{}, PRSearchOptions{All: true, Org: "acme", State: "open"})
	require.NoError(t, err)
	assert.Equal(t, []string{"gh_acme"}, used)
}
