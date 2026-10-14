package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("explicit TODO PR links over HTTP", func() {
	It("creates, replaces and clears an explicit link", func() {
		dir := GinkgoT().TempDir()
		server := &Server{ghOpts: github.Options{WorkDir: dir}}
		pr := &types.PR{Number: 27, URL: "https://github.com/acme/service/pull/27"}
		raw, err := json.Marshal(todoNewPayload{TodoCreatePayload: TodoCreatePayload{Title: "Fix parser", PR: pr}})
		Expect(err).NotTo(HaveOccurred())
		req := httptest.NewRequest(http.MethodPost, "/api/todos/new?dir="+url.QueryEscape(dir), bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusCreated), rec.Body.String())
		var created todoNewResponse
		Expect(json.Unmarshal(rec.Body.Bytes(), &created)).To(Succeed())
		Expect(created.Todo.PR).To(Equal(pr))
		for _, payload := range []string{`{"pr":{"number":28,"url":"https://dev.azure.com/acme/product/_git/service/pullrequest/28"}}`, `{"clearPR":true}`} {
			req = httptest.NewRequest(http.MethodPatch, "/api/todos/item?dir="+url.QueryEscape(dir)+"&ref="+url.QueryEscape(created.Todo.Ref), strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			rec = httptest.NewRecorder()
			server.Handler().ServeHTTP(rec, req)
			Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
			var updated todoSummary
			Expect(json.Unmarshal(rec.Body.Bytes(), &updated)).To(Succeed())
			if payload != `{"clearPR":true}` {
				Expect(updated.PR).To(Equal(&types.PR{Number: 28, URL: "https://dev.azure.com/acme/product/_git/service/pullrequest/28"}))
			}
		}
		var updated todoSummary
		Expect(json.Unmarshal(rec.Body.Bytes(), &updated)).To(Succeed())
		Expect(updated.PR).To(BeNil())
	})
	DescribeTable("validates PR links at create boundaries", func(route string) {
		dir := GinkgoT().TempDir()
		server := &Server{ghOpts: github.Options{WorkDir: dir}}
		req := httptest.NewRequest(http.MethodPost, route+"?dir="+url.QueryEscape(dir), strings.NewReader(`{"title":"Fix parser","pr":{"number":28,"url":"https://github.com/acme/service/pull/27"}}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusBadRequest), rec.Body.String())
	}, Entry("new", "/api/todos/new"), Entry("create", "/api/todos"))
	It("parses explicit links and unlink operations in form PATCH requests", func() {
		prURL := "https://github.com/acme/service/pull/27"
		for _, value := range []url.Values{{"pr": {prURL}}, {"clearPR": {"true"}}} {
			req := httptest.NewRequest(http.MethodPatch, "/api/todos/item", strings.NewReader(value.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			payload, _, err := parseTodoUpdatePayload(req)
			Expect(err).NotTo(HaveOccurred())
			if value.Has("pr") {
				Expect(payload.PR).To(Equal(&types.PR{Number: 27, URL: prURL}))
			} else {
				Expect(payload.ClearPR).To(BeTrue())
			}
		}
	})
})
