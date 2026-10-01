package runtime

import (
	"context"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Settling: once the Captain turn that answered a parked run is over, the gavel
// run it continued is finished through the lifecycle's own outcomes.
var _ = Describe("settling turns Captain ran for a parked todo", func() {
	var f *answerFixture

	BeforeEach(func(ctx SpecContext) { f = newAnswerFixture(ctx) })

	It("settles a turn that finished with plain text as a completed run", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Answered and finished in Captain")
		f.answerInCaptain(ctx, parked)
		f.insertMessage(ctx, parked.executionSessionID, "assistant", textParts(followupSummary))
		f.captainTurn(ctx, parked, captaindb.PromptRunStateSucceeded, "")

		Expect(f.concurrentGets(ctx, parked.todo.ID)).To(HaveEach(types.StatusPending))

		read, err := f.reader.Get(ctx, parked.todo.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(read.Questions).To(BeEmpty())
		Expect(read.LastRunSummary).To(Equal(followupSummary), "a chat turn persists no result text; the transcript carries it")
		Expect(f.gavelRun(ctx, parked).State).To(Equal(captaindb.PromptRunStateSucceeded))
		outcomes := f.eventsOf(ctx, parked.todo, lifecycle.EventLifecycleOutcome)
		Expect(outcomes).To(HaveLen(2), "the ask, then the Captain turn")
		Expect(payloadOf(outcomes[1])).To(And(
			HaveKeyWithValue("source", "captain"),
			HaveKeyWithValue("status", string(types.StatusPending)),
			HaveKeyWithValue("promptRunId", parked.runID.String()),
		))
		Expect(f.eventsOf(ctx, parked.todo, native.EventAskAnswered)).To(HaveLen(1))
	})

	It("settles a turn while the session's process is still alive", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Parked TUI stays alive")
		f.liveProcess(ctx, parked.executionSessionID)
		f.answerInCaptain(ctx, parked)
		f.insertMessage(ctx, parked.executionSessionID, "assistant", textParts(followupSummary))
		f.captainTurn(ctx, parked, captaindb.PromptRunStateSucceeded, "")

		read, err := f.reader.Get(ctx, parked.todo.ID)

		Expect(err).NotTo(HaveOccurred())
		Expect(read.Status).To(Equal(types.StatusPending))
	})

	It("waits for the turn's reply to be ingested before classifying it", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Reply not ingested yet")
		f.answerInCaptain(ctx, parked)
		f.captainTurn(ctx, parked, captaindb.PromptRunStateSucceeded, "")

		read, err := f.reader.Get(ctx, parked.todo.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(read.Status).To(Equal(types.StatusInProgress))
		Expect(f.gavelRun(ctx, parked).State).To(Equal(captaindb.PromptRunStateWaiting))

		f.insertMessage(ctx, parked.executionSessionID, "assistant", structuredOutput(askEnvelope(followupQuestion)))
		read, err = f.reader.Get(ctx, parked.todo.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(read.Status).To(Equal(types.StatusAsk))
	})

	It("defers a turn Captain still records as in progress", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Captain turn still running")
		f.answerInCaptain(ctx, parked)
		f.insertMessage(ctx, parked.executionSessionID, "assistant", textParts(followupSummary))
		_, err := f.dispatcher.Captain().CreatePromptRun(ctx, captaindb.CreatePromptRunInput{
			SessionID: parked.executionSessionID, Origin: "captain", PromptMarkdown: captainAnswer,
		})
		Expect(err).NotTo(HaveOccurred())

		read, err := f.reader.Get(ctx, parked.todo.ID)

		Expect(err).NotTo(HaveOccurred())
		Expect(read.Status).To(Equal(types.StatusInProgress))
		Expect(f.gavelRun(ctx, parked).State).To(Equal(captaindb.PromptRunStateWaiting))
	})

	// Captain writes a turn's row when the turn ends; the monitor may ingest the
	// answer that started it later still. The turn is bound by what it sent.
	It("settles a turn whose answer was ingested after the turn was recorded", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Answer ingested after the turn row")
		f.captainTurn(ctx, parked, captaindb.PromptRunStateSucceeded, "")
		f.answerInCaptain(ctx, parked)
		f.insertMessage(ctx, parked.executionSessionID, "assistant", textParts(followupSummary))

		read, err := f.reader.Get(ctx, parked.todo.ID)

		Expect(err).NotTo(HaveOccurred())
		Expect(read.Status).To(Equal(types.StatusPending))
		Expect(f.gavelRun(ctx, parked).State).To(Equal(captaindb.PromptRunStateSucceeded))
	})

	It("does not bind a Captain turn that sent something other than the answer", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Unrelated Captain turn")
		f.answerInCaptain(ctx, parked)
		f.insertMessage(ctx, parked.executionSessionID, "assistant", textParts(followupSummary))
		f.captainTurnFor(ctx, parked, "Summarise this session instead.", captaindb.PromptRunStateSucceeded, "")

		read, err := f.reader.Get(ctx, parked.todo.ID)

		Expect(err).NotTo(HaveOccurred())
		Expect(read.Status).To(Equal(types.StatusInProgress))
		Expect(f.gavelRun(ctx, parked).State).To(Equal(captaindb.PromptRunStateWaiting))
	})

	It("parks the todo again when the turn asked new questions", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Answered in Captain and asked again")
		f.answerInCaptain(ctx, parked)
		f.insertMessage(ctx, parked.executionSessionID, "assistant", structuredOutput(askEnvelope(followupQuestion)))
		f.captainTurn(ctx, parked, captaindb.PromptRunStateSucceeded, "")

		read, err := f.reader.Get(ctx, parked.todo.ID)

		Expect(err).NotTo(HaveOccurred())
		Expect(read.Status).To(Equal(types.StatusAsk))
		Expect(read.Questions).To(HaveLen(1))
		Expect(read.Questions[0].Text).To(Equal(followupQuestion))
		Expect(f.gavelRun(ctx, parked).State).To(Equal(captaindb.PromptRunStateWaiting))

		again, err := f.reader.Get(ctx, parked.todo.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(again.Status).To(Equal(types.StatusAsk), "the new ask is not answered by the message that answered the old one")
		Expect(f.eventsOf(ctx, parked.todo, lifecycle.EventLifecycleOutcome)).To(HaveLen(2))
	})

	It("reads a final reply that is itself an envelope", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Envelope as reply text")
		f.answerInCaptain(ctx, parked)
		f.insertMessage(ctx, parked.executionSessionID, "assistant",
			textParts(`{"summary":"One more answer needed.","endStatus":"ask","questions":[{"text":"`+followupQuestion+`"}]}`))
		f.captainTurn(ctx, parked, captaindb.PromptRunStateSucceeded, "")

		read, err := f.reader.Get(ctx, parked.todo.ID)

		Expect(err).NotTo(HaveOccurred())
		Expect(read.Status).To(Equal(types.StatusAsk))
		Expect(read.Questions[0].Text).To(Equal(followupQuestion))
	})

	DescribeTable("settles a turn that did not succeed, with no reply and its answer ingested late",
		func(ctx SpecContext, state captaindb.PromptRunState, want types.Status) {
			parked := f.parkRunAsk(ctx, "Captain turn "+string(state))
			f.captainTurn(ctx, parked, state, "the Captain turn ended "+string(state))
			f.answerInCaptain(ctx, parked)

			read, err := f.reader.Get(ctx, parked.todo.ID)

			Expect(err).NotTo(HaveOccurred())
			Expect(read.Status).To(Equal(want))
			Expect(f.gavelRun(ctx, parked).State).To(Equal(state))
		},
		Entry("failed", captaindb.PromptRunStateFailed, types.StatusFailed),
		Entry("cancelled", captaindb.PromptRunStateCancelled, types.StatusPending),
	)

	It("finishes a settlement whose reader went away", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Reader cancelled mid-settle")
		f.answerInCaptain(ctx, parked)
		f.insertMessage(ctx, parked.executionSessionID, "assistant", textParts(followupSummary))
		f.captainTurn(ctx, parked, captaindb.PromptRunStateSucceeded, "")
		candidates, err := f.reader.askCandidates(ctx, reconcileScope{IssueID: uuid.MustParse(parked.todo.ID)})
		Expect(err).NotTo(HaveOccurred())
		Expect(candidates).To(HaveLen(1))
		answer, err := f.reader.detectExternalAnswer(ctx, candidates[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(answer).NotTo(BeNil())
		cancelled, cancel := context.WithCancel(ctx)
		cancel()

		Expect(f.reader.settleExternalTurn(cancelled, candidates[0], *answer, f.options.RootPath)).To(Succeed())

		Expect(f.gavelRun(ctx, parked).State).To(Equal(captaindb.PromptRunStateSucceeded))
		Expect(f.eventsOf(ctx, parked.todo, lifecycle.EventLifecycleOutcome)).To(HaveLen(2))
	})

	It("fails the read of a turn its step cannot classify once, and settles it", func(ctx SpecContext) {
		parked := f.parkAsk(ctx, "Plan answered with prose", "plan", types.ModePlan)
		f.answerInCaptain(ctx, parked)
		f.insertMessage(ctx, parked.executionSessionID, "assistant", textParts(followupSummary))
		f.captainTurn(ctx, parked, captaindb.PromptRunStateSucceeded, "")

		_, err := f.reader.Get(ctx, parked.todo.ID)
		Expect(err).To(MatchError(ContainSubstring("no outcome matched")))

		read, err := f.reader.Get(ctx, parked.todo.ID)
		Expect(err).NotTo(HaveOccurred(), "the turn was recorded; later reads have nothing left to reconcile")
		Expect(f.gavelRun(ctx, parked).State).To(Equal(captaindb.PromptRunStateSucceeded))
		Expect(read.ID).To(Equal(parked.todo.ID))
	})

	It("keeps listing the workspace when one todo cannot be reconciled", func(ctx SpecContext) {
		broken := f.parkAsk(ctx, "Plan answered with prose", "plan", types.ModePlan)
		f.answerInCaptain(ctx, broken)
		f.insertMessage(ctx, broken.executionSessionID, "assistant", textParts(followupSummary))
		f.captainTurn(ctx, broken, captaindb.PromptRunStateSucceeded, "")
		answered := f.parkRunAsk(ctx, "Answered next to it")
		f.answerInCaptain(ctx, answered)

		listed, err := f.reader.List(ctx, todos.DiscoveryFilters{})

		Expect(err).NotTo(HaveOccurred())
		Expect(listed).To(HaveLen(2))
		counts, err := f.newProvider(ctx).CountByStatus(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(counts).To(HaveKeyWithValue(types.StatusInProgress, 1))
	})
})
