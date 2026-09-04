package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
)

// resolvedName is the exact catalog id a runtime hands its driver. Run options
// are resolved at the API boundary, so a test that posts an alias ("codex")
// asserts against what the registry expands it to rather than restating it.
func resolvedName(t *testing.T, model api.Model) string {
	t.Helper()
	resolved, err := captainai.Resolve(model)
	if err != nil {
		t.Fatalf("resolve %+v: %v", model, err)
	}
	return resolved.Name
}

// specPayload builds the inlined api.Spec a test payload sends for the
// model/effort pair — the common case most run-payload tests need.
func specPayload(model, effort string) api.Spec {
	return api.Spec{Model: api.Model{Name: model, Effort: api.Effort(effort)}}
}

func TestTodoAPINativeCRUD(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	createBody := `{"title":"Fix workspace","body":"Implement todo tab","priority":"high","status":"pending"}`
	rec := httptest.NewRecorder()
	s.handleTodos(rec, httptest.NewRequest(http.MethodPost, "/api/todos", strings.NewReader(createBody)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %q", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var created todoSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create: %v", err)
	}
	if created.Title != "Fix workspace" || created.Status != types.StatusPending || created.Priority != types.PriorityHigh {
		t.Fatalf("unexpected created todo: %+v", created)
	}
	if created.Lifecycle == nil || len(created.Lifecycle.Steps) == 0 {
		t.Fatalf("create response missing lifecycle steps: %+v", created)
	}
	rec = httptest.NewRecorder()
	s.handleTodos(rec, httptest.NewRequest(http.MethodGet, "/api/todos", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var list todoListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if list.Counts.Total != 1 || list.Counts.Open != 1 || list.Counts.Pending != 1 {
		t.Fatalf("unexpected counts: %+v", list.Counts)
	}

	rec = httptest.NewRecorder()
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodGet, "/api/todos/item?ref="+url.QueryEscape(created.Ref), nil))
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
	patchBody := `{"ref":` + strconvQuote(created.Ref) + `,"status":"pending"}`
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodPatch, "/api/todos/item", strings.NewReader(patchBody)))
	if rec.Code != http.StatusOK {
		t.Fatalf("patch status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var patched todoSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &patched); err != nil {
		t.Fatalf("unmarshal patch: %v", err)
	}
	if patched.Status != types.StatusPending {
		t.Fatalf("status = %q, want pending", patched.Status)
	}
	// A stale-lifecycle regression: every write path must attach the lifecycle
	// so the dashboard's cache overwrite doesn't blank the step strip.
	if patched.Lifecycle == nil || patched.Lifecycle.Next != "plan" {
		t.Fatalf("patch response lifecycle.next = %+v, want next=\"plan\"", patched.Lifecycle)
	}

	rec = httptest.NewRecorder()
	completeBody := `{"ref":` + strconvQuote(created.Ref) + `,"status":"completed"}`
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodPatch, "/api/todos/item", strings.NewReader(completeBody)))
	if rec.Code != http.StatusOK {
		t.Fatalf("patch status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &patched); err != nil {
		t.Fatalf("unmarshal patch: %v", err)
	}
	if patched.Status != types.StatusCompleted {
		t.Fatalf("status = %q, want completed", patched.Status)
	}

	rec = httptest.NewRecorder()
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodDelete, "/api/todos/item?ref="+url.QueryEscape(created.Ref), nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if _, err := uiTestProviderFor(workDir).Get(t.Context(), created.Ref); err == nil {
		t.Fatal("archived TODO remained in the native test provider")
	}
}

