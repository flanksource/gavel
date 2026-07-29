package ui

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
)

// promptReq builds a request against /api/settings/prompts/{id} with the path
// value populated (the mux would otherwise supply it).
func promptReq(method, id, query, body string) *http.Request {
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, "/api/settings/prompts/"+id+"?"+query, r)
	req.SetPathValue("id", id)
	return req
}

func putBody(t *testing.T, req promptDetailRequest) string {
	t.Helper()
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return string(b)
}

// ptr returns a pointer to v, for the optional pointer fields of
// promptDetailRequest (which distinguish an absent field from an empty value).
func ptr[T any](v T) *T { return &v }

// Every registered prompt's dotted config path must resolve to a settable
// PromptSpec, so a new prompt is editable with no per-id server code.
func TestPromptOverridePtr_AllRegistered(t *testing.T) {
	cfg := verify.GavelConfig{}
	for _, p := range registeredPrompts() {
		ov, err := promptOverridePtr(&cfg, p.ConfigPath)
		if err != nil {
			t.Errorf("%s (%s): %v", p.ID, p.ConfigPath, err)
			continue
		}
		if ov == nil {
			t.Errorf("%s: nil override pointer", p.ID)
		}
	}
}

// The returned pointer writes through to the config field.
func TestPromptOverridePtr_IsSettable(t *testing.T) {
	cfg := verify.GavelConfig{}
	ov, err := promptOverridePtr(&cfg, prompts.CommitMessage)
	if err != nil {
		t.Fatalf("promptOverridePtr: %v", err)
	}
	*ov = verify.PromptSpec{Spec: api.Spec{Prompt: api.Prompt{User: "hi"}}}
	if cfg.Commit.Message.Spec.Prompt.User != "hi" {
		t.Errorf("write did not reach cfg.Commit.Message: %+v", cfg.Commit.Message)
	}
}

func TestPromptOverridePtr_BadPath(t *testing.T) {
	cfg := verify.GavelConfig{}
	if _, err := promptOverridePtr(&cfg, "commit.nope"); err == nil {
		t.Error("expected an error for a non-existent config path")
	}
}

func TestModeledSpecKeys(t *testing.T) {
	keys := modeledSpecKeys()
	for _, want := range []string{"model", "backend", "effort", "prompt", "permissions", "budget", "setup", "cliArgs"} {
		if !keys[want] {
			t.Errorf("modeledSpecKeys missing %q", want)
		}
	}
	// dotprompt-only keys the editor does not model must not be cleared by a merge.
	for _, notWant := range []string{"config", "output", "input"} {
		if keys[notWant] {
			t.Errorf("modeledSpecKeys should not contain unmodeled key %q", notWant)
		}
	}
}

