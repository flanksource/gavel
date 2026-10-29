//go:build cli_smoke

package commands

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira CLI configured exclusions", func() {
	It("honors --cwd for get, list and sync when launched elsewhere", func() {
		binary := os.Getenv("GAVEL_TEST_BINARY")
		Expect(binary).NotTo(BeEmpty(), "set GAVEL_TEST_BINARY to the binary produced by make build")
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		workDir := GinkgoT().TempDir()
		config := "externalIssues: {jira: {ignoreFields: [Internal*, comment, remoteLinks, attachment]}}\n"
		Expect(os.WriteFile(filepath.Join(workDir, ".gavel.yaml"), []byte(config), 0o600)).To(Succeed())
		commandServer(func(w http.ResponseWriter, r *http.Request) {
			const issue = `{"key":"EXAMPLE-1","fields":{"summary":"Visible title","project":{"id":"10001"},"customfield_1":"Hidden value","customfield_2":"Visible value"}}`
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
		for _, operation := range [][]string{
			{"get", "EXAMPLE-1"},
			{"list", "--jql", "project = EXAMPLE", "--fields", "customfield_1,customfield_2"},
			{"sync", "EXAMPLE-1", "--sync-dir", filepath.Join(workDir, "issues"), "--extract"},
		} {
			args := append([]string{"--cwd", workDir, "external-issues", "jira"}, operation...)
			command := exec.Command(binary, append(args, "--json", "--no-progress")...)
			command.Dir = GinkgoT().TempDir()
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			Expect(command.Run()).To(Succeed(), stderr.String())
			output := stdout.String()
			if operation[0] == "sync" {
				data, err := os.ReadFile(filepath.Join(workDir, "issues", "10001", "EXAMPLE-1.md"))
				Expect(err).NotTo(HaveOccurred())
				output = string(data)
			}
			Expect(output).To(SatisfyAll(ContainSubstring("Visible title"), ContainSubstring("Visible value"), Not(ContainSubstring("Hidden value"))))
		}
	})
})
