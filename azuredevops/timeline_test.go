package azuredevops

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Azure timeline normalization", func() {
	It("sorts jobs and tasks and selects their latest attempt", func() {
		records := latestRecords([]timelineRecord{{ID: "old-job", Identifier: "Test", Type: "Job", Order: 1, Attempt: 1}, {ID: "new-job", Identifier: "Test", Type: "Job", Order: 1, Attempt: 2}, {ID: "task-b", Type: "Task", ParentID: "new-job", Order: 3}, {ID: "task-a", Type: "Task", ParentID: "new-job", Order: 2}, {ID: "deploy", Identifier: "Deploy", Type: "Job", Order: 4}})
		var ids []string
		for _, r := range records {
			ids = append(ids, r.ID)
		}
		Expect(ids).To(Equal([]string{"new-job", "task-a", "task-b", "deploy"}))
	})
	It("rejects unknown execution results instead of marking them successful", func() {
		_, _, err := executionState("completed", "unexpected")
		Expect(err).To(MatchError(ContainSubstring("unexpected Azure execution state")))
	})
	DescribeTable("preserves native partial and cancelled results", func(result, want string) {
		status, conclusion, err := executionState("completed", result)
		Expect(err).NotTo(HaveOccurred())
		Expect(status).To(Equal("COMPLETED"))
		Expect(conclusion).To(Equal(want))
	}, Entry("cancelled", "canceled", "CANCELLED"), Entry("partially succeeded build", "partiallySucceeded", "PARTIAL_SUCCESS"), Entry("task issues", "succeededWithIssues", "PARTIAL_SUCCESS"), Entry("skipped", "skipped", "SKIPPED"))
})
