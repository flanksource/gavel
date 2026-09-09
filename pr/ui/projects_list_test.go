package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withProjects points projectsPath at a temp file holding the named projects,
// in order. Unlike withProject it does not create Procfiles — these tests are
// about how the list endpoint fans out, not what is in each directory.
func withProjects(t *testing.T, names ...string) {
	t.Helper()
	original := projectsPath
	projectsPath = filepath.Join(t.TempDir(), "projects.json")
	t.Cleanup(func() { projectsPath = original })
	projects := make([]Project, 0, len(names))
	for _, name := range names {
		projects = append(projects, Project{Name: name, Dir: t.TempDir(), Repos: []string{"example/" + name}})
	}
	require.NoError(t, SaveProjects(projects))
}

func getProjects(t *testing.T) (*httptest.ResponseRecorder, []projectInfo) {
	t.Helper()
	rec := httptest.NewRecorder()
	(&Server{}).handleProjects(rec, httptest.NewRequest(http.MethodGet, "/api/projects", nil))
	var got []projectInfo
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got), "body = %q", rec.Body.String())
	return rec, got
}

// TestHandleProjectsCountsConcurrently pins the fan-out: with as many projects
// as the concurrency limit allows, every count lookup must be in flight at once.
// A serial loop never reaches the barrier and the test fails on the deadline
// rather than hanging.
func TestHandleProjectsCountsConcurrently(t *testing.T) {
	names := make([]string, projectInfoConcurrency)
	for i := range names {
		names[i] = fmt.Sprintf("project-%d", i)
	}
	withProjects(t, names...)

	entered := make(chan struct{}, projectInfoConcurrency)
	release := make(chan struct{})
	original := projectTodoCounts
	projectTodoCounts = func(context.Context, Project) (todoCounts, error) {
		entered <- struct{}{}
		<-release
		return todoCounts{}, nil
	}
	t.Cleanup(func() { projectTodoCounts = original })

	done := make(chan struct{})
	go func() {
		defer close(done)
		rec := httptest.NewRecorder()
		(&Server{}).handleProjects(rec, httptest.NewRequest(http.MethodGet, "/api/projects", nil))
	}()

	deadline := time.After(10 * time.Second)
	for i := 0; i < projectInfoConcurrency; i++ {
		select {
		case <-entered:
		case <-deadline:
			t.Fatalf("only %d of %d count lookups started concurrently", i, projectInfoConcurrency)
		}
	}
	close(release)
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("handleProjects did not return after releasing the count lookups")
	}
}

// TestHandleProjectsPreservesConfiguredOrder guards the indexed writes: parallel
// counting must not reorder projects.json.
func TestHandleProjectsPreservesConfiguredOrder(t *testing.T) {
	want := []string{"alpha", "bravo", "charlie", "delta", "echo", "foxtrot"}
	withProjects(t, want...)

	original := projectTodoCounts
	// Reverse-ordered delays: the last project finishes first, so an
	// append-as-you-finish implementation would emit the reverse order.
	delays := map[string]time.Duration{}
	for i, name := range want {
		delays[name] = time.Duration(len(want)-i) * 5 * time.Millisecond
	}
	projectTodoCounts = func(_ context.Context, project Project) (todoCounts, error) {
		time.Sleep(delays[project.Name])
		return todoCounts{Total: 1, Open: 1}, nil
	}
	t.Cleanup(func() { projectTodoCounts = original })

	rec, got := getProjects(t)
	assert.Equal(t, http.StatusOK, rec.Code)
	names := make([]string, 0, len(got))
	for _, info := range got {
		names = append(names, info.Name)
	}
	assert.Equal(t, want, names, "projects must be returned in projects.json order")
}

// TestHandleProjectsIsolatesPerProjectCountFailures is the regression that
// matters for the dashboard: the projects payload also drives Processes and PRs,
// so one workspace whose TODO store is unreachable must not 500 the whole list.
func TestHandleProjectsIsolatesPerProjectCountFailures(t *testing.T) {
	withProjects(t, "healthy", "broken", "also-healthy")

	original := projectTodoCounts
	projectTodoCounts = func(_ context.Context, project Project) (todoCounts, error) {
		if project.Name == "broken" {
			return todoCounts{}, fmt.Errorf("connect to postgres: connection refused")
		}
		return todoCounts{Total: 5, Open: 3}, nil
	}
	t.Cleanup(func() { projectTodoCounts = original })

	rec, got := getProjects(t)
	assert.Equal(t, http.StatusOK, rec.Code, "one failing project must not fail the list")
	require.Len(t, got, 3)

	for _, name := range []string{"healthy", "also-healthy"} {
		info := projectNamed(t, got, name)
		require.NotNil(t, info.TodoCounts, "%s must still report counts", name)
		assert.Equal(t, todoCounts{Total: 5, Open: 3}, *info.TodoCounts)
		assert.Empty(t, info.Error)
	}

	broken := projectNamed(t, got, "broken")
	assert.Nil(t, broken.TodoCounts, "failed counts must be absent, not a misleading zero")
	assert.Contains(t, broken.Error, "connection refused", "the failure must stay visible in the payload")
}

// TestHandleProjectByNameStillFailsOnCountError keeps the single-entity endpoint
// strict: there is no other project to serve, so the error is the response.
func TestHandleProjectByNameStillFailsOnCountError(t *testing.T) {
	withProjects(t, "broken")

	original := projectTodoCounts
	projectTodoCounts = func(context.Context, Project) (todoCounts, error) {
		return todoCounts{}, fmt.Errorf("connect to postgres: connection refused")
	}
	t.Cleanup(func() { projectTodoCounts = original })

	rec := httptest.NewRecorder()
	(&Server{}).handleProjectByName(rec, projectByNameReq("GET", "broken", ""))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "connection refused")
}

func projectNamed(t *testing.T, infos []projectInfo, name string) projectInfo {
	t.Helper()
	for _, info := range infos {
		if info.Name == name {
			return info
		}
	}
	t.Fatalf("project %q missing from response %+v", name, infos)
	return projectInfo{}
}
