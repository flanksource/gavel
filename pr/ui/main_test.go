package ui

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain redirects HOME for the whole package.
//
// Since todo runs resolve their spec through todos/headless, every layer
// verify.LoadGavelConfig reads is an input to these tests — and one of those
// layers is ~/.gavel.yaml. Without this, whoever runs the suite is a hidden
// input to it: a developer whose own config sets `ai.mode` or
// `todos.run.model` gets different assertions than CI, and the failure surfaces
// as an unrelated-looking rejection deep in runtime/model validation.
//
// It is done once here rather than per test because the dependency is a
// property of the package, not of the individual test that happens to trip over
// it — a test added later inherits the isolation instead of having to remember
// it.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "gavel-prui-home")
	if err != nil {
		panic(err)
	}
	os.Setenv("HOME", home)
	// Pin a model in the isolated home layer. gavel has no built-in default any
	// more, so an otherwise-empty home makes every todo run fail with "model name
	// is required" long before it reaches whatever the test is asserting. Naming
	// it here keeps the isolation — the value is fixed and visible — while still
	// giving the resolver the one thing it now insists on.
	if err := os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte(homeModelConfig), 0o600); err != nil {
		panic(err)
	}
	code := m.Run()
	os.RemoveAll(home)
	os.Exit(code)
}

// homeModelConfig is the home layer every test in this package inherits. gavel
// has no built-in default model, so without it a resolved run fails with "model
// name is required" before reaching what the test is actually asserting.
const homeModelConfig = "ai:\n  model: agent:claude-sonnet-5\n"

// writeHomeModel supplies homeModelConfig to a test that replaces TestMain's
// home with one of its own.
func writeHomeModel(t *testing.T, home string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(home, ".gavel.yaml"), []byte(homeModelConfig), 0o600); err != nil {
		t.Fatal(err)
	}
}