// A create body with a key the payload does not declare is rejected before it
// ever reaches the provider, and the error names the offending key so the
// caller knows what to drop.
func TestTodoAPICreateRejectsUnknownField(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	rec := httptest.NewRecorder()
	body := `{"title":"Fix workspace","bogus":true}`
	s.handleTodos(rec, httptest.NewRequest(http.MethodPost, "/api/todos", strings.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("create status = %d, want %d; body = %q", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "bogus") {
		t.Fatalf("body does not name the offending key: %q", rec.Body.String())
	}
}

// in_progress, review, ask, failed and unverified are projections of the last
// run's execution state. Storage declines to persist them, so accepting one
// would return 200/201 while changing nothing.
func TestTodoAPIRejectsProjectedStatusWrites(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	rec := httptest.NewRecorder()
	createBody := `{"title":"Seed","body":"seed body","status":"pending"}`
	s.handleTodos(rec, httptest.NewRequest(http.MethodPost, "/api/todos", strings.NewReader(createBody)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %q", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var created todoSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create: %v", err)
	}

	for _, status := range []types.Status{
		types.StatusInProgress, types.StatusReview, types.StatusAsk,
		types.StatusFailed, types.StatusUnverified,
	} {
		t.Run("patch "+string(status), func(t *testing.T) {
			rec := httptest.NewRecorder()
			body := `{"ref":` + strconvQuote(created.Ref) + `,"status":"` + string(status) + `"}`
			s.handleTodoItem(rec, httptest.NewRequest(http.MethodPatch, "/api/todos/item", strings.NewReader(body)))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("patch %q = %d, want 400; body = %q", status, rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), "projected") {
				t.Fatalf("patch %q error = %q, want it to explain the status is projected", status, rec.Body.String())
			}
		})

		t.Run("create "+string(status), func(t *testing.T) {
			rec := httptest.NewRecorder()
			body := `{"title":"Projected","body":"x","status":"` + string(status) + `"}`
			s.handleTodos(rec, httptest.NewRequest(http.MethodPost, "/api/todos", strings.NewReader(body)))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("create %q = %d, want 400; body = %q", status, rec.Code, rec.Body.String())
			}
		})
	}

	// Filtering by a projected status stays legal — it is a read of what the
	// last run projected, not a write.
	rec = httptest.NewRecorder()
	s.handleTodos(rec, httptest.NewRequest(http.MethodGet, "/api/todos?status=in_progress", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("list?status=in_progress = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
}

func TestTodoAPILinks(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	create := func(title string) todoSummary {
		t.Helper()
		rec := httptest.NewRecorder()
		body := `{"title":` + strconvQuote(title) + `,"body":"body","status":"pending"}`
		s.handleTodos(rec, httptest.NewRequest(http.MethodPost, "/api/todos", strings.NewReader(body)))
		if rec.Code != http.StatusCreated {
			t.Fatalf("create %q = %d; body = %q", title, rec.Code, rec.Body.String())
		}
		var created todoSummary
		if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
			t.Fatalf("unmarshal create: %v", err)
		}
		return created
	}
	blocked := create("Blocked work")
	blocker := create("Blocking work")

	rec := httptest.NewRecorder()
	linkBody := `{"ref":` + strconvQuote(blocked.Ref) + `,"target":` + strconvQuote(blocker.Ref) + `,"relation":"depends-on"}`
	s.handleTodoLinks(rec, httptest.NewRequest(http.MethodPost, "/api/todos/links", strings.NewReader(linkBody)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("link = %d, want 201; body = %q", rec.Code, rec.Body.String())
	}
	var created todos.Link
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal link: %v", err)
	}
	if created.Relation != types.RelationDependsOn || created.TargetTitle != "Blocking work" {
		t.Fatalf("unexpected link: %+v", created)
	}

	rec = httptest.NewRecorder()
	s.handleTodoLinks(rec, httptest.NewRequest(http.MethodGet, "/api/todos/links?ref="+url.QueryEscape(blocked.Ref), nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("list links = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var listed todoLinksResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("unmarshal links: %v", err)
	}
	if len(listed.Links) != 1 || listed.Links[0].Relation != types.RelationDependsOn {
		t.Fatalf("unexpected links for the blocked TODO: %+v", listed.Links)
	}

	// The blocker sees the same edge as the derived read-only blocks relation.
	rec = httptest.NewRecorder()
	s.handleTodoLinks(rec, httptest.NewRequest(http.MethodGet, "/api/todos/links?ref="+url.QueryEscape(blocker.Ref), nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("list reverse links = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("unmarshal reverse links: %v", err)
	}
	if len(listed.Links) != 1 || listed.Links[0].Relation != types.RelationBlocks {
		t.Fatalf("unexpected links for the blocking TODO: %+v", listed.Links)
	}

	// blocks is derived, so it cannot be written.
	rec = httptest.NewRecorder()
	blocksBody := `{"ref":` + strconvQuote(blocker.Ref) + `,"target":` + strconvQuote(blocked.Ref) + `,"relation":"blocks"}`
	s.handleTodoLinks(rec, httptest.NewRequest(http.MethodPost, "/api/todos/links", strings.NewReader(blocksBody)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("link blocks = %d, want 400; body = %q", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	s.handleTodoLinks(rec, httptest.NewRequest(http.MethodDelete, "/api/todos/links", strings.NewReader(linkBody)))
	if rec.Code != http.StatusOK {
		t.Fatalf("unlink = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil {
		t.Fatalf("unmarshal unlink: %v", err)
	}
	if len(listed.Links) != 0 {
		t.Fatalf("links remained after unlink: %+v", listed.Links)
	}
}

// A link write with a key the payload does not declare is rejected before it
// resolves a target or a linker, and the error names the offending key.
func TestTodoAPILinkCreateRejectsUnknownField(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	rec := httptest.NewRecorder()
	body := `{"ref":"a","target":"b","relation":"depends-on","bogus":true}`
	s.handleTodoLinks(rec, httptest.NewRequest(http.MethodPost, "/api/todos/links", strings.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("link status = %d, want %d; body = %q", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "bogus") {
		t.Fatalf("body does not name the offending key: %q", rec.Body.String())
	}
}

// A label write with a key the payload does not declare is rejected before it
// resolves a label store, and the error names the offending key.
func TestTodoAPILabelSetRejectsUnknownField(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	rec := httptest.NewRecorder()
	body := `{"name":"bug","bogus":true}`
	s.handleTodoLabels(rec, httptest.NewRequest(http.MethodPost, "/api/todos/labels", strings.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("label status = %d, want %d; body = %q", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "bogus") {
		t.Fatalf("body does not name the offending key: %q", rec.Body.String())
	}
}

func TestTodoAPIGetResolvesSessionUUIDAndPreservesExactSession(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	provider := uiTestProviderFor(workDir)
	created, err := provider.Create(t.Context(), todos.CreateRequest{
		Title:  "Open me from my session",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("create todo: %v", err)
	}
	sessionID := "019f5b29-7890-7c11-8e7a-838e5d373e39"
	if err := provider.UpdateState(t.Context(), created, todos.StateUpdate{SessionID: &sessionID}); err != nil {
		t.Fatalf("record session: %v", err)
	}

	rec := httptest.NewRecorder()
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodGet, "/api/todos/item?ref="+url.QueryEscape(sessionID), nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("get by session status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var detail todoSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &detail); err != nil {
		t.Fatalf("unmarshal detail: %v", err)
	}
	if detail.Ref != created.ID || detail.LookupSessionID != sessionID {
		t.Fatalf("session lookup = %+v, want todo %q and session %q", detail, created.ID, sessionID)
	}
}

// The list response must expose hasPlan/hasVerification on every item (not
// just detail responses) so the todo row can render its plan/verification
// indicators without a round-trip per row — see HasPlan (todos/plans.go) and
// ExtractVerificationFixture (todos/verification_fixture.go).
func TestTodoAPIListExposesHasPlanAndHasVerification(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	provider := uiTestProviderFor(workDir)

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
	if err := provider.UpdateState(t.Context(), awaitingReview, todos.StateUpdate{Status: &reviewStatus, PlanPath: &planPath}); err != nil {
		t.Fatalf("mark awaiting review: %v", err)
	}

	rec := httptest.NewRecorder()
	s.handleTodos(rec, httptest.NewRequest(http.MethodGet, "/api/todos", nil))
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
	req := httptest.NewRequest(http.MethodPost, "/todos/new?dir="+url.QueryEscape(workDir)+"&title=Draft+from+query&priority=low", nil)
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
	req := httptest.NewRequest(http.MethodPost, "/api/todos/new?dir="+url.QueryEscape(workDir), strings.NewReader(body))
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
	req := httptest.NewRequest(http.MethodPost, "/api/todos/new?dir="+url.QueryEscape(workDir), bytes.NewReader(raw))
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
	req := httptest.NewRequest(http.MethodPost, "/api/todos/new?dir="+url.QueryEscape(workDir), bytes.NewReader(raw))
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
	req := httptest.NewRequest(http.MethodPost, "/api/todos/new?dir="+url.QueryEscape(workDir), &body)
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

func TestTodoAPIRunStartsSelectedTodo(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:    "Run me",
		Priority: types.PriorityMedium,
		Status:   types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	oldStart := run.Start
	var got todoRunRequest
	run.Start = func(req todoRunRequest) (todoRunStartResult, error) {
		got = req
		return todoRunStartResult{Status: "started", SessionID: "11111111-1111-4111-8111-111111111111"}, nil
	}
	t.Cleanup(func() { run.Start = oldStart })

	body, _ := json.Marshal(todoRunPayload{
		Ref: todos.TODOReference(created),
		Spec: api.Spec{
			Model:  api.Model{Name: "codex", Mode: api.ModeCmux, Effort: "high"},
			Budget: api.Budget{Cost: 1.25, MaxTurns: 12, Timeout: "45m"},
			// Dirty-worktree now rides the spec's
			// Setup.Checkout.Worktree.Uncommitted (Workspace section) instead of a
			// sibling flag.
			Setup: &shell.Setup{Checkout: &shell.Checkout{
				Worktree: &shell.Worktree{Mode: shell.WorktreeNew, Uncommitted: shell.CloneClone},
			}},
		},
	})
	rec := httptest.NewRecorder()
	s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("run status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var resp todoRunResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal run response: %v", err)
	}
	if resp.Status != "started" || resp.Provider != "openai" || resp.RuntimeMode != "cmux" {
		t.Fatalf("unexpected run response: %+v", resp)
	}
	if resp.Count != 1 {
		t.Fatalf("run count = %d, want 1", resp.Count)
	}
	if got.Todo == nil || got.Todo.Title != "Run me" {
		t.Fatalf("run starter did not receive selected todo: %+v", got.Todo)
	}
	if got.Dir != workDir {
		t.Fatalf("unexpected run source: dir=%q", got.Dir)
	}
	// The request layer carries exactly what the dialog sent; the fold — model
	// expansion, the lifecycle step's own layers — happens when the run resolves.
	request := got.Options.Request
	if request.Name != "codex" || request.Mode != captainai.ModeCmux || request.Effort != "high" || request.Budget.Cost != 1.25 || request.Budget.MaxTurns != 12 || !specDirty(request) {
		t.Fatalf("unexpected request layer: %+v", got.Options)
	}
	if got.Options.Host != lifecycle.HostDashboard || got.Options.Step != "" {
		t.Fatalf("run options = %+v, want the dashboard host and the lifecycle's own step choice", got.Options)
	}
	if resp.Step == "" || resp.Model != resolvedName(t, api.Model{Name: "codex", Mode: captainai.ModeCmux}) {
		t.Fatalf("response did not report the resolved step and model: %+v", resp)
	}
}

// dashboardRunSpec resolves a dashboard payload for a fresh todo in dir the way
// /api/todos/run does — wire validation, then the lifecycle host's fold — and
// returns the spec captain would be handed.
func dashboardRunSpec(t *testing.T, dir string, payload todoRunPayload) (api.Spec, error) {
	t.Helper()
	provider := uiTestProviderFor(dir)
	todo, err := provider.Create(t.Context(), todos.CreateRequest{Title: "Resolvable", Status: types.StatusPending})
	if err != nil {
		t.Fatalf("seed todo: %v", err)
	}
	opts, err := buildTodoRunOptions(payload, nil)
	if err != nil {
		return api.Spec{}, err
	}
	prepared, err := run.Resolve(t.Context(), todoRunRequest{Provider: provider, Todo: todo, Dir: dir, Options: opts})
	if err != nil {
		return api.Spec{}, err
	}
	return prepared.Resolution.Spec, nil
}

func TestTodoAPIRunPreviewReturnsPrompt(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
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
		s.handleTodoRunPreview(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run/preview", strings.NewReader(string(body))))
		if rec.Code != http.StatusOK {
			t.Fatalf("preview status = %d, want 200; body = %q", rec.Code, rec.Body.String())
		}
		var resp todoRunPreviewResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal preview response: %v", err)
		}
		return resp
	}

	cmuxResp := preview(todoRunPayload{Ref: ref, Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeCmux, Effort: "high"}}})
	if cmuxResp.Count != 1 || cmuxResp.Provider != "anthropic" || cmuxResp.RuntimeMode != "cmux" {
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

	planResp := preview(todoRunPayload{Ref: ref, Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeCmux, Effort: "medium"}}, Step: "plan"})
	if !strings.Contains(planResp.Prompt, "## Fix the parser") {
		t.Fatalf("plan preview should contain the title: %q", planResp.Prompt)
	}
	if planResp.Step != "plan" {
		t.Fatalf("plan preview step = %q, want plan", planResp.Step)
	}

	agentResp := preview(todoRunPayload{Ref: ref, Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeAgent, Effort: "medium"}}})
	if strings.HasPrefix(agentResp.Prompt, "# Fix the parser") {
		t.Fatalf("agent preview should be the bare claude prompt, not the cmux instruction: %q", agentResp.Prompt)
	}
	if !strings.Contains(agentResp.Prompt, "The parser drops trailing commas.") {
		t.Fatalf("agent preview should include the body: %q", agentResp.Prompt)
	}
}

func TestTodoRunPreviewAbsolutizesAttachmentURLs(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:  "Screenshot todo",
		Body:   "See the bug:\n\n![screen.png](" + attachmentURLPrefix + "abc.png)",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}
	ref := todos.TODOReference(created)

	body, _ := json.Marshal(todoRunPayload{Ref: ref, Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeAgent}}})
	rec := httptest.NewRecorder()
	s.handleTodoRunPreview(rec, httptest.NewRequest(http.MethodPost, "http://gavel.example:9092/api/todos/run/preview", strings.NewReader(string(body))))
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

// isolatedTodoWorkspace returns a workspace whose .gavel.yaml layers are empty
// apart from a model, so a resolution test reads only what it declares. The home
// layer is already redirected package-wide by TestMain.
//
// The model is not optional any more: gavel has no built-in default, so a run
// resolved from a genuinely empty workspace fails with "model name is required"
// before it reaches whatever the test is actually about. Declaring it here keeps
// these tests about run options rather than about configuration.
func isolatedTodoWorkspace(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("ai:\n  model: agent:claude-sonnet-5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// Auto-commit is sourced from Workflow.Commits. The lifecycle's run step
// declares one, so a default dashboard run commits; a payload stanza replaces
// it; and an empty list cannot clear it — captain's merge reads an empty slice
// as "not stated", which is why the CLI refuses `--commit=false` outright
// rather than pretending the request layer removed anything.
func TestDashboardRunCommitFromWorkflow(t *testing.T) {
	dir := isolatedTodoWorkspace(t)
	// The run step is named: a pending todo with no plan would otherwise be
	// planned first, and a plan never commits.
	base := todoRunPayload{Step: "run", Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeCmux, Effort: "medium"}}}

	cases := []struct {
		name     string
		workflow *api.Workflow
		want     []api.Commit
	}{
		{"no workflow inherits the lifecycle step's commit", nil, []api.Commit{{On: api.CommitOnRun, Stage: "worktree", Gates: api.CommitGatesFull}}},
		{"a commit policy replaces it", &api.Workflow{Commits: []api.Commit{{On: api.CommitOnRun, Gates: api.CommitGatesCheap}}}, []api.Commit{{On: api.CommitOnRun, Gates: api.CommitGatesCheap}}},
		{"an empty commit list leaves the step's commit", &api.Workflow{Commits: []api.Commit{}}, []api.Commit{{On: api.CommitOnRun, Stage: "worktree", Gates: api.CommitGatesFull}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := base
			payload.Spec.Workflow = tc.workflow
			spec, err := dashboardRunSpec(t, dir, payload)
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}
			if !reflect.DeepEqual(run.Commits(spec), tc.want) {
				t.Fatalf("commits = %+v, want %+v", run.Commits(spec), tc.want)
			}
		})
	}
}

func TestDashboardRunToolPreferences(t *testing.T) {
	dir := isolatedTodoWorkspace(t)
	base := todoRunPayload{Step: "run", Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeCmux, Effort: "medium"}}}

	t.Run("valid prefs and permission mode are threaded", func(t *testing.T) {
		payload := base
		payload.Spec.Permissions = api.Permissions{
			Mode: api.PermissionAcceptEdits,
			Tools: api.Tools{
				"Bash": api.ToolPolicyAsk, "Write": api.ToolPolicyDeny, "Read": api.ToolPolicyAuto, "Glob": api.ToolPolicyAuto,
			},
		}
		spec, err := dashboardRunSpec(t, dir, payload)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if spec.Permissions.Mode != api.PermissionAcceptEdits {
			t.Fatalf("PermissionMode = %q, want acceptEdits", spec.Permissions.Mode)
		}
		policies := spec.Permissions.Tools.Policies()
		if policies["Bash"] != api.ToolPolicyAsk || policies["Write"] != api.ToolPolicyDeny || policies["Read"] != api.ToolPolicyAuto {
			t.Fatalf("tool policies = %v, want Bash=ask Write=deny Read=auto", policies)
		}
		// auto is carried rather than dropped now that Tools is the policy map.
		// It means the same thing either way — an absent tool inherits the posture
		// and auto defers to it — but the map now says what was configured.
		if policies["Glob"] != api.ToolPolicyAuto {
			t.Fatalf("tool policies = %v, want Glob carried as auto", policies)
		}
	})

	t.Run("empty prefs resolve to the dashboard posture", func(t *testing.T) {
		spec, err := dashboardRunSpec(t, dir, base)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		// The dashboard resolves as the approval-serving host, so a payload that
		// states no posture still leaves with permissions.mode: default — the mode
		// that makes the broker the thing a tool call is checked against. The
		// prompt's own preset may allow the read-only tools; nothing here polices
		// Bash, which is what the broker exists to ask about.
		if policy := spec.Permissions.Tools.Policies()["Bash"]; policy != "" || spec.Permissions.Mode != api.PermissionDefault {
			t.Fatalf("want an unpoliced Bash and the dashboard posture, got %v / %q", spec.Permissions.Tools, spec.Permissions.Mode)
		}
	})

	t.Run("invalid tool policy fails at the wire", func(t *testing.T) {
		payload := base
		payload.Spec.Permissions = api.Permissions{Tools: api.Tools{"Bash": "maybe"}}
		if _, err := buildTodoRunOptions(payload, nil); err == nil {
			t.Fatal("expected error for invalid tool policy, got nil")
		}
	})

	t.Run("invalid permission mode fails at the wire", func(t *testing.T) {
		payload := base
		payload.Spec.Permissions = api.Permissions{Mode: "yolo"}
		if _, err := buildTodoRunOptions(payload, nil); err == nil {
			t.Fatal("expected error for invalid permission mode, got nil")
		}
	})
}

// api.Spec declares a value-receiver MarshalJSON; when todoRunPayload embedded it
// that method was promoted onto the payload, so marshaling emitted a bare spec and
// every sibling field vanished — silently emptying the review API's `options`
// object and stripping the ref off every request body a test builds. Spec is now a
// named field under its own `spec` key, and this asserts the whole payload survives
// the wire in both directions.
func TestTodoRunPayloadRoundTripsSpecAndSiblings(t *testing.T) {
	payload := todoRunPayload{
		Dir:    "/repos/gavel",
		Ref:    "todo-1",
		Refs:   []string{"todo-1", "todo-2"},
		Step:   "plan",
		Resume: true,
		Force:  true,
		Spec: api.Spec{
			Model:  api.Model{Name: "claude", Mode: api.ModeCmux, Effort: "medium", Fallbacks: api.ModelList{{Name: "claude-sonnet-5"}}},
			Prompt: api.Prompt{User: "Implement the reviewed plan.", System: "Keep the patch narrow."},
			Budget: api.Budget{Cost: 2.5, MaxTurns: 8, Timeout: "20m"},
			Memory: api.Memory{Skills: []string{"gavel-todos"}},
			Permissions: api.Permissions{
				Mode:    api.PermissionAcceptEdits,
				Tools:   api.Tools{"Bash": api.ToolPolicyAsk},
				MCP:     api.MCP{Servers: []string{"postgres"}},
				Plugins: api.ResourcePolicies{"review": api.ResourceEnabled},
				Skills:  api.ResourcePolicies{"gavel-todos": api.ResourceEnabled},
			},
			Setup:     &shell.Setup{Cwd: "workspace", Checkout: &shell.Checkout{Mode: shell.CheckoutLocal, Path: ".", Worktree: &shell.Worktree{Mode: shell.WorktreeNew, Prefix: "todo"}}},
			Workflow:  &api.Workflow{Verify: &api.Verify{Commands: []string{"go test ./todos"}, Scope: api.VerifyScopeChanged, MaxIterations: 3}, Commits: []api.Commit{{On: api.CommitOnRun, Gates: api.CommitGatesFull}}},
			SessionID: "sess-1",
			CLIArgs:   map[string]any{"fullAuto": true},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	var got todoRunPayload
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if !reflect.DeepEqual(payload, got) {
		t.Fatalf("payload did not round-trip\n got: %+v\nwant: %+v\nwire: %s", got, payload, body)
	}
}

// The response's Commit reports the RESOLVED run: a payload stanza commits, and
// so does a payload that says nothing, because the lifecycle's run step declares
// a commit of its own.
func TestTodoAPIRunThreadsCommitOption(t *testing.T) {
	cases := []struct {
		name     string
		workflow *api.Workflow
		want     bool
	}{
		{"a commit policy auto-commits", &api.Workflow{Commits: []api.Commit{{On: api.CommitOnRun}}}, true},
		{"absent workflow inherits the run step's commit", nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			workDir := t.TempDir()
			s := &Server{ghOpts: github.Options{WorkDir: workDir}}
			created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
				Title:  "Run me",
				Status: types.StatusPending,
			})
			if err != nil {
				t.Fatalf("seed create: %v", err)
			}

			oldStart := run.Start
			var got todoRunRequest
			run.Start = func(req todoRunRequest) (todoRunStartResult, error) {
				got = req
				return todoRunStartResult{Status: "started", SessionID: "11111111-1111-4111-8111-111111111111"}, nil
			}
			t.Cleanup(func() { run.Start = oldStart })

			payload := todoRunPayload{
				Ref:  todos.TODOReference(created),
				Step: "run",
				Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeCmux, Effort: "medium"}},
			}
			payload.Spec.Workflow = tc.workflow
			body, _ := json.Marshal(payload)
			rec := httptest.NewRecorder()
			s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(string(body))))
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
			if got := specCommit(resolvedRun(t, got).Spec); got != tc.want {
				t.Fatalf("run starter commit = %v, want %v", got, tc.want)
			}
		})
	}
}

