package runtime

import (
	"context"
	"encoding/json"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("TODO parents", Ordered, func() {
	const session = "session-4c1d"
	var (
		db        *gorm.DB
		provider  *Provider
		elsewhere *Provider
	)

	BeforeAll(func(ctx SpecContext) {
		dsn := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_todo_parents"}).DSN()

		GinkgoT().Setenv(database.EnvDSN, dsn)
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDSN, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		db = opened.Gorm()

		provider, err = New(ctx, db, WorkspaceOptions{
			Name: "parents", RootPath: GinkgoT().TempDir(), Repositories: []string{"acme/parents"},
		})
		Expect(err).NotTo(HaveOccurred())
		elsewhere, err = New(ctx, db, WorkspaceOptions{
			Name: "parents-elsewhere", RootPath: GinkgoT().TempDir(), Repositories: []string{"acme/parents-elsewhere"},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	create := func(ctx context.Context, in *Provider, request todos.CreateRequest) *types.TODO {
		GinkgoHelper()
		created, err := in.Create(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		return created
	}

	createdPayload := func(todo *types.TODO) map[string]any {
		GinkgoHelper()
		Expect(todo.ProviderEvents).NotTo(BeEmpty())
		Expect(todo.ProviderEvents[0].Kind).To(Equal("created"))
		var payload map[string]any
		Expect(json.Unmarshal(todo.ProviderEvents[0].Payload, &payload)).To(Succeed())
		return payload
	}

	Describe("Create", func() {
		It("attaches the new TODO to an explicit parent named by short id", func(ctx SpecContext) {
			parent := create(ctx, provider, todos.CreateRequest{Title: "Rework the importer"})

			child := create(ctx, provider, todos.CreateRequest{Title: "Import the aliases", Parent: parent.ShortID})

			Expect(child.ParentID).To(Equal(parent.ID))
			Expect(createdPayload(child)).To(HaveKeyWithValue("parent", parent.ID))
		})

		It("refuses a parent that is a child, naming the TODO it belongs under", func(ctx SpecContext) {
			parent := create(ctx, provider, todos.CreateRequest{Title: "Top of the tree"})
			child := create(ctx, provider, todos.CreateRequest{Title: "Middle of the tree", Parent: parent.ID})

			_, err := provider.Create(ctx, todos.CreateRequest{Title: "Too deep", Parent: child.ID})

			Expect(err).To(MatchError(native.ErrInvalidParent))
			Expect(err).To(MatchError(ContainSubstring(parent.ShortID)))
		})

		It("refuses Parent together with NoParent", func(ctx SpecContext) {
			parent := create(ctx, provider, todos.CreateRequest{Title: "Contradicted parent"})

			_, err := provider.Create(ctx, todos.CreateRequest{Title: "Contradiction", Parent: parent.ID, NoParent: true})

			Expect(err).To(MatchError(ContainSubstring("cannot both be set")))
		})

		It("makes a TODO created inside a run a child of the running TODO", func(ctx SpecContext) {
			running := create(ctx, provider, todos.CreateRequest{Title: "The running TODO"})

			child := create(ctx, provider, todos.CreateRequest{
				Title: "Found while running", Origin: &todos.CreateOrigin{IssueID: running.ID, SessionID: session},
			})

			Expect(child.ParentID).To(Equal(running.ID))
			Expect(createdPayload(child)).To(SatisfyAll(
				HaveKeyWithValue("parent", running.ID),
				HaveKeyWithValue("originIssue", running.ID),
				HaveKeyWithValue("originSession", session),
			))
		})

		It("attaches to the running TODO's parent when the running TODO is a child", func(ctx SpecContext) {
			parent := create(ctx, provider, todos.CreateRequest{Title: "Parent of the running TODO"})
			running := create(ctx, provider, todos.CreateRequest{Title: "A running child", Parent: parent.ID})

			sibling := create(ctx, provider, todos.CreateRequest{
				Title: "Found by a child", Origin: &todos.CreateOrigin{IssueID: running.ID},
			})

			Expect(sibling.ParentID).To(Equal(parent.ID))
			Expect(createdPayload(sibling)).To(HaveKeyWithValue("originIssue", running.ID))
		})

		It("prefers an explicit parent over the origin", func(ctx SpecContext) {
			running := create(ctx, provider, todos.CreateRequest{Title: "Running, not the parent"})
			parent := create(ctx, provider, todos.CreateRequest{Title: "Named parent"})

			child := create(ctx, provider, todos.CreateRequest{
				Title: "Explicitly placed", Parent: parent.ID, Origin: &todos.CreateOrigin{IssueID: running.ID},
			})

			Expect(child.ParentID).To(Equal(parent.ID))
		})

		It("keeps the TODO top-level with NoParent and still records the origin", func(ctx SpecContext) {
			running := create(ctx, provider, todos.CreateRequest{Title: "Running, opted out"})

			standalone := create(ctx, provider, todos.CreateRequest{
				Title: "Unrelated follow-up", NoParent: true, Origin: &todos.CreateOrigin{IssueID: running.ID, SessionID: session},
			})

			Expect(standalone.ParentID).To(BeEmpty())
			Expect(createdPayload(standalone)).To(SatisfyAll(
				Not(HaveKey("parent")),
				HaveKeyWithValue("originIssue", running.ID),
				HaveKeyWithValue("originSession", session),
			))
		})

		It("leaves a TODO top-level when the origin is in another workspace, recording the reference", func(ctx SpecContext) {
			running := create(ctx, elsewhere, todos.CreateRequest{Title: "Running elsewhere"})

			created := create(ctx, provider, todos.CreateRequest{
				Title: "Created across workspaces", Origin: &todos.CreateOrigin{IssueID: running.ID},
			})

			Expect(created.ParentID).To(BeEmpty())
			Expect(createdPayload(created)).To(SatisfyAll(
				Not(HaveKey("parent")), HaveKeyWithValue("originIssue", running.ID),
			))
		})

		It("leaves a TODO top-level when the origin is not in this database", func(ctx SpecContext) {
			unknown := uuid.NewString()

			created := create(ctx, provider, todos.CreateRequest{
				Title: "Origin from another database", Origin: &todos.CreateOrigin{IssueID: unknown},
			})

			Expect(created.ParentID).To(BeEmpty())
			Expect(createdPayload(created)).To(HaveKeyWithValue("originIssue", unknown))
		})

		// The variable the origin comes from is free text — `gavel commit` takes
		// whatever is in it — so a run may export a short id or an alias.
		It("auto-links an origin given as a short id and records the TODO it resolved to", func(ctx SpecContext) {
			running := create(ctx, provider, todos.CreateRequest{Title: "Running, named by short id"})

			child := create(ctx, provider, todos.CreateRequest{
				Title: "Found by a short id", Origin: &todos.CreateOrigin{IssueID: running.ShortID},
			})

			Expect(child.ParentID).To(Equal(running.ID))
			Expect(createdPayload(child)).To(HaveKeyWithValue("originIssue", running.ID))
		})

		It("leaves a TODO top-level when the origin names nothing, recording it as given", func(ctx SpecContext) {
			created := create(ctx, provider, todos.CreateRequest{
				Title: "Unresolvable origin", Origin: &todos.CreateOrigin{IssueID: " not-a-todo-id "},
			})

			Expect(created.ParentID).To(BeEmpty())
			Expect(createdPayload(created)).To(SatisfyAll(
				Not(HaveKey("parent")), HaveKeyWithValue("originIssue", "not-a-todo-id"),
			))
		})

		It("creates with NoParent whatever the origin holds", func(ctx SpecContext) {
			created := create(ctx, provider, todos.CreateRequest{
				Title: "Opted out under an odd origin", NoParent: true, Origin: &todos.CreateOrigin{IssueID: "6408"},
			})

			Expect(created.ParentID).To(BeEmpty())
			Expect(createdPayload(created)).To(HaveKeyWithValue("originIssue", "6408"))
		})

		It("creates under an explicit parent whatever the origin holds", func(ctx SpecContext) {
			parent := create(ctx, provider, todos.CreateRequest{Title: "Parent under an odd origin"})

			child := create(ctx, provider, todos.CreateRequest{
				Title: "Placed under an odd origin", Parent: parent.ID, Origin: &todos.CreateOrigin{IssueID: "6408"},
			})

			Expect(child.ParentID).To(Equal(parent.ID))
			Expect(createdPayload(child)).To(HaveKeyWithValue("originIssue", "6408"))
		})
	})

	Describe("a closed parent", func() {
		closed := func(ctx context.Context, title string) *types.TODO {
			GinkgoHelper()
			todo := create(ctx, provider, todos.CreateRequest{Title: title})
			Expect(provider.Delete(ctx, todo)).To(Succeed())
			return todo
		}

		It("cannot be named as the parent of a new TODO", func(ctx SpecContext) {
			parent := closed(ctx, "Closed before the create")

			_, err := provider.Create(ctx, todos.CreateRequest{Title: "Filed under a closed parent", Parent: parent.ID})

			Expect(err).To(MatchError(native.ErrInvalidParent))
			Expect(err).To(MatchError(ContainSubstring(parent.ShortID + ` ("Closed before the create") is closed`)))
		})

		It("cannot adopt an existing TODO", func(ctx SpecContext) {
			parent := closed(ctx, "Closed before the adoption")
			todo := create(ctx, provider, todos.CreateRequest{Title: "Stays top-level"})

			Expect(provider.SetParent(ctx, todo, parent.ID)).To(MatchError(native.ErrInvalidParent))
			Expect(todo.ParentID).To(BeEmpty())
		})

		DescribeTable("leaves a TODO created inside a run top-level, recording the origin",
			func(ctx SpecContext, running func(ctx context.Context) *types.TODO) {
				origin := running(ctx)

				created := create(ctx, provider, todos.CreateRequest{
					Title: "Found under closed work", Origin: &todos.CreateOrigin{IssueID: origin.ID},
				})

				Expect(created.ParentID).To(BeEmpty())
				Expect(createdPayload(created)).To(SatisfyAll(
					Not(HaveKey("parent")), HaveKeyWithValue("originIssue", origin.ID),
				))
			},
			Entry("when the running TODO is closed", func(ctx context.Context) *types.TODO {
				return closed(ctx, "Closed while its run was still filing TODOs")
			}),
			Entry("when the running TODO is a child of a closed parent", func(ctx context.Context) *types.TODO {
				parent := create(ctx, provider, todos.CreateRequest{Title: "Parent closed around its child"})
				running := create(ctx, provider, todos.CreateRequest{Title: "Child left running", Parent: parent.ID})
				_, err := todos.Archive(ctx, provider, parent, todos.ArchiveOptions{Children: todos.ChildrenArchive})
				Expect(err).NotTo(HaveOccurred())
				return running
			}),
		)
	})

	Describe("closing a parent", func() {
		var parent, open, done *types.TODO

		current := func(ctx context.Context, todo *types.TODO) *types.TODO {
			GinkgoHelper()
			reloaded, err := provider.Get(ctx, todo.ID)
			Expect(err).NotTo(HaveOccurred())
			return reloaded
		}

		BeforeEach(func(ctx SpecContext) {
			parent = create(ctx, provider, todos.CreateRequest{Title: "Parent being closed"})
			open = create(ctx, provider, todos.CreateRequest{Title: "Open child", Parent: parent.ID})
			done = create(ctx, provider, todos.CreateRequest{Title: "Finished child", Parent: parent.ID})
			completed := types.StatusCompleted
			Expect(provider.UpdateState(ctx, done, todos.StateUpdate{Status: &completed})).To(Succeed())
		})

		It("Delete refuses while a child is open, saying how many and what the choices are", func(ctx SpecContext) {
			err := provider.Delete(ctx, parent)

			Expect(err).To(MatchError(native.ErrOpenChildren))
			Expect(err).To(MatchError(
				native.ErrOpenChildren.Error() + ": 1 under issue " + parent.ID + "; " + todos.ChildrenChoice))
			Expect(current(ctx, parent).Status).To(Equal(types.StatusPending))
			Expect(current(ctx, open).ParentID).To(Equal(parent.ID))
		})

		It("archives the open child with it and leaves the finished one attached", func(ctx SpecContext) {
			outcome, err := todos.Archive(ctx, provider, parent, todos.ArchiveOptions{Children: todos.ChildrenArchive})

			Expect(err).NotTo(HaveOccurred())
			Expect(outcome.String()).To(Equal("open children archived too: " + open.ShortID))
			Expect(parent.Status).To(Equal(types.StatusCompleted))
			Expect(current(ctx, open)).To(SatisfyAll(
				HaveField("Status", types.StatusCompleted), HaveField("ParentID", parent.ID)))
			Expect(current(ctx, done).ParentID).To(Equal(parent.ID))
		})

		It("makes the open child a top-level TODO and leaves the finished one attached", func(ctx SpecContext) {
			outcome, err := todos.Archive(ctx, provider, parent, todos.ArchiveOptions{Children: todos.ChildrenDetach})

			Expect(err).NotTo(HaveOccurred())
			Expect(outcome.String()).To(Equal("open children made top-level TODOs: " + open.ShortID))
			Expect(parent.Status).To(Equal(types.StatusCompleted))
			Expect(current(ctx, open)).To(SatisfyAll(
				HaveField("Status", types.StatusPending), HaveField("ParentID", BeEmpty())))
			Expect(current(ctx, done).ParentID).To(Equal(parent.ID))
		})

		It("hands the open child to a survivor, after which the parent closes", func(ctx SpecContext) {
			survivor := create(ctx, provider, todos.CreateRequest{Title: "Takes the work over"})

			outcome, err := todos.Archive(ctx, provider, parent, todos.ArchiveOptions{
				Survivor: survivor, Children: todos.ChildrenDetach,
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(outcome.String()).To(Equal("open children moved to " + survivor.ShortID + ": " + open.ShortID))
			Expect(parent.Status).To(Equal(types.StatusCompleted))
			Expect(current(ctx, open).ParentID).To(Equal(survivor.ID))
		})

		It("closes without a choice once nothing under it is open", func(ctx SpecContext) {
			Expect(provider.Delete(ctx, open)).To(Succeed())

			Expect(provider.Delete(ctx, parent)).To(Succeed())
			Expect(parent.Status).To(Equal(types.StatusCompleted))
		})
	})

	Describe("SetParent", func() {
		It("attaches and detaches, refreshing the TODO and recording parent_changed", func(ctx SpecContext) {
			parent := create(ctx, provider, todos.CreateRequest{Title: "Adopting parent"})
			todo := create(ctx, provider, todos.CreateRequest{Title: "Adopted later"})
			versionBefore := todo.Version

			Expect(provider.SetParent(ctx, todo, parent.ShortID)).To(Succeed())
			Expect(todo.ParentID).To(Equal(parent.ID))
			Expect(todo.Version).To(Equal(versionBefore + 1))
			Expect(todo.ProviderEvents[len(todo.ProviderEvents)-1].Kind).To(Equal("parent_changed"))

			Expect(provider.SetParent(ctx, todo, "")).To(Succeed())
			Expect(todo.ParentID).To(BeEmpty())
			Expect(todo.Version).To(Equal(versionBefore + 2))
		})

		It("refuses a parent that is a child, naming the TODO it belongs under", func(ctx SpecContext) {
			parent := create(ctx, provider, todos.CreateRequest{Title: "Real parent"})
			child := create(ctx, provider, todos.CreateRequest{Title: "Only a child", Parent: parent.ID})
			todo := create(ctx, provider, todos.CreateRequest{Title: "Wants a child as parent"})

			err := provider.SetParent(ctx, todo, child.ID)

			Expect(err).To(MatchError(native.ErrInvalidParent))
			Expect(err).To(MatchError(ContainSubstring(parent.ShortID)))
			Expect(todo.ParentID).To(BeEmpty())
		})

		It("refuses a parent from another workspace", func(ctx SpecContext) {
			foreign := create(ctx, elsewhere, todos.CreateRequest{Title: "Foreign parent"})
			todo := create(ctx, provider, todos.CreateRequest{Title: "Stays local"})

			Expect(provider.SetParent(ctx, todo, foreign.ID)).To(MatchError(native.ErrNotFound))
		})
	})

	It("List filters children by parent and CountByStatus leaves them out", func(ctx SpecContext) {
		counted, err := New(ctx, db, WorkspaceOptions{
			Name: "parents-counts", RootPath: GinkgoT().TempDir(), Repositories: []string{"acme/parents-counts"},
		})
		Expect(err).NotTo(HaveOccurred())
		parent := create(ctx, counted, todos.CreateRequest{Title: "Listed parent"})
		first := create(ctx, counted, todos.CreateRequest{Title: "Listed child one", Parent: parent.ID})
		second := create(ctx, counted, todos.CreateRequest{Title: "Listed child two", Parent: parent.ID})

		titles := func(filters todos.DiscoveryFilters) []string {
			listed, err := counted.List(ctx, filters)
			Expect(err).NotTo(HaveOccurred())
			names := make([]string, len(listed))
			for i, todo := range listed {
				names[i] = todo.Title
			}
			return names
		}
		Expect(titles(todos.DiscoveryFilters{})).To(ConsistOf(parent.Title, first.Title, second.Title))
		Expect(titles(todos.DiscoveryFilters{TopLevelOnly: true})).To(ConsistOf(parent.Title))
		Expect(titles(todos.DiscoveryFilters{ParentID: parent.ID})).To(ConsistOf(first.Title, second.Title))

		counts, err := counted.CountByStatus(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(counts).To(Equal(map[types.Status]int{types.StatusPending: 1}))
	})

	It("reports a database that predates the parent column as behind, naming the column", func(ctx SpecContext) {
		Expect(requireCurrentSchema(ctx, db)).To(Succeed())

		// The spec's context ends the transaction if an assertion fails first, so
		// the column is only ever missing inside it.
		tx := db.WithContext(ctx).Begin()
		Expect(tx.Error).NotTo(HaveOccurred())
		Expect(tx.Exec(`ALTER TABLE public.todo_issues DROP COLUMN parent_issue_id`).Error).NotTo(HaveOccurred())
		behind := requireCurrentSchema(ctx, tx)
		_, constructed := New(ctx, tx, WorkspaceOptions{
			Name: "parents-behind", RootPath: GinkgoT().TempDir(), Repositories: []string{"acme/parents-behind"},
		})
		var workspaces int64
		Expect(tx.Table("todo_workspaces").Where("repo_key = ?", "github.com/acme/parents-behind").
			Count(&workspaces).Error).NotTo(HaveOccurred())
		Expect(tx.Rollback().Error).NotTo(HaveOccurred())

		want := ErrSchemaBehind + " (todo_issues.parent_issue_id is missing)"
		Expect(behind).To(MatchError(want))
		Expect(constructed).To(MatchError(want), "every constructor shares the guard, not only Open")
		Expect(workspaces).To(BeZero(), "the guard runs before a workspace is initialized")
		Expect(requireCurrentSchema(ctx, db)).To(Succeed())
	})
})
