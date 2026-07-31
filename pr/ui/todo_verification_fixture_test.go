package ui

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// verificationFixtureBody builds a TODO body carrying the given fixture markdown
// inside a top-level Verification section.
func verificationFixtureBody(description, fixture string) string {
	return description + "\n\n## Verification\n\n" + fixture + "\n"
}

// TestHandleTodoVerificationFixtureSavesSection exercises the Verification
// tab's save path and confirms dedicated fixture markdown survives a re-read.
func TestHandleTodoVerificationFixtureSavesSection(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	createPayload := todoCreatePayload{
		Title:    "Fix parser",
		Body:     verificationFixtureBody("Some description.", "```test\nquery: SELECT 1\n```"),
		Priority: types.PriorityHigh,
	}
	createRaw, err := json.Marshal(createPayload)
	if err != nil {
		t.Fatalf("marshal create payload: %v", err)
	}
	rec := httptest.NewRecorder()
	s.handleTodos(rec, httptest.NewRequest(http.MethodPost, "/api/todos", bytes.NewReader(createRaw)))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %q", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var created todoSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create: %v", err)
	}
	if !strings.Contains(created.VerificationMarkdown, "query: SELECT 1") {
		t.Fatalf("created todo missing seeded verification fixture: %+v", created)
	}

	savePayload := todoVerificationFixturePayload{
		Ref:     created.Ref,
		Fixture: "## Focused query\n\n```test\nquery: SELECT 2\n```\n\n## Assertions\n\n- query passes",
	}
	saveRaw, err := json.Marshal(savePayload)
	if err != nil {
		t.Fatalf("marshal save payload: %v", err)
	}
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/todos/verification/fixture", bytes.NewReader(saveRaw))
	s.handleTodoVerificationFixture(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("save status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var saved todoSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &saved); err != nil {
		t.Fatalf("unmarshal save response: %v", err)
	}
	if !strings.Contains(saved.VerificationMarkdown, "query: SELECT 2") {
		t.Fatalf("saved todo missing updated verification fixture: %+v", saved)
	}
	if strings.Contains(saved.VerificationMarkdown, "SELECT 1") {
		t.Fatalf("saved todo should not retain the prior fixture: %+v", saved)
	}
	if saved.VerificationMarkdown != savePayload.Fixture {
		t.Fatalf("saved verification = %q, want %q", saved.VerificationMarkdown, savePayload.Fixture)
	}
	if !strings.Contains(saved.Body, "Some description.") {
		t.Fatalf("save must preserve unrelated body content: %+v", saved)
	}
	if strings.Contains(saved.Body, "Verification") {
		t.Fatalf("save must keep verification out of body: %+v", saved)
	}

	rec = httptest.NewRecorder()
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodGet, "/api/todos/item?ref="+url.QueryEscape(created.Ref), nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var reread todoSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &reread); err != nil {
		t.Fatalf("unmarshal reread: %v", err)
	}
	if !strings.Contains(reread.VerificationMarkdown, "query: SELECT 2") {
		t.Fatalf("re-read todo missing persisted verification fixture: %+v", reread)
	}
	if reread.VerificationMarkdown != savePayload.Fixture {
		t.Fatalf("re-read verification = %q, want %q", reread.VerificationMarkdown, savePayload.Fixture)
	}
}

// TestHandleTodoVerificationFixtureRequiresRef mirrors the other todo mutation
// handlers' validation: a missing ref is a 400, not a panic or silent no-op.
func TestHandleTodoVerificationFixtureRequiresRef(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/todos/verification/fixture", strings.NewReader(`{"fixture":"x"}`))
	s.handleTodoVerificationFixture(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestHandleTodoVerificationRunExecutesPersistedFixture(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	provider := uiTestProviderFor(workDir)
	created, err := provider.Create(t.Context(), todos.CreateRequest{
		Title: "Run verification fixture",
		Body: verificationFixtureBody("Some description.", `### command: verification smoke

`+"```bash"+`
echo dashboard-verification-ok
`+"```"+`

- contains: dashboard-verification-ok`),
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	payload, err := json.Marshal(todoCriteriaPayload{Ref: todos.TODOReference(created)})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/todos/verification/run", bytes.NewReader(payload))
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("run status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var response struct {
		Verification types.CheckResult `json:"verification"`
		Todo         todoSummary       `json:"todo"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if !response.Verification.AllPassed || response.Verification.Output == nil {
		t.Fatalf("verification did not pass with evidence: %+v", response.Verification)
	}
	if len(response.Verification.Output.Results) != 1 ||
		!strings.Contains(response.Verification.Output.Results[0].Stdout, "dashboard-verification-ok") {
		t.Fatalf("unexpected verification evidence: %+v", response.Verification.Output)
	}
	if response.Todo.Status != types.StatusVerified {
		t.Fatalf("todo status = %q, want verified", response.Todo.Status)
	}
}

// A verification that cannot start creates no prompt run, so the dashboard's
// attempt list can never show it — the response's error field is the only place
// the failure can reach the user.
func TestHandleTodoVerificationRunReportsPreExecutionFailure(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}
	provider := uiTestProviderFor(workDir)
	created, err := provider.Create(t.Context(), todos.CreateRequest{
		Title: "No definition of done",
		Body:  "Nothing to verify here.",
	})
	if err != nil {
		t.Fatalf("seed create: %v", err)
	}

	payload, err := json.Marshal(todoCriteriaPayload{Ref: todos.TODOReference(created)})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/todos/verification/run", bytes.NewReader(payload)))
	if rec.Code != http.StatusOK {
		t.Fatalf("run status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var response struct {
		Verification types.CheckResult `json:"verification"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Verification.AllPassed {
		t.Fatalf("verification without a definition of done must not pass: %+v", response.Verification)
	}
	if !strings.Contains(response.Verification.ErrorText, "no verification fixture") {
		t.Fatalf("verification error = %q, want the missing definition-of-done reason", response.Verification.ErrorText)
	}
}

func TestLegacyTodoVerifyRoutesRemoved(t *testing.T) {
	s := &Server{}
	for _, path := range []string{"/api/todos/verify", "/api/todos/verify/preview"} {
		t.Run(path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`)))
			if rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404; body = %q", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandleTodoVerificationSchemaReturnsProviderDocument(t *testing.T) {
	s := &Server{}
	s.SetFixtureSchemaProvider(func() (any, error) {
		return map[string]any{
			"schemaVersion": 1,
			"fences": map[string]any{
				"test": map[string]any{
					"aliases": []string{"yaml test"},
					"schema":  map[string]any{"type": "object"},
				},
			},
		}, nil
	})

	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/todos/verification/schema", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusOK, rec.Body.String())
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	if got["schemaVersion"].(float64) != 1 {
		t.Fatalf("unexpected schema version: %#v", got["schemaVersion"])
	}
	fences := got["fences"].(map[string]any)
	if _, ok := fences["test"]; !ok {
		t.Fatalf("schema response missing test fence: %#v", got)
	}
}

func TestHandleTodoVerificationSchemaReportsProviderErrors(t *testing.T) {
	s := &Server{}

	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/todos/verification/schema", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("missing provider status = %d, want %d; body = %q", rec.Code, http.StatusServiceUnavailable, rec.Body.String())
	}

	s.SetFixtureSchemaProvider(func() (any, error) {
		return nil, errors.New("schema failed")
	})
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/todos/verification/schema", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("provider error status = %d, want %d; body = %q", rec.Code, http.StatusInternalServerError, rec.Body.String())
	}
}
