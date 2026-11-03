package queryapp_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/gavel/testrunner/ui/queryapp"
)

func TestQueryApp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Query app over run trace stores")
}

const runID = "run-1"

func get(handler http.Handler, target string) *httptest.ResponseRecorder {
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))
	return response
}

var _ = Describe("the query app over a run's trace store", func() {
	It("opens a window's link on the app's page under the run's mount, and leaves /api/ to the store", func() {
		srv := testui.NewServer()
		store := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("store " + r.URL.Path)) })
		srv.ServeTraces(runID, store, nil)
		Expect(queryapp.Serve(srv)).To(Succeed())
		mount := testui.QueryMount + "/" + runID

		link := get(srv.Handler(), mount+"/traces/http_proxy?stream=s1")
		Expect(link.Code).To(Equal(http.StatusFound))
		Expect(link.Header().Get("Location")).To(Equal("../profile-traces-http-proxy?stream=s1"))
		Expect(get(srv.Handler(), mount+"/profile-traces-http-proxy").Body.String()).To(ContainSubstring(`<base href="` + mount + `/"`))
		Expect(get(srv.Handler(), mount+"/api/v1/traces/kinds").Body.String()).To(Equal("store /api/v1/traces/kinds"))
		Expect(get(srv.Handler(), mount+"/api").Body.String()).To(Equal("store /api"))
	})
})
