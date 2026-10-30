package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/commons-db/observability"
	commonscontext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/har"
	commonshttp "github.com/flanksource/commons/http"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/commons/properties"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHTTPx(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP client observability")
}

func httpProperty(key, value string) {
	previous := properties.Get(key)
	properties.Set(key, value)
	DeferCleanup(func() {
		if previous == "" {
			properties.Unset(key)
		} else {
			properties.Set(key, previous)
		}
	})
}

var _ = Describe("commons HTTP client wiring", func() {
	var endpoint *httptest.Server
	BeforeEach(func() {
		previous := harRegistry
		harRegistry = har.NewRegistry(nil)
		DeferCleanup(func() { harRegistry = previous })
		endpoint = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"message":"response-content"}`)
		}))
		DeferCleanup(endpoint.Close)
	})

	It("taps requests for the shared Gavel HTTP observer", func() {
		collector := har.NewCollector(har.DefaultConfig())
		DeferCleanup(observability.ObserveHTTP(Feature, collector))
		client := commonshttp.NewClient()
		Expect(Configure(context.Background(), "jira", client)).To(Succeed())
		response, err := client.R(context.Background()).Get(endpoint.URL)
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Body.Close()).To(Succeed())
		Expect(collector.Entries()).To(HaveLen(1))
		Expect(collector.Entries()[0].Response.Content.Text).To(Equal(`{"message":"response-content"}`))
	})

	DescribeTable("inherits the request logger and shared trace configuration", func(level logger.LogLevel, key, value string, headers, body bool) {
		var output bytes.Buffer
		log := logger.NewWithWriter(&output)
		log.SetLogLevel(level)
		ctx := commonscontext.NewContext(context.Background(), commonscontext.WithLogger(log))
		if key != "" {
			httpProperty(key, value)
		}
		client := commonshttp.NewClient()
		Expect(Configure(ctx, "jira", client)).To(Succeed())
		response, err := client.R(ctx).Header("Authorization", "Bearer example-private-token").Header("X-Example", "visible-header").Get(endpoint.URL)
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Body.Close()).To(Succeed())
		Expect(output.String()).To(ContainSubstring("GET"))
		Expect(output.String()).NotTo(ContainSubstring("example-private-token"))
		Expect(bytes.Contains(output.Bytes(), []byte("visible-header"))).To(Equal(headers))
		Expect(bytes.Contains(output.Bytes(), []byte("response-content"))).To(Equal(body))
	},
		Entry("verbosity access line", logger.Debug, "", "", false, false),
		Entry("verbosity headers", logger.Trace, "", "", true, false),
		Entry("verbosity response body", logger.Trace2, "", "", true, true),
		Entry("global trace property", logger.Info, "http.log", "all,access", true, true),
		Entry("Jira trace property", logger.Info, "http.jira.log", "all,access", true, true),
	)

	It("does not install success logging at the default level", func() {
		var output bytes.Buffer
		log := logger.NewWithWriter(&output)
		log.SetLogLevel(logger.Info)
		ctx := commonscontext.NewContext(context.Background(), commonscontext.WithLogger(log))
		client := commonshttp.NewClient()
		Expect(Configure(ctx, "jira", client)).To(Succeed())
		response, err := client.R(ctx).Get(endpoint.URL)
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Body.Close()).To(Succeed())
		Expect(output.String()).To(BeEmpty())
	})

	It("captures HAR for both clients sharing the global output file", func() {
		path := filepath.Join(GinkgoT().TempDir(), "requests.har")
		httpProperty("http.har", path)
		for _, feature := range []string{"jira", "other"} {
			client := commonshttp.NewClient()
			Expect(Configure(context.Background(), feature, client)).To(Succeed())
			response, err := client.R(context.Background()).Header("Authorization", "Bearer example-private-token").Get(endpoint.URL)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Body.Close()).To(Succeed())
		}
		Expect(harRegistry.Flush()).To(Succeed())
		data, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		var archive har.File
		Expect(json.Unmarshal(data, &archive)).To(Succeed())
		Expect(archive.Log.Entries).To(HaveLen(2))
		Expect(archive.Log.Entries[0].Response.Content.Text).To(Equal(`{"message":"response-content"}`))
		Expect(string(data)).NotTo(ContainSubstring("example-private-token"))
	})

	It("uses Jira's HAR destination and metadata level before global properties", func() {
		dir := GinkgoT().TempDir()
		globalPath, jiraPath := filepath.Join(dir, "global.har"), filepath.Join(dir, "jira.har")
		httpProperty("http.har", globalPath)
		httpProperty("http.jira.har", jiraPath)
		httpProperty("http.jira.har.level", "metadata")
		for _, feature := range []string{"jira", "other"} {
			client := commonshttp.NewClient()
			Expect(Configure(context.Background(), feature, client)).To(Succeed())
			response, err := client.R(context.Background()).Get(endpoint.URL)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Body.Close()).To(Succeed())
		}
		Expect(harRegistry.Flush()).To(Succeed())
		for _, path := range []string{globalPath, jiraPath} {
			data, err := os.ReadFile(path)
			Expect(err).NotTo(HaveOccurred())
			var archive har.File
			Expect(json.Unmarshal(data, &archive)).To(Succeed())
			Expect(archive.Log.Entries).To(HaveLen(1))
			if path == jiraPath {
				Expect(archive.Log.Entries[0].Response.Content.Text).To(BeEmpty())
				Expect(archive.Log.Entries[0].Response.Content.Size).To(Equal(int64(-1)))
			} else {
				Expect(archive.Log.Entries[0].Response.Content.Text).To(Equal(`{"message":"response-content"}`))
			}
		}
	})

	It("honors disabled HAR capture", func() {
		path := filepath.Join(GinkgoT().TempDir(), "disabled.har")
		httpProperty("http.har", path)
		httpProperty("http.har.level", "off")
		client := commonshttp.NewClient()
		Expect(Configure(context.Background(), "jira", client)).To(Succeed())
		response, err := client.R(context.Background()).Get(endpoint.URL)
		Expect(err).NotTo(HaveOccurred())
		Expect(response.Body.Close()).To(Succeed())
		Expect(harRegistry.Flush()).To(Succeed())
		_, err = os.Stat(path)
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	It("rejects invalid HAR levels before configuring the client", func() {
		httpProperty("http.har", filepath.Join(GinkgoT().TempDir(), "invalid.har"))
		httpProperty("http.har.level", "unexpected")
		Expect(Configure(context.Background(), "jira", commonshttp.NewClient())).To(MatchError(ContainSubstring("http.har.level")))
	})
})
