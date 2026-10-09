package lifecycle

import (
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
		source   = "captain"
		summary  = "Migrated the staging database."
		question = "Should the staging data be anonymised first?"
	)
	runDefinition := todoprompt.Definition{Name: "run", Class: types.ModeRun, Envelope: todoprompt.EnvelopeResult}
	runStep := Step{Name: "run"}
	runID := uuid.MustParse("0199f0aa-0000-7000-8000-00000000c0de")

	ginkgo.It("reads a plain-text success as a completed run", func() {
		out, err := (&Host{}).externalOutcome(runStep, runDefinition, ExternalTurn{
			Source: source, PromptRunID: runID, State: RunSucceeded, Summary: summary,
		})

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(out.Source).To(gomega.Equal(source))
		gomega.Expect(out.Admission.PromptRunID).To(gomega.Equal(runID))
		gomega.Expect(out.Result.Run.State).To(gomega.Equal(RunSucceeded))
		gomega.Expect(out.Result.Envelope).To(gomega.Equal(Envelope{Summary: summary, EndStatus: string(types.EndCompleted)}))
		gomega.Expect(out.Execution.Success).To(gomega.BeTrue())
		gomega.Expect(out.Execution.EndStatus).To(gomega.Equal(types.EndCompleted))
		gomega.Expect(out.Execution.Summary).To(gomega.Equal(summary))
		gomega.Expect(out.Execution.ExecutorName).To(gomega.Equal(source))
	})

	ginkgo.It("decodes a returned ask envelope into a waiting run with its questions", func() {
		out, err := (&Host{}).externalOutcome(runStep, runDefinition, ExternalTurn{
			Source: source, PromptRunID: runID, State: RunSucceeded,
			Envelope: map[string]any{
				"summary": "Waiting on one more answer.", "endStatus": "ask",
				"questions": []any{map[string]any{"text": question}},
			},
		})

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(out.Result.Run.State).To(gomega.Equal(RunWaiting))
		gomega.Expect(out.Result.Envelope.EndStatus).To(gomega.Equal(string(types.EndAsk)))
		gomega.Expect(out.Execution.EndStatus).To(gomega.Equal(types.EndAsk))
		gomega.Expect(out.Execution.Questions).To(gomega.Equal([]types.AgentQuestion{{Text: question}}))
	})

	ginkgo.It("fails a turn whose envelope does not decode", func() {
		out, err := (&Host{}).externalOutcome(runStep, runDefinition, ExternalTurn{
			Source: source, PromptRunID: runID, State: RunSucceeded,
			Envelope: map[string]any{"endStatus": "ask"},
		})

		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(out.Result.Run.State).To(gomega.Equal(RunFailed))
		gomega.Expect(out.Execution.Success).To(gomega.BeFalse())
		gomega.Expect(out.Execution.ErrorMessage).To(gomega.ContainSubstring("summary"))
	})

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
