package runtime

import (
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Recording an answer: a gavel run that asked is answered from Captain's
// session page, which resumes the same provider session without gavel.
var _ = Describe("recording answers given outside gavel", func() {
	var f *answerFixture

	BeforeEach(func(ctx SpecContext) { f = newAnswerFixture(ctx) })

	It("moves a todo answered in Captain to in progress and records the answer once", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Answered while the turn is live")
		f.answerInCaptain(ctx, parked)

		Expect(f.concurrentGets(ctx, parked.todo.ID)).To(HaveEach(types.StatusInProgress))

		answered := f.eventsOf(ctx, parked.todo, native.EventAskAnswered)
		Expect(answered).To(HaveLen(1))
		Expect(answered[0].Source).To(Equal("captain"))
		Expect(answered[0].Actor).To(Equal("captain"))
		Expect(answered[0].SourceID).To(HavePrefix("ask-answer:" + parked.runID.String() + ":"))
		Expect(answered[0].Body).To(Equal("**Answer (Captain):** " + captainAnswer))
		Expect(payloadOf(answered[0])).To(HaveKeyWithValue("promptRunId", parked.runID.String()))
		Expect(f.gavelRun(ctx, parked).State).To(Equal(captaindb.PromptRunStateWaiting), "a live answer only records; the run settles when Captain's turn ends")
	})

	It("records a dashboard answer when the resumed turn is admitted, even with the parked TUI still alive", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Answered through the dashboard")
		f.liveProcess(ctx, parked.executionSessionID)
		parkedVersion := f.gavelRun(ctx, parked).Version

		f.resumeFromDashboard(ctx, parked, captainAnswer)
		// The resumed turn delivers the answer to the same transcript.
		f.answerInCaptain(ctx, parked)

		answered := f.eventsOf(ctx, parked.todo, native.EventAskAnswered)
		Expect(answered).To(HaveLen(1))
		Expect(answered[0].Source).To(Equal("gavel"))
		Expect(answered[0].Body).To(Equal("**Answer:** " + captainAnswer))
		Expect(payloadOf(answered[0])).To(And(
			HaveKeyWithValue("promptRunId", parked.runID.String()),
			HaveKeyWithValue("runVersion", BeNumerically("==", parkedVersion)),
		))

		read, err := f.dispatcher.Get(ctx, parked.todo.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(read.Status).To(Equal(types.StatusAsk), "the dispatching provider drives this run until its turn starts")
		read, err = f.reader.Get(ctx, parked.todo.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(read.Status).To(Equal(types.StatusAsk), "gavel's own resumed message is not an answer from Captain")
		Expect(f.eventsOf(ctx, parked.todo, native.EventAskAnswered)).To(HaveLen(1))
	})

	It("still records a Captain answer after a dashboard resume that never started", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Dashboard dispatcher died after admission")
		f.resumeFromDashboard(ctx, parked, "use the production database")
		f.dispatcher.clearPrepared(uuid.MustParse(parked.todo.ID), parked.runID)
		f.answerInCaptain(ctx, parked)

		read, err := f.reader.Get(ctx, parked.todo.ID)

		Expect(err).NotTo(HaveOccurred())
		Expect(read.Status).To(Equal(types.StatusInProgress))
		Expect(f.eventsOf(ctx, parked.todo, native.EventAskAnswered)).To(HaveLen(2))
	})

	It("resumes an ask in place after the dispatcher that parked it has exited", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Parked by a dispatcher that exited")
		// The dispatching process is gone: its run claim is released.
		f.dispatcher.clearPrepared(uuid.MustParse(parked.todo.ID), parked.runID)
		dashboard := f.newProvider(ctx)

		f.resumeWith(ctx, dashboard, parked, captainAnswer)

		run := f.gavelRun(ctx, parked)
		Expect(run.State).To(Equal(captaindb.PromptRunStateWaiting), "the parked run is resumed, not reclaimed")
		answered := f.eventsOf(ctx, parked.todo, native.EventAskAnswered)
		Expect(answered).To(HaveLen(1))
		Expect(answered[0].Source).To(Equal("gavel"))
		var cancelled int64
		Expect(f.db.WithContext(ctx).Raw(`
			SELECT count(*) FROM captain_prompt_runs run
			JOIN todo_issue_prompt_runs link ON link.prompt_run_id = run.id
			WHERE link.issue_id = ? AND run.state = 'cancelled'`, parked.todo.ID).Scan(&cancelled).Error).To(Succeed())
		Expect(cancelled).To(BeZero())
		owner, err := f.reader.Repository().PromptRunOwner(ctx, parked.runID)
		Expect(err).NotTo(HaveOccurred())
		alive, reason := owner.Alive(time.Now())
		Expect(alive).To(BeTrue(), "the resuming process claims the run: %s", reason)
	})

	It("still reclaims an orphaned parked run for a dispatch that resumes nothing", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Orphaned and dispatched afresh")
		f.dispatcher.clearPrepared(uuid.MustParse(parked.todo.ID), parked.runID)
		fresh := f.newProvider(ctx)
		todo, err := fresh.Get(ctx, parked.todo.ID)
		Expect(err).NotTo(HaveOccurred())

		admission, err := fresh.PrepareRun(ctx, todo, todos.RunPreparation{Mode: types.ModeRun, Prompt: "run", ExecutorName: "claude"})

		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { fresh.clearPrepared(uuid.MustParse(todo.ID), admission.PromptRunID) })
		Expect(admission.PromptRunID).NotTo(Equal(parked.runID))
		Expect(f.gavelRun(ctx, parked).State).To(Equal(captaindb.PromptRunStateCancelled))
	})

	It("does not read an injected system line as an answer", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Interrupted after the ask")
		f.insertMessage(ctx, parked.executionSessionID, "system", textParts("[Request interrupted by user]"))

		read, err := f.reader.Get(ctx, parked.todo.ID)

		Expect(err).NotTo(HaveOccurred())
		Expect(read.Status).To(Equal(types.StatusAsk))
		Expect(f.eventsOf(ctx, parked.todo, native.EventAskAnswered)).To(BeEmpty())
	})

	It("does not read a tool result as an answer", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Only a tool result followed")
		f.insertMessage(ctx, parked.executionSessionID, "user", []map[string]any{{
			"type": "dynamic-tool", "toolName": "StructuredOutput", "toolCallId": "toolu_result", "state": "output-available",
		}})

		read, err := f.reader.Get(ctx, parked.todo.ID)

		Expect(err).NotTo(HaveOccurred())
		Expect(read.Status).To(Equal(types.StatusAsk))
		Expect(f.eventsOf(ctx, parked.todo, native.EventAskAnswered)).To(BeEmpty())
	})

	It("writes nothing for a todo that is not asking", func(ctx SpecContext) {
		todo, err := f.dispatcher.Create(ctx, todos.CreateRequest{Title: "Never asked", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())

		listed, err := f.reader.List(ctx, todos.DiscoveryFilters{})
		Expect(err).NotTo(HaveOccurred())
		counts, err := f.reader.CountByStatus(ctx)
		Expect(err).NotTo(HaveOccurred())
		got, err := f.reader.Get(ctx, todo.ID)

		Expect(err).NotTo(HaveOccurred())
		Expect(listed).To(HaveLen(1))
		Expect(counts).To(Equal(map[types.Status]int{types.StatusPending: 1}))
		Expect(got.Version).To(Equal(todo.Version))
		Expect(got.ProviderEvents).To(HaveLen(len(todo.ProviderEvents)))
	})

	It("reconciles on the list and count reads too", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Answered before a list")
		f.answerInCaptain(ctx, parked)

		counts, err := f.reader.CountByStatus(ctx)
		Expect(err).NotTo(HaveOccurred())
		listed, err := f.reader.List(ctx, todos.DiscoveryFilters{})
		Expect(err).NotTo(HaveOccurred())

		Expect(counts).To(Equal(map[types.Status]int{types.StatusInProgress: 1}))
		Expect(listed).To(HaveLen(1))
		Expect(listed[0].Status).To(Equal(types.StatusInProgress))
		Expect(f.eventsOf(ctx, parked.todo, native.EventAskAnswered)).To(HaveLen(1))
	})

	It("resolves a global reference through the workspace-less provider without reconciling, and the owning workspace read settles once", func(ctx SpecContext) {
		parked := f.parkRunAsk(ctx, "Resolved through a deep link")
		f.answerInCaptain(ctx, parked)
		f.insertMessage(ctx, parked.executionSessionID, "assistant", textParts(followupSummary))
		f.captainTurn(ctx, parked, captaindb.PromptRunStateSucceeded, "")
		global, err := NewGlobal(f.db)
		Expect(err).NotTo(HaveOccurred())

		resolved, err := global.GetGlobal(ctx, parked.todo.ID)
		Expect(err).NotTo(HaveOccurred())
		_, _, err = global.GetGlobalBySession(ctx, parked.executionSessionID.String())
		Expect(err).To(Or(Not(HaveOccurred()), MatchError(native.ErrNotFound)))
		Expect(f.gavelRun(ctx, parked).State).To(Equal(captaindb.PromptRunStateWaiting), "a global lookup writes nothing")

		owned := f.newProvider(ctx)
		read, err := owned.Get(ctx, resolved.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(read.Status).To(Equal(types.StatusPending))
		Expect(f.eventsOf(ctx, parked.todo, "lifecycle_outcome")).To(HaveLen(2))
	})
})
