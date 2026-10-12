package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"strings"
)

const fixtureRepository = `{"id":"repo-id","name":"service","defaultBranch":"refs/heads/main","project":{"id":"project-id","name":"product"}}`

func testClient(handler http.HandlerFunc) *Client {
	server := httptest.NewServer(handler)
	DeferCleanup(server.Close)
	c := New(Repository{Organization: "acme", Project: "product", Name: "service"})
	c.baseURL, c.auth.pat = server.URL, "fixture-pat"
	return c
}

var _ = Describe("Azure REST client", func() {
	It("creates draft PRs using full refs and the browser URL", func() {
		var payload map[string]json.RawMessage
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			Expect(r.Header.Get("Authorization")).To(HavePrefix("Basic "))
			Expect(r.URL.Query().Get("api-version")).To(Equal("7.1"))
			if r.Method == http.MethodGet {
				fmt.Fprint(w, fixtureRepository)
				return
			}
			Expect(r.URL.Path).To(Equal("/acme/product/_apis/git/repositories/repo-id/pullrequests"))
			Expect(json.NewDecoder(r.Body).Decode(&payload)).To(Succeed())
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"pullRequestId":27,"title":"Add feature","status":"active"}`)
		})
		result, err := c.CreatePR(context.Background(), model.CreatePRInput{Title: "Add feature", Body: "Why", Head: "feature/topic", Base: "main", Draft: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(payload).To(Equal(map[string]json.RawMessage{"title": json.RawMessage(`"Add feature"`), "description": json.RawMessage(`"Why"`), "sourceRefName": json.RawMessage(`"refs/heads/feature/topic"`), "targetRefName": json.RawMessage(`"refs/heads/main"`), "isDraft": json.RawMessage(`true`)}))
		Expect(result.Number).To(Equal(27))
		Expect(result.URL).To(Equal("https://dev.azure.com/acme/product/_git/service/pullrequest/27"))
		Expect(result.Base).To(Equal("main"))
	})
	It("returns permission failures without invoking another credential source", func() {
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"message":"Build read permission required"}`)
		})
		c.auth.fetch = func(context.Context) (accessToken, error) {
			Fail("must not replace a configured PAT")
			return accessToken{}, nil
		}
		err := c.Preflight(context.Background())
		Expect(err).To(MatchError(ContainSubstring("HTTP 403")))
		Expect(err.Error()).To(ContainSubstring("Build read permission required"))
	})
	It("rejects malformed repository metadata", func() {
		c := testClient(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{}`) })
		Expect(c.Preflight(context.Background())).To(MatchError(ContainSubstring("repository")))
	})
	It("collects paginated open PRs", func() {
		pages := 0
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			if !strings.HasSuffix(r.URL.Path, "/pullrequests") {
				fmt.Fprint(w, fixtureRepository)
				return
			}
			Expect(r.URL.Query().Get("searchCriteria.status")).To(Equal("active"))
			pages++
			if pages == 1 {
				w.Header().Set("x-ms-continuationtoken", "next")
				fmt.Fprint(w, `{"value":[{"pullRequestId":27,"title":"Feature","status":"active","sourceRefName":"refs/heads/feature"}]}`)
				return
			}
			Expect(r.URL.Query().Get("continuationToken")).To(Equal("next"))
			fmt.Fprint(w, `{"value":[{"pullRequestId":28,"title":"Another","status":"active","sourceRefName":"refs/heads/another"}]}`)
		})
		prs, err := c.OpenPRs(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(prs).To(HaveLen(2))
		Expect(prs[0].HeadRefName).To(Equal("feature"))
		Expect(pages).To(Equal(2))
	})
})
