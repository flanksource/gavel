package github

import (
	"encoding/json"
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchCommitDiffUsesRESTDiffMediaType(t *testing.T) {
	const sha = "392799902c7fa82130fb4a43e6ea8734e11d1e98"
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		require.Equal(t, "/repos/flanksource/postgres/commits/"+sha, r.URL.Path)
		require.Equal(t, "application/vnd.github.diff", r.Header.Get("Accept"))
		_, _ = io.WriteString(w, "diff --git a/test/go.mod b/test/go.mod\n+github.com/jackc/pgx/v5 v5.7.5\n")
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)

	payload, err := FetchCommitDiff(Options{Token: "tok", Repo: "flanksource/postgres"}, sha)
	require.NoError(t, err)
	assert.Contains(t, payload.Diff, "test/go.mod")
	assert.False(t, payload.Truncated)
	assert.False(t, payload.Binary)
}

func TestFetchPRFilePatchFindsFileAcrossPages(t *testing.T) {
	pages := 0
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		require.Equal(t, "/repos/flanksource/postgres/pulls/35/files", r.URL.Path)
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		pages++
		w.Header().Set("Content-Type", "application/json")
		if page == 1 {
			files := make([]restPRFile, 100)
			for i := range files {
				files[i] = restPRFile{Filename: "docs/file-" + strconv.Itoa(i) + ".md", Status: "modified", Patch: stringPtr("@@ noop @@")}
			}
			_ = json.NewEncoder(w).Encode(files)
			return
		}
		_ = json.NewEncoder(w).Encode([]restPRFile{
			{
				Filename:  "test/go.mod",
				Status:    "modified",
				Additions: 1,
				Deletions: 1,
				Patch:     stringPtr("@@ -1 +1 @@\n-github.com/jackc/pgx/v5 v5.7.4\n+github.com/jackc/pgx/v5 v5.7.5"),
			},
		})
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)

	payload, err := FetchPRFilePatch(Options{Token: "tok", Repo: "flanksource/postgres"}, 35, "test/go.mod")
	require.NoError(t, err)
	assert.Equal(t, 2, pages)
	assert.Contains(t, payload.Diff, "diff --git a/test/go.mod b/test/go.mod")
	assert.Contains(t, payload.Diff, "+github.com/jackc/pgx/v5 v5.7.5")
}

func TestFetchPRFilePatchReturnsBinaryForMissingPatch(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]restPRFile{
			{Filename: "image.png", Status: "modified"},
		})
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)

	payload, err := FetchPRFilePatch(Options{Token: "tok", Repo: "flanksource/postgres"}, 35, "image.png")
	require.NoError(t, err)
	assert.True(t, payload.Binary)
	assert.Empty(t, payload.Diff)
}

func stringPtr(s string) *string {
	return &s
}
