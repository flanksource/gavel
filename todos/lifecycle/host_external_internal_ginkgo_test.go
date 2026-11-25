package lifecycle

import (
	"github.com/flanksource/gavel/lifecycle"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// externalOutcome turns a turn gavel did not run into the same two records a
// dispatched step produces, so the lifecycle's outcomes classify it and
// OnOutcome persists it exactly like one of gavel's own.
var _ = ginkgo.Describe("an external turn's outcome", func() {
	const (
		source  = "captain"
		comment = "Migrated the staging database."
	)
	runDefinition := todoprompt.Definition{Name: "run", Class: types.ModeRun}
	runStep := lifecycle.Step{Name: "run", Events: []string{"completed", "blocked"}}
	runID := uuid.MustParse("0199f0aa-0000-7000-8000-00000000c0de")

	ginkgo.It("decodes the returned envelope like a dispatched run's", func() {
		out, err := (&Host{}).externalOutcome(runStep, runDefinition, ExternalTurn{
			Source: source, PromptRunID: runID, State: RunSucceeded,
			Envelope: map[string]any{"success": true, "event": "completed", "comment": comment},
		})

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(out.Source).To(gomega.Equal(source))
		gomega.Expect(out.Admission.PromptRunID).To(gomega.Equal(runID))
		gomega.Expect(out.Result.Run.State).To(gomega.Equal(RunSucceeded))
		gomega.Expect(out.Result.Envelope).To(gomega.Equal(lifecycle.Envelope{
			Success: true, Event: "completed", Comment: comment,
			Extra: map[string]any{"success": true, "event": "completed", "comment": comment},
		}))
		gomega.Expect(out.Execution.Success).To(gomega.BeTrue())
		gomega.Expect(out.Execution.Summary).To(gomega.Equal(comment))
		gomega.Expect(out.Execution.ExecutorName).To(gomega.Equal(source))
	})

	ginkgo.DescribeTable("fails a successful turn without a valid envelope",
		func(envelope map[string]any, message string) {
			out, err := (&Host{}).externalOutcome(runStep, runDefinition, ExternalTurn{
				Source: source, PromptRunID: runID, State: RunSucceeded, Summary: comment, Envelope: envelope,
			})

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(out.Result.Run.State).To(gomega.Equal(RunFailed))
			gomega.Expect(out.Execution.Success).To(gomega.BeFalse())
			gomega.Expect(out.Execution.ErrorMessage).To(gomega.ContainSubstring(message))
		},
		ginkgo.Entry("plain text only", nil, "returned no result envelope"),
		ginkgo.Entry("an undeclared event", map[string]any{"success": true, "event": "ask", "comment": "Which?"}, `event "ask"`),
	)

	ginkgo.DescribeTable("carries a turn that did not succeed",
		func(state, reported string, cancelled bool) {
			out, err := (&Host{}).externalOutcome(runStep, runDefinition, ExternalTurn{
				Source: source, PromptRunID: runID, State: state, Error: reported,
			})

			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(out.Result.Run.State).To(gomega.Equal(state))
			gomega.Expect(out.Result.Run.Error).To(gomega.Equal(reported))
			gomega.Expect(out.Execution.Success).To(gomega.BeFalse())
			gomega.Expect(out.Execution.Cancelled).To(gomega.Equal(cancelled))
			gomega.Expect(out.Execution.ErrorMessage).To(gomega.Equal(reported))
		},
		ginkgo.Entry("failed", RunFailed, "provider exited 1", false),
		ginkgo.Entry("cancelled", RunCancelled, "stopped from Captain", true),
	)

	ginkgo.It("names a failure the turn did not explain", func() {
		out, err := (&Host{}).externalOutcome(runStep, runDefinition, ExternalTurn{Source: source, PromptRunID: runID, State: RunFailed})

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(out.Execution.ErrorMessage).To(gomega.Equal("the captain turn ended failed"))
	})

	ginkgo.DescribeTable("rejects a turn it cannot classify",
		func(turn ExternalTurn, message string) {
			_, err := (&Host{}).externalOutcome(runStep, runDefinition, turn)

			gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(message)))
		},
		ginkgo.Entry("a turn still waiting", ExternalTurn{Source: source, PromptRunID: runID, State: RunWaiting}, `state "waiting"`),
		ginkgo.Entry("no source", ExternalTurn{PromptRunID: runID, State: RunSucceeded}, "source"),
		ginkgo.Entry("no prompt run", ExternalTurn{Source: source, State: RunSucceeded}, "prompt run"),
	)
})
