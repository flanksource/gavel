package ui

import (
	"os"
	"testing"
)

// TestMain redirects HOME for the whole package.
//
// Since todo runs resolve their spec through todos/spec, every layer
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
	code := m.Run()
	os.RemoveAll(home)
	os.Exit(code)
}
