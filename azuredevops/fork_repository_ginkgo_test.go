package azuredevops

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const fixtureForkRepository = `{"id":"fork-id","name":"source-service","project":{"id":"fork-project-id","name":"source-project"}}`

var _ = Describe("fork source repository resolution", func() {
	DescribeTable("resolves partial fork metadata while listing open PRs", func(repository string) {
		lookups := 0
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			Expect(r.Method).To(Equal(http.MethodGet))
			switch r.URL.Path {
			case "/acme/product/_apis/git/repositories/service":
				fmt.Fprint(w, fixtureRepository)
			case "/acme/product/_apis/git/repositories/repo-id/pullrequests":
				fmt.Fprintf(w, `{"value":[%s,"forkSource":{"repository":%s}}]}`, strings.TrimSuffix(fixturePR, "}"), repository)
			case "/acme/_apis/git/repositories/fork-id":
				lookups++
				fmt.Fprint(w, fixtureForkRepository)
			default:
				Fail("unexpected request: " + r.URL.String())
			}
		})
		prs, err := c.OpenPRs(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(prs).To(HaveLen(1))
		Expect(prs[0].HeadRepositoryURL).To(Equal("https://dev.azure.com/acme/source-project/_git/source-service"))
		Expect(prs[0].HeadRefName).To(Equal("feature"))
		Expect(prs[0].BaseRefName).To(Equal("main"))
		Expect(lookups).To(Equal(1))
	},
		Entry("repository ID only", `{"id":"fork-id"}`),
		Entry("project ID without project name", `{"id":"fork-id","name":"source-service","project":{"id":"fork-project-id"}}`),
		Entry("missing repository name", `{"id":"fork-id","project":{"name":"source-project"}}`),
	)

	It("uses the resolved fork for branch and build queries and resolves it once per status request", func() {
		lookups := 0
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			switch {
			case strings.HasSuffix(r.URL.Path, "/pullrequests/27"):
				fmt.Fprint(w, strings.TrimSuffix(fixturePR, "}")+`,"forkSource":{"repository":{"id":"fork-id"}}}`)
			case r.URL.Path == "/acme/_apis/git/repositories/fork-id":
				lookups++
				fmt.Fprint(w, fixtureForkRepository)
			case strings.HasSuffix(r.URL.Path, "/refs"):
				Expect(r.URL.Path).To(Equal("/acme/source-project/_apis/git/repositories/fork-id/refs"))
				fmt.Fprint(w, `{"value":[{"name":"refs/heads/feature","objectId":"head-sha"}]}`)
			case strings.HasSuffix(r.URL.Path, "/builds"):
				if r.URL.Query().Get("branchName") == "refs/heads/feature" {
					Expect(r.URL.Path).To(Equal("/acme/source-project/_apis/build/builds"))
					Expect(r.URL.Query().Get("repositoryId")).To(Equal("fork-id"))
				}
				fmt.Fprint(w, `{"value":[]}`)
			case strings.HasSuffix(r.URL.Path, "/evaluations"):
				fmt.Fprint(w, `{"value":[]}`)
			default:
				fmt.Fprint(w, fixtureRepository)
			}
		})
		result, err := c.Status(context.Background(), 27, model.StatusOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.PR.HeadRepositoryURL).To(Equal("https://dev.azure.com/acme/source-project/_git/source-service"))
		Expect(result.PR.MergeReadiness.State).To(Equal("ready"))
		Expect(lookups).To(Equal(1))
	})

	DescribeTable("rejects a fork lookup that cannot establish the source identity", func(status int, response, expected string) {
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasSuffix(r.URL.Path, "/pullrequests"):
				fmt.Fprint(w, `{"value":[`+strings.TrimSuffix(fixturePR, "}")+`,"forkSource":{"repository":{"id":"fork-id"}}}]}`)
			case r.URL.Path == "/acme/_apis/git/repositories/fork-id":
				w.WriteHeader(status)
				fmt.Fprint(w, response)
			default:
				fmt.Fprint(w, fixtureRepository)
			}
		})
		_, err := c.OpenPRs(context.Background())
		Expect(err).To(MatchError(ContainSubstring(expected)))
	},
		Entry("permission denied", http.StatusForbidden, `{"message":"denied"}`, "HTTP 403"),
		Entry("malformed JSON", http.StatusOK, `not JSON`, "decode Azure fork source repository"),
		Entry("missing project", http.StatusOK, `{"id":"fork-id","name":"source-service"}`, "missing repository or project identity"),
		Entry("different repository", http.StatusOK, fixtureRepository, "repository ID"),
	)

	It("rejects a fork reference without a repository ID instead of using the target repository", func() {
		requests := 0
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			requests++
			if strings.HasSuffix(r.URL.Path, "/pullrequests") {
				fmt.Fprint(w, `{"value":[`+strings.TrimSuffix(fixturePR, "}")+`,"forkSource":{"repository":{}}}]}`)
				return
			}
			fmt.Fprint(w, fixtureRepository)
		})
		_, err := c.OpenPRs(context.Background())
		Expect(err).To(MatchError(ContainSubstring("PR #27 fork source is missing repository ID")))
		Expect(requests).To(Equal(2))
	})
})
