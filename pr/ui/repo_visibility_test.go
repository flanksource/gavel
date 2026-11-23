// Tests a hosted organization member's own PR list: the repositories and owners they hide are left
// out of their PR list and its live stream, and a change to what they hide reaches the stream.
package ui

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// visibleRepoRow records one of org's repositories and returns its workspace, repo:<id>.
func visibleRepoRow(t *testing.T, db *gorm.DB, org uuid.UUID, source, name, remoteURL string) string {
	t.Helper()
	var repo struct{ ID uuid.UUID }
	remote := any(nil)
	if remoteURL != "" {
		remote = remoteURL
	}
	require.NoError(t, db.Raw(`INSERT INTO repos (org_id, name, source, remote_url) VALUES (?, ?, ?, ?) RETURNING id`, org, name, source, remote).Scan(&repo).Error)
	return RepoWorkspace(repo.ID)
}

// prKeys are the repo#number keys of the PRs in a snapshot's JSON, sorted.
func prKeys(t *testing.T, body string) []string {
	t.Helper()
	var snap struct {
		PRs []struct {
			Repo   string
			Number int
		}
	}
	require.NoError(t, json.Unmarshal([]byte(body), &snap), body)
	keys := []string{}
	for _, pr := range snap.PRs {
		keys = append(keys, fmt.Sprintf("%s#%d", pr.Repo, pr.Number))
	}
	sort.Strings(keys)
	return keys
}

func TestAHostedMembersPRListLeavesOutWhatTheyHide(t *testing.T) {
	db, store, org := openOrgDB(t)
	t.Setenv("GAVEL_SECRET_KEY", base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("r"), 32)))
	ada, bob := uuid.New(), uuid.New()
	for name, id := range map[string]uuid.UUID{"ada": ada, "bob": bob} {
		require.NoError(t, db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, "user_"+name).Error)
	}
	visibleRepoRow(t, db, org, "github", "acme/api", "")
	web := visibleRepoRow(t, db, org, "github", "acme/web", "")
	visibleRepoRow(t, db, org, "gitlab", "platform/sub/svc", "https://gitlab.example.com/platform/sub/svc.git")
	visibleRepoRow(t, db, org, "azure_devops", "Platform/rules", "https://dev.azure.com/acme/Platform/_git/rules")

	srv := &Server{githubCache: store, fetchedAt: time.Now()}
	srv.SetOrg(org, db)
	srv.SetHostedMode(true)
	srv.publishResults(github.PRSearchResults{
		{Repo: "acme/api", Number: 1, URL: "https://github.com/acme/api/pull/1"},
		{Repo: "acme/web", Number: 2, URL: "https://github.com/acme/web/pull/2"},
		{Provider: types.PRProviderGitLab, Repo: "platform/sub/svc", Number: 3, URL: "https://gitlab.example.com/platform/sub/svc/-/merge_requests/3"},
		{Provider: types.PRProviderAzure, Repo: "acme/Platform/rules", Number: 4, URL: "https://dev.azure.com/acme/Platform/_git/rules/pullrequest/4"},
		{Repo: "globex/app", Number: 5, URL: "https://github.com/globex/app/pull/5"},
	})
	base := serveActingAs(t, srv, org)
	call := func(user uuid.UUID, method, path string, body any) string {
		t.Helper()
		status, reply := connectionsPageCall(t, base, method, path, user, "org:member", body, nil)
		require.Less(t, status, 300, reply)
		return reply
	}
	call(ada, http.MethodPost, "/api/filters/hidden-projects/values", map[string]string{"value": web})
	call(ada, http.MethodPost, "/api/filters/hidden-repo-owners/values", map[string]string{"value": "gitlab.example.com/platform"})
	call(ada, http.MethodPost, "/api/filters/hidden-repo-owners/values", map[string]string{"value": "dev.azure.com/acme"})

	all := []string{"acme/Platform/rules#4", "acme/api#1", "acme/web#2", "globex/app#5", "platform/sub/svc#3"}
	assert.Equal(t, []string{"acme/api#1", "globex/app#5"}, prKeys(t, call(ada, http.MethodGet, "/api/prs", nil)),
		"ada hid acme/web, the platform group on gitlab.example.com and the acme Azure DevOps organization")
	assert.Equal(t, all, prKeys(t, call(bob, http.MethodGet, "/api/prs", nil)), "what ada hides is ada's alone")

	t.Run("the stream follows a change to what they hide", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/prs/stream", nil)
		require.NoError(t, err)
		req.Header.Set("X-Test-User", ada.String())
		req.Header.Set("X-Test-Role", "org:member")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close() //nolint:errcheck
		frames := make(chan string, 8)
		go func() {
			defer close(frames)
			scanner := bufio.NewScanner(resp.Body)
			scanner.Buffer(make([]byte, 1<<20), 1<<20)
			for scanner.Scan() {
				if data, ok := strings.CutPrefix(scanner.Text(), "data:"); ok && strings.Contains(data, `"prs"`) {
					frames <- strings.TrimSpace(data)
				}
			}
		}()
		next := func() string {
			t.Helper()
			select {
			case frame := <-frames:
				return frame
			case <-time.After(5 * time.Second):
				t.Fatal("no PR frame within 5s")
			}
			return ""
		}
		assert.Equal(t, []string{"acme/api#1", "globex/app#5"}, prKeys(t, next()))

		call(ada, http.MethodDelete, "/api/filters/hidden-repo-owners/values", map[string]string{"value": "dev.azure.com/acme"})
		assert.Equal(t, []string{"acme/Platform/rules#4", "acme/api#1", "globex/app#5"}, prKeys(t, next()), "unhiding the Azure DevOps organization brings its PR back")
	})
}
