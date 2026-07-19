package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

type listFailureProvider struct {
	todos.Provider
	err error
}

func (p *listFailureProvider) List(context.Context, todos.DiscoveryFilters) (types.TODOS, error) {
	return nil, p.err
}

func TestRunTodosListHidesDoneByDefault(t *testing.T) {
	workDir := seedListTodos(t)
	stubRuntimeWithTestProvider(t)

	got := runListWithGlobals(t, workDir, TodosListOptions{})

	if len(got) != 2 {
		t.Fatalf("default list length = %d, want 2: %+v", len(got), got)
	}
	for _, todo := range got {
		if todo.Status == types.StatusCompleted || todo.Status == types.StatusVerified {
			t.Fatalf("default list included done todo: %+v", todo)
		}
	}
}

func TestRunTodosListDoneIncludesVerifiedAndCompleted(t *testing.T) {
	workDir := seedListTodos(t)
	stubRuntimeWithTestProvider(t)

	got := runListWithGlobals(t, workDir, TodosListOptions{Done: true})

	if len(got) != 4 {
		t.Fatalf("--done list length = %d, want 4: %+v", len(got), got)
	}
	if !hasStatus(got, types.StatusCompleted) {
		t.Fatalf("--done list did not include completed todo: %+v", got)
	}
	if !hasStatus(got, types.StatusVerified) {
		t.Fatalf("--done list did not include verified todo: %+v", got)
	}
}

func TestRunTodosListStatusCompletedOverridesDefaultHide(t *testing.T) {
	workDir := seedListTodos(t)
	stubRuntimeWithTestProvider(t)

	got := runListWithGlobals(t, workDir, TodosListOptions{Status: string(types.StatusCompleted)})

	if len(got) != 1 {
		t.Fatalf("completed list length = %d, want 1: %+v", len(got), got)
	}
	if got[0].Status != types.StatusCompleted {
		t.Fatalf("expected completed todo, got %+v", got[0])
	}
}

func TestRunTodosListStatusVerifiedOverridesDefaultHide(t *testing.T) {
	workDir := seedListTodos(t)
	stubRuntimeWithTestProvider(t)

	got := runListWithGlobals(t, workDir, TodosListOptions{Status: string(types.StatusVerified)})

	if len(got) != 1 || got[0].Status != types.StatusVerified {
		t.Fatalf("verified list = %+v, want one verified todo", got)
	}
}

func TestRunTodosListAllAggregatesRegisteredProjects(t *testing.T) {
	stubRuntimeWithTestProvider(t)
	firstDir := seedProjectTodos(t, "First pending", "First done")
	secondDir := seedProjectTodos(t, "Second pending", "Second done")
	oldLoad := loadTodoProjects
	loadTodoProjects = func() []ui.Project {
		return []ui.Project{
			{Name: "first", Dir: firstDir},
			{Name: "second", Dir: secondDir},
			{Name: "duplicate", Dir: firstDir},
		}
	}
	t.Cleanup(func() { loadTodoProjects = oldLoad })

	got := runListWithGlobals(t, firstDir, TodosListOptions{All: true})
	if len(got) != 2 {
		t.Fatalf("--all list length = %d, want 2: %+v", len(got), got)
	}
	if got[0].Title != "First pending" || got[1].Title != "Second pending" {
		t.Fatalf("--all titles = %q, %q; want both projects", got[0].Title, got[1].Title)
	}
	for _, todo := range got {
		if todo.CWD == "" {
			t.Fatalf("aggregated todo missing cwd: %+v", todo)
		}
		if todo.Workspace == "" {
			t.Fatalf("aggregated todo missing workspace: %+v", todo)
		}
	}

	got = runListWithGlobals(t, firstDir, TodosListOptions{All: true, Done: true})
	if len(got) != 4 {
		t.Fatalf("--all --done list length = %d, want 4: %+v", len(got), got)
	}

	got = runListWithGlobals(t, firstDir, TodosListOptions{All: true, Status: string(types.StatusCompleted)})
	if len(got) != 2 || got[0].Status != types.StatusCompleted || got[1].Status != types.StatusCompleted {
		t.Fatalf("--all --status completed = %+v, want two completed todos", got)
	}
}

