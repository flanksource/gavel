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

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// specPayload builds the inlined api.Spec a test payload sends for the
// model/effort pair — the common case most run-payload tests need.
func specPayload(model, effort string) api.Spec {
	return api.Spec{Model: api.Model{Name: model, Effort: api.Effort(effort)}}
}

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

// The list response must expose hasPlan/hasVerification on every item (not
// just detail responses) so the todo row can render its plan/verification
// indicators without a round-trip per row — see HasPlan (todos/plans.go) and
// ExtractVerificationFixture (todos/verification_fixture.go).
func TestTodoAPIListExposesHasPlanAndHasVerification(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	provider := todos.NewFileProvider(workDir, "")

	plain, err := provider.Create(t.Context(), todos.CreateRequest{
		Title:  "Plain todo",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed plain todo: %v", err)
	}

	withFixture, err := provider.Create(t.Context(), todos.CreateRequest{
		Title:  "Todo with a verification fixture",
		Body:   "## Verification\n\n```yaml test\ncommand: go test ./...\n```\n",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed todo with fixture: %v", err)
	}

	awaitingReview, err := provider.Create(t.Context(), todos.CreateRequest{
		Title:  "Todo awaiting plan review",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed todo awaiting review: %v", err)
	}
	planPath := filepath.Join(t.TempDir(), "plan.md")
	if err := os.WriteFile(planPath, []byte("# Plan\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	reviewStatus := types.StatusReview
	if err := todos.UpdateTODOState(awaitingReview, todos.StateUpdate{Status: &reviewStatus, PlanPath: &planPath}); err != nil {
		t.Fatalf("mark awaiting review: %v", err)
	}

	rec := httptest.NewRecorder()
	s.handleTodos(rec, httptest.NewRequest(http.MethodGet, "/api/todos?provider=todos", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var list todoListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	byTitle := map[string]todoSummary{}
	for _, item := range list.Items {
		byTitle[item.Title] = item
	}

	if got := byTitle[plain.Title]; got.HasPlan || got.HasVerification {
		t.Errorf("plain todo = %+v, want both flags false", got)
	}
	if got := byTitle[withFixture.Title]; !got.HasVerification || got.HasPlan {
		t.Errorf("fixture todo = %+v, want hasVerification=true hasPlan=false", got)
	}
	if got := byTitle[awaitingReview.Title]; !got.HasPlan || got.HasVerification {
		t.Errorf("review todo = %+v, want hasPlan=true hasVerification=false", got)
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

func TestTodoNewEndpointFoldsCriteriaIntoBody(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	// A "create todo from PR" request carries the selected failing tests and
	// review comments as acceptance criteria; the server folds them into the body
	// so they round-trip back as the todo's parsed criteria.
	payload := todoNewPayload{todoCreatePayload: todoCreatePayload{
		Title: "Fix failing tests in flanksource/gavel#7",
		Body:  "From flanksource/gavel#7",
		Criteria: []types.AcceptanceCriterion{
			{Text: "Test `TestParser` passes"},
			{Text: "Resolve @reviewer's comment on parser.go:42"},
		},
	}}
	autoSave := true
	payload.AutoSave = &autoSave
	raw, _ := json.Marshal(payload)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/todos/new?provider=todos&dir="+url.QueryEscape(workDir), bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("new status = %d, want %d; body = %q", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var resp todoNewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal new response: %v", err)
	}
	if len(resp.Todo.Criteria) != 2 {
		t.Fatalf("criteria not parsed back onto todo: %+v", resp.Todo.Criteria)
	}
	if resp.Todo.Criteria[0].Text != "Test `TestParser` passes" || resp.Todo.Criteria[1].Text != "Resolve @reviewer's comment on parser.go:42" {
		t.Fatalf("unexpected criteria: %+v", resp.Todo.Criteria)
	}
	if !strings.Contains(resp.Todo.Body, "## Acceptance Criteria") {
		t.Fatalf("body missing acceptance-criteria section: %q", resp.Todo.Body)
	}
	if !strings.Contains(resp.Todo.Body, "From flanksource/gavel#7") {
		t.Fatalf("body dropped the PR reference: %q", resp.Todo.Body)
	}
}

func TestTodoNewEndpointAddsPRVerificationFixture(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	payload := todoNewPayload{todoCreatePayload: todoCreatePayload{
		Title: "Fix PR feedback",
		Body:  "From flanksource/gavel#7",
		Criteria: []types.AcceptanceCriterion{
			{Text: "Test `TestParser` passes"},
		},
		PRVerification: &todoPRVerificationPayload{
			PRNumber:   7,
			Repo:       "flanksource/gavel",
			CommentIDs: []int64{102, 101},
			Actions:    []string{"*"},
		},
	}}
	autoSave := true
	payload.AutoSave = &autoSave
	raw, _ := json.Marshal(payload)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/todos/new?provider=todos&dir="+url.QueryEscape(workDir), bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("new status = %d, want %d; body = %q", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var resp todoNewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal new response: %v", err)
	}
	if len(resp.Todo.Criteria) != 1 || resp.Todo.Criteria[0].Text != "Test `TestParser` passes" {
		t.Fatalf("criteria not preserved: %+v", resp.Todo.Criteria)
	}
	want := "gavel pr status 7 --repo flanksource/gavel --comments 101,102 --actions '*'"
	if !strings.Contains(resp.Todo.VerificationMarkdown, want) {
		t.Fatalf("verification fixture missing command %q: %q", want, resp.Todo.VerificationMarkdown)
	}
	if strings.Contains(resp.Todo.Body, "CI check") || strings.Contains(resp.Todo.Body, "Address @") {
		t.Fatalf("PR gates should not be duplicated as criteria: %q", resp.Todo.Body)
	}
}

func TestTodoNewEndpointMultipartFiles(t *testing.T) {
	attachmentsDir = t.TempDir()
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
	if resp.Attachments[0].URL == "" {
		t.Fatalf("attachment missing served URL: %+v", resp.Attachments[0])
	}
	if !strings.Contains(resp.Todo.Body, "## Attachments") || !strings.Contains(resp.Todo.Body, resp.Attachments[0].URL) {
		t.Fatalf("created body missing attachment reference: %q", resp.Todo.Body)
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

func TestTodoAPIPatchEditsTitleBodyAndComments(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	created, err := todos.NewFileProvider(workDir, "").Create(t.Context(), todos.CreateRequest{
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
		s.handleTodoItem(rec, httptest.NewRequest(http.MethodPatch, "/api/todos/item?provider=todos", strings.NewReader(payload)))
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
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodPatch, "/api/todos/item?provider=todos", strings.NewReader(`{"ref":`+strconvQuote(ref)+`,"title":"   "}`)))
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

	created, err := todos.NewFileProvider(workDir, "").Create(t.Context(), todos.CreateRequest{
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
	req := httptest.NewRequest(http.MethodPatch, "/api/todos/item?provider=todos&dir="+url.QueryEscape(workDir), &body)
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
		Ref:   todos.TODOReference(created),
		Agent: "codex",
		Mode:  "cmux",
		Spec: api.Spec{
			Model:  api.Model{Name: "codex", Effort: "high"},
			Budget: api.Budget{Cost: 1.25, MaxTurns: 12, Timeout: "45m"},
			// Dirty-worktree now rides the spec's Setup.Checkout.Dirty (Workspace
			// section) instead of a sibling flag.
			Setup: &shell.Setup{Checkout: &shell.Checkout{Dirty: &shell.Dirty{Stash: shell.StashAll}}},
		},
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
	if got.Options.Name != "gpt-5.5" || got.Options.Backend != "codex-cmux" || got.Options.Effort != "high" || got.Options.Budget.Cost != 1.25 || got.Options.Budget.MaxTurns != 12 || !specDirty(got.Options.Spec) {
		t.Fatalf("unexpected run options: %+v", got.Options)
	}
}

func TestTodoAPIRunPreviewReturnsPrompt(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := todos.NewFileProvider(workDir, "").Create(t.Context(), todos.CreateRequest{
		Title:    "Fix the parser",
		Body:     "The parser drops trailing commas.",
		Priority: types.PriorityMedium,
		Status:   types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}
	ref := todos.TODOReference(created)

	preview := func(payload todoRunPayload) todoRunPreviewResponse {
		t.Helper()
		body, _ := json.Marshal(payload)
		rec := httptest.NewRecorder()
		s.handleTodoRunPreview(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run/preview?provider=todos", strings.NewReader(string(body))))
		if rec.Code != http.StatusOK {
			t.Fatalf("preview status = %d, want 200; body = %q", rec.Code, rec.Body.String())
		}
		var resp todoRunPreviewResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal preview response: %v", err)
		}
		return resp
	}

	cmuxResp := preview(todoRunPayload{Ref: ref, Agent: "claude", Mode: "cmux", Spec: specPayload("claude", "high")})
	if cmuxResp.Count != 1 || cmuxResp.Mode != "cmux" {
		t.Fatalf("unexpected preview meta: %+v", cmuxResp)
	}
	if !strings.Contains(cmuxResp.Prompt, "## Fix the parser") {
		t.Fatalf("cmux preview should contain the title heading: %q", cmuxResp.Prompt)
	}
	if !strings.Contains(cmuxResp.Prompt, "The parser drops trailing commas.") {
		t.Fatalf("cmux preview should inline the body: %q", cmuxResp.Prompt)
	}
	if !strings.Contains(cmuxResp.Prompt, "Think hard and reason thoroughly") {
		t.Fatalf("cmux preview should include the effort directive: %q", cmuxResp.Prompt)
	}
	if !strings.Contains(cmuxResp.Prompt, "## Instructions") {
		t.Fatalf("cmux preview should include the instructions section: %q", cmuxResp.Prompt)
	}

	planResp := preview(todoRunPayload{Ref: ref, Agent: "claude", Mode: "cmux", Spec: specPayload("claude", "medium"), Plan: true})
	if !strings.Contains(planResp.Prompt, "## Fix the parser") {
		t.Fatalf("plan preview should contain the title: %q", planResp.Prompt)
	}

	inlineResp := preview(todoRunPayload{Ref: ref, Agent: "claude", Mode: "inline", Spec: specPayload("claude", "medium")})
	if strings.HasPrefix(inlineResp.Prompt, "# Fix the parser") {
		t.Fatalf("inline preview should be the bare claude prompt, not the cmux instruction: %q", inlineResp.Prompt)
	}
	if !strings.Contains(inlineResp.Prompt, "The parser drops trailing commas.") {
		t.Fatalf("inline preview should include the body: %q", inlineResp.Prompt)
	}
}

func TestTodoRunPreviewAbsolutizesAttachmentURLs(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := todos.NewFileProvider(workDir, "").Create(t.Context(), todos.CreateRequest{
		Title:  "Screenshot todo",
		Body:   "See the bug:\n\n![screen.png](" + attachmentURLPrefix + "abc.png)",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}
	ref := todos.TODOReference(created)

	body, _ := json.Marshal(todoRunPayload{Ref: ref, Agent: "claude", Mode: "inline", Spec: api.Spec{Model: api.Model{Name: "claude"}}})
	rec := httptest.NewRecorder()
	s.handleTodoRunPreview(rec, httptest.NewRequest(http.MethodPost, "http://gavel.example:9092/api/todos/run/preview?provider=todos", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("preview status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var resp todoRunPreviewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal preview response: %v", err)
	}

	wantURL := "http://gavel.example:9092" + attachmentURLPrefix + "abc.png"
	if !strings.Contains(resp.Prompt, wantURL) {
		t.Fatalf("prompt should reference the absolute attachment URL %q: %q", wantURL, resp.Prompt)
	}
	if strings.Contains(resp.Prompt, "]("+attachmentURLPrefix) {
		t.Fatalf("prompt should not retain relative attachment links: %q", resp.Prompt)
	}
}

func TestNormalizeTodoRunOptionsCommitFromWorkflow(t *testing.T) {
	base := todoRunPayload{Agent: "claude", Mode: "cmux", Spec: specPayload("claude", "medium")}
	commitWorkflow := func(commit bool) *api.Workflow {
		return &api.Workflow{PostRun: &api.PostRun{Commit: commit}}
	}

	// Auto-commit is sourced from Workflow.PostRun.Commit — no server-side default.
	// The run dialog seeds postRun.commit=true so a default dashboard run commits;
	// an absent workflow (or an explicit false) means no commit.
	cases := []struct {
		name     string
		workflow *api.Workflow
		want     bool
	}{
		{"no workflow means no auto-commit", nil, false},
		{"postRun.commit true auto-commits", commitWorkflow(true), true},
		{"postRun.commit false skips commit", commitWorkflow(false), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := base
			payload.Workflow = tc.workflow
			opts, err := normalizeTodoRunOptions(payload)
			if err != nil {
				t.Fatalf("normalize: %v", err)
			}
			if got := specCommit(opts.Spec); got != tc.want {
				t.Fatalf("specCommit = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestNormalizeTodoRunOptionsToolPreferences(t *testing.T) {
	base := todoRunPayload{Agent: "claude", Mode: "cmux", Spec: specPayload("claude", "medium")}

	t.Run("valid prefs and permission mode are threaded", func(t *testing.T) {
		payload := base
		payload.Spec.Permissions = api.Permissions{
			Mode: api.PermissionAcceptEdits,
			Tools: api.Tools{Modes: map[string]api.ToolMode{
				"Bash": api.ToolModeAsk, "Write": api.ToolModeDisabled, "Read": api.ToolModeEnabled,
			}},
		}
		opts, err := normalizeTodoRunOptions(payload)
		if err != nil {
			t.Fatalf("normalize: %v", err)
		}
		if opts.Permissions.Mode != api.PermissionAcceptEdits {
			t.Fatalf("PermissionMode = %q, want acceptEdits", opts.Permissions.Mode)
		}
		toolModes := toolModesFromPermissions(opts.Permissions.Tools)
		if toolModes["Bash"] != "ask" || toolModes["Write"] != "disabled" || toolModes["Read"] != "enabled" {
			t.Fatalf("ToolModes = %v, want Bash=ask Write=disabled Read=enabled", toolModes)
		}
	})

	t.Run("empty prefs normalize to nil", func(t *testing.T) {
		opts, err := normalizeTodoRunOptions(base)
		if err != nil {
			t.Fatalf("normalize: %v", err)
		}
		if toolModesFromPermissions(opts.Permissions.Tools) != nil || opts.Permissions.Mode != "" {
			t.Fatalf("want nil ToolModes and empty PermissionMode, got %v / %q", opts.Permissions.Tools, opts.Permissions.Mode)
		}
	})

	t.Run("invalid tool mode fails loud", func(t *testing.T) {
		payload := base
		payload.Spec.Permissions = api.Permissions{Tools: api.Tools{Modes: map[string]api.ToolMode{"Bash": "maybe"}}}
		if _, err := normalizeTodoRunOptions(payload); err == nil {
			t.Fatal("expected error for invalid tool mode, got nil")
		}
	})

	t.Run("invalid permission mode fails loud", func(t *testing.T) {
		payload := base
		payload.Spec.Permissions = api.Permissions{Mode: "yolo"}
		if _, err := normalizeTodoRunOptions(payload); err == nil {
			t.Fatal("expected error for invalid permission mode, got nil")
		}
	})
}

func TestTodoAPIRunThreadsCommitOption(t *testing.T) {
	cases := []struct {
		name     string
		workflow *api.Workflow
		want     bool
	}{
		{"postRun commit auto-commits", &api.Workflow{PostRun: &api.PostRun{Commit: true}}, true},
		{"absent workflow skips commit", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			workDir := t.TempDir()
			s := &Server{ghOpts: github.Options{WorkDir: workDir}}
			created, err := todos.NewFileProvider(workDir, "").Create(t.Context(), todos.CreateRequest{
				Title:  "Run me",
				Status: types.StatusPending,
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

			payload := todoRunPayload{
				Ref:   todos.TODOReference(created),
				Agent: "claude",
				Mode:  "cmux",
				Spec:  specPayload("claude", "medium"),
			}
			payload.Workflow = tc.workflow
			body, _ := json.Marshal(payload)
			rec := httptest.NewRecorder()
			s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run?provider=todos", strings.NewReader(string(body))))
			if rec.Code != http.StatusOK {
				t.Fatalf("run status = %d, want 200; body = %q", rec.Code, rec.Body.String())
			}
			var resp todoRunResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("unmarshal run response: %v", err)
			}
			if resp.Commit != tc.want {
				t.Fatalf("response Commit = %v, want %v", resp.Commit, tc.want)
			}
			if got := specCommit(got.Options.Spec); got != tc.want {
				t.Fatalf("run starter commit = %v, want %v", got, tc.want)
			}
		})
	}
}

// A dry run (Workflow.PostRun.DryRun) still executes the agent — it only skips the
// post-run auto-commit — so handleTodoRun starts it and reports Commit:false.
func TestTodoAPIRunDryRunStartsButSkipsCommit(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := todos.NewFileProvider(workDir, "").Create(t.Context(), todos.CreateRequest{
		Title:  "Run me",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	oldStart := startTodoRun
	started := false
	startTodoRun = func(todoRunRequest) error { started = true; return nil }
	t.Cleanup(func() { startTodoRun = oldStart })

	payload := todoRunPayload{
		Ref:   todos.TODOReference(created),
		Agent: "claude",
		Mode:  "cmux",
		Spec:  specPayload("claude", "medium"),
	}
	// commit=true but dryRun=true: the run executes, the commit is suppressed.
	payload.Workflow = &api.Workflow{PostRun: &api.PostRun{Commit: true, DryRun: true}}
	body, _ := json.Marshal(payload)
	rec := httptest.NewRecorder()
	s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run?provider=todos", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("run status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var resp todoRunResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal run response: %v", err)
	}
	if resp.Status != "started" {
		t.Fatalf("status = %q, want started (a dry run still executes)", resp.Status)
	}
	if resp.Commit {
		t.Fatal("response Commit = true, want false (dry run suppresses the commit)")
	}
	if !started {
		t.Fatal("dry run did not start the agent run")
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
		Refs:  []string{todos.TODOReference(first), todos.TODOReference(second), todos.TODOReference(first)},
		Agent: "claude",
		Mode:  "cmux",
		Spec:  specPayload("sonnet", "medium"),
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
		Ref:   todos.TODOReference(created),
		Agent: "codex",
		Mode:  "inline",
		Spec:  specPayload("codex", "medium"),
	})
	rec := httptest.NewRecorder()
	s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run?provider=todos", strings.NewReader(string(body))))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("run status = %d, want 400; body = %q", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "codex runs require a cmux driver") {
		t.Fatalf("unexpected error body: %q", rec.Body.String())
	}
}

func TestTodoAPIRunPlanThreadsPlanOption(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := todos.NewFileProvider(workDir, "").Create(t.Context(), todos.CreateRequest{
		Title:  "Plan me",
		Status: types.StatusPending,
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
		Ref:   todos.TODOReference(created),
		Agent: "claude",
		Mode:  "cmux",
		Spec:  specPayload("claude", "medium"),
		Plan:  true,
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
	if !resp.Plan || resp.RunMode != "plan" {
		t.Fatalf("response did not echo plan mode: %+v", resp)
	}
	if got.Options.RunMode != types.ModePlan {
		t.Fatalf("run starter did not receive plan mode: %+v", got.Options)
	}
}

// Plan runs work on every driver now (the plan template's frontmatter carries
// the plan posture), so a non-cmux plan run is accepted, and the new runMode
// field supersedes the plan bool.
func TestTodoAPIRunModeField(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := todos.NewFileProvider(workDir, "").Create(t.Context(), todos.CreateRequest{
		Title:  "Plan me",
		Status: types.StatusPending,
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
		Ref:     todos.TODOReference(created),
		Driver:  "claude-headless",
		Spec:    specPayload("claude", "medium"),
		RunMode: "plan",
	})
	rec := httptest.NewRecorder()
	s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run?provider=todos", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("run status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if got.Options.RunMode != types.ModePlan {
		t.Fatalf("runMode not resolved: %+v", got.Options)
	}

	// verify routes through its own endpoint, not the run endpoint.
	body, _ = json.Marshal(todoRunPayload{
		Ref:     todos.TODOReference(created),
		Driver:  "claude-headless",
		RunMode: "verify",
	})
	rec = httptest.NewRecorder()
	s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run?provider=todos", strings.NewReader(string(body))))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("verify runMode status = %d, want 400; body = %q", rec.Code, rec.Body.String())
	}
}

func TestNormalizeTodoRunOptionsDriverField(t *testing.T) {
	// Explicit driver wins and sets the legacy mode label for downstream paths.
	opts, err := normalizeTodoRunOptions(todoRunPayload{Driver: "claude-headless", Spec: api.Spec{Model: api.Model{Effort: "medium"}}})
	if err != nil {
		t.Fatalf("claude-headless driver: %v", err)
	}
	if opts.Driver != "claude-headless" || opts.Agent != "claude" || opts.Mode != "inline" {
		t.Fatalf("got driver=%q agent=%q mode=%q", opts.Driver, opts.Agent, opts.Mode)
	}

	// codex-cmux with no model resolves the codex agent and defaults to the
	// captain whoami-backed cmux model.
	opts, err = normalizeTodoRunOptions(todoRunPayload{Driver: "codex-cmux"})
	if err != nil {
		t.Fatalf("codex-cmux driver: %v", err)
	}
	if opts.Agent != "codex" || opts.Backend != "codex-cmux" || opts.Name != "gpt-5.5" || opts.Mode != "cmux" {
		t.Fatalf("got agent=%q backend=%q model=%q mode=%q", opts.Agent, opts.Backend, opts.Name, opts.Mode)
	}

	if _, err := normalizeTodoRunOptions(todoRunPayload{Driver: "claude-tui"}); err == nil {
		t.Fatal("invalid driver should be rejected")
	}
}

func TestTodoRunContextListsCaptainBackends(t *testing.T) {
	prev := runCaptainWhoami
	calls := 0
	runCaptainWhoami = func(opts captaincli.WhoamiOptions) (any, error) {
		calls++
		if opts.Backend != "" || !opts.Models {
			t.Fatalf("whoami opts = %+v, want one unfiltered model snapshot", opts)
		}
		adapters := make([]captaincli.AdapterStatus, 0, 5)
		for _, backend := range []string{"claude-cmux", "claude-agent", "claude-cli"} {
			adapters = append(adapters, captaincli.AdapterStatus{
				Backend:       backend,
				Type:          "cli",
				Authenticated: true,
				Binary:        "/usr/local/bin/claude",
				ModelCount:    3,
				Models:        []string{"claude-sonnet-5", "claude-fable-5", "claude-opus-4-8"},
				ModelDetails: []captainai.ModelDef{
					{ID: "claude-sonnet-5", Name: "Claude Sonnet 5", CapabilitiesKnown: true, Reasoning: true, SupportedEfforts: []api.Effort{api.EffortLow, api.EffortMedium, api.EffortHigh, api.EffortXHigh, api.EffortMax}, DefaultEffort: api.EffortMedium},
					{ID: "claude-fable-5", Name: "Claude Fable 5", CapabilitiesKnown: true, Reasoning: true, SupportedEfforts: []api.Effort{api.EffortLow, api.EffortMedium, api.EffortHigh, api.EffortXHigh, api.EffortMax}, DefaultEffort: api.EffortMedium},
					{ID: "claude-opus-4-8", Name: "Claude Opus 4.8", CapabilitiesKnown: true, Reasoning: true, SupportedEfforts: []api.Effort{api.EffortLow, api.EffortMedium, api.EffortHigh, api.EffortXHigh, api.EffortMax}, DefaultEffort: api.EffortMedium},
				},
			})
		}
		for _, backend := range []string{"codex-cmux", "codex-agent"} {
			adapters = append(adapters, captaincli.AdapterStatus{
				Backend:       backend,
				Type:          "cli",
				Authenticated: true,
				Binary:        "/usr/local/bin/codex",
				ModelCount:    4,
				Models:        []string{"gpt-5.5", "gpt-5.6-sol", "gpt-5.6-luna", "gpt-5.6-terra"},
				ModelDetails: []captainai.ModelDef{
					{ID: "gpt-5.5", Name: "GPT-5.5", CapabilitiesKnown: true, Reasoning: true, SupportedEfforts: []api.Effort{api.EffortLow, api.EffortMedium, api.EffortHigh, api.EffortXHigh}, DefaultEffort: api.EffortMedium},
					{ID: "gpt-5.6-sol", Name: "GPT-5.6 Sol", CapabilitiesKnown: true, Reasoning: true, SupportedEfforts: []api.Effort{api.EffortLow, api.EffortMedium, api.EffortHigh, api.EffortXHigh, api.EffortMax, api.EffortUltra}, DefaultEffort: api.EffortMedium},
					{ID: "gpt-5.6-luna", Name: "GPT-5.6 Luna", CapabilitiesKnown: true, Reasoning: true, SupportedEfforts: []api.Effort{api.EffortLow, api.EffortMedium, api.EffortHigh, api.EffortXHigh, api.EffortMax}, DefaultEffort: api.EffortMedium},
					{ID: "gpt-5.6-terra", Name: "GPT-5.6 Terra", CapabilitiesKnown: true, Reasoning: true, SupportedEfforts: []api.Effort{api.EffortLow, api.EffortMedium, api.EffortHigh, api.EffortXHigh, api.EffortMax}, DefaultEffort: api.EffortMedium},
				},
			})
		}
		return captaincli.WhoamiResult{Adapters: adapters}, nil
	}
	t.Cleanup(func() { runCaptainWhoami = prev })

	resp := todoRunContext()
	if calls != 1 {
		t.Fatalf("captain whoami calls = %d, want 1", calls)
	}
	if !stringSliceContains(resp.Efforts, "xhigh") {
		t.Fatalf("efforts = %v, want captain xhigh effort", resp.Efforts)
	}
	if resp.DefaultBackend != "claude-agent" {
		t.Fatalf("default backend = %q, want claude-agent", resp.DefaultBackend)
	}
	backends := map[string]todoRunBackendOption{}
	for _, backend := range resp.Backends {
		backends[backend.ID] = backend
	}
	for _, id := range []string{"claude-cmux", "claude-agent", "claude-cli", "codex-cmux", "codex-agent"} {
		if backends[id].ID == "" {
			t.Fatalf("missing backend %q in %+v", id, resp.Backends)
		}
		if len(backends[id].Models) == 0 {
			t.Fatalf("backend %q has no model list: %+v", id, backends[id])
		}
	}
	if backends["codex-cli"].ID != "" {
		t.Fatalf("unexpected codex-cli backend in %+v", resp.Backends)
	}
	if backends["claude-cmux"].Driver != "claude-cmux" || backends["claude-agent"].Driver != "claude-headless" || backends["codex-agent"].Driver != "codex-headless" {
		t.Fatalf("unexpected backend drivers: claude-cmux=%q claude-agent=%q codex-agent=%q", backends["claude-cmux"].Driver, backends["claude-agent"].Driver, backends["codex-agent"].Driver)
	}
	if !todoRunModelsContain(backends["claude-cmux"].Models, "claude-sonnet-5") {
		t.Fatalf("claude-cmux models = %+v, want claude-sonnet-5 from captain whoami", backends["claude-cmux"].Models)
	}
	if !todoRunModelsContain(backends["claude-cmux"].Models, "claude-opus-4-8") {
		t.Fatalf("claude-cmux models = %+v, want claude-opus-4-8 from captain whoami", backends["claude-cmux"].Models)
	}
	if !todoRunModelsContain(backends["claude-agent"].Models, "claude-fable-5") {
		t.Fatalf("claude-agent models = %+v, want Fable from captain whoami", backends["claude-agent"].Models)
	}
	if len(backends["claude-cmux"].Models) != 3 || backends["claude-cmux"].Models[0].ID != "claude-sonnet-5" || backends["claude-cmux"].Models[1].ID != "claude-fable-5" || backends["claude-cmux"].Models[2].ID != "claude-opus-4-8" {
		t.Fatalf("claude-cmux models = %+v, want Captain model-detail order", backends["claude-cmux"].Models)
	}
	if todoRunModelsContain(backends["claude-cmux"].Models, "claude-agent-sonnet") {
		t.Fatalf("claude-cmux models = %+v, should not expose synthetic claude-agent aliases", backends["claude-cmux"].Models)
	}
	if !todoRunModelsContain(backends["codex-cmux"].Models, "gpt-5.5") {
		t.Fatalf("codex-cmux models = %+v, want gpt-5.5 from captain whoami", backends["codex-cmux"].Models)
	}
	for _, id := range []string{"gpt-5.6-sol", "gpt-5.6-luna", "gpt-5.6-terra"} {
		if !todoRunModelsContain(backends["codex-agent"].Models, id) {
			t.Fatalf("codex-agent models = %+v, want %s from captain whoami", backends["codex-agent"].Models, id)
		}
	}
	sol := todoRunModelByID(backends["codex-agent"].Models, "gpt-5.6-sol")
	if !sol.CapabilitiesKnown || !sol.Reasoning || sol.DefaultEffort != "medium" || !stringSliceContains(sol.SupportedEfforts, "ultra") || sol.Temperature == nil || *sol.Temperature {
		t.Fatalf("gpt-5.6-sol capabilities = %+v, want exact effort and temperature metadata", sol)
	}
	if todoRunModelsContain(backends["codex-cmux"].Models, "gpt-5-codex") {
		t.Fatalf("codex-cmux models = %+v, should not include code variant gpt-5-codex", backends["codex-cmux"].Models)
	}
}

func TestNormalizeTodoRunOptionsCaptainBackend(t *testing.T) {
	opts, err := normalizeTodoRunOptions(todoRunPayload{Driver: "claude-headless", Spec: api.Spec{Model: api.Model{Backend: "claude-cli", Effort: "xhigh"}}})
	if err != nil {
		t.Fatalf("claude-cli backend: %v", err)
	}
	if opts.Backend != "claude-cli" || opts.Name != "claude-sonnet-5" || opts.Effort != "xhigh" {
		t.Fatalf("unexpected claude backend options: %+v", opts)
	}

	opts, err = normalizeTodoRunOptions(todoRunPayload{Driver: "codex-headless", Spec: api.Spec{Model: api.Model{Name: "codex"}}})
	if err != nil {
		t.Fatalf("default codex backend: %v", err)
	}
	if opts.Backend != "codex-agent" || opts.Name != "gpt-5.5" {
		t.Fatalf("unexpected codex backend defaults: %+v", opts)
	}

	if _, err := normalizeTodoRunOptions(todoRunPayload{Driver: "codex-headless", Spec: api.Spec{Model: api.Model{Backend: "codex-cli", Name: "gpt-5.5"}}}); err == nil {
		t.Fatal("codex-cli backend should be rejected for codex-headless")
	}

	opts, err = normalizeTodoRunOptions(todoRunPayload{Driver: "claude-cmux", Spec: api.Spec{Model: api.Model{Name: "sonnet-4-6"}}})
	if err != nil {
		t.Fatalf("versioned claude cmux model: %v", err)
	}
	if opts.Backend != "claude-cmux" || opts.Name != "claude-sonnet-4-6" {
		t.Fatalf("unexpected versioned claude model options: %+v", opts)
	}

	opts, err = normalizeTodoRunOptions(todoRunPayload{Driver: "claude-headless", Spec: api.Spec{Model: api.Model{Backend: "claude-agent", Name: "opus-4-8"}}})
	if err != nil {
		t.Fatalf("stale claude agent opus model: %v", err)
	}
	if opts.Backend != "claude-agent" || opts.Name != "claude-opus-4-8" {
		t.Fatalf("unexpected normalized opus model options: %+v", opts)
	}

	if _, err := normalizeTodoRunOptions(todoRunPayload{Driver: "codex-headless", Spec: api.Spec{Model: api.Model{Backend: "claude-agent", Name: "claude-sonnet-5"}}}); err == nil {
		t.Fatal("mismatched backend and driver should be rejected")
	}
	if _, err := normalizeTodoRunOptions(todoRunPayload{Driver: "claude-headless", Spec: api.Spec{Model: api.Model{Backend: "claude-agent", Name: "gpt-5-codex"}}}); err == nil {
		t.Fatal("mismatched backend and model should be rejected")
	}
}

func TestTodoAPIRunThreadsCaptainBackend(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := todos.NewFileProvider(workDir, "").Create(t.Context(), todos.CreateRequest{
		Title:  "Run headless",
		Status: types.StatusPending,
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
		Ref:    todos.TODOReference(created),
		Driver: "claude-headless",
		Spec:   api.Spec{Model: api.Model{Backend: "claude-agent", Name: "claude-sonnet-5", Effort: "medium"}},
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
	if resp.Driver != "claude-headless" || resp.Backend != "claude-agent" || resp.Model != "claude-sonnet-5" {
		t.Fatalf("response did not thread backend/model: %+v", resp)
	}
	if got.Options.Backend != "claude-agent" || got.Options.Name != "claude-sonnet-5" {
		t.Fatalf("run starter did not receive backend/model: %+v", got.Options)
	}
}

func TestTodoRunModelLabelFormatsVersionedClaudeModels(t *testing.T) {
	cases := map[string]string{
		"sonnet-5":              "Sonnet 5",
		"sonnet-4-6":            "Sonnet 4.6",
		"opus-4-8":              "Opus 4.8",
		"claude-agent-opus-4-6": "Opus 4.6",
		"claude-sonnet-5":       "Sonnet 5",
		"claude-agent-sonnet":   "Sonnet",
		"gpt-5-codex":           "GPT 5 Codex",
		"codex-gpt-5-codex":     "GPT 5 Codex",
		"claude-code-haiku-4-5": "Haiku 4.5",
		"claude-agent-fable-5":  "Fable 5",
	}
	for in, want := range cases {
		if got := modelLabel(in); got != want {
			t.Errorf("modelLabel(%q) = %q, want %q", in, got, want)
		}
	}
}

func stringSliceContains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func todoRunModelsContain(items []todoRunModelOption, want string) bool {
	for _, item := range items {
		if item.ID == want {
			return true
		}
	}
	return false
}

func todoRunModelByID(items []todoRunModelOption, want string) todoRunModelOption {
	for _, item := range items {
		if item.ID == want {
			return item
		}
	}
	return todoRunModelOption{}
}

func strconvQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
