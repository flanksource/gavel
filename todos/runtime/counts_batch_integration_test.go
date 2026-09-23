package runtime

import (
	"fmt"
	"os"
	"path/filepath"

	rpchttp "github.com/flanksource/clicky/rpc/http"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

// countProjectsStatements is what countProjects costs whatever the number of
// projects: one workspace lookup, one ask-candidate scan, one grouped count.
const countProjectsStatements = 3

var _ = Describe("counting TODOs across configured projects", func() {
	var (
		db       *gorm.DB
		recorder *statementRecorder
	)

	BeforeEach(func(ctx SpecContext) {
		handle := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_todo_counts_batch"})
		GinkgoT().Setenv(database.EnvDSN, handle.DSN())
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		db = opened.Gorm()
		recorder = recordStatements(db)
	})

	project := func(name string, repositories ...string) WorkspaceOptions {
		GinkgoHelper()
		root := filepath.Join(GinkgoT().TempDir(), name)
		Expect(os.MkdirAll(root, 0o755)).To(Succeed())
		return WorkspaceOptions{Name: name, RootPath: root, Repositories: repositories}
	}

	create := func(ctx SpecContext, provider *Provider, status types.Status, n int) {
		GinkgoHelper()
		for i := range n {
			todo, err := provider.Create(ctx, todos.CreateRequest{
				Title: fmt.Sprintf("%s-%d", status, i), Body: "Body", Status: status,
			})
			Expect(err).NotTo(HaveOccurred())
			if todo.Status != status {
				Expect(provider.UpdateState(ctx, todo, todos.StateUpdate{Status: &status})).To(Succeed())
			}
		}
	}

	withPlan := func(ctx SpecContext, provider *Provider, title string) *types.TODO {
		GinkgoHelper()
		todo, err := provider.Create(ctx, todos.CreateRequest{
			Title: title, Body: "Body", Status: types.StatusPending,
			Plan: &todos.CreatePlanRequest{Markdown: "# Plan\n\nDo the thing.\n"},
		})
		Expect(err).NotTo(HaveOccurred())
		return todo
	}

	perProject := func(ctx SpecContext, options WorkspaceOptions) map[types.Status]int {
		GinkgoHelper()
		provider, err := New(ctx, db, options)
		Expect(err).NotTo(HaveOccurred())
		counts, err := provider.CountByStatus(ctx)
		Expect(err).NotTo(HaveOccurred())
		return counts
	}

	It("counts every project exactly as that project's own CountByStatus does, without creating workspaces", func(ctx SpecContext) {
		mixed := project("mixed", "acme/mixed", "acme/mixed-mirror")
		mixedProvider, err := New(ctx, db, mixed)
		Expect(err).NotTo(HaveOccurred())
		create(ctx, mixedProvider, types.StatusPending, 3)
		create(ctx, mixedProvider, types.StatusDraft, 2)
		create(ctx, mixedProvider, types.StatusCompleted, 4)
		create(ctx, mixedProvider, types.StatusVerified, 2)
		running := withPlan(ctx, mixedProvider, "Approved, then admitted to a run")
		_, err = mixedProvider.ApprovePlan(ctx, running, "reviewer", "ship it")
		Expect(err).NotTo(HaveOccurred())
		running, err = mixedProvider.Get(ctx, running.ID)
		Expect(err).NotTo(HaveOccurred())
		admission, err := mixedProvider.PrepareRun(ctx, running, todos.RunPreparation{Mode: types.ModeRun, ExecutorName: "codex"})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { mixedProvider.ownership.stop(admission.PromptRunID) })
		Expect(withPlan(ctx, mixedProvider, "Plan awaiting review").Status).To(Equal(types.StatusReview))
		approved := withPlan(ctx, mixedProvider, "Plan approved")
		_, err = mixedProvider.ApprovePlan(ctx, approved, "reviewer", "looks right")
		Expect(err).NotTo(HaveOccurred())

		// Found only by its repository: the project was opened from elsewhere.
		byRepo := project("by-repo", "acme/by-repo")
		elsewhere := byRepo
		elsewhere.RootPath = project("by-repo-elsewhere").RootPath
		byRepoProvider, err := New(ctx, db, elsewhere)
		Expect(err).NotTo(HaveOccurred())
		create(ctx, byRepoProvider, types.StatusFailed, 1)
		create(ctx, byRepoProvider, types.StatusSkipped, 2)

		// Found only by a retained path: the workspace moved, the project did not.
		retained := project("retained")
		retainedProvider, err := New(ctx, db, WorkspaceOptions{Name: "retained", RootPath: retained.RootPath, Repositories: []string{"acme/retained"}})
		Expect(err).NotTo(HaveOccurred())
		create(ctx, retainedProvider, types.StatusPending, 1)
		_, err = New(ctx, db, WorkspaceOptions{Name: "retained", RootPath: project("retained-moved").RootPath, Repositories: []string{"acme/retained"}})
		Expect(err).NotTo(HaveOccurred())

		never := project("never-opened", "acme/never-opened")
		conflicted := project("conflicted", "acme/by-repo")
		conflicted.RootPath = mixed.RootPath
		invalid := WorkspaceOptions{Name: "  ", RootPath: mixed.RootPath}

		projects := []WorkspaceOptions{mixed, byRepo, retained, never, conflicted, invalid, mixed}
		got, err := countProjects(ctx, db, projects)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(HaveLen(len(projects)))

		repository, err := native.NewRepository(db)
		Expect(err).NotTo(HaveOccurred())
		_, err = repository.GetWorkspaceByPath(ctx, never.RootPath)
		Expect(err).To(MatchError(native.ErrNotFound), "a count must not create the workspace of a never-opened project")

		Expect(got[4].Err).To(MatchError(native.ErrWorkspaceConflict))
		Expect(got[5].Err).To(MatchError(native.ErrInvalidInput))
		Expect(got[3]).To(Equal(ProjectStatusCounts{Counts: map[types.Status]int{}}))
		Expect(got[0].Counts).To(HaveKeyWithValue(types.StatusReview, 1), "the unreviewed plan must be counted as review")
		for index, options := range map[int]WorkspaceOptions{0: mixed, 1: byRepo, 2: retained, 3: never, 6: mixed} {
			Expect(got[index].Err).NotTo(HaveOccurred(), options.Name)
			Expect(got[index].Counts).To(Equal(perProject(ctx, options)), "project %s", options.Name)
		}
	})

	It("issues the same three statements for one project as for eight", func(ctx SpecContext) {
		measure := func(n int) (statements []string, timed int64) {
			GinkgoHelper()
			projects := make([]WorkspaceOptions, n)
			for i := range projects {
				repositories := []string{}
				for r := range i % 3 {
					repositories = append(repositories, fmt.Sprintf("acme/n%d-p%d-r%d", n, i, r))
				}
				projects[i] = project(fmt.Sprintf("n%d-p%d", n, i), repositories...)
				provider, err := New(ctx, db, projects[i])
				Expect(err).NotTo(HaveOccurred())
				create(ctx, provider, types.StatusPending, 2)
				withPlan(ctx, provider, "Plan awaiting review")
			}
			timedCtx, timings := rpchttp.WithTimings(ctx)
			statements = recorder.during(func() {
				counted, err := countProjects(timedCtx, db, projects)
				Expect(err).NotTo(HaveOccurred())
				for _, c := range counted {
					Expect(c).To(Equal(ProjectStatusCounts{Counts: map[types.Status]int{types.StatusPending: 2, types.StatusReview: 1}}))
				}
			})
			timed, _ = timings.Counter("sql", "queries")
			return statements, timed
		}

		one, oneTimed := measure(1)
		eight, eightTimed := measure(8)

		Expect(one).To(HaveLen(countProjectsStatements), "%q", one)
		Expect(eight).To(HaveLen(countProjectsStatements), "%q", eight)
		Expect([]int64{oneTimed, eightTimed}).To(Equal([]int64{countProjectsStatements, countProjectsStatements}),
			"the Server-Timing sql queries counter must see every statement")
	})
})

var _ = Describe("counting TODOs across configured projects with answers given outside gavel", func() {
	It("reconciles a Captain answer before counting, as CountByStatus does", func(ctx SpecContext) {
		f := newAnswerFixture(ctx)
		parked := f.parkRunAsk(ctx, "Answered before the projects list")
		f.answerInCaptain(ctx, parked)

		got, err := countProjects(ctx, f.db, []WorkspaceOptions{f.options})

		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]ProjectStatusCounts{{Counts: map[types.Status]int{types.StatusInProgress: 1}}}))
		Expect(f.eventsOf(ctx, parked.todo, native.EventAskAnswered)).To(HaveLen(1))
	})
})