func TestHandleSettingsPromptDetail_GetDefault(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	rec := httptest.NewRecorder()
	(&Server{}).handleSettingsPromptDetail(rec, promptReq("GET", prompts.CommitMessage, "scope=global", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var got promptDetailResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Source != "default" {
		t.Errorf("source = %q, want default", got.Source)
	}
	if strings.TrimSpace(got.Raw) == "" {
		t.Error("default prompt has empty raw text")
	}
}

func TestHandleSettingsPromptDetail_UnknownID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	rec := httptest.NewRecorder()
	(&Server{}).handleSettingsPromptDetail(rec, promptReq("GET", "nope.nope", "scope=global", ""))
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestHandleSettingsPromptDetail_PutInlineRoundTrip(t *testing.T) {
	dir := withProject(t, "gavel", "flanksource/gavel", "")

	body := putBody(t, promptDetailRequest{
		Source: "inline",
		Spec:   ptr(map[string]any{"model": "claude-test"}),
		Body:   ptr("Write a message for {{diff}}."),
	})
	rec := httptest.NewRecorder()
	(&Server{}).handleSettingsPromptDetail(rec, promptReq("PUT", prompts.CommitMessage, "project=gavel", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".gavel.yaml")); err != nil {
		t.Fatalf(".gavel.yaml was not written: %v", err)
	}

	rec = httptest.NewRecorder()
	(&Server{}).handleSettingsPromptDetail(rec, promptReq("GET", prompts.CommitMessage, "project=gavel", ""))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var got promptDetailResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Source != "inline" {
		t.Errorf("source = %q, want inline", got.Source)
	}
	if got.Spec == nil || (*got.Spec)["model"] != "claude-test" {
		t.Errorf("spec.model = %v, want claude-test", got.Spec)
	}
	if got.Body == nil || !strings.Contains(*got.Body, "Write a message") {
		t.Errorf("body not persisted: %v", got.Body)
	}
}

func TestHandleSettingsPromptDetail_PutDefaultClearsOverride(t *testing.T) {
	dir := withProject(t, "gavel", "flanksource/gavel", "")

	body := putBody(t, promptDetailRequest{
		Source: "inline",
		Spec:   ptr(map[string]any{"model": "claude-test"}),
		Body:   ptr("Custom message body."),
	})
	rec := httptest.NewRecorder()
	(&Server{}).handleSettingsPromptDetail(rec, promptReq("PUT", prompts.CommitMessage, "project=gavel", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("inline PUT status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	(&Server{}).handleSettingsPromptDetail(
		rec,
		promptReq("PUT", prompts.CommitMessage, "project=gavel", putBody(t, promptDetailRequest{Source: "default"})),
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("default PUT status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var got promptDetailResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Source != "default" {
		t.Errorf("source = %q, want default", got.Source)
	}
	if strings.Contains(got.Raw, "Custom message body.") {
		t.Errorf("default response still contains old override:\n%s", got.Raw)
	}
	cfg, err := verify.LoadSingleGavelConfig(filepath.Join(dir, ".gavel.yaml"))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if !cfg.Commit.Message.IsEmpty() {
		t.Errorf("prompt override was not cleared: %+v", cfg.Commit.Message)
	}
}

// Editing one modeled key (model) via a structured edit must not drop an
// unmodeled frontmatter block (output.schema) the base document carried. Only a
// file override retains dotprompt-only keys, so this exercises the file source.
func TestHandleSettingsPromptDetail_RoundTripPreservesUnmodeledKey(t *testing.T) {
	withProject(t, "gavel", "flanksource/gavel", "")

	baseRaw := "---\n" +
		"model: old-model\n" +
		"output:\n  schema:\n    type: object\n    properties:\n      title:\n        type: string\n" +
		"---\nWrite {{topic}}.\n"
	body := putBody(t, promptDetailRequest{
		Source:  "file",
		Path:    "commit-message.prompt",
		Spec:    ptr(map[string]any{"model": "new-model"}),
		Body:    ptr("Write {{topic}}."),
		BaseRaw: ptr(baseRaw),
	})
	rec := httptest.NewRecorder()
	(&Server{}).handleSettingsPromptDetail(rec, promptReq("PUT", prompts.CommitMessage, "project=gavel", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var got promptDetailResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Spec == nil || (*got.Spec)["model"] != "new-model" {
		t.Errorf("spec.model = %v, want new-model", got.Spec)
	}
	if !strings.Contains(got.Raw, "output") || !strings.Contains(got.Raw, "schema") {
		t.Errorf("unmodeled output.schema was dropped from the saved prompt:\n%s", got.Raw)
	}
}

// A project-scoped save must not touch the global ~/.gavel.yaml layer.
func TestHandleSettingsPromptDetail_ScopeIsolation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	withProject(t, "gavel", "flanksource/gavel", "")

	body := putBody(t, promptDetailRequest{Source: "inline", Spec: ptr(map[string]any{"model": "proj-model"}), Body: ptr("project body")})
	rec := httptest.NewRecorder()
	(&Server{}).handleSettingsPromptDetail(rec, promptReq("PUT", prompts.CommitMessage, "project=gavel", body))
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}

	if _, err := os.Stat(filepath.Join(home, ".gavel.yaml")); !os.IsNotExist(err) {
		t.Errorf("global .gavel.yaml should not exist after a project PUT (err=%v)", err)
	}
	rec = httptest.NewRecorder()
	(&Server{}).handleSettingsPromptDetail(rec, promptReq("GET", prompts.CommitMessage, "scope=global", ""))
	var global promptDetailResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &global); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if global.Source != "default" {
		t.Errorf("global source = %q, want default (untouched)", global.Source)
	}
}

// An invalid spec (a key the spec does not model) fails validation and writes
// nothing.
func TestHandleSettingsPromptDetail_InvalidSpecRejected(t *testing.T) {
	dir := withProject(t, "gavel", "flanksource/gavel", "")

	body := putBody(t, promptDetailRequest{Source: "inline", Spec: ptr(map[string]any{"bogusKey": "x"}), Body: ptr("body")})
	rec := httptest.NewRecorder()
	(&Server{}).handleSettingsPromptDetail(rec, promptReq("PUT", prompts.CommitMessage, "project=gavel", body))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %q", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".gavel.yaml")); !os.IsNotExist(err) {
		t.Errorf(".gavel.yaml must not be written when validation fails (err=%v)", err)
	}
}

// The route resolves through the real mux — the {id} path value is populated by
// routing, not a test-only SetPathValue.
func TestHandleSettingsPromptDetail_ThroughMux(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	rec := httptest.NewRecorder()
	(&Server{}).Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/api/settings/prompts/"+prompts.CommitMessage+"?scope=global", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var got promptDetailResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ID != prompts.CommitMessage {
		t.Errorf("id = %q, want %q (path value not routed?)", got.ID, prompts.CommitMessage)
	}
	if got.Source != "default" {
		t.Errorf("source = %q, want default", got.Source)
	}
}
