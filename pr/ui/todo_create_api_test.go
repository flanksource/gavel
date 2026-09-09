package ui

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos/types"
)

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
