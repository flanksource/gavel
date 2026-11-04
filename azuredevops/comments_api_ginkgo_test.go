package azuredevops

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Azure comments publication", func() {
	It("binds inline comments to the captured iteration and rejects stale heads before writing", func(ctx SpecContext) {
		writes := 0
		client := testClient(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			switch {
			case strings.HasSuffix(r.URL.Path, "/iterations/2/changes"):
				fmt.Fprint(w, `{"changeEntries":[{"changeTrackingId":9,"item":{"path":"/widget.go"}}]}`)
			case strings.HasSuffix(r.URL.Path, "/iterations"):
				fmt.Fprint(w, `{"value":[{"id":2,"sourceRefCommit":{"commitId":"head-a"},"commonRefCommit":{"commitId":"base-a"}}]}`)
			case strings.HasSuffix(r.URL.Path, "/pullrequests/27"):
				fmt.Fprint(w, `{"lastMergeSourceCommit":{"commitId":"head-a"}}`)
			case r.Method == http.MethodPost:
				writes++
				var payload map[string]any
				Expect(json.NewDecoder(r.Body).Decode(&payload)).To(Succeed())
				Expect(payload["pullRequestThreadContext"]).To(Equal(map[string]any{"changeTrackingId": float64(9), "iterationContext": map[string]any{"firstComparingIteration": float64(0), "secondComparingIteration": float64(2)}}))
				fmt.Fprint(w, `{"id":7,"status":"active","comments":[{"id":11,"content":"Check this","commentType":"text","publishedDate":"2026-11-04T10:00:00Z","author":{"id":"native-user"}}],"threadContext":{"filePath":"/widget.go"},"pullRequestThreadContext":{"changeTrackingId":9}}`)
			default:
				fmt.Fprint(w, fixtureRepository)
			}
		})
		s := comments.CodeSelector{Kind: "diff-range", Version: 1, Path: "widget.go", BaseCommit: "base-a", HeadCommit: "head-a", Side: "new", StartLine: 7, EndLine: 7}
		selector, err := json.Marshal(s)
		Expect(err).NotTo(HaveOccurred())
		p := comments.Publication{Comment: comments.Comment{Body: "Check this", Target: comments.Target{URI: client.repo.URL() + "/pullrequest/27", Revision: "head-a", Selector: selector}}}
		remote, err := client.ApplyComment(ctx, "publish", p)
		Expect(err).NotTo(HaveOccurred())
		Expect(remote.Binding.ThreadID).To(Equal("7"))
		Expect(string(remote.Binding.Native)).To(ContainSubstring("changeTrackingId"))
		p.Comment.Target.Revision = "old-head"
		_, err = client.ApplyComment(ctx, "publish", p)
		Expect(err).To(HaveOccurred())
		Expect(writes).To(Equal(1))
	})
})
