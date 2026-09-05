package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

func TestTodoAPIPatchPriority(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
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
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodPatch, "/api/todos/item", strings.NewReader(body)))
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
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodPatch, "/api/todos/item", strings.NewReader(empty)))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("empty patch status = %d, want 400", rec.Code)
	}

	// PATCH with an invalid priority is a 400.
	rec = httptest.NewRecorder()
	bad := `{"ref":` + strconvQuote(ref) + `,"priority":"urgent"}`
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodPatch, "/api/todos/item", strings.NewReader(bad)))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("invalid priority status = %d, want 400", rec.Code)
	}
}

func TestTodoAPIPatchEditsTitleBodyAndComments(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:  "Before edit",
		Body:   "Before body",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}
	ref := todos.TODOReference(created)

	patch := func(payload string) todoSummary {
		t.Helper()
		rec := httptest.NewRecorder()
		s.handleTodoItem(rec, httptest.NewRequest(http.MethodPatch, "/api/todos/item", strings.NewReader(payload)))
		if rec.Code != http.StatusOK {
			t.Fatalf("patch %s status = %d, want 200; body = %q", payload, rec.Code, rec.Body.String())
		}
		var out todoSummary
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("unmarshal patch response: %v", err)
		}
		return out
	}

	// Edit title + body.
	edited := patch(`{"ref":` + strconvQuote(ref) + `,"title":"After edit","body":"After body"}`)
	if edited.Title != "After edit" {
		t.Fatalf("title = %q, want After edit", edited.Title)
	}
	if !strings.Contains(edited.Body, "After body") || strings.Contains(edited.Body, "Before body") {
		t.Fatalf("body not replaced: %q", edited.Body)
	}

	// Add a comment only.
	commented := patch(`{"ref":` + strconvQuote(ref) + `,"comment":"please double-check"}`)
	if !strings.Contains(commented.Body, "## Comments") || !strings.Contains(commented.Body, "please double-check") {
		t.Fatalf("comment not recorded in body: %q", commented.Body)
	}

	// Close, then reopen with a comment in one request.
	if got := patch(`{"ref":` + strconvQuote(ref) + `,"status":"completed"}`); got.Status != types.StatusCompleted {
		t.Fatalf("close status = %q, want completed", got.Status)
	}
	reopened := patch(`{"ref":` + strconvQuote(ref) + `,"status":"pending","comment":"reopening to address feedback"}`)
	if reopened.Status != types.StatusPending {
		t.Fatalf("reopen status = %q, want pending", reopened.Status)
	}
	if !strings.Contains(reopened.Body, "reopening to address feedback") {
		t.Fatalf("reopen comment not recorded: %q", reopened.Body)
	}

	// An empty-title edit is rejected.
	rec := httptest.NewRecorder()
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodPatch, "/api/todos/item", strings.NewReader(`{"ref":`+strconvQuote(ref)+`,"title":"   "}`)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty-title patch status = %d, want 400", rec.Code)
	}
}

func TestTodoAPIPatchMultipartCommentWithAttachment(t *testing.T) {
	origAttachmentsDir := attachmentsDir
	attachmentsDir = t.TempDir()
	t.Cleanup(func() { attachmentsDir = origAttachmentsDir })

	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:  "Existing issue",
		Body:   "Original body",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}
	ref := todos.TODOReference(created)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range map[string]string{
		"ref":     ref,
		"comment": "Captured UI element",
	} {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write field %s: %v", key, err)
		}
	}
	part, err := writer.CreateFormFile("attachment", "screen.png")
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
	req := httptest.NewRequest(http.MethodPatch, "/api/todos/item?dir="+url.QueryEscape(workDir), &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	s.handleTodoItem(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("multipart patch status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var patched todoSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &patched); err != nil {
		t.Fatalf("unmarshal patch response: %v", err)
	}
	if !strings.Contains(patched.Body, "Captured UI element") {
		t.Fatalf("comment not recorded: %q", patched.Body)
	}
	if !strings.Contains(patched.Body, "## Attachments") || !strings.Contains(patched.Body, "screen.png") || !strings.Contains(patched.Body, attachmentURLPrefix) {
		t.Fatalf("attachment not appended to comment: %q", patched.Body)
	}
}

func TestTodoAPINativeProviderListsWorkspace(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	if _, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:    "List me",
		Priority: types.PriorityHigh,
		Status:   types.StatusPending,
	}); err != nil {
		t.Fatalf("create todo: %v", err)
	}

	rec := httptest.NewRecorder()
	s.handleTodos(rec, httptest.NewRequest(http.MethodGet, "/api/todos?dir="+url.QueryEscape(workDir), nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var list todoListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if list.Counts.Total != 1 || len(list.Items) != 1 || list.Items[0].Title != "List me" {
		t.Fatalf("native provider did not list the workspace: %+v", list)
	}
}

func TestHandleProjectsIncludesTodoCounts(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "")
	original := projectTodoCounts
	projectTodoCounts = func(context.Context, Project) (todoCounts, error) {
		return todoCounts{Total: 1, Open: 1, InProgress: 1}, nil
	}
	t.Cleanup(func() { projectTodoCounts = original })

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

	created, err := uiTestProviderFor(srcDir).Create(t.Context(), todos.CreateRequest{
		Title:    "Relocate me",
		Body:     "Body that should travel with the todo.",
		Priority: types.PriorityHigh,
		Status:   types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	body, _ := json.Marshal(todoTransferPayload{
		Ref:     todos.TODOReference(created),
		FromDir: srcDir,
		ToDir:   dstDir,
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
	if resp.Dir != dstDir {
		t.Fatalf("unexpected transfer target dir: %q", resp.Dir)
	}
	if resp.Todo.Title != "Relocate me" || resp.Todo.Priority != types.PriorityHigh {
		t.Fatalf("transferred todo lost fields: %+v", resp.Todo)
	}
	// Gone from source, present in target.
	if _, err := uiTestProviderFor(srcDir).Get(t.Context(), created.ID); err == nil {
		t.Fatal("expected source TODO removed")
	}
	items, err := uiTestProviderFor(dstDir).List(t.Context(), todos.DiscoveryFilters{})
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

	created, err := uiTestProviderFor(dir).Create(t.Context(), todos.CreateRequest{
		Title:  "Stay put",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	body, _ := json.Marshal(todoTransferPayload{
		Ref:     todos.TODOReference(created),
		FromDir: dir,
		ToDir:   dir,
	})
	rec := httptest.NewRecorder()
	s.handleTodoTransfer(rec, httptest.NewRequest(http.MethodPost, "/api/todos/transfer", strings.NewReader(string(body))))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("same-workspace transfer status = %d, want 400; body = %q", rec.Code, rec.Body.String())
	}
	// The original must survive a rejected transfer.
	if _, err := uiTestProviderFor(dir).Get(t.Context(), created.ID); err != nil {
		t.Fatalf("expected source todo to survive rejected transfer: %v", err)
	}
}
