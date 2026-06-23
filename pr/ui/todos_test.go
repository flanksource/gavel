package ui

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

func TestTodoAPIFileProviderCRUD(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	createBody := `{"title":"Fix workspace","body":"Implement todo tab","priority":"high","status":"in_progress"}`
	rec := httptest.NewRecorder()
	s.handleTodos(rec, httptest.NewRequest(http.MethodPost, "/api/todos?provider=todos", strings.NewReader(createBody)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %q", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var created todoSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create: %v", err)
	}
	if created.Title != "Fix workspace" || created.Status != types.StatusInProgress || created.Priority != types.PriorityHigh {
		t.Fatalf("unexpected created todo: %+v", created)
	}
	if _, err := os.Stat(filepath.Join(workDir, ".todos", "fix-workspace.md")); err != nil {
		t.Fatalf("expected TODO file to be created: %v", err)
	}

	rec = httptest.NewRecorder()
	s.handleTodos(rec, httptest.NewRequest(http.MethodGet, "/api/todos?provider=todos", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var list todoListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if list.Counts.Total != 1 || list.Counts.Open != 1 || list.Counts.InProgress != 1 {
		t.Fatalf("unexpected counts: %+v", list.Counts)
	}

	rec = httptest.NewRecorder()
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodGet, "/api/todos/item?provider=todos&ref="+url.QueryEscape(created.Ref), nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var detail todoSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal detail: %v", err)
	}
	if !strings.Contains(detail.Body, "Implement todo tab") {
		t.Fatalf("detail body missing content: %+v", detail)
	}

	rec = httptest.NewRecorder()
	patchBody := `{"ref":` + strconvQuote(created.Ref) + `,"status":"completed"}`
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodPatch, "/api/todos/item?provider=todos", strings.NewReader(patchBody)))
	if rec.Code != http.StatusOK {
		t.Fatalf("patch status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var patched todoSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &patched); err != nil {
		t.Fatalf("unmarshal patch: %v", err)
	}
	if patched.Status != types.StatusCompleted {
		t.Fatalf("status = %q, want completed", patched.Status)
	}

	rec = httptest.NewRecorder()
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodDelete, "/api/todos/item?provider=todos&ref="+url.QueryEscape(created.Ref), nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(created.Ref); !os.IsNotExist(err) {
		t.Fatalf("expected TODO file to be removed, stat err=%v", err)
	}
}

func TestTodoNewEndpointQueryDefaultsDraft(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/todos/new?provider=todos&dir="+url.QueryEscape(workDir)+"&title=Draft+from+query&priority=low", nil)
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("new status = %d, want %d; body = %q", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var resp todoNewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal new response: %v", err)
	}
	if resp.AutoSave {
		t.Fatalf("autoSave default = true, want false")
	}
	if resp.Todo.Title != "Draft from query" || resp.Todo.Status != types.StatusDraft || resp.Todo.Priority != types.PriorityLow {
		t.Fatalf("unexpected created draft: %+v", resp)
	}
}

