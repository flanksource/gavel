package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/clicky/formatters/excel"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/xuri/excelize/v2"
)

func writeSyncWorkbook(path string, keys []string) {
	book := excelize.NewFile()
	DeferCleanup(func() { Expect(book.Close()).To(Succeed()) })
	Expect(book.SetSheetRow("Sheet1", "A1", &[]any{"Issue Key"})).To(Succeed())
	Expect(book.SetSheetRow("Sheet1", "A2", &[]any{"IGNORED-1"})).To(Succeed())
	_, err := book.NewSheet("Sprint Backlog")
	Expect(err).NotTo(HaveOccurred())
	Expect(book.SetSheetRow("Sprint Backlog", "A1", &[]any{"Issue Key"})).To(Succeed())
	for i, key := range keys {
		Expect(book.SetSheetRow("Sprint Backlog", fmt.Sprintf("A%d", i+2), &[]any{key})).To(Succeed())
	}
	Expect(book.SaveAs(path)).To(Succeed())
}

var _ = Describe("Jira sync command", func() {
	It("resyncs a renamed file in place with explicit false attachment overrides", func() {
		commandServer(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case r.URL.Path == "/rest/api/3/field" || strings.HasSuffix(r.URL.Path, "/remotelink"):
				fmt.Fprint(w, `[]`)
			case strings.HasSuffix(r.URL.Path, "/comment"):
				fmt.Fprint(w, `{"total":0,"comments":[]}`)
			default:
				fmt.Fprint(w, `{"id":"10001","key":"EXAMPLE-1","fields":{"summary":"Metadata sync","project":{"id":"10001"}}}`)
			}
		})
		dir := GinkgoT().TempDir()
		_, err := commandOutput([]string{"sync", "EXAMPLE-1", "--sync-dir", dir, "--extract"}, "json")
		Expect(err).NotTo(HaveOccurred())
		path := filepath.Join(dir, "10001", "renamed.md")
		Expect(os.Rename(filepath.Join(dir, "10001", "EXAMPLE-1.md"), path)).To(Succeed())
		output, err := commandOutput([]string{"sync", "--from", path, "--attachments=false", "--extract=false"}, "json")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring("renamed.md"))
		data, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(SatisfyAll(ContainSubstring("attachments: false"), ContainSubstring("extract: false")))
	})
	It("rejects a missing worksheet before discovering Jira credentials", func() {
		path := filepath.Join(GinkgoT().TempDir(), "issues.xlsx")
		writeSyncWorkbook(path, []string{"EXAMPLE-1"})
		_, err := commandOutput([]string{"sync", "@" + path + ":Missing!Issue Key", "--site", "https://user:password@example.test"}, "json")
		Expect(err).To(MatchError(And(ContainSubstring(path), ContainSubstring(`worksheet "Missing" not found`), ContainSubstring("Sprint Backlog"))))
	})

	It("rejects an empty spreadsheet column before discovering Jira credentials", func() {
		path := filepath.Join(GinkgoT().TempDir(), "empty.xlsx")
		Expect(excel.WriteFile([]map[string]any{{"Issue Key": ""}}, path)).To(Succeed())
		_, err := commandOutput([]string{"sync", "@" + path + ":Issue Key", "--site", "https://user:password@example.test"}, "json")
		Expect(err).To(MatchError("Jira sync requires at least one issue key or ID"))
	})

	It("syncs multiple positional IDs and selected Excel worksheets using one catalog refresh", func() {
		for _, selector := range []string{"", "Issue Key", "'Sprint Backlog'!Issue Key", "#2!Issue Key"} {
			var keys []string
			catalogs, identities := 0, 0
			commandServer(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.URL.Path == "/rest/api/3/myself":
					identities++
					fmt.Fprint(w, `{"accountId":"account"}`)
				case r.URL.Path == "/rest/api/3/field":
					catalogs++
					fmt.Fprint(w, `[]`)
				case strings.HasSuffix(r.URL.Path, "/comment"):
					fmt.Fprint(w, `{"total":0,"comments":[]}`)
				case strings.HasSuffix(r.URL.Path, "/remotelink"):
					fmt.Fprint(w, `[]`)
				case r.URL.Path == "/rest/api/3/issue/EXAMPLE-1" || r.URL.Path == "/rest/api/3/issue/EXAMPLE-2":
					key := strings.TrimPrefix(r.URL.Path, "/rest/api/3/issue/")
					keys = append(keys, key)
					fmt.Fprintf(w, `{"key":%q,"fields":{"summary":"Batch sync","project":{"id":"10001"}}}`, key)
				default:
					http.NotFound(w, r)
				}
			})
			dir := GinkgoT().TempDir()
			args := []string{"sync", "EXAMPLE-1", "EXAMPLE-2", "--sync-dir", dir, "--no-cache"}
			if selector != "" {
				path := filepath.Join(dir, "issues.xlsx")
				if selector == "Issue Key" {
					Expect(excel.WriteFile([]map[string]any{{"Issue Key": "EXAMPLE-1"}, {"Issue Key": ""}, {"Issue Key": "EXAMPLE-2"}}, path)).To(Succeed())
				} else {
					writeSyncWorkbook(path, []string{"EXAMPLE-1", "", "EXAMPLE-2"})
				}
				args = []string{"sync", "@" + path + ":" + selector, "--sync-dir", dir, "--no-cache"}
			}
			output, err := commandOutput(args, "json")
			Expect(err).NotTo(HaveOccurred())
			var results []struct {
				Key  string `json:"key"`
				Path string `json:"path"`
			}
			Expect(json.Unmarshal([]byte(output), &results)).To(Succeed())
			Expect(keys).To(Equal([]string{"EXAMPLE-1", "EXAMPLE-2"}))
			Expect(results).To(HaveLen(2))
			Expect(catalogs).To(Equal(1))
			Expect(identities).To(Equal(1))
			for _, result := range results {
				Expect(result.Path).To(Equal(filepath.Join(dir, "10001", result.Key+".md")))
				Expect(result.Path).To(BeAnExistingFile())
			}
		}
	})

	It("registers sync with the default directory and attachment/extraction flags", func() {
		command, _, err := commandRoot().Find([]string{"external-issues", "jira", "sync"})
		Expect(err).NotTo(HaveOccurred())
		Expect(command.Name()).To(Equal("sync"))
		Expect(command.Flags().Lookup("sync-dir").Usage).To(ContainSubstring(".gavel/jira"))
		Expect(command.Flags().Lookup("attachments")).NotTo(BeNil())
		Expect(command.Flags().Lookup("extract")).NotTo(BeNil())
	})

	It("syncs an issue through the authenticated shared Jira client", func() {
		commandServer(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/rest/api/3/issue/EXAMPLE-1":
				fmt.Fprint(w, `{"key":"EXAMPLE-1","fields":{"summary":"Sync example","project":{"id":"10001"},"attachment":[{"id":"20001","filename":"notes.txt","content":"https://example.test/notes.txt"}]}}`)
			case "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case "/rest/api/3/field", "/rest/api/3/issue/EXAMPLE-1/remotelink":
				fmt.Fprint(w, `[]`)
			case "/rest/api/3/issue/EXAMPLE-1/comment":
				fmt.Fprint(w, `{"total":0,"comments":[]}`)
			case "/rest/api/3/attachment/content/20001":
				fmt.Fprint(w, "Local notes")
			default:
				http.NotFound(w, r)
			}
		})
		dir := GinkgoT().TempDir()
		for _, flag := range []string{"--attachments", "--extract"} {
			output, err := commandOutput([]string{"sync", "EXAMPLE-1", "--sync-dir", dir, flag}, "json")
			Expect(err).NotTo(HaveOccurred())
			var result []struct {
				Path string `json:"path"`
			}
			Expect(json.Unmarshal([]byte(output), &result)).To(Succeed())
			Expect(result).To(HaveLen(1))
			Expect(result[0].Path).To(Equal(filepath.Join(dir, "10001", "EXAMPLE-1.md")))
			data, err := os.ReadFile(result[0].Path)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(ContainSubstring("[notes.txt](EXAMPLE-1.attachments/20001/notes.txt)"))
		}
	})
})
