package runners

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestGoTestDetect(t *testing.T) {
	tmpDir := t.TempDir()

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

func TestGoTestDiscoverPackages(t *testing.T) {
	tmpDir := t.TempDir()

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
	packages, err := runner.DiscoverPackages(tmpDir)
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
	packages, err := runner.DiscoverPackages(tmpDir)
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
