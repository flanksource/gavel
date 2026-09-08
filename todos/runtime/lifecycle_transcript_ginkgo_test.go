package runtime

import (
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("run start transcript binding", func() {
	var provider *Provider

	BeforeEach(func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_run_transcript"})
		GinkgoT().Setenv(database.EnvDSN, db.DSN())
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		// Transcript registration resolves paths under the agent's home; an
		// empty one keeps these specs off the developer's real session logs.
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		provider, err = New(ctx, opened.Gorm(), WorkspaceOptions{
			Name: "run-transcript", RootPath: GinkgoT().TempDir(), Repositories: []string{"acme/run-transcript"},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	// startRun admits a run and reports its start exactly as an executor does.
	startRun := func(ctx SpecContext, title string, metadata todos.RunStartMetadata) *captaindb.Session {
		GinkgoHelper()
		todo, err := provider.Create(ctx, todos.CreateRequest{Title: title, Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		admission, err := provider.PrepareRun(ctx, todo, todos.RunPreparation{
			Mode: types.ModeRun, Prompt: "run", ExecutorName: "claude",
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { provider.clearPrepared(uuid.MustParse(todo.ID), admission.PromptRunID) })
		Expect(provider.RecordRunStart(ctx, todo, metadata)).To(Succeed())
		run, err := provider.Captain().GetPromptRun(ctx, admission.PromptRunID)
		Expect(err).NotTo(HaveOccurred())
		operation, err := provider.Captain().GetSession(ctx, run.SessionID)
		Expect(err).NotTo(HaveOccurred())
		return operation
	}

	// Captain's own recovery path looks for a `transcript` child under a run's
	// session. Recorded as `agent` — a sub-agent — the row holding the provider
	// log was invisible to it even when everything else had worked.
	It("makes the agent session findable as the run's transcript child", func(ctx SpecContext) {
		operation := startRun(ctx, "Bind the provider transcript", todos.RunStartMetadata{
			SessionID: "0199f0aa-0000-7000-8000-000000000001", Provider: "claude", Mode: "run",
		})

		transcript, err := provider.Captain().GetTranscriptSession(ctx, operation.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(transcript.Source).To(Equal("claude"))
		Expect(transcript.ParentRelation).To(Equal(captaindb.SessionParentRelationTranscript))
		Expect(transcript.ProviderSessionID).To(Equal("0199f0aa-0000-7000-8000-000000000001"))
	})

	It("resolves the transcript-bearing sibling from the provider identity alone", func(ctx SpecContext) {
		providerSessionID := "0199f0aa-0000-7000-8000-000000000002"
		operation := startRun(ctx, "Hop to the transcript sibling", todos.RunStartMetadata{
			SessionID: providerSessionID, Provider: "claude", Mode: "run",
		})

		resolved, err := provider.Captain().GetTranscriptSessionByIdentity(ctx, providerSessionID)
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.ID).NotTo(Equal(operation.ID), "the admission root holds no transcript")
		Expect(resolved.Source).To(Equal("claude"))
	})

	// The worktree is built by setup and reported only on the spec that trails
	// it. A run that blocks before its first turn boundary never reports one any
	// other way, which is why no session in the live database had one.
	It("records the worktree the agent works in, not the repository it was filed against", func(ctx SpecContext) {
		worktree := "/work/repo/.shell/worktrees/shell-9f3a-0199f0aa"
		operation := startRun(ctx, "Record the run worktree", todos.RunStartMetadata{
			SessionID: "0199f0aa-0000-7000-8000-000000000003", Provider: "claude", Mode: "run",
			Spec: &api.Spec{Setup: &shell.Setup{Cwd: worktree}},
		})

		Expect(operation.CWD).To(Equal(worktree))

		transcript, err := provider.Captain().GetTranscriptSession(ctx, operation.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(transcript.CWD).To(Equal(worktree), "the agent row inherits the directory the agent runs in")
	})

	It("leaves the recorded worktree alone on a turn that carries no spec", func(ctx SpecContext) {
		worktree := "/work/repo/.shell/worktrees/shell-aaaa-0199f0aa"
		todo, err := provider.Create(ctx, todos.CreateRequest{Title: "Keep the worktree", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		admission, err := provider.PrepareRun(ctx, todo, todos.RunPreparation{
			Mode: types.ModeRun, Prompt: "run", ExecutorName: "claude",
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { provider.clearPrepared(uuid.MustParse(todo.ID), admission.PromptRunID) })
		Expect(provider.RecordRunStart(ctx, todo, todos.RunStartMetadata{
			SessionID: "0199f0aa-0000-7000-8000-000000000004", Provider: "claude", Mode: "run",
			Spec: &api.Spec{Setup: &shell.Setup{Cwd: worktree}},
		})).To(Succeed())

		Expect(provider.RecordRunStart(ctx, todo, todos.RunStartMetadata{
			SessionID: "0199f0aa-0000-7000-8000-000000000004", Provider: "claude", Mode: "run",
		})).To(Succeed())

		run, err := provider.Captain().GetPromptRun(ctx, admission.PromptRunID)
		Expect(err).NotTo(HaveOccurred())
		operation, err := provider.Captain().GetSession(ctx, run.SessionID)
		Expect(err).NotTo(HaveOccurred())
		Expect(operation.CWD).To(Equal(worktree))
	})
})
