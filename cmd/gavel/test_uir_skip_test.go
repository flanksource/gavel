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
	addIgnoredGeneratedConstantFixture(t, repo)
	dsn := filepath.Join(t.TempDir(), "uir.db")
	t.Setenv("UIR_DSN", dsn)
	command := exec.Command(binary, "add", repo)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("index fixture: %v: %s", err, output)
	}
	export := filepath.Join(repo, "uir-run.json")
	options := testrunner.RunOptions{
		Context: context.Background(), WorkDir: repo, Dirs: testWorkspaceDirs(t, repo),
		Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
		UIRSkip: true, UIRBinary: binary, UIRExport: export, UIRInputContext: "postgres=test",
	}
	_, err := runTests(options, false)
	if err != nil {
		t.Fatalf("paired full-baseline round trip: %v", err)
	}
	markers := []string{"known-ran", "new-ran", "dot-ran", "generated-ran", "golden-ran", "suite-ran", "fuzz-ran", "example-ran"}
	if runtime.GOARCH == "amd64" {
		markers = append(markers, "assembly-ran")
	}
	for _, marker := range markers {
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
	skipped := []string{"known-ran", "new-ran", "dot-ran", "generated-ran", "golden-ran"}
	if runtime.GOARCH == "amd64" {
		skipped = append(skipped, "assembly-ran")
	}
	for _, marker := range skipped {
		if _, err := os.Stat(filepath.Join(repo, marker)); !os.IsNotExist(err) {
			t.Fatalf("real selected plan did not omit %s: %v", marker, err)
		}
	}
	for _, marker := range []string{"suite-ran", "fuzz-ran", "example-ran"} {
		if _, err := os.Stat(filepath.Join(repo, marker)); err != nil {
			t.Fatalf("real selected plan omitted uninventoried %s: %v", marker, err)
		}
	}

	generated := filepath.Join(repo, "gen", "zz_gen.go")
	if err := os.WriteFile(generated, []byte("package gen\nconst generatedValue = \"after\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldExitCode := exitCode
	exitCode = 0
	_, err = runTests(options, false)
	generatedExitCode := exitCode
	exitCode = oldExitCode
	if err != nil {
		t.Fatalf("paired ignored generated Go mutation run: %v", err)
	}
	if generatedExitCode != 1 {
		t.Fatal("changed ignored build-selected Go source did not fail the real paired run")
	}
	if _, err := os.Stat(filepath.Join(repo, "generated-ran")); err != nil {
		t.Fatalf("changed ignored build-selected Go source was skipped: %v", err)
	}
	if err := os.WriteFile(generated, []byte("package gen\nconst generatedValue = \"before\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(repo, "fixtures", "want.golden"), []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldExitCode = exitCode
	exitCode = 0
	_, err = runTests(options, false)
	goldenExitCode := exitCode
	exitCode = oldExitCode
	if err != nil {
		t.Fatalf("paired golden mutation run: %v", err)
	}
	if goldenExitCode != 1 {
		t.Fatal("changed golden input did not fail the real paired run")
	}
	if _, err := os.Stat(filepath.Join(repo, "golden-ran")); err != nil {
		t.Fatalf("changed golden input was skipped: %v", err)
	}

	if runtime.GOARCH == "amd64" {
		if err := os.WriteFile(filepath.Join(repo, "fixtures", "want.golden"), []byte("want\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(filepath.Join(repo, "assembly-ran")); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
		assembly := filepath.Join(repo, "assembly", "add_amd64.s")
		content, readErr := os.ReadFile(assembly)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if err := os.WriteFile(assembly, []byte(strings.Replace(string(content), "ADDQ", "SUBQ", 1)), 0o644); err != nil {
			t.Fatal(err)
		}
		oldExitCode = exitCode
		exitCode = 0
		_, err = runTests(options, false)
		assemblyExitCode := exitCode
		exitCode = oldExitCode
		if err != nil {
			t.Fatalf("paired assembly mutation run: %v", err)
		}
		if assemblyExitCode != 1 {
			t.Fatal("changed assembly input did not fail the real paired run")
		}
		if _, err := os.Stat(filepath.Join(repo, "assembly-ran")); err != nil {
			t.Fatalf("changed assembly input was skipped: %v", err)
		}
	}

	coverageRepo := newUIRSkipFixture(t, false)
	addIgnoredGeneratedCoverageFixture(t, coverageRepo)
	command = exec.Command(binary, "add", coverageRepo)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("index ignored generated coverage fixture: %v: %s", err, output)
	}
	coverageOptions := options
	coverageOptions.WorkDir = coverageRepo
	coverageOptions.UIRExport = filepath.Join(coverageRepo, "uir-run.json")
	oldExitCode = exitCode
	exitCode = 0
	_, err = runTests(coverageOptions, false)
	coverageExitCode := exitCode
	exitCode = oldExitCode
	if err != nil || coverageExitCode != 0 {
		t.Fatalf("ignored generated coverage metadata changed passing status: err=%v exit=%d", err, coverageExitCode)
	}
	if _, err := os.Stat(filepath.Join(coverageRepo, "generated-coverage-ran")); err != nil {
		t.Fatalf("ignored generated coverage test did not run: %v", err)
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
		Context: context.Background(), WorkDir: repo, Dirs: testWorkspaceDirs(t, repo),
		Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
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
	wantInventory := 4
	if runtime.GOARCH == "amd64" {
		wantInventory++
	}
	if len(document.Runs[0].Inventory) != wantInventory {
		t.Fatalf("exported inventory = %+v", document.Runs[0].Inventory)
	}
}

func TestUIRRunAllPreservesOrdinaryGoTestSemantics(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake UIR executable is a Python script")
	}
	repo := newUIRSkipFixture(t, false)
	_, err := runTests(testrunner.RunOptions{
		Context: context.Background(), WorkDir: repo, Dirs: testWorkspaceDirs(t, repo),
		Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
		UIRSkip: true, UIRBinary: newFakeUIR(t, false, true), UIRExport: filepath.Join(repo, "uir-run.json"), UIRInputContext: "postgres=test",
	}, false)
	if err != nil {
		t.Fatalf("runTests: %v", err)
	}
	markers := []string{"known-ran", "new-ran", "dot-ran", "golden-ran", "suite-ran", "fuzz-ran", "example-ran"}
	if runtime.GOARCH == "amd64" {
		markers = append(markers, "assembly-ran")
	}
	for _, marker := range markers {
		if _, err := os.Stat(filepath.Join(repo, marker)); err != nil {
			t.Fatalf("run_all did not preserve %s: %v", marker, err)
		}
	}
}

func TestUIRSkipInputInventoryFailureRunsFullWithoutExport(t *testing.T) {
	repo := newUIRSkipFixture(t, false)
	export := filepath.Join(repo, "uir-run.json")
	if err := os.WriteFile(filepath.Join(repo, "export_embed.go"), []byte("package fixture\n\nimport _ \"embed\"\n\n//go:embed uir-run.json\nvar exported []byte\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(export, []byte("seed"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := runTests(testrunner.RunOptions{
		Context: context.Background(), WorkDir: repo, Dirs: testWorkspaceDirs(t, repo),
		Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
		UIRSkip: true, UIRBinary: filepath.Join(repo, "missing-uir"), UIRExport: export, UIRInputContext: "postgres=test",
	}, false)
	if err != nil {
		t.Fatalf("input inventory failure did not fail open: %v", err)
	}
	for _, marker := range []string{"known-ran", "new-ran", "dot-ran", "suite-ran", "fuzz-ran", "example-ran"} {
		if _, err := os.Stat(filepath.Join(repo, marker)); err != nil {
			t.Fatalf("input inventory fallback omitted %s: %v", marker, err)
		}
	}
	if data, err := os.ReadFile(export); err != nil || string(data) != "seed" {
		t.Fatalf("unsafe fallback wrote export: data=%q err=%v", data, err)
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
				Context: context.Background(), WorkDir: repo, Dirs: testWorkspaceDirs(t, repo),
				Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
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
				Context: context.Background(), WorkDir: repo, Dirs: testWorkspaceDirs(t, repo),
				Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
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
		Context: context.Background(), WorkDir: repo, Dirs: testWorkspaceDirs(t, repo),
		Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
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

func TestUIRSkipRerunsFullAfterSelectedIdentityMutation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake UIR executable is a Python script")
	}
	repo := newUIRSkipFixture(t, true)
	testFile := filepath.Join(repo, "pkg", "value_test.go")
	testSource, readErr := os.ReadFile(testFile)
	if readErr != nil {
		t.Fatal(readErr)
	}
	testSource = []byte(strings.Replace(string(testSource),
		`if err := os.WriteFile("../known-ran", []byte("known"), 0o600); err != nil { t.Fatal(err) }`,
		`if err := os.WriteFile("../known-ran", []byte("known"), 0o600); err != nil { t.Fatal(err) }
	if Value != 1 { t.Fatalf("Value = %d", Value) }`, 1))
	if err := os.WriteFile(testFile, testSource, 0o644); err != nil {
		t.Fatal(err)
	}
	oldExitCode := exitCode
	exitCode = 0
	t.Cleanup(func() { exitCode = oldExitCode })
	_, err := runTests(testrunner.RunOptions{
		Context: context.Background(), WorkDir: repo, Dirs: testWorkspaceDirs(t, repo),
		Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
		UIRSkip: true, UIRBinary: newFakeUIR(t, false, false), UIRExport: filepath.Join(repo, "uir-run.json"), UIRInputContext: "postgres=test",
	}, false)
	if err != nil {
		t.Fatalf("safe full rerun failed: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(repo, "new-ran")); statErr != nil {
		t.Fatalf("mutation fixture did not execute: %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(repo, "known-ran")); statErr != nil {
		t.Fatalf("successful invalidated selection was not rerun in full: %v", statErr)
	}
	if exitCode != 1 {
		t.Fatal("full safety rerun did not preserve the failure hidden by the invalidated selection")
	}
	if _, statErr := os.Stat(filepath.Join(repo, "imported")); !os.IsNotExist(statErr) {
		t.Fatalf("mutated execution was imported: %v", statErr)
	}
}

func TestUIRSkipFullIdentityMutationPreservesPassingStatusWithoutBaseline(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake UIR executable is a Python script")
	}
	repo := newUIRSkipFixture(t, false)
	addOutputWritingTest(t, repo, false)
	options := testrunner.RunOptions{
		Context: context.Background(), WorkDir: repo, Dirs: testWorkspaceDirs(t, repo),
		Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
		UIRSkip: true, UIRBinary: newFakeUIR(t, false, true), UIRExport: filepath.Join(repo, "uir-run.json"), UIRInputContext: "postgres=test",
	}
	for run := 1; run <= 2; run++ {
		_, err := runTests(options, false)
		if err != nil {
			t.Fatalf("passing full test command %d became a metadata failure: %v", run, err)
		}
		if _, statErr := os.Stat(filepath.Join(repo, "out.log")); statErr != nil {
			t.Fatalf("output-writing test %d did not run: %v", run, statErr)
		}
		if _, statErr := os.Stat(filepath.Join(repo, "imported")); !os.IsNotExist(statErr) {
			t.Fatalf("identity-mutated full execution %d became a baseline: %v", run, statErr)
		}
	}
}

func TestUIRSkipFullIdentityMutationPreservesFailingStatusWithoutBaseline(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake UIR executable is a Python script")
	}
	repo := newUIRSkipFixture(t, false)
	addOutputWritingTest(t, repo, true)
	oldExitCode := exitCode
	exitCode = 0
	t.Cleanup(func() { exitCode = oldExitCode })
	_, err := runTests(testrunner.RunOptions{
		Context: context.Background(), WorkDir: repo, Dirs: testWorkspaceDirs(t, repo),
		Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
		UIRSkip: true, UIRBinary: newFakeUIR(t, false, true), UIRExport: filepath.Join(repo, "uir-run.json"), UIRInputContext: "postgres=test",
	}, false)
	if err != nil {
		t.Fatalf("test failure became an infrastructure error: %v", err)
	}
	if exitCode != 1 {
		t.Fatal("failing full test command lost its failure status")
	}
	if _, statErr := os.Stat(filepath.Join(repo, "imported")); !os.IsNotExist(statErr) {
		t.Fatalf("failing identity-mutated execution became a baseline: %v", statErr)
	}
}

func TestUIRSkipImportFailurePreservesPassingStatus(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake UIR executable is a Python script")
	}
	repo := newUIRSkipFixture(t, false)
	_, err := runTests(testrunner.RunOptions{
		Context: context.Background(), WorkDir: repo, Dirs: testWorkspaceDirs(t, repo),
		Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
		UIRSkip: true, UIRBinary: newFakeUIRWithImportFailure(t), UIRExport: filepath.Join(repo, "uir-run.json"), UIRInputContext: "postgres=test",
	}, false)
	if err != nil {
		t.Fatalf("passing test command became an import failure: %v", err)
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
		"fixtures/want.golden": "want\n",
		"golden/golden_test.go": `package golden
import ("os"; "testing")
func TestGolden(t *testing.T) {
	if err := os.WriteFile("../golden-ran", []byte("golden"), 0o600); err != nil { t.Fatal(err) }
	got, err := os.ReadFile("../fixtures/want.golden")
	if err != nil { t.Fatal(err) }
	if string(got) != "want\n" { t.Fatalf("golden = %q", got) }
}
`,
	}
	if runtime.GOARCH == "amd64" {
		files["assembly/add.go"] = "package assembly\nfunc Add(a, b int) int\n"
		files["assembly/add_amd64.s"] = "#include \"textflag.h\"\nTEXT ·Add(SB), NOSPLIT, $0-24\n\tMOVQ a+0(FP), AX\n\tADDQ b+8(FP), AX\n\tMOVQ AX, ret+16(FP)\n\tRET\n"
		files["assembly/add_test.go"] = `package assembly
import ("os"; "testing")
func TestAssembly(t *testing.T) {
	if err := os.WriteFile("../assembly-ran", []byte("assembly"), 0o600); err != nil { t.Fatal(err) }
	if got := Add(1, 2); got != 3 { t.Fatalf("Add(1, 2) = %d", got) }
}
`
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

func addOutputWritingTest(t *testing.T, repo string, fail bool) {
	t.Helper()
	failure := ""
	if fail {
		failure = `
	t.Fatal("intentional failure")`
	}
	if err := os.WriteFile(filepath.Join(repo, "pkg", "writer_test.go"), []byte(`package pkg
import ("os"; "testing")
func TestWritesOutput(t *testing.T) {
	previous, _ := os.ReadFile("../out.log")
	if err := os.WriteFile("../out.log", append(previous, 'x'), 0o600); err != nil { t.Fatal(err) }`+failure+`
}
`), 0o644); err != nil {
		t.Fatal(err)
	}
}

func addIgnoredGeneratedConstantFixture(t *testing.T, repo string) {
	t.Helper()
	gitignore, err := os.ReadFile(filepath.Join(repo, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		".gitignore":      string(gitignore) + "gen/zz_gen.go\n",
		"gen/use.go":      "package gen\nfunc Value() string { return generatedValue }\n",
		"gen/zz_gen.go":   "package gen\nconst generatedValue = \"before\"\n",
		"gen/use_test.go": "package gen\nimport (\"os\"; \"testing\")\nfunc TestGenerated(t *testing.T) { if err := os.WriteFile(\"../generated-ran\", []byte(\"yes\"), 0o600); err != nil { t.Fatal(err) }; if got := Value(); got != \"before\" { t.Fatalf(\"Value() = %q\", got) } }\n",
	}
	writeUIRSkipFixtureFiles(t, repo, files)
	gitIn(t, repo, "add", ".gitignore", "gen/use.go", "gen/use_test.go")
	gitIn(t, repo, "commit", "-qm", "add ignored generated source fixture")
}

func addIgnoredGeneratedCoverageFixture(t *testing.T, repo string) {
	t.Helper()
	gitignore, err := os.ReadFile(filepath.Join(repo, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		".gitignore":            string(gitignore) + "coveragegen/zz_gen.go\n",
		"coveragegen/use.go":    "package coveragegen\nfunc Value() string { return generatedValue() }\n",
		"coveragegen/zz_gen.go": "package coveragegen\nfunc generatedValue() string { return \"ok\" }\n",
		"coveragegen/use_test.go": `package coveragegen
import ("os"; "testing")
func TestGeneratedCoverage(t *testing.T) {
	if err := os.WriteFile("../generated-coverage-ran", []byte("yes"), 0o600); err != nil { t.Fatal(err) }
	if got := Value(); got != "ok" { t.Fatalf("Value() = %q", got) }
}
`,
	}
	writeUIRSkipFixtureFiles(t, repo, files)
	gitIn(t, repo, "add", ".gitignore", "coveragegen/use.go", "coveragegen/use_test.go")
	gitIn(t, repo, "commit", "-qm", "add ignored generated coverage fixture")
}

func writeUIRSkipFixtureFiles(t *testing.T, repo string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(repo, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
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

func newFakeUIRWithImportFailure(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "uir")
	script := `#!/usr/bin/env python3
import json
import pathlib
import sys

args = sys.argv[1:]
if args[2:4] == ["tests", "plan"]:
    print(json.dumps({"version": 2, "plan_id": "plan-1", "inventory_digest": "inventory-1", "mode": "run_all", "skippable": []}))
elif args[2:4] == ["tests", "import"]:
    print("fixture import failure", file=sys.stderr)
    raise SystemExit(1)
else:
    raise SystemExit("unexpected arguments: " + repr(args))
`
	if err := os.WriteFile(binary, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return binary
}
