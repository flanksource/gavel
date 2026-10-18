package testui_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	testui "github.com/flanksource/gavel/testrunner/ui"
)

var _ = Describe("the test UI in dev mode", func() {
	const fromVite = "served by vite"
	var handler http.Handler

	BeforeEach(func() {
		server := testui.NewServer()
		server.SetDevProxy(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(fromVite))
		}))
		handler = server.Handler()
	})

	get := func(target string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, target, nil)
		// What a browser sends; without it "/" is a content-negotiated export.
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}

	DescribeTable("hands the page and its modules to Vite",
		func(target string) { Expect(get(target).Body.String()).To(Equal(fromVite)) },
		Entry("the page", "/"),
		Entry("the entry module", "/src/index.tsx"),
		Entry("vite's client", "/@vite/client"),
		Entry("a sibling clicky-ui source file", "/@fs/work/clicky-ui/packages/ui/src/data.ts"),
	)

	It("keeps the results API on the Go server", func() {
		response := get("/api/tests")

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body.String()).NotTo(Equal(fromVite))
		Expect(response.Header().Get("Content-Type")).To(ContainSubstring("application/json"))
	})

	It("answers an API path with no Go route as a JSON 404, not Vite's HTML fallback", func() {
		response := get("/api/not-a-route")

		Expect(response.Code).To(Equal(http.StatusNotFound))
		Expect(response.Body.String()).To(MatchJSON(`{"error":"no API route for GET /api/not-a-route"}`))
	})

	It("still renders an export on the Go server", func() {
		Expect(get("/?format=json").Body.String()).NotTo(Equal(fromVite))
	})
})
