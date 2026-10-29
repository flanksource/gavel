package jira

import (
	"context"
	"fmt"
	"github.com/flanksource/gavel/externalissues/syncdoc"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira filesystem sync", func() {
	var client *Client
	var root string
	var title, filename string
	var downloads int
	var downloadStatus int
	var issueReads []string
	var commentBody string
	var etag string

	BeforeEach(func() {
		root = GinkgoT().TempDir()
		title, filename = "First version", "report #1.pdf"
		downloads, downloadStatus = 0, http.StatusOK
		issueReads = nil
		commentBody = `"Complete comment"`
		etag = ""
		client = testClient(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/rest/api/3/issue/EXAMPLE-1":
				issueReads = append(issueReads, "EXAMPLE-1")
				fmt.Fprintf(w, `{"id":"1","key":"EXAMPLE-1","fields":{"summary":%q,"project":{"id":"10001","key":"EXAMPLE","name":"Example"},"description":"Complete description","attachment":[{"id":"20001","filename":%q,"content":"https://example.test/report.pdf"}]}}`, title, filename)
			case "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case "/rest/api/3/field":
				fmt.Fprint(w, `[]`)
			case "/rest/api/3/issue/EXAMPLE-1/comment":
				fmt.Fprintf(w, `{"total":1,"comments":[{"id":"c1","body":%s}]}`, commentBody)
			case "/rest/api/3/issue/EXAMPLE-1/remotelink":
				fmt.Fprint(w, `[]`)
			case "/rest/api/3/attachment/content/20001":
				if etag != "" {
					w.Header().Set("ETag", etag)
				}
				if etag != "" && r.Header.Get("If-None-Match") == etag {
					w.WriteHeader(http.StatusNotModified)
					return
				}
				Expect(r.URL.Query().Get("redirect")).To(Equal("false"))
				downloads++
				w.WriteHeader(downloadStatus)
				fmt.Fprint(w, "original attachment")
			case "/rest/api/3/issue/EXAMPLE-2", "/rest/api/3/issue/EXAMPLE-3":
				issueReads = append(issueReads, strings.TrimPrefix(r.URL.Path, "/rest/api/3/issue/"))
				http.NotFound(w, r)
			default:
				http.NotFound(w, r)
			}
		})
	})

	It("validates the entire input list before requesting issues", func() {
		for _, keys := range [][]string{nil, {}, {"EXAMPLE-1", " "}} {
			_, err := client.Sync(context.Background(), SyncOptions{Keys: keys, Dir: root})
			Expect(err).To(HaveOccurred())
		}
		Expect(issueReads).To(BeEmpty())
	})

	It("stops on a failed issue and reports its position while keeping completed files", func() {
		results, err := client.Sync(context.Background(), SyncOptions{Keys: []string{"EXAMPLE-1", "EXAMPLE-2", "EXAMPLE-3"}, Dir: root, Store: CatalogStore{Dir: root}})
		Expect(err).To(MatchError(SatisfyAll(ContainSubstring("EXAMPLE-2 (2/3)"), ContainSubstring("HTTP 404"))))
		Expect(results).To(HaveLen(1))
		Expect(results[0].Path).To(BeAnExistingFile())
		Expect(issueReads).To(Equal([]string{"EXAMPLE-1", "EXAMPLE-2"}))
	})

	It("writes a full issue under its project ID without downloading attachments by default", func() {
		result, err := client.Sync(context.Background(), SyncOptions{Keys: []string{"EXAMPLE-1"}, Dir: root, Store: CatalogStore{Dir: root}})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(SyncResults{{Key: "EXAMPLE-1", Path: filepath.Join(root, "10001", "EXAMPLE-1.md")}}))
		data, err := os.ReadFile(result[0].Path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(SatisfyAll(ContainSubstring("Complete description"), ContainSubstring("Complete comment"), ContainSubstring(`[report \#1.pdf](https://example.test/report.pdf)`)))
		Expect(downloads).To(BeZero())
	})

	It("downloads originals and replaces attachment URLs with escaped relative paths", func() {
		options := SyncOptions{Keys: []string{"EXAMPLE-1"}, Dir: root, Attachments: true, Store: CatalogStore{Dir: root}}
		result, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		attachment := filepath.Join(root, "10001", "EXAMPLE-1.attachments", "20001", filename)
		Expect(result[0].Attachments).To(Equal([]string{attachment}))
		data, err := os.ReadFile(attachment)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal("original attachment"))
		data, err = os.ReadFile(result[0].Path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(SatisfyAll(ContainSubstring("EXAMPLE-1.attachments/20001/report%20%231.pdf"), ContainSubstring("https://example.test/report.pdf")))
		title = "Updated version"
		_, err = client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		data, err = os.ReadFile(result[0].Path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring(title))
		Expect(downloads).To(Equal(1))
	})

	It("writes the issue when a comment contains unsupported ADF", func() {
		commentBody = `{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Complete comment","marks":[{"type":"futureMark"}]}]},{"type":"mediaInline","attrs":{"id":"file-1","type":"file"}}]}`
		result, err := client.Sync(context.Background(), SyncOptions{Keys: []string{"EXAMPLE-1"}, Dir: root, Store: CatalogStore{Dir: root}})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(1))
		data, err := os.ReadFile(result[0].Path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(SatisfyAll(ContainSubstring("Complete description"), ContainSubstring("Complete comment"), ContainSubstring("file-1")))
	})

	It("reuses a locally verified download and preserves enrichments", func() {
		etag = `"version-1"`
		options := SyncOptions{Keys: []string{"EXAMPLE-1"}, Dir: root, Attachments: true, Store: CatalogStore{Dir: root}}
		results, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		dir, err := os.OpenRoot(filepath.Dir(results[0].Path))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(dir.Close)
		doc, err := syncdoc.Read(dir, "EXAMPLE-1.md")
		Expect(err).NotTo(HaveOccurred())
		record := doc.Metadata.Attachments["20001"]
		Expect(record.URL).To(Equal("https://example.test/report.pdf"))
		record.Enrichments = map[string]syncdoc.Enrichment{"ocr": {SourceSHA256: record.Local.SHA256, Result: yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "retained text"}}}
		doc.Extra = map[string]yaml.Node{"owner": {Kind: yaml.ScalarNode, Tag: "!!str", Value: "example"}}
		data, err := doc.Markdown("Old generated body")
		Expect(err).NotTo(HaveOccurred())
		Expect(dir.WriteFile("EXAMPLE-1.md", data, 0600)).To(Succeed())
		_, err = client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		Expect(downloads).To(Equal(1))
		doc, err = syncdoc.Read(dir, "EXAMPLE-1.md")
		Expect(err).NotTo(HaveOccurred())
		Expect(doc.Extra["owner"].Value).To(Equal("example"))
		Expect(doc.Metadata.Attachments["20001"].Enrichments["ocr"].Stale).To(BeFalse())
		options.Force = true
		_, err = client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		Expect(downloads).To(Equal(2))
	})

	It("repairs corrupt downloads without sending a cache validator", func() {
		etag = `"version-1"`
		options := SyncOptions{Keys: []string{"EXAMPLE-1"}, Dir: root, Attachments: true, Store: CatalogStore{Dir: root}}
		results, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(results[0].Attachments[0], []byte("corrupt"), 0600)).To(Succeed())
		_, err = client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		Expect(downloads).To(Equal(2))
	})

	It("retains the previous Markdown when an attachment download fails", func() {
		options := SyncOptions{Keys: []string{"EXAMPLE-1"}, Dir: root, Store: CatalogStore{Dir: root}}
		result, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		before, err := os.ReadFile(result[0].Path)
		Expect(err).NotTo(HaveOccurred())
		options.Attachments, downloadStatus, title = true, http.StatusForbidden, "Not published"
		_, err = client.Sync(context.Background(), options)
		Expect(err).To(MatchError(SatisfyAll(ContainSubstring("20001"), ContainSubstring("HTTP 403"))))
		after, err := os.ReadFile(result[0].Path)
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(Equal(before))
		entries, err := os.ReadDir(filepath.Dir(result[0].Path))
		Expect(err).NotTo(HaveOccurred())
		for _, entry := range entries {
			Expect(entry.Name()).NotTo(HavePrefix(".sync-"))
		}
	})

	It("rejects unsafe attachment filenames before writing outside the sync directory", func() {
		for _, name := range []string{"../escape.pdf", "/absolute.pdf", `folder\escape.pdf`, ".", ".."} {
			filename = name
			_, err := client.Sync(context.Background(), SyncOptions{Keys: []string{"EXAMPLE-1"}, Dir: root, Attachments: true, Store: CatalogStore{Dir: root}})
			Expect(err).To(MatchError(ContainSubstring("filename")), strings.ReplaceAll(name, "\\", "/"))
		}
		Expect(downloads).To(BeZero())
	})

	It("refuses to follow project or attachment directories outside the sync root", func() {
		outside := GinkgoT().TempDir()
		Expect(os.Symlink(outside, filepath.Join(root, "10001"))).To(Succeed())
		options := SyncOptions{Keys: []string{"EXAMPLE-1"}, Dir: root, Attachments: true, Store: CatalogStore{Dir: root}}
		_, err := client.Sync(context.Background(), options)
		Expect(err).To(HaveOccurred())
		Expect(os.Remove(filepath.Join(root, "10001"))).To(Succeed())
		Expect(os.Mkdir(filepath.Join(root, "10001"), 0o700)).To(Succeed())
		Expect(os.Symlink(outside, filepath.Join(root, "10001", "EXAMPLE-1.attachments"))).To(Succeed())
		_, err = client.Sync(context.Background(), options)
		Expect(err).To(HaveOccurred())
		entries, err := os.ReadDir(outside)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(BeEmpty())
	})
})
