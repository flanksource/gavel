package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/uirplan"
)

func TestUIRSkipRealBinaryRoundTrip(t *testing.T) {
	binary := os.Getenv("GAVEL_UIR_INTEGRATION_BINARY")
	if binary == "" {
		t.Skip("set GAVEL_UIR_INTEGRATION_BINARY to run the paired-binary integration")
	}
	repo := newUIRSkipFixture(t, false)
	dsn := filepath.Join(t.TempDir(), "uir.db")
	t.Setenv("UIR_DSN", dsn)
	command := exec.Command(binary, "add", repo)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("index fixture: %v: %s", err, output)
	}
	export := filepath.Join(repo, "uir-run.json")
	options := testrunner.RunOptions{
		Context: context.Background(), WorkDir: repo, Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
		UIRSkip: true, UIRBinary: binary, UIRExport: export, UIRInputContext: "postgres=test",
	}
	_, err := runTests(options, false)
	if err != nil {
		t.Fatalf("paired full-baseline round trip: %v", err)
	}
	for _, marker := range []string{"known-ran", "new-ran", "dot-ran", "suite-ran", "fuzz-ran", "example-ran"} {
		if err := os.Remove(filepath.Join(repo, marker)); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
	}
	_, err = runTests(options, false)
	if err != nil {
		t.Fatalf("paired selected round trip: %v", err)
	}
	data, err := os.ReadFile(export)
	if err != nil || !strings.Contains(string(data), `"version":2`) || !strings.Contains(string(data), `"selection_mode":"partial"`) || !strings.Contains(string(data), `"inventory"`) {
		t.Fatalf("paired export: %s: %v", data, err)
	}
	for _, marker := range []string{"known-ran", "new-ran", "dot-ran"} {
		if _, err := os.Stat(filepath.Join(repo, marker)); !os.IsNotExist(err) {
			t.Fatalf("real selected plan did not omit %s: %v", marker, err)
		}
	}
	for _, marker := range []string{"suite-ran", "fuzz-ran", "example-ran"} {
		if _, err := os.Stat(filepath.Join(repo, marker)); err != nil {
			t.Fatalf("real selected plan omitted uninventoried %s: %v", marker, err)
		}
	}
}

