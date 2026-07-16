package runtime

import (
	"testing"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
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
