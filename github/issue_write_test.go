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

// issueServer records the one request it receives and answers with issueJSON.
type issueServer struct {
	method   string
	path     string
	payload  map[string]any
	status   int
	response string
}

func (s *issueServer) start(t *testing.T) {
	t.Helper()
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		s.method, s.path = r.Method, r.URL.Path
		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 {
			require.NoError(t, json.Unmarshal(body, &s.payload))
		}
		w.WriteHeader(s.status)
		_, _ = w.Write([]byte(s.response))
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)
}

func TestSaveIssue_PostsWhenNoNumberIsGiven(t *testing.T) {
	srv := &issueServer{status: nethttp.StatusCreated, response: `{"number":42,
		"html_url":"https://github.com/acme/api/issues/42","title":"Fix parser",
		"state":"open","node_id":"I_node_42"}`}
	srv.start(t)

	out, err := SaveIssue(Options{Token: "tok", Repo: "acme/api"}, IssueInput{
		Title:  "Fix parser",
		Body:   "the parser drops trailing commas",
		Labels: []string{"bug", "parser"},
	})
	require.NoError(t, err)

	assert.Equal(t, nethttp.MethodPost, srv.method)
	assert.Equal(t, "/repos/acme/api/issues", srv.path)
	assert.Equal(t, "Fix parser", srv.payload["title"])
	assert.Equal(t, "the parser drops trailing commas", srv.payload["body"])
	assert.Equal(t, []any{"bug", "parser"}, srv.payload["labels"])

	assert.Equal(t, 42, out.Number)
	assert.Equal(t, "https://github.com/acme/api/issues/42", out.URL)
	assert.Equal(t, "I_node_42", out.NodeID)
	assert.Equal(t, "acme/api", out.Repo)
	assert.False(t, out.Updated)
}

func TestSaveIssue_PatchesTheNumberedIssue(t *testing.T) {
	srv := &issueServer{status: nethttp.StatusOK,
		response: `{"number":7,"html_url":"https://github.com/acme/api/issues/7","state":"open"}`}
	srv.start(t)

	out, err := SaveIssue(Options{Token: "tok", Repo: "acme/api"}, IssueInput{
		Number: 7,
		Title:  "Fix parser",
		Body:   "rewritten body",
	})
	require.NoError(t, err)

	assert.Equal(t, nethttp.MethodPatch, srv.method)
	assert.Equal(t, "/repos/acme/api/issues/7", srv.path)
	assert.Equal(t, "rewritten body", srv.payload["body"])
	assert.Equal(t, 7, out.Number)
	assert.True(t, out.Updated)
}

func TestSaveIssue_OmitsEmptyLabelsAndAssignees(t *testing.T) {
	srv := &issueServer{status: nethttp.StatusCreated,
		response: `{"number":1,"html_url":"https://github.com/acme/api/issues/1"}`}
	srv.start(t)

	_, err := SaveIssue(Options{Token: "tok", Repo: "acme/api"}, IssueInput{Title: "No labels"})
	require.NoError(t, err)

	assert.NotContains(t, srv.payload, "labels")
	assert.NotContains(t, srv.payload, "assignees")
}

func TestSaveIssue_SurfacesErrorBodyVerbatim(t *testing.T) {
	srv := &issueServer{status: nethttp.StatusUnprocessableEntity,
		response: `{"message":"Validation Failed","errors":[{"field":"labels"}]}`}
	srv.start(t)

	_, err := SaveIssue(Options{Token: "tok", Repo: "acme/api"}, IssueInput{Title: "Boom"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 422")
	assert.Contains(t, err.Error(), "Validation Failed")
}

func TestSaveIssue_HintsWhenIssuesAreDisabled(t *testing.T) {
	srv := &issueServer{status: nethttp.StatusGone, response: `{"message":"Issues are disabled for this repo"}`}
	srv.start(t)

	_, err := SaveIssue(Options{Token: "tok", Repo: "acme/api"}, IssueInput{Title: "Boom"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "issues are disabled on acme/api")
}

func TestSaveIssue_HintsWhenRepoIsNotFound(t *testing.T) {
	srv := &issueServer{status: nethttp.StatusNotFound, response: `{"message":"Not Found"}`}
	srv.start(t)

	_, err := SaveIssue(Options{Token: "tok", Repo: "acme/api"}, IssueInput{Title: "Boom"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "repo acme/api not found")
	assert.Contains(t, err.Error(), "issues: write")
}

func TestSaveIssue_HintsWhenTheEditedIssueIsNotFound(t *testing.T) {
	srv := &issueServer{status: nethttp.StatusNotFound, response: `{"message":"Not Found"}`}
	srv.start(t)

	_, err := SaveIssue(Options{Token: "tok", Repo: "acme/api"}, IssueInput{Number: 9, Title: "Boom"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update issue #9 on acme/api")
	assert.Contains(t, err.Error(), "issue acme/api#9 not found")
}

func TestSaveIssue_EmptyTitleFailsBeforeRequest(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(_ nethttp.ResponseWriter, _ *nethttp.Request) {
		t.Errorf("HTTP call should not happen for an empty title")
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)

	_, err := SaveIssue(Options{Token: "tok", Repo: "acme/api"}, IssueInput{Title: "  "})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Title is required")
}
