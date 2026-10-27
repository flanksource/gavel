package lint

import (
	"os"
	"path/filepath"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/linters"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gstruct"
)

func writeLintFixture(root string, files map[string]string) {
	for path, contents := range files {
		absolute := filepath.Join(root, path)
		Expect(os.MkdirAll(filepath.Dir(absolute), 0o755)).To(Succeed())
		Expect(os.WriteFile(absolute, []byte(contents), 0o755)).To(Succeed())
	}
}

var _ = Describe("Execute linter isolation", func() {
	var root string
	BeforeEach(func() {
		clicky.ClearGlobalTasks()
		DeferCleanup(clicky.ClearGlobalTasks)
		root = GinkgoT().TempDir()
		Expect(os.Mkdir(filepath.Join(root, ".git"), 0o755)).To(Succeed())
		GinkgoT().Setenv("PATH", GinkgoT().TempDir())
		GinkgoT().Setenv("CI", "false")
	})

	It("does not schedule globally configured uir for an explicit file it cannot lint", func() {
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		writeLintFixture(home, map[string]string{".uir.yaml": "version: 1\n"})
		writeLintFixture(root, map[string]string{"report.mdx": "# Report\n"})

		results, err := Execute(Options{WorkDir: root, Files: []string{"report.mdx"}, Timeout: "5s"})

		Expect(err).NotTo(HaveOccurred())
		Expect(results).NotTo(ContainElement(HaveField("Linter", "uir")))
	})

	It("reports a missing required executable on that linter while other linters still run", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		writeLintFixture(root, map[string]string{
			".uir.yaml":                "version: 1\n",
			"sample.go":                "package sample\n",
			"package.json":             "{}\n",
			".oxlintrc.json":           "{}\n",
			"src/index.ts":             "export const value = 1;\n",
			"node_modules/.bin/oxlint": "#!/bin/sh\nprintf '%s\\n' '{\"diagnostics\":[],\"number_of_files\":1,\"number_of_rules\":1}'\n",
		})

		results, err := Execute(Options{WorkDir: root, Linters: []string{"uir", "oxlint"}, Timeout: "5s"})

		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(ConsistOf(
			gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Linter":  Equal("uir"),
				"WorkDir": Equal(root),
				"Success": BeFalse(),
				"Error":   ContainSubstring("uir executable not found on PATH"),
			})),
			gstruct.PointTo(gstruct.MatchFields(gstruct.IgnoreExtras, gstruct.Fields{
				"Linter":  Equal("oxlint"),
				"Success": BeTrue(),
				"Error":   BeEmpty(),
			})),
		))
		Expect(linters.HasFailures(results)).To(BeTrue())
	})
})
