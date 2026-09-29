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

var _ = Describe("LaunchCoordinator session and prompt-run attachment", Ordered, func() {
	var f *coordinatorFixture
	var owner native.RunOwner

	BeforeAll(func(ctx SpecContext) {
		f = openCoordinatorFixture(ctx, "gavel_native_launch_coordinator")
		var err error
		owner, err = native.LocalOwner()
		Expect(err).NotTo(HaveOccurred())
	})

	// admit stands in for Captain's promptrun admission: it ensures the session
	// tree and the run in one Captain transaction and calls the host link,
	// exactly the hook Gavel hands promptrun.Recording.Link.
	admit := func(
		ctx context.Context,
		issueID, sessionID uuid.UUID,
		run captaindb.CreatePromptRunInput,
		attachment native.PromptRunAttachment,
	) (*native.PromptRunLaunch, error) {
		var launch *native.PromptRunLaunch
		root, session := todoSessionTree(issueID, sessionID, string(attachment.StepKind))
		err := f.captain.Transaction(ctx, func(tx *captaindb.DB) error {
			sessions, err := sessiontree.EnsureTx(ctx, tx, []captaindb.CreateSessionInput{root, session})
			if err != nil {
				return err
			}
			run.SessionID = sessions[1].ID
			admitted, err := tx.CreatePromptRun(ctx, run)
			if err != nil {
				return err
			}
			launch, err = f.coordinator.AttachPromptRun(ctx, tx, admitted, attachment)
			return err
		})
		return launch, err
	}

	planRun := func(admissionKey string) captaindb.CreatePromptRunInput {
		// The projection classifies a run by what it was asked to do: a plan run
		// is one whose rendered spec runs in plan mode.
		return captaindb.CreatePromptRunInput{
			AdmissionKey: admissionKey, Origin: "gavel.todos.plan", PromptMarkdown: "Draft a plan",
			RenderedSpec: map[string]any{"permissions": map[string]any{"mode": "plan"}},
		}
	}

	It("activates the admitted run on its issue and owns dispatch only on first admission", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "coordinated launch")
		sessionID := uuid.New()
		attachment := native.PromptRunAttachment{
			IssueID: issue.ID, StepKind: native.StepPlan, Ordinal: 0,
			ExpectedIssueVersion: issue.Version, Actor: "launch-test", Owner: &owner,
		}

		launch, err := admit(ctx, issue.ID, sessionID, planRun("gavel:attach:plan:0"), attachment)
		Expect(err).NotTo(HaveOccurred())
		Expect(launch.DispatchOwned).To(BeTrue())
		Expect(launch.Session.ID).To(Equal(sessionID))
		Expect(launch.PromptRun.SessionID).To(Equal(sessionID))
		Expect(launch.Issue.ActivePromptRunID).To(Equal(&launch.PromptRun.ID))
		Expect(launch.Issue.ExecutionState).To(Equal(native.ExecutionPlanning))
		Expect(launch.Issue.Version).To(Equal(issue.Version+1), "attaching the run is the one mutation")

		replay, err := admit(ctx, issue.ID, sessionID, planRun("gavel:attach:plan:0"), attachment)
		Expect(err).NotTo(HaveOccurred())
		Expect(replay.PromptRun.ID).To(Equal(launch.PromptRun.ID))
		Expect(replay.Issue.Version).To(Equal(launch.Issue.Version))
		Expect(replay.DispatchOwned).To(BeFalse(), "a replayed admission must not start a second agent")
	})

	It("rolls the Captain admission back when the issue version is stale", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "stale coordinated launch")
		sessionID := uuid.New()
		admissionKey := "gavel:attach-stale:run:0"

		_, err := admit(ctx, issue.ID, sessionID, captaindb.CreatePromptRunInput{
			AdmissionKey: admissionKey, Origin: "gavel.todos.run",
		}, native.PromptRunAttachment{
			IssueID: issue.ID, StepKind: native.StepRun, Ordinal: 0,
			ExpectedIssueVersion: issue.Version - 1, Actor: "launch-test", Owner: &owner,
		})

		Expect(err).To(MatchError(native.ErrVersionConflict))
		f.expectNoSession(ctx, issue.ID)
		f.expectNoSession(ctx, sessionID)
		var runs int64
		Expect(f.db.Table("captain_prompt_runs").Where("admission_key = ?", admissionKey).Count(&runs).Error).To(Succeed())
		Expect(runs).To(BeZero())
		stored, err := f.repository.GetIssue(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(stored.ActivePromptRunID).To(BeNil())
	})

	It("rejects an attachment that names a different prompt run than the admitted one", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "mismatched run")
		other := uuid.New()

		_, err := admit(ctx, issue.ID, uuid.New(), planRun("gavel:attach-mismatch:plan:0"), native.PromptRunAttachment{
			IssueID: issue.ID, PromptRunID: other, StepKind: native.StepPlan,
			ExpectedIssueVersion: issue.Version, Actor: "launch-test", Owner: &owner,
		})

		Expect(err).To(MatchError(native.ErrInvalidInput))
		Expect(err.Error()).To(ContainSubstring(other.String()))
	})

	It("refuses a Captain handle that is not scoped to the admission transaction", func(ctx SpecContext) {
		issue := f.createIssue(ctx, "untransacted attach")

		_, err := f.coordinator.AttachPromptRun(ctx, f.captain, &captaindb.PromptRun{ID: uuid.New(), SessionID: uuid.New()},
			native.PromptRunAttachment{IssueID: issue.ID, StepKind: native.StepRun, ExpectedIssueVersion: issue.Version, Owner: &owner})

		Expect(err).To(MatchError(native.ErrInvalidInput))
	})

	It("creates an issue together with its canonical root session", func(ctx SpecContext) {
		issueID := uuid.New()
		root, _ := todoSessionTree(issueID, uuid.New(), "")

		created, err := f.coordinator.CreateIssueWithSession(ctx, native.CreateIssueSessionInput{
			Issue:       native.CreateIssueInput{ID: issueID, WorkspaceID: f.workspace.ID, Title: "rooted issue"},
			RootSession: root,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(created.Issue.ID).To(Equal(issueID))
		Expect(created.Root.ID).To(Equal(issueID))
		Expect(created.Root.ParentSessionID).To(BeNil())
		Expect(created.Root.RootSessionID).To(BeNil())
	})

	It("rolls the root session back when the issue cannot be created", func(ctx SpecContext) {
		issueID := uuid.New()
		root, _ := todoSessionTree(issueID, uuid.New(), "")

		_, err := f.coordinator.CreateIssueWithSession(ctx, native.CreateIssueSessionInput{
			Issue:       native.CreateIssueInput{ID: issueID, WorkspaceID: uuid.New(), Title: "orphan workspace"},
			RootSession: root,
		})

		Expect(err).To(HaveOccurred())
		f.expectNoSession(ctx, issueID)
	})

	It("updates an issue and hangs the operation session under its root", func(ctx SpecContext) {
		issueID := uuid.New()
		root, session := todoSessionTree(issueID, uuid.New(), "verify")
		created, err := f.coordinator.CreateIssueWithSession(ctx, native.CreateIssueSessionInput{
			Issue:       native.CreateIssueInput{ID: issueID, WorkspaceID: f.workspace.ID, Title: "verification target"},
			RootSession: root,
		})
		Expect(err).NotTo(HaveOccurred())
		verification := "go test ./..."

		updated, err := f.coordinator.UpdateIssueWithSession(ctx, native.UpdateIssueSessionInput{
			IssueID: issueID, ExpectedIssueVersion: created.Issue.Version,
			Patch:       native.IssuePatch{Verification: &verification, Actor: "launch-test"},
			RootSession: root, Session: session,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(updated.Version).To(Equal(created.Issue.Version + 1))
		stored, err := f.captain.GetSession(ctx, session.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(stored.ParentSessionID).To(Equal(&issueID))
		Expect(stored.RootSessionID).To(Equal(&issueID))
	})

	It("rolls the operation session back when the issue update loses its version race", func(ctx SpecContext) {
		issueID := uuid.New()
		root, session := todoSessionTree(issueID, uuid.New(), "verify")
		created, err := f.coordinator.CreateIssueWithSession(ctx, native.CreateIssueSessionInput{
			Issue:       native.CreateIssueInput{ID: issueID, WorkspaceID: f.workspace.ID, Title: "stale update"},
			RootSession: root,
		})
		Expect(err).NotTo(HaveOccurred())
		verification := "go test ./..."

		_, err = f.coordinator.UpdateIssueWithSession(ctx, native.UpdateIssueSessionInput{
			IssueID: issueID, ExpectedIssueVersion: created.Issue.Version - 1,
			Patch:       native.IssuePatch{Verification: &verification, Actor: "launch-test"},
			RootSession: root, Session: session,
		})

		Expect(err).To(MatchError(native.ErrVersionConflict))
		f.expectNoSession(ctx, session.ID)
	})
})
