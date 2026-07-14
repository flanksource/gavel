package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleSettingsSchema(t *testing.T) {
	rec := httptest.NewRecorder()
	(&Server{}).handleSettingsSchema(rec, httptest.NewRequest("GET", "/api/settings/schema", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var schema map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	props, _ := schema["properties"].(map[string]any)
	if _, ok := props["verify"]; !ok {
		t.Errorf("schema is missing the verify section: %v", props)
	}
}

func TestHandleSettingsPrompts_MatchSchema(t *testing.T) {
	// The registry every prompt the UI can edit.
	rec := httptest.NewRecorder()
	(&Server{}).handleSettingsPrompts(rec, httptest.NewRequest("GET", "/api/settings/prompts", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var list []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("prompts is not valid JSON: %v", err)
	}
	registry := map[string]bool{}
	for _, p := range list {
		id, _ := p["id"].(string)
		if id == "" {
			t.Errorf("prompt descriptor missing id: %v", p)
			continue
		}
		registry[id] = true
		if s, _ := p["default"].(string); strings.TrimSpace(s) == "" {
			t.Errorf("prompt %q has an empty default template", id)
		}
		if s, _ := p["title"].(string); s == "" {
			t.Errorf("prompt %q has an empty title", id)
		}
	}

	// Every x-prompt-id stamped on a schema node must have a registry descriptor,
	// otherwise the settings form would render an override field with no default.
	rec = httptest.NewRecorder()
	(&Server{}).handleSettingsSchema(rec, httptest.NewRequest("GET", "/api/settings/schema", nil))
	var schema map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	schemaIDs := map[string]bool{}
	collectPromptIDs(schema, schemaIDs)
	if len(schemaIDs) == 0 {
		t.Fatal("no x-prompt-id nodes found in schema; the cross-check would be vacuous")
	}
	for id := range schemaIDs {
		if !registry[id] {
			t.Errorf("schema x-prompt-id %q has no registry descriptor", id)
		}
	}
	for id := range registry {
		if !schemaIDs[id] {
			t.Errorf("registry prompt %q has no override field in the schema", id)
		}
	}
}

// collectPromptIDs walks a decoded JSON Schema and records every x-prompt-id.
func collectPromptIDs(node any, out map[string]bool) {
	m, ok := node.(map[string]any)
	if !ok {
		return
	}
	if id, ok := m["x-prompt-id"].(string); ok && id != "" {
		out[id] = true
	}
	for _, v := range m {
		collectPromptIDs(v, out)
	}
}

func TestHandleSettingsGavel_ProjectRoundTrip(t *testing.T) {
	dir := withProject(t, "gavel", "flanksource/gavel", "")

	// PUT a config into the project's .gavel.yaml.
	body := `{"verify":{"model":"gemini","promptTemplate":{"inline":"be strict"}}}`
	rec := httptest.NewRecorder()
	(&Server{}).handleSettingsGavel(rec, httptest.NewRequest("PUT", "/api/settings/gavel?project=gavel", strings.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".gavel.yaml")); err != nil {
		t.Fatalf(".gavel.yaml was not written: %v", err)
	}

	// GET it back and confirm the value survived the round trip.
	rec = httptest.NewRecorder()
	(&Server{}).handleSettingsGavel(rec, httptest.NewRequest("GET", "/api/settings/gavel?project=gavel", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var got gavelSettingsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !got.Exists || got.Config.Verify.Model != "gemini" {
		t.Errorf("GET = %+v, want exists with verify.model=gemini", got)
	}
	if text, ok := got.Config.Verify.PromptTemplate.InlineText(); !ok || text != "be strict" {
		t.Errorf("promptTemplate not persisted: %+v", got.Config.Verify.PromptTemplate)
	}
}

func TestHandleSettingsGavel_GlobalMissingReturnsEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	rec := httptest.NewRecorder()
	(&Server{}).handleSettingsGavel(rec, httptest.NewRequest("GET", "/api/settings/gavel?scope=global", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}
	var got gavelSettingsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Exists {
		t.Errorf("global config should not exist yet: %+v", got)
	}
}

func TestHandleSettingsGavelTrace_LayeredProvenance(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// User layer sets verify.model; project layer sets commit + overrides nothing
	// in verify, so the two layers own distinct fields for the provenance badges.
	if err := os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte("verify:\n  model: gemini\n"), 0o644); err != nil {
		t.Fatalf("write home config: %v", err)
	}
	dir := withProject(t, "gavel", "flanksource/gavel", "")
	if err := os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("commit:\n  model: claude-opus\n"), 0o644); err != nil {
		t.Fatalf("write project config: %v", err)
	}

	rec := httptest.NewRecorder()
	(&Server{}).handleSettingsGavelTrace(rec, httptest.NewRequest("GET", "/api/settings/gavel/trace?project=gavel", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %q", rec.Code, rec.Body.String())
	}

	var trace struct {
		Sources []struct {
			Origin string `json:"origin"`
			Config struct {
				Verify struct {
					Model string `json:"model"`
				} `json:"verify"`
				Commit struct {
					Model string `json:"model"`
				} `json:"commit"`
			} `json:"config"`
		} `json:"sources"`
		Merged struct {
			Verify struct {
				Model string `json:"model"`
			} `json:"verify"`
			Commit struct {
				Model string `json:"model"`
			} `json:"commit"`
		} `json:"merged"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &trace); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	byOrigin := map[string]string{}
	for _, s := range trace.Sources {
		if s.Config.Verify.Model != "" {
			byOrigin["verify.model"] = s.Origin
		}
		if s.Config.Commit.Model != "" {
			byOrigin["commit.model"] = s.Origin
		}
	}
	if byOrigin["verify.model"] != "user-home" {
		t.Errorf("verify.model provenance = %q, want user-home", byOrigin["verify.model"])
	}
	if byOrigin["commit.model"] != "target-directory" {
		t.Errorf("commit.model provenance = %q, want target-directory", byOrigin["commit.model"])
	}
	if trace.Merged.Verify.Model != "gemini" || trace.Merged.Commit.Model != "claude-opus" {
		t.Errorf("merged did not combine both layers: %+v", trace.Merged)
	}
}

func TestHandleSettingsGavel_UnknownScopeRejected(t *testing.T) {
	rec := httptest.NewRecorder()
	(&Server{}).handleSettingsGavel(rec, httptest.NewRequest("GET", "/api/settings/gavel?project=ghost", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("unknown project status = %d, want 400", rec.Code)
	}
}
