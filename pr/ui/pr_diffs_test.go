package ui

import (
	"encoding/json"
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlePRCommitDiffValidatesParams(t *testing.T) {
	s := &Server{}
	rec := httptest.NewRecorder()
	s.handlePRCommitDiff(rec, httptest.NewRequest(nethttp.MethodGet, "/api/prs/commits/diff?repo=bad&sha=not-a-sha", nil))

	assert.Equal(t, nethttp.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "repo must be owner/name")
}

func TestHandlePRCommitDiffReturnsPayload(t *testing.T) {
	const sha = "392799902c7fa82130fb4a43e6ea8734e11d1e98"
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		require.Equal(t, "/repos/flanksource/postgres/commits/"+sha, r.URL.Path)
		require.Equal(t, "application/vnd.github.diff", r.Header.Get("Accept"))
		_, _ = io.WriteString(w, "diff --git a/test/go.mod b/test/go.mod\n+pgx\n")
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)

	s := &Server{ghOpts: github.Options{Token: "tok"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(nethttp.MethodGet, "/api/prs/commits/diff?repo=flanksource%2Fpostgres&sha="+sha, nil)
	s.handlePRCommitDiff(rec, req)

	require.Equal(t, nethttp.StatusOK, rec.Code)
	var resp prDiffResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "flanksource/postgres", resp.Repo)
	assert.Equal(t, sha, resp.Commit)
	assert.Contains(t, resp.Diff, "test/go.mod")
}

func TestHandlePRFileDiffReturnsPayload(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		require.Equal(t, "/repos/flanksource/postgres/pulls/35/files", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{
				"filename":  "test/go.mod",
				"status":    "modified",
				"additions": 1,
				"deletions": 1,
				"patch":     "@@ -1 +1 @@\n-old\n+new",
			},
		})
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)

	s := &Server{ghOpts: github.Options{Token: "tok"}}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(nethttp.MethodGet, "/api/prs/files/diff?repo=flanksource%2Fpostgres&number=35&path=test%2Fgo.mod", nil)
	s.handlePRFileDiff(rec, req)

	require.Equal(t, nethttp.StatusOK, rec.Code)
	var resp prDiffResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "flanksource/postgres", resp.Repo)
	assert.Equal(t, 35, resp.Number)
	assert.Equal(t, "test/go.mod", resp.Path)
	assert.Contains(t, resp.Diff, "+new")
}
