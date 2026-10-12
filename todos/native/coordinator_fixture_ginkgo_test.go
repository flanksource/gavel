package native_test

import (
	"context"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

// coordinatorFixture is one migrated embedded database shared by an Ordered
// container: the Gavel repository and the Captain handle use the same pool, as
// LaunchCoordinator requires.
type coordinatorFixture struct {
	db          *gorm.DB
	repository  *native.Repository
	captain     *captaindb.DB
	coordinator *native.LaunchCoordinator
	workspace   *native.Workspace
}

func openCoordinatorFixture(ctx context.Context, name string) *coordinatorFixture {
	GinkgoHelper()
	dsn := dbtest.ForGinkgo(dbtest.Options{Name: name}).DSN()

	GinkgoT().Setenv(database.EnvDSN, dsn)
	GinkgoT().Setenv(database.EnvDisable, "")
	GinkgoT().Setenv(database.LegacyEnvDSN, "")
	GinkgoT().Setenv(database.LegacyEnvDisable, "")
	GinkgoT().Setenv("HOME", GinkgoT().TempDir())

	opened, err := database.Open(ctx, database.WithMigrations())
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })

	fixture := &coordinatorFixture{db: opened.Gorm()}
	fixture.repository, err = native.NewRepository(opened.Gorm())
	Expect(err).NotTo(HaveOccurred())
	fixture.captain, err = captaindb.Use(opened.Gorm())
	Expect(err).NotTo(HaveOccurred())
	fixture.coordinator, err = native.NewLaunchCoordinator(fixture.captain, fixture.repository)
	Expect(err).NotTo(HaveOccurred())
	fixture.workspace, err = fixture.repository.CreateWorkspace(ctx, native.CreateWorkspaceInput{
		RepoKey: "github.com/acme/" + name, RootPath: GinkgoT().TempDir(),
	})
	Expect(err).NotTo(HaveOccurred())
	return fixture
}

func (f *coordinatorFixture) createIssue(ctx context.Context, title string) *native.Issue {
	GinkgoHelper()
	issue, err := f.repository.CreateIssue(ctx, native.CreateIssueInput{WorkspaceID: f.workspace.ID, Title: title})
	Expect(err).NotTo(HaveOccurred())
	return issue
}

func (f *coordinatorFixture) lastEventKind(ctx context.Context, issueID uuid.UUID) string {
	GinkgoHelper()
	events, err := f.repository.ListEvents(ctx, issueID)
	Expect(err).NotTo(HaveOccurred())
	Expect(events).NotTo(BeEmpty())
	return events[len(events)-1].Kind
}

func (f *coordinatorFixture) eventCount(ctx context.Context, issueID uuid.UUID) int {
	GinkgoHelper()
	events, err := f.repository.ListEvents(ctx, issueID)
	Expect(err).NotTo(HaveOccurred())
	return len(events)
}

func (f *coordinatorFixture) expectNoSession(ctx context.Context, id uuid.UUID) {
	GinkgoHelper()
	_, err := f.captain.GetSession(ctx, id)
	Expect(err).To(MatchError(captaindb.ErrSessionNotFound), "session %s should have been rolled back", id)
}

// todoSessionTree is the TODO root plus one operation session under it, the
// shape Gavel hands Captain for every issue-scoped session.
func todoSessionTree(issueID, sessionID uuid.UUID, operation string) (captaindb.CreateSessionInput, captaindb.CreateSessionInput) {
	root := captaindb.CreateSessionInput{
		ID: issueID, Source: "gavel", Provider: "todos", CWD: "/workspace/acme", AgentType: "todo",
	}
	session := captaindb.CreateSessionInput{
		ID: sessionID, Source: "gavel", Provider: "test", CWD: "/workspace/acme",
		ParentSessionID: &issueID, AgentType: operation,
	}
	return root, session
}

// releaseOnce unblocks a held independent Captain transaction; deferring it as
// well keeps a failed spec from leaving that transaction open.
func releaseOnce(release chan struct{}) {
	select {
	case <-release:
	default:
		close(release)
	}
}

// waitForBlockedPlanLock waits until another backend is queued on a row lock,
// i.e. the coordinator is parked behind an independent Captain writer.
func waitForBlockedPlanLock(ctx context.Context, db *gorm.DB) {
	GinkgoHelper()
	Eventually(func(g Gomega) {
		var waiters int64
		g.Expect(db.WithContext(ctx).Raw(`
			SELECT COUNT(*) FROM pg_stat_activity
			WHERE datname = current_database() AND wait_event_type = 'Lock'`,
		).Scan(&waiters).Error).To(Succeed())
		g.Expect(waiters).To(BeNumerically(">", 0))
	}).WithTimeout(5 * time.Second).WithPolling(10 * time.Millisecond).Should(Succeed())
}
