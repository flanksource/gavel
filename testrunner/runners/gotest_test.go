package runners

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func writeGoMod(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}
}

func TestGoTestDetect(t *testing.T) {
	tmpDir := t.TempDir()
	writeGoMod(t, tmpDir)

	// No test files yet
	runner := NewGoTest(tmpDir)
	found, err := runner.Detect(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected no test files detected")
	}

	// Create a test file
	testFile := filepath.Join(tmpDir, "example_test.go")
	if err := os.WriteFile(testFile, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	found, err = runner.Detect(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected test file to be detected")
	}
}

func TestGoTestDetectSkipsGitIgnoredFiles(t *testing.T) {
	tmpDir := t.TempDir()
	writeGoMod(t, tmpDir)

	if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("ignored/\n"), 0644); err != nil {
		t.Fatalf("failed to create .gitignore: %v", err)
	}

	ignoredDir := filepath.Join(tmpDir, "ignored")
	if err := os.MkdirAll(ignoredDir, 0755); err != nil {
		t.Fatalf("failed to create ignored directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(ignoredDir, "ignored_test.go"), []byte("package ignored\n"), 0644); err != nil {
		t.Fatalf("failed to create ignored test file: %v", err)
	}

	runner := NewGoTest(tmpDir)
	found, err := runner.Detect(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected gitignored test files to be skipped")
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "visible_test.go"), []byte("package visible\n"), 0644); err != nil {
		t.Fatalf("failed to create visible test file: %v", err)
	}

	found, err = runner.Detect(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected visible test file to be detected")
	}
}

