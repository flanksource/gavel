package lint

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Command linters declared in .gavel.yaml", func() {
	var root, binary string
	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		root = GinkgoT().TempDir()
		content, err := os.ReadFile("../linters/command/testdata/report.sh")
		Expect(err).NotTo(HaveOccurred())
		binary = filepath.Join(root, "rules-lint")
		Expect(os.WriteFile(binary, content, 0o700)).To(Succeed())
		GinkgoT().Setenv("COMMAND_TEST_ARGS", filepath.Join(root, "args"))
		Expect(os.MkdirAll(filepath.Join(root, "tr", "Fee"), 0o700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "tr", "Fee", "XMLData.xml"), []byte("<Transaction/>\n"), 0o600)).To(Succeed())
	})
	declare := func(name string) {
		config := "lint:\n  linters:\n    " + name + ":\n      command: [" + binary + ", lint, rules]\n      includes: [\"**/XMLData.xml\"]\n"
		Expect(os.WriteFile(filepath.Join(root, ".gavel.yaml"), []byte(config), 0o600)).To(Succeed())
	}

	It("runs the declared command over the requested files and reports its findings under the declared name", func() {
		declare("rules")
		GinkgoT().Setenv("COMMAND_TEST_REPORT", `[{"linter":"oipa-transaction","success":true,"violations":[{"file":"tr/Fee/XMLData.xml","line":3,"severity":"warning"}]}]`)
		GinkgoT().Setenv("COMMAND_TEST_EXIT", "0")

		results, err := Execute(Options{WorkDir: root, Linters: []string{"rules"}, Files: []string{"tr/Fee/XMLData.xml"}})

		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveExactElements(SatisfyAll(
			HaveField("Linter", "rules"),
			HaveField("Success", true),
			HaveField("Violations", HaveExactElements(SatisfyAll(
				HaveField("File", HaveSuffix("tr/Fee/XMLData.xml")),
				HaveField("Line", 3),
				HaveField("Severity", models.SeverityWarning),
			))),
		)))
		Expect(linters.HasFailures(results)).To(BeFalse())
	})

	It("refuses a command declared under a built-in linter's name", func() {
		declare("ruff")
		_, err := Execute(Options{WorkDir: root, Linters: []string{"ruff"}})
		Expect(err).To(MatchError(ContainSubstring("lint.linters.ruff declares a command, but ruff is a built-in linter")))
	})
})
