package golangci

import (
	"reflect"
	"testing"

	"github.com/flanksource/gavel/linters"
)

func TestDryRunCommandDefault(t *testing.T) {
	g := NewGolangciLint("/repo")
	g.SetOptions(linters.RunOptions{WorkDir: "/repo", ForceJSON: true})

	cmd, args := g.DryRunCommand()
	if cmd != "golangci-lint" {
		t.Errorf("cmd = %q, want golangci-lint", cmd)
	}
	// Note: `run` itself satisfies hasPathArg (it doesn't start with "-"),
	// so no implicit "." is appended. This mirrors the current Run() behavior.
	want := []string{"run", "--output.json.path=stdout", "--output.text.path=stderr"}
	if !reflect.DeepEqual(args, want) {
		t.Errorf("args = %v, want %v", args, want)
	}
}

func TestDryRunCommandWithFiles(t *testing.T) {
	g := NewGolangciLint("/repo")
	g.SetOptions(linters.RunOptions{
		WorkDir:   "/repo",
		ForceJSON: true,
		Files:     []string{"./pkg/a", "./pkg/b"},
		ExtraArgs: []string{"--new-from-rev=abc123"},
	})

	_, args := g.DryRunCommand()
	want := []string{
		"run",
		"--output.json.path=stdout",
		"--output.text.path=stderr",
		"--new-from-rev=abc123",
		"./pkg/a",
		"./pkg/b",
	}
	if !reflect.DeepEqual(args, want) {
		t.Errorf("args = %v, want %v", args, want)
	}
}
