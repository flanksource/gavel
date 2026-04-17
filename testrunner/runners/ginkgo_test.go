package runners

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner/parsers"
)

func writeGinkgoGoMod(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}
}

func TestGinkgoDetect(t *testing.T) {
	tmpDir := t.TempDir()
	writeGinkgoGoMod(t, tmpDir)

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
	writeGinkgoGoMod(t, tmpDir)

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

func TestGinkgoDetectSkipsNestedProjectRoots(t *testing.T) {
	tmpDir := t.TempDir()
	writeGinkgoGoMod(t, tmpDir)

	if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create root .git: %v", err)
	}

	nestedModule := filepath.Join(tmpDir, "nested-module")
	if err := os.MkdirAll(filepath.Join(nestedModule, "pkg"), 0o755); err != nil {
		t.Fatalf("failed to create nested module package: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedModule, "go.mod"), []byte("module example.com/nested\n"), 0o644); err != nil {
		t.Fatalf("failed to create nested go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedModule, "pkg", "suite_test.go"), []byte(`package pkg

import . "github.com/onsi/ginkgo/v2"
`), 0o644); err != nil {
		t.Fatalf("failed to create nested module ginkgo test: %v", err)
	}

	nestedRepo := filepath.Join(tmpDir, "nested-repo")
	if err := os.MkdirAll(filepath.Join(nestedRepo, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create nested repo .git: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(nestedRepo, "pkg"), 0o755); err != nil {
		t.Fatalf("failed to create nested repo package: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedRepo, "pkg", "suite_test.go"), []byte(`package pkg

import . "github.com/onsi/ginkgo/v2"
`), 0o644); err != nil {
		t.Fatalf("failed to create nested repo ginkgo test: %v", err)
	}

	runner := NewGinkgo(tmpDir)
	found, err := runner.Detect(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected nested module and repo ginkgo tests to be ignored from the parent root")
	}

	found, err = runner.Detect(nestedModule)
	if err != nil {
		t.Fatalf("unexpected error detecting nested module directly: %v", err)
	}
	if !found {
		t.Error("expected nested module ginkgo tests to be detected when starting inside that module")
	}

	found, err = runner.Detect(nestedRepo)
	if err != nil {
		t.Fatalf("unexpected error detecting nested repo directly: %v", err)
	}
	if !found {
		t.Error("expected nested repo ginkgo tests to be detected when starting inside that repo")
	}
}

func TestGinkgoDetectsGinkgoImportsWithoutGoMod(t *testing.T) {
	// Like GoTest.Detect, ginkgo detection is purely about the file-level
	// signal (ginkgo import in a _test.go). If the user runs gavel inside a
	// directory that isn't a module, `ginkgo run` will surface the error.
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "suite_test.go"), []byte(`package main

import . "github.com/onsi/ginkgo/v2"
`), 0o644); err != nil {
		t.Fatalf("failed to create ginkgo test file: %v", err)
	}

	runner := NewGinkgo(tmpDir)
	found, err := runner.Detect(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected ginkgo tests to be detected even without a go.mod")
	}

	packages, err := runner.DiscoverPackages(tmpDir, true)
	if err != nil {
		t.Fatalf("unexpected error discovering packages: %v", err)
	}
	if len(packages) != 1 {
		t.Fatalf("expected 1 ginkgo package, got %v", packages)
	}
}

func TestGinkgoDiscoverPackages(t *testing.T) {
	tmpDir := t.TempDir()
	writeGinkgoGoMod(t, tmpDir)

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
	packages, err := runner.DiscoverPackages(tmpDir, true)
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

func TestGinkgoDiscoverPackagesSkipsNestedProjectRoots(t *testing.T) {
	tmpDir := t.TempDir()
	writeGinkgoGoMod(t, tmpDir)

	if err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create root .git: %v", err)
	}

	rootPkg := filepath.Join(tmpDir, "rootpkg")
	if err := os.MkdirAll(rootPkg, 0o755); err != nil {
		t.Fatalf("failed to create root package: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rootPkg, "suite_test.go"), []byte(`package rootpkg

import . "github.com/onsi/ginkgo/v2"
`), 0o644); err != nil {
		t.Fatalf("failed to create root ginkgo test: %v", err)
	}

	nestedModule := filepath.Join(tmpDir, "nested-module")
	if err := os.MkdirAll(filepath.Join(nestedModule, "pkg"), 0o755); err != nil {
		t.Fatalf("failed to create nested module package: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedModule, "go.mod"), []byte("module example.com/nested\n"), 0o644); err != nil {
		t.Fatalf("failed to create nested go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedModule, "pkg", "suite_test.go"), []byte(`package pkg

import . "github.com/onsi/ginkgo/v2"
`), 0o644); err != nil {
		t.Fatalf("failed to create nested module ginkgo test: %v", err)
	}

	nestedRepo := filepath.Join(tmpDir, "nested-repo")
	if err := os.MkdirAll(filepath.Join(nestedRepo, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create nested repo .git: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(nestedRepo, "pkg"), 0o755); err != nil {
		t.Fatalf("failed to create nested repo package: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedRepo, "pkg", "suite_test.go"), []byte(`package pkg

import . "github.com/onsi/ginkgo/v2"
`), 0o644); err != nil {
		t.Fatalf("failed to create nested repo ginkgo test: %v", err)
	}

	runner := NewGinkgo(tmpDir)

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

func TestHasGinkgoImports(t *testing.T) {
	cases := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name: "dot import v2",
			content: `package p
import . "github.com/onsi/ginkgo/v2"
`,
			expected: true,
		},
		{
			name: "named import v2",
			content: `package p
import "github.com/onsi/ginkgo/v2"
`,
			expected: true,
		},
		{
			name: "v1 dot import",
			content: `package p
import . "github.com/onsi/ginkgo"
`,
			expected: true,
		},
		{
			name: "sub-package of v2",
			content: `package p
import "github.com/onsi/ginkgo/v2/reporters"
`,
			expected: true,
		},
		{
			name: "alias import",
			content: `package p
import gk "github.com/onsi/ginkgo/v2"
var _ = gk.Describe
`,
			expected: true,
		},
		{
			name: "no ginkgo",
			content: `package p
import "testing"

func TestX(t *testing.T) {}
`,
			expected: false,
		},
		{
			name: "ginkgo only in a raw-string literal (regression)",
			content: "package p\n" +
				"\nimport \"testing\"\n\n" +
				"func TestGinkgoImportDetection(t *testing.T) {\n" +
				"\tfixture := `package main\n" +
				"import . \"github.com/onsi/ginkgo/v2\"\n" +
				"var _ = Describe(\"x\", func() {})\n" +
				"`\n" +
				"\t_ = fixture\n" +
				"}\n",
			expected: false,
		},
		{
			name: "ginkgo only in a double-quoted string literal",
			content: `package p
import "testing"

func TestX(t *testing.T) {
	_ = "github.com/onsi/ginkgo/v2"
}
`,
			expected: false,
		},
		{
			name: "ginkgo only in a line comment",
			content: `package p
// we used to import "github.com/onsi/ginkgo/v2" here
import "testing"

func TestX(t *testing.T) {}
`,
			expected: false,
		},
		{
			name: "unrelated onsi package",
			content: `package p
import "github.com/onsi/gomega"
`,
			expected: false,
		},
		{
			name:     "parse error fails closed",
			content:  `this is not valid go`,
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "sample_test.go")
			if err := os.WriteFile(path, []byte(tc.content), 0o644); err != nil {
				t.Fatalf("write fixture: %v", err)
			}
			got := hasGinkgoImports(path)
			if got != tc.expected {
				t.Errorf("hasGinkgoImports(%q) = %v, want %v\ncontent:\n%s",
					tc.name, got, tc.expected, tc.content)
			}
		})
	}
}
