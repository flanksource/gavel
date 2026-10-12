package lifecycle

import (
	"context"
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/promptrun"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/todos"
	todoprompt "github.com/flanksource/gavel/todos/prompt"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Triage new recorded outcome", func() {
	ginkgo.It("records an unapproved relationship as failed", func() {
		prepared := &preparedStep{definition: todoprompt.Definition{Name: "triage.new", Class: types.ModePlan, Envelope: todoprompt.EnvelopeTriageNew}, class: types.ModePlan}
		input := &stepInput{execution: &todos.ExecutionResult{}}
		classify := (&Host{}).recordedOutcome(context.Background(), Step{Name: "triage.new"}, prepared, input, time.Now())
		outcome, err := classify(promptrun.Result{Passed: true, Response: &api.Response{Text: `{"summary":"Covered by existing work","endStatus":"completed","title":"Repair parser","labels":["bug"],"action":"duplicate-of","target":"target","rationale":"Same scope","proposalId":"fabricated"}`}}, nil, false)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(outcome.State).To(gomega.Equal(captaindb.PromptRunStateFailed))
		gomega.Expect(outcome.Error).To(gomega.ContainSubstring("approval"))
	})
})

var _ = ginkgo.Describe("Triage new durable approval", ginkgo.Ordered, func() {
	var db *captaindb.DB
	ginkgo.BeforeAll(func(ctx ginkgo.SpecContext) {
		opened, err := captaindb.Open(ctx, captaindb.WithDSN(dbtest.ForGinkgo(dbtest.Options{Name: "gavel_triage_new_approvals"}).DSN()), captaindb.WithMigrations())
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		db = opened
		ginkgo.DeferCleanup(func() { gomega.Expect(opened.Close()).To(gomega.Succeed()) })
	})
	ginkgo.DescribeTable("requires an exact approved decision in this prompt run", func(ctx ginkgo.SpecContext, state string, amended bool, accepted bool) {
		session, err := db.CreateOrGetSession(ctx, captaindb.CreateSessionInput{ID: uuid.New(), Source: "gavel", Provider: "openai"})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		run, err := db.CreatePromptRun(ctx, captaindb.CreatePromptRunInput{SessionID: session.ID})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		proposed := map[string]any{"proposalId": "prepared-proposal", "targetId": "target"}
		request, err := db.CreateToolApprovalRequest(ctx, captaindb.CreateToolApprovalRequestInput{SessionID: session.ID, PromptRunID: run.ID, ToolCallID: uuid.NewString(), Tool: "triage_review", Input: proposed, RequestedBy: "gavel", ExpiresAt: time.Now().Add(time.Hour)})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		waiting, err := db.GetPromptRun(ctx, run.ID)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		stateWaiting := captaindb.PromptRunStateWaiting
		_, err = db.UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{ID: run.ID, ExpectedVersion: waiting.Version, State: &stateWaiting})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		switch state {
		case "approved", "denied":
			decision := captaindb.ResolveToolApprovalRequestInput{SessionID: session.ID, RequestID: request.ID, Approved: state == "approved", ResolvedBy: "operator"}
			if amended {
				decision.UpdatedInput = map[string]any{"targetId": "different"}
			}
			_, err = db.ResolveToolApprovalRequest(ctx, decision)
		case "expired":
			err = db.ExpireToolApprovalRequest(ctx, request.ID, captaindb.TurnRequestStateExpired, "expired")
		case "cancelled":
			err = db.CancelPendingTurnRequests(ctx, session.ID, run.ID, "cancelled")
		}
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		input := &stepInput{input: promptrun.Input{Record: &promptrun.Recording{DB: db, PromptRunID: run.ID}}}
		review := (&Host{}).newTriageReview(input, &types.TODO{ID: "source"})
		if accepted {
			gomega.Expect(review.Approved(ctx, proposed)).To(gomega.Succeed())
		} else {
			gomega.Expect(review.Approved(ctx, proposed)).To(gomega.HaveOccurred())
		}
		gomega.Expect(review.Approved(ctx, map[string]any{"targetId": "different"})).To(gomega.HaveOccurred())
		current, err := db.GetPromptRun(ctx, run.ID)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		finished := captaindb.PromptRunStateSucceeded
		_, err = db.UpdatePromptRun(ctx, captaindb.UpdatePromptRunInput{ID: run.ID, ExpectedVersion: current.Version, State: &finished})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		other, err := db.CreatePromptRun(ctx, captaindb.CreatePromptRunInput{SessionID: session.ID})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		input.input.Record.PromptRunID = other.ID
		gomega.Expect(review.Approved(ctx, proposed)).To(gomega.HaveOccurred())
	},
		ginkgo.Entry("pending", "pending", false, false),
		ginkgo.Entry("denied", "denied", false, false),
		ginkgo.Entry("expired", "expired", false, false),
		ginkgo.Entry("cancelled", "cancelled", false, false),
		ginkgo.Entry("approved exact input", "approved", false, true),
		ginkgo.Entry("approved amended input", "approved", true, false),
	)
})
