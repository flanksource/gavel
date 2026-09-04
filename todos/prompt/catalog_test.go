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

// The catalog is exactly the three built-ins, so an unknown name must say what IS
// available rather than failing bare.
func TestCatalogUnknownNameEnumeratesAvailable(t *testing.T) {
	_, err := mustCatalog(t, verify.TodosConfig{}).Lookup("nope")
	if err == nil {
		t.Fatal("unknown prompt must error")
	}
	for _, want := range []string{"nope", "run", "plan", "triage"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

// The prompt set is a code contract — each name pairs a behaviour class with the
// envelope the run loop parses against — so configuration re-points the built-ins
// and cannot add a fourth prompt.
func TestCatalogIsExactlyTheBuiltins(t *testing.T) {
	names := mustCatalog(t, verify.TodosConfig{
		Run:    verify.PromptSpec{Spec: api.Spec{Prompt: api.Prompt{User: "implement it"}}},
		Triage: verify.PromptSpec{Spec: api.Spec{Prompt: api.Prompt{User: "triage it"}}},
	}).Names()
	if strings.Join(names, ",") != "run,plan,triage" {
		t.Fatalf("Names() = %v, want run,plan,triage", names)
	}
}

// The typed todos.<name> field re-points a built-in's body without touching the
// envelope the loop parses its result against.
func TestCatalogTypedInlineOverrideReplacesTheBody(t *testing.T) {
	def, err := mustCatalog(t, verify.TodosConfig{
		Triage: verify.PromptSpec{Spec: api.Spec{Prompt: api.Prompt{User: "triage it"}}},
	}).Lookup("triage")
	if err != nil {
		t.Fatalf("Lookup(triage): %v", err)
	}
	template, err := def.Template(t.TempDir())
	if err != nil {
		t.Fatalf("Template: %v", err)
	}
	if template != "triage it" {
		t.Errorf("template = %q, want the configured body", template)
	}
	if def.Origin != ".gavel.yaml todos.triage" {
		t.Errorf("origin = %q, want .gavel.yaml todos.triage", def.Origin)
	}
	if def.Envelope != EnvelopeTriage || def.Class != types.ModePlan {
		t.Errorf("class/envelope = %q/%q, want plan/triage", def.Class, def.Envelope)
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

func TestCatalogListIsLifecycleOrderedNotAlphabetical(t *testing.T) {
	var names []string
	for _, def := range mustCatalog(t, verify.TodosConfig{}).List() {
		names = append(names, def.Name)
	}
	want := []string{"run", "plan", "triage"}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Errorf("List order = %v, want %v", names, want)
	}
}
