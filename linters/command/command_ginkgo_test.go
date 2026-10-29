package command

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCommand(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Command linter") }

var _ = Describe("Command linter", func() {
	var root, binary, argsFile string
	BeforeEach(func() {
		root = GinkgoT().TempDir()
		content, err := os.ReadFile("testdata/report.sh")
		Expect(err).NotTo(HaveOccurred())
		binary = filepath.Join(root, "rules-lint")
		Expect(os.WriteFile(binary, content, 0o700)).To(Succeed())
		argsFile = filepath.Join(root, "args")
		GinkgoT().Setenv("COMMAND_TEST_ARGS", argsFile)
	})
	run := func(report string, exit string, files ...string) (*Linter, []models.Violation, error) {
		GinkgoT().Setenv("COMMAND_TEST_REPORT", report)
		GinkgoT().Setenv("COMMAND_TEST_EXIT", exit)
		linter := New(root, "rules", verify.LintLinterConfig{Command: []string{binary, "lint", "rules", "--context", "dev"}, Includes: []string{"**/XMLData.xml"}})
		linter.SetOptions(linters.RunOptions{WorkDir: root, Files: files})
		found, err := linter.Run(commonsContext.NewContext(context.Background()), nil)
		return linter, found, err
	}

	It("runs the declared argv with JSON flags and the files, keeping the report's findings and counts", func() {
		report := `[{"linter":"oipa-transaction","success":true,"file_count":1,"rule_count":2,"violations":[{"file":"tr/Fee/XMLData.xml","line":12,"severity":"error"}]}]`
		linter, found, err := run(report, "1", "tr/Fee/XMLData.xml")
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(Equal([]models.Violation{{File: "tr/Fee/XMLData.xml", Line: 12, Severity: models.SeverityError}}))
		Expect(os.ReadFile(argsFile)).To(BeEquivalentTo("lint rules --context dev --format json tr/Fee/XMLData.xml\n"))
		Expect([]int{linter.GetFileCount(), linter.GetRuleCount()}).To(Equal([]int{1, 2}))
		Expect(linter.DefaultIncludes()).To(Equal([]string{"**/XMLData.xml"}))
	})

	It("fails on a report that disagrees with the exit status, keeping its findings", func() {
		_, found, err := run(`[{"linter":"x","success":true,"violations":[{"file":"a.xml","severity":"error"}]}]`, "0")
		Expect(err).To(MatchError(ContainSubstring("lint report disagrees with exit status")))
		Expect(found).To(HaveLen(1))
	})

	It("fails on output that is no report", func() {
		_, _, err := run("rules not found", "2")
		Expect(err).To(MatchError(ContainSubstring("decode lint report")))
	})

	It("declares a linter for each configured command, and none for a built-in's settings", func() {
		declared := Declared(root, verify.LintConfig{Linters: map[string]verify.LintLinterConfig{
			"rules": {Command: []string{"oipa-cli", "lint", "rules"}},
			"ruff":  {Args: []string{"--select", "E"}},
		}})
		Expect(declared).To(HaveLen(1))
		Expect(declared[0].Name()).To(Equal("rules"))
		Expect(declared[0].HasDefaultActivation(root)).To(BeTrue())
	})
})
