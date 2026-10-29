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

var _ = Describe("Jira sync local links", func() {
	It("links existing cross-project issues and issues already published in the batch without fetching targets", func() {
		root := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(root, "10002"), 0o700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "10002", "EXAMPLE-2.md"), []byte("Existing"), 0o600)).To(Succeed())
		var reads []string
		client := testClient(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case r.URL.Path == "/rest/api/3/field", strings.HasSuffix(r.URL.Path, "/remotelink"):
				fmt.Fprint(w, `[]`)
			case strings.HasSuffix(r.URL.Path, "/comment"):
				fmt.Fprint(w, `{"total":0,"comments":[]}`)
			case r.URL.Path == "/rest/api/3/issue/EXAMPLE-1", r.URL.Path == "/rest/api/3/issue/EXAMPLE-4":
				key := strings.TrimPrefix(r.URL.Path, "/rest/api/3/issue/")
				reads = append(reads, key)
				target := "EXAMPLE-2"
				if key == "EXAMPLE-4" {
					target = "EXAMPLE-1"
				}
				fmt.Fprintf(w, `{"key":%q,"fields":{"summary":"References","project":{"id":"10001"},"parent":{"key":%q}}}`, key, target)
			default:
				Fail("unexpected Jira request: " + r.URL.Path)
			}
		})
		results, err := client.Sync(context.Background(), SyncOptions{Keys: []string{"EXAMPLE-1", "EXAMPLE-4"}, Dir: root, Store: CatalogStore{Dir: filepath.Join(root, "catalog")}})
		Expect(err).NotTo(HaveOccurred())
		Expect(reads).To(Equal([]string{"EXAMPLE-1", "EXAMPLE-4"}))
		for index, target := range []string{"../10002/EXAMPLE-2.md", "EXAMPLE-1.md"} {
			body, err := os.ReadFile(results[index].Path)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(ContainSubstring("](" + target + ")"))
		}
	})
})
