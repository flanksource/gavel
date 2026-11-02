package jira

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/externalissues/syncdoc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira client ignored fields", func() {
	const issueJSON = `{"id":"1","key":"EXAMPLE-1","fields":{"summary":"Visible title","project":{"id":"10001","key":"EXAMPLE"},"description":"Hidden description","updated":"2026-01-01T00:00:00Z","comment":{"total":1},"attachment":[{"id":"20001","filename":"hidden.pdf","content":"https://example.test/hidden.pdf"}],"customfield_1":"Hidden custom","customfield_2":"Visible custom"}}`
	var client, unfiltered *Client
	var paths []string
	var store CatalogStore

	BeforeEach(func() {
		paths = nil
		store = CatalogStore{Dir: GinkgoT().TempDir()}
		unfiltered = testClient(func(w http.ResponseWriter, r *http.Request) {
			paths = append(paths, r.URL.Path)
			switch r.URL.Path {
			case "/rest/api/3/issue/EXAMPLE-1", "/rest/api/3/issue/1":
				fmt.Fprint(w, issueJSON)
			case "/rest/api/3/search/jql":
				fmt.Fprint(w, `{"issues":[`+issueJSON+`]}`)
			case "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account-1"}`)
			case "/rest/api/3/field":
				fmt.Fprint(w, `[{"id":"customfield_1","name":"Internal notes"},{"id":"customfield_2","name":"Public notes"},{"id":"comment","name":"Discussion"},{"id":"attachment","name":"Files"}]`)
			case "/rest/api/3/issue/EXAMPLE-1/comment":
				fmt.Fprint(w, `{"total":1,"comments":[{"id":"c1","body":"Hidden comment"}]}`)
			case "/rest/api/3/issue/EXAMPLE-1/remotelink":
				fmt.Fprint(w, `[{"id":1,"object":{"title":"Hidden remote link"}}]`)
			case "/rest/api/3/attachment/content/20001":
				fmt.Fprint(w, "original attachment")
			default:
				http.NotFound(w, r)
			}
		})
		patterns := []string{"Internal*", "Discussion", "Remote links", "Files", "Description", "updated"}
		var err error
		client, err = NewClient(context.Background(), ClientOptions{Site: unfiltered.Site, Credentials: unfiltered.credentials, IgnoreFields: patterns})
		Expect(err).NotTo(HaveOccurred())
		patterns[0] = "Public*"
	})

	It("copies client patterns and skips excluded comments and remote links", func() {
		issue, err := client.Get(context.Background(), GetOptions{Key: "EXAMPLE-1", Full: true, Store: store})
		Expect(err).NotTo(HaveOccurred())
		Expect(issue.Fields).To(HaveLen(1))
		Expect(issue.Fields).To(HaveKey("customfield_2"))
		Expect(issue.Description).To(BeNil())
		Expect(issue.Updated).To(BeNil())
		Expect(issue.Comments).To(BeEmpty())
		Expect(issue.CommentCount).To(BeNil())
		Expect(issue.Attachments).To(BeEmpty())
		Expect(issue.AttachmentCount).To(BeNil())
		Expect(paths).To(Equal([]string{"/rest/api/3/issue/EXAMPLE-1", "/rest/api/3/myself", "/rest/api/3/field"}))
	})

	It("excludes requested search fields by catalog name", func() {
		issues, err := client.List(context.Background(), SearchOptions{JQL: "project = EXAMPLE", Fields: []string{"customfield_1", "customfield_2"}, Store: store})
		Expect(err).NotTo(HaveOccurred())
		Expect(issues).To(HaveLen(1))
		Expect(issues[0].Fields).To(HaveLen(1))
		Expect(issues[0].Fields).To(HaveKey("customfield_2"))
	})

	It("omits excluded fields and attachment metadata on initial sync", func() {
		results, err := client.Sync(context.Background(), SyncOptions{Keys: []string{"EXAMPLE-1"}, Dir: GinkgoT().TempDir(), Attachments: true, Extract: true, Store: store})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Attachments).To(BeEmpty())
		data, err := os.ReadFile(results[0].Path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(SatisfyAll(ContainSubstring("Visible title"), ContainSubstring("Visible custom"), Not(ContainSubstring("Hidden")), Not(ContainSubstring("hidden.pdf")), Not(ContainSubstring("updated_at:"))))
		Expect(paths).NotTo(ContainElement("/rest/api/3/attachment/content/20001"))
	})

	It("removes historical attachment metadata on resync while retaining local files", func() {
		results, err := unfiltered.Sync(context.Background(), SyncOptions{Keys: []string{"EXAMPLE-1"}, Dir: GinkgoT().TempDir(), Attachments: true, Store: store})
		Expect(err).NotTo(HaveOccurred())
		attachment := results[0].Attachments[0]
		paths = nil
		results, err = client.Sync(context.Background(), SyncOptions{From: results[0].Path, Store: store})
		Expect(err).NotTo(HaveOccurred())
		Expect(attachment).To(BeAnExistingFile())
		root, err := os.OpenRoot(filepath.Dir(results[0].Path))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(root.Close)
		doc, err := syncdoc.Read(root, filepath.Base(results[0].Path))
		Expect(err).NotTo(HaveOccurred())
		Expect(doc.Metadata.Attachments).To(BeEmpty())
		Expect(doc.Metadata.Source.UpdatedAt).To(BeNil())
		Expect(string(doc.Original)).NotTo(ContainSubstring("hidden.pdf"))
		Expect(paths).NotTo(ContainElement("/rest/api/3/attachment/content/20001"))
	})

	It("retains explicit comment and attachment operations", func() {
		comments, err := client.Comments(context.Background(), "EXAMPLE-1")
		Expect(err).NotTo(HaveOccurred())
		Expect(comments).To(HaveLen(1))
		attachments, err := client.Attachments(context.Background(), "EXAMPLE-1")
		Expect(err).NotTo(HaveOccurred())
		Expect(attachments).To(HaveLen(1))
	})
})