func TestGoTestDetectSkipsNestedProjectRoots(t *testing.T) {
	tmpDir := t.TempDir()
	writeGoMod(t, tmpDir)

	if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	nestedModule := filepath.Join(tmpDir, "nested-module")
	if err := os.MkdirAll(filepath.Join(nestedModule, "pkg"), 0o755); err != nil {
		t.Fatalf("failed to create nested module: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedModule, "go.mod"), []byte("module example.com/nested\n"), 0o644); err != nil {
		t.Fatalf("failed to create nested go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedModule, "pkg", "nested_test.go"), []byte("package pkg\n"), 0o644); err != nil {
		t.Fatalf("failed to create nested module test: %v", err)
	}

	nestedRepo := filepath.Join(tmpDir, "nested-repo")
	if err := os.MkdirAll(filepath.Join(nestedRepo, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create nested repo .git: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(nestedRepo, "pkg"), 0o755); err != nil {
		t.Fatalf("failed to create nested repo package: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedRepo, "pkg", "nested_test.go"), []byte("package pkg\n"), 0o644); err != nil {
		t.Fatalf("failed to create nested repo test: %v", err)
	}

	runner := NewGoTest(tmpDir)
	found, err := runner.Detect(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected nested module and repo tests to be ignored from the parent root")
	}

	found, err = runner.Detect(nestedModule)
	if err != nil {
		t.Fatalf("unexpected error detecting nested module directly: %v", err)
	}
	if !found {
		t.Error("expected nested module tests to be detected when starting inside that module")
	}

	found, err = runner.Detect(nestedRepo)
	if err != nil {
		t.Fatalf("unexpected error detecting nested repo directly: %v", err)
	}
	if !found {
		t.Error("expected nested repo tests to be detected when starting inside that repo")
	}
}

func TestGoTestDetectsTestFilesWithoutGoMod(t *testing.T) {
	// Detection is purely about the presence of *_test.go. If the user runs
	// gavel inside a directory that has Go test files but is not inside a
	// module, `go test` itself will emit a useful error — gavel's job is
	// to surface the intent, not pre-validate toolchain state.
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "example_test.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	runner := NewGoTest(tmpDir)
	found, err := runner.Detect(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected go test file to be detected even without a go.mod")
	}

	packages, err := runner.DiscoverPackages(tmpDir, true)
	if err != nil {
		t.Fatalf("unexpected error discovering packages: %v", err)
	}
	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %v", packages)
	}
}

func TestGoTestDiscoverPackages(t *testing.T) {
	tmpDir := t.TempDir()
	writeGoMod(t, tmpDir)

	// Create test structure
	pkgA := filepath.Join(tmpDir, "pkg", "a")
	pkgB := filepath.Join(tmpDir, "pkg", "b")

	for _, pkg := range []string{pkgA, pkgB} {
		if err := os.MkdirAll(pkg, 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		if err := os.WriteFile(filepath.Join(pkg, "test_test.go"), []byte("package a\n"), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	runner := NewGoTest(tmpDir)
	packages, err := runner.DiscoverPackages(tmpDir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(packages) != 2 {
		t.Errorf("expected 2 packages, got %d", len(packages))
	}

	// Check that packages are relative paths
	for _, pkg := range packages {
		if !strings.HasPrefix(pkg, "./") {
			t.Errorf("expected relative path, got %s", pkg)
		}
	}
}

func TestGoTestDiscoverPackagesNonRecursive(t *testing.T) {
	tmpDir := t.TempDir()
	writeGoMod(t, tmpDir)

	// Create test file in root and in a subdirectory
	if err := os.WriteFile(filepath.Join(tmpDir, "root_test.go"), []byte("package root\n"), 0644); err != nil {
		t.Fatal(err)
	}
	subDir := filepath.Join(tmpDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "sub_test.go"), []byte("package sub\n"), 0644); err != nil {
		t.Fatal(err)
	}

	runner := NewGoTest(tmpDir)

	// Non-recursive should only find root
	packages, err := runner.DiscoverPackages(tmpDir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d: %v", len(packages), packages)
	}
	if packages[0] != "./." {
		t.Errorf("expected './.', got %s", packages[0])
	}

	// Recursive should find both
	packages, err = runner.DiscoverPackages(tmpDir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(packages) != 2 {
		t.Fatalf("expected 2 packages, got %d: %v", len(packages), packages)
	}
}

func TestGoTestDiscoverPackagesSkipsNestedProjectRoots(t *testing.T) {
	tmpDir := t.TempDir()
	writeGoMod(t, tmpDir)

	if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create root .git: %v", err)
	}

	rootPkg := filepath.Join(tmpDir, "rootpkg")
	if err := os.MkdirAll(rootPkg, 0o755); err != nil {
		t.Fatalf("failed to create root package: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootPkg, "root_test.go"), []byte("package rootpkg\n"), 0o644); err != nil {
		t.Fatalf("failed to create root test file: %v", err)
	}

	nestedModule := filepath.Join(tmpDir, "nested-module")
	if err := os.MkdirAll(filepath.Join(nestedModule, "pkg"), 0o755); err != nil {
		t.Fatalf("failed to create nested module package: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedModule, "go.mod"), []byte("module example.com/nested\n"), 0o644); err != nil {
		t.Fatalf("failed to create nested go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedModule, "pkg", "nested_test.go"), []byte("package pkg\n"), 0o644); err != nil {
		t.Fatalf("failed to create nested module test file: %v", err)
	}

	nestedRepo := filepath.Join(tmpDir, "nested-repo")
	if err := os.MkdirAll(filepath.Join(nestedRepo, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create nested repo .git: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(nestedRepo, "pkg"), 0o755); err != nil {
		t.Fatalf("failed to create nested repo package: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedRepo, "pkg", "nested_test.go"), []byte("package pkg\n"), 0o644); err != nil {
		t.Fatalf("failed to create nested repo test file: %v", err)
	}

	runner := NewGoTest(tmpDir)

	packages, err := runner.DiscoverPackages(tmpDir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(packages) != 1 || packages[0] != "./rootpkg" {
		t.Fatalf("expected only root package, got %v", packages)
	}

	modulePackages, err := runner.DiscoverPackages(nestedModule, true)
	if err != nil {
		t.Fatalf("unexpected error discovering nested module directly: %v", err)
	}
	if len(modulePackages) != 1 || modulePackages[0] != "./nested-module/pkg" {
		t.Fatalf("expected nested module package when starting inside it, got %v", modulePackages)
	}

	repoPackages, err := runner.DiscoverPackages(nestedRepo, true)
	if err != nil {
		t.Fatalf("unexpected error discovering nested repo directly: %v", err)
	}
	if len(repoPackages) != 1 || repoPackages[0] != "./nested-repo/pkg" {
		t.Fatalf("expected nested repo package when starting inside it, got %v", repoPackages)
	}
}

func TestGoTestPackageHasTests(t *testing.T) {
	tmpDir := t.TempDir()
	pkgDir := filepath.Join(tmpDir, "pkg")

	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	runner := NewGoTest(tmpDir)

	// Should not have tests yet
	found, err := runner.PackageHasTests("./pkg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected no test files")
	}

	// Add a test file
	if err := os.WriteFile(filepath.Join(pkgDir, "test_test.go"), []byte("package pkg\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	found, err = runner.PackageHasTests("./pkg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected test file to be found")
	}
}

func TestGoTestName(t *testing.T) {
	runner := NewGoTest("/tmp")
	if runner.Name() != parsers.GoTest {
		t.Errorf("expected framework GoTest, got %v", runner.Name())
	}
}

func TestGoTestParserType(t *testing.T) {
	runner := NewGoTest("/tmp")
	parser := runner.Parser()
	if parser.Name() != "go test json" {
		t.Errorf("expected parser name 'go test json', got %s", parser.Name())
	}
}

func TestGoTestGinkgoImportDetection(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with Ginkgo v2 import
	ginkgoTestContent := `package test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test", func() {
	It("should test", func() {
		Expect(true).To(BeTrue())
	})
})
`
	testFile := filepath.Join(tmpDir, "ginkgo_test.go")
	if err := os.WriteFile(testFile, []byte(ginkgoTestContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	runner := NewGoTest(tmpDir)

	// Ginkgo test files should not be considered GoTest
	found, err := runner.PackageHasTests(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected no GoTest files (only Ginkgo imports)")
	}

	// Now add a regular GoTest file
	goTestContent := `package test

import "testing"

func TestRegular(t *testing.T) {
	if 1 != 1 {
		t.Error("math is broken")
	}
}
`
	goTestFile := filepath.Join(tmpDir, "regular_test.go")
	if err := os.WriteFile(goTestFile, []byte(goTestContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Now should find GoTest files
	found, err = runner.PackageHasTests(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected to find GoTest files (mixed with Ginkgo)")
	}
}

func TestGoTestDiscoverPackagesExcludesGinkgoOnly(t *testing.T) {
	tmpDir := t.TempDir()
	writeGoMod(t, tmpDir)

	// Create a Ginkgo-only package
	ginkgoPkg := filepath.Join(tmpDir, "ginkgo_pkg")
	if err := os.MkdirAll(ginkgoPkg, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	ginkgoTestContent := `package ginkgo_pkg

import . "github.com/onsi/ginkgo/v2"

var _ = Describe("Ginkgo Test", func() {})
`
	if err := os.WriteFile(filepath.Join(ginkgoPkg, "ginkgo_test.go"), []byte(ginkgoTestContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create a regular GoTest package
	goTestPkg := filepath.Join(tmpDir, "gotest_pkg")
	if err := os.MkdirAll(goTestPkg, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	goTestContent := `package gotest_pkg

import "testing"

func TestExample(t *testing.T) {}
`
	if err := os.WriteFile(filepath.Join(goTestPkg, "example_test.go"), []byte(goTestContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	runner := NewGoTest(tmpDir)
	packages, err := runner.DiscoverPackages(tmpDir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only find the GoTest package, not the Ginkgo-only package
	if len(packages) != 1 {
		t.Errorf("expected 1 package, got %d", len(packages))
	}

	if !strings.Contains(packages[0], "gotest_pkg") {
		t.Errorf("expected to find gotest_pkg, got %v", packages)
	}
}

func TestGoTestBenchmarkDetection(t *testing.T) {
	tmpDir := t.TempDir()

	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	runner := NewGoTest(tmpDir)

	t.Run("empty package has no tests or benches", func(t *testing.T) {
		if runner.PackageHasBenchmarks(".") {
			t.Error("empty package should not report benchmarks")
		}
		if runner.PackageHasGoTests(".") {
			t.Error("empty package should not report tests")
		}
	})

	t.Run("bench-only with TestMain", func(t *testing.T) {
		write("main_test.go", `package t

import "testing"

func TestMain(m *testing.M) { m.Run() }
`)
		write("bench_test.go", `package t

import "testing"

func BenchmarkFoo(b *testing.B) {
	for i := 0; i < b.N; i++ {
	}
}
`)
		if !runner.PackageHasBenchmarks(".") {
			t.Error("expected benchmarks to be detected")
		}
		if runner.PackageHasGoTests(".") {
			t.Error("TestMain alone must not count as runnable tests")
		}
	})

	t.Run("mixed tests and benches", func(t *testing.T) {
		write("real_test.go", `package t

import "testing"

func TestReal(t *testing.T) {}
`)
		if !runner.PackageHasBenchmarks(".") {
			t.Error("expected benchmarks to still be detected")
		}
		if !runner.PackageHasGoTests(".") {
			t.Error("expected real Test* func to count")
		}
	})

	t.Run("ginkgo file with benchmark import is ignored", func(t *testing.T) {
		ginkgoDir := filepath.Join(tmpDir, "ginkgo_bench")
		if err := os.MkdirAll(ginkgoDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(ginkgoDir, "bench_test.go"), []byte(`package ginkgo_bench

import (
	"testing"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("x", func() {})

func BenchmarkIgnored(b *testing.B) {}
`), 0644); err != nil {
			t.Fatal(err)
		}
		if runner.PackageHasBenchmarks("./ginkgo_bench") {
			t.Error("Ginkgo-tagged file should be ignored for bench detection")
		}
	})
}

func TestGoTestBuildCommandHonoursTimeoutFlag(t *testing.T) {
	tmpDir := t.TempDir()
	writeGoMod(t, tmpDir)
	runner := NewGoTest(tmpDir)

	// Caller passes -timeout once at the front; a user-supplied override
	// comes in as a later extraArg and must win via Go's last-flag-wins rule.
	testRun, err := runner.BuildCommand("./pkg", "-timeout=4m30s", "-run", "TestFoo", "-timeout=1s")
	if err != nil {
		t.Fatalf("BuildCommand: %v", err)
	}
	args := testRun.Process.Args
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-timeout=4m30s") {
		t.Errorf("expected -timeout=4m30s in args, got: %s", joined)
	}
	if !strings.Contains(joined, "-timeout=1s") {
		t.Errorf("expected user-supplied -timeout=1s to survive, got: %s", joined)
	}
	firstIdx := strings.Index(joined, "-timeout=4m30s")
	lastIdx := strings.Index(joined, "-timeout=1s")
	if firstIdx < 0 || lastIdx < 0 || firstIdx > lastIdx {
		t.Errorf("auto-timeout must come before user override, got: %s", joined)
	}
	if args[0] != "test" || args[1] != "-json" {
		t.Errorf("expected `test -json` prefix, got: %v", args[:2])
	}
}
