package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
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

func TestTodoAPICreateWarnsUnknownField(t *testing.T) {
	warnings := captureTodoRequestWarnings(t)
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	rec := httptest.NewRecorder()
	body := `{"title":"Fix workspace","bogus":true}`
	s.handleTodos(rec, httptest.NewRequest(http.MethodPost, "/api/todos", strings.NewReader(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %q", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if !strings.Contains(warnings.String(), "POST /api/todos request body: unknown field bogus") {
		t.Fatalf("missing request warning: %q", warnings.String())
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

func TestTodoAPILinkCreateWarnsUnknownField(t *testing.T) {
	warnings := captureTodoRequestWarnings(t)
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	provider := uiTestProviderFor(workDir)
	from, err := provider.Create(t.Context(), todos.CreateRequest{Title: "Blocked work", Status: types.StatusPending})
	if err != nil {
		t.Fatal(err)
	}
	to, err := provider.Create(t.Context(), todos.CreateRequest{Title: "Dependency", Status: types.StatusPending})
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	body := fmt.Sprintf(`{"ref":%q,"target":%q,"relation":"depends-on","bogus":true}`, from.ID, to.ID)
	s.handleTodoLinks(rec, httptest.NewRequest(http.MethodPost, "/api/todos/links", strings.NewReader(body)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("link status = %d, want %d; body = %q", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if !strings.Contains(warnings.String(), "POST /api/todos/links request body: unknown field bogus") {
		t.Fatalf("missing request warning: %q", warnings.String())
	}
}

func TestTodoAPILabelSetWarnsUnknownFieldBeforeProviderValidation(t *testing.T) {
	warnings := captureTodoRequestWarnings(t)
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	rec := httptest.NewRecorder()
	body := `{"name":"bug","bogus":true}`
	s.handleTodoLabels(rec, httptest.NewRequest(http.MethodPost, "/api/todos/labels", strings.NewReader(body)))
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "TODO provider does not support label definitions") {
		t.Fatalf("label request should reach provider validation: status = %d; body = %q", rec.Code, rec.Body.String())
	}
	if !strings.Contains(warnings.String(), "POST /api/todos/labels request body: unknown field bogus") {
		t.Fatalf("missing request warning: %q", warnings.String())
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
