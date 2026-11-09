package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/flanksource/gavel/testrunner"
)

func TestUIRSkipPreservesNestedModuleExecution(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake UIR executable is a Python script")
	}
	for _, test := range []struct {
		name      string
		selective bool
		wantMode  string
	}{
		{name: "run all", wantMode: "full"},
		{name: "select root test", selective: true, wantMode: "partial"},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := newNestedUIRSkipFixture(t)
			export := filepath.Join(repo, "uir-run.json")
			_, err := runTests(testrunner.RunOptions{
				Context: context.Background(), WorkDir: repo, Recursive: true, Frameworks: []string{"go"}, SkipHooks: true,
				UIRSkip: true, UIRBinary: newNestedFakeUIR(t, test.selective), UIRExport: export, UIRInputContext: "postgres=test",
			}, false)
			if err != nil {
				t.Fatalf("runTests: %v", err)
			}

			assertMarker(t, filepath.Join(repo, "root-required-ran"), true)
			assertMarker(t, filepath.Join(repo, "child", "child-ran"), true)
			assertMarker(t, filepath.Join(repo, "root-skippable-ran"), !test.selective)
			assertNestedUIRExport(t, export, test.wantMode, test.selective)
			assertNestedUIRInventory(t, filepath.Join(repo, ".gavel", "planned-inventory.json"))
			if _, err := os.Stat(filepath.Join(repo, "imported")); err != nil {
				t.Fatalf("root-module UIR import was not invoked: %v", err)
			}
		})
	}
}

func newNestedUIRSkipFixture(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	files := map[string]string{
		".gitignore": "*-ran\nimported\nuir-run.json\n",
		"go.mod":     "module example.com/root\n\ngo 1.22\n",
		"pkg/root_test.go": `package pkg
import ("os"; "testing")
func TestRootSkippable(t *testing.T) { if err := os.WriteFile("../root-skippable-ran", []byte("yes"), 0o600); err != nil { t.Fatal(err) } }
func TestRootRequired(t *testing.T) { if err := os.WriteFile("../root-required-ran", []byte("yes"), 0o600); err != nil { t.Fatal(err) } }
`,
		"child/go.mod": "module example.com/child\n\ngo 1.22\n",
		"child/pkg/child_test.go": `package pkg
import ("os"; "testing")
func TestChild(t *testing.T) { if err := os.WriteFile("../child-ran", []byte("yes"), 0o600); err != nil { t.Fatal(err) } }
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

func newNestedFakeUIR(t *testing.T, selective bool) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "uir")
	selectTest := "False"
	planMode := "run_all"
	if selective {
		selectTest = "True"
		planMode = "selected"
	}
	script := `#!/usr/bin/env python3
import json
import pathlib
import sys

args = sys.argv[1:]
location = pathlib.Path(args[args.index("--location") + 1])
if args[2:4] == ["tests", "plan"]:
    request = json.loads(pathlib.Path(args[4]).read_text())
    (location / ".gavel").mkdir(exist_ok=True)
    (location / ".gavel" / "planned-inventory.json").write_text(json.dumps(request["inventory"]))
    skippable = []
    if ` + selectTest + `:
        skippable = [next(test["id"] for test in request["inventory"] if test["name"] == "TestRootSkippable")]
    print(json.dumps({"version": 2, "plan_id": "plan-1", "inventory_digest": "inventory-1", "mode": "` + planMode + `", "skippable": skippable}))
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

func assertMarker(t *testing.T, path string, want bool) {
	t.Helper()
	_, err := os.Stat(path)
	if want && err != nil {
		t.Fatalf("expected marker %s: %v", path, err)
	}
	if !want && !os.IsNotExist(err) {
		t.Fatalf("unexpected marker %s: %v", path, err)
	}
}

func assertNestedUIRExport(t *testing.T, path, mode string, selective bool) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Version int `json:"version"`
		Runs    []struct {
			SelectionMode string `json:"selection_mode"`
			Complete      bool   `json:"complete"`
			Tests         []struct {
				Package string `json:"package"`
				Name    string `json:"name"`
			} `json:"tests"`
			Inventory []struct {
				Package string `json:"package"`
			} `json:"inventory"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if document.Version != 2 || len(document.Runs) != 1 || document.Runs[0].SelectionMode != mode || !document.Runs[0].Complete {
		t.Fatalf("export metadata = %+v", document.Runs)
	}
	if len(document.Runs[0].Inventory) != 2 {
		t.Fatalf("export inventory = %+v", document.Runs[0].Inventory)
	}
	wantTests := map[string]bool{"TestRootRequired": true, "TestRootSkippable": !selective}
	wantCount := 2
	if selective {
		wantCount = 1
	}
	if len(document.Runs[0].Tests) != wantCount {
		t.Fatalf("exported tests = %+v", document.Runs[0].Tests)
	}
	for _, result := range document.Runs[0].Tests {
		if result.Package != "example.com/root/pkg" || !wantTests[result.Name] {
			t.Fatalf("nested-module identity leaked into root export: %+v", result)
		}
	}
}

func assertNestedUIRInventory(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var inventory []struct {
		Package string `json:"package"`
	}
	if err := json.Unmarshal(data, &inventory); err != nil {
		t.Fatal(err)
	}
	if len(inventory) != 2 {
		t.Fatalf("root inventory has %d tests, want 2: %+v", len(inventory), inventory)
	}
	for _, test := range inventory {
		if test.Package != "example.com/root/pkg" {
			t.Fatalf("nested module entered the root plan: %+v", inventory)
		}
	}
}
