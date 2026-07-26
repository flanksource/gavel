package todosync

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

func TestScanSourceCommentsFindsMarkersAndIgnoresStrings(t *testing.T) {
	workDir := t.TempDir()
	path := filepath.Join(workDir, "pkg", "app.go")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	source := `package pkg

const ignored = "// TODO: not a comment"

func run() {
	// TODO(alice): wire CLI sync
	_ = 1 /* FIXME: handle failures */
}
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	comments, scanned, err := ScanSourceComments(SourceCommentSyncOptions{WorkDir: workDir})
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if scanned != 1 {
		t.Fatalf("scanned files = %d, want 1", scanned)
	}
	if len(comments) != 2 {
		t.Fatalf("comments = %+v, want 2", comments)
	}
	if comments[0].Marker != "TODO" || comments[0].Message != "(alice) wire CLI sync" {
		t.Fatalf("unexpected TODO comment: %+v", comments[0])
	}
	if comments[1].Marker != "FIXME" || comments[1].Message != "handle failures" {
		t.Fatalf("unexpected FIXME comment: %+v", comments[1])
	}
}

func TestScanSourceCommentIDSurvivesLineMove(t *testing.T) {
	workDir := t.TempDir()
	path := filepath.Join(workDir, "main.go")
	if err := os.WriteFile(path, []byte("// TODO: keep stable\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	first, _, err := ScanSourceComments(SourceCommentSyncOptions{WorkDir: workDir})
	if err != nil {
		t.Fatalf("first scan: %v", err)
	}
	if err := os.WriteFile(path, []byte("\n\n// TODO: keep stable\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	second, _, err := ScanSourceComments(SourceCommentSyncOptions{WorkDir: workDir})
	if err != nil {
		t.Fatalf("second scan: %v", err)
	}
	if len(first) != 1 || len(second) != 1 {
		t.Fatalf("expected one comment in each scan: first=%+v second=%+v", first, second)
	}
	if first[0].ID != second[0].ID {
		t.Fatalf("source id changed across line move: %s != %s", first[0].ID, second[0].ID)
	}
	if second[0].Line != 3 {
		t.Fatalf("second line = %d, want 3", second[0].Line)
	}
}

func TestSyncSourceCommentsNativeProviderLifecycle(t *testing.T) {
	ctx := context.Background()
	workDir := t.TempDir()
	sourcePath := filepath.Join(workDir, "main.go")
	provider := &sourceSyncProvider{}

	if err := os.WriteFile(sourcePath, []byte("// TODO: move sync to gavel\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	created, err := SyncSourceComments(ctx, provider, SourceCommentSyncOptions{WorkDir: workDir})
	if err != nil {
		t.Fatalf("initial sync: %v", err)
	}
	if len(created.Created) != 1 || created.Matches != 1 {
		t.Fatalf("initial result = %+v, want one created match", created)
	}
	todo := onlyGeneratedTodo(t, provider)
	if todo.Status != types.StatusPending {
		t.Fatalf("status = %s, want pending", todo.Status)
	}
	if len(todo.Path) != 1 || todo.Path[0] != "main.go:1" {
		t.Fatalf("path = %#v, want main.go:1", todo.Path)
	}
	if todo.Metadata["source"] != sourceTodoKind {
		t.Fatalf("metadata source = %#v", todo.Metadata["source"])
	}

	if err := os.WriteFile(sourcePath, []byte("\n// TODO: move sync to gavel\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	updated, err := SyncSourceComments(ctx, provider, SourceCommentSyncOptions{WorkDir: workDir})
	if err != nil {
		t.Fatalf("line move sync: %v", err)
	}
	if len(updated.Updated) != 1 || len(updated.Created) != 0 {
		t.Fatalf("line move result = %+v, want one update", updated)
	}
	todo = onlyGeneratedTodo(t, provider)
	if len(todo.Path) != 1 || todo.Path[0] != "main.go:2" {
		t.Fatalf("moved path = %#v, want main.go:2", todo.Path)
	}

	if err := os.WriteFile(sourcePath, []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	completed, err := SyncSourceComments(ctx, provider, SourceCommentSyncOptions{WorkDir: workDir})
	if err != nil {
		t.Fatalf("completion sync: %v", err)
	}
	if len(completed.Completed) != 1 {
		t.Fatalf("completion result = %+v, want one completed", completed)
	}
	todo = onlyGeneratedTodo(t, provider)
	if todo.Status != types.StatusCompleted {
		t.Fatalf("status after completion = %s, want completed", todo.Status)
	}

	if err := os.WriteFile(sourcePath, []byte("// TODO: move sync to gavel\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	reopened, err := SyncSourceComments(ctx, provider, SourceCommentSyncOptions{WorkDir: workDir})
	if err != nil {
		t.Fatalf("reopen sync: %v", err)
	}
	if len(reopened.Updated) != 1 || len(reopened.Created) != 0 {
		t.Fatalf("reopen result = %+v, want one update", reopened)
	}
	todo = onlyGeneratedTodo(t, provider)
	if todo.Status != types.StatusPending {
		t.Fatalf("status after reopen = %s, want pending", todo.Status)
	}
}

func onlyGeneratedTodo(t *testing.T, provider *sourceSyncProvider) *types.TODO {
	t.Helper()
	items, err := provider.List(context.Background(), todos.DiscoveryFilters{IncludeLabels: []string{SourceTodoLabel}})
	if err != nil {
		t.Fatalf("list generated todos: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("generated todos = %+v, want one", items)
	}
	detail, err := provider.Get(context.Background(), items[0].ID)
	if err != nil {
		t.Fatalf("get generated todo: %v", err)
	}
	return detail
}

// sourceSyncProvider is an in-memory native-provider test double. The source
// sync tests exercise reconciliation without reintroducing `.todos` storage.
type sourceSyncProvider struct{ todo *types.TODO }

func (p *sourceSyncProvider) List(_ context.Context, filters todos.DiscoveryFilters) (types.TODOS, error) {
	if p.todo == nil || !filters.Matches(p.todo) {
		return nil, nil
	}
	return types.TODOS{p.todo}, nil
}

func (p *sourceSyncProvider) CountByStatus(_ context.Context) (map[types.Status]int, error) {
	counts := map[types.Status]int{}
	if p.todo != nil {
		counts[p.todo.Status]++
	}
	return counts, nil
}

func (p *sourceSyncProvider) Get(_ context.Context, _ string) (*types.TODO, error) {
	return p.todo, nil
}

func (p *sourceSyncProvider) Create(_ context.Context, req todos.CreateRequest) (*types.TODO, error) {
	p.todo = &types.TODO{
		ID: "source-sync-issue", Labels: append([]string(nil), req.Labels...), MarkdownBody: req.Body,
		TODOFrontmatter: types.TODOFrontmatter{
			Title: req.Title, Priority: req.Priority, Status: req.Status, Path: req.Path,
			FrontMatter: fixturesFrontMatter(req.Metadata),
		},
	}
	return p.todo, nil
}

func fixturesFrontMatter(metadata map[string]any) fixtures.FrontMatter {
	return fixtures.FrontMatter{Metadata: metadata}
}

func (p *sourceSyncProvider) Delete(context.Context, *types.TODO) error { return nil }

func (p *sourceSyncProvider) Edit(_ context.Context, todo *types.TODO, edit todos.EditRequest) error {
	if edit.Title != nil {
		todo.Title = *edit.Title
	}
	if edit.Body != nil {
		todo.MarkdownBody = *edit.Body
	}
	if edit.Path != nil {
		todo.Path = *edit.Path
	}
	if len(edit.Labels) > 0 {
		todo.Labels = append([]string(nil), edit.Labels...)
	}
	if todo.Metadata == nil {
		todo.Metadata = map[string]any{}
	}
	for key, value := range edit.Metadata {
		todo.Metadata[key] = value
	}
	return nil
}

func (p *sourceSyncProvider) Comment(context.Context, *types.TODO, string) error { return nil }

func (p *sourceSyncProvider) UpdateState(_ context.Context, todo *types.TODO, update todos.StateUpdate) error {
	if update.Status != nil {
		todo.Status = *update.Status
	}
	return nil
}

func (p *sourceSyncProvider) UpdateLatestFailure(context.Context, *types.TODO, *types.TestResultInfo) error {
	return nil
}

func (p *sourceSyncProvider) SaveAttempt(context.Context, *types.TODO, *todos.ExecutionResult) error {
	return nil
}
