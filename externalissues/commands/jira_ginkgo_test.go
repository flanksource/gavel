package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/clicky"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

func TestCommands(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "External Jira Commands") }

func commandServer(handler http.HandlerFunc) *httptest.Server {
	for _, key := range []string{"JIRA_OAUTH_ACCESS_TOKEN", "JIRA_CLOUD_ID", "ATLASSIAN_CLOUD_ID"} {
		GinkgoT().Setenv(key, "")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer GinkgoRecover()
		user, token, ok := r.BasicAuth()
		Expect(ok).To(BeTrue())
		Expect(user).To(Equal("api@example.test"))
		Expect(token).To(Equal("command-token"))
		handler(w, r)
	}))
	DeferCleanup(server.Close)
	GinkgoT().Setenv("JIRA_URL", server.URL)
	GinkgoT().Setenv("JIRA_EMAIL", "api@example.test")
	GinkgoT().Setenv("JIRA_API_TOKEN", "command-token")
	GinkgoT().Setenv("GAVEL_JIRA_CACHE_DIR", GinkgoT().TempDir())
	return server
}

func commandRoot() *cobra.Command {
	clicky.Flags.FormatOptions = clicky.FormatOptions{}
	clicky.UseFormatter(clicky.FormatOptions{})
	root := &cobra.Command{Use: "gavel", SilenceErrors: true, SilenceUsage: true}
	clicky.BindAllFlagsToCommand(root, "format", "tasks")
	root.PersistentPreRun = func(*cobra.Command, []string) { clicky.Flags.UseFlags() }
	Register(root)
	return root
}

func commandOutput(args []string, format string) (string, error) {
	output := filepath.Join(GinkgoT().TempDir(), "result")
	root := commandRoot()
	root.SetArgs(append(append([]string{"external-issues", "jira"}, args...), "--format", format+"="+output))
	if err := root.Execute(); err != nil {
		return "", err
	}
	data, err := os.ReadFile(output)
	return string(data), err
}

var _ = Describe("direct Jira commands", func() {
	It("registers issue, comment and attachment operations", func() {
		root := commandRoot()
		for _, args := range [][]string{{"get"}, {"list"}, {"comment", "list"}, {"comment", "create"}, {"attachment", "list"}, {"attachment", "upload"}, {"attachment", "download"}, {"attachment", "delete"}} {
			command, _, err := root.Find(append([]string{"external-issues", "jira"}, args...))
			Expect(err).NotTo(HaveOccurred())
			Expect(command.Name()).To(Equal(args[len(args)-1]))
		}
	})

	It("shares a site catalog across list rows and replaces it with no-cache", func() {
		calls := 0
		commandServer(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/rest/api/3/search/jql":
				var body map[string]any
				Expect(json.NewDecoder(r.Body).Decode(&body)).To(Succeed())
				Expect(body["jql"]).To(Equal("project = EXAMPLE"))
				Expect(body["fields"]).To(ContainElement("customfield_1"))
				fmt.Fprint(w, `{"issues":[{"key":"EXAMPLE-1","fields":{"summary":"First","project":{"id":"1"},"customfield_1":"One"}},{"key":"EXAMPLE-2","fields":{"summary":"Second","project":{"id":"2"},"customfield_1":"Two"}}]}`)
			case "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case "/rest/api/3/field":
				calls++
				if calls == 1 {
					fmt.Fprint(w, `[{"id":"customfield_1","name":"Before"}]`)
				} else {
					fmt.Fprint(w, `[{"id":"customfield_1","name":"After"}]`)
				}
			default:
				http.NotFound(w, r)
			}
		})
		for _, noCache := range []bool{false, false, true} {
			args := []string{"list", "--jql", "project = EXAMPLE", "--fields", "customfield_1"}
			if noCache {
				args = append(args, "--no-cache")
			}
			output, err := commandOutput(args, "json")
			Expect(err).NotTo(HaveOccurred())
			var rows []struct {
				Fields map[string]string `json:"fields"`
			}
			Expect(json.Unmarshal([]byte(output), &rows)).To(Succeed())
			Expect(rows).To(HaveLen(2))
			name := "before"
			if noCache {
				name = "after"
			}
			Expect(rows[0].Fields).To(HaveKeyWithValue(name, "One"))
			Expect(rows[1].Fields).To(HaveKeyWithValue(name, "Two"))
		}
		Expect(calls).To(Equal(2))
	})

	It("converts comment Markdown to ADF before posting", func() {
		commandServer(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.Method).To(Equal("POST"))
			Expect(r.URL.Path).To(Equal("/rest/api/3/issue/EXAMPLE-1/comment"))
			var body map[string]json.RawMessage
			Expect(json.NewDecoder(r.Body).Decode(&body)).To(Succeed())
			Expect(string(body["body"])).To(MatchJSON(`{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Hello","marks":[{"type":"strong"}]}]}]}`))
			fmt.Fprint(w, `{"id":"comment-1","body":"Hello"}`)
		})
		output, err := commandOutput([]string{"comment", "create", "EXAMPLE-1", "--body", "**Hello**"}, "json")
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring("comment-1"))
	})

	It("uploads multipart files and downloads bytes without overwriting existing files", func() {
		dir := GinkgoT().TempDir()
		input := filepath.Join(dir, "report.txt")
		Expect(os.WriteFile(input, []byte("report body"), 0o600)).To(Succeed())
		commandServer(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				Expect(r.Header.Get("X-Atlassian-Token")).To(Equal("no-check"))
				reader, err := r.MultipartReader()
				Expect(err).NotTo(HaveOccurred())
				part, err := reader.NextPart()
				Expect(err).NotTo(HaveOccurred())
				Expect(part.FileName()).To(Equal("report.txt"))
				data, err := io.ReadAll(part)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal("report body"))
				fmt.Fprint(w, `[{"id":"file-1","filename":"report.txt","content":"https://example.test/file"}]`)
			} else {
				Expect(r.URL.Query().Get("redirect")).To(Equal("false"))
				fmt.Fprint(w, "report body")
			}
		})
		_, err := commandOutput([]string{"attachment", "upload", "EXAMPLE-1", "--file", input}, "json")
		Expect(err).NotTo(HaveOccurred())
		output := filepath.Join(dir, "download.txt")
		root := commandRoot()
		root.SetArgs([]string{"external-issues", "jira", "attachment", "download", "file-1", "--output", output})
		Expect(root.Execute()).To(Succeed())
		data, err := os.ReadFile(output)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal("report body"))
		root = commandRoot()
		root.SetArgs([]string{"external-issues", "jira", "attachment", "download", "file-1", "--output", output})
		Expect(root.Execute()).To(HaveOccurred())
	})

	It("requires confirmation before an attachment deletion", func() {
		root := commandRoot()
		root.SetArgs([]string{"external-issues", "jira", "attachment", "delete", "file-1"})
		Expect(root.Execute()).To(MatchError(ContainSubstring("--yes")))
	})

	It("deletes a confirmed attachment directly", func() {
		commandServer(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.Method).To(Equal("DELETE"))
			Expect(r.URL.Path).To(Equal("/rest/api/3/attachment/file-1"))
			w.WriteHeader(http.StatusNoContent)
		})
		root := commandRoot()
		root.SetArgs([]string{"external-issues", "jira", "attachment", "delete", "file-1", "--yes"})
		Expect(root.Execute()).To(Succeed())
	})
})
