package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	httpclient "github.com/flanksource/commons/http"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) {
	if filepath.Base(os.Args[0]) == "twg" {
		data, err := json.Marshal(os.Args[1:])
		if err != nil || os.WriteFile(os.Getenv("JIRA_TEST_TWG_ARGS"), data, 0o600) != nil {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stdout, "export TWG_SITE=$'example.atlassian.net'\nexport TWG_OAUTH_ACCESS_TOKEN=$'exported-token'")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

var _ = Describe("Jira authentication boundaries", func() {
	It("rejects invalid bearer tokens before an HTTP error can expose them", func() {
		_, err := NewClient(context.Background(), ClientOptions{Site: "example", Credentials: Credentials{AccessToken: "private-token\ninvalid"}})
		Expect(err).To(MatchError("jira OAuth access token contains invalid characters"))
	})

	It("discovers the selected TWG profile through its refresh/export command without a shell", func() {
		dir := GinkgoT().TempDir()
		binary, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Symlink(binary, filepath.Join(dir, "twg"))).To(Succeed())
		argsFile := filepath.Join(dir, "args.json")
		GinkgoT().Setenv("JIRA_TEST_TWG_ARGS", argsFile)
		GinkgoT().Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
		GinkgoT().Setenv("JIRA_CLOUD_ID", "selected-cloud")
		credentials, site, err := discoverCredentials(context.Background(), ClientOptions{Site: "example", Profile: "work"})
		Expect(err).NotTo(HaveOccurred())
		Expect(site).To(Equal("example"))
		Expect(credentials).To(Equal(Credentials{AccessToken: "exported-token", CloudID: "selected-cloud"}))
		data, err := os.ReadFile(argsFile)
		Expect(err).NotTo(HaveOccurred())
		var args []string
		Expect(json.Unmarshal(data, &args)).To(Succeed())
		Expect(args).To(Equal([]string{"env", "--site", "example", "--profile", "work", "--exec", "--", filepath.Join(dir, "twg"), "env", "--export"}))
	})

	It("selects an OAuth site and sends bearer requests through the cloud gateway", func() {
		var paths []string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			Expect(r.Header.Get("Authorization")).To(Equal("Bearer oauth-test-token"))
			paths = append(paths, r.URL.Path)
			switch r.URL.Path {
			case "/oauth/token/accessible-resources":
				fmt.Fprint(w, `[{"id":"other","url":"https://other.atlassian.net"},{"id":"cloud-id","url":"https://example.atlassian.net"}]`)
			case "/ex/jira/cloud-id/rest/api/3/field":
				fmt.Fprint(w, `[]`)
			case "/ex/jira/cloud-id/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			default:
				http.NotFound(w, r)
			}
		}))
		DeferCleanup(server.Close)
		endpoint, err := url.Parse(server.URL)
		Expect(err).NotTo(HaveOccurred())
		httpClient := httpclient.NewClient().InsecureSkipVerify(false).Use(func(next http.RoundTripper) http.RoundTripper {
			return transportFunc(func(r *http.Request) (*http.Response, error) {
				copy := r.Clone(r.Context())
				copy.URL.Scheme = endpoint.Scheme
				copy.URL.Host = endpoint.Host
				return next.RoundTrip(copy)
			})
		})
		client, err := NewClient(context.Background(), ClientOptions{Site: "example", Credentials: Credentials{AccessToken: "oauth-test-token"}, HTTP: httpClient})
		Expect(err).NotTo(HaveOccurred())
		_, err = client.Catalog(context.Background(), CatalogOptions{Store: CatalogStore{Dir: GinkgoT().TempDir()}})
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(Equal([]string{"/oauth/token/accessible-resources", "/ex/jira/cloud-id/rest/api/3/myself", "/ex/jira/cloud-id/rest/api/3/field"}))
	})

	It("does not follow redirects with Jira credentials", func() {
		called := false
		other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))
		DeferCleanup(other.Close)
		client := testClient(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, other.URL, http.StatusFound) })
		_, err := client.Comments(context.Background(), "EXAMPLE-1")
		Expect(err).To(MatchError(ContainSubstring("HTTP 302")))
		Expect(called).To(BeFalse())
	})

	It("verifies TLS certificates and rejects credential-bearing site URLs", func() {
		_, err := NewClient(context.Background(), ClientOptions{Site: "https://user:password@example.test", Credentials: Credentials{Username: "api", APIToken: "test-token"}})
		Expect(err).To(MatchError("invalid Jira site URL"))
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `[]`) }))
		DeferCleanup(server.Close)
		client, err := NewClient(context.Background(), ClientOptions{Site: server.URL, Credentials: Credentials{Username: "api", APIToken: "test-token"}})
		Expect(err).NotTo(HaveOccurred())
		_, err = client.Comments(context.Background(), "EXAMPLE-1")
		Expect(err).To(HaveOccurred())
		Expect(strings.ToLower(err.Error())).To(ContainSubstring("certificate"))
	})
})
