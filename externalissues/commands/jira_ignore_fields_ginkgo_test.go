package commands

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/workspace"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira command field exclusions", func() {
	var workDir string

	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		workDir = GinkgoT().TempDir()
		config := "externalIssues:\n  jira:\n    ignoreFields: [Internal*, description, comment, attachment, remoteLinks]\n"
		Expect(os.WriteFile(filepath.Join(workDir, ".gavel.yaml"), []byte(config), 0o600)).To(Succeed())
		commandServer(func(w http.ResponseWriter, r *http.Request) {
			const issue = `{"id":"1","key":"EXAMPLE-1","fields":{"summary":"Visible title","project":{"id":"10001"},"description":"Hidden description","customfield_1":"Hidden value","customfield_2":"Visible value","attachment":[{"id":"2","filename":"hidden.pdf"}]}}`
			switch r.URL.Path {
			case "/rest/api/3/issue/EXAMPLE-1":
				fmt.Fprint(w, issue)
			case "/rest/api/3/search/jql":
				fmt.Fprint(w, `{"issues":[`+issue+`]}`)
			case "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case "/rest/api/3/field":
				fmt.Fprint(w, `[{"id":"customfield_1","name":"Internal notes"},{"id":"customfield_2","name":"Public notes"}]`)
			default:
				http.Error(w, "excluded content must not be requested", http.StatusInternalServerError)
			}
		})
	})

	DescribeTable("uses the invocation directory for all issue output formats",
		func(format string) {
			root := commandRoot()
			root.SetContext(workspace.With(context.Background(), workspace.Dirs{Invocation: workDir}))
			output := filepath.Join(GinkgoT().TempDir(), "issue")
			root.SetArgs([]string{"external-issues", "jira", "get", "EXAMPLE-1", "--full", "--format", format + "=" + output})
			Expect(root.Execute()).To(Succeed())
			data, err := os.ReadFile(output)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(SatisfyAll(ContainSubstring("Visible title"), ContainSubstring("Visible value"), Not(ContainSubstring("Hidden")), Not(ContainSubstring("hidden.pdf"))))
		},
		Entry("JSON", "json"), Entry("YAML", "yaml"), Entry("Markdown", "markdown"), Entry("HTML", "html"), Entry("terminal", "pretty"),
	)

	It("applies configured exclusions to requested list fields and synced documents", func() {
		for _, args := range [][]string{
			{"list", "--jql", "project = EXAMPLE", "--fields", "customfield_1,customfield_2"},
			{"sync", "EXAMPLE-1", "--sync-dir", filepath.Join(workDir, "issues"), "--extract"},
		} {
			root := commandRoot()
			root.SetContext(workspace.With(context.Background(), workspace.Dirs{Invocation: workDir}))
			output := filepath.Join(GinkgoT().TempDir(), "result.json")
			root.SetArgs(append(append([]string{"external-issues", "jira"}, args...), "--format", "json="+output))
			Expect(root.Execute()).To(Succeed())
			if args[0] == "sync" {
				output = filepath.Join(workDir, "issues", "10001", "EXAMPLE-1.md")
			}
			data, err := os.ReadFile(output)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(SatisfyAll(ContainSubstring("Visible title"), ContainSubstring("Visible value"), Not(ContainSubstring("Hidden")), Not(ContainSubstring("hidden.pdf"))))
		}
	})
})
