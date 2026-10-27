package testui_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/flanksource/clicky"
	clickytask "github.com/flanksource/clicky/task"
	commonsContext "github.com/flanksource/commons/context"
	testui "github.com/flanksource/gavel/testrunner/ui"
)

// A `gavel lint prompt` run's review tasks live in the lint prompt group; the
// snapshot must surface them so the lint tab can show what is running.
func TestSnapshotSurfacesLivePromptLintTasks(t *testing.T) {
	clicky.ClearGlobalTasks()
	t.Cleanup(clicky.ClearGlobalTasks)
	const review = "messaging-review:report.mdx"

	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	group := clicky.StartGroup[string](testui.PromptLintTaskGroupName, clickytask.WithConcurrency(1))
	group.Add(review, func(_ commonsContext.Context, task *clickytask.Task) (string, error) {
		<-release
		task.Success()
		return "ok", nil
	})
	time.Sleep(50 * time.Millisecond) // let the task reach running state

	_, handler := newTestServer(t)
	resp := doRequest(t, handler, http.MethodGet, "/api/tests", nil)
	var got testui.Snapshot
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if len(got.Tests) != 1 || got.Tests[0].Name != testui.PromptLintTaskGroupName {
		t.Fatalf("tests = %+v, want one %q group", got.Tests, testui.PromptLintTaskGroupName)
	}
	if children := got.Tests[0].Children; len(children) != 1 || children[0].Name != review {
		t.Fatalf("group children = %+v, want the %q task", children, review)
	}
	if !got.Status.Running {
		t.Fatal("snapshot with a live prompt lint task must report running")
	}
}
