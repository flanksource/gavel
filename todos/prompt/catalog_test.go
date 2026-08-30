package prompt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
)

func mustCatalog(t *testing.T, cfg verify.TodosConfig) *Catalog {
	t.Helper()
	catalog, err := NewCatalog(cfg)
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	return catalog
}

func TestCatalogBuiltinsCarryClassAndEnvelope(t *testing.T) {
	catalog := mustCatalog(t, verify.TodosConfig{})
	for _, tc := range []struct {
		name         string
		wantClass    types.RunMode
		wantEnvelope EnvelopeKind
	}{
		{"run", types.ModeRun, EnvelopeResult},
		{"plan", types.ModePlan, EnvelopePlan},
		{"triage", types.ModePlan, EnvelopeTriage},
	} {
		def, err := catalog.Lookup(tc.name)
		if err != nil {
			t.Fatalf("Lookup(%s): %v", tc.name, err)
		}
		if def.Class != tc.wantClass {
			t.Errorf("%s class = %q, want %q", tc.name, def.Class, tc.wantClass)
		}
		if def.Envelope != tc.wantEnvelope {
			t.Errorf("%s envelope = %q, want %q", tc.name, def.Envelope, tc.wantEnvelope)
		}
		if def.Origin != "builtin" {
			t.Errorf("%s origin = %q, want builtin", tc.name, def.Origin)
		}
		if strings.TrimSpace(def.Builtin) == "" {
			t.Errorf("%s has no embedded template", tc.name)
		}
	}
}

// Triage shares the plan behaviour class deliberately: it neither commits nor
// runs the definition of done. If it ever became its own class it would need a
// widened step_kind CHECK constraint, which this whole design exists to avoid.
func TestTriageIsPlanClassSoItNeitherCommitsNorVerifies(t *testing.T) {
	def, err := mustCatalog(t, verify.TodosConfig{}).Lookup("triage")
	if err != nil {
		t.Fatalf("Lookup(triage): %v", err)
	}
	if def.Class == types.ModeRun {
		t.Fatal("triage must not be run-class: it would commit the agent's work and run the DoD verifiers")
	}
}

func TestCatalogEmptyNameIsTheRunPrompt(t *testing.T) {
	def, err := mustCatalog(t, verify.TodosConfig{}).Lookup("")
	if err != nil {
		t.Fatalf("Lookup(\"\"): %v", err)
	}
	if def.Name != DefaultName {
		t.Fatalf("empty name resolved to %q, want %q", def.Name, DefaultName)
	}
}

