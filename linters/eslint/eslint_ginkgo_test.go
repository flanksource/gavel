package eslint

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestESLint(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ESLint")
}

var _ = Describe("ESLintMessage.ToViolation", func() {
	It("keeps the reported end line and column", func() {
		message := ESLintMessage{RuleId: "no-unused-vars", Severity: 2, Message: "unused", Line: 3, Column: 7, EndLine: 5, EndColumn: 12}
		violation := message.ToViolation("/workspace", "src/app.ts")
		Expect([]int{violation.Line, violation.Column, violation.EndLine, violation.EndColumn}).To(Equal([]int{3, 7, 5, 12}))
	})
})
