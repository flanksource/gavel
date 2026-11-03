package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJira(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Jira Suite")
}

func testClient(handler http.HandlerFunc) *Client {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer GinkgoRecover()
		user, password, ok := r.BasicAuth()
		Expect(ok).To(BeTrue())
		Expect(user).To(Equal("api@example.test"))
		Expect(password).To(Equal("test-api-token"))
		w.Header().Set("Content-Type", "application/json")
		handler(w, r)
	}))
	DeferCleanup(server.Close)
	client, err := NewClient(context.Background(), ClientOptions{Site: server.URL, Credentials: Credentials{Username: "api@example.test", APIToken: "test-api-token"}})
	Expect(err).NotTo(HaveOccurred())
	return client
}

var _ = Describe("direct Jira HTTP client", func() {
	It("reads native issues, resolves read-only fields and follows comment pagination", func() {
		client := testClient(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/rest/api/3/issue/EXAMPLE-1":
				Expect(r.URL.Query().Get("fields")).To(Equal("*all"))
				Expect(r.URL.Query().Get("updateHistory")).To(Equal("false"))
				fmt.Fprint(w, `{"id":"1","key":"EXAMPLE-1","fields":{"summary":"Example","customfield_1":"value","attachment":[]}}`)
			case "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account-1"}`)
			case "/rest/api/3/field":
				fmt.Fprint(w, `[{"id":"customfield_1","name":"Read-only field","schema":{"type":"string"}}]`)
			case "/rest/api/3/issue/EXAMPLE-1/remotelink":
				fmt.Fprint(w, `[{"id":1,"object":{"title":"Runbook","url":"https://example.test/runbook"}}]`)
			case "/rest/api/3/issue/EXAMPLE-1/comment":
				if r.URL.Query().Get("startAt") == "0" {
					fmt.Fprint(w, `{"total":2,"comments":[{"id":"c1","body":"First"}]}`)
				} else {
					fmt.Fprint(w, `{"total":2,"comments":[{"id":"c2","body":"Second"}]}`)
				}
			default:
				http.Error(w, "unexpected path", http.StatusNotFound)
			}
		})
		issue, err := client.Get(context.Background(), GetOptions{Key: "EXAMPLE-1", Store: CatalogStore{Dir: GinkgoT().TempDir()}})
		Expect(err).NotTo(HaveOccurred())
		Expect(issue.Fields["customfield_1"].Name).To(Equal("Read-only field"))
		Expect(string(issue.Fields["remoteLinks"].Value)).To(ContainSubstring("Runbook"))
		Expect(issue.Comments).To(HaveLen(2))
		Expect(*issue.CommentCount).To(Equal(2))
		data, err := json.Marshal(issue)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(SatisfyAll(ContainSubstring(`"read_only_field":"value"`), Not(ContainSubstring(`"native"`))))
		Expect(issue.URL).To(Equal(client.Site + "/browse/EXAMPLE-1"))
	})

	It("returns contextual HTTP errors without exposing credentials", func() {
		client := testClient(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "denied test-api-token", http.StatusForbidden)
		})
		_, err := client.Get(context.Background(), GetOptions{Key: "EXAMPLE-1"})
		Expect(err).To(MatchError(SatisfyAll(ContainSubstring("HTTP 403"), ContainSubstring("/issue/EXAMPLE-1"), Not(ContainSubstring("test-api-token")))))
	})

	It("rejects incomplete comment pages instead of returning a truncated thread", func() {
		client := testClient(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"total":2,"comments":[]}`) })
		_, err := client.Comments(context.Background(), "EXAMPLE-1")
		Expect(err).To(MatchError(ContainSubstring("pagination stopped")))
	})

	It("returns an empty list without fetching identity or field metadata", func() {
		client := testClient(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.URL.Path).To(Equal("/rest/api/3/search/jql"))
			fmt.Fprint(w, `{"issues":[]}`)
		})
		issues, err := client.List(context.Background(), SearchOptions{JQL: "project = EXAMPLE"})
		Expect(err).NotTo(HaveOccurred())
		Expect(issues).To(BeEmpty())
	})

	It("rebuilds the field catalog for an empty search when no-cache is requested", func() {
		var paths []string
		client := testClient(func(w http.ResponseWriter, r *http.Request) {
			paths = append(paths, r.URL.Path)
			switch r.URL.Path {
			case "/rest/api/3/search/jql":
				fmt.Fprint(w, `{"issues":[]}`)
			case "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account-1"}`)
			case "/rest/api/3/field":
				fmt.Fprint(w, `[]`)
			default:
				http.NotFound(w, r)
			}
		})
		issues, err := client.List(context.Background(), SearchOptions{JQL: "project = EXAMPLE", NoCache: true, Store: CatalogStore{Dir: GinkgoT().TempDir()}})
		Expect(err).NotTo(HaveOccurred())
		Expect(issues).To(BeEmpty())
		Expect(paths).To(Equal([]string{"/rest/api/3/search/jql", "/rest/api/3/myself", "/rest/api/3/field"}))
	})
})
