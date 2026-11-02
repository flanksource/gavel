package lifecycle

import (
	"context"
	"time"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/lifecycle"
	"github.com/flanksource/gavel/todos"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/tools"
	"github.com/flanksource/gavel/todos/types"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// collect is where captain's result becomes the facts the outcomes read. These
// specs pin the one invariant that matters there: the run state the lifecycle
// classifies agrees with the state the runtime records for the prompt run.
var _ = ginkgo.Describe("collecting a finished run", func() {
	const completed = `{"success":true,"event":"completed","comment":"Built it."}`
	runStep := lifecycle.Step{Name: "run", Events: []string{"completed", "blocked"}}

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
			definition: todoprompt.Definition{Name: "run", Class: types.ModeRun},
			class:      types.ModeRun, timeout: time.Minute,
		}
		execution = &todos.ExecutionResult{ExecutorName: "scripted"}
	})

	collect := func(out promptrun.Result) *StepOutcome {
		return host.collect(runStep, prepared, dispatched{out: out, execution: execution}, time.Now())
	}

	ginkgo.It("reports a decoded envelope as a succeeded run with the loop's stop reason", func() {
		out := collect(promptrun.Result{
			Response: &api.Response{Text: completed},
			Loop:     &captainai.LoopResult{StopReason: "condition-met"},
		})

		gomega.Expect(out.Result.Run.State).To(gomega.Equal(RunSucceeded))
		gomega.Expect(out.Result.Run.StopReason).To(gomega.Equal("condition-met"))
		gomega.Expect(out.Result.Envelope).To(gomega.Equal(lifecycle.Envelope{Success: true, Event: "completed", Comment: "Built it."}))
		gomega.Expect(out.Execution.Summary).To(gomega.Equal("Built it."))
		gomega.Expect(out.Execution.OutputJSON).To(gomega.Equal(map[string]any{"success": true, "event": "completed", "comment": "Built it."}))
		gomega.Expect(out.Execution.Success).To(gomega.BeTrue())
	})

	ginkgo.It("collect copies the run workspace onto the execution", func() {
		worktree := &api.WorktreeState{
			Repo: "/src/repo", Path: "/src/repo/.worktrees/shell-1", Branch: "shell/1",
			Base: "b000000", Setup: "5e70000", Head: "4ead000", Removed: true,
		}
		commits := []api.CommitRecord{{SHA: "c100000", Message: "feat: one"}, {SHA: "4ead000", Message: "fix: two"}}
		out := collect(promptrun.Result{Response: &api.Response{Text: completed, Workspace: &api.Workspace{
			Cwd: worktree.Path, Worktree: worktree, Commits: commits, Diff: "transient diff",
			Notices: []api.Notice{{Text: "committed c100000"}},
		}}})

		gomega.Expect(out.Execution.Workspace).To(gomega.Equal(&api.WorkspaceRecord{
			Cwd: worktree.Path, Worktree: worktree, Commits: commits,
		}))
	})

	ginkgo.It("records no workspace for a run that reported none", func() {
		out := collect(promptrun.Result{Response: &api.Response{Text: completed}})
		gomega.Expect(out.Execution.Workspace).To(gomega.BeNil())
	})

	ginkgo.It("uses Captain's final usage and cost even after result events", func() {
		saw := false
		host.handleEvent(exec, captainai.Event{Kind: captainai.EventResult, Success: true,
			Usage: &captainai.Usage{InputTokens: 20}, CostUSD: 2}, execution, &types.TODO{}, &saw, todos.RunStartMetadata{})
		out := collect(promptrun.Result{Response: &api.Response{Text: completed},
			Usage: captainai.Usage{InputTokens: 3, OutputTokens: 2}, CostUSD: 0.5})
		gomega.Expect(out.Execution.TokensUsed).To(gomega.Equal(5))
		gomega.Expect(out.Execution.CostUSD).To(gomega.Equal(0.5))
	})

	ginkgo.It("renders Captain approval IDs and terminal reasons in the session transcript", func() {
		saw := false
		event := captainai.Event{Kind: captainai.EventPermission, Tool: "Bash",
			ApprovalID: "approval-1", Input: map[string]any{"command": "pwd"}}
		host.handleEvent(exec, event, execution, &types.TODO{}, &saw, todos.RunStartMetadata{})
		gomega.Expect(exec.GetTranscript().Entries).To(gomega.HaveLen(2))
		gomega.Expect(exec.GetTranscript().Entries[0].Metadata).To(gomega.HaveKeyWithValue("approvalId", "approval-1"))
		gomega.Expect(exec.GetTranscript().Entries[0].Metadata).To(gomega.HaveKeyWithValue("input", event.Input))

		event.Reason = "expired"
		host.handleEvent(exec, event, execution, &types.TODO{}, &saw, todos.RunStartMetadata{})
		gomega.Expect(exec.GetTranscript().Entries).To(gomega.HaveLen(4))
		gomega.Expect(exec.GetTranscript().Entries[2].Content).To(gomega.Equal("approval expired: Bash"))
		gomega.Expect(exec.GetTranscript().Entries[2].Metadata).To(gomega.HaveKeyWithValue("reason", "expired"))
	})

	// The dashboard labels a request by its kind; the tool, input and approval id
	// every reader already uses stay beside it.
	ginkgo.It("carries the approval kind in the notification beside the existing detail", func() {
		saw := false
		var notified []todos.Notification
		exec = todos.NewExecutorContext(context.Background(), logger.StandardLogger(), &todos.UserInteraction{
			NotifyFunc: func(n todos.Notification) { notified = append(notified, n) },
		})
		input := map[string]any{"command": "make build"}
		event := captainai.Event{Kind: captainai.EventPermission, Tool: "exec_command", ApprovalID: "approval-2", Input: input,
			Request: &api.ApprovalRequest{Tool: "exec_command", Input: input, Kind: api.ApprovalKindCommand,
				Command: &api.CommandApproval{Command: "make build"}}}

		host.handleEvent(exec, event, execution, &types.TODO{}, &saw, todos.RunStartMetadata{})

		expected := map[string]any{"tool": "exec_command", "approvalId": "approval-2", "input": input, "kind": api.ApprovalKindCommand}
		gomega.Expect(exec.GetTranscript().Entries[0].Metadata).To(gomega.Equal(expected))
		gomega.Expect(notified).To(gomega.HaveLen(1))
		gomega.Expect(notified[0].Data).To(gomega.Equal(expected))
	})

	ginkgo.It("preserves Captain's cancellation classification in the recording callback", func() {
		classify := host.recordedOutcome(context.Background(), runStep, prepared,
			&stepInput{execution: execution}, time.Now())
		outcome, err := classify(promptrun.Result{Response: &api.Response{Text: completed}}, context.Canceled, true)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(outcome.State).To(gomega.Equal(captaindb.PromptRunStateCancelled))
		gomega.Expect(outcome.Phase).To(gomega.Equal(captaindb.PromptRunPhaseFinished))
	})

	// The agent's claim is half the verdict: an envelope that does not claim
	// success fails the prompt run even though the run itself produced it.
	ginkgo.It("files an envelope that claims no success as a failed run carrying its comment", func() {
		classify := host.recordedOutcome(context.Background(), runStep, prepared,
			&stepInput{execution: execution}, time.Now())
		outcome, err := classify(promptrun.Result{Response: &api.Response{
			Text: `{"success":false,"event":"blocked","comment":"The schema migration needs a DBA."}`,
		}}, nil, false)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(outcome.State).To(gomega.Equal(captaindb.PromptRunStateFailed))
		gomega.Expect(outcome.Error).To(gomega.Equal("The schema migration needs a DBA."))
	})

	ginkgo.It("fails a run whose agent asked natively instead of through todo_ask", func() {
		out := collect(promptrun.Result{Response: &api.Response{TerminalOutcome: &api.TerminalOutcome{
			Kind: api.TerminalOutcomeQuestions, Questions: []api.TerminalQuestion{{Text: "Which database?"}},
		}}})

		gomega.Expect(out.Result.Run.State).To(gomega.Equal(RunFailed))
		gomega.Expect(out.Result.Run.Error).To(gomega.ContainSubstring("todo prompts ask through todo_ask"))
		gomega.Expect(out.Execution.Success).To(gomega.BeFalse())
	})

	ginkgo.It("fails a run whose event is not one the step declares", func() {
		out := collect(promptrun.Result{Response: &api.Response{Text: `{"success":true,"event":"ask","comment":"Which database?"}`}})

		gomega.Expect(out.Result.Run.State).To(gomega.Equal(RunFailed))
		gomega.Expect(out.Result.Run.Error).To(gomega.ContainSubstring(`event "ask" is not one of completed, blocked`))
	})

	ginkgo.It("fails a run whose provider result was not a success, even with an envelope", func() {
		saw := false
		host.handleEvent(exec, captainai.Event{Kind: captainai.EventResult, Success: false}, execution, &types.TODO{}, &saw, todos.RunStartMetadata{})

		out := collect(promptrun.Result{Response: &api.Response{Text: completed}})

		gomega.Expect(out.Result.Run.State).To(gomega.Equal(RunFailed))
		gomega.Expect(out.Result.Run.Error).To(gomega.Equal("agent reported an unsuccessful result"))
		gomega.Expect(out.Execution.Success).To(gomega.BeFalse())
		gomega.Expect(out.Result.Envelope.Event).To(gomega.Equal("completed"), "the envelope still travels for the outcomes to read")
	})

	// A reply that answered the question in the wrong shape used to leave nothing
	// behind but the word "failed": the decode error replaced the only account of
	// what the agent said.
	ginkgo.It("keeps an undecodable response verbatim, alongside the decode error", func() {
		const prose = "I read the code and this TODO is already done."
		out := collect(promptrun.Result{Response: &api.Response{Text: prose}})

		gomega.Expect(out.Execution.ResponseText).To(gomega.Equal(prose))
		gomega.Expect(out.Execution.Envelope).To(gomega.BeNil(), "no envelope was captured")
		gomega.Expect(out.Result.Run.State).To(gomega.Equal(RunFailed))
		gomega.Expect(out.Result.Run.Error).To(gomega.Equal(out.Execution.ErrorMessage))
	})

	ginkgo.Describe("a plan-mode run", func() {
		planStep := lifecycle.Step{Name: "plan", Events: []string{"new", "updated", "unchanged"}}
		ginkgo.BeforeEach(func() {
			prepared.class, prepared.existingPlan = types.ModePlan, "1. Old plan"
		})

		ginkgo.It("takes the plan todo_plan recorded, with the envelope's event as its status", func() {
			run := &tools.Run{}
			defs, err := tools.For([]string{tools.SetTodos}, tools.Options{
				Provider: struct{ todos.Provider }{}, Subject: &types.TODO{ID: "t1"}, Run: run, PlanStep: true,
			})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			_, err = toolNamed(defs, "todo_plan").Handler(context.Background(), map[string]any{"content": "1. New plan"})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			out := host.collect(planStep, prepared, dispatched{out: promptrun.Result{Response: &api.Response{
				Text: `{"success":true,"event":"updated","comment":"Revised the plan."}`,
			}}, execution: execution, tools: run}, time.Now())

			gomega.Expect(out.Execution.Plan).To(gomega.Equal(&types.PlanResult{Status: types.PlanUpdated, Content: "1. New plan"}))
			gomega.Expect(out.Execution.ToolWrites).To(gomega.Equal([]todos.ToolWrite{{Name: "todo_plan", Ref: "t1"}}))
		})

		ginkgo.It("maps a native plan outcome onto the plan events", func() {
			out := host.collect(planStep, prepared, dispatched{out: promptrun.Result{Response: &api.Response{
				TerminalOutcome: &api.TerminalOutcome{Kind: api.TerminalOutcomePlan, Plan: &api.TerminalPlan{Content: "1. Old plan"}},
			}}, execution: execution}, time.Now())

			gomega.Expect(out.Result.Envelope.Event).To(gomega.Equal("unchanged"))
			gomega.Expect(out.Execution.Plan.Status).To(gomega.Equal(types.PlanUnchanged))
			gomega.Expect(out.Execution.Success).To(gomega.BeTrue())
		})
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

	ginkgo.It("synthesises the verify step's envelope from the report", func() {
		prepared.class = types.ModeVerify
		passed := true
		report := &api.VerifyReport{Ran: true, Passed: false, Reason: "1 of 2 criteria failed",
			Checklist: []api.VerifyChecklistItem{{Item: "Greets", Passed: &passed}, {Item: "Formats"}}}
		out := host.collect(lifecycle.Step{Name: StepVerify, Events: VerifyEvents}, prepared,
			dispatched{out: promptrun.Result{Report: report}, execution: execution}, time.Now())

		gomega.Expect(out.Result.Envelope).To(gomega.Equal(lifecycle.Envelope{
			Success: false, Event: "failed", Comment: "1 of 2 criteria failed\n\n- [x] Greets\n- [ ] Formats",
		}))
	})
})

func toolNamed(defs []api.ToolDefinition, name string) api.ToolDefinition {
	for _, def := range defs {
		if def.Name == name {
			return def
		}
	}
	ginkgo.Fail("no tool named " + name)
	return api.ToolDefinition{}
}
