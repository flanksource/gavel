package lifecycle_test

import (
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// RunOutcome is the terminal state Captain files for a finished step: the same
// reading of gavel's envelope the lifecycle's outcomes act on, so the run row
// and the todo never disagree about how the run ended.
var _ = Describe("RunOutcome", func() {
	const summary = "Built it."
	askJSON := map[string]any{"summary": "Which database?", "endStatus": "ask"}

	DescribeTable("maps a finished run onto Captain's terminal state",
		func(result *todos.ExecutionResult, verifyStep bool, want promptrun.Outcome) {
			Expect(lifecycle.RunOutcome(result, verifyStep)).To(Equal(want))
		},
		Entry("a missing result fails the run in generate",
			nil, false,
			promptrun.Outcome{State: captaindb.PromptRunStateFailed, Phase: captaindb.PromptRunPhaseGenerate, Error: "agent run returned no result"}),
		Entry("a completed envelope succeeds with its summary and output",
			&todos.ExecutionResult{Success: true, Summary: summary, EndStatus: types.EndCompleted, OutputJSON: map[string]any{"endStatus": "completed"}}, false,
			promptrun.Outcome{State: captaindb.PromptRunStateSucceeded, Phase: captaindb.PromptRunPhaseFinished, Text: summary, JSON: map[string]any{"endStatus": "completed"}}),
		Entry("an ask envelope parks the run waiting in generate, keeping the envelope a reader matches on",
			&todos.ExecutionResult{Success: true, Summary: "Which database?", EndStatus: types.EndAsk, OutputJSON: askJSON}, false,
			promptrun.Outcome{State: captaindb.PromptRunStateWaiting, Phase: captaindb.PromptRunPhaseGenerate, Text: "Which database?", JSON: askJSON}),
		Entry("a stop cancels the run, whatever its envelope said",
			&todos.ExecutionResult{Cancelled: true, EndStatus: types.EndAsk, Summary: todos.ErrExecutionCancelled.Error(), ErrorMessage: todos.ErrExecutionCancelled.Error()}, false,
			promptrun.Outcome{State: captaindb.PromptRunStateCancelled, Phase: captaindb.PromptRunPhaseFinished, Text: todos.ErrExecutionCancelled.Error(), Error: todos.ErrExecutionCancelled.Error()}),
		Entry("a failed envelope fails the run in generate with its summary as the reason",
			&todos.ExecutionResult{Success: false, Summary: "Could not build.", EndStatus: types.EndFailed}, false,
			promptrun.Outcome{State: captaindb.PromptRunStateFailed, Phase: captaindb.PromptRunPhaseGenerate, Text: "Could not build.", Error: "Could not build."}),
		Entry("a failed definition of done fails the run in verify",
			&todos.ExecutionResult{Success: true, Summary: summary, EndStatus: types.EndCompleted, DoD: &todos.DoDOutcome{Ran: true, Passed: false, Report: &api.VerifyReport{Ran: true}}}, false,
			promptrun.Outcome{State: captaindb.PromptRunStateFailed, Phase: captaindb.PromptRunPhaseVerify, Text: summary, Error: summary}),
		Entry("a verify step that errored fails in verify with its error",
			&todos.ExecutionResult{Success: false, ErrorMessage: "verification produced no report"}, true,
			promptrun.Outcome{State: captaindb.PromptRunStateFailed, Phase: captaindb.PromptRunPhaseVerify, Error: "verification produced no report"}),
		Entry("a run with no reason at all still names one",
			&todos.ExecutionResult{Success: false}, false,
			promptrun.Outcome{State: captaindb.PromptRunStateFailed, Phase: captaindb.PromptRunPhaseGenerate, Error: "agent run failed"}),
	)
})
