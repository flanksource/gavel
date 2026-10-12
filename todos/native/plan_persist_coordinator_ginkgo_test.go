package native_test

import (
	"context"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/sessiontree"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// agentPlanSource admits a TODO plan-run session tree and prompt run through
// Captain's own APIs, the provenance an agent-produced plan carries.
func agentPlanSource(ctx context.Context, f *coordinatorFixture) (uuid.UUID, uuid.UUID) {
	GinkgoHelper()
	root, session := todoSessionTree(uuid.New(), uuid.New(), "plan")
	var runID uuid.UUID
	Expect(f.captain.Transaction(ctx, func(tx *captaindb.DB) error {
		if _, err := sessiontree.EnsureTx(ctx, tx, []captaindb.CreateSessionInput{root, session}); err != nil {
			return err
		}
		run, err := tx.CreatePromptRun(ctx, captaindb.CreatePromptRunInput{SessionID: session.ID, Origin: "gavel.todos.plan"})
		if err != nil {
			return err
		}
		runID = run.ID
		return nil
	})).To(Succeed())
	return session.ID, runID
}

var _ = Describe("LaunchCoordinator.PersistAndSelectPlan", Ordered, func() {
	var f *coordinatorFixture

	BeforeAll(func(ctx SpecContext) {
		f = openCoordinatorFixture(ctx, "gavel_native_plan_persist")
	})

	agentPlan := func(ctx context.Context, variant string) captaindb.CreatePlanInput {
		sessionID, runID := agentPlanSource(ctx, f)
		return captaindb.CreatePlanInput{
			SourceSessionID: sessionID, SourcePromptRunID: &runID, Variant: variant,
			Title: "Durable coordinated plan", Path: "/workspace/deleted-agent-plan.md",
		}
	}
	revision := func(markdown, author string) captaindb.AppendPlanRevisionInput {
		return captaindb.AppendPlanRevisionInput{PlanMarkdown: markdown, CreatedBy: author}
	}
	selection := func(issue *native.Issue) native.PlanSelectionAttachment {
		return native.PlanSelectionAttachment{IssueID: issue.ID, ExpectedIssueVersion: issue.Version, Actor: "plan-test"}
	}
	persist := func(ctx context.Context, plan captaindb.CreatePlanInput, rev captaindb.AppendPlanRevisionInput, attachment native.PlanSelectionAttachment) (*native.PersistedPlan, error) {
		return f.coordinator.PersistAndSelectPlan(ctx, native.PersistPlanInput{Plan: plan, Revision: rev, Attachment: attachment})
	}
	const firstMarkdown = "# Durable plan\n\n1. Implement the native seam."

	It("saves the revision, selects the plan, and replays exactly with the original issue version", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "persist and replay")
		plan := agentPlan(ctx, "primary")

		persisted, err := persist(ctx, plan, revision(firstMarkdown, "planner"), selection(issue))
		Expect(err).NotTo(HaveOccurred())
		Expect(persisted.Revision.PlanID).To(Equal(persisted.Plan.ID))
		Expect(persisted.Revision.PlanMarkdown).To(Equal(firstMarkdown))
		Expect(persisted.Plan.LatestRevision.ID).To(Equal(persisted.Revision.ID))
		Expect(persisted.Issue.SelectedPlanID).To(Equal(&persisted.Plan.ID))
		Expect(persisted.Issue.Version).To(Equal(issue.Version + 1))
		Expect(f.lastEventKind(ctx, issue.ID)).To(Equal("plan_persisted_and_selected"))

		replay, err := persist(ctx, plan, revision(firstMarkdown, "planner"), selection(issue))
		Expect(err).NotTo(HaveOccurred())
		Expect(replay.Revision.ID).To(Equal(persisted.Revision.ID))
		Expect(replay.Issue.Version).To(Equal(persisted.Issue.Version))
	})

	It("replays equivalent content by content hash alone, whoever authored the retry", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "content-hash replay")
		plan := agentPlan(ctx, "primary")
		persisted, err := persist(ctx, plan, revision(firstMarkdown, "planner"), selection(issue))
		Expect(err).NotTo(HaveOccurred())

		replay, err := persist(ctx, plan, revision("\r\n"+firstMarkdown+"\r\n", "someone-else"), selection(issue))

		Expect(err).NotTo(HaveOccurred())
		Expect(replay.Revision.ID).To(Equal(persisted.Revision.ID))
		Expect(replay.Issue.Version).To(Equal(persisted.Issue.Version))
	})

	It("still validates the plan identity on a stale exact replay", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "identity validation")
		plan := agentPlan(ctx, "primary")
		_, err := persist(ctx, plan, revision(firstMarkdown, "planner"), selection(issue))
		Expect(err).NotTo(HaveOccurred())

		wrongSession := plan
		wrongSession.SourceSessionID, _ = agentPlanSource(ctx, f)
		_, err = persist(ctx, wrongSession, revision(firstMarkdown, "planner"), selection(issue))
		Expect(err).To(MatchError(captaindb.ErrPlanConflict))

		wrongID := plan
		wrongID.ID = uuid.New()
		_, err = persist(ctx, wrongID, revision(firstMarkdown, "planner"), selection(issue))
		Expect(err).To(MatchError(captaindb.ErrPlanConflict))
		stored, err := f.captain.ListPlans(ctx, captaindb.PlanFilter{SourcePromptRunID: plan.SourcePromptRunID})
		Expect(err).NotTo(HaveOccurred())
		Expect(stored).To(HaveLen(1), "a rejected identity creates no second plan")
	})

	It("rolls a stale caller's new revision back", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "stale revision")
		plan := agentPlan(ctx, "primary")
		persisted, err := persist(ctx, plan, revision(firstMarkdown, "planner"), selection(issue))
		Expect(err).NotTo(HaveOccurred())

		_, err = persist(ctx, plan, revision("# Durable plan v2", "planner"), selection(issue))

		Expect(err).To(MatchError(native.ErrVersionConflict))
		revisions, err := f.captain.ListPlanRevisions(ctx, persisted.Plan.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(revisions).To(HaveLen(1))
		stored, err := f.repository.GetIssue(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(stored.Version).To(Equal(persisted.Issue.Version))
	})

	It("returns an approved plan to pending with its approval cleared when new content is saved", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "revise approved plan")
		plan := agentPlan(ctx, "primary")
		persisted, err := persist(ctx, plan, revision(firstMarkdown, "planner"), selection(issue))
		Expect(err).NotTo(HaveOccurred())
		approved, err := f.coordinator.ApproveAndSelectPlan(ctx, captaindb.ApprovePlanRevisionInput{
			PlanID: persisted.Plan.ID, RevisionID: persisted.Revision.ID, ApprovedBy: "reviewer", Comment: "ship it",
		}, selection(persisted.Issue))
		Expect(err).NotTo(HaveOccurred())

		revised, err := persist(ctx, plan, revision("# Durable plan v2\n\n2. Record changes.", "planner"), selection(approved.Issue))

		Expect(err).NotTo(HaveOccurred())
		Expect(revised.Revision.Revision).To(Equal(2))
		Expect(revised.Plan.ApprovalState).To(Equal(captaindb.PlanApprovalPending))
		Expect(revised.Plan.ApprovedRevisionID).To(BeNil())
		Expect(revised.Plan.ApprovedBy).To(BeEmpty())
		Expect(revised.Plan.ApprovalComment).To(BeEmpty())
		Expect(revised.Issue.Version).To(Equal(approved.Issue.Version + 1))
		Expect(f.lastEventKind(ctx, issue.ID)).To(Equal("plan_revision_persisted"))
	})

	It("takes the plan lock before the issue lock and observes an independent revision as a no-op", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "independent writer")
		plan := agentPlan(ctx, "primary")
		persisted, err := persist(ctx, plan, revision(firstMarkdown, "planner"), selection(issue))
		Expect(err).NotTo(HaveOccurred())
		independent := captaindb.AppendPlanRevisionInput{
			PlanID: persisted.Plan.ID, PlanMarkdown: "# Written independently by Captain", CreatedBy: "captain-independent",
		}
		ready, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
		defer releaseOnce(release)
		var independentRevision *captaindb.PlanRevision
		go func() {
			defer GinkgoRecover()
			done <- f.captain.Transaction(ctx, func(tx *captaindb.DB) error {
				var appendErr error
				if independentRevision, appendErr = tx.AppendPlanRevision(ctx, independent); appendErr != nil {
					return appendErr
				}
				close(ready)
				<-release
				return nil
			})
		}()
		Eventually(ready).Should(BeClosed())
		eventsBefore := f.eventCount(ctx, issue.ID)

		type persistResult struct {
			result *native.PersistedPlan
			err    error
		}
		coordinated := make(chan persistResult, 1)
		go func() {
			result, persistErr := persist(ctx, plan, independent, selection(persisted.Issue))
			coordinated <- persistResult{result, persistErr}
		}()
		waitForBlockedPlanLock(ctx, f.db)
		releaseOnce(release)
		Expect(<-done).To(Succeed())
		outcome := <-coordinated

		Expect(outcome.err).NotTo(HaveOccurred())
		Expect(outcome.result.Revision.ID).To(Equal(independentRevision.ID))
		Expect(outcome.result.Issue.Version).To(Equal(persisted.Issue.Version))
		Expect(f.eventCount(ctx, issue.ID)).To(Equal(eventsBefore))
	})

	It("bootstraps the TODO root and plan session for a human plan in the same transaction", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "human plan")
		root, session := todoSessionTree(issue.ID, uuid.New(), "plan")

		persisted, err := f.coordinator.PersistAndSelectPlan(ctx, native.PersistPlanInput{
			Plan:        captaindb.CreatePlanInput{Title: issue.Title, Variant: "primary", SpecProfile: "gavel.todo.plan"},
			Revision:    revision(firstMarkdown, "human"),
			Attachment:  selection(issue),
			RootSession: &root, Session: &session,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(persisted.Plan.SourceSessionID).To(Equal(session.ID))
		stored, err := f.captain.GetSession(ctx, session.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(stored.RootSessionID).To(Equal(&issue.ID))
	})

	It("rolls the bootstrapped sessions back with a stale human plan", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "stale human plan")
		root, session := todoSessionTree(issue.ID, uuid.New(), "plan")
		stale := selection(issue)
		stale.ExpectedIssueVersion--

		_, err := f.coordinator.PersistAndSelectPlan(ctx, native.PersistPlanInput{
			Plan:        captaindb.CreatePlanInput{Title: issue.Title, Variant: "primary"},
			Revision:    revision(firstMarkdown, "human"),
			Attachment:  stale,
			RootSession: &root, Session: &session,
		})

		Expect(err).To(MatchError(native.ErrVersionConflict))
		f.expectNoSession(ctx, issue.ID)
		f.expectNoSession(ctx, session.ID)
	})

	It("requires session inputs for a plan without a source session", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "sessionless plan")

		_, err := persist(ctx, captaindb.CreatePlanInput{Title: issue.Title}, revision(firstMarkdown, "human"), selection(issue))

		Expect(err).To(MatchError(native.ErrInvalidInput))
	})
})
