// Tests the todo entity on a hosted dashboard: its operations act for the request's organization
// and user, its runs go to a sandbox, and actions that run code on serve or read its files are refused.
package ui

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hostedRaw calls base as user, returning the status and the raw body.
func hostedRaw(t *testing.T, base, method, path string, user uuid.UUID) (int, string) {
	t.Helper()
	req, err := http.NewRequest(method, base+path, nil)
	require.NoError(t, err)
	req.Header.Set("X-Test-User", user.String())
	req.Header.Set("X-Test-Role", "org:member")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(body)
}

// bulkOutcome decodes a bulk action's answer.
func bulkOutcome(t *testing.T, body string) (outcome struct {
	Applied int
	Results []struct{ Error string }
}) {
	t.Helper()
	require.NoError(t, json.Unmarshal([]byte(body), &outcome), body)
	return outcome
}

func TestHostedTodoEntityActsForTheOrgsUser(t *testing.T) {
	h := newHostedRuns(t, nil)
	_, private := hostedCall(t, h.base, http.MethodPost, "/api/todos", h.ada, "org:member", map[string]any{"dir": h.workspace, "title": "Private plan"})
	_, shared := hostedCall(t, h.base, http.MethodPost, "/api/todos", h.ada, "org:member", map[string]any{"dir": h.workspace, "title": "Shared plan", "visibility": "public"})
	sharedID := shared["id"].(string)

	status, catalog := hostedRaw(t, h.base, http.MethodGet, "/api/entities", h.grace)
	require.Equal(t, http.StatusOK, status, catalog)
	var entities []struct{ Name string }
	require.NoError(t, json.Unmarshal([]byte(catalog), &entities))
	assert.Contains(t, entities, struct{ Name string }{"todo"})

	status, body := hostedRaw(t, h.base, http.MethodPost, "/api/v1/todo?dir="+h.workspace+"&title=Reads+a+file&body=@/etc/hostname", h.ada)
	require.Equal(t, http.StatusOK, status, body)
	var created struct {
		ID           string `json:"id"`
		MarkdownBody string `json:"markdown_body"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &created), body)
	_, item := hostedCall(t, h.base, http.MethodGet, "/api/todos/item?dir="+h.workspace+"&ref="+created.ID, h.ada, "org:member", nil)
	assert.Equal(t, "@/etc/hostname", item["body"], "a hosted todo's body is its text, never a file on serve")

	status, listed := hostedRaw(t, h.base, http.MethodGet, "/api/v1/todo?dir="+h.workspace, h.grace)
	require.Equal(t, http.StatusOK, status, listed)
	assert.Contains(t, listed, "Shared plan")
	assert.NotContains(t, listed, "Private plan", "another member doesn't list a private todo")

	status, body = hostedRaw(t, h.base, http.MethodPost, "/api/v1/todo/"+sharedID+"/priority?to=high", h.grace)
	require.Equal(t, http.StatusOK, status, body)
	refused := bulkOutcome(t, body)
	assert.Equal(t, 0, refused.Applied, "another member can't change the creator's todo: %s", body)
	require.Len(t, refused.Results, 1)
	assert.Contains(t, refused.Results[0].Error, "only its creator or an org admin can change it")

	status, body = hostedRaw(t, h.base, http.MethodPost, "/api/v1/todo/"+private["id"].(string)+"/priority?to=high", h.ada)
	require.Equal(t, http.StatusOK, status, body)
	assert.Equal(t, 1, bulkOutcome(t, body).Applied, body)

	status, body = hostedRaw(t, h.base, http.MethodPost, "/api/v1/todo/"+sharedID+"/plan", h.ada)
	require.Equal(t, http.StatusOK, status, body)
	assert.Equal(t, 1, bulkOutcome(t, body).Applied, "the plan started: %s", body)
	var runs int64
	require.Eventually(t, func() bool {
		require.NoError(t, h.db.Raw(`SELECT count(*) FROM sandbox_runs WHERE todo_id = ?`, sharedID).Scan(&runs).Error)
		return runs == 1
	}, 30*time.Second, 50*time.Millisecond, "the plan runs in a sandbox job")

	status, body = hostedRaw(t, h.base, http.MethodPost, "/api/v1/todo/steps?dir="+h.workspace, h.ada)
	require.Equal(t, http.StatusOK, status, body)
	assert.Contains(t, body, `"plan"`, "the repository's lifecycle")

	for _, path := range []string{"/" + sharedID + "/check", "/" + sharedID + "/merge", "/sync", "/import", "/" + sharedID + "/export"} {
		status, body := hostedRaw(t, h.base, http.MethodPost, "/api/v1/todo"+path, h.ada)
		assert.Equal(t, http.StatusNotImplemented, status, "%s: %s", path, body)
		assert.Contains(t, body, "hosted mode", path)
	}
}

func TestHostedTodosOnAPushedRepositoryHaveNoGitHubIssues(t *testing.T) {
	h := newHostedRuns(t, nil)
	_, shared := hostedCall(t, h.base, http.MethodPost, "/api/todos", h.ada, "org:member", map[string]any{"dir": h.workspace, "title": "Shared", "visibility": "public"})
	ref := shared["id"].(string)
	const noIssues = "repository app was pushed to gavel, not linked from GitHub, so its todos have no GitHub issues"

	status, refused := hostedCall(t, h.base, http.MethodPost, "/api/todos/github", h.grace, "org:member", map[string]any{"dir": h.workspace, "ref": ref})
	assert.Equal(t, http.StatusForbidden, status, "another member can't push it: %v", refused)

	status, refused = hostedCall(t, h.base, http.MethodPost, "/api/todos/github", h.ada, "org:member", map[string]any{"dir": h.workspace, "ref": ref})
	assert.Equal(t, http.StatusBadRequest, status, refused)
	assert.Equal(t, noIssues, refused["error"])

	status, body := hostedRaw(t, h.base, http.MethodPost, "/api/v1/todo/"+ref+"/push", h.ada)
	require.Equal(t, http.StatusOK, status, body)
	outcome := bulkOutcome(t, body)
	require.Len(t, outcome.Results, 1, body)
	assert.Equal(t, noIssues, outcome.Results[0].Error)
}

func TestHostedTodoCatalogOffersOnlyServedActions(t *testing.T) {
	h := newHostedRuns(t, nil)
	status, body := hostedRaw(t, h.base, http.MethodGet, "/api/entities", h.ada)
	require.Equal(t, http.StatusOK, status, body)
	var entities []struct {
		Name        string
		Actions     []struct{ Name string } `json:"actions"`
		BulkActions []struct{ Name string } `json:"bulk_actions"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &entities))
	var offered []string
	for _, entity := range entities {
		if entity.Name != "todo" {
			continue
		}
		for _, action := range append(entity.Actions, entity.BulkActions...) {
			offered = append(offered, action.Name)
		}
	}
	assert.Contains(t, offered, "status")
	assert.Contains(t, offered, "push")
	for _, refused := range []string{"check", "merge", "sync", "import", "export"} {
		assert.NotContains(t, offered, refused, "a hosted organization isn't offered %s", refused)
	}
}
