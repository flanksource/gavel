package runtime

import (
	"testing"

	capapi "github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("verification progress projection", func() {
	It("replaces only the live snapshot while preserving other run results", func() {
		original := map[string]any{
			"session":          map[string]any{"id": "session-1"},
			"definitionOfDone": map[string]any{"note": "keep", "progress": "stale"},
		}
		report := capapi.VerifyReport{Kind: "fixture", Iteration: 4, State: capapi.VerifyStateRunning}

		projected := progressResultJSON(original, report)

		Expect(projected).To(HaveKeyWithValue("session", original["session"]))
		Expect(projected["definitionOfDone"]).To(Equal(map[string]any{"note": "keep", "progress": report}))
		Expect(original["definitionOfDone"]).To(Equal(map[string]any{"note": "keep", "progress": "stale"}))
	})
})

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
		func(status native.IssueStatus, execution native.ExecutionState, step native.StepKind, approval captaindb.PlanApprovalState, expected types.Status) {
			Expect(todoStatusWithPlan(status, execution, step, approval)).To(Equal(expected))
		},
		Entry("open pending plan", native.StatusOpen, native.ExecutionIdle, native.StepPlan, captaindb.PlanApprovalPending, types.StatusReview),
		Entry("draft pending plan", native.StatusDraft, native.ExecutionIdle, native.StepPlan, captaindb.PlanApprovalPending, types.StatusReview),
		Entry("draft revision requested", native.StatusDraft, native.ExecutionIdle, native.StepPlan, captaindb.PlanApprovalRevisionRequested, types.StatusReview),
		Entry("failed plan with recovered content", native.StatusOpen, native.ExecutionFailed, native.StepPlan, captaindb.PlanApprovalPending, types.StatusReview),
		Entry("failed implementation", native.StatusOpen, native.ExecutionFailed, native.StepRun, captaindb.PlanApprovalPending, types.StatusFailed),
		Entry("draft approved plan", native.StatusDraft, native.ExecutionIdle, native.StepPlan, captaindb.PlanApprovalApproved, types.StatusPending),
		Entry("draft rejected plan", native.StatusDraft, native.ExecutionIdle, native.StepPlan, captaindb.PlanApprovalRejected, types.StatusPending),
		Entry("running open issue", native.StatusOpen, native.ExecutionRunning, native.StepPlan, captaindb.PlanApprovalPending, types.StatusInProgress),
		Entry("verified issue", native.StatusVerified, native.ExecutionIdle, native.StepPlan, captaindb.PlanApprovalPending, types.StatusVerified),
		Entry("closed issue", native.StatusClosed, native.ExecutionIdle, native.StepPlan, captaindb.PlanApprovalPending, types.StatusCompleted),
	)
})
