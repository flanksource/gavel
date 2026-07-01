package github

import (
	nethttp "net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// page1 carries: a private repo whose default branch is failing, a public repo
// that's green, an archived repo (must be skipped), and an empty repo with no
// default branch (must be skipped). hasNextPage is true so the fetcher pages on.
const orgStatusPage1 = `{
  "data": { "organization": { "repositories": {
    "pageInfo": { "hasNextPage": true, "endCursor": "CUR1" },
    "nodes": [
      {
        "nameWithOwner": "acme/api", "isPrivate": true, "isArchived": false,
        "url": "https://github.com/acme/api", "pushedAt": "2026-06-08T00:00:00Z",
        "defaultBranchRef": { "name": "main", "target": { "statusCheckRollup": {
          "state": "FAILURE", "contexts": { "nodes": [
            {"__typename":"CheckRun","name":"build","status":"COMPLETED","conclusion":"SUCCESS","detailsUrl":""},
            {"__typename":"CheckRun","name":"test","status":"COMPLETED","conclusion":"FAILURE","detailsUrl":"https://github.com/acme/api/actions/runs/1"}
          ]}}}}
      },
      {
        "nameWithOwner": "acme/web", "isPrivate": false, "isArchived": false,
        "url": "https://github.com/acme/web", "pushedAt": "2026-06-07T00:00:00Z",
        "defaultBranchRef": { "name": "main", "target": { "statusCheckRollup": {
          "state": "SUCCESS", "contexts": { "nodes": [
            {"__typename":"CheckRun","name":"build","status":"COMPLETED","conclusion":"SUCCESS","detailsUrl":""}
          ]}}}}
      },
      {
        "nameWithOwner": "acme/old", "isPrivate": false, "isArchived": true,
        "url": "https://github.com/acme/old", "pushedAt": "2026-06-06T00:00:00Z",
        "defaultBranchRef": { "name": "main", "target": { "statusCheckRollup": null } }
      },
      {
        "nameWithOwner": "acme/empty", "isPrivate": false, "isArchived": false,
        "url": "https://github.com/acme/empty", "pushedAt": "2026-06-05T00:00:00Z",
        "defaultBranchRef": null
      }
    ]
  }}}
}`

// page2 carries a single repo pushed BEFORE the cutoff. It must be dropped, and
// because it's older than the cutoff the fetcher must stop paging here even
// though hasNextPage is true — a third request would be a bug.
const orgStatusPage2 = `{
  "data": { "organization": { "repositories": {
    "pageInfo": { "hasNextPage": true, "endCursor": "CUR2" },
    "nodes": [
      {
        "nameWithOwner": "acme/stale", "isPrivate": false, "isArchived": false,
        "url": "https://github.com/acme/stale", "pushedAt": "2026-01-01T00:00:00Z",
        "defaultBranchRef": { "name": "main", "target": { "statusCheckRollup": {
          "state": "SUCCESS", "contexts": { "nodes": [] } } } }
      }
    ]
  }}}
}`

func orgStatusServer(t *testing.T) (string, *int32) {
	t.Helper()
	var calls int32
	mux := nethttp.NewServeMux()
	mux.HandleFunc("/graphql", func(w nethttp.ResponseWriter, _ *nethttp.Request) {
		n := atomic.AddInt32(&calls, 1)
		switch n {
		case 1:
			_, _ = w.Write([]byte(orgStatusPage1))
		case 2:
			_, _ = w.Write([]byte(orgStatusPage2))
		default:
			t.Errorf("unexpected GraphQL call #%d: paging should have stopped at the stale page", n)
			nethttp.Error(w, "too many calls", nethttp.StatusInternalServerError)
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL, &calls
}

func TestFetchOrgDefaultBranchStatus(t *testing.T) {
	url, calls := orgStatusServer(t)
	t.Setenv("GITHUB_API_URL", url)

	since := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	res, _, err := FetchOrgDefaultBranchStatus(
		Options{Token: "test"},
		OrgStatusOptions{Org: "acme", Since: since},
	)
	require.NoError(t, err)

	// archived + empty + stale are excluded; only api and web survive.
	repos := make(map[string]RepoBranchStatus, len(res.Repos))
	for _, r := range res.Repos {
		repos[r.Repo] = r
	}
	assert.Len(t, res.Repos, 2)
	assert.Contains(t, repos, "acme/api")
	assert.Contains(t, repos, "acme/web")
	assert.NotContains(t, repos, "acme/old", "archived repo must be skipped")
	assert.NotContains(t, repos, "acme/empty", "repo with no default branch must be skipped")
	assert.NotContains(t, repos, "acme/stale", "repo pushed before --since must be dropped")

	assert.True(t, repos["acme/api"].Private, "private repo must be surfaced as private")
	require.NotNil(t, repos["acme/api"].CheckStatus)
	assert.Equal(t, 1, repos["acme/api"].CheckStatus.Failed)
	assert.Equal(t, 1, repos["acme/api"].CheckStatus.Passed)
	assert.Equal(t, "main", repos["acme/api"].DefaultBranch)

	require.NotNil(t, repos["acme/web"].CheckStatus)
	assert.Equal(t, 0, repos["acme/web"].CheckStatus.Failed)

	assert.Equal(t, int32(2), atomic.LoadInt32(calls), "must stop paging once a page predates the cutoff")
}

func TestFetchOrgDefaultBranchStatus_FailedOnly(t *testing.T) {
	url, _ := orgStatusServer(t)
	t.Setenv("GITHUB_API_URL", url)

	since := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	res, _, err := FetchOrgDefaultBranchStatus(
		Options{Token: "test"},
		OrgStatusOptions{Org: "acme", Since: since, FailedOnly: true},
	)
	require.NoError(t, err)

	require.Len(t, res.Repos, 1, "FailedOnly keeps only the failing repo")
	assert.Equal(t, "acme/api", res.Repos[0].Repo)
}
