package runtime

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("durable run provenance", func() {
	var provider *Provider
	BeforeEach(func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_run_provenance"})
		GinkgoT().Setenv(database.EnvDSN, db.DSN())
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		provider, err = New(ctx, opened.Gorm(), WorkspaceOptions{
			Name: "run-provenance", RootPath: GinkgoT().TempDir(), Repositories: []string{"acme/run-provenance"},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("keeps the admitted profile and trace after recording the prepared execution spec", func(ctx SpecContext) {
		todo, err := provider.Create(ctx, todos.CreateRequest{Title: "Keep resolution provenance", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		profile, trace := renderedProfileFixture()
		admission, err := provider.PrepareRun(ctx, todo, todos.RunPreparation{
			Mode: types.ModeRun, Prompt: "run", ExecutorName: "example",
			Spec: trace[0].Spec, RuntimeProfile: profile, SpecTrace: trace,
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { provider.clearPrepared(uuid.MustParse(todo.ID), admission.PromptRunID) })
		admitted, err := provider.Captain().GetPromptRun(ctx, admission.PromptRunID)
		Expect(err).NotTo(HaveOccurred())
		Expect(admitted.RenderedSpec).To(HaveKey("runtimeProfile"))
		Expect(admitted.RenderedSpec).To(HaveKey("specTrace"))
		prepared := api.Spec{Setup: &shell.Setup{Cwd: "/work/prepared"}}
		Expect(provider.RecordRunStart(ctx, todo, todos.RunStartMetadata{Mode: "run", Spec: &prepared})).To(Succeed())
		started, err := provider.Captain().GetPromptRun(ctx, admission.PromptRunID)
		Expect(err).NotTo(HaveOccurred())
		Expect(started.RenderedSpec).To(HaveKeyWithValue("runtimeProfile", admitted.RenderedSpec["runtimeProfile"]))
		Expect(started.RenderedSpec).To(HaveKeyWithValue("specTrace", admitted.RenderedSpec["specTrace"]))
		Expect(started.RenderedSpec["setup"]).To(HaveKeyWithValue("cwd", "/work/prepared"))
	})
})