// A dry run (every Workflow.Commits stanza marked dryRun) still executes the
// agent — it only reports what it would commit — so handleTodoRun starts it and
// reports Commit:false.
func TestTodoAPIRunDryRunStartsButSkipsCommit(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:  "Run me",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	oldStart := run.Start
	started := false
	run.Start = func(todoRunRequest) (todoRunStartResult, error) {
		started = true
		return todoRunStartResult{Status: "started", SessionID: "11111111-1111-4111-8111-111111111111"}, nil
	}
	t.Cleanup(func() { run.Start = oldStart })

	payload := todoRunPayload{
		Ref:  todos.TODOReference(created),
		Step: "run",
		Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeCmux, Effort: "medium"}},
	}
	// A commit policy is declared but marked dryRun: the run executes, the commit
	// is reported rather than cut.
	payload.Spec.Workflow = &api.Workflow{Commits: []api.Commit{{On: api.CommitOnRun, DryRun: true}}}
	body, _ := json.Marshal(payload)
	rec := httptest.NewRecorder()
	s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(string(body))))
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

func TestTodoAPIRunRejectsMultipleTodos(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	provider := uiTestProviderFor(workDir)
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

	oldStart := run.Start
	run.Start = func(todoRunRequest) (todoRunStartResult, error) {
		t.Fatal("grouped native run must be rejected before dispatch")
		return todoRunStartResult{}, nil
	}
	t.Cleanup(func() { run.Start = oldStart })

	body, _ := json.Marshal(todoRunPayload{
		Refs: []string{todos.TODOReference(first), todos.TODOReference(second), todos.TODOReference(first)},
		Spec: api.Spec{Model: api.Model{Name: "sonnet", Mode: api.ModeCmux, Effort: "medium"}},
	})
	rec := httptest.NewRecorder()
	s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(string(body))))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("run status = %d, want 400; body = %q", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "one issue at a time") {
		t.Fatalf("unexpected grouped-run error: %q", rec.Body.String())
	}
}