// The available prompt set is project-specific, so an unknown name must say what
// IS available rather than failing bare.
func TestCatalogUnknownNameEnumeratesAvailable(t *testing.T) {
	cfg := verify.TodosConfig{Prompts: map[string]verify.NamedPromptSpec{
		"security": {PromptSpec: verify.PromptSpec{Spec: api.Spec{Prompt: api.Prompt{User: "audit it"}}}},
	}}
	_, err := mustCatalog(t, cfg).Lookup("nope")
	if err == nil {
		t.Fatal("unknown prompt must error")
	}
	for _, want := range []string{"nope", "run", "plan", "triage", "security"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

func TestCatalogCustomPromptDefaultsToPlanClass(t *testing.T) {
	cfg := verify.TodosConfig{Prompts: map[string]verify.NamedPromptSpec{
		"security": {PromptSpec: verify.PromptSpec{Spec: api.Spec{Prompt: api.Prompt{User: "audit it"}}}},
	}}
	def, err := mustCatalog(t, cfg).Lookup("security")
	if err != nil {
		t.Fatalf("Lookup(security): %v", err)
	}
	// A prompt that was never told it may commit must not inherit the right to.
	if def.Class != types.ModePlan {
		t.Errorf("class = %q, want plan by default", def.Class)
	}
	if def.Envelope != EnvelopePlan {
		t.Errorf("envelope = %q, want plan", def.Envelope)
	}
	template, err := def.Template(t.TempDir())
	if err != nil {
		t.Fatalf("Template: %v", err)
	}
	if template != "audit it" {
		t.Errorf("template = %q, want the configured body", template)
	}
}

func TestCatalogCustomPromptHonoursDeclaredClassAndMetadata(t *testing.T) {
	cfg := verify.TodosConfig{Prompts: map[string]verify.NamedPromptSpec{
		"docs": {
			PromptSpec:  verify.PromptSpec{Spec: api.Spec{Prompt: api.Prompt{User: "write docs"}}},
			Class:       "run",
			Title:       "Docs pass",
			Description: "Updates documentation",
		},
	}}
	def, err := mustCatalog(t, cfg).Lookup("docs")
	if err != nil {
		t.Fatalf("Lookup(docs): %v", err)
	}
	if def.Class != types.ModeRun || def.Envelope != EnvelopeResult {
		t.Errorf("class/envelope = %q/%q, want run/result", def.Class, def.Envelope)
	}
	if def.Title != "Docs pass" || def.Description != "Updates documentation" {
		t.Errorf("metadata not carried: %+v", def)
	}
}

// A built-in's envelope is a code contract the run loop parses against, so
// redeclaring its class would make the loop expect a shape the prompt cannot
// produce. That must fail loudly rather than be quietly ignored.
func TestCatalogRejectsRedeclaringABuiltinClass(t *testing.T) {
	cfg := verify.TodosConfig{Prompts: map[string]verify.NamedPromptSpec{
		"triage": {Class: "run", PromptSpec: verify.PromptSpec{Spec: api.Spec{Prompt: api.Prompt{User: "x"}}}},
	}}
	if _, err := NewCatalog(cfg); err == nil {
		t.Fatal("redeclaring a built-in prompt's class must error")
	}
}

func TestCatalogRejectsAPromptWithNoTemplate(t *testing.T) {
	cfg := verify.TodosConfig{Prompts: map[string]verify.NamedPromptSpec{
		"empty": {Title: "Nothing"},
	}}
	def, err := mustCatalog(t, cfg).Lookup("empty")
	if err != nil {
		t.Fatalf("Lookup(empty): %v", err)
	}
	if _, err := def.Template(t.TempDir()); err == nil {
		t.Fatal("a prompt with no file and no body must error rather than render an empty prompt")
	}
}

// The typed todos.<name> field is the ergonomic override for a built-in; a file:
// there replaces the embedded template and is reported as the origin.
func TestCatalogTypedOverrideReplacesBuiltinTemplate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "triage.prompt")
	if err := os.WriteFile(path, []byte("---\nmodel: claude\n---\ncustom triage"), 0o600); err != nil {
		t.Fatalf("write override: %v", err)
	}
	def, err := mustCatalog(t, verify.TodosConfig{Triage: verify.PromptSpec{File: path}}).Lookup("triage")
	if err != nil {
		t.Fatalf("Lookup(triage): %v", err)
	}
	template, err := def.Template(dir)
	if err != nil {
		t.Fatalf("Template: %v", err)
	}
	if !strings.Contains(template, "custom triage") {
		t.Errorf("template = %q, want the override file's contents", template)
	}
	if def.Origin != path {
		t.Errorf("origin = %q, want %q", def.Origin, path)
	}
	// The envelope is not configurable: a triage override still returns a triage
	// envelope, or the run loop would parse the wrong type.
	if def.Envelope != EnvelopeTriage {
		t.Errorf("envelope = %q, want triage", def.Envelope)
	}
}

func TestCatalogListPutsBuiltinsFirst(t *testing.T) {
	cfg := verify.TodosConfig{Prompts: map[string]verify.NamedPromptSpec{
		"alpha": {PromptSpec: verify.PromptSpec{Spec: api.Spec{Prompt: api.Prompt{User: "a"}}}},
		"zeta":  {PromptSpec: verify.PromptSpec{Spec: api.Spec{Prompt: api.Prompt{User: "z"}}}},
	}}
	var names []string
	for _, def := range mustCatalog(t, cfg).List() {
		names = append(names, def.Name)
	}
	want := []string{"run", "plan", "triage", "alpha", "zeta"}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Errorf("List order = %v, want %v", names, want)
	}
}
