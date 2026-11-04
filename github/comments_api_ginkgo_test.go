package github

import (
	"encoding/json"
	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("comments publication API", func() {
	It("publishes a captured diff range only after checking the head and retains native thread context", func(ctx SpecContext) {
		writes := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			w.Header().Set("Content-Type", "application/json")
			switch r.URL.Path {
			case "/repos/acme/demo/pulls/1":
				Expect(r.Method).To(Equal(http.MethodGet))
				_, err := w.Write([]byte(`{"head":{"sha":"head-a"}}`))
				Expect(err).NotTo(HaveOccurred())
			case "/repos/acme/demo/pulls/1/comments":
				Expect(r.Method).To(Equal(http.MethodPost))
				writes++
				var body map[string]any
				Expect(json.NewDecoder(r.Body).Decode(&body)).To(Succeed())
				Expect(body).To(HaveKeyWithValue("commit_id", "head-a"))
				Expect(body).To(HaveKeyWithValue("side", "RIGHT"))
				Expect(body).To(HaveKeyWithValue("line", float64(7)))
				_, err := w.Write([]byte(`{"id":12,"body":"Handle the error","user":{"login":"bot"},"created_at":"2026-11-04T10:00:00Z"}`))
				Expect(err).NotTo(HaveOccurred())
			case "/graphql":
				_, err := w.Write([]byte(`{"data":{"repository":{"pullRequest":{"reviewThreads":{"pageInfo":{"hasNextPage":false},"nodes":[{"id":"thread-a","viewerCanResolve":true,"comments":{"pageInfo":{"hasNextPage":false},"nodes":[{"databaseId":12,"body":"Handle the error","url":"https://github.com/acme/demo/pull/1#discussion-12","createdAt":"2026-11-04T10:00:00Z","author":{"login":"bot"},"path":"widget.go","line":7,"originalLine":6,"diffSide":"RIGHT","originalCommit":{"oid":"older-head"}}]}}]}}}}}`))
				Expect(err).NotTo(HaveOccurred())
			case "/repos/acme/demo/issues/1/comments":
				_, err := w.Write([]byte(`[]`))
				Expect(err).NotTo(HaveOccurred())
			default:
				Fail("unexpected path " + r.URL.Path)
			}
		}))
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GITHUB_API_URL", server.URL)
		api := CommentAPI{Options: Options{Repo: "acme/demo", Token: "fixture-token"}}
		selector, _ := json.Marshal(comments.CodeSelector{Kind: "diff-range", Version: 1, Path: "widget.go", BaseCommit: "base-a", HeadCommit: "head-a", Side: "new", StartLine: 7, EndLine: 7})
		publication := comments.Publication{Comment: comments.Comment{Body: "Handle the error", Target: comments.Target{URI: "https://github.com/acme/demo/pull/1", Revision: "head-a", Selector: selector}}}
		snapshot, err := api.Apply(ctx, "publish", publication)
		Expect(err).NotTo(HaveOccurred())
		Expect(writes).To(Equal(1))
		Expect(snapshot.Binding.ThreadID).To(Equal("thread-a"))
		Expect(string(snapshot.Binding.Native)).To(ContainSubstring("older-head"))
		publication.Comment.Target.Revision = "stale-head"
		_, err = api.Apply(ctx, "publish", publication)
		Expect(err).To(HaveOccurred())
		Expect(writes).To(Equal(1))
	})
})