func TestTodoAPIRunRejectsLegacyCompositeDriver(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:  "Run me",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	// The retired keys are rejected by name so the client learns what replaced
	// them; any other unknown key is refused as unknown.
	for _, tc := range []struct{ field, body string }{
		{"driver", `{"ref":%q,"driver":"codex-headless","spec":{"model":"codex"}}`},
		{"runMode", `{"ref":%q,"runMode":"plan"}`},
		{"prompt", `{"ref":%q,"prompt":"triage"}`},
		{"plan", `{"ref":%q,"plan":true}`},
	} {
		body := fmt.Sprintf(tc.body, todos.TODOReference(created))
		rec := httptest.NewRecorder()
		s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(body)))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: run status = %d, want 400; body = %q", tc.field, rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), `\"`+tc.field+`\" is not supported`) || !strings.Contains(rec.Body.String(), `\"step\"`) {
			t.Fatalf("%s: error must name the retired field and its replacement: %q", tc.field, rec.Body.String())
		}
	}
	rec := httptest.NewRecorder()
	s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run",
		strings.NewReader(fmt.Sprintf(`{"ref":%q,"unexpected":1}`, todos.TODOReference(created)))))
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "unknown field") {
		t.Fatalf("unknown key: status = %d, body = %q, want a 400 naming the unknown field", rec.Code, rec.Body.String())
	}
}

