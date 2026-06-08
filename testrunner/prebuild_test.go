package testrunner

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestGoPackagesToWarm_UnionsGoFrameworksAndDedupes(t *testing.T) {
	byFw := map[parsers.Framework][]string{
		parsers.GoTest:     {"./b", "./a"},
		parsers.Ginkgo:     {"./a", "./c"},
		parsers.Vitest:     {"./web"},
		parsers.Playwright: {"./e2e"},
	}

	got := goPackagesToWarm(byFw)
	want := []string{"./a", "./b", "./c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("goPackagesToWarm = %v, want %v (JS frameworks excluded, deduped, sorted)", got, want)
	}
}

func TestGoPreBuildArgs_CompilesWithoutRunning(t *testing.T) {
	got := goPreBuildArgs([]string{"./a", "./b"})
	want := []string{"test", "-count=0", "./a", "./b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("goPreBuildArgs = %v, want %v", got, want)
	}
}

func TestPreBuildGoPackages_NoPackagesIsNoop(t *testing.T) {
	o := &TestOrchestrator{}
	if err := o.preBuildGoPackages(nil); err != nil {
		t.Fatalf("preBuildGoPackages(nil) = %v, want nil", err)
	}
}

func TestPreBuildGoPackages_CompilesValidPackage(t *testing.T) {
	repo := t.TempDir()
	writeGoTestPackage(t, repo, "sample", "example.com/warm")

	o := &TestOrchestrator{}
	o.WorkDir = repo

	if err := o.preBuildGoPackages([]string{"./sample"}); err != nil {
		t.Fatalf("preBuildGoPackages on a valid package = %v, want nil", err)
	}
}

func TestPreBuildGoPackages_FailsFastOnCompileError(t *testing.T) {
	repo := t.TempDir()
	writeGoTestPackage(t, repo, "sample", "example.com/warm")
	// Introduce a deliberate syntax error so compilation fails.
	broken := filepath.Join(repo, "sample", "broken_test.go")
	if err := os.WriteFile(broken, []byte("package sample\n\nfunc this is not valid go {\n"), 0o644); err != nil {
		t.Fatalf("write broken file: %v", err)
	}

	o := &TestOrchestrator{}
	o.WorkDir = repo

	err := o.preBuildGoPackages([]string{"./sample"})
	if err == nil {
		t.Fatal("preBuildGoPackages on a broken package = nil, want compile error")
	}
	if !strings.Contains(err.Error(), "pre-build") {
		t.Fatalf("error %q should be tagged as a pre-build failure", err.Error())
	}
	if !strings.Contains(err.Error(), "broken_test.go") {
		t.Fatalf("error %q should surface the compiler output naming the broken file", err.Error())
	}
}
