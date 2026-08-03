package testrunner

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("build tags", func() {
	It("maps tags to Go frameworks without removing unsupported frameworks", func() {
		orchestrator := &TestOrchestrator{
			RunOptions: RunOptions{Tags: []string{"integration", "postgres"}},
			registry:   DefaultRegistry(GinkgoT().TempDir()),
		}

		frameworks, args, err := orchestrator.resolveFrameworkArgs(
			[]Framework{parsers.GoTest, parsers.Ginkgo, parsers.Jest},
			[]string{"-count=1"},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(frameworks).To(Equal([]Framework{parsers.GoTest, parsers.Ginkgo, parsers.Jest}))
		Expect(args[parsers.GoTest]).To(Equal([]string{"-count=1", "-tags=integration,postgres"}))
		Expect(args[parsers.Ginkgo]).To(Equal([]string{"-count=1", "--tags=integration,postgres"}))
		Expect(args[parsers.Jest]).To(Equal([]string{"-count=1"}))
	})

	It("rejects tags when no selected framework supports them", func() {
		orchestrator := &TestOrchestrator{
			RunOptions: RunOptions{Tags: []string{"integration"}},
			registry:   DefaultRegistry(GinkgoT().TempDir()),
		}

		_, _, err := orchestrator.resolveFrameworkArgs([]Framework{parsers.Jest, parsers.Vitest}, nil)
		Expect(err).To(MatchError(ContainSubstring("--tags is not supported")))
	})

	It("includes tags in the Go pre-build invocation", func() {
		Expect(goPreBuildArgs([]string{"./api", "./store"}, []string{"integration", "postgres"})).To(
			Equal([]string{"test", "-count=0", "-tags=integration,postgres", "./api", "./store"}),
		)
	})

	It("salts run-cache fingerprints with tags", func() {
		workDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(workDir, "go.mod"), []byte("module example.com/tags\n\ngo 1.24\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, "tags.go"), []byte("package tags\n"), 0o644)).To(Succeed())

		plain, err := newSelectorContext(workDir, nil)
		Expect(err).NotTo(HaveOccurred())
		tagged, err := newSelectorContext(workDir, []string{"integration"})
		Expect(err).NotTo(HaveOccurred())

		plainFingerprint, err := plain.hasher.Effective("example.com/tags")
		Expect(err).NotTo(HaveOccurred())
		taggedFingerprint, err := tagged.hasher.Effective("example.com/tags")
		Expect(err).NotTo(HaveOccurred())
		Expect(taggedFingerprint.Hex).NotTo(Equal(plainFingerprint.Hex))
	})
})
