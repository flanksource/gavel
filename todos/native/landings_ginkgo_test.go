package native_test

import (
	"encoding/json"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("run landings", func() {
	var (
		repository *native.Repository
		issue      *native.Issue
		runID      uuid.UUID
	)

	BeforeEach(func(ctx SpecContext) {
		handle := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_native_landings"})
		GinkgoT().Setenv(database.EnvDSN, handle.DSN())
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDSN, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		repository, err = native.NewRepository(opened.Gorm())
		Expect(err).NotTo(HaveOccurred())

		workspace, err := repository.CreateWorkspace(ctx, native.CreateWorkspaceInput{
			RepoKey: "github.com/acme/landings", RootPath: GinkgoT().TempDir(),
		})
		Expect(err).NotTo(HaveOccurred())
		issue, err = repository.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: workspace.ID, Title: "Land the run"})
		Expect(err).NotTo(HaveOccurred())
		runID = uuid.New()
		_, err = repository.LinkPromptRun(ctx, issue.ID, runID, native.StepRun, 0, issue.Version, "test")
		Expect(err).NotTo(HaveOccurred())
	})

	It("RecordLanding writes the row and a run_landed event", func(ctx SpecContext) {
		prNumber := 42
		recorded, err := repository.RecordLanding(ctx, issue.ID, native.RunLanding{
			PromptRunID: runID, Via: native.LandingPR, TargetBranch: "main",
			LandedSHA: "0123456789abcdef0123456789abcdef01234567",
			PRNumber:  &prNumber, PRURL: "https://github.com/acme/landings/pull/42", CommitCount: 3,
		}, "tester")
		Expect(err).NotTo(HaveOccurred())
		Expect(recorded.LandedAt).NotTo(BeZero())
		Expect(recorded.CommitCount).To(Equal(3))

		landings, err := repository.ListLandings(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(landings).To(Equal([]native.RunLanding{*recorded}))
		Expect(landings[0].PromptRunID).To(Equal(runID))
		Expect(landings[0].PRNumber).To(HaveValue(Equal(prNumber)))
		Expect(landings[0].BranchDeletedAt).To(BeNil())

		events, err := repository.ListEvents(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		last := events[len(events)-1]
		Expect(last.Kind).To(Equal("run_landed"))
		Expect(last.Actor).To(Equal("tester"))
		var payload map[string]any
		Expect(json.Unmarshal(last.Payload, &payload)).To(Succeed())
		Expect(payload).To(Equal(map[string]any{
			"promptRunId": runID.String(), "via": "pr", "targetBranch": "main",
			"landedSha": "0123456789abcdef0123456789abcdef01234567",
			"prNumber":  float64(prNumber), "prUrl": "https://github.com/acme/landings/pull/42",
			"commitCount": float64(3),
		}))
	})

	It("ListLandingsForRuns returns only the named runs' landings, keyed by run", func(ctx SpecContext) {
		otherRun := uuid.New()
		current, err := repository.GetIssue(ctx, issue.ID)
		Expect(err).NotTo(HaveOccurred())
		_, err = repository.LinkPromptRun(ctx, issue.ID, otherRun, native.StepRun, 1, current.Version, "test")
		Expect(err).NotTo(HaveOccurred())
		recorded, err := repository.RecordLanding(ctx, issue.ID, native.RunLanding{
			PromptRunID: runID, Via: native.LandingMerge, TargetBranch: "main", LandedSHA: "abc1234", CommitCount: 1,
		}, "tester")
		Expect(err).NotTo(HaveOccurred())
		_, err = repository.RecordLanding(ctx, issue.ID, native.RunLanding{
			PromptRunID: otherRun, Via: native.LandingMerge, TargetBranch: "main", LandedSHA: "def5678", CommitCount: 1,
		}, "tester")
		Expect(err).NotTo(HaveOccurred())

		landings, err := repository.ListLandingsForRuns(ctx, []uuid.UUID{runID, uuid.New()})
		Expect(err).NotTo(HaveOccurred())
		Expect(landings).To(Equal(map[uuid.UUID]native.RunLanding{runID: *recorded}))

		empty, err := repository.ListLandingsForRuns(ctx, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(empty).To(BeEmpty())
	})

	It("refuses to land the same run twice", func(ctx SpecContext) {
		landing := native.RunLanding{PromptRunID: runID, Via: native.LandingMerge, TargetBranch: "main", LandedSHA: "abc1234", CommitCount: 1}
		_, err := repository.RecordLanding(ctx, issue.ID, landing, "tester")
		Expect(err).NotTo(HaveOccurred())
		_, err = repository.RecordLanding(ctx, issue.ID, landing, "tester")
		Expect(err).To(MatchError(native.ErrAlreadyLanded))
	})

	It("refuses a landing for a run that is not linked to the issue", func(ctx SpecContext) {
		_, err := repository.RecordLanding(ctx, issue.ID, native.RunLanding{
			PromptRunID: uuid.New(), Via: native.LandingMerge, TargetBranch: "main", LandedSHA: "abc1234", CommitCount: 1,
		}, "tester")
		Expect(err).To(MatchError(native.ErrLinkConflict))
	})

	DescribeTable("rejects an inconsistent landing before writing", func(landing native.RunLanding) {
		landing.PromptRunID = runID
		_, err := repository.RecordLanding(GinkgoT().Context(), issue.ID, landing, "tester")
		Expect(err).To(MatchError(native.ErrInvalidInput))
		landings, err := repository.ListLandings(GinkgoT().Context(), issue.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(landings).To(BeEmpty())
	},
		Entry("unknown via", native.RunLanding{CommitCount: 1, Via: "rebase", TargetBranch: "main", LandedSHA: "abc1234"}),
		Entry("missing landed sha", native.RunLanding{CommitCount: 1, Via: native.LandingMerge, TargetBranch: "main"}),
		Entry("missing target branch", native.RunLanding{CommitCount: 1, Via: native.LandingMerge, LandedSHA: "abc1234"}),
		Entry("pr without a number", native.RunLanding{CommitCount: 1, Via: native.LandingPR, TargetBranch: "main", LandedSHA: "abc1234", PRURL: "https://x/1"}),
		Entry("no landed commits", native.RunLanding{Via: native.LandingMerge, TargetBranch: "main", LandedSHA: "abc1234"}),
		Entry("merge carrying a PR", native.RunLanding{CommitCount: 1, Via: native.LandingMerge, TargetBranch: "main", LandedSHA: "abc1234", PRURL: "https://x/1"}),
	)
})
