package todoprojection

import (
	"context"
	"github.com/flanksource/gavel/todos/native"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TODO overview", func() {
	It("carries monitored session timing and cost through the repository", func() {
		db := openMigratedDatabase()
		started := time.Now().UTC().Add(-time.Minute).Truncate(time.Second)
		fixture := newActiveRunFixture(db, started)
		Expect(db.Exec(`UPDATE captain_sessions SET path = '/logs/session.jsonl' WHERE id = ?`, fixture.agentSessionID).Error).To(Succeed())
		Expect(db.Exec(`INSERT INTO captain_sessions (id, source, provider_session_id, host_id) SELECT ?, 'codex', provider_session_id, 'other' FROM captain_sessions WHERE id = ?`, uuid.New(), fixture.agentSessionID).Error).To(Succeed())
		Expect(db.Exec(`UPDATE captain_sessions SET started_at = ?, last_activity_at = ?, lifecycle_status = 'succeeded', activity_state = 'idle' WHERE id = ?`, started, started.Add(15*time.Second), fixture.agentSessionID).Error).To(Succeed())
		child, transcript := uuid.New(), uuid.New()
		Expect(db.Exec(`INSERT INTO captain_sessions (id, source, provider_session_id, host_id, parent_session_id, root_session_id, parent_relation, lifecycle_status, last_activity_at) VALUES (?, 'codex', ?, 'local', ?, ?, 'agent', 'succeeded', ?)`, child, child.String(), fixture.agentSessionID, fixture.agentSessionID, started.Add(20*time.Second)).Error).To(Succeed())
		Expect(db.Exec(`INSERT INTO captain_sessions (id, source, provider_session_id, host_id, parent_session_id, root_session_id, parent_relation) VALUES (?, 'codex', ?, 'local', ?, ?, 'transcript')`, transcript, transcript.String(), fixture.agentSessionID, fixture.agentSessionID).Error).To(Succeed())
		Expect(db.Exec(`UPDATE captain_session_overview_rows SET cost_usd = 2.5 WHERE id = ?`, child).Error).To(Succeed())
		Expect(db.Exec(`UPDATE captain_session_overview_rows SET cost_usd = 100 WHERE id = ?`, transcript).Error).To(Succeed())
		Expect(db.Exec(`UPDATE captain_session_overview_rows SET cost_usd = 1.25 WHERE id = ?`, fixture.agentSessionID).Error).To(Succeed())
		repository, err := native.NewRepository(db)
		Expect(err).NotTo(HaveOccurred())
		issue, err := repository.GetIssue(context.Background(), fixture.issueID)
		Expect(err).NotTo(HaveOccurred())
		Expect(issue.Overview.SessionSummary).NotTo(BeNil())
		Expect(issue.Overview.SessionSummary.DurationMS).To(Equal(int64(20000)))
		Expect(issue.Overview.SessionSummary.CostUSD).To(Equal(3.75))
		Expect(issue.Overview.SessionSummary.State).To(Equal("completed"))
		Expect(issue.Overview.SessionSummary.InProgress).To(BeFalse())
		Expect(db.Exec(`UPDATE captain_sessions SET lifecycle_status = 'running' WHERE id = ?`, child).Error).To(Succeed())
		issue, err = repository.GetIssue(context.Background(), fixture.issueID)
		Expect(err).NotTo(HaveOccurred())
		Expect(issue.Overview.SessionSummary.InProgress).To(BeTrue())
		Expect(issue.Overview.SessionSummary.DurationMS).To(BeNumerically(">=", int64(60000)))
	})
	It("joins the exact selected plan independently of the active run", func() {
		db := openMigratedDatabase()
		fixture := newActiveRunFixture(db, time.Now().UTC())
		planID := uuid.New()
		Expect(db.Exec(`INSERT INTO captain_plans (id, source_session_id, title) VALUES (?, ?, 'Selected older plan')`, planID, fixture.admissionSessionID).Error).To(Succeed())
		Expect(db.Exec(`INSERT INTO captain_plan_revisions (plan_id, revision, plan_markdown, content_hash) VALUES (?, 1, 'first', 'a'), (?, 2, 'second', 'b')`, planID, planID).Error).To(Succeed())
		Expect(db.Exec(`INSERT INTO todo_issue_plans (issue_id, plan_id, ordinal) VALUES (?, ?, 0)`, fixture.issueID, planID).Error).To(Succeed())
		Expect(db.Exec(`UPDATE todo_issues SET selected_plan_id = ? WHERE id = ?`, planID, fixture.issueID).Error).To(Succeed())
		var row struct {
			IssueID       uuid.UUID
			PlanRevision  int
			ApprovalState string
			StepKind      string
			Attempts      int
		}
		Expect(db.Raw(`SELECT issue_id, plan_revision, approval_state, step_kind, attempts FROM todo_issue_overview WHERE issue_id = ?`, fixture.issueID).Scan(&row).Error).To(Succeed())
		Expect(row.IssueID).To(Equal(fixture.issueID))
		Expect(row.PlanRevision).To(Equal(2))
		Expect(row.ApprovalState).To(Equal("pending"))
		Expect(row.StepKind).To(Equal("run"))
		Expect(row.Attempts).To(Equal(1))
	})
})
