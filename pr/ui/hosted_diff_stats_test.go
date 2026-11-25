// Tests a hosted todo's diff stats: its run is measured in serve's copy of the organization's
// repository, where the run's commits come back, and a run that can't be measured hides nothing.
package ui

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// listedDiff is the diff stats todo carries in a todo list response, and whether it is listed.
func listedDiff(items []any, todo uuid.UUID) (any, bool) {
	for _, item := range items {
		summary := item.(map[string]any)
		if summary["id"] == todo.String() {
			return summary["diff"], true
		}
	}
	return nil, false
}

func TestHostedTodosShowTheirRunsDiffStats(t *testing.T) {
	h := newHostedLanding(t)
	todo, _, _ := h.run(t, "Fix the build", "fix.go", "fix_test.go")
	want := map[string]any{"commits": float64(2), "files": float64(2), "adds": float64(2), "dels": float64(0)}

	status, listed := hostedCall(t, h.base, http.MethodGet, "/api/todos?dir="+h.workspace, h.creator, "org:member", nil)
	require.Equal(t, http.StatusOK, status, listed)
	diff, ok := listedDiff(listed["items"].([]any), todo)
	require.True(t, ok, "the todo is listed: %v", listed)
	assert.Equal(t, want, diff, "measured in serve's copy, not at the sandbox's /workspace/src")

	status, batch := hostedCall(t, h.base, http.MethodPost, "/api/todos/batch", h.creator, "org:member", map[string]any{"dirs": []string{h.workspace}})
	require.Equal(t, http.StatusOK, status, batch)
	result := batch["results"].([]any)[0].(map[string]any)
	assert.Nil(t, result["error"])
	diff, ok = listedDiff(result["items"].([]any), todo)
	require.True(t, ok, "the todo is listed: %v", result)
	assert.Equal(t, want, diff)

	status, detail := hostedCall(t, h.base, http.MethodGet, "/api/todos/item?dir="+h.workspace+"&ref="+todo.String(), h.creator, "org:member", nil)
	require.Equal(t, http.StatusOK, status, detail)
	assert.Equal(t, want, detail["diff"])
}

func TestHostedTodosStayListedWhenTheirRunCantBeMeasured(t *testing.T) {
	h := newHostedLanding(t)
	todo, _, head := h.run(t, "Fix the build", "fix.go")
	// A run whose recorded start is in no repository serve has.
	workspace, err := json.Marshal(map[string]any{"cwd": "/workspace/src", "worktree": map[string]any{
		"repo": "/workspace/src", "branch": "agent", "setup": "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", "head": head,
	}})
	require.NoError(t, err)
	require.NoError(t, h.db.Exec(`UPDATE captain_prompt_runs SET workspace = ?`, string(workspace)).Error)

	status, listed := hostedCall(t, h.base, http.MethodGet, "/api/todos?dir="+h.workspace, h.creator, "org:member", nil)
	require.Equal(t, http.StatusOK, status, listed)
	diff, ok := listedDiff(listed["items"].([]any), todo)
	assert.True(t, ok, "the todo is listed: %v", listed)
	assert.Nil(t, diff, "without diff stats")

	status, batch := hostedCall(t, h.base, http.MethodPost, "/api/todos/batch", h.creator, "org:member", map[string]any{"dirs": []string{h.workspace}})
	require.Equal(t, http.StatusOK, status, batch)
	result := batch["results"].([]any)[0].(map[string]any)
	assert.Nil(t, result["error"], "the repository's todos aren't replaced by an error")
	_, ok = listedDiff(result["items"].([]any), todo)
	assert.True(t, ok, "the todo is listed: %v", result)

	status, detail := hostedCall(t, h.base, http.MethodGet, "/api/todos/item?dir="+h.workspace+"&ref="+todo.String(), h.creator, "org:member", nil)
	require.Equal(t, http.StatusOK, status, detail)
	assert.Nil(t, detail["diff"])
}
