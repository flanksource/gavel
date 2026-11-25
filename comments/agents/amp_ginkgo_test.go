package agents_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/flanksource/captain/pkg/session/amp"
	"github.com/flanksource/clicky/route"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/agents"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const ampThreadID = "T-7f395a45-7fae-4983-8de0-d02e61d30183"

type ampDelivery struct {
	mu       sync.Mutex
	messages []string
	failure  error
}

func (*ampDelivery) OwnedOrb(_ context.Context, id string) (amp.Thread, error) {
	return amp.Thread{ID: id, Title: "Amp review"}, nil
}

func (*ampDelivery) List(context.Context) ([]amp.Thread, error) {
	repos := []amp.Repository{{URL: "https://github.com/Example/project.git"}}
	return []amp.Thread{
		{ID: ampThreadID, Title: "Amp review", Repositories: repos},
		{ID: ampThreadID, Title: "A duplicate", Repositories: repos},
		{ID: "T-local", Hostname: "workstation", Repositories: repos},
		{ID: "T-archived", Archived: true, Repositories: repos},
		{ID: "T-other", Repositories: []amp.Repository{{URL: "https://github.com/example/other"}}},
	}, nil
}

func (a *ampDelivery) Send(_ context.Context, id, message string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.messages = append(a.messages, id+"\n"+message)
	if a.failure != nil {
		return "", a.failure
	}
	return "M-receipt", nil
}

func (a *ampDelivery) sent() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]string(nil), a.messages...)
}

var _ = Describe("Amp comment delivery", Ordered, func() {
	var f *fixture
	var client *ampDelivery
	target := comments.DeliverTarget{Kind: "session", SessionID: amp.ThreadURLPrefix + ampThreadID}

	BeforeAll(func(ctx SpecContext) {
		f = newFixture(ctx, "comment_agent_amp", nil)
		client = &ampDelivery{}
		delivery, err := agents.Delivery(agents.Options{
			Bridge: f.bridge, Groups: f.store, Amp: client,
			Workdir:      func(context.Context, comments.Target) (string, error) { return f.dir, nil },
			Repositories: func(context.Context, string) ([]string, error) { return []string{"example/project"}, nil },
		})
		Expect(err).NotTo(HaveOccurred())
		f.svc = f.service(comments.Options{Delivery: delivery})
	})

	It("offers repository-matched Orb threads alongside Captain sessions", func(ctx SpecContext) {
		sessions, err := f.svc.AgentSessions(ctx, comments.SessionQuery{ScopeURI: scope})
		Expect(err).NotTo(HaveOccurred())
		Expect(sessions).To(HaveLen(3))
		Expect(sessions[2]).To(Equal(comments.AgentSession{ID: ampThreadID, Title: "Amp review", Source: "amp", Status: "available"}))
	})

	It("previews and posts to a pasted Amp thread instead of the draft mention", func(ctx SpecContext) {
		draft := comments.CreateDeliveryDraft{Target: pageTarget("amp-preview"), Body: "@rename-pass check this page", Deliver: &target}
		data, err := json.Marshal(draft)
		Expect(err).NotTo(HaveOccurred())
		router := route.NewRouter(nil)
		commentapi.RegisterDeliveryRoutes(router, func(context.Context) (*comments.Service, error) { return f.svc, nil })
		response := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/comments/delivery/preview", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(response, req)
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		var preview comments.CreateDeliveryPreview
		Expect(json.Unmarshal(response.Body.Bytes(), &preview)).To(Succeed())
		Expect(preview.Session.ID).To(Equal(ampThreadID))
		Expect(preview.Route.Target.SessionID).To(Equal(ampThreadID))
		request := comments.CreateRequest{Target: draft.Target, Body: draft.Body, Deliver: &target, WorkflowID: "human-review", RequestID: uuid.NewString()}
		thread, err := f.svc.Create(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Eventually(f.runOf(ctx, thread.ID)).WithTimeout(5 * time.Second).Should(HaveField("MessageID", "M-receipt"))
		Expect(f.get(ctx, thread.ID).Status).To(Equal(thread.Status))
		request.Deliver = &preview.Route.Target
		request.Deliver.Edits = ""
		replayed, err := f.svc.Create(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(replayed.ID).To(Equal(thread.ID))
		Expect(client.sent()).To(HaveLen(1))
		live, _ := f.bridge.Snapshot()
		Expect(live).To(BeEmpty())
	})

	It("sends a batch once with a shared receipt and preserves the comment status", func(ctx SpecContext) {
		first := f.create(ctx, "amp-batch", "Align the heading")
		second := f.create(ctx, "amp-batch", "Update the footer")
		before := len(client.sent())
		request := comments.DeliverBatchRequest{RequestID: uuid.NewString(), Target: target, Threads: []comments.DeliverThread{
			{ID: first.ID, ExpectedVersion: first.Version}, {ID: second.ID, ExpectedVersion: second.Version},
		}}
		rows, err := f.svc.DeliverBatch(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(rows).To(HaveLen(2))
		Expect(contentOf(rows[0]).AgentRun).To(Equal(contentOf(rows[1]).AgentRun))
		Expect(contentOf(rows[0]).AgentRun.Target.Channel).To(Equal("amp-http"))
		Expect(rows[0].Status).To(Equal(first.Status))
		Expect(rows[1].Status).To(Equal(second.Status))
		events, err := f.svc.Events(ctx, scope, 0, 100)
		Expect(err).NotTo(HaveOccurred())
		var accepted []comments.Operation
		for _, event := range events {
			if event.CommentID == first.ID && event.Kind == "delivery:done" {
				var operation comments.Operation
				Expect(json.Unmarshal(event.Data, &operation)).To(Succeed())
				accepted = append(accepted, operation)
			}
		}
		Expect(accepted).To(HaveLen(1))
		receipt, err := json.Marshal(contentOf(rows[0]).AgentRun)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(accepted[0].Receipt)).To(MatchJSON(string(receipt)))
		_, _ = f.svc.DeliverBatch(ctx, request)
		request.RequestID = uuid.NewString()
		_, err = f.svc.DeliverBatch(ctx, request)
		Expect(err).To(MatchError(comments.ErrConflict))
		messages := client.sent()
		Expect(messages).To(HaveLen(before + 1))
		Expect(messages[before]).To(MatchRegexp(`(?s)Align the heading.*Update the footer`))
		Expect(messages[before]).NotTo(ContainSubstring("GAVEL_COMMENT_ACTOR"))
	})

	It("records an unconfirmed handoff and does not resend the same operation", func(ctx SpecContext) {
		client.failure = errors.New("Amp delivery unconfirmed")
		DeferCleanup(func() { client.failure = nil })
		thread := f.create(ctx, "amp-uncertain", "Review the spacing")
		before := len(client.sent())
		request := comments.DeliverRequest{CommentID: thread.ID, ExpectedVersion: thread.Version, RequestID: uuid.NewString(), Target: target}
		failed, err := f.svc.Deliver(ctx, request)
		Expect(err).To(MatchError(comments.ErrDelivery))
		run := contentOf(failed).AgentRun
		Expect([]string{run.State, run.SessionID, run.MessageID}).To(Equal([]string{"error", ampThreadID, ""}))
		_, err = f.svc.Deliver(ctx, request)
		Expect(err).To(MatchError(comments.ErrDelivery))
		Expect(client.sent()).To(HaveLen(before + 1))
	})
})
