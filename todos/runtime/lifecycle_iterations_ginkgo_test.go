package runtime

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/captain/pkg/ai/agent"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The iteration rows the host files after a run are the only place a verify
// step's report lives: captain_prompt_run_overview.latest_verification_result
// is read from them by the attempt listing, the phase index and the run
// history. These specs file a verify-only run's row the way the host does and
// read it back through each of those.
var _ = Describe("run iterations on the native runtime", Ordered, func() {
	var (
		ctx      context.Context
		provider *Provider
	)

	BeforeAll(func() {
		if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
			Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres native runtime tests")
		}
		ctx = context.Background()
		dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
			DataDir: filepath.Join(GinkgoT().TempDir(), "postgres"), Database: "gavel_todo_run_iterations",
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(stop()).To(Succeed()) })

		GinkgoT().Setenv(database.EnvDSN, dsn)
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDSN, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())

		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })

		provider, err = New(ctx, opened.Gorm(), WorkspaceOptions{
			Name: "run-iterations", RootPath: GinkgoT().TempDir(), Repositories: []string{"acme/run-iterations"},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	// verifyOnlyResult is what promptrun returns for a step that generated
	// nothing: no loop, one fixture verdict on iteration 1.
	verifyOnlyResult := func(passed bool) promptrun.Result {
		started := time.Date(2026, 9, 3, 21, 23, 48, 0, time.UTC)
		finished := started.Add(45 * time.Second)
		node := api.VerifyNode{Name: "echo lifecycle-smoke", Passed: passed, Failed: !passed}
		report := api.NewNodeReport(api.VerifyKindFixture, "fixture", node)
		report.Ran, report.Iteration, report.StartedAt, report.FinishedAt = true, 1, &started, &finished
		return promptrun.Result{
			Verdicts: []agent.VerifyResult{{Valid: passed, Iteration: 1, Report: &report}},
			Report:   &report, Passed: passed,
		}
	}

	It("files a verify-only run's verdict where every reader finds it", func() {
		todo, err := provider.Create(ctx, todos.CreateRequest{Title: "Check the widget", Body: "Prove it", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		admission, err := provider.PrepareRun(ctx, todo, todos.RunPreparation{Mode: types.ModeVerify, Prompt: "verify", ExecutorName: "agent-claude"})
		Expect(err).NotTo(HaveOccurred())
		Expect(provider.RecordRunStart(ctx, todo, todos.RunStartMetadata{
			Mode: string(types.ModeVerify), Driver: "agent", Agent: "claude",
			Provider: "anthropic", RuntimeMode: "agent", ResolvedModel: "claude-sonnet", Effort: "medium",
		})).To(Succeed())

		records := promptrun.IterationRecords(verifyOnlyResult(true), false)
		Expect(records).To(HaveLen(1), "a verify-only run is one iteration")
		Expect(provider.RecordRunIterations(ctx, admission.PromptRunID, records)).To(Succeed())

		verifications, err := provider.captain.LatestPromptRunVerifications(ctx, []uuid.UUID{admission.PromptRunID})
		Expect(err).NotTo(HaveOccurred())
		Expect(verifications).To(HaveKey(admission.PromptRunID))
		Expect(verifications[admission.PromptRunID].Iteration).To(Equal(1))
		Expect(verifications[admission.PromptRunID].Report.Passed).To(BeTrue())
		Expect(verifications[admission.PromptRunID].Report.Tests).To(HaveLen(1))

		issueID, err := uuid.Parse(todo.ID)
		Expect(err).NotTo(HaveOccurred())
		phaseRuns, err := provider.repository.ListIssuePhaseRuns(ctx, provider.workspace.ID)
		Expect(err).NotTo(HaveOccurred())
		var verifyRun *native.IssuePhaseRun
		for i := range phaseRuns {
			if phaseRuns[i].IssueID == issueID && phaseRuns[i].Phase == native.StepVerify {
				verifyRun = &phaseRuns[i]
			}
		}
		Expect(verifyRun).NotTo(BeNil(), "the phase index lists the verify run")
		var stored api.VerifyReport
		Expect(json.Unmarshal([]byte(verifyRun.VerificationResult), &stored)).To(Succeed())
		Expect(stored.Passed).To(BeTrue())
		Expect(stored.Summary.Passed).To(Equal(1))

		history, err := provider.repository.ListIssueRunHistory(ctx, issueID, "lifecycle_outcome")
		Expect(err).NotTo(HaveOccurred())
		Expect(history).To(HaveLen(1))
		Expect(history[0].PromptRunID).To(Equal(admission.PromptRunID))
		Expect(string(history[0].Phase)).To(Equal("verify"))
	})

	It("restates a failed verdict on replay and refuses a run without an identity", func() {
		todo, err := provider.Create(ctx, todos.CreateRequest{Title: "Check the gadget", Body: "Prove it", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		admission, err := provider.PrepareRun(ctx, todo, todos.RunPreparation{Mode: types.ModeVerify, Prompt: "verify", ExecutorName: "agent-claude"})
		Expect(err).NotTo(HaveOccurred())

		Expect(provider.RecordRunIterations(ctx, admission.PromptRunID, promptrun.IterationRecords(verifyOnlyResult(true), false))).To(Succeed())
		Expect(provider.RecordRunIterations(ctx, admission.PromptRunID, promptrun.IterationRecords(verifyOnlyResult(false), false))).To(Succeed())

		verifications, err := provider.captain.LatestPromptRunVerifications(ctx, []uuid.UUID{admission.PromptRunID})
		Expect(err).NotTo(HaveOccurred())
		Expect(verifications[admission.PromptRunID].Report.Passed).To(BeFalse(), "the replay is the run's current statement")

		err = provider.RecordRunIterations(ctx, uuid.Nil, []captaindb.UpsertPromptRunIterationInput{{Iteration: 1}})
		Expect(err).To(MatchError(ContainSubstring("prompt run ID is required")))
	})
})
