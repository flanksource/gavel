package jira

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/externalissues"
	"github.com/flanksource/gavel/externalissues/syncdoc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

func readCachedIssue(path string) (*syncdoc.Document, *os.Root) {
	GinkgoHelper()
	root, err := os.OpenRoot(filepath.Dir(path))
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(root.Close)
	doc, err := syncdoc.Read(root, filepath.Base(path))
	Expect(err).NotTo(HaveOccurred())
	return doc, root
}

func writeCachedIssue(path string, doc *syncdoc.Document) {
	GinkgoHelper()
	data, err := doc.Markdown("Previous generated issue")
	Expect(err).NotTo(HaveOccurred())
	Expect(os.WriteFile(path, data, 0600)).To(Succeed())
}

var _ = Describe("Jira attachment cache", func() {
	const attachmentID = "20001"
	var client *Client
	var options SyncOptions
	var body, title, name, etag, id string
	var remoteSize int
	var contentRequests []string
	var contentStatus int

	BeforeEach(func() {
		body, title, name, id = "Original attachment", "Initial title", "notes.txt", attachmentID
		etag, remoteSize, contentRequests, contentStatus = "", len(body), nil, http.StatusOK
		dir := GinkgoT().TempDir()
		options = SyncOptions{Keys: []string{"EXAMPLE-1"}, Dir: dir, Attachments: true, Store: CatalogStore{Dir: dir}}
		client = testClient(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/rest/api/3/issue/EXAMPLE-1":
				Expect(json.NewEncoder(w).Encode(map[string]any{"id": "1", "key": "EXAMPLE-1", "fields": map[string]any{
					"summary": title, "project": map[string]string{"id": "10001"},
					"attachment": []map[string]any{{"id": id, "filename": name, "size": remoteSize, "content": "https://example.test/original"}},
				}})).To(Succeed())
			case r.URL.Path == "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case r.URL.Path == "/rest/api/3/field", strings.HasSuffix(r.URL.Path, "/remotelink"):
				fmt.Fprint(w, `[]`)
			case strings.HasSuffix(r.URL.Path, "/comment"):
				fmt.Fprint(w, `{"total":0,"comments":[]}`)
			case strings.HasPrefix(r.URL.Path, "/rest/api/3/attachment/content/"):
				contentRequests = append(contentRequests, filepath.Base(r.URL.Path))
				Expect(r.Header.Get("If-None-Match")).To(BeEmpty())
				if etag != "" {
					w.Header().Set("ETag", etag)
				}
				w.WriteHeader(contentStatus)
				fmt.Fprint(w, body)
			default:
				http.NotFound(w, r)
			}
		})
	})

	DescribeTable("reuses verified originals without content requests and preserves metadata", func(validator string) {
		etag = validator
		first, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		doc, _ := readCachedIssue(first[0].Path)
		record := doc.Metadata.Attachments[id]
		record.Local.Extra = map[string]yaml.Node{"owner": {Kind: yaml.ScalarNode, Value: "example"}}
		record.Enrichments = map[string]syncdoc.Enrichment{"ocr": {SourceSHA256: record.Local.SHA256}}
		original, err := yaml.Marshal(record.Local)
		Expect(err).NotTo(HaveOccurred())
		writeCachedIssue(first[0].Path, doc)
		title, contentRequests = "Updated title", nil
		options.Extract, options.NoCache = true, true
		task.ClearTasks()
		DeferCleanup(task.ClearTasks)
		second, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		Expect(contentRequests).To(BeEmpty())
		Expect(second[0].Attachments).To(Equal(first[0].Attachments))
		doc, _ = readCachedIssue(second[0].Path)
		current, err := yaml.Marshal(doc.Metadata.Attachments[id].Local)
		Expect(err).NotTo(HaveOccurred())
		Expect(current).To(Equal(original))
		Expect(doc.Metadata.Attachments[id].Enrichments["ocr"].Stale).To(BeFalse())
		data, err := os.ReadFile(second[0].Path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring(title))
		Expect(jiraProgressSnapshot("Download Jira attachment " + id).ID).To(BeEmpty())
	}, Entry("without an ETag", ""), Entry("with a strong ETag", `"original"`), Entry("with a weak ETag", `W/"original"`))

	DescribeTable("repairs only the affected original", func(state string) {
		first, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		doc, _ := readCachedIssue(first[0].Path)
		switch state {
		case "missing":
			Expect(os.Remove(first[0].Attachments[0])).To(Succeed())
		case "corrupt":
			Expect(os.WriteFile(first[0].Attachments[0], []byte(strings.Repeat("x", len(body))), 0600)).To(Succeed())
		case "hash":
			doc.Metadata.Attachments[id].Local.SHA256 = ""
			writeCachedIssue(first[0].Path, doc)
		case "record":
			doc.Metadata.Attachments[id].Local = nil
			writeCachedIssue(first[0].Path, doc)
		case "size":
			body += " appended content"
			remoteSize = len(body)
		case "new":
			id = "20002"
		case "force":
			options.Force = true
		}
		contentRequests = nil
		second, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		Expect(contentRequests).To(Equal([]string{id}))
		data, err := os.ReadFile(second[0].Attachments[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal(body))
	}, Entry("missing file", "missing"), Entry("same-size corruption", "corrupt"), Entry("missing hash", "hash"), Entry("missing record", "record"), Entry("changed positive size", "size"), Entry("new attachment ID", "new"), Entry("forced refresh", "force"))

	DescribeTable("reuses files with unspecified size and empty files", func(empty bool) {
		remoteSize = 0
		if empty {
			body = ""
		}
		_, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		contentRequests = nil
		_, err = client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		Expect(contentRequests).To(BeEmpty())
	}, Entry("size omitted", false), Entry("empty original", true))

	It("reuses a recorded original at its current renamed path", func() {
		first, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		name, contentRequests = "renamed.txt", nil
		second, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		Expect(contentRequests).To(BeEmpty())
		Expect(second[0].Attachments).To(Equal([]string{filepath.Join(filepath.Dir(first[0].Path), "EXAMPLE-1.attachments", id, name)}))
		Expect(second[0].Attachments[0]).To(BeAnExistingFile())
	})

	It("fails on unreadable originals without downloading or replacing the issue", func() {
		first, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		before, err := os.ReadFile(first[0].Path)
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chmod(first[0].Attachments[0], 0000)).To(Succeed())
		DeferCleanup(os.Chmod, first[0].Attachments[0], os.FileMode(0600))
		contentRequests = nil
		_, err = client.Sync(context.Background(), options)
		Expect(errors.Is(err, os.ErrPermission)).To(BeTrue(), "%v", err)
		Expect(contentRequests).To(BeEmpty())
		after, err := os.ReadFile(first[0].Path)
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(Equal(before))
	})

	It("preserves the issue and corrupt original when its repair download fails", func() {
		first, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		before, err := os.ReadFile(first[0].Path)
		Expect(err).NotTo(HaveOccurred())
		const corrupt = "Damaged content"
		Expect(os.WriteFile(first[0].Attachments[0], []byte(corrupt), 0600)).To(Succeed())
		contentStatus = http.StatusForbidden
		_, err = client.Sync(context.Background(), options)
		Expect(err).To(MatchError(ContainSubstring("HTTP 403")))
		after, err := os.ReadFile(first[0].Path)
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(Equal(before))
		original, err := os.ReadFile(first[0].Attachments[0])
		Expect(err).NotTo(HaveOccurred())
		Expect(string(original)).To(Equal(corrupt))
	})

	It("honors cancellation before staging a verified original", func() {
		first, err := client.Sync(context.Background(), options)
		Expect(err).NotTo(HaveOccurred())
		doc, root := readCachedIssue(first[0].Path)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		stage := GinkgoT().TempDir()
		contentRequests = nil
		err = client.syncDownload(ctx, syncDownloadOptions{Attachment: &externalissues.Attachment{ID: id, Size: int64(remoteSize)}, Record: doc.Metadata.Attachments[id], Root: root, Stage: stage, Relative: "notes.txt"})
		Expect(err).To(MatchError(context.Canceled))
		Expect(filepath.Join(stage, "notes.txt")).NotTo(BeAnExistingFile())
		Expect(contentRequests).To(BeEmpty())
	})
})