func TestTodoAPIRunPlanThreadsPlanOption(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:  "Plan me",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	oldStart := run.Start
	var got todoRunRequest
	run.Start = func(req todoRunRequest) (todoRunStartResult, error) {
		got = req
		return todoRunStartResult{Status: "started", SessionID: "11111111-1111-4111-8111-111111111111"}, nil
	}
	t.Cleanup(func() { run.Start = oldStart })

	body, _ := json.Marshal(todoRunPayload{
		Ref:  todos.TODOReference(created),
		Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeCmux, Effort: "medium"}},
		Step: "plan",
	})
	rec := httptest.NewRecorder()
	s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("run status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var resp todoRunResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal run response: %v", err)
	}
	if resp.Step != "plan" || !strings.Contains(resp.Reason, "step") {
		t.Fatalf("response did not echo the requested step and why: %+v", resp)
	}
	if got.Options.Step != "plan" {
		t.Fatalf("run starter did not receive the plan step: %+v", got.Options)
	}
}

// The step is the whole vocabulary: a plan step runs on every runtime (the plan
// prompt's frontmatter carries the plan posture), verify is a step like any
// other and is refused only when the todo has nothing to verify, and an unknown
// step is refused by name with the lifecycle's own steps listed.
func TestTodoAPIRunStepField(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:  "Plan me",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}
	got, _ := stubRunStart(t)

	post := func(payload todoRunPayload) *httptest.ResponseRecorder {
		t.Helper()
		body, _ := json.Marshal(payload)
		rec := httptest.NewRecorder()
		s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(string(body))))
		return rec
	}

	rec := post(todoRunPayload{
		Ref:  todos.TODOReference(created),
		Spec: api.Spec{Model: api.Model{Name: "claude", Mode: api.ModeAgent, Effort: "medium"}},
		Step: "plan",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("run status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if got.Options.Step != "plan" {
		t.Fatalf("step not threaded: %+v", got.Options)
	}

	rec = post(todoRunPayload{Ref: todos.TODOReference(created), Step: "verify"})
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "definition of done") {
		t.Fatalf("verify on a todo with nothing to verify: status = %d, body = %q, want a 400 naming the missing definition of done", rec.Code, rec.Body.String())
	}

	rec = post(todoRunPayload{Ref: todos.TODOReference(created), Step: "shape-it"})
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "shape-it") || !strings.Contains(rec.Body.String(), "plan") {
		t.Fatalf("unknown step: status = %d, body = %q, want a 400 naming the step and listing the lifecycle's", rec.Code, rec.Body.String())
	}
}

