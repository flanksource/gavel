package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/captain/pkg/api"
)

// TestLoadTodosDefaultsResolvesModelFallbacks pins the config→effective-model
// wiring: a .gavel.yaml todos.run with a compact model and a fallbacks list
// resolves into todosDefaults whose Expand()+Candidates() put the primary first
// and include the configured fallback. Guards the fallback-drop bug where only
// the model name flowed end to end and the fallback chain was silently discarded.
func TestLoadTodosDefaultsResolvesModelFallbacks(t *testing.T) {
	saved := todosDef
	t.Cleanup(func() { todosDef = saved })

	dir := t.TempDir()
	// Isolate HOME so a real ~/.gavel.yaml cannot contribute todos config.
	t.Setenv("HOME", t.TempDir())

	const cfg = "todos:\n  run:\n    model: \"opus:high\"\n    fallbacks:\n      - sonnet\n"
	if err := os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatalf("write .gavel.yaml: %v", err)
	}

	if err := loadTodosDefaults(dir); err != nil {
		t.Fatalf("loadTodosDefaults: %v", err)
	}
	if todosDef.Model != "opus:high" {
		t.Errorf("todosDef.Model = %q, want opus:high", todosDef.Model)
	}
	if len(todosDef.Fallbacks) != 1 || todosDef.Fallbacks[0].Name != "sonnet" {
		t.Fatalf("todosDef.Fallbacks = %+v, want [sonnet]", todosDef.Fallbacks)
	}

	eff, err := (api.Model{Name: todosDef.Model, Fallbacks: todosDef.Fallbacks}).Expand()
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	if eff.Name != "opus" || eff.Effort != api.EffortHigh {
		t.Fatalf("expanded primary = %+v, want opus/high", eff)
	}
	cands := eff.Candidates()
	if len(cands) != 2 || cands[0].Name != "opus" || cands[1].Name != "sonnet" {
		t.Fatalf("Candidates() = %+v, want [opus sonnet]", cands)
	}
}
