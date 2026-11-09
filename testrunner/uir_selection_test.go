package testrunner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestUIRSelectionBuildsAnchoredPackageRunsAndOmitsSkippedPackages(t *testing.T) {
	repo := t.TempDir()
	files := map[string]string{
		"go.mod": "module example.com/selection\n\ngo 1.22\n",
		"selected/selected_test.go": `package selected
import ("os"; "testing")
func TestExact(t *testing.T) { if err := os.WriteFile("../exact-ran", []byte("yes"), 0o600); err != nil { t.Fatal(err) } }
func TestOther(t *testing.T) { if err := os.WriteFile("../other-ran", []byte("yes"), 0o600); err != nil { t.Fatal(err) } }
`,
		"skipped/skipped_test.go": `package skipped
import ("os"; "testing")
func TestSkippedOnly(t *testing.T) { if err := os.WriteFile("../skipped-ran", []byte("yes"), 0o600); err != nil { t.Fatal(err) } }
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

	var results []parsers.Test
	_, err := Run(RunOptions{
		WorkDir: repo, Recursive: true, Frameworks: []string{"go"}, ResultsOut: &results,
		UIRSelected: map[string][]string{
			"./selected": {"TestExact"},
			"./skipped":  {},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Name != "TestExact" {
		t.Fatalf("results = %+v", results)
	}
	if !strings.Contains(results[0].Command, "-run=^(?:TestExact)$") {
		t.Fatalf("command is not anchored to the package selection: %s", results[0].Command)
	}
	for _, marker := range []string{"other-ran", "skipped-ran"} {
		if _, statErr := os.Stat(filepath.Join(repo, marker)); !os.IsNotExist(statErr) {
			t.Fatalf("unselected test wrote %s: %v", marker, statErr)
		}
	}
	if _, err := os.Stat(filepath.Join(repo, "exact-ran")); err != nil {
		t.Fatalf("selected test did not run: %v", err)
	}
}
