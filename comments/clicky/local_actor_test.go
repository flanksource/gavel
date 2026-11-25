package clicky_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flanksource/gavel/comments"
	api "github.com/flanksource/gavel/comments/clicky"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCommentsClicky(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Comments clicky") }

var _ = Describe("LocalActor", func() {
	you := comments.Actor{Kind: "user", ID: "local:501", Label: "You"}
	agent := comments.Actor{Kind: "agent", ID: "claude:session-1", Label: "claude-opus"}
	serve := func(remote, header string) (int, comments.Actor) {
		var seen comments.Actor
		handler := api.LocalActor(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor, ok := api.ActorFrom(r.Context())
			Expect(ok).To(BeTrue())
			seen = actor
		}), you)
		request := httptest.NewRequest("GET", "/api/v1/comments", nil)
		request.RemoteAddr = remote
		if header != "" {
			request.Header.Set(api.ActorHeader, header)
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response.Code, seen
	}

	It("acts as the default actor for a loopback caller without an assertion", func() {
		code, actor := serve("127.0.0.1:5000", "")
		Expect([]any{code, actor}).To(Equal([]any{http.StatusOK, you}))
	})
	It("acts as the asserted agent for a loopback caller, round-tripping EncodeActor", func() {
		code, actor := serve("[::1]:5000", api.EncodeActor(agent))
		Expect([]any{code, actor}).To(Equal([]any{http.StatusOK, agent}))
	})
	DescribeTable("refuses",
		func(remote, header string, status int) {
			code, _ := serve(remote, header)
			Expect(code).To(Equal(status))
		},
		Entry("a caller that is not on loopback", "10.0.0.5:5000", "", http.StatusForbidden),
		Entry("an assertion that is not base64url JSON", "127.0.0.1:5000", "not-json", http.StatusBadRequest),
		Entry("an assertion of an unknown actor kind", "127.0.0.1:5000", api.EncodeActor(comments.Actor{Kind: "system", ID: "x"}), http.StatusBadRequest),
		Entry("an assertion without an id", "127.0.0.1:5000", api.EncodeActor(comments.Actor{Kind: "agent"}), http.StatusBadRequest),
	)
	It("carries an actor set in-process with WithActor", func() {
		actor, ok := api.ActorFrom(api.WithActor(GinkgoT().Context(), agent))
		Expect([]any{actor, ok}).To(Equal([]any{agent, true}))
	})
})
