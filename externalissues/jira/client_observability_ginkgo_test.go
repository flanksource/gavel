package jira

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/flanksource/commons-db/observability"
	"github.com/flanksource/commons/har"
	httpclient "github.com/flanksource/commons/http"
	"github.com/flanksource/gavel/internal/httpx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira HTTP observability", func() {
	It("captures default-client requests through Gavel's shared observer", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"comments":[],"total":0}`)
		}))
		DeferCleanup(server.Close)
		collector := har.NewCollector(har.DefaultConfig())
		DeferCleanup(observability.ObserveHTTP(httpx.Feature, collector))
		client, err := NewClient(context.Background(), ClientOptions{Site: server.URL, Credentials: Credentials{Username: "api", APIToken: "example-api-token"}})
		Expect(err).NotTo(HaveOccurred())
		_, err = client.Comments(context.Background(), "EXAMPLE-1")
		Expect(err).NotTo(HaveOccurred())
		Expect(collector.Entries()).To(HaveLen(1))
		Expect(collector.Entries()[0].Request.URL).To(ContainSubstring("/rest/api/3/issue/EXAMPLE-1/comment"))
		Expect(collector.Entries()[0].Response.Content.Text).To(Equal(`{"comments":[],"total":0}`))
	})

	It("wires injected clients before OAuth discovery and captures failed requests", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.URL.Path == "/oauth/token/accessible-resources" {
				_, _ = io.WriteString(w, `[{"id":"cloud","url":"https://example.atlassian.net"}]`)
				return
			}
			w.WriteHeader(http.StatusForbidden)
			_, _ = io.WriteString(w, `{"message":"permission denied"}`)
		}))
		DeferCleanup(server.Close)
		endpoint, err := url.Parse(server.URL)
		Expect(err).NotTo(HaveOccurred())
		httpClient := httpclient.NewClient().Use(func(next http.RoundTripper) http.RoundTripper {
			return transportFunc(func(r *http.Request) (*http.Response, error) {
				request := r.Clone(r.Context())
				request.URL.Scheme, request.URL.Host = endpoint.Scheme, endpoint.Host
				return next.RoundTrip(request)
			})
		})
		collector := har.NewCollector(har.DefaultConfig())
		DeferCleanup(observability.ObserveHTTP(httpx.Feature, collector))
		client, err := NewClient(context.Background(), ClientOptions{Site: "example", Credentials: Credentials{AccessToken: "example-oauth-token"}, HTTP: httpClient})
		Expect(err).NotTo(HaveOccurred())
		_, err = client.Comments(context.Background(), "EXAMPLE-1")
		Expect(err).To(MatchError(ContainSubstring("HTTP 403")))
		Expect(collector.Entries()).To(HaveLen(2))
		Expect(collector.Entries()[0].Request.URL).To(ContainSubstring("/oauth/token/accessible-resources"))
		Expect(collector.Entries()[1].Response.Status).To(Equal(http.StatusForbidden))
	})
})
