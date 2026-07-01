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

func TestMergeMethodFor(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"rebase", "REBASE", false},
		{"squash", "SQUASH", false},
		{"merge", "MERGE", false},
		{"REBASE", "REBASE", false},
		{"  Squash  ", "SQUASH", false},
		{"fast-forward", "", true},
		{"", "", true},
	}
	for _, tc := range cases {
		got, err := MergeMethodFor(tc.in)
		if tc.wantErr {
			require.Error(t, err, "input %q", tc.in)
			assert.Contains(t, err.Error(), tc.in)
			continue
		}
		require.NoError(t, err, "input %q", tc.in)
		assert.Equal(t, tc.want, got, "input %q", tc.in)
	}
}

func TestEnableAutoMerge_SendsMutation(t *testing.T) {
	var gotVars map[string]any
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		require.Equal(t, "/graphql", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Variables map[string]any `json:"variables"`
		}
		require.NoError(t, json.Unmarshal(body, &payload))
		gotVars = payload.Variables
		_, _ = w.Write([]byte(`{"data":{"enablePullRequestAutoMerge":{"pullRequest":{"number":7}}}}`))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)

	err := EnableAutoMerge(Options{Token: "tok"}, "PR_node_1", "rebase")
	require.NoError(t, err)
	assert.Equal(t, "PR_node_1", gotVars["prId"])
	assert.Equal(t, "REBASE", gotVars["method"])
}

func TestEnableAutoMerge_SurfacesGraphQLError(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = w.Write([]byte(`{"errors":[{"message":"Auto merge is not allowed for this repository"}]}`))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)

	err := EnableAutoMerge(Options{Token: "tok"}, "PR_node_1", "rebase")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Auto merge is not allowed")
}

func TestEnableAutoMerge_EmptyNodeIDFailsBeforeRequest(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		t.Errorf("HTTP call should not happen for an empty node ID")
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)

	err := EnableAutoMerge(Options{Token: "tok"}, "  ", "rebase")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "node ID is required")
}

func TestEnableAutoMerge_InvalidMergeType(t *testing.T) {
	err := EnableAutoMerge(Options{Token: "tok"}, "PR_node_1", "fast-forward")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid merge type")
}
