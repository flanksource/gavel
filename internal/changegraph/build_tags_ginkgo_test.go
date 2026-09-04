package changegraph_test

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/internal/changegraph"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("tagged package graphs", func() {
	It("loads dependencies enabled by the requested build tags", func() {
		workDir := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(workDir, "app"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(workDir, "tagdep"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, "go.mod"), []byte("module example.com/tags\n\ngo 1.24\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, "app", "app.go"), []byte("package app\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, "app", "integration.go"), []byte("//go:build integration\n\npackage app\n\nimport _ \"example.com/tags/tagdep\"\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, "tagdep", "dep.go"), []byte("package tagdep\n"), 0o644)).To(Succeed())

		plain, err := changegraph.Load(workDir, nil)
		Expect(err).NotTo(HaveOccurred())
		tagged, err := changegraph.Load(workDir, []string{"integration"})
		Expect(err).NotTo(HaveOccurred())

		Expect(plain.Packages["example.com/tags/app"].Imports).NotTo(ContainElement("example.com/tags/tagdep"))
		Expect(tagged.Packages["example.com/tags/app"].Imports).To(ContainElement("example.com/tags/tagdep"))
	})
})
