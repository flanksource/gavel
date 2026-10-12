package native_test

import (
	"context"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LaunchCoordinator plan approval", Ordered, func() {
	var f *coordinatorFixture

	BeforeAll(func(ctx SpecContext) {
		f = openCoordinatorFixture(ctx, "gavel_native_plan_approve")
	})

	// unselectedPlan is a Captain plan with one revision that no issue selects
	// yet, written through Captain's own plan APIs.
	unselectedPlan := func(ctx context.Context, variant string) (*captaindb.Plan, *captaindb.PlanRevision) {
		sessionID, runID := agentPlanSource(ctx, f)
		plan, err := f.captain.CreateOrGetPlan(ctx, captaindb.CreatePlanInput{
			SourceSessionID: sessionID, SourcePromptRunID: &runID, Variant: variant, Title: "Approval plan " + variant,
		})
		Expect(err).NotTo(HaveOccurred())
		revision, err := f.captain.AppendPlanRevision(ctx, captaindb.AppendPlanRevisionInput{
			PlanID: plan.ID, PlanMarkdown: "# Approved plan " + variant, CreatedBy: "planner",
		})
		Expect(err).NotTo(HaveOccurred())
		return plan, revision
	}
	selection := func(issue *native.Issue, actor string) native.PlanSelectionAttachment {
		return native.PlanSelectionAttachment{IssueID: issue.ID, ExpectedIssueVersion: issue.Version, Actor: actor}
	}

	It("rolls the Captain approval back when the issue version is stale", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "stale approval")
		plan, revision := unselectedPlan(ctx, "stale")
		stale := selection(issue, "reviewer")
		stale.ExpectedIssueVersion--

		_, err := f.coordinator.ApproveAndSelectPlan(ctx, captaindb.ApprovePlanRevisionInput{
			PlanID: plan.ID, RevisionID: revision.ID, ApprovedBy: "reviewer",
		}, stale)

		Expect(err).To(MatchError(native.ErrVersionConflict))
		stored, err := f.captain.GetPlan(ctx, plan.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(stored.ApprovalState).To(Equal(captaindb.PlanApprovalPending))
		Expect(stored.ApprovedRevisionID).To(BeNil())
		links, err := f.repository.ListPlans(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(links).To(BeEmpty())
	})

	It("approves and selects, replays exactly, and records a changed approval once", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "approval selection")
		plan, revision := unselectedPlan(ctx, "selection")
		approval := captaindb.ApprovePlanRevisionInput{
			PlanID: plan.ID, RevisionID: revision.ID, ApprovedBy: "reviewer", Comment: "approved",
		}

		approved, err := f.coordinator.ApproveAndSelectPlan(ctx, approval, selection(issue, "reviewer"))
		Expect(err).NotTo(HaveOccurred())
		Expect(approved.Plan.ApprovalState).To(Equal(captaindb.PlanApprovalApproved))
		Expect(approved.Plan.ApprovedRevisionID).To(Equal(&revision.ID))
		Expect(approved.Issue.SelectedPlanID).To(Equal(&plan.ID))
		Expect(approved.Issue.Version).To(Equal(issue.Version + 1))
		Expect(f.lastEventKind(ctx, issue.ID)).To(Equal("plan_approved_and_selected"))

		replay, err := f.coordinator.ApproveAndSelectPlan(ctx, approval, selection(issue, "reviewer"))
		Expect(err).NotTo(HaveOccurred())
		Expect(replay.Issue.Version).To(Equal(approved.Issue.Version), "an exact replay with the original version is a no-op")

		changed := approval
		changed.ApprovedBy, changed.Comment = "second-reviewer", "approved after another review"
		_, err = f.coordinator.ApproveAndSelectPlan(ctx, changed, selection(issue, "reviewer"))
		Expect(err).To(MatchError(native.ErrVersionConflict), "a stale caller cannot change the approval")
		afterStale, err := f.captain.GetPlan(ctx, plan.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(afterStale.ApprovedBy).To(Equal(approval.ApprovedBy))

		changedApproved, err := f.coordinator.ApproveAndSelectPlan(ctx, changed, selection(approved.Issue, "reviewer"))
		Expect(err).NotTo(HaveOccurred())
		Expect(changedApproved.Plan.ApprovedBy).To(Equal(changed.ApprovedBy))
		Expect(changedApproved.Plan.ApprovalComment).To(Equal(changed.Comment))
		Expect(changedApproved.Issue.Version).To(Equal(approved.Issue.Version + 1))
		Expect(f.lastEventKind(ctx, issue.ID)).To(Equal("plan_approval_changed"))
	})

	It("waits on the plan lock and observes an identical independent approval as a no-op", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "independent approval")
		plan, revision := unselectedPlan(ctx, "independent")
		approved, err := f.coordinator.ApproveAndSelectPlan(ctx, captaindb.ApprovePlanRevisionInput{
			PlanID: plan.ID, RevisionID: revision.ID, ApprovedBy: "reviewer",
		}, selection(issue, "reviewer"))
		Expect(err).NotTo(HaveOccurred())
		independent := captaindb.ApprovePlanRevisionInput{
			PlanID: plan.ID, RevisionID: revision.ID, ApprovedBy: "captain-independent", Comment: "approved independently",
		}
		ready, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
		defer releaseOnce(release)
		go func() {
			defer GinkgoRecover()
			done <- f.captain.Transaction(ctx, func(tx *captaindb.DB) error {
				if _, approveErr := tx.ApprovePlanRevision(ctx, independent); approveErr != nil {
					return approveErr
				}
				close(ready)
				<-release
				return nil
			})
		}()
		Eventually(ready).Should(BeClosed())
		eventsBefore := f.eventCount(ctx, issue.ID)

		type approveResult struct {
			result *native.ApprovedPlanSelection
			err    error
		}
		coordinated := make(chan approveResult, 1)
		go func() {
			result, approveErr := f.coordinator.ApproveAndSelectPlan(ctx, independent, selection(approved.Issue, "plan-test"))
			coordinated <- approveResult{result, approveErr}
		}()
		waitForBlockedPlanLock(ctx, f.db)
		releaseOnce(release)
		Expect(<-done).To(Succeed())
		outcome := <-coordinated

		Expect(outcome.err).NotTo(HaveOccurred())
		Expect(outcome.result.Plan.ApprovedBy).To(Equal(independent.ApprovedBy))
		Expect(outcome.result.Issue.Version).To(Equal(approved.Issue.Version))
		Expect(f.eventCount(ctx, issue.ID)).To(Equal(eventsBefore))
	})

	It("lets exactly one of two concurrent approvals select its plan on one issue", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "approval race")
		first, firstRevision := unselectedPlan(ctx, "race-a")
		second, secondRevision := unselectedPlan(ctx, "race-b")
		approve := func(plan *captaindb.Plan, revision *captaindb.PlanRevision, reviewer string) func() error {
			return func() error {
				_, err := f.coordinator.ApproveAndSelectPlan(ctx, captaindb.ApprovePlanRevisionInput{
					PlanID: plan.ID, RevisionID: revision.ID, ApprovedBy: reviewer,
				}, selection(issue, reviewer))
				return err
			}
		}

		errs := runConcurrently(approve(first, firstRevision, "reviewer-a"), approve(second, secondRevision, "reviewer-b"))

		Expect(errs).To(ConsistOf(BeNil(), MatchError(native.ErrVersionConflict)))
		stored, err := f.repository.GetIssue(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		approvedPlans := 0
		for _, candidate := range []*captaindb.Plan{first, second} {
			plan, getErr := f.captain.GetPlan(ctx, candidate.ID)
			Expect(getErr).NotTo(HaveOccurred())
			if plan.ApprovalState == captaindb.PlanApprovalApproved {
				approvedPlans++
				Expect(stored.SelectedPlanID).To(Equal(&candidate.ID))
			} else {
				Expect(plan.ApprovalState).To(Equal(captaindb.PlanApprovalPending), "the losing approval rolls back")
			}
		}
		Expect(approvedPlans).To(Equal(1))
		links, err := f.repository.ListPlans(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(links).To(HaveLen(1))
	})

	It("creates an issue with a supplied plan approved at its latest revision", func(ctx SpecContext) {
		issueID := uuid.New()
		root, session := todoSessionTree(issueID, uuid.New(), "plan")
		const markdown = "# Supplied plan\n\n1. Do it."

		created, err := f.coordinator.CreateIssueWithPlan(ctx, native.CreateIssuePlanInput{
			Issue:       native.CreateIssueInput{ID: issueID, WorkspaceID: f.workspace.ID, Title: "supplied plan"},
			RootSession: root, Session: session,
			Plan:     captaindb.CreatePlanInput{Title: "supplied plan", Variant: "primary", SpecProfile: "gavel.todo.plan"},
			Revision: captaindb.AppendPlanRevisionInput{PlanMarkdown: markdown, CreatedBy: "human"},
			Approval: &native.InitialPlanApproval{ApprovedBy: "human", Comment: "reviewed on create"},
			Actor:    "plan-test",
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(created.Plan.SourceSessionID).To(Equal(session.ID))
		Expect(created.Plan.ApprovalState).To(Equal(captaindb.PlanApprovalApproved))
		Expect(created.Plan.ApprovedRevisionID).To(Equal(&created.Revision.ID))
		Expect(created.Plan.ApprovedBy).To(Equal("human"))
		Expect(created.Issue.SelectedPlanID).To(Equal(&created.Plan.ID))
		Expect(f.lastEventKind(ctx, issueID)).To(Equal("plan_approved_and_selected"))
	})

	It("creates an issue with a supplied pending plan selected", func(ctx SpecContext) {
		issueID := uuid.New()
		root, session := todoSessionTree(issueID, uuid.New(), "plan")

		created, err := f.coordinator.CreateIssueWithPlan(ctx, native.CreateIssuePlanInput{
			Issue:       native.CreateIssueInput{ID: issueID, WorkspaceID: f.workspace.ID, Title: "pending supplied plan"},
			RootSession: root, Session: session,
			Plan:     captaindb.CreatePlanInput{Title: "pending supplied plan", Variant: "primary"},
			Revision: captaindb.AppendPlanRevisionInput{PlanMarkdown: "# Pending plan", CreatedBy: "human"},
			Actor:    "plan-test",
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(created.Plan.ApprovalState).To(Equal(captaindb.PlanApprovalPending))
		Expect(created.Issue.SelectedPlanID).To(Equal(&created.Plan.ID))
		Expect(f.lastEventKind(ctx, issueID)).To(Equal("plan_created_and_selected"))
	})
})
