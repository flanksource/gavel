package testui_test

import (
	"encoding/json"
	"net/http"
	"testing"

	clickytask "github.com/flanksource/clicky/task"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

func TestFrozenRunningTaskDoesNotKeepAFinishedRunRunning(t *testing.T) {
	srv, handler := newTestServer(t)
	srv.BeginRun("tasks")
	srv.SetTaskSource(func() []clickytask.TaskSnapshot {
		return []clickytask.TaskSnapshot{
			{ID: "prepare", GroupID: "prepare", Type: "group", Name: "Prepare", Status: "running"},
			{ID: "build", GroupID: "prepare", Name: "Build", Status: "running"},
		}
	})
	srv.MarkDone()

	var snap testui.Snapshot
	resp := doRequest(t, handler, http.MethodGet, "/api/tests", nil)
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if snap.Status.Running {
		t.Fatalf("a finished run reported running because a task was running when it was frozen")
	}
}