func TestUIRSkipRunsUnskippableInventoryTestsAcrossSubprocesses(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake UIR executable is a Python script")
	}
	repo := newUIRSkipFixture(t, false)
	fake := newFakeUIR(t, false, false)
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
	for _, marker := range []string{"dot-ran", "suite-ran", "fuzz-ran", "example-ran"} {
		if _, err := os.Stat(filepath.Join(repo, marker)); err != nil {
			t.Fatalf("ordinary go test behavior did not write %s: %v", marker, err)
		}
	}
	if _, err := os.Stat(filepath.Join(repo, "imported")); err != nil {
		t.Fatalf("fake UIR import was not invoked: %v", err)
	}
	data, err := os.ReadFile(export)
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Version int `json:"version"`
		Runs    []struct {
			SelectionMode string `json:"selection_mode"`
			Complete      bool   `json:"complete"`
			Tests         []struct {
				Name string `json:"name"`
			} `json:"tests"`
			Inventory []struct {
				Name string `json:"name"`
			} `json:"inventory"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if document.Version != 2 || len(document.Runs) != 1 || document.Runs[0].SelectionMode != "partial" || !document.Runs[0].Complete {
		t.Fatalf("export metadata = %+v", document.Runs)
	}
	if len(document.Runs[0].Tests) < 5 {
		t.Fatalf("exported tests = %+v", document.Runs[0].Tests)
	}
	if len(document.Runs[0].Inventory) != 3 {
		t.Fatalf("exported inventory = %+v", document.Runs[0].Inventory)
	}
}

func TestUIRRunAllPreservesOrdinaryGoTestSemantics(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake UIR executable is a Python script")
	}
	repo := newUIRSkipFixture(t, false)
	_, err := runTests(testrunner.RunOptions{
		Context: context.Background(), WorkDir: repo, Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
		UIRSkip: true, UIRBinary: newFakeUIR(t, false, true), UIRExport: filepath.Join(repo, "uir-run.json"), UIRInputContext: "postgres=test",
	}, false)
	if err != nil {
		t.Fatalf("runTests: %v", err)
	}
	for _, marker := range []string{"known-ran", "new-ran", "dot-ran", "suite-ran", "fuzz-ran", "example-ran"} {
		if _, err := os.Stat(filepath.Join(repo, marker)); err != nil {
			t.Fatalf("run_all did not preserve %s: %v", marker, err)
		}
	}
}

func TestUIRSkipPreservesVetFailuresInRunAllAndSelected(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake UIR executable is a Python script")
	}
	for _, runAll := range []bool{true, false} {
		t.Run(map[bool]string{true: "run_all", false: "selected"}[runAll], func(t *testing.T) {
			repo := newUIRSkipFixture(t, false)
			if err := os.MkdirAll(filepath.Join(repo, "broken"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(repo, "broken", "broken.go"), []byte("package broken\nimport \"fmt\"\nfunc Broken() { fmt.Printf(\"%d\", \"wrong\") }\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			oldExitCode := exitCode
			exitCode = 0
			t.Cleanup(func() { exitCode = oldExitCode })
			_, runErr := runTests(testrunner.RunOptions{
				Context: context.Background(), WorkDir: repo, Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
				UIRSkip: true, UIRBinary: newFakeUIR(t, false, runAll), UIRExport: filepath.Join(repo, "uir-run.json"), UIRInputContext: "postgres=test",
			}, false)
			if runErr != nil {
				t.Fatalf("runTests infrastructure error: %v", runErr)
			}
			if exitCode != 1 {
				t.Fatal("vet failure did not fail the run")
			}
		})
	}
}

func TestUIRInvalidSuccessfulPlanRunsFullTests(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake UIR executable is a shell script")
	}
	responses := map[string]string{
		"unknown inventory id": "{\"version\":2,\"plan_id\":\"plan-1\",\"inventory_digest\":\"inventory-1\",\"mode\":\"selected\",\"skippable\":[\"unknown\"]}",
		"old protocol":         "{\"version\":1,\"plan_id\":\"plan-1\",\"inventory_digest\":\"inventory-1\",\"mode\":\"run_all\",\"skippable\":[]}",
	}
	for name, response := range responses {
		t.Run(name, func(t *testing.T) {
			repo := newUIRSkipFixture(t, false)
			binary := filepath.Join(t.TempDir(), "uir")
			if err := os.WriteFile(binary, []byte("#!/bin/sh\nprintf '%s\\n' '"+response+"'\n"), 0o755); err != nil {
				t.Fatal(err)
			}
			export := filepath.Join(repo, "uir-run.json")
			_, err := runTests(testrunner.RunOptions{
				Context: context.Background(), WorkDir: repo, Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
				UIRSkip: true, UIRBinary: binary, UIRExport: export, UIRInputContext: "postgres=test",
			}, false)
			if err != nil {
				t.Fatalf("runTests: %v", err)
			}
			for _, marker := range []string{"known-ran", "new-ran"} {
				if _, err := os.Stat(filepath.Join(repo, marker)); err != nil {
					t.Fatalf("invalid plan suppressed %s: %v", marker, err)
				}
			}
			data, err := os.ReadFile(export)
			if err != nil || !strings.Contains(string(data), `"version":1`) {
				t.Fatalf("fail-open export is not legacy v1: %s: %v", data, err)
			}
		})
	}
}

func TestUIRSkipIdentityMutationBeforeExecutionFailsOpen(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake UIR executable is a Python script")
	}
	repo := newUIRSkipFixture(t, false)
	_, err := runTests(testrunner.RunOptions{
		Context: context.Background(), WorkDir: repo, Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
		UIRSkip: true, UIRBinary: newFakeUIR(t, true, false), UIRExport: filepath.Join(repo, "uir-run.json"), UIRInputContext: "postgres=test",
	}, false)
	if err != nil {
		t.Fatalf("runTests error = %v", err)
	}
	for _, marker := range []string{"known-ran", "new-ran"} {
		if _, statErr := os.Stat(filepath.Join(repo, marker)); statErr != nil {
			t.Fatalf("full tests did not run after identity mutation: %s: %v", marker, statErr)
		}
	}
	data, readErr := os.ReadFile(filepath.Join(repo, "uir-run.json"))
	if readErr != nil || !strings.Contains(string(data), `"version":1`) {
		t.Fatalf("fail-open export is not legacy v1: %s: %v", data, readErr)
	}
}

func TestUIRSkipRejectsIdentityMutationDuringExecution(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake UIR executable is a Python script")
	}
	repo := newUIRSkipFixture(t, true)
	_, err := runTests(testrunner.RunOptions{
		Context: context.Background(), WorkDir: repo, Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
		UIRSkip: true, UIRBinary: newFakeUIR(t, false, false), UIRExport: filepath.Join(repo, "uir-run.json"), UIRInputContext: "postgres=test",
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
	if !completedUIRSelection("/root", inventory, []string{"known"}, results) {
		t.Fatal("selection with every unskippable result was incomplete")
	}
	if completedUIRSelection("/root", inventory, nil, results) {
		t.Fatal("selection missing a non-skippable inventory result was complete")
	}
	if !completedUIRSelection("/root", inventory, []string{"known"}, []parsers.Test{{Package: "example.com/fixture/pkg", Name: "TestNew", Skipped: true}}) {
		t.Fatal("terminal skip made the execution incomplete")
	}
	if !completedUIRSelection("/root", inventory, []string{"known"}, append(results,
		parsers.Test{WorkDir: "/root/child", Package: "example.com/child", Name: "TestChild", Failed: true})) {
		t.Fatal("child-module failure made the root execution incomplete")
	}
	if completedUIRSelection("/root", inventory, []string{"known"}, []parsers.Test{{
		WorkDir: "/root/child", Package: "example.com/fixture/pkg", Name: "TestNew", Passed: true,
	}}) {
		t.Fatal("child-module result satisfied the root inventory")
	}
	if completedUIRSelection("/root", inventory, []string{"known"}, append(results,
		parsers.Test{WorkDir: "/root", Package: "example.com/fixture/other", Name: "Build Failed", Failed: true})) {
		t.Fatal("root-module failure left the execution complete")
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
		".gitignore":   "*-ran\nimported\nuir-run.json\n",
		"go.mod":       "module example.com/fixture\n\ngo 1.22\n",
		"pkg/value.go": "package pkg\n\nconst Value = 1\n",
		"pkg/value_test.go": `package pkg

import (
	"fmt"
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

func RunSpecs(t *testing.T) { if err := os.WriteFile("../suite-ran", []byte("suite"), 0o600); err != nil { t.Fatal(err) } }
func TestSuite(t *testing.T) { RunSpecs(t) }

func FuzzSeed(f *testing.F) {
	f.Add("seed")
	f.Fuzz(func(t *testing.T, value string) { if err := os.WriteFile("../fuzz-ran", []byte(value), 0o600); err != nil { t.Fatal(err) } })
}

func Example() {
	if err := os.WriteFile("../example-ran", []byte("example"), 0o600); err != nil { panic(err) }
	fmt.Println("ok")
	// Output: ok
}
`,
		"pkg/dot_test.go": `package pkg
import ("os"; . "testing")
func TestDot(t *T) { if err := os.WriteFile("../dot-ran", []byte("dot"), 0o600); err != nil { t.Fatal(err) } }
`,
		"examples/example_test.go": `package examples
import ("fmt"; "os")
func Example() {
	if err := os.WriteFile("../example-ran", []byte("example"), 0o600); err != nil { panic(err) }
	fmt.Println("only")
	// Output: only
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

func newFakeUIR(t *testing.T, mutate, runAll bool) string {
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
    run_all = ` + map[bool]string{true: "True", false: "False"}[runAll] + `
    print(json.dumps({"version": 2, "plan_id": "plan-1", "inventory_digest": "inventory-1", "mode": "run_all" if run_all else "selected", "skippable": [] if run_all else [known]}))
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
