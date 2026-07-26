package runtime

import (
	"testing"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRuntimeCancellation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TODO runtime cancellation")
}

var _ = Describe("cancelled attempt projection", func() {
	It("records a cancelled prompt run instead of a failure", func() {
		state, phase, lifecycle, activity, reason := terminalState(&todos.ExecutionResult{
			Cancelled: true,
			Summary:   "run stopped by user",
		}, native.StepRun)

		Expect(state).To(Equal(captaindb.PromptRunStateCancelled))
		Expect(phase).To(Equal(captaindb.PromptRunPhaseFinished))
		Expect(lifecycle).To(Equal(captaindb.SessionLifecycleCancelled))
		Expect(activity).To(Equal(captaindb.SessionActivityIdle))
		Expect(reason).To(Equal("run stopped by user"))
	})
})

var _ = Describe("plan review status projection", func() {
	DescribeTable("projects plan approval without hiding terminal or active states",
		func(status native.IssueStatus, execution native.ExecutionState, approval captaindb.PlanApprovalState, expected types.Status) {
			Expect(todoStatusWithPlan(status, execution, approval)).To(Equal(expected))
		},
		Entry("open pending plan", native.StatusOpen, native.ExecutionIdle, captaindb.PlanApprovalPending, types.StatusReview),
		Entry("draft pending plan", native.StatusDraft, native.ExecutionIdle, captaindb.PlanApprovalPending, types.StatusReview),
		Entry("draft revision requested", native.StatusDraft, native.ExecutionIdle, captaindb.PlanApprovalRevisionRequested, types.StatusReview),
		Entry("draft approved plan", native.StatusDraft, native.ExecutionIdle, captaindb.PlanApprovalApproved, types.StatusPending),
		Entry("draft rejected plan", native.StatusDraft, native.ExecutionIdle, captaindb.PlanApprovalRejected, types.StatusPending),
		Entry("running open issue", native.StatusOpen, native.ExecutionRunning, captaindb.PlanApprovalPending, types.StatusInProgress),
		Entry("verified issue", native.StatusVerified, native.ExecutionIdle, captaindb.PlanApprovalPending, types.StatusVerified),
		Entry("closed issue", native.StatusClosed, native.ExecutionIdle, captaindb.PlanApprovalPending, types.StatusCompleted),
	)
})
