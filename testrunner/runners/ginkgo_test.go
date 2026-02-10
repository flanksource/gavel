package runners

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func TestGinkgoDetect(t *testing.T) {
	tmpDir := t.TempDir()

	// No ginkgo tests yet
	runner := NewGinkgo(tmpDir)
	found, err := runner.Detect(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected no ginkgo tests detected")
	}

	// Create a ginkgo test file
	testFile := filepath.Join(tmpDir, "suite_test.go")
	content := `package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Suite", func() {
	It("works", func() {
		Expect(true).To(BeTrue())
	})
})
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	found, err = runner.Detect(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected ginkgo test file to be detected")
	}
}

func TestGinkgoDetectOldImport(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with old ginkgo v1 import
	testFile := filepath.Join(tmpDir, "suite_test.go")
	content := `package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	runner := NewGinkgo(tmpDir)
	found, err := runner.Detect(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected old ginkgo import to be detected")
	}
}

func TestGinkgoDiscoverPackages(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	pkgDir := filepath.Join(tmpDir, "pkg")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Create a ginkgo test file
	content := `package pkg

import . "github.com/onsi/ginkgo/v2"
`
	if err := os.WriteFile(filepath.Join(pkgDir, "suite_test.go"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	runner := NewGinkgo(tmpDir)
	packages, err := runner.DiscoverPackages(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(packages) != 1 {
		t.Errorf("expected 1 package, got %d", len(packages))
	}

	if !strings.HasPrefix(packages[0], "./") {
		t.Errorf("expected relative path, got %s", packages[0])
	}
}

func TestGinkgoPackageHasTests(t *testing.T) {
	tmpDir := t.TempDir()
	pkgDir := filepath.Join(tmpDir, "pkg")

	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	runner := NewGinkgo(tmpDir)

	// Should not have ginkgo tests yet
	found, err := runner.PackageHasTests("./pkg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected no ginkgo tests")
	}

	// Add a ginkgo test file
	content := `package pkg

import . "github.com/onsi/ginkgo/v2"
`
	if err := os.WriteFile(filepath.Join(pkgDir, "test_test.go"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	found, err = runner.PackageHasTests("./pkg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected ginkgo test file to be found")
	}
}

func TestGinkgoName(t *testing.T) {
	runner := NewGinkgo("/tmp")
	if runner.Name() != parsers.Ginkgo {
		t.Errorf("expected framework Ginkgo, got %v", runner.Name())
	}
}

func TestGinkgoParserType(t *testing.T) {
	runner := NewGinkgo("/tmp")
	parser := runner.Parser()
	if parser.Name() != "ginkgo json" {
		t.Errorf("expected parser name 'ginkgo json', got %s", parser.Name())
	}
}
