package agents_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/clicky/route"
	"github.com/flanksource/gavel/comments"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Starting a named session for a page ahead of its threads", func() {
	const prefix = "gavel-localhost:5801"
	var f *fixture
	page := pageTarget("home")
	request := comments.StartSessionRequest{ScopeURI: scope, Target: page.URI, Prefix: prefix}
	BeforeEach(func(ctx SpecContext) { f = newFixture(ctx, "comment_start_session", pageRoutes{}) })

	It("answers the session its run reports, named <prefix>-<6 chars>, recorded as the page's managed session and resumed by a later delivery", func(ctx SpecContext) {
		started, err := f.svc.StartAgentSession(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		Expect(started.Title).To(MatchRegexp(`^` + prefix + `-[a-z0-9]{6}$`))
		Expect([]any{started.ID, started.Status, started.Managed}).To(Equal([]any{"managed-1", "", true}))

		_, launched := f.bridge.Snapshot()
		Expect(launched).To(HaveLen(1))
		var vars map[string]any
		Expect(json.Unmarshal([]byte(launched[0].Options.Vars), &vars)).To(Succeed())
		Expect([]any{launched[0].Name, launched[0].Reference, launched[0].Fresh}).To(Equal([]any{started.Title, "editor-comments", true}))
		Expect(launched[0].Tools).To(BeEmpty(), "a session without threads gets no thread tools")
		Expect([]any{vars["target"], vars["scope"], vars["dir"], vars["propose"], vars["waiting"]}).To(Equal([]any{page.URI, scope, f.dir, false, true}))
		Expect(vars["message"]).To(And(ContainSubstring(page.URI), ContainSubstring("Edits policy: direct"), ContainSubstring("later")))

		managed, err := f.store.ManagedSessions(ctx, "tenant-agents", scope)
		Expect(err).NotTo(HaveOccurred())
		Expect(managed).To(HaveLen(1))
		Expect([]any{managed[0].ID, managed[0].Label, managed[0].GroupKey, managed[0].Prompt, managed[0].CWD}).To(Equal([]any{"managed-1", started.Title, page.URI, "editor-comments", f.dir}))

		listed, err := f.svc.AgentSessions(ctx, comments.SessionQuery{ScopeURI: scope})
		Expect(err).NotTo(HaveOccurred())
		Expect(listed).To(ContainElement(started), "the live managed session is listed under its name")

		Expect(f.bridge.End(started.ID)).To(Succeed())
		thread := f.create(ctx, "home", "Fix the header")
		_, err = f.svc.Deliver(ctx, comments.DeliverRequest{CommentID: thread.ID, ExpectedVersion: thread.Version, RequestID: "to-started", Target: comments.DeliverTarget{Kind: "session", SessionID: started.ID}})
		Expect(err).NotTo(HaveOccurred())
		_, launched = f.bridge.Snapshot()
		Expect(launched).To(HaveLen(2))
		Expect([]any{launched[1].Options.Resume, len(launched[1].Tools) > 0}).To(Equal([]any{started.ID, true}), "the thread resumes the named session with its thread tools")
	})

	It("rejects a blank prefix without launching anything", func(ctx SpecContext) {
		blank := request
		blank.Prefix = "  "
		_, err := f.svc.StartAgentSession(ctx, blank)
		Expect(err).To(MatchError(comments.ErrInvalid))
		_, launched := f.bridge.Snapshot()
		Expect(launched).To(BeEmpty())
	})

	It("fails when the run ends before reporting a session id", func(ctx SpecContext) {
		f.bridge.DeferEvents = true
		failed := make(chan error, 1)
		go func() {
			defer GinkgoRecover()
			_, err := f.svc.StartAgentSession(ctx, request)
			failed <- err
		}()
		Eventually(func() int { _, launched := f.bridge.Snapshot(); return len(launched) }).Should(Equal(1))
		_, launched := f.bridge.Snapshot()
		launched[0].OnComplete(captaincli.PromptRunSummary{Error: "session admission refused"})
		var err error
		Eventually(failed).Should(Receive(&err))
		Expect(err).To(And(MatchError(comments.ErrDelivery), MatchError(ContainSubstring("session admission refused"))))
		managed, err := f.store.ManagedSessions(ctx, "tenant-agents", scope)
		Expect(err).NotTo(HaveOccurred())
		Expect(managed).To(BeEmpty())
	})

	It("answers 201 with the session over HTTP, and 400 for a blank prefix", func() {
		router := route.NewRouter(nil)
		commentapi.RegisterDeliveryRoutes(router, func(context.Context) (*comments.Service, error) { return f.svc, nil })
		post := func(body map[string]any) *httptest.ResponseRecorder {
			data, err := json.Marshal(body)
			Expect(err).NotTo(HaveOccurred())
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/comments/agent-sessions", bytes.NewReader(data)))
			return response
		}
		response := post(map[string]any{"scope": scope, "target": page.URI, "prefix": prefix, "edits": "propose"})
		Expect(response.Code).To(Equal(http.StatusCreated), response.Body.String())
		var session comments.AgentSession
		Expect(json.Unmarshal(response.Body.Bytes(), &session)).To(Succeed())
		Expect([]any{session.ID, session.Managed}).To(Equal([]any{"managed-1", true}))
		Expect(session.Title).To(MatchRegexp(`^` + prefix + `-[a-z0-9]{6}$`))
		Expect(post(map[string]any{"scope": scope, "target": page.URI, "prefix": ""}).Code).To(Equal(http.StatusBadRequest))
		Expect(post(map[string]any{"scope": scope, "prefix": prefix}).Code).To(Equal(http.StatusBadRequest))
	})
})
