package telltale_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/kinds/telltale"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTelltale(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Telltale fixture kind") }
func step(config string) *fixtures.StepContext {
	return &fixtures.StepContext{Spec: fixtures.StepSpec{ID: "diagnostic", Kind: "telltale", Config: json.RawMessage(config)}, Scope: fixtures.NewStepScope(os.Environ())}
}

var _ = Describe("Telltale operations", func() {
	It("preserves exact parameters and exposes the response for CEL", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			Expect(r.Method).To(Equal("POST"))
			Expect(r.URL.EscapedPath()).To(Equal("/api/agents/java%2Fexample/call"))
			Expect(r.Header.Get("Authorization")).To(Equal("Bearer fixture-token"))
			var body map[string]json.RawMessage
			Expect(json.NewDecoder(r.Body).Decode(&body)).To(Succeed())
			Expect(body["params"]).To(MatchJSON(`{"source":"return args.id","args":{"id":9007199254740993}}`))
			Expect(body["allowMutation"]).To(MatchJSON(`true`))
			_, _ = w.Write([]byte(`{"value":9007199254740993,"stdout":"hello"}`))
		}))
		DeferCleanup(server.Close)
		outcome, err := (telltale.Kind{}).Execute(context.Background(), step(`{"server":"`+server.URL+`","token":"fixture-token","agent":"java/example","operation":"jvm.groovy.eval","allowMutation":true,"params":{"source":"return args.id","args":{"id":9007199254740993}}}`))
		Expect(err).NotTo(HaveOccurred())
		Expect(outcome.Outputs["stdout"]).To(ContainSubstring(`9007199254740993`))
		Expect(outcome.Outputs["json"]).To(HaveKey("value"))
	})
	It("returns null and arrays without forcing them into objects", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`null`)) }))
		DeferCleanup(server.Close)
		outcome, err := (telltale.Kind{}).Execute(context.Background(), step(`{"server":"`+server.URL+`","agent":"demo","operation":"jvm.jmx.invoke"}`))
		Expect(err).NotTo(HaveOccurred())
		Expect(outcome.Outputs).To(HaveKeyWithValue("json", BeNil()))
	})
	DescribeTable("refuses invalid requests before network access", func(config string) {
		_, err := (telltale.Kind{}).Execute(context.Background(), step(config))
		Expect(err).To(HaveOccurred())
	}, Entry("no server", `{"agent":"demo","operation":"jvm.status"}`), Entry("no agent", `{"server":"http://127.0.0.1","operation":"jvm.status"}`), Entry("ambiguous action", `{"server":"http://127.0.0.1","agent":"demo","operation":"jvm.status","capture":"jvm_redefine","duration":"1s"}`), Entry("no duration", `{"server":"http://127.0.0.1","agent":"demo","capture":"jvm_redefine"}`))
	It("surfaces HTTP failures with operation context", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "mutation disabled", 400) }))
		DeferCleanup(server.Close)
		_, err := (telltale.Kind{}).Execute(context.Background(), step(`{"server":"`+server.URL+`","agent":"demo","operation":"jvm.groovy.eval"}`))
		Expect(err).To(MatchError(And(ContainSubstring("400"), ContainSubstring("mutation disabled"))))
	})
})
