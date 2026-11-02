package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/externalissues/syncdoc"
	"github.com/flanksource/gavel/workspace"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

func describeIgnoreFieldResync(run func(string, []string)) {
	const original = "Private attachment content"
	var workDir, syncDir, configPath string
	var requests []string

	BeforeEach(func() {
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		workDir, syncDir = GinkgoT().TempDir(), GinkgoT().TempDir()
		configPath = filepath.Join(home, ".gavel.yaml")
		Expect(os.WriteFile(configPath, []byte("externalIssues: {jira: {ignoreFields: [Initially hidden]}}\n"), 0o600)).To(Succeed())
		requests = nil
		commandServer(func(w http.ResponseWriter, r *http.Request) {
			requests = append(requests, r.URL.Path)
			switch {
			case r.URL.Path == "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case r.URL.Path == "/rest/api/3/field":
				fmt.Fprint(w, `[{"id":"customfield_1","name":"Internal notes"},{"id":"customfield_2","name":"Sensitive detail"},{"id":"customfield_3","name":"Private status"},{"id":"customfield_4","name":"Public notes"},{"id":"customfield_5","name":"Initially hidden"}]`)
			case strings.HasSuffix(r.URL.Path, "/comment"):
				fmt.Fprint(w, `{"total":1,"comments":[{"id":"comment-1","body":"Hidden comment"}]}`)
			case strings.HasSuffix(r.URL.Path, "/remotelink"):
				fmt.Fprint(w, `[{"id":1,"object":{"title":"Hidden remote link","url":"https://example.test/private"}}]`)
			case r.URL.Path == "/rest/api/3/attachment/content/20001":
				fmt.Fprint(w, original)
			case strings.HasPrefix(r.URL.Path, "/rest/api/3/issue/"):
				key := strings.TrimPrefix(r.URL.Path, "/rest/api/3/issue/")
				if key == "1" {
					key = "EXAMPLE-1"
				}
				id := strings.TrimPrefix(key, "EXAMPLE-")
				Expect(json.NewEncoder(w).Encode(map[string]any{"id": id, "key": key, "fields": map[string]any{
					"summary": "Visible title", "project": map[string]string{"id": "10001", "key": "EXAMPLE"},
					"description": "Hidden description", "priority": map[string]string{"name": "Hidden priority"},
					"labels": []string{"hidden-label"}, "updated": "2026-01-01T00:00:00Z",
					"customfield_1": "Hidden by name", "customfield_2": "Hidden by ID", "customfield_3": "Hidden by wildcard",
					"customfield_4": "Visible custom value", "customfield_5": "Initially hidden value",
					"attachment": []map[string]any{{"id": "20001", "filename": "private.txt", "size": len(original), "content": "https://example.test/private.txt"}},
				}})).To(Succeed())
			default:
				http.NotFound(w, r)
			}
		})
	})

	DescribeTable("removes newly ignored content with unchanged Jira data and a cached field catalog",
		func(from bool) {
			By("saving two issues before extending the ignore list")
			run(workDir, []string{"sync", "EXAMPLE-1", "EXAMPLE-2", "--sync-dir", syncDir, "--extract"})
			path := filepath.Join(syncDir, "10001", "EXAMPLE-1.md")
			unselected := filepath.Join(syncDir, "10001", "EXAMPLE-2.md")
			untouched, err := os.ReadFile(unselected)
			Expect(err).NotTo(HaveOccurred())
			root, err := os.OpenRoot(filepath.Dir(path))
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(root.Close)
			doc, err := syncdoc.Read(root, filepath.Base(path))
			Expect(err).NotTo(HaveOccurred())
			Expect(doc.Metadata.Source.UpdatedAt).NotTo(BeNil())
			Expect(doc.Metadata.Attachments).To(HaveKey("20001"))
			Expect(doc.Metadata.Sync.Extract).To(BeTrue())
			for _, value := range []string{"Hidden by name", "Hidden by ID", "Hidden by wildcard", "Hidden description", "Hidden priority", "hidden-label", "Hidden comment", "Hidden remote link", "private.txt"} {
				Expect(string(doc.Original)).To(ContainSubstring(value))
			}
			Expect(string(doc.Original)).NotTo(ContainSubstring("Initially hidden value"))
			Expect(requests).To(ContainElement("/rest/api/3/field"))

			By("retaining custom frontmatter and replacing old generated extraction content")
			doc.Extra = map[string]yaml.Node{"reviewer": {Kind: yaml.ScalarNode, Tag: "!!str", Value: "Local reviewer"}}
			_, body, ok := bytes.Cut(doc.Original, []byte("\n---\n"))
			Expect(ok).To(BeTrue())
			data, err := doc.Markdown(string(body) + "\n\n## Extracted attachments\n\nHidden extracted content")
			Expect(err).NotTo(HaveOccurred())
			Expect(os.WriteFile(path, data, 0o600)).To(Succeed())
			Expect(os.WriteFile(configPath, []byte("externalIssues: {jira: {ignoreFields: [Initially hidden, Internal notes, customfield_2, 'Private*', description, priority, labels, updated, comment, attachment, remoteLinks]}}\n"), 0o600)).To(Succeed())

			By("resyncing only the selected issue using the current configuration")
			requests = nil
			args := []string{"sync", "EXAMPLE-1", "--sync-dir", syncDir, "--extract"}
			issuePath := "/rest/api/3/issue/EXAMPLE-1"
			if from {
				args = []string{"sync", "--from", path}
				issuePath = "/rest/api/3/issue/1"
			}
			run(workDir, args)
			doc, err = syncdoc.Read(root, filepath.Base(path))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(doc.Original)).To(SatisfyAll(ContainSubstring("Visible title"), ContainSubstring("Visible custom value"), ContainSubstring("Local reviewer"), Not(ContainSubstring("Hidden")), Not(ContainSubstring("hidden-label")), Not(ContainSubstring("private.txt")), Not(ContainSubstring("Extracted attachments"))))
			Expect(doc.Metadata.Attachments).To(BeEmpty())
			Expect(doc.Metadata.Source.UpdatedAt).To(BeNil())
			Expect(doc.Extra["reviewer"].Value).To(Equal("Local reviewer"))
			for _, name := range []string{"Internal notes", "Sensitive detail", "Private status", "## Comments"} {
				Expect(string(doc.Original)).NotTo(ContainSubstring(name))
			}
			data, err = root.ReadFile(filepath.Join("EXAMPLE-1.attachments", "20001", "private.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal(original))
			data, err = os.ReadFile(unselected)
			Expect(err).NotTo(HaveOccurred())
			Expect(data).To(Equal(untouched))
			Expect(requests).To(ConsistOf(issuePath, "/rest/api/3/myself"))
		},
		Entry("positional issue sync", false),
		Entry("saved-file sync", true),
	)
}

var _ = Describe("Jira ignore-field resync commands", func() {
	describeIgnoreFieldResync(func(workDir string, args []string) {
		root := commandRoot()
		root.SetContext(workspace.With(context.Background(), workspace.Dirs{Invocation: workDir}))
		output := filepath.Join(GinkgoT().TempDir(), "result.json")
		root.SetArgs(append(append([]string{"external-issues", "jira"}, args...), "--format", "json="+output))
		Expect(root.Execute()).To(Succeed())
	})
})
