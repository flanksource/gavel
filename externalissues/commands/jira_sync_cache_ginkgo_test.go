package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/externalissues/jira"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira sync attachment cache", func() {
	It("downloads every unsupported attachment type and reuses originals after moving the saved issue", func() {
		files := map[string]string{"20001": "image.png", "20002": "costing.xls", "20003": "billing.xlsx", "20004": "source.bin"}
		var contentRequests []string
		commandServer(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case r.URL.Path == "/rest/api/3/field", strings.HasSuffix(r.URL.Path, "/remotelink"):
				fmt.Fprint(w, `[]`)
			case strings.HasSuffix(r.URL.Path, "/comment"):
				fmt.Fprint(w, `{"total":0,"comments":[]}`)
			case strings.HasPrefix(r.URL.Path, "/rest/api/3/issue/"):
				attachments := make([]map[string]string, 0, len(files))
				for id, name := range files {
					attachments = append(attachments, map[string]string{"id": id, "filename": name})
				}
				Expect(json.NewEncoder(w).Encode(map[string]any{"id": "10001", "key": "EXAMPLE-1", "fields": map[string]any{"summary": "All original types", "project": map[string]string{"id": "10001"}, "attachment": attachments}})).To(Succeed())
			case strings.HasPrefix(r.URL.Path, "/rest/api/3/attachment/content/"):
				id := filepath.Base(r.URL.Path)
				contentRequests = append(contentRequests, id)
				fmt.Fprint(w, files[id])
			default:
				http.NotFound(w, r)
			}
		})
		dir := GinkgoT().TempDir()
		output, err := commandOutput([]string{"sync", "EXAMPLE-1", "--sync-dir", dir, "--extract"}, "json")
		Expect(err).NotTo(HaveOccurred())
		var results jira.SyncResults
		Expect(json.Unmarshal([]byte(output), &results)).To(Succeed())
		Expect(results).To(HaveLen(1))
		Expect(results[0].Attachments).To(HaveLen(len(files)))
		Expect(results[0].Extracted).To(BeEmpty())
		Expect(contentRequests).To(ConsistOf("20001", "20002", "20003", "20004"))
		data, err := os.ReadFile(results[0].Path)
		Expect(err).NotTo(HaveOccurred())
		for id, name := range files {
			path := filepath.Join(dir, "10001", "EXAMPLE-1.attachments", id, name)
			original, err := os.ReadFile(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(original)).To(Equal(name))
			Expect(string(data)).To(ContainSubstring("[" + name + "](EXAMPLE-1.attachments/" + id + "/" + name + ")"))
		}
		moved := filepath.Join(dir, "moved")
		Expect(os.Rename(filepath.Join(dir, "10001"), moved)).To(Succeed())
		path := filepath.Join(moved, "renamed.md")
		Expect(os.Rename(filepath.Join(moved, "EXAMPLE-1.md"), path)).To(Succeed())
		contentRequests = nil
		_, err = commandOutput([]string{"sync", "--from", path}, "json")
		Expect(err).NotTo(HaveOccurred())
		Expect(contentRequests).To(BeEmpty())
		Expect(os.Remove(filepath.Join(moved, "EXAMPLE-1.attachments", "20001", files["20001"]))).To(Succeed())
		_, err = commandOutput([]string{"sync", "--from", path}, "json")
		Expect(err).NotTo(HaveOccurred())
		Expect(contentRequests).To(Equal([]string{"20001"}))
		contentRequests = nil
		_, err = commandOutput([]string{"sync", "--from", path, "--force"}, "json")
		Expect(err).NotTo(HaveOccurred())
		Expect(contentRequests).To(ConsistOf("20001", "20002", "20003", "20004"))
	})
})
