package github

import (
	"encoding/json"
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergePullRequest_SendsMutation(t *testing.T) {
	var gotVars map[string]any
	var gotQuery string
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		require.Equal(t, "/graphql", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		require.NoError(t, json.Unmarshal(body, &payload))
		gotVars = payload.Variables
		gotQuery = payload.Query
		_, _ = w.Write([]byte(`{"data":{"mergePullRequest":{"pullRequest":{"number":7,"state":"MERGED"}}}}`))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)

	err := MergePullRequest(Options{Token: "tok"}, "PR_node_1", "squash")
	require.NoError(t, err)
	assert.Contains(t, gotQuery, "mergePullRequest")
	assert.Equal(t, "PR_node_1", gotVars["prId"])
	assert.Equal(t, "SQUASH", gotVars["method"])
}

func TestMergePullRequest_SurfacesGraphQLError(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = w.Write([]byte(`{"errors":[{"message":"Pull request is not mergeable"}]}`))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)

	err := MergePullRequest(Options{Token: "tok"}, "PR_node_1", "rebase")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Pull request is not mergeable")
}

func TestMergePullRequest_EmptyNodeIDFailsBeforeRequest(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		t.Errorf("HTTP call should not happen for an empty node ID")
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)

	err := MergePullRequest(Options{Token: "tok"}, "  ", "rebase")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "node ID is required")
}

func TestMergePullRequest_InvalidMergeType(t *testing.T) {
	err := MergePullRequest(Options{Token: "tok"}, "PR_node_1", "fast-forward")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid merge type")
}

func TestApprovePullRequest_SendsMutation(t *testing.T) {
	var gotVars map[string]any
	var gotQuery string
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		require.Equal(t, "/graphql", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		require.NoError(t, json.Unmarshal(body, &payload))
		gotVars = payload.Variables
		gotQuery = payload.Query
		_, _ = w.Write([]byte(`{"data":{"addPullRequestReview":{"pullRequestReview":{"state":"APPROVED"}}}}`))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)

	err := ApprovePullRequest(Options{Token: "tok"}, "PR_node_1", "LGTM")
	require.NoError(t, err)
	assert.Contains(t, gotQuery, "addPullRequestReview")
	assert.Equal(t, "PR_node_1", gotVars["prId"])
	assert.Equal(t, "LGTM", gotVars["body"])
}

func TestApprovePullRequest_SurfacesGraphQLError(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = w.Write([]byte(`{"errors":[{"message":"Can not approve your own pull request"}]}`))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)

	err := ApprovePullRequest(Options{Token: "tok"}, "PR_node_1", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Can not approve your own pull request")
}

func TestApprovePullRequest_EmptyNodeIDFailsBeforeRequest(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		t.Errorf("HTTP call should not happen for an empty node ID")
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)

	err := ApprovePullRequest(Options{Token: "tok"}, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "node ID is required")
}
