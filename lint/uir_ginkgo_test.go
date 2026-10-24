package lint

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/baseline"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/linters/uir"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UIR native configuration", func() {
	It("fails when the required UIR binary is missing and still previews a dry run", func() {
		root := GinkgoT().TempDir()
		GinkgoT().Setenv("PATH", "")
		_, _, err := resolveLinterExecutable(context.Background(), uir.New(root), root, root, true, false)
		Expect(err).To(MatchError(ContainSubstring("uir executable not found")))
		binary, _, err := resolveLinterExecutable(context.Background(), uir.New(root), root, root, true, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(binary).To(Equal("uir"))
	})
	It("schedules nested configs independently without inheriting their parent's rules", func() {
		root := GinkgoT().TempDir()
		nested := filepath.Join(root, "nested")
		Expect(os.MkdirAll(nested, 0o700)).To(Succeed())
		for _, directory := range []string{root, nested} {
			Expect(os.WriteFile(filepath.Join(directory, ".uir.yaml"), []byte("version: 1\n"), 0o600)).To(Succeed())
		}
		invocations := resolveLinterInvocations(uir.New(root), Options{WorkDir: root})
		Expect(invocations).To(HaveLen(2))
		Expect(invocations[0].linter).NotTo(BeIdenticalTo(invocations[1].linter))
		for _, invocation := range invocations {
			Expect(isInvocationCovered(invocation, []string{root})).To(BeFalse())
		}
	})
	It("activates only with configuration and registers the adapter", func() {
		root := GinkgoT().TempDir()
		adapter := uir.New(root)
		selected, _ := shouldSelectLinter(root, verify.GavelConfig{}, adapter, false)
		Expect(selected).To(BeFalse())
		Expect(os.WriteFile(filepath.Join(root, ".uir.yaml"), []byte("version: 1\n"), 0o600)).To(Succeed())
		selected, _ = shouldSelectLinter(root, verify.GavelConfig{}, adapter, false)
		Expect(selected).To(BeTrue())
		Expect(buildLinterRegistry(root).Has("uir")).To(BeTrue())
	})
	It("expands config changes to the project but filters ordinary file scopes after analysis", func() {
		root := GinkgoT().TempDir()
		config := filepath.Join(root, ".uir.yaml")
		Expect(os.WriteFile(config, []byte("version: 1\n"), 0o600)).To(Succeed())
		makeResult := func() *linters.LinterResult {
			return &linters.LinterResult{Linter: "uir", WorkDir: root, Success: true, Violations: []models.Violation{{File: "a.go"}, {File: "b.go"}, {Severity: models.SeverityWarning}}}
		}
		result := makeResult()
		Expect(applyPostLintFilters([]*linters.LinterResult{result}, root, []string{config})).To(Succeed())
		Expect(result.Violations).To(HaveLen(3))
		result = makeResult()
		Expect(applyPostLintFilters([]*linters.LinterResult{result}, root, []string{"a.go"})).To(Succeed())
		Expect(result.Violations).To(HaveLen(2))
	})
	It("uses existing baseline keys and retains execution failures", func() {
		finding := models.NewViolationBuilder().WithFile("a.go").WithMessage("too large").WithRuleFromLinter("uir", "body").Build()
		finding.Severity = models.SeverityError
		original := []*linters.LinterResult{{Linter: "uir", Success: true, Violations: []models.Violation{finding}}}
		current := []*linters.LinterResult{{Linter: "uir", Success: true, Violations: []models.Violation{finding}}}
		filtered := baseline.FilterNewViolations(current, baseline.ExtractViolationKeys(original))
		Expect(linters.HasFailures(filtered)).To(BeFalse())
		current[0].Success, current[0].Error = false, "execution failed"
		filtered = baseline.FilterNewViolations(current, baseline.ExtractViolationKeys(original))
		Expect(linters.HasFailures(filtered)).To(BeTrue())
	})
})
