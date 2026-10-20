package linters

import (
	"testing"

	"github.com/flanksource/gavel/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLintVerdict(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Lint verdict") }

var _ = Describe("Lint verdict and run diagnostics", func() {
	DescribeTable("fails only errors and execution failures", func(result *LinterResult, failed bool) {
		Expect(HasFailures([]*LinterResult{result})).To(Equal(failed))
	},
		Entry("warning", &LinterResult{Success: true, Violations: []models.Violation{{Severity: models.SeverityWarning}}}, false),
		Entry("info", &LinterResult{Success: true, Violations: []models.Violation{{Severity: models.SeverityInfo}}}, false),
		Entry("error", &LinterResult{Success: true, Violations: []models.Violation{{Severity: models.SeverityError}}}, true),
		Entry("unknown severity", &LinterResult{Success: true, Violations: []models.Violation{{Severity: "fatal"}}}, true),
		Entry("crash", &LinterResult{Error: "failed"}, true),
		Entry("timeout", &LinterResult{Success: true, TimedOut: true}, true),
		Entry("skipped", &LinterResult{Skipped: true}, false),
		Entry("nil", nil, false),
	)
	It("keeps fileless diagnostics when the user requests one file", func() {
		result := &LinterResult{WorkDir: GinkgoT().TempDir(), Violations: []models.Violation{{Severity: models.SeverityWarning}, {File: "other.go"}}}
		Expect(FilterViolationsByUserScope([]*LinterResult{result}, result.WorkDir, []string{"selected.go"})).To(Equal(1))
		Expect(result.Violations).To(Equal([]models.Violation{{Severity: models.SeverityWarning}}))
	})
})
