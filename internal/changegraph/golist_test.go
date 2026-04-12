package changegraph_test

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/gavel/internal/changegraph"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// buildFixtureModule creates a small multi-package Go module with the
// dependency chain a → b → c. It returns the absolute module directory.
//
//	example.com/fix/a imports example.com/fix/b
//	example.com/fix/b imports example.com/fix/c
//	example.com/fix/c has a testdata/ subdirectory with a fixture file
//
// This lets us assert that editing files in c ripples back to {a,b,c}, and
// that editing files outside c's declared GoFiles (under testdata/) still
// marks c dirty.
func buildFixtureModule() string {
	root := GinkgoT().TempDir()

	writeFile := func(rel, content string) {
		p := filepath.Join(root, rel)
		Expect(os.MkdirAll(filepath.Dir(p), 0o755)).To(Succeed())
		Expect(os.WriteFile(p, []byte(content), 0o644)).To(Succeed())
	}

	writeFile("go.mod", "module example.com/fix\n\ngo 1.21\n")
	writeFile("a/a.go", `package a

import "example.com/fix/b"

func A() string { return b.B() }
`)
	writeFile("b/b.go", `package b

import "example.com/fix/c"

func B() string { return c.C() }
`)
	writeFile("c/c.go", `package c

func C() string { return "c" }
`)
	writeFile("c/testdata/fixture.yaml", "value: original\n")

	return root
}

// runGoModTidy makes sure the fixture module is listable. `go list` will
// fail if go.sum is missing deps, but our fixture has no external deps, so
// tidy is effectively a no-op — we still run it to keep the test resilient
// to future go-version changes.
func runGoModTidy(dir string) {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "go mod tidy: %s", out)
}

var _ = Describe("Graph.AffectedPackages", func() {
	var root string
	var graph *changegraph.Graph

	BeforeEach(func() {
		root = buildFixtureModule()
		runGoModTidy(root)

		var err error
		graph, err = changegraph.Load(root)
		Expect(err).NotTo(HaveOccurred())
		Expect(graph.Packages).To(HaveKey("example.com/fix/a"))
		Expect(graph.Packages).To(HaveKey("example.com/fix/b"))
		Expect(graph.Packages).To(HaveKey("example.com/fix/c"))
	})

	It("propagates changes to a leaf package up to reverse deps", func() {
		fs := changegraph.NewFileSet()
		fs.Add("c/c.go")
		Expect(graph.AffectedPackages(fs)).To(ConsistOf(
			"example.com/fix/a",
			"example.com/fix/b",
			"example.com/fix/c",
		))
	})

	It("only marks the edited package when it has no reverse deps", func() {
		fs := changegraph.NewFileSet()
		fs.Add("a/a.go")
		Expect(graph.AffectedPackages(fs)).To(ConsistOf("example.com/fix/a"))
	})

	It("marks a package dirty when a file under its testdata/ changes", func() {
		fs := changegraph.NewFileSet()
		fs.Add("c/testdata/fixture.yaml")
		// testdata files are resolved via the ancestor walk and should
		// propagate to every reverse dep of c.
		Expect(graph.AffectedPackages(fs)).To(ConsistOf(
			"example.com/fix/a",
			"example.com/fix/b",
			"example.com/fix/c",
		))
	})

	It("marks every package dirty when go.mod changes", func() {
		fs := changegraph.NewFileSet()
		fs.Add("go.mod")
		Expect(graph.AffectedPackages(fs)).To(ConsistOf(
			"example.com/fix/a",
			"example.com/fix/b",
			"example.com/fix/c",
		))
	})

	It("marks every package dirty when a root-level cross-cutting file changes", func() {
		fs := changegraph.NewFileSet()
		// A file at the module root that is not part of any package and not
		// go.mod/go.sum. Should be treated as cross-cutting and bust all.
		fs.Add("Makefile")
		Expect(graph.AffectedPackages(fs)).To(ConsistOf(
			"example.com/fix/a",
			"example.com/fix/b",
			"example.com/fix/c",
		))
	})

	It("returns nil for an empty file set", func() {
		Expect(graph.AffectedPackages(changegraph.NewFileSet())).To(BeNil())
	})
})
