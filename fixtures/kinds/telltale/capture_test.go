package telltale_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/flanksource/gavel/fixtures/kinds/telltale"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("owned Telltale captures", func() {
	It("starts redefinition with exact target and consent, observes it and stops it at teardown", func() {
		stopped := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			switch r.Method {
			case "POST":
				var body map[string]json.RawMessage
				Expect(json.NewDecoder(r.Body).Decode(&body)).To(Succeed())
				Expect(body["traces"]).To(MatchJSON(`{"jvm_redefine":{"enabled":true,"target":{"seam":"agent","agent":"demo","allowMutation":true},"class":"example.Worker","source":"replacement"}}`))
				_, _ = w.Write([]byte(`{"runId":"owned-run","status":"running"}`))
			case "GET":
				status := "running"
				if stopped {
					status = "succeeded"
				}
				_, _ = w.Write([]byte(`{"runs":[{"runId":"owned-run","status":"` + status + `"}]}`))
			case "DELETE":
				Expect(r.URL.Path).To(Equal("/api/traces/owned-run"))
				stopped = true
				w.WriteHeader(202)
			}
		}))
		DeferCleanup(server.Close)
		cfg := `{"server":"` + server.URL + `","agent":"demo","capture":"jvm_redefine","duration":"1s","allowMutation":true,"params":{"class":"example.Worker","source":"replacement"}}`
		ctx := context.Background()
		s := step(cfg)
		result, err := (telltale.Kind{}).Execute(ctx, s)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Outputs).To(HaveKeyWithValue("runId", "owned-run"))
		observed, err := (telltale.Kind{}).Observe(ctx, s, result.Handle)
		Expect(err).NotTo(HaveOccurred())
		Expect(observed["status"]).To(Equal("running"))
		Expect(observed["exitCode"]).To(Equal(0))
		Expect(result.Cleanup(ctx)).To(Succeed())
		Expect(stopped).To(BeTrue())
		Expect(result.Cleanup(ctx)).To(Succeed())
	})
	It("cancels an HTTP operation when the step context is cancelled", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
		DeferCleanup(server.Close)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := (telltale.Kind{}).Execute(ctx, step(`{"server":"`+server.URL+`","agent":"demo","operation":"jvm.groovy.eval"}`))
		Expect(err).To(HaveOccurred())
	})
})
