package testui_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	testui "github.com/flanksource/gavel/testrunner/ui"
)

var _ = Describe("the test UI page", func() {
	get := func(server *testui.Server, path string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Set("Accept", "text/html")
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, request)
		return response
	}

	It("loads its bundle as a module from under the server's base path", func() {
		server := testui.NewServer()
		server.SetBasePath("/runs/run-1")

		page := get(server, "/").Body.String()

		Expect(page).To(ContainSubstring(`window.__gavelBasePath = "/runs/run-1";`))
		Expect(page).To(MatchRegexp(`<script (type="module" src="/runs/run-1/_testui/testui\.js"|src="/runs/run-1/_testui/testui\.js" type="module")></script>`))
	})

	It("serves the entry and the chunks it imports", func() {
		server := testui.NewServer()
		entry := get(server, testui.BundlePath)
		Expect(entry.Code).To(Equal(http.StatusOK))
		Expect(entry.Header().Get("Content-Type")).To(ContainSubstring("javascript"))

		chunk := regexp.MustCompile(`"\./(chunks/[^"]+\.js)"`).FindStringSubmatch(entry.Body.String())
		Expect(chunk).NotTo(BeNil(), "the entry imports at least one chunk")
		response := get(server, testui.AssetsMount+"/"+chunk[1])
		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Header().Get("Cache-Control")).To(ContainSubstring("immutable"))
	})

	DescribeTable("serves nothing else from the embedded dist",
		func(path string) { Expect(get(testui.NewServer(), path).Code).To(Equal(http.StatusNotFound)) },
		Entry("the library build", testui.AssetsMount+"/lib/index.js"),
		Entry("the chunks directory listing", testui.AssetsMount+"/chunks/"),
	)
})
