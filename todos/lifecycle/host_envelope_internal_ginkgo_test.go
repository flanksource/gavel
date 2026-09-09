package lifecycle

import (
	"context"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// collect is where captain's result becomes the facts the outcomes read. These
// specs pin the one invariant that matters there: the run state the lifecycle
// classifies agrees with the state the runtime records for the prompt run.
var _ = ginkgo.Describe("collecting a finished run", func() {
	const completed = `{"summary":"Built it.","endStatus":"completed"}`

	var (
		host      *Host
		exec      *todos.ExecutorContext
		prepared  *preparedStep
		execution *todos.ExecutionResult
	)
	ginkgo.BeforeEach(func() {
		host = &Host{}
		exec = todos.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
		prepared = &preparedStep{
			definition: todoprompt.Definition{Name: "run", Class: types.ModeRun, Envelope: todoprompt.EnvelopeResult},
			class:      types.ModeRun, timeout: time.Minute,
		}
		execution = &todos.ExecutionResult{ExecutorName: "scripted"}
	})

	collect := func(out promptrun.Result) *StepOutcome {
		return host.collect(exec, &types.TODO{}, Step{Name: "run"}, prepared, dispatched{out: out, execution: execution}, time.Now())
	}

	ginkgo.It("reports a decoded envelope as a succeeded run with the loop's stop reason", func() {
		out := collect(promptrun.Result{
			Response: &api.Response{Text: completed},
			Loop:     &captainai.LoopResult{StopReason: "condition-met"},
		})

		gomega.Expect(out.Result.Run.State).To(gomega.Equal(RunSucceeded))
		gomega.Expect(out.Result.Run.StopReason).To(gomega.Equal("condition-met"))
		gomega.Expect(out.Result.Envelope.EndStatus).To(gomega.Equal("completed"))
		gomega.Expect(out.Execution.Success).To(gomega.BeTrue())
	})

	ginkgo.It("fails a run whose provider result was not a success, even with an envelope", func() {
		saw := false
		host.handleEvent(exec, captainai.Event{Kind: captainai.EventResult, Success: false}, execution, &types.TODO{}, &saw, todos.RunStartMetadata{})

		out := collect(promptrun.Result{Response: &api.Response{Text: completed}})

		gomega.Expect(out.Result.Run.State).To(gomega.Equal(RunFailed))
		gomega.Expect(out.Result.Run.Error).To(gomega.Equal("agent reported an unsuccessful result"))
		gomega.Expect(out.Execution.Success).To(gomega.BeFalse())
		gomega.Expect(out.Result.Envelope.EndStatus).To(gomega.Equal("completed"), "the envelope still travels for the outcomes to read")
	})

	ginkgo.It("fails a run whose provider reported an error event", func() {
		saw := false
		host.handleEvent(exec, captainai.Event{Kind: captainai.EventError, Error: "socket closed"}, execution, &types.TODO{}, &saw, todos.RunStartMetadata{})

		out := collect(promptrun.Result{Response: &api.Response{Text: completed}})

		gomega.Expect(out.Result.Run.State).To(gomega.Equal(RunFailed))
		gomega.Expect(out.Result.Run.Error).To(gomega.Equal("socket closed"))
	})

	ginkgo.It("fails a run whose envelope says it failed", func() {
		out := collect(promptrun.Result{Response: &api.Response{Text: `{"summary":"Could not build it.","endStatus":"failed"}`}})

		gomega.Expect(out.Result.Run.State).To(gomega.Equal(RunFailed))
		gomega.Expect(out.Result.Run.Error).To(gomega.Equal("Could not build it."))
		gomega.Expect(out.Result.Envelope.EndStatus).To(gomega.Equal("failed"))
		gomega.Expect(out.Execution.Success).To(gomega.BeFalse())
	})

	ginkgo.It("keeps a waiting run waiting", func() {
		out := collect(promptrun.Result{Response: &api.Response{
			Text: `{"summary":"Which database?","endStatus":"ask","questions":[{"text":"Which database?"}]}`,
		}})

		gomega.Expect(out.Result.Run.State).To(gomega.Equal(RunWaiting))
		gomega.Expect(out.Execution.Success).To(gomega.BeFalse())
	})

	ginkgo.Describe("the definition-of-done record", func() {
		ginkgo.BeforeEach(func() {
			prepared.request.Workflow = &api.Workflow{Verify: &api.Verify{Fixture: "# dod"}}
		})

		ginkgo.It("ran only when the report says it ran", func() {
			out := collect(promptrun.Result{Response: &api.Response{Text: completed}, Report: &api.VerifyReport{Ran: false}, Passed: true})

			gomega.Expect(out.Execution.DoD).NotTo(gomega.BeNil())
			gomega.Expect(out.Execution.DoD.Ran).To(gomega.BeFalse())
			gomega.Expect(out.Execution.DoD.Passed).To(gomega.BeFalse())
		})

		ginkgo.It("passed when the report ran and the final verdict passed", func() {
			report := &api.VerifyReport{Ran: true, Passed: true}
			out := collect(promptrun.Result{
				Response: &api.Response{Text: completed}, Report: report, Passed: true,
				Verdicts: []agent.VerifyResult{{Valid: true, Report: report, Iteration: 1}},
			})

			gomega.Expect(out.Execution.DoD.Ran).To(gomega.BeTrue())
			gomega.Expect(out.Execution.DoD.Passed).To(gomega.BeTrue())
			gomega.Expect(out.Result.Verify).To(gomega.Equal(report))
		})
	})
})
