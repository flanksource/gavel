package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/uirplan"
)

func TestUIRSkipRunsUnskippableInventoryTestsAcrossSubprocesses(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake UIR executable is a Python script")
	}
	repo := newUIRSkipFixture(t, false)
	fake := newFakeUIR(t, false)
	export := filepath.Join(repo, "uir-run.json")

	_, err := runTests(testrunner.RunOptions{
		Context: context.Background(), WorkDir: repo, Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
		UIRSkip: true, UIRBinary: fake, UIRExport: export, UIRInputContext: "postgres=test",
	}, false)
	if err != nil {
		t.Fatalf("runTests: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, "known-ran")); !os.IsNotExist(err) {
		t.Fatalf("planner-skippable TestKnown ran: %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(repo, "new-ran")); err != nil || string(data) != "new" {
		t.Fatalf("unskippable TestNew did not run: data=%q err=%v", data, err)
	}
	if _, err := os.Stat(filepath.Join(repo, "imported")); err != nil {
		t.Fatalf("fake UIR import was not invoked: %v", err)
	}
	data, err := os.ReadFile(export)
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Runs []struct {
			SelectionMode string `json:"selection_mode"`
			Complete      bool   `json:"complete"`
			Tests         []struct {
				Name string `json:"name"`
			} `json:"tests"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if len(document.Runs) != 1 || document.Runs[0].SelectionMode != "partial" || !document.Runs[0].Complete {
		t.Fatalf("export metadata = %+v", document.Runs)
	}
	if len(document.Runs[0].Tests) != 1 || document.Runs[0].Tests[0].Name != "TestNew" {
		t.Fatalf("exported tests = %+v", document.Runs[0].Tests)
	}
}

func TestUIRSkipRejectsIdentityMutationBeforeExecution(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake UIR executable is a Python script")
	}
	repo := newUIRSkipFixture(t, false)
	_, err := runTests(testrunner.RunOptions{
		Context: context.Background(), WorkDir: repo, Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
		UIRSkip: true, UIRBinary: newFakeUIR(t, true), UIRExport: filepath.Join(repo, "uir-run.json"), UIRInputContext: "postgres=test",
	}, false)
	if err == nil || !strings.Contains(err.Error(), "indexed inputs changed between planning and execution") {
		t.Fatalf("runTests error = %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(repo, "new-ran")); !os.IsNotExist(statErr) {
		t.Fatalf("tests ran after identity mutation: %v", statErr)
	}
}

func TestUIRSkipRejectsIdentityMutationDuringExecution(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake UIR executable is a Python script")
	}
	repo := newUIRSkipFixture(t, true)
	_, err := runTests(testrunner.RunOptions{
		Context: context.Background(), WorkDir: repo, Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
		UIRSkip: true, UIRBinary: newFakeUIR(t, false), UIRExport: filepath.Join(repo, "uir-run.json"), UIRInputContext: "postgres=test",
	}, false)
	if err == nil || !strings.Contains(err.Error(), "inputs changed during execution; results were not imported") {
		t.Fatalf("runTests error = %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(repo, "new-ran")); statErr != nil {
		t.Fatalf("mutation fixture did not execute: %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(repo, "imported")); !os.IsNotExist(statErr) {
		t.Fatalf("mutated execution was imported: %v", statErr)
	}
}

func TestCompletedUIRSelectionRequiresEveryUnskippableResult(t *testing.T) {
	inventory := []uirplan.Test{
		{ID: "known", Package: "example.com/fixture/pkg", Name: "TestKnown"},
		{ID: "new", Package: "example.com/fixture/pkg", Name: "TestNew"},
	}
	results := []parsers.Test{{Package: "example.com/fixture/pkg", Name: "TestNew", Passed: true}}
	if !completedUIRSelection(inventory, []string{"known"}, results) {
		t.Fatal("selection with every unskippable result was incomplete")
	}
	if completedUIRSelection(inventory, nil, results) {
		t.Fatal("selection missing a non-skippable inventory result was complete")
	}
	if completedUIRSelection(inventory, []string{"known"}, []parsers.Test{{Package: "example.com/fixture/pkg", Name: "TestNew", Skipped: true}}) {
		t.Fatal("skipped execution result was complete")
	}
}

func newUIRSkipFixture(t *testing.T, mutateDuringTest bool) string {
	t.Helper()
	repo := t.TempDir()
	mutation := ""
	if mutateDuringTest {
		mutation = `
	if err := os.WriteFile("value.go", []byte("package pkg\n\nconst Value = 2\n"), 0o600); err != nil { t.Fatal(err) }`
	}
	files := map[string]string{
		"go.mod":       "module example.com/fixture\n\ngo 1.22\n",
		"pkg/value.go": "package pkg\n\nconst Value = 1\n",
		"pkg/value_test.go": `package pkg

import (
	"os"
	"testing"
)

func TestKnown(t *testing.T) {
	if err := os.WriteFile("../known-ran", []byte("known"), 0o600); err != nil { t.Fatal(err) }
}

func TestNew(t *testing.T) {
	if err := os.WriteFile("../new-ran", []byte("new"), 0o600); err != nil { t.Fatal(err) }
` + mutation + `
}
`,
	}
	for name, content := range files {
		path := filepath.Join(repo, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitIn(t, repo, "init", "-q")
	gitIn(t, repo, "add", ".")
	gitIn(t, repo, "commit", "-qm", "fixture")
	return repo
}

func newFakeUIR(t *testing.T, mutate bool) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "uir")
	script := `#!/usr/bin/env python3
import json
import pathlib
import sys

args = sys.argv[1:]
location = pathlib.Path(args[args.index("--location") + 1])
if args[2:4] == ["tests", "plan"]:
    request = json.loads(pathlib.Path(args[4]).read_text())
    known = next(test["id"] for test in request["inventory"] if test["name"] == "TestKnown")
    if ` + map[bool]string{true: "True", false: "False"}[mutate] + `:
        (location / "pkg" / "value.go").write_text("package pkg\n\nconst Value = 2\n")
    print(json.dumps({"version": 1, "plan_id": "plan-1", "inventory_digest": "inventory-1", "mode": "selected", "skippable": [known]}))
elif args[2:4] == ["tests", "import"]:
    (location / "imported").write_text("yes")
    print("{}")
else:
    raise SystemExit("unexpected arguments: " + repr(args))
`
	if err := os.WriteFile(binary, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return binary
}
