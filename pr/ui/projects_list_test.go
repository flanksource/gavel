package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withProjects points projectsPath at a temp file holding the named projects,
// in order. Unlike withProject it does not create Procfiles — these tests are
// about how the list endpoint reads counts, not what is in each directory.
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

// perProjectTodoCounts adapts a per-project stub to the batched count seam, so
// a test states what each project's counts are and the seam keeps its one
// result per project, in order.
func perProjectTodoCounts(fn func(Project) (todoCounts, error)) func(context.Context, []Project) []todoCountsResult {
	return func(_ context.Context, projects []Project) []todoCountsResult {
		results := make([]todoCountsResult, len(projects))
		for i, project := range projects {
			results[i].Counts, results[i].Err = fn(project)
		}
		return results
	}
}

// stubTodoCounts swaps the batched count seam for the duration of a test.
func stubTodoCounts(t *testing.T, stub func(context.Context, []Project) []todoCountsResult) {
	t.Helper()
	original := projectsTodoCounts
	projectsTodoCounts = stub
	t.Cleanup(func() { projectsTodoCounts = original })
}

func getProjects(t *testing.T) (*httptest.ResponseRecorder, []projectInfo) {
	t.Helper()
	rec := httptest.NewRecorder()
	(&Server{}).handleProjects(rec, httptest.NewRequest(http.MethodGet, "/api/projects", nil))
	var got []projectInfo
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got), "body = %q", rec.Body.String())
	return rec, got
}

// TestHandleProjectsCountsInOneBatch pins the batching: however many projects
// are configured, the list reads their counts through one call to the seam,
// handed every project in the store's order — the call CountProjects turns into
// a fixed number of queries.
func TestHandleProjectsCountsInOneBatch(t *testing.T) {
	withProjects(t, "echo", "Alpha", "foxtrot", "charlie", "Delta", "bravo")
	var (
		mu      sync.Mutex
		batches [][]string
	)
	stubTodoCounts(t, func(ctx context.Context, projects []Project) []todoCountsResult {
		names := make([]string, 0, len(projects))
		for _, project := range projects {
			names = append(names, project.Name)
		}
		mu.Lock()
		batches = append(batches, names)
		mu.Unlock()
		return make([]todoCountsResult, len(projects))
	})

	rec, got := getProjects(t)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Len(t, got, 6)
	assert.Equal(t, [][]string{{"Alpha", "bravo", "charlie", "Delta", "echo", "foxtrot"}}, batches)
}

// TestHandleProjectsReturnsStoreOrder guards both the store's case-insensitive
// sort and the indexed hand-off: every project must carry its own counts, not
// a neighbour's.
func TestHandleProjectsReturnsStoreOrder(t *testing.T) {
	withProjects(t, "echo", "Alpha", "foxtrot", "charlie", "Delta", "bravo")
	want := []string{"Alpha", "bravo", "charlie", "Delta", "echo", "foxtrot"}
	totals := map[string]int{}
	for i, name := range want {
		totals[name] = (i + 1) * 10
	}
	stubTodoCounts(t, perProjectTodoCounts(func(project Project) (todoCounts, error) {
		return todoCounts{Total: totals[project.Name], Open: totals[project.Name]}, nil
	}))

	rec, got := getProjects(t)
	assert.Equal(t, http.StatusOK, rec.Code)
	names := make([]string, 0, len(got))
	for _, info := range got {
		names = append(names, info.Name)
		require.NotNil(t, info.TodoCounts, info.Name)
		assert.Equal(t, totals[info.Name], info.TodoCounts.Total, "%s must carry its own counts", info.Name)
	}
	assert.Equal(t, want, names, "projects must be returned in the store's sorted order")
}

// TestHandleProjectsIsolatesPerProjectCountFailures is the regression that
// matters for the dashboard: the projects payload also drives Processes and PRs,
// so one workspace whose TODO store is unreachable must not 500 the whole list.
func TestHandleProjectsIsolatesPerProjectCountFailures(t *testing.T) {
	withProjects(t, "healthy", "broken", "also-healthy")
	stubTodoCounts(t, perProjectTodoCounts(func(project Project) (todoCounts, error) {
		if project.Name == "broken" {
			return todoCounts{}, fmt.Errorf("connect to postgres: connection refused")
		}
		return todoCounts{Total: 5, Open: 3}, nil
	}))

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

// TestHandleProjectsReportsAMisalignedCountBatch keeps the hand-off loud: a
// count batch that does not answer every project cannot be matched to the
// projects it belongs to, so every entry reports it rather than showing counts
// that may be someone else's.
func TestHandleProjectsReportsAMisalignedCountBatch(t *testing.T) {
	withProjects(t, "first", "second")
	stubTodoCounts(t, func(context.Context, []Project) []todoCountsResult {
		return []todoCountsResult{{Counts: todoCounts{Total: 1}}}
	})

	rec, got := getProjects(t)

	assert.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, got, 2)
	for _, info := range got {
		assert.Nil(t, info.TodoCounts, info.Name)
		assert.Contains(t, info.Error, "1 results for 2 projects", info.Name)
	}
}

// TestHandleProjectByNameStillFailsOnCountError keeps the single-entity endpoint
// strict: there is no other project to serve, so the error is the response.
func TestHandleProjectByNameStillFailsOnCountError(t *testing.T) {
	withProjects(t, "broken")
	stubTodoCounts(t, perProjectTodoCounts(func(Project) (todoCounts, error) {
		return todoCounts{}, fmt.Errorf("connect to postgres: connection refused")
	}))

	rec := httptest.NewRecorder()
	(&Server{}).handleProjectByName(rec, projectByNameReq("GET", "broken", ""))
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "connection refused")
}

// TestCountProjectsTodosSkipsProjectsWithoutADirectory pins the one case the
// batch answers without the database: a project with no directory has no
// workspace, and reports zero counts rather than an error.
func TestCountProjectsTodosSkipsProjectsWithoutADirectory(t *testing.T) {
	got := countProjectsTodos(t.Context(), []Project{{Name: "no-dir"}, {Name: "blank-dir", Dir: "  "}})

	assert.Equal(t, []todoCountsResult{{}, {}}, got)
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
