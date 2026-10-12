package ui

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The chat window creates its thread with POST /api/chat/sessions before the
// first message. When the dashboard never mounted the chat service that request
// fell through to the mux's plain 404, and the window reported "Creating a chat
// thread failed with status 404".
var _ = Describe("dashboard chat routes", func() {
	BeforeEach(func() {
		GinkgoT().Setenv(database.EnvDisable, "off")
		Expect(database.ResetDisabledSharedForTest()).To(Succeed())
		DeferCleanup(database.ResetDisabledSharedForTest)
	})

	DescribeTable("reach the chat service rather than the mux's not-found",
		func(method, path string) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(method, path, strings.NewReader("{}"))
			request.Header.Set("Content-Type", "application/json")
			(&Server{}).Handler().ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusServiceUnavailable), recorder.Body.String())
			Expect(recorder.Body.String()).To(ContainSubstring("Gavel chat requires the TODO database"))
		},
		Entry("create a thread", http.MethodPost, "/api/chat/sessions"),
		Entry("list threads", http.MethodGet, "/api/chat/sessions"),
		Entry("list models", http.MethodGet, "/api/chat/models"),
		Entry("send a message", http.MethodPost, "/api/chat"),
	)
})
