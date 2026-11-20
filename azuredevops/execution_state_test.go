// Tests Azure DevOps execution states as the dashboard spells them: native partial and cancelled
// results kept, unknown results refused.
package azuredevops

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Azure execution states", func() {
	It("rejects unknown execution results instead of marking them successful", func() {
		_, _, err := ExecutionState("completed", "unexpected")
		Expect(err).To(MatchError(ContainSubstring("unexpected Azure execution state")))
	})
	DescribeTable("preserves native partial and cancelled results", func(result, want string) {
		status, conclusion, err := ExecutionState("completed", result)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal("COMPLETED"))
		Expect(conclusion).To(Equal(want))
	}, Entry("cancelled", "canceled", "CANCELLED"), Entry("partially succeeded build", "partiallySucceeded", "PARTIAL_SUCCESS"), Entry("task issues", "succeededWithIssues", "PARTIAL_SUCCESS"), Entry("skipped", "skipped", "SKIPPED"))
})
