package todoprojection

import (
	"context"
	"os"
	"path/filepath"
	"time"

	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/gavel/internal/database"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

// The projection is asynchronous: Captain commits, then notifies. Every
// expectation on the watermark therefore waits for it rather than reading once.
const projectionDeadline = 10 * time.Second

var _ = Describe("Projection listening to Captain row changes", Ordered, func() {
	var (
		db        *gorm.DB
		fixture   activeRunFixture
		base      time.Time
		stop      context.CancelFunc
		runResult chan error
	)
	watermarkOf := func(issueID uuid.UUID) func() time.Time {
		return func() time.Time {
			var at time.Time
			Expect(db.Raw(`SELECT updated_at FROM todo_issues WHERE id = ?`, issueID).Scan(&at).Error).To(Succeed())
			return at.UTC()
		}
	}
	var watermark func() time.Time
	minutesAfterBase := func(minutes int) time.Time { return base.Add(time.Duration(minutes) * time.Minute) }

	BeforeAll(func() {
		db = openMigratedDatabase()
		// Captain stamps updated_at with clock_timestamp() on every UPDATE, so
		// activity is dated in the future to be the only value that can win.
		base = time.Now().Add(time.Hour).UTC().Truncate(time.Microsecond)
		fixture = newActiveRunFixture(db, base)
		watermark = watermarkOf(fixture.issueID)
		// The watermark lags the agent activity recorded while nobody listened.
		Expect(db.Exec(`UPDATE captain_sessions SET last_activity_at = ? WHERE id = ?`,
			minutesAfterBase(1), fixture.agentSessionID).Error).To(Succeed())

		projection, err := New(db)
		Expect(err).NotTo(HaveOccurred())
		var ctx context.Context
		ctx, stop = context.WithCancel(context.Background())
		DeferCleanup(func() { stop() })
		runResult = make(chan error, 1)
		go func() { runResult <- projection.Run(ctx) }()
		Eventually(projection.Ready(), projectionDeadline).Should(BeClosed())
	})

	It("catches up on activity it missed once its LISTEN is established", func() {
		Expect(watermark()).To(Equal(minutesAfterBase(1)))
	})

	It("advances the TODO from its agent session family", func() {
		Expect(db.Exec(`UPDATE captain_sessions SET last_activity_at = ?, activity_state = 'ask' WHERE id = ?`,
			minutesAfterBase(2), fixture.agentSessionID).Error).To(Succeed())
		Eventually(watermark, projectionDeadline).Should(Equal(minutesAfterBase(2)))
	})

	It("advances the TODO from a turn request on its active run", func() {
		Expect(db.Exec(`
			INSERT INTO captain_turn_requests
				(id, session_id, prompt_run_id, kind, state, request, version, created_at)
			VALUES (?, ?, ?, 'question', 'pending', '{}'::jsonb, 0, ?)`,
			fixture.requestID, fixture.agentSessionID, fixture.runID, minutesAfterBase(3)).Error).To(Succeed())
		Eventually(watermark, projectionDeadline).Should(Equal(minutesAfterBase(3)))
	})

	It("advances the TODO from a session in its own root session tree", func() {
		Expect(db.Exec(`
			INSERT INTO captain_sessions
				(id, source, provider, host_id, parent_session_id, root_session_id, last_activity_at)
			VALUES (?, 'gavel', 'test', 'local', ?, ?, ?)`,
			uuid.New(), fixture.issueID, fixture.issueID, minutesAfterBase(4)).Error).To(Succeed())
		Eventually(watermark, projectionDeadline).Should(Equal(minutesAfterBase(4)))
	})

	It("never moves the watermark back for an older observation", func() {
		Expect(db.Exec(`UPDATE captain_sessions SET last_activity_at = ? WHERE id = ?`,
			minutesAfterBase(0), fixture.agentSessionID).Error).To(Succeed())
		Consistently(watermark, time.Second).Should(Equal(minutesAfterBase(4)))
	})

	It("dates a deleted turn request by when the projection observed it", func() {
		// A deleted row has no timestamps left, so this TODO's activity is in the
		// past: only the observation time of the delete can advance it. Captain
		// has no trigger on a turn request DELETE, so nothing else moves it.
		past := newActiveRunFixture(db, time.Now().Add(-time.Hour).UTC())
		pastWatermark := watermarkOf(past.issueID)
		Expect(db.Exec(`
			INSERT INTO captain_turn_requests (id, session_id, prompt_run_id, kind, state, request, version)
			VALUES (?, ?, ?, 'question', 'pending', '{}'::jsonb, 0)`,
			past.requestID, past.agentSessionID, past.runID).Error).To(Succeed())
		// The insert's own notifications (and those of the session Captain's
		// activity trigger touches) arrive over time; wait for a quiet window.
		var settled time.Time
		Eventually(func() bool {
			settled = pastWatermark()
			time.Sleep(300 * time.Millisecond)
			return pastWatermark().Equal(settled)
		}, projectionDeadline).Should(BeTrue())
		Expect(settled).To(BeTemporally(">", time.Now().Add(-time.Minute)), "the insert itself must have been projected")
		Expect(db.Exec(`DELETE FROM captain_turn_requests WHERE id = ?`, past.requestID).Error).To(Succeed())
		Eventually(pastWatermark, projectionDeadline).Should(BeTemporally(">", settled))
	})

	It("stops with the context that runs it", func() {
		stop()
		Eventually(runResult, projectionDeadline).Should(Receive(MatchError(context.Canceled)))
	})
})

type activeRunFixture struct {
	issueID, admissionSessionID, agentSessionID, runID, requestID uuid.UUID
}

// newActiveRunFixture links one running prompt run as a fresh TODO's active
// run, with every Captain timestamp and the TODO's watermark at base: the
// admission root is the source='gavel' session, the monitored agent session
// shares its provider identity, and the TODO's own root session has its id.
func newActiveRunFixture(db *gorm.DB, base time.Time) activeRunFixture {
	GinkgoHelper()
	fixture := activeRunFixture{
		issueID: uuid.New(), admissionSessionID: uuid.New(), agentSessionID: uuid.New(),
		runID: uuid.New(), requestID: uuid.New(),
	}
	workspaceID := uuid.New()
	providerSessionID := uuid.NewString()
	for _, statement := range []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO todo_workspaces (id, repo_key) VALUES (?, ?)`,
			[]any{workspaceID, "github.com/example/projection-" + workspaceID.String()}},
		{`INSERT INTO todo_issues (id, workspace_id, title) VALUES (?, ?, 'Projected')`,
			[]any{fixture.issueID, workspaceID}},
		{`INSERT INTO captain_sessions (id, source, provider, host_id, updated_at) VALUES (?, 'gavel', 'test', 'local', ?)`,
			[]any{fixture.issueID, base}},
		{`INSERT INTO captain_sessions (id, source, provider, provider_session_id, host_id, updated_at)
		  VALUES (?, 'gavel', 'test', ?, 'local', ?)`,
			[]any{fixture.admissionSessionID, providerSessionID, base}},
		{`INSERT INTO captain_sessions (id, source, provider_session_id, host_id, lifecycle_status, updated_at)
		  VALUES (?, 'codex', ?, 'local', 'running', ?)`,
			[]any{fixture.agentSessionID, providerSessionID, base}},
		{`INSERT INTO captain_prompt_runs (id, session_id, root_session_id, origin, phase, state, version, updated_at)
		  VALUES (?, ?, ?, 'gavel-test', 'generate', 'running', 0, ?)`,
			[]any{fixture.runID, fixture.admissionSessionID, fixture.admissionSessionID, base}},
		{`INSERT INTO todo_issue_prompt_runs (issue_id, prompt_run_id, step_kind, ordinal) VALUES (?, ?, 'run', 0)`,
			[]any{fixture.issueID, fixture.runID}},
		{`UPDATE todo_issues SET active_prompt_run_id = ?, updated_at = ? WHERE id = ?`,
			[]any{fixture.runID, base, fixture.issueID}},
	} {
		Expect(db.Exec(statement.sql, statement.args...).Error).To(Succeed(), statement.sql)
	}
	return fixture
}

func openMigratedDatabase() *gorm.DB {
	GinkgoHelper()
	if os.Getenv("GAVEL_DB_EMBEDDED_TEST") == "" {
		Skip("set GAVEL_DB_EMBEDDED_TEST=1 to run embedded-postgres projection tests")
	}
	dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
		DataDir:  filepath.Join(GinkgoT().TempDir(), "postgres"),
		Database: "gavel_todo_projection",
	})
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(stop()).To(Succeed()) })
	GinkgoT().Setenv(database.EnvDSN, dsn)
	GinkgoT().Setenv(database.EnvDisable, "")
	GinkgoT().Setenv(database.LegacyEnvDSN, "")
	GinkgoT().Setenv(database.LegacyEnvDisable, "")
	GinkgoT().Setenv("HOME", GinkgoT().TempDir())
	opened, err := database.Open(context.Background(), database.WithMigrations())
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
	return opened.Gorm()
}
