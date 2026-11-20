// Tests an Azure DevOps pull request's dashboard detail, read through its connection's provider:
// the PR, the CI its pipelines ran with the tails of failed steps' logs, and its comments.
package ui

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/gavel/connections"
	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/tenant"
	"github.com/flanksource/gavel/todos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	azureHead  = "1111111111111111111111111111111111111111"
	azureMerge = "2222222222222222222222222222222222222222"
)

// fakeAzure is an Azure DevOps API answering "METHOD path" routes.
func fakeAzure(t *testing.T, routes map[string]string) *httptest.Server {
	t.Helper()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := routes[r.Method+" "+r.URL.Path]
		if !ok {
			http.Error(w, "no route "+r.Method+" "+r.URL.Path, http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return server
}

func azureRemote(t *testing.T, server *httptest.Server) *git.Repository {
	t.Helper()
	client, err := git.FromConnection(dbcontext.New(), models.Connection{Name: "acme", Type: models.ConnectionTypeAzureDevops,
		URL: server.URL + "/acme", Password: "azure-pat", InsecureTLS: true})
	require.NoError(t, err)
	remote, err := client.Repository(server.URL + "/acme/Platform/_git/rules")
	require.NoError(t, err)
	return remote
}

// azurePRRoutes are the Azure DevOps API routes of pull request 9 of Platform/rules: its CI
// failed in the compile step, a review thread on rules.go is open, a discussion is resolved, and
// a system note was left on it.
func azurePRRoutes() map[string]string {
	builds := `{"value":[{"id":31,"buildNumber":"b31","status":"completed","result":"failed","sourceBranch":"refs/pull/9/merge",
		"sourceVersion":"` + azureMerge + `","queueTime":"2026-11-02T11:00:00Z","definition":{"id":4,"name":"ci"},
		"_links":{"web":{"href":"https://dev.azure.com/acme/Platform/_build/results?buildId=31"}}}]}`
	return map[string]string{
		"GET /acme/Platform/_apis/git/repositories/rules/pullrequests/9": `{"pullRequestId":9,"status":"active","title":"Tighten the rules",
			"description":"Closes the gap.","isDraft":true,"mergeStatus":"conflicts","creationDate":"2026-11-02T10:00:00Z",
			"sourceRefName":"refs/heads/tighten","targetRefName":"refs/heads/main","createdBy":{"displayName":"Ada Lovelace","imageUrl":"https://avatars.example/ada"},
			"lastMergeSourceCommit":{"commitId":"` + azureHead + `"},"lastMergeCommit":{"commitId":"` + azureMerge + `"},
			"repository":{"id":"rules-id","name":"rules","project":{"name":"Platform"}}}`,
		"GET /acme/Platform/_apis/git/repositories/rules/pullrequests/9/threads": `{"value":[
			{"id":5,"status":"active","threadContext":{"filePath":"/rules.go","rightFileStart":{"line":12}},"comments":[
				{"id":1,"content":"Please split this up.","commentType":"text","publishedDate":"2026-11-02T12:00:00Z",
				 "author":{"displayName":"Grace Hopper","imageUrl":"https://avatars.example/grace"}}]},
			{"id":6,"status":"fixed","comments":[
				{"id":1,"content":"Rename the flag?","commentType":"text","publishedDate":"2026-11-02T12:30:00Z","author":{"displayName":"Ada Lovelace"}}]},
			{"id":7,"comments":[
				{"id":1,"content":"Updated the PR","commentType":"system","publishedDate":"2026-11-02T12:05:00Z","author":{"displayName":"Azure"}}]}]}`,
		"GET /acme/Platform/_apis/build/builds":                                    builds,
		"GET /acme/_apis/projects/Platform":                                        `{"id":"platform-id"}`,
		"GET /acme/Platform/_apis/policy/evaluations":                              `{"value":[]}`,
		"GET /acme/Platform/_apis/git/repositories/rules/pullrequests/9/reviewers": `{"value":[]}`,
		"GET /acme/Platform/_apis/git/repositories/rules-id/refs":                  `{"value":[{"name":"refs/heads/tighten","objectId":"` + azureHead + `"}]}`,
		"GET /acme/Platform/_apis/build/builds/31/timeline": `{"records":[
			{"id":"job-2","type":"Job","name":"build","state":"completed","result":"failed","order":1,"log":{"id":7}},
			{"id":"checkout","parentId":"job-2","type":"Task","name":"checkout","state":"completed","result":"succeeded","order":1,"log":{"id":9}},
			{"id":"compile","parentId":"job-2","type":"Task","name":"compile","state":"completed","result":"failed","order":2,"log":{"id":8}}]}`,
		"GET /acme/Platform/_apis/build/builds/31/logs":   `{"value":[{"id":8,"lineCount":2}]}`,
		"GET /acme/Platform/_apis/build/builds/31/logs/8": "go build ./...\nundefined: rule\n",
	}
}

func TestAnAzurePullRequestsDetailComesThroughItsConnection(t *testing.T) {
	server := fakeAzure(t, azurePRRoutes())
	remote := azureRemote(t, server)

	detail, err := connectionPRDetail(context.Background(), remote, 9, nil)
	require.NoError(t, err)
	require.NotNil(t, detail.PR)
	pr := detail.PR
	assert.Equal(t, 9, pr.Number)
	assert.Equal(t, "azuredevops", pr.Provider, "the provider the PR actions read")
	assert.Equal(t, "Tighten the rules", pr.Title)
	assert.Equal(t, "Closes the gap.", pr.Body)
	assert.Equal(t, "OPEN", pr.State)
	assert.True(t, pr.IsDraft)
	assert.Equal(t, "CONFLICTING", pr.Mergeable)
	assert.Equal(t, "tighten", pr.HeadRefName)
	assert.Equal(t, "main", pr.BaseRefName)
	assert.Equal(t, azureHead, pr.HeadRefOID)
	assert.Equal(t, model.PRAuthor{Name: "Ada Lovelace", AvatarURL: "https://avatars.example/ada"}, pr.Author)
	assert.Equal(t, server.URL+"/acme/Platform/_git/rules/pullrequest/9", pr.URL)

	jobURL := "https://dev.azure.com/acme/Platform/_build/results?buildId=31&view=logs&j=job-2"
	assert.Equal(t, model.StatusChecks{{RunID: 31, JobID: "job-2", Name: "build", WorkflowName: "ci", Status: "COMPLETED",
		Conclusion: "FAILURE", DetailsURL: jobURL}}, pr.StatusCheckRollup)
	require.Contains(t, detail.Runs, int64(31))
	assert.Equal(t, &model.WorkflowRun{DatabaseID: 31, WorkflowID: 4, Name: "ci", Status: "COMPLETED", Conclusion: "FAILURE",
		URL: "https://dev.azure.com/acme/Platform/_build/results?buildId=31", HeadSHA: azureMerge,
		Jobs: []model.Job{{NativeID: "job-2", Name: "build", Status: "COMPLETED", Conclusion: "FAILURE", URL: jobURL, Steps: []model.Step{
			{Name: "checkout", Number: 1, Status: "COMPLETED", Conclusion: "SUCCESS"},
			{Name: "compile", Number: 2, Status: "COMPLETED", Conclusion: "FAILURE", Logs: "go build ./...\nundefined: rule"},
		}}}}, detail.Runs[31], "a failed step carries its log's tail")

	threads := []model.PRComment{
		{ThreadID: "5", ID: 1, Body: "Please split this up.", Author: "Grace Hopper", AvatarURL: "https://avatars.example/grace",
			URL: server.URL + "/acme/Platform/_git/rules/pullrequest/9?discussionId=5", CreatedAt: time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC),
			Path: "/rules.go", Line: 12, IsReviewThread: true, CanResolve: true},
		{ThreadID: "6", ID: 1, Body: "Rename the flag?", Author: "Ada Lovelace",
			URL: server.URL + "/acme/Platform/_git/rules/pullrequest/9?discussionId=6", CreatedAt: time.Date(2026, 11, 2, 12, 30, 0, 0, time.UTC),
			IsResolved: true, IsReviewThread: true, CanResolve: true},
	}
	assert.Equal(t, threads, pr.ReviewThreads, "a system note isn't a comment")
	assert.Equal(t, threads, pr.Comments, "every Azure DevOps comment is in a thread")
	require.NotEmpty(t, detail.Comments)
	assert.Equal(t, "Please split this up.", detail.Comments[0].Body, "the open review thread is actionable")
}

func TestTheDashboardServesAnAzurePullRequestsDetailThroughItsRemote(t *testing.T) {
	server := fakeAzure(t, azurePRRoutes())
	s := NewServer(gavelctx.New(t.Context()), 0, github.Options{}, SearchConfig{})
	listed := github.PRListItem{Provider: types.PRProviderAzure, Repo: "acme/Platform/rules", Number: 9,
		URL: server.URL + "/acme/Platform/_git/rules/pullrequest/9"}
	s.publishResults(github.PRSearchResults{listed})

	detail := s.fetchPRDetail("acme/Platform/rules", 9)
	assert.Equal(t, "this dashboard has no connection to read azure pull request acme/Platform/rules#9 through", detail.Error)

	var asked []github.PRListItem
	s.SetPRRemote(func(_ context.Context, pr github.PRListItem) (*git.Repository, error) {
		asked = append(asked, pr)
		return azureRemote(t, server), nil
	})
	detail = s.fetchPRDetail("acme/Platform/rules", 9)
	require.Empty(t, detail.Error)
	assert.Equal(t, "Tighten the rules", detail.PR.Title)
	assert.Contains(t, detail.Runs, int64(31))
	assert.Equal(t, []github.PRListItem{listed}, asked, "the remote is asked for the listed PR")

	recorder := httptest.NewRecorder()
	s.handleDetail(recorder, httptest.NewRequest(http.MethodGet, "/api/prs/detail?repo=acme/Platform/rules&number=9", nil))
	stream := recorder.Body.String()
	assert.Contains(t, stream, "event: pr\n")
	assert.Contains(t, stream, `"title":"Tighten the rules"`)
	assert.Contains(t, stream, "event: runs\n")
	assert.Contains(t, stream, "event: done\n")
	assert.NotContains(t, stream, "event: error")
}

func TestTheDashboardMergesAnAzurePullRequestThroughItsRemote(t *testing.T) {
	var completion map[string]any
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/acme/Platform/_apis/git/repositories/rules/pullrequests/9" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&completion)
		_, _ = w.Write([]byte(`{"pullRequestId":9,"status":"completed"}`))
	}))
	t.Cleanup(server.Close)
	s := NewServer(gavelctx.New(t.Context()), 0, github.Options{}, SearchConfig{})
	prURL := server.URL + "/acme/Platform/_git/rules/pullrequest/9"
	s.publishResults(github.PRSearchResults{{Provider: types.PRProviderAzure, Repo: "acme/Platform/rules", Number: 9, URL: prURL}})
	s.SetPRRemote(func(context.Context, github.PRListItem) (*git.Repository, error) { return azureRemote(t, server), nil })

	recorder := httptest.NewRecorder()
	s.handlePRMerge(recorder, httptest.NewRequest(http.MethodPost, "/api/prs/merge",
		strings.NewReader(`{"url":"`+prURL+`","method":"squash","headSha":"`+azureHead+`"}`)))
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	assert.Equal(t, "completed", completion["status"])
	assert.Equal(t, map[string]any{"commitId": azureHead}, completion["lastMergeSourceCommit"], "the head that was reviewed")
	assert.Equal(t, "squash", completion["completionOptions"].(map[string]any)["mergeStrategy"])

	recorder = httptest.NewRecorder()
	s.handlePRMerge(recorder, httptest.NewRequest(http.MethodPost, "/api/prs/merge",
		strings.NewReader(`{"url":"`+prURL+`","method":"squash","auto":true,"headSha":"`+azureHead+`"}`)))
	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "auto-merge isn't offered for Azure DevOps pull requests")
}

