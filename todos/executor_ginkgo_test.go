package todos

import (
	"context"
	"testing"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestExecutorCancellation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TODO executor cancellation")
}

var _ = Describe("cancelled execution", func() {
	It("persists the attempt without marking the todo failed or committing", func() {
		provider := &recordingProvider{}
		runner := NewTODOExecutor(".", stoppedExecutor{}, "", provider)
		todo := &types.TODO{ID: "todo-1", FilePath: "todo-1", TODOFrontmatter: types.TODOFrontmatter{Title: "Stop this run"}}

		result, err := runner.Execute(NewExecutorContext(context.Background(), logger.StandardLogger(), nil), todo)

		Expect(err).To(MatchError(context.Canceled))
		Expect(result.Cancelled).To(BeTrue())
		Expect(provider.saveCalls).To(Equal(1))
		Expect(todo.Status).To(Equal(types.StatusPending))
		Expect(todo.Attempts).To(Equal(1))
		Expect(ShouldCommitAfter(&ExecutionResult{Success: true, Cancelled: true}, true)).To(BeFalse())
	})
})

type stoppedExecutor struct{}

func (stoppedExecutor) Name() string { return "headless-test" }

func (stoppedExecutor) Execute(*ExecutorContext, *types.TODO) (*ExecutionResult, error) {
	return &ExecutionResult{ExecutorName: "headless-test", Cancelled: true, ErrorMessage: "run stopped by user"}, context.Canceled
}
