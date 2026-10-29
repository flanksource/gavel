package jira

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira sync inline media", func() {
	It("renders local description and comment images without redownloading cached originals", func() {
		downloads := 0
		client := testClient(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/rest/api/3/issue/EXAMPLE-1":
				fmt.Fprint(w, `{"id":"1","key":"EXAMPLE-1","fields":{"summary":"Inline images","project":{"id":"10001"},"description":{"type":"doc","version":1,"content":[{"type":"mediaSingle","content":[{"type":"media","attrs":{"id":"description-uuid","type":"file","alt":"description.png"}}]}]},"attachment":[{"id":"20001","filename":"description.png","mimeType":"image/png","content":"https://example.test/description.png"},{"id":"20002","filename":"comment.png","mimeType":"image/png","content":"https://example.test/comment.png"}]}}`)
			case r.URL.Path == "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case r.URL.Path == "/rest/api/3/field", strings.HasSuffix(r.URL.Path, "/remotelink"):
				fmt.Fprint(w, `[]`)
			case strings.HasSuffix(r.URL.Path, "/comment"):
				fmt.Fprint(w, `{"total":1,"comments":[{"id":"comment-1","author":{"displayName":"Alex"},"body":{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"mediaInline","attrs":{"id":"comment-uuid","type":"file","alt":"comment.png"}}]}]}}]}`)
			case strings.HasPrefix(r.URL.Path, "/rest/api/3/attachment/content/"):
				downloads++
				fmt.Fprint(w, "image original "+filepath.Base(r.URL.Path))
			default:
				Fail("unexpected Jira request: " + r.URL.Path)
			}
		})
		dir := GinkgoT().TempDir()
		options := SyncOptions{Keys: []string{"EXAMPLE-1"}, Dir: dir, Extract: true, Store: CatalogStore{Dir: dir}}
		for range 2 {
			results, err := client.Sync(context.Background(), options)
			Expect(err).NotTo(HaveOccurred())
			data, err := os.ReadFile(results[0].Path)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(SatisfyAll(ContainSubstring("![description.png](EXAMPLE-1.attachments/20001/description.png)"), ContainSubstring("![comment.png](EXAMPLE-1.attachments/20002/comment.png)"), Not(ContainSubstring("atlassian-media:"))))
			Expect(results[0].Extracted).To(BeEmpty())
			Expect(downloads).To(Equal(2))
		}
	})
})
