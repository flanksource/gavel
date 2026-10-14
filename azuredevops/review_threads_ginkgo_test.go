package azuredevops

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Azure native review threads", func() {
	It("collects every page and resolves only the requested native thread", func() {
		var patched []string
		c := testClient(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			switch {
			case r.Method == http.MethodPatch:
				Expect(r.URL.Path).To(HaveSuffix("/pullrequests/27/threads/8"))
				var payload map[string]string
				Expect(json.NewDecoder(r.Body).Decode(&payload)).To(Succeed())
				Expect(payload).To(Equal(map[string]string{"status": "fixed"}))
				patched = append(patched, r.URL.Path)
				fmt.Fprint(w, `{"id":8,"status":"fixed"}`)
			case strings.HasSuffix(r.URL.Path, "/threads"):
				if r.URL.Query().Get("continuationToken") == "next" {
					fmt.Fprint(w, `{"value":[{"id":8,"status":1,"comments":[{"id":12,"content":"Fix the parser","commentType":"text"}]}]}`)
				} else {
					w.Header().Set("x-ms-continuationtoken", "next")
					fmt.Fprint(w, `{"value":[{"id":7,"status":"active","comments":[{"id":11,"content":"Keep this feedback open","commentType":"text"}]}]}`)
				}
			default:
				fmt.Fprint(w, fixtureRepository)
			}
		})
		threads, err := c.ReviewThreads(context.Background(), 27)
		Expect(err).NotTo(HaveOccurred())
		Expect(threads).To(HaveLen(2))
		Expect(threads[1].ThreadID).To(Equal("8"))
		Expect(threads[1].IsReviewThread).To(BeTrue())
		Expect(c.ResolveThread(context.Background(), 27, "8")).To(Succeed())
		Expect(patched).To(HaveLen(1))
		Expect(c.ResolveThread(context.Background(), 27, "99")).To(MatchError(ContainSubstring("does not belong")))
		Expect(patched).To(HaveLen(1))
	})
})
