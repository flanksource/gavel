package lint

import (
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/linters/eslint"
	"github.com/flanksource/gavel/linters/golangci"
	"github.com/flanksource/gavel/linters/uir"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lint source selection", func() {
	DescribeTable("maps build tags to native commands", func(factory func(string) linters.Linter, expected string) {
		root := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		_, args, err := linters.PrepareCommand(factory(root), linters.RunOptions{WorkDir: root, Tags: []string{"integration", "feature"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(args).To(ContainElement(expected))
	},
		Entry("golangci", func(root string) linters.Linter { return golangci.NewGolangciLint(root) }, "--build-tags=integration,feature"),
		Entry("UIR", func(root string) linters.Linter { return uir.New(root) }, "--tags=integration,feature"),
	)

	It("rejects unsupported selections and permits mixed selections", func() {
		registry := linters.NewRegistry()
		registry.Register(eslint.NewESLint(""))
		registry.Register(golangci.NewGolangciLint(""))
		Expect(validateTags(registry, []string{"eslint"}, []string{"integration"})).To(MatchError(ContainSubstring("--tags is not supported")))
		Expect(validateTags(registry, []string{"eslint", "golangci-lint"}, []string{"integration"})).To(Succeed())
		Expect(validateTags(registry, []string{"eslint"}, nil)).To(Succeed())
	})
})
