package native_test

import (
	"context"
	"encoding/json"

	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("issue parents", Ordered, func() {
	var (
		fixture *coordinatorFixture
		other   *native.Workspace
	)

	BeforeAll(func(ctx SpecContext) {
		fixture = openCoordinatorFixture(ctx, "gavel_native_parents")
		var err error
		other, err = fixture.repository.CreateWorkspace(ctx, native.CreateWorkspaceInput{
			RepoKey: "github.com/acme/parents-elsewhere", RootPath: GinkgoT().TempDir(),
		})
		Expect(err).NotTo(HaveOccurred())
	})

	createChild := func(ctx context.Context, title string, parent *native.Issue) *native.Issue {
		GinkgoHelper()
		child, err := fixture.repository.CreateIssue(ctx, native.CreateIssueInput{
			WorkspaceID: fixture.workspace.ID, Title: title, ParentID: &parent.ID,
		})
		Expect(err).NotTo(HaveOccurred())
		return child
	}

	setStatus := func(ctx context.Context, issue *native.Issue, status native.IssueStatus) *native.Issue {
		GinkgoHelper()
		updated, err := fixture.repository.UpdateIssue(ctx, issue.ID, issue.Version, native.IssuePatch{Status: &status, Actor: "tester"})
		Expect(err).NotTo(HaveOccurred())
		return updated
	}

	lastEvent := func(ctx context.Context, issueID uuid.UUID) (native.Event, map[string]any) {
		GinkgoHelper()
		events, err := fixture.repository.ListEvents(ctx, issueID)
		Expect(err).NotTo(HaveOccurred())
		Expect(events).NotTo(BeEmpty())
		last := events[len(events)-1]
		var payload map[string]any
		Expect(json.Unmarshal(last.Payload, &payload)).To(Succeed())
		return last, payload
	}

	Describe("CreateIssue", func() {
		It("stores the parent and records it with the origin in the created event", func(ctx SpecContext) {
			parent := fixture.createIssue(ctx, "Split the importer")
			origin := native.IssueOrigin{IssueID: " " + parent.ID.String() + " ", SessionID: " session-7f3a "}

			child, err := fixture.repository.CreateIssue(ctx, native.CreateIssueInput{
				WorkspaceID: fixture.workspace.ID, Title: "Import the aliases",
				ParentID: &parent.ID, Origin: origin, Actor: "tester",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(child.ParentID).To(HaveValue(Equal(parent.ID)))

			event, payload := lastEvent(ctx, child.ID)
			Expect(event.Kind).To(Equal("created"))
			Expect(payload).To(SatisfyAll(
				HaveKeyWithValue("parent", parent.ID.String()),
				HaveKeyWithValue("originIssue", parent.ID.String()),
				HaveKeyWithValue("originSession", "session-7f3a"),
			))
		})

		It("records neither a parent nor an origin for a plain top-level issue", func(ctx SpecContext) {
			issue := fixture.createIssue(ctx, "Standalone")
			Expect(issue.ParentID).To(BeNil())

			_, payload := lastEvent(ctx, issue.ID)
			Expect(payload).NotTo(SatisfyAny(HaveKey("parent"), HaveKey("originIssue"), HaveKey("originSession")))
		})

		It("records an origin that is not an issue id exactly as it was given", func(ctx SpecContext) {
			issue, err := fixture.repository.CreateIssue(ctx, native.CreateIssueInput{
				WorkspaceID: fixture.workspace.ID, Title: "Created under an alias", Origin: native.IssueOrigin{IssueID: "jira-42"},
			})
			Expect(err).NotTo(HaveOccurred())

			_, payload := lastEvent(ctx, issue.ID)
			Expect(payload).To(SatisfyAll(
				HaveKeyWithValue("originIssue", "jira-42"), Not(HaveKey("parent")), Not(HaveKey("originSession")),
			))
		})

		It("refuses a parent that is itself a child and creates nothing", func(ctx SpecContext) {
			parent := fixture.createIssue(ctx, "Grandparent")
			child := createChild(ctx, "Middle", parent)
			before, err := fixture.repository.ListIssues(ctx, fixture.workspace.ID)
			Expect(err).NotTo(HaveOccurred())

			_, err = fixture.repository.CreateIssue(ctx, native.CreateIssueInput{
				WorkspaceID: fixture.workspace.ID, Title: "Grandchild", ParentID: &child.ID,
			})
			Expect(err).To(MatchError(native.ErrInvalidParent))
			Expect(err).To(MatchError(ContainSubstring(parent.ID.String())), "the error names the parent the todo belongs under")

			after, err := fixture.repository.ListIssues(ctx, fixture.workspace.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(after).To(HaveLen(len(before)))
		})

		It("refuses a parent from another workspace", func(ctx SpecContext) {
			elsewhere, err := fixture.repository.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: other.ID, Title: "Elsewhere"})
			Expect(err).NotTo(HaveOccurred())

			_, err = fixture.repository.CreateIssue(ctx, native.CreateIssueInput{
				WorkspaceID: fixture.workspace.ID, Title: "Cross-workspace child", ParentID: &elsewhere.ID,
			})
			Expect(err).To(MatchError(native.ErrInvalidParent))
		})

		DescribeTable("refuses a parent that is no longer open and creates nothing",
			func(ctx SpecContext, status native.IssueStatus) {
				parent := setStatus(ctx, fixture.createIssue(ctx, "Finished before it had children"), status)

				_, err := fixture.repository.CreateIssue(ctx, native.CreateIssueInput{
					WorkspaceID: fixture.workspace.ID, Title: "Born under a closed parent", ParentID: &parent.ID,
				})

				Expect(err).To(MatchError(native.ErrInvalidParent))
				Expect(err).To(MatchError(ContainSubstring("is " + string(status))))
			},
			Entry("closed", native.StatusClosed),
			Entry("cancelled", native.StatusCancelled),
		)

		It("reports a parent that does not exist as not found", func(ctx SpecContext) {
			missing := uuid.New()
			_, err := fixture.repository.CreateIssue(ctx, native.CreateIssueInput{
				WorkspaceID: fixture.workspace.ID, Title: "Orphan", ParentID: &missing,
			})
			Expect(err).To(MatchError(native.ErrNotFound))
		})
	})

	Describe("SetIssueParent", func() {
		It("attaches, re-parents and detaches, recording parent_changed each time", func(ctx SpecContext) {
			first := fixture.createIssue(ctx, "First parent")
			second := fixture.createIssue(ctx, "Second parent")
			issue := fixture.createIssue(ctx, "Moves around")

			attached, err := fixture.repository.SetIssueParent(ctx, issue.ID, &first.ID, issue.Version, "tester")
			Expect(err).NotTo(HaveOccurred())
			Expect(attached.ParentID).To(HaveValue(Equal(first.ID)))
			Expect(attached.Version).To(Equal(issue.Version + 1))
			event, payload := lastEvent(ctx, issue.ID)
			Expect(event.Kind).To(Equal("parent_changed"))
			Expect(event.Actor).To(Equal("tester"))
			Expect(payload).To(Equal(map[string]any{"from": nil, "to": first.ID.String()}))

			moved, err := fixture.repository.SetIssueParent(ctx, issue.ID, &second.ID, attached.Version, "tester")
			Expect(err).NotTo(HaveOccurred())
			Expect(moved.ParentID).To(HaveValue(Equal(second.ID)))
			_, payload = lastEvent(ctx, issue.ID)
			Expect(payload).To(Equal(map[string]any{"from": first.ID.String(), "to": second.ID.String()}))

			detached, err := fixture.repository.SetIssueParent(ctx, issue.ID, nil, moved.Version, "tester")
			Expect(err).NotTo(HaveOccurred())
			Expect(detached.ParentID).To(BeNil())
			_, payload = lastEvent(ctx, issue.ID)
			Expect(payload).To(Equal(map[string]any{"from": second.ID.String(), "to": nil}))
		})

		It("changes nothing when the issue already has that parent", func(ctx SpecContext) {
			parent := fixture.createIssue(ctx, "Steady parent")
			child := createChild(ctx, "Steady child", parent)
			events := fixture.eventCount(ctx, child.ID)

			same, err := fixture.repository.SetIssueParent(ctx, child.ID, &parent.ID, child.Version, "tester")
			Expect(err).NotTo(HaveOccurred())
			Expect(same.Version).To(Equal(child.Version))
			Expect(fixture.eventCount(ctx, child.ID)).To(Equal(events))
		})

		It("refuses a stale version", func(ctx SpecContext) {
			parent := fixture.createIssue(ctx, "Version parent")
			issue := fixture.createIssue(ctx, "Version child")

			_, err := fixture.repository.SetIssueParent(ctx, issue.ID, &parent.ID, issue.Version+1, "tester")
			Expect(err).To(MatchError(native.ErrVersionConflict))
		})

		DescribeTable("refuses a parent that would break the single level",
			func(ctx SpecContext, arrange func(context.Context) (issue *native.Issue, parentID uuid.UUID)) {
				issue, parentID := arrange(ctx)
				events := fixture.eventCount(ctx, issue.ID)

				_, err := fixture.repository.SetIssueParent(ctx, issue.ID, &parentID, issue.Version, "tester")
				Expect(err).To(MatchError(native.ErrInvalidParent))

				unchanged, err := fixture.repository.GetIssue(ctx, issue.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(unchanged.ParentID).To(Equal(issue.ParentID))
				Expect(unchanged.Version).To(Equal(issue.Version))
				Expect(fixture.eventCount(ctx, issue.ID)).To(Equal(events))
			},
			Entry("itself", func(ctx context.Context) (*native.Issue, uuid.UUID) {
				issue := fixture.createIssue(ctx, "Self parent")
				return issue, issue.ID
			}),
			Entry("a parent that is itself a child", func(ctx context.Context) (*native.Issue, uuid.UUID) {
				child := createChild(ctx, "Already a child", fixture.createIssue(ctx, "Top"))
				return fixture.createIssue(ctx, "Wants a grandparent"), child.ID
			}),
			Entry("an issue that already has children", func(ctx context.Context) (*native.Issue, uuid.UUID) {
				parent := fixture.createIssue(ctx, "Has children")
				createChild(ctx, "Existing child", parent)
				return parent, fixture.createIssue(ctx, "Would-be grandparent").ID
			}),
			Entry("a parent in another workspace", func(ctx context.Context) (*native.Issue, uuid.UUID) {
				elsewhere, err := fixture.repository.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: other.ID, Title: "Foreign parent"})
				Expect(err).NotTo(HaveOccurred())
				return fixture.createIssue(ctx, "Local issue"), elsewhere.ID
			}),
			Entry("a closed parent", func(ctx context.Context) (*native.Issue, uuid.UUID) {
				closed := setStatus(ctx, fixture.createIssue(ctx, "Closed parent"), native.StatusClosed)
				return fixture.createIssue(ctx, "Wants a closed parent"), closed.ID
			}),
			Entry("a cancelled parent", func(ctx context.Context) (*native.Issue, uuid.UUID) {
				cancelled := setStatus(ctx, fixture.createIssue(ctx, "Cancelled parent"), native.StatusCancelled)
				return fixture.createIssue(ctx, "Wants a cancelled parent"), cancelled.ID
			}),
		)

		It("lets only one of two issues become the other's child", func(ctx SpecContext) {
			first := fixture.createIssue(ctx, "First of a pair")
			second := fixture.createIssue(ctx, "Second of a pair")

			errs := runConcurrently(
				func() error {
					_, err := fixture.repository.SetIssueParent(ctx, first.ID, &second.ID, first.Version, "tester")
					return err
				},
				func() error {
					_, err := fixture.repository.SetIssueParent(ctx, second.ID, &first.ID, second.Version, "tester")
					return err
				},
			)
			Expect(errs).To(ConsistOf(BeNil(), MatchError(native.ErrInvalidParent)))
		})
	})

	Describe("cancelling an issue", func() {
		cancel := func(ctx context.Context, issue *native.Issue) (*native.Issue, error) {
			cancelled := native.StatusCancelled
			return fixture.repository.UpdateIssue(ctx, issue.ID, issue.Version, native.IssuePatch{Status: &cancelled, Actor: "tester"})
		}

		It("is refused while it has open children, counting only the open ones, and changes nothing", func(ctx SpecContext) {
			parent := fixture.createIssue(ctx, "Parent with work left")
			createChild(ctx, "Still open", parent)
			setStatus(ctx, createChild(ctx, "Verified, not yet closed", parent), native.StatusVerified)
			setStatus(ctx, createChild(ctx, "Already closed", parent), native.StatusClosed)
			events := fixture.eventCount(ctx, parent.ID)

			_, err := cancel(ctx, parent)

			Expect(err).To(MatchError(native.ErrOpenChildren))
			Expect(err).To(MatchError(ContainSubstring("2 under issue " + parent.ID.String())))
			unchanged, err := fixture.repository.GetIssue(ctx, parent.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(unchanged.Status).To(Equal(native.StatusOpen))
			Expect(unchanged.Version).To(Equal(parent.Version))
			Expect(fixture.eventCount(ctx, parent.ID)).To(Equal(events))
		})

		It("goes through once every child is closed or cancelled, and they stay attached", func(ctx SpecContext) {
			parent := fixture.createIssue(ctx, "Parent with nothing left")
			closed := setStatus(ctx, createChild(ctx, "Closed child", parent), native.StatusClosed)
			setStatus(ctx, createChild(ctx, "Cancelled child", parent), native.StatusCancelled)

			cancelled, err := cancel(ctx, parent)

			Expect(err).NotTo(HaveOccurred())
			Expect(cancelled.Status).To(Equal(native.StatusCancelled))
			kept, err := fixture.repository.GetIssue(ctx, closed.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(kept.ParentID).To(HaveValue(Equal(parent.ID)))
		})

		It("cannot be raced by a child attaching to it", func(ctx SpecContext) {
			parent := fixture.createIssue(ctx, "Parent being cancelled")
			joining := fixture.createIssue(ctx, "Joining late")

			errs := runConcurrently(
				func() error {
					_, err := cancel(ctx, parent)
					return err
				},
				func() error {
					_, err := fixture.repository.SetIssueParent(ctx, joining.ID, &parent.ID, joining.Version, "tester")
					return err
				},
			)

			Expect(errs).To(ConsistOf(BeNil(), Or(MatchError(native.ErrOpenChildren), MatchError(native.ErrInvalidParent))),
				"whichever lands first, the other is refused: no open child ends up under a cancelled parent")
		})
	})

	It("CountIssuesByStatus counts top-level issues only, while ListIssues keeps the children", func(ctx SpecContext) {
		workspace, err := fixture.repository.CreateWorkspace(ctx, native.CreateWorkspaceInput{
			RepoKey: "github.com/acme/parents-counts", RootPath: GinkgoT().TempDir(),
		})
		Expect(err).NotTo(HaveOccurred())
		parent, err := fixture.repository.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: "Counted"})
		Expect(err).NotTo(HaveOccurred())
		for _, title := range []string{"Hidden one", "Hidden two"} {
			_, err := fixture.repository.CreateIssue(ctx, native.CreateIssueInput{
				WorkspaceID: workspace.ID, Title: title, ParentID: &parent.ID,
			})
			Expect(err).NotTo(HaveOccurred())
		}

		counts, err := fixture.repository.CountIssuesByStatus(ctx, []uuid.UUID{workspace.ID})
		Expect(err).NotTo(HaveOccurred())
		Expect(counts).To(Equal([]native.IssueStatusCount{{
			WorkspaceID: workspace.ID, Status: native.StatusOpen, ExecutionState: native.ExecutionIdle, Count: 1,
		}}))

		listed, err := fixture.repository.ListIssues(ctx, workspace.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(listed).To(HaveLen(3))
	})

	DescribeTable("MoveIssueWorkspace refuses an issue in a hierarchy",
		func(ctx SpecContext, pick func(parent, child *native.Issue) *native.Issue) {
			parent := fixture.createIssue(ctx, "Anchored parent")
			child := createChild(ctx, "Anchored child", parent)
			moving := pick(parent, child)

			_, err := fixture.repository.MoveIssueWorkspace(ctx, moving.ID, other.ID, moving.Version, "tester")
			Expect(err).To(MatchError(native.ErrIssueInHierarchy))

			unmoved, err := fixture.repository.GetIssue(ctx, moving.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(unmoved.WorkspaceID).To(Equal(fixture.workspace.ID))
			Expect(unmoved.Version).To(Equal(moving.Version))
		},
		Entry("a child", func(_, child *native.Issue) *native.Issue { return child }),
		Entry("a parent with children", func(parent, _ *native.Issue) *native.Issue { return parent }),
	)

	It("DeleteIssue refuses a parent that still has children and deletes a child", func(ctx SpecContext) {
		parent := fixture.createIssue(ctx, "Parent under deletion")
		child := createChild(ctx, "Child under deletion", parent)

		Expect(fixture.repository.DeleteIssue(ctx, parent.ID, parent.Version)).To(MatchError(native.ErrIssueInHierarchy))
		_, err := fixture.repository.GetIssue(ctx, parent.ID)
		Expect(err).NotTo(HaveOccurred())

		Expect(fixture.repository.DeleteIssue(ctx, child.ID, child.Version)).To(Succeed())
		Expect(fixture.repository.DeleteIssue(ctx, parent.ID, parent.Version)).To(Succeed())
	})
})