func TestTodoNewEndpointJSONAutoSaveDefaultsPending(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	body := `{"title":"JSON todo","body":"Created from json","autoSave":true}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/todos/new?provider=todos&dir="+url.QueryEscape(workDir), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("new json status = %d, want %d; body = %q", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var resp todoNewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal new json response: %v", err)
	}
	if !resp.AutoSave {
		t.Fatalf("autoSave = false, want true")
	}
	if resp.Todo.Status != types.StatusPending {
		t.Fatalf("status = %q, want pending", resp.Todo.Status)
	}
	if !strings.Contains(resp.Todo.Body, "Created from json") {
		t.Fatalf("created body missing json content: %+v", resp.Todo)
	}
}

func TestTodoNewEndpointMultipartFiles(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range map[string]string{
		"title":    "Screenshot todo",
		"body":     "Screenshot context.",
		"status":   string(types.StatusVerified),
		"priority": string(types.PriorityHigh),
		"autoSave": "true",
	} {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write field %s: %v", key, err)
		}
	}
	part, err := writer.CreateFormFile("screenshot", "screen.png")
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := part.Write([]byte("png bytes")); err != nil {
		t.Fatalf("write file part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/todos/new?provider=todos&dir="+url.QueryEscape(workDir), &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("new multipart status = %d, want %d; body = %q", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var resp todoNewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal new multipart response: %v", err)
	}
	if resp.Todo.Status != types.StatusVerified || resp.Todo.Priority != types.PriorityHigh {
		t.Fatalf("unexpected multipart todo: %+v", resp.Todo)
	}
	if len(resp.Attachments) != 1 || resp.Attachments[0].Filename != "screen.png" || resp.Attachments[0].Field != "screenshot" {
		t.Fatalf("unexpected attachments: %+v", resp.Attachments)
	}
	if !strings.Contains(resp.Todo.Body, "## Attachments") || !strings.Contains(resp.Todo.Body, "screen.png") {
		t.Fatalf("created body missing attachment summary: %q", resp.Todo.Body)
	}
}

func TestTodoAPIPatchPriority(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	created, err := todos.NewFileProvider(workDir, "").Create(t.Context(), todos.CreateRequest{
		Title:    "Tune severity",
		Priority: types.PriorityMedium,
		Status:   types.StatusPending,
	})
	if err != nil {
		t.Fatalf("create todo: %v", err)
	}
	ref := todos.TODOReference(created)

	// PATCH priority only (no status) sets severity and leaves status alone.
	rec := httptest.NewRecorder()
	body := `{"ref":` + strconvQuote(ref) + `,"priority":"low"}`
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodPatch, "/api/todos/item?provider=todos", strings.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("patch priority status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var patched todoSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &patched); err != nil {
		t.Fatalf("unmarshal patch: %v", err)
	}
	if patched.Priority != types.PriorityLow {
		t.Errorf("priority = %q, want low", patched.Priority)
	}
	if patched.Status != types.StatusPending {
		t.Errorf("status changed to %q, want pending preserved", patched.Status)
	}

	// PATCH with neither status nor priority is a 400.
	rec = httptest.NewRecorder()
	empty := `{"ref":` + strconvQuote(ref) + `}`
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodPatch, "/api/todos/item?provider=todos", strings.NewReader(empty)))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("empty patch status = %d, want 400", rec.Code)
	}

	// PATCH with an invalid priority is a 400.
	rec = httptest.NewRecorder()
	bad := `{"ref":` + strconvQuote(ref) + `,"priority":"urgent"}`
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodPatch, "/api/todos/item?provider=todos", strings.NewReader(bad)))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("invalid priority status = %d, want 400", rec.Code)
	}
}

func TestTodoAPIAutoProviderListsWorkspace(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	// Seed a .todos workspace (no .grite present), so the auto provider must
	// resolve the file provider for this directory.
	if _, err := todos.NewFileProvider(workDir, "").Create(t.Context(), todos.CreateRequest{
		Title:    "Auto detect me",
		Priority: types.PriorityHigh,
		Status:   types.StatusPending,
	}); err != nil {
		t.Fatalf("create todo: %v", err)
	}

	rec := httptest.NewRecorder()
	s.handleTodos(rec, httptest.NewRequest(http.MethodGet, "/api/todos?provider=auto&dir="+url.QueryEscape(workDir), nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var list todoListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if list.Counts.Total != 1 || len(list.Items) != 1 || list.Items[0].Title != "Auto detect me" {
		t.Fatalf("auto provider did not list the .todos workspace: %+v", list)
	}
}

func TestAutoTodoProviderSelection(t *testing.T) {
	// A directory with a .todos store resolves to the file provider.
	withTodos := t.TempDir()
	if err := os.MkdirAll(filepath.Join(withTodos, ".todos"), 0o755); err != nil {
		t.Fatalf("mkdir .todos: %v", err)
	}
	if got := autoTodoProvider(withTodos); !isFileProvider(got) {
		t.Errorf("autoTodoProvider(dir with .todos) = %T, want *todos.FileProvider", got)
	}

	// A directory without .todos resolves to Grite, which tracks issues globally
	// per repo and must NOT be gated on a .grite marker dir.
	plain := t.TempDir()
	if got := autoTodoProvider(plain); !isGriteProvider(got) {
		t.Errorf("autoTodoProvider(plain dir) = %T, want *todos.GriteProvider", got)
	}
}

func TestProviderForDirSelection(t *testing.T) {
	dir := t.TempDir()
	if got := providerForDir(dir, "grite"); !isGriteProvider(got) {
		t.Errorf("providerForDir(_, grite) = %T, want *todos.GriteProvider", got)
	}
	if got := providerForDir(dir, "todos"); !isFileProvider(got) {
		t.Errorf("providerForDir(_, todos) = %T, want *todos.FileProvider", got)
	}
	// Empty/auto falls back to detection; no .todos here, so Grite.
	if got := providerForDir(dir, ""); !isGriteProvider(got) {
		t.Errorf("providerForDir(_, '') = %T, want *todos.GriteProvider (auto)", got)
	}
}

func TestTodoProviderHonorsExplicitGriteWithDir(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	// Grite scoped to an explicit workspace dir must be allowed (previously
	// rejected with "dir is only supported with provider=todos").
	p, src, err := s.todoProvider(todoSource{Provider: "grite", Dir: workDir})
	if err != nil {
		t.Fatalf("grite with dir errored: %v", err)
	}
	if !isGriteProvider(p) {
		t.Errorf("provider = %T, want *todos.GriteProvider", p)
	}
	if src.Dir != workDir {
		t.Errorf("resolved dir = %q, want %q", src.Dir, workDir)
	}
}

func isFileProvider(p todos.Provider) bool {
	_, ok := p.(*todos.FileProvider)
	return ok
}

// isGriteProvider reports whether p is grite-backed. resolveGrite returns a
// *todos.CachedGriteProvider when the gavel DB is configured and a plain
// *todos.GriteProvider otherwise, so both count as "grite".
func isGriteProvider(p todos.Provider) bool {
	switch p.(type) {
	case *todos.GriteProvider, *todos.CachedGriteProvider:
		return true
	default:
		return false
	}
}

func TestHandleProjectsIncludesTodoCounts(t *testing.T) {
	dir := withProject(t, "gavel", "flanksource/gavel", "")
	provider := todos.NewFileProvider(dir, "")
	if _, err := provider.Create(t.Context(), todos.CreateRequest{
		Title:    "Wire todos",
		Priority: types.PriorityMedium,
		Status:   types.StatusInProgress,
	}); err != nil {
		t.Fatalf("create todo: %v", err)
	}

	rec := httptest.NewRecorder()
	(&Server{}).handleProjects(rec, httptest.NewRequest(http.MethodGet, "/api/projects", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var got []projectInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 1 || got[0].TodoCounts.Open != 1 || got[0].TodoCounts.InProgress != 1 {
		t.Fatalf("unexpected project todo counts: %+v", got)
	}
}

func TestTodoAPITransferMovesBetweenWorkspaces(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: srcDir}}

	created, err := todos.NewFileProvider(srcDir, "").Create(t.Context(), todos.CreateRequest{
		Title:    "Relocate me",
		Body:     "Body that should travel with the todo.",
		Priority: types.PriorityHigh,
		Status:   types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	body, _ := json.Marshal(todoTransferPayload{
		Ref:          todos.TODOReference(created),
		FromDir:      srcDir,
		FromProvider: "todos",
		ToDir:        dstDir,
		ToProvider:   "todos",
	})
	rec := httptest.NewRecorder()
	s.handleTodoTransfer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/transfer", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("transfer status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var resp todoTransferResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal transfer: %v", err)
	}
	if resp.Dir != dstDir || resp.Provider != todos.ProviderFiles {
		t.Fatalf("unexpected transfer target: dir=%q provider=%q", resp.Dir, resp.Provider)
	}
	if resp.Todo.Title != "Relocate me" || resp.Todo.Priority != types.PriorityHigh {
		t.Fatalf("transferred todo lost fields: %+v", resp.Todo)
	}
	if !strings.HasPrefix(resp.Todo.FilePath, dstDir) {
		t.Fatalf("transferred todo not in target dir %q: %s", dstDir, resp.Todo.FilePath)
	}

	// Gone from source, present in target.
	if _, err := os.Stat(created.FilePath); !os.IsNotExist(err) {
		t.Fatalf("expected source todo removed, stat err=%v", err)
	}
	items, err := todos.NewFileProvider(dstDir, "").List(t.Context(), todos.DiscoveryFilters{})
	if err != nil {
		t.Fatalf("target list: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Relocate me" {
		t.Fatalf("unexpected target contents: %+v", items)
	}
}

func TestTodoAPITransferRejectsSameWorkspace(t *testing.T) {
	dir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: dir}}

	created, err := todos.NewFileProvider(dir, "").Create(t.Context(), todos.CreateRequest{
		Title:  "Stay put",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	body, _ := json.Marshal(todoTransferPayload{
		Ref:          todos.TODOReference(created),
		FromDir:      dir,
		FromProvider: "todos",
		ToDir:        dir,
		ToProvider:   "todos",
	})
	rec := httptest.NewRecorder()
	s.handleTodoTransfer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/transfer", strings.NewReader(string(body))))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("same-workspace transfer status = %d, want 400; body = %q", rec.Code, rec.Body.String())
	}
	// The original must survive a rejected transfer.
	if _, err := os.Stat(created.FilePath); err != nil {
		t.Fatalf("expected source todo to survive rejected transfer: %v", err)
	}
}

func TestTodoAPIRunStartsSelectedTodo(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := todos.NewFileProvider(workDir, "").Create(t.Context(), todos.CreateRequest{
		Title:    "Run me",
		Priority: types.PriorityMedium,
		Status:   types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	oldStart := startTodoRun
	var got todoRunRequest
	startTodoRun = func(req todoRunRequest) error {
		got = req
		return nil
	}
	t.Cleanup(func() { startTodoRun = oldStart })

	body, _ := json.Marshal(todoRunPayload{
		Ref:      todos.TODOReference(created),
		Agent:    "codex",
		Mode:     "cmux",
		Model:    "codex",
		Effort:   "high",
		Timeout:  "45m",
		MaxCost:  1.25,
		MaxTurns: 12,
		Dirty:    true,
	})
	rec := httptest.NewRecorder()
	s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run?provider=todos", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("run status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var resp todoRunResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal run response: %v", err)
	}
	if resp.Status != "started" || resp.Provider != todos.ProviderFiles || resp.Agent != "codex" || resp.Mode != "cmux" {
		t.Fatalf("unexpected run response: %+v", resp)
	}
	if resp.Count != 1 {
		t.Fatalf("run count = %d, want 1", resp.Count)
	}
	if len(got.Todos) != 1 || got.Todos[0].Title != "Run me" {
		t.Fatalf("run starter did not receive selected todo: %+v", got.Todos)
	}
	if got.Source.Dir != workDir || got.Backend != todos.ProviderFiles {
		t.Fatalf("unexpected run source: dir=%q backend=%q", got.Source.Dir, got.Backend)
	}
	if got.Options.Model != "codex" || got.Options.Effort != "high" || got.Options.MaxBudget != 1.25 || got.Options.MaxTurns != 12 || !got.Options.Dirty {
		t.Fatalf("unexpected run options: %+v", got.Options)
	}
}

func TestTodoAPIRunStartsMultipleTodosInOneSession(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	provider := todos.NewFileProvider(workDir, "")
	first, err := provider.Create(t.Context(), todos.CreateRequest{
		Title:  "First todo",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed first: %v", err)
	}
	second, err := provider.Create(t.Context(), todos.CreateRequest{
		Title:  "Second todo",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed second: %v", err)
	}

	oldStart := startTodoRun
	var got todoRunRequest
	startTodoRun = func(req todoRunRequest) error {
		got = req
		return nil
	}
	t.Cleanup(func() { startTodoRun = oldStart })

	// Duplicate the first ref to confirm the handler de-duplicates refs.
	body, _ := json.Marshal(todoRunPayload{
		Refs:   []string{todos.TODOReference(first), todos.TODOReference(second), todos.TODOReference(first)},
		Agent:  "claude",
		Mode:   "cmux",
		Model:  "sonnet",
		Effort: "medium",
	})
	rec := httptest.NewRecorder()
	s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run?provider=todos", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("run status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var resp todoRunResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal run response: %v", err)
	}
	if resp.Status != "started" || resp.Count != 2 || len(resp.Refs) != 2 {
		t.Fatalf("unexpected multi-run response: %+v", resp)
	}
	if resp.Message != "Started run for 2 todos" {
		t.Fatalf("message = %q, want %q", resp.Message, "Started run for 2 todos")
	}
	if len(got.Todos) != 2 || got.Todos[0].Title != "First todo" || got.Todos[1].Title != "Second todo" {
		t.Fatalf("run starter did not receive both todos in order: %+v", got.Todos)
	}
}

func TestTodoAPIRunRejectsUnsupportedInlineCodex(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := todos.NewFileProvider(workDir, "").Create(t.Context(), todos.CreateRequest{
		Title:  "Run me",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	body, _ := json.Marshal(todoRunPayload{
		Ref:    todos.TODOReference(created),
		Agent:  "codex",
		Mode:   "inline",
		Model:  "codex",
		Effort: "medium",
	})
	rec := httptest.NewRecorder()
	s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run?provider=todos", strings.NewReader(string(body))))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("run status = %d, want 400; body = %q", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "codex runs require cmux mode") {
		t.Fatalf("unexpected error body: %q", rec.Body.String())
	}
}

func strconvQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
