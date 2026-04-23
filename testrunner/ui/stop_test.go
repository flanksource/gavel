package testui_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/flanksource/clicky"
	clickytask "github.com/flanksource/clicky/task"
	commonsContext "github.com/flanksource/commons/context"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

func TestStopRequiresPOST(t *testing.T) {
	srv, handler := newTestServer(t)
	srv.SetStopFunc(func() {})

	resp := doRequest(t, handler, http.MethodGet, "/api/stop", nil)
	if resp.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", resp.Code)
	}
}

func TestStopWithoutHandlerReturns501(t *testing.T) {
	_, handler := newTestServer(t)

	resp := doRequest(t, handler, http.MethodPost, "/api/stop", strings.NewReader(`{"scope":"global"}`))
	if resp.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want 501", resp.Code)
	}
}

func TestGlobalStopMarksSnapshotStopped(t *testing.T) {
	srv, handler := newTestServer(t)
	srv.BeginRun("initial")
	var stopCalled atomic.Bool
	srv.SetStopFunc(func() {
		stopCalled.Store(true)
		srv.MarkDone()
	})

	resp := doRequest(t, handler, http.MethodPost, "/api/stop", strings.NewReader(`{"scope":"global"}`))
	if resp.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202: %s", resp.Code, resp.Body.String())
	}
	if !stopCalled.Load() {
		t.Fatalf("expected stop callback to be invoked")
	}

	var snap testui.Snapshot
	resp = doRequest(t, handler, http.MethodGet, "/api/tests", nil)
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if snap.Status.Running {
		t.Fatalf("snapshot should be stopped")
	}
	if !snap.Status.Stopped {
		t.Fatalf("snapshot should be marked stopped")
	}
	if snap.Status.StopMessage != "Stopped by user" {
		t.Fatalf("stop message = %q, want %q", snap.Status.StopMessage, "Stopped by user")
	}
}

func TestTaskStopUsesTaskID(t *testing.T) {
	clicky.ClearGlobalTasks()
	t.Cleanup(func() {
		clicky.CancelAllGlobalTasks()
		clicky.ClearGlobalTasks()
	})

	srv, handler := newTestServer(t)
	srv.SetStopFunc(func() {})
	srv.BeginRun("initial")
	srv.MarkDone()

	group := clicky.StartGroup[string](testui.TestTaskGroupName, clickytask.WithConcurrency(1))
	task := group.Add("dummy", func(ctx commonsContext.Context, t *clickytask.Task) (string, error) {
		t.SetName("go test -json ./pkg/foo")
		<-ctx.Done()
		return "", ctx.Err()
	})

	time.Sleep(50 * time.Millisecond)

	snapResp := doRequest(t, handler, http.MethodGet, "/api/tests", nil)
	var snap testui.Snapshot
	if err := json.NewDecoder(snapResp.Body).Decode(&snap); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(snap.Tests) == 0 || len(snap.Tests[0].Children) == 0 {
		t.Fatalf("expected virtual task snapshot, got %+v", snap.Tests)
	}
	taskID := snap.Tests[0].Children[0].TaskID
	if taskID == "" {
		t.Fatalf("expected task_id in snapshot")
	}

	body, _ := json.Marshal(testui.StopRequest{Scope: "task", TaskID: taskID})
	resp := doRequest(t, handler, http.MethodPost, "/api/stop", bytes.NewReader(body))
	if resp.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202: %s", resp.Code, resp.Body.String())
	}

	wait := task.WaitFor()
	if wait.Status != clickytask.StatusCancelled {
		t.Fatalf("task status = %q, want canceled", wait.Status)
	}
}
