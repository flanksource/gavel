package ui

import (
	"os"
	"path/filepath"
	"testing"
)

// These transport fixtures replace unsupported explicit per-tool grants with
// auto policies. This makes no claim about the runtime's read-only isolation.
func configureAutomaticPlanToolPolicies(t testing.TB, dir string) {
	t.Helper()
	config := "todos:\n  plan:\n    permissions:\n      tools: {Read: auto, Glob: auto, Grep: auto}\n  triage:\n    permissions:\n      tools: {Read: auto, Glob: auto, Grep: auto}\n"
	if err := os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte(config), 0o600); err != nil {
		t.Fatalf("configure automatic plan tool policies: %v", err)
	}
}