func TestDashboardRunRuntimeFromSpec(t *testing.T) {
	dir := isolatedTodoWorkspace(t)
	spec, err := dashboardRunSpec(t, dir, todoRunPayload{Spec: api.Spec{Model: api.Model{Mode: api.ModeAgent, Effort: "medium"}}})
	if err != nil {
		t.Fatalf("agent runtime: %v", err)
	}
	if spec.Mode != api.ModeAgent {
		t.Fatalf("got mode=%q", spec.Mode)
	}

	spec, err = dashboardRunSpec(t, dir, todoRunPayload{Spec: api.Spec{Model: api.Model{Name: "codex", Mode: api.ModeCmux}}})
	if err != nil {
		t.Fatalf("codex cmux runtime: %v", err)
	}
	if providerKey(spec.Model) != "openai" || spec.Name != resolvedName(t, api.Model{Name: "codex", Mode: api.ModeCmux}) || spec.Mode != api.ModeCmux {
		t.Fatalf("got provider=%q model=%q mode=%q", providerKey(spec.Model), spec.Name, spec.Mode)
	}
}

func TestTodoRunContextListsCaptainRuntimeModes(t *testing.T) {
	prev := runCaptainWhoami
	calls := 0
	runCaptainWhoami = func(opts captaincli.WhoamiOptions) (any, error) {
		calls++
		if opts.Mode != "" || opts.Provider != "" || !opts.Models {
			t.Fatalf("whoami opts = %+v, want one unfiltered model snapshot", opts)
		}
		adapters := make([]captaincli.AdapterStatus, 0, 5)
		for _, runtimeMode := range []string{"cmux", "agent", "cli"} {
			adapters = append(adapters, captaincli.AdapterStatus{
				Provider:      "anthropic",
				Mode:          runtimeMode,
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
		for _, runtimeMode := range []string{"cmux", "agent"} {
			adapters = append(adapters, captaincli.AdapterStatus{
				Provider:      "openai",
				Mode:          runtimeMode,
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
		return captaincli.WhoamiResult{
			Adapters:        adapters,
			Runtimes:        api.RuntimeCatalog(),
			DefaultProvider: "anthropic",
			ProviderDefaults: map[string]captaincli.ProviderDefaultView{
				"anthropic": {Mode: "agent", Model: "claude-sonnet-5"},
				"openai":    {Mode: "agent", Model: "gpt-5.5"},
			},
		}, nil
	}
	t.Cleanup(func() { runCaptainWhoami = prev })

	resp, err := todoRunContext("")
	if err != nil {
		t.Fatalf("todo run context: %v", err)
	}
	if calls != 1 {
		t.Fatalf("captain whoami calls = %d, want 1", calls)
	}
	if !stringSliceContains(resp.Efforts, "xhigh") {
		t.Fatalf("efforts = %v, want captain xhigh effort", resp.Efforts)
	}
	if resp.DefaultMode != "agent" {
		t.Fatalf("default mode = %q, want agent", resp.DefaultMode)
	}
	modeFor := func(provider, id string) todoRunModeOption {
		t.Helper()
		for _, option := range resp.Modes {
			if option.Provider == provider && option.ID == id {
				return option
			}
		}
		t.Fatalf("missing %s %s runtime in %+v", provider, id, resp.Modes)
		return todoRunModeOption{}
	}
	claudeCmux := modeFor("anthropic", "cmux")
	claudeAgent := modeFor("anthropic", "agent")
	claudeCLI := modeFor("anthropic", "cli")
	codexCmux := modeFor("openai", "cmux")
	codexAgent := modeFor("openai", "agent")
	for _, option := range []todoRunModeOption{claudeCmux, claudeAgent, claudeCLI, codexCmux, codexAgent} {
		if len(option.Models) == 0 {
			t.Fatalf("runtime %s/%s has no model list: %+v", option.Provider, option.ID, option)
		}
	}
	if !todoRunModelsContain(claudeCmux.Models, "claude-sonnet-5") {
		t.Fatalf("claude cmux models = %+v, want claude-sonnet-5 from captain whoami", claudeCmux.Models)
	}
	if !todoRunModelsContain(claudeCmux.Models, "claude-opus-4-8") {
		t.Fatalf("claude cmux models = %+v, want claude-opus-4-8 from captain whoami", claudeCmux.Models)
	}
	if !todoRunModelsContain(claudeAgent.Models, "claude-fable-5") {
		t.Fatalf("claude agent models = %+v, want Fable from captain whoami", claudeAgent.Models)
	}
	if len(claudeCmux.Models) != 3 || claudeCmux.Models[0].ID != "claude-sonnet-5" || claudeCmux.Models[1].ID != "claude-fable-5" || claudeCmux.Models[2].ID != "claude-opus-4-8" {
		t.Fatalf("claude cmux models = %+v, want Captain model-detail order", claudeCmux.Models)
	}
	if todoRunModelsContain(claudeCmux.Models, "claude-agent-sonnet") {
		t.Fatalf("claude cmux models = %+v, should not expose synthetic aliases", claudeCmux.Models)
	}
	if !todoRunModelsContain(codexCmux.Models, "gpt-5.5") {
		t.Fatalf("codex cmux models = %+v, want gpt-5.5 from captain whoami", codexCmux.Models)
	}
	for _, id := range []string{"gpt-5.6-sol", "gpt-5.6-luna", "gpt-5.6-terra"} {
		if !todoRunModelsContain(codexAgent.Models, id) {
			t.Fatalf("codex agent models = %+v, want %s from captain whoami", codexAgent.Models, id)
		}
	}
	sol := todoRunModelByID(codexAgent.Models, "gpt-5.6-sol")
	if !sol.CapabilitiesKnown || !sol.Reasoning || sol.DefaultEffort != "medium" || !stringSliceContains(sol.SupportedEfforts, "ultra") || sol.Temperature == nil || *sol.Temperature {
		t.Fatalf("gpt-5.6-sol capabilities = %+v, want exact effort and temperature metadata", sol)
	}
	if todoRunModelsContain(codexCmux.Models, "gpt-5-codex") {
		t.Fatalf("codex cmux models = %+v, should not include code variant gpt-5-codex", codexCmux.Models)
	}
}

// TestNormalizeTodoRunOptionsCaptainRuntime pins the (provider, mode) each
// payload normalizes to. The provider follows from the model name, so it is
// asserted through providerKey rather than read off the spec.
func TestDashboardRunCaptainRuntime(t *testing.T) {
	dir := isolatedTodoWorkspace(t)
	spec, err := dashboardRunSpec(t, dir, todoRunPayload{Spec: api.Spec{Model: api.Model{Mode: api.ModeCLI, Effort: "xhigh"}}})
	if err != nil {
		t.Fatalf("claude cli runtime: %v", err)
	}
	if spec.Mode != api.ModeCLI || providerKey(spec.Model) != "anthropic" || spec.Effort != "xhigh" {
		t.Fatalf("unexpected claude cli spec: %+v", spec.Model)
	}

	spec, err = dashboardRunSpec(t, dir, todoRunPayload{Spec: api.Spec{Model: api.Model{Name: "codex", Mode: api.ModeAgent}}})
	if err != nil {
		t.Fatalf("default codex runtime: %v", err)
	}
	if spec.Mode != api.ModeAgent || providerKey(spec.Model) != "openai" {
		t.Fatalf("unexpected codex agent defaults: %+v", spec.Model)
	}

	spec, err = dashboardRunSpec(t, dir, todoRunPayload{Spec: api.Spec{Model: api.Model{Mode: api.ModeCLI, Name: "gpt-5.5"}}})
	if err != nil {
		t.Fatalf("the model's mode should override the driver: %v", err)
	}
	if spec.Mode != api.ModeCLI || providerKey(spec.Model) != "openai" {
		t.Fatalf("the model's mode did not take precedence: %+v", spec.Model)
	}

	spec, err = dashboardRunSpec(t, dir, todoRunPayload{Spec: api.Spec{Model: api.Model{Name: "sonnet-4-6", Mode: api.ModeCmux}}})
	if err != nil {
		t.Fatalf("versioned claude cmux model: %v", err)
	}
	if spec.Mode != api.ModeCmux || spec.Name != "claude-sonnet-4-6" {
		t.Fatalf("unexpected versioned claude model spec: %+v", spec.Model)
	}

	spec, err = dashboardRunSpec(t, dir, todoRunPayload{Spec: api.Spec{Model: api.Model{Mode: api.ModeAgent, Name: "opus-4-8"}}})
	if err != nil {
		t.Fatalf("claude agent opus model: %v", err)
	}
	if spec.Mode != api.ModeAgent || spec.Name != "claude-opus-4-8" {
		t.Fatalf("unexpected normalized opus model spec: %+v", spec.Model)
	}

	if _, err := dashboardRunSpec(t, dir, todoRunPayload{Spec: api.Spec{Model: api.Model{Mode: "claude-agent", Name: "claude-sonnet-5"}}}); err == nil {
		t.Fatal("a composite adapter id in the mode field should be rejected")
	}
}

func TestTodoAPIRunThreadsCaptainRuntimeMode(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	created, err := uiTestProviderFor(workDir).Create(t.Context(), todos.CreateRequest{
		Title:  "Run headless",
		Status: types.StatusPending,
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	oldStart := run.Start
	var got todoRunRequest
	run.Start = func(req todoRunRequest) (todoRunStartResult, error) {
		got = req
		return todoRunStartResult{Status: "started", SessionID: "11111111-1111-4111-8111-111111111111"}, nil
	}
	t.Cleanup(func() { run.Start = oldStart })

	body, _ := json.Marshal(todoRunPayload{
		Ref:  todos.TODOReference(created),
		Spec: api.Spec{Model: api.Model{Mode: api.ModeAgent, Name: "claude-sonnet-5", Effort: "medium"}},
	})
	rec := httptest.NewRecorder()
	s.handleTodoRun(rec, httptest.NewRequest(http.MethodPost, "/api/todos/run", strings.NewReader(string(body))))
	if rec.Code != http.StatusOK {
		t.Fatalf("run status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var resp todoRunResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal run response: %v", err)
	}
	if resp.RuntimeMode != "agent" || resp.Model != "claude-sonnet-5" {
		t.Fatalf("response did not thread mode/model: %+v", resp)
	}
	if got.Options.Request.Mode != api.ModeAgent || got.Options.Request.Name != "claude-sonnet-5" {
		t.Fatalf("run starter did not receive mode/model: %+v", got.Options)
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