func TestTheDashboardApprovesAnAzurePullRequestAsItsConnection(t *testing.T) {
	const me = "8f1c2a8e-6b2d-4c1f-9a77-2f6f0b8d5e11"
	var vote map[string]any
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /acme/_apis/connectionData":
			_, _ = w.Write([]byte(`{"authenticatedUser":{"id":"` + me + `"}}`))
		case "PUT /acme/Platform/_apis/git/repositories/rules/pullrequests/9/reviewers/" + me:
			_ = json.NewDecoder(r.Body).Decode(&vote)
			_, _ = w.Write([]byte(`{"id":"` + me + `","vote":10}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	s := NewServer(gavelctx.New(t.Context()), 0, github.Options{}, SearchConfig{})
	s.publishResults(github.PRSearchResults{{Provider: types.PRProviderAzure, Repo: "acme/Platform/rules", Number: 9,
		URL: server.URL + "/acme/Platform/_git/rules/pullrequest/9"}})
	s.SetPRRemote(func(context.Context, github.PRListItem) (*git.Repository, error) { return azureRemote(t, server), nil })

	recorder := httptest.NewRecorder()
	s.handlePRApprove(recorder, httptest.NewRequest(http.MethodPost, "/api/prs/approve",
		strings.NewReader(`{"repo":"acme/Platform/rules","number":9}`)))
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	assert.Equal(t, map[string]any{"vote": float64(10)}, vote, "an Azure pull request needs no GitHub node ID")
}

func TestAHostedOrgReachesItsAzurePullRequestsByURL(t *testing.T) {
	db, _, org := openOrgDB(t)
	t.Setenv("GAVEL_SECRET_KEEPER_CONNECTION", "")
	t.Setenv("GAVEL_SECRET_KEY", base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("h"), 32)))
	server := fakeAzure(t, azurePRRoutes())
	scope := tenant.Scope{OrgID: org, Role: "org:admin"}
	store, closeStore, err := connections.OpenStore(t.Context(), connections.Options{DB: db, Scope: scope})
	require.NoError(t, err)
	saved, err := store.SaveConnection(t.Context(), &models.Connection{Name: "acme-azure", Type: models.ConnectionTypeAzureDevops,
		URL: server.URL + "/acme", Password: "azure-pat", InsecureTLS: true})
	require.NoError(t, err)
	require.NoError(t, closeStore())
	require.NoError(t, db.Exec(`INSERT INTO repos (org_id, name, source, connection_id, remote_url, provider_repo_id, default_branch)
		VALUES (?, 'Platform/rules', 'azure_devops', ?, ?, 'rules-id', 'main')`, org, saved.ID, server.URL+"/acme/Platform/_git/rules").Error)
	prs := OrganizationPRs(nil, nil, OrgConnections{DB: db, Scope: scope})

	client, number, err := prs(t.Context(), server.URL+"/acme/Platform/_git/rules/pullrequest/9")
	require.NoError(t, err)
	assert.Equal(t, 9, number)
	assert.Equal(t, "azuredevops", client.Kind())
	branch, err := client.DefaultBranch(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "main", branch)
	status, err := client.Status(t.Context(), number, model.StatusOptions{Threads: true})
	require.NoError(t, err)
	assert.Equal(t, "Tighten the rules", status.PR.Title)

	_, _, err = prs(t.Context(), "https://github.com/acme/api/pull/7")
	require.ErrorContains(t, err, "this organization has no GitHub App installation to read")
}

func TestAHostedOrgReachesItsGitLabMergeRequestsByURL(t *testing.T) {
	db, _, org := openOrgDB(t)
	t.Setenv("GAVEL_SECRET_KEEPER_CONNECTION", "")
	t.Setenv("GAVEL_SECRET_KEY", base64.StdEncoding.EncodeToString(bytes.Repeat([]byte("h"), 32)))
	var server *httptest.Server
	server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v4/projects/acme/api/merge_requests/5" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"iid":5,"state":"opened","title":"Speed up the API","web_url":"` + server.URL + `/acme/api/-/merge_requests/5",
			"source_branch":"fast","target_branch":"main","source_project_id":1,"target_project_id":1,
			"sha":"2222222222222222222222222222222222222222","author":{"name":"Grace Hopper"}}`))
	}))
	t.Cleanup(server.Close)
	scope := tenant.Scope{OrgID: org, Role: "org:admin"}
	store, closeStore, err := connections.OpenStore(t.Context(), connections.Options{DB: db, Scope: scope})
	require.NoError(t, err)
	saved, err := store.SaveConnection(t.Context(), &models.Connection{Name: "acme-gitlab", Type: models.ConnectionTypeGitlab,
		URL: server.URL, Password: "glpat", InsecureTLS: true})
	require.NoError(t, err)
	require.NoError(t, closeStore())
	require.NoError(t, db.Exec(`INSERT INTO repos (org_id, name, source, connection_id, remote_url, provider_repo_id, default_branch)
		VALUES (?, 'acme/api', 'gitlab', ?, ?, '1', 'main')`, org, saved.ID, server.URL+"/acme/api.git").Error)
	prs := OrganizationPRs(nil, nil, OrgConnections{DB: db, Scope: scope})

	client, number, err := prs(t.Context(), server.URL+"/acme/api/-/merge_requests/5")
	require.NoError(t, err)
	assert.Equal(t, 5, number)
	assert.Equal(t, "gitlab", client.Kind())
	status, err := client.Status(t.Context(), number, model.StatusOptions{})
	require.NoError(t, err)
	assert.Equal(t, "Speed up the API", status.PR.Title)
	assert.Equal(t, "2222222222222222222222222222222222222222", status.PR.HeadRefOID)

	_, _, err = prs(t.Context(), server.URL+"/acme/api/-/merge_requests/none")
	require.EqualError(t, err, `pull request URL "`+server.URL+`/acme/api/-/merge_requests/none" has no positive number`)
}
