package agents_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/flanksource/clicky/route"
	"github.com/flanksource/gavel/comments"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Looking up a comment group's selected session", Ordered, func() {
	var f *fixture
	BeforeAll(func(ctx SpecContext) {
		f = newFixture(ctx, "comment_group_selection", pageRoutes{})
		for i, entry := range []struct{ id, group string }{{"older", "shared"}, {"current", "shared"}, {"other", "another"}} {
			Expect(f.store.RecordSession(ctx, "tenant-agents", comments.ManagedSession{
				ID: entry.id, ScopeURI: scope, GroupKey: entry.group, Prompt: "editor-comments",
				Label: entry.id, CWD: f.dir, StartedAt: time.Date(2026, 1, 1, 0, 0, i, 0, time.UTC),
			})).To(Succeed())
		}
	})
	It("returns only the current group session, including an ended session that can resume", func(ctx SpecContext) {
		sessions, err := f.svc.AgentSessions(ctx, comments.SessionQuery{ScopeURI: scope, GroupKey: "shared"})
		Expect(err).NotTo(HaveOccurred())
		Expect(sessions).To(Equal([]comments.AgentSession{{ID: "current", Title: "current", CWD: f.dir, Status: "ended", Managed: true, Resumable: true}}))
	})
	It("returns no sessions for a group that has not started", func(ctx SpecContext) {
		sessions, err := f.svc.AgentSessions(ctx, comments.SessionQuery{ScopeURI: scope, GroupKey: "unknown"})
		Expect(err).NotTo(HaveOccurred())
		Expect(sessions).To(BeEmpty())
	})
	It("requires a scope when looking up a group", func(ctx SpecContext) {
		_, err := f.svc.AgentSessions(ctx, comments.SessionQuery{GroupKey: "shared", All: true})
		Expect(err).To(MatchError(comments.ErrInvalid))
	})
	It("carries the group filter through HTTP and rejects a group without scope", func() {
		router := route.NewRouter(nil)
		commentapi.RegisterDeliveryRoutes(router, func(context.Context) (*comments.Service, error) { return f.svc, nil })
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/comments/agent-sessions?scope="+url.QueryEscape(scope)+"&group=shared", nil))
		Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		var sessions []comments.AgentSession
		Expect(json.Unmarshal(response.Body.Bytes(), &sessions)).To(Succeed())
		Expect(sessions).To(HaveLen(1))
		Expect(sessions[0].ID).To(Equal("current"))
		response = httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/comments/agent-sessions?group=shared", nil))
		Expect(response.Code).To(Equal(http.StatusBadRequest))
	})
})
