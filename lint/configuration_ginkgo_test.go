package lint

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/linters/markdownlint"
	"github.com/flanksource/gavel/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nested lint configuration", func() {
	var root, child string
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		root = GinkgoT().TempDir()
		child = filepath.Join(root, "child")
		Expect(os.Mkdir(filepath.Join(root, ".git"), 0o755)).To(Succeed())
		Expect(os.Mkdir(child, 0o755)).To(Succeed())
		for _, dir := range []string{root, child} {
			Expect(os.WriteFile(filepath.Join(dir, "guide.md"), []byte("# Guide\n"), 0o600)).To(Succeed())
		}
	})
	It("partitions one native project at configuration boundaries", func() {
		Expect(os.WriteFile(filepath.Join(root, ".gavel.yaml"), []byte("lint: {options: {fix: true}}\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(child, ".gavel.yaml"), []byte("lint: {linters: {markdownlint: {options: {fix: false}, args: [--child]}}}\n"), 0o600)).To(Succeed())
		invocations, err := partitionInvocations([]linterInvocation{{linter: markdownlint.NewMarkdownlint(root), projectRoot: root}}, Options{WorkDir: root})
		Expect(err).NotTo(HaveOccurred())
		Expect(invocations).To(HaveLen(2))
		for _, inv := range invocations {
			opts, _, _, _, err := inv.configuration(Options{WorkDir: root, OptionFields: map[string]bool{}})
			Expect(err).NotTo(HaveOccurred())
			Expect(opts.Fix).To(Equal(inv.configDir == root))
			Expect(inv.files).To(HaveLen(1))
		}
	})
	It("applies nested ignore rules relative to their declaring directory", func() {
		Expect(os.WriteFile(filepath.Join(child, ".gavel.yaml"), []byte("lint: {ignore: [{rule: MD001, file: guide.md}]}\n"), 0o600)).To(Succeed())
		results := []*linters.LinterResult{{Linter: "markdownlint", WorkDir: root, Violations: []models.Violation{{File: "guide.md", Rule: &models.Rule{Method: "MD001"}}, {File: "child/guide.md", Rule: &models.Rule{Method: "MD001"}}}}}
		Expect(applyPostLintFilters(results, root, nil)).To(Succeed())
		Expect(results[0].Violations).To(HaveExactElements(WithTransform(func(v models.Violation) string { return v.File }, Equal("guide.md"))))
	})
	It("keeps files owned by a different linter's nested project", func() {
		Expect(os.WriteFile(filepath.Join(child, ".gavel.yaml"), []byte("lint: {options: {fix: false}}\n"), 0o600)).To(Succeed())
		other, exists := builtinLinterRegistry(child).Get("ruff")
		Expect(exists).To(BeTrue())
		invocations, err := partitionInvocations([]linterInvocation{
			{linter: markdownlint.NewMarkdownlint(root), projectRoot: root},
			{linter: other, projectRoot: child},
		}, Options{WorkDir: root})
		Expect(err).NotTo(HaveOccurred())
		Expect(invocations).To(HaveLen(3))
	})
})
