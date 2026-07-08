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
	"github.com/flanksource/gavel/todos/types"
)

// verificationFixtureBody builds a TODO body carrying the given fixture markdown
// inside a "## Verification" section, matching how ParseTODOContent expects it.
func verificationFixtureBody(description, fixture string) string {
	return description + "\n\n## Verification\n\n" + fixture + "\n"
}

// TestHandleTodoVerificationFixtureSavesSection exercises the Verification
// tab's save path: POST /api/todos/verification/fixture rewrites the todo's
// "## Verification" section in place and the refreshed todo reflects it via
// VerificationMarkdown, mirroring TestTodoAPIFileProviderCRUD's setup.
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
	s.handleTodos(rec, httptest.NewRequest(http.MethodPost, "/api/todos?provider=todos", bytes.NewReader(createRaw)))
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
		Fixture: "```test\nquery: SELECT 2\n```",
	}
	saveRaw, err := json.Marshal(savePayload)
	if err != nil {
		t.Fatalf("marshal save payload: %v", err)
	}
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/todos/verification/fixture?provider=todos", bytes.NewReader(saveRaw))
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
	if !strings.Contains(saved.Body, "Some description.") {
		t.Fatalf("save must preserve unrelated body content: %+v", saved)
	}

	rec = httptest.NewRecorder()
	s.handleTodoItem(rec, httptest.NewRequest(http.MethodGet, "/api/todos/item?provider=todos&ref="+url.QueryEscape(created.Ref), nil))
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
}

// TestHandleTodoVerificationFixtureRequiresRef mirrors the other todo mutation
// handlers' validation: a missing ref is a 400, not a panic or silent no-op.
func TestHandleTodoVerificationFixtureRequiresRef(t *testing.T) {
	workDir := t.TempDir()
	s := &Server{ghOpts: github.Options{WorkDir: workDir}}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/todos/verification/fixture?provider=todos", strings.NewReader(`{"fixture":"x"}`))
	s.handleTodoVerificationFixture(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %q", rec.Code, http.StatusBadRequest, rec.Body.String())
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
