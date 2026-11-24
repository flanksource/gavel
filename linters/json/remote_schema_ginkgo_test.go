package jsonlint

import (
	"context"
	"github.com/flanksource/gavel/linters/json/jsonlinttest"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Remote schema destinations", func() {
	DescribeTable("refuses non-public destinations even with remote schemas enabled", func(location string) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		DeferCleanup(cancel)
		_, err := (schemaLoader{ctx: ctx, remoteSchemas: true}).Load(location)
		Expect(err).To(MatchError(ContainSubstring("defensive HTTP client refuses")))
	},
		Entry("loopback IPv4", "http://127.0.0.1/schema.json"),
		Entry("loopback IPv6", "http://[::1]/schema.json"),
		Entry("mapped loopback IPv4", "http://[::ffff:127.0.0.1]/schema.json"),
		Entry("loopback DNS", "http://localhost/schema.json"),
		Entry("private IPv4", "http://10.0.0.1/schema.json"),
		Entry("private IPv6", "http://[fd00::1]/schema.json"),
		Entry("metadata address", "http://169.254.169.254/schema.json"),
		Entry("metadata hostname", "http://metadata.google.internal/schema.json"),
		Entry("unspecified", "http://0.0.0.0/schema.json"),
		Entry("carrier-grade NAT", "http://100.64.0.1/schema.json"),
		Entry("multicast", "http://224.0.0.1/schema.json"),
		Entry("credentials", "https://user:password@example.com/schema.json"),
	)

	It("refuses loopback requests without reaching the server", func() {
		var requests atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requests.Add(1)
			_, err := w.Write([]byte(`true`))
			Expect(err).NotTo(HaveOccurred())
		}))
		DeferCleanup(server.Close)
		_, err := (schemaLoader{ctx: context.Background(), remoteSchemas: true}).Load(server.URL)
		Expect(err).To(HaveOccurred())
		Expect(requests.Load()).To(BeZero())
	})

	It("follows redirects between public destinations", func() {
		var requests atomic.Int32
		server := jsonlinttest.ServePublicSchema(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			if r.URL.Path == "/redirect" {
				http.Redirect(w, r, "/schema.json", http.StatusFound)
				return
			}
			_, err := w.Write([]byte(`true`))
			Expect(err).NotTo(HaveOccurred())
		}))
		DeferCleanup(server.Close)
		value, err := (schemaLoader{ctx: context.Background(), remoteSchemas: true}).Load(server.URL + "/redirect")
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(BeTrue())
		Expect(requests.Load()).To(Equal(int32(2)))
	})

	DescribeTable("refuses private destinations reached through a public schema", func(redirect bool) {
		var requests atomic.Int32
		private := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requests.Add(1)
			_, err := w.Write([]byte(`true`))
			Expect(err).NotTo(HaveOccurred())
		}))
		DeferCleanup(private.Close)
		server := jsonlinttest.ServePublicSchema(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if redirect {
				http.Redirect(w, r, private.URL, http.StatusFound)
				return
			}
			_, err := w.Write([]byte(`{"$ref":"` + private.URL + `/schema.json"}`))
			Expect(err).NotTo(HaveOccurred())
		}))
		DeferCleanup(server.Close)
		_, err := newValidator(schemaLoader{ctx: context.Background(), remoteSchemas: true}).compile(server.URL+"/schema.json", ".")
		Expect(err).To(MatchError(ContainSubstring("defensive HTTP client refuses")))
		Expect(requests.Load()).To(BeZero())
	}, Entry("redirect", true), Entry("reference", false))

	It("disables proxies that could bypass address validation", func() {
		server := jsonlinttest.ServePublicSchema(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, err := w.Write([]byte(`true`))
			Expect(err).NotTo(HaveOccurred())
		}))
		DeferCleanup(server.Close)
		var proxies atomic.Int32
		http.DefaultTransport.(*http.Transport).Proxy = func(*http.Request) (*url.URL, error) {
			proxies.Add(1)
			return url.Parse("http://127.0.0.1:1")
		}
		value, err := (schemaLoader{ctx: context.Background(), remoteSchemas: true}).Load(server.URL)
		Expect(err).NotTo(HaveOccurred())
		Expect(value).To(BeTrue())
		Expect(proxies.Load()).To(BeZero())
	})
})