func TestRunTodosListAllReturnsEveryDatabaseOpenAndListFailure(t *testing.T) {
	openFailure := fmt.Errorf("wrapped open failure: %w", database.ErrUnavailable)
	listFailure := errors.New("database list failed")
	openDir := filepath.Join(t.TempDir(), "open-failure")
	listDir := filepath.Join(t.TempDir(), "list-failure")

	oldOpen := openRuntimeTodosProvider
	oldLoad := loadTodoProjects
	oldWorkingDir := workingDir
	openRuntimeTodosProvider = func(_ context.Context, workDir string) (todos.Provider, error) {
		switch workDir {
		case openDir:
			return nil, openFailure
		case listDir:
			return &listFailureProvider{err: listFailure}, nil
		default:
			t.Fatalf("unexpected workspace %q", workDir)
			return nil, nil
		}
	}
	loadTodoProjects = func() []ui.Project {
		return []ui.Project{
			{Name: "open project", Dir: openDir},
			{Name: "list project", Dir: listDir},
		}
	}
	workingDir = t.TempDir()
	t.Cleanup(func() {
		openRuntimeTodosProvider = oldOpen
		loadTodoProjects = oldLoad
		workingDir = oldWorkingDir
	})

	result, err := runTodosList(TodosListOptions{All: true})
	if result != nil {
		t.Fatalf("runTodosList result = %#v, want nil on aggregate failure", result)
	}
	if !errors.Is(err, database.ErrUnavailable) {
		t.Fatalf("runTodosList error = %v, want ErrUnavailable", err)
	}
	if !errors.Is(err, listFailure) {
		t.Fatalf("runTodosList error = %v, want list failure", err)
	}
	for _, want := range []string{"open project", "list project", "open native TODO workspace", "list native TODOs"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("runTodosList error = %q, want %q", err, want)
		}
	}
}

func TestFilterTODOsSinceUsesNewestCreatedOrUpdatedTimestamp(t *testing.T) {
	cutoff := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	old := cutoff.Add(-24 * time.Hour)
	newer := cutoff.Add(time.Hour)
	createdAtCutoff := cutoff

	got := filterTODOsSince(types.TODOS{
		{ID: "created", TODOFrontmatter: types.TODOFrontmatter{Created: &newer}},
		{ID: "updated", TODOFrontmatter: types.TODOFrontmatter{Created: &old, LastRun: &newer}},
		{ID: "boundary", TODOFrontmatter: types.TODOFrontmatter{Created: &createdAtCutoff}},
		{ID: "old", TODOFrontmatter: types.TODOFrontmatter{Created: &old, LastRun: &old}},
		{ID: "unknown"},
	}, cutoff)

	if len(got) != 3 || got[0].ID != "created" || got[1].ID != "updated" || got[2].ID != "boundary" {
		t.Fatalf("filterTODOsSince = %+v, want created, updated, boundary", got)
	}
}

func TestRunTodosListRejectsInvalidSince(t *testing.T) {
	workDir := seedListTodos(t)
	oldWorkingDir := workingDir
	workingDir = workDir
	t.Cleanup(func() {
		workingDir = oldWorkingDir
	})

	if _, err := runTodosList(TodosListOptions{Since: "not-a-date"}); err == nil || !strings.Contains(err.Error(), "unable to parse since") {
		t.Fatalf("runTodosList error = %v, want invalid --since error", err)
	}
}

func TestTodoListSinceAcceptsDateMath(t *testing.T) {
	before := time.Now().Add(-7*24*time.Hour - time.Second)
	got, err := parseSince("7d")
	if err != nil {
		t.Fatalf("parseSince(7d): %v", err)
	}
	after := time.Now().Add(-7*24*time.Hour + time.Second)
	if got.Before(before) || got.After(after) {
		t.Fatalf("parseSince(7d) = %s, want approximately seven days ago", got)
	}
}

func seedListTodos(t *testing.T) string {
	t.Helper()
	workDir := t.TempDir()
	provider := testProviderFor(workDir)
	for _, item := range []struct {
		title  string
		status types.Status
	}{
		{"Pending item", types.StatusPending},
		{"Running item", types.StatusInProgress},
		{"Verified item", types.StatusVerified},
		{"Completed item", types.StatusCompleted},
	} {
		if _, err := provider.Create(t.Context(), todos.CreateRequest{
			Title:  item.title,
			Status: item.status,
		}); err != nil {
			t.Fatalf("seed %q: %v", item.title, err)
		}
	}
	return workDir
}

func seedProjectTodos(t *testing.T, pendingTitle, completedTitle string) string {
	t.Helper()
	workDir := t.TempDir()
	provider := testProviderFor(workDir)
	for _, item := range []struct {
		title  string
		status types.Status
	}{
		{pendingTitle, types.StatusPending},
		{completedTitle, types.StatusCompleted},
	} {
		if _, err := provider.Create(t.Context(), todos.CreateRequest{Title: item.title, Status: item.status}); err != nil {
			t.Fatalf("seed %q: %v", item.title, err)
		}
	}
	return workDir
}

func runListWithGlobals(t *testing.T, workDir string, opts TodosListOptions) types.TODOS {
	t.Helper()
	oldWorkingDir := workingDir
	workingDir = workDir
	t.Cleanup(func() {
		workingDir = oldWorkingDir
	})

	out, err := runTodosList(opts)
	if err != nil {
		t.Fatalf("runTodosList: %v", err)
	}
	got, ok := out.(types.TODOS)
	if !ok {
		t.Fatalf("runTodosList returned %T, want types.TODOS", out)
	}
	return got
}

func hasStatus(todoList types.TODOS, status types.Status) bool {
	for _, todo := range todoList {
		if todo.Status == status {
			return true
		}
	}
	return false
}
