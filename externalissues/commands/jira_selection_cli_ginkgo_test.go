//go:build cli_smoke

package commands

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira sync selection compiled CLI", func() {
	It("selects on ignored fields with the real global flag wiring", func() {
		binary := os.Getenv("GAVEL_TEST_BINARY")
		Expect(binary).NotTo(BeEmpty(), "set GAVEL_TEST_BINARY to make build output")
		workDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(workDir, ".gavel.yaml"), []byte("externalIssues: {jira: {ignoreFields: [priority, comment, remoteLinks]}}\n"), 0o600)).To(Succeed())
		commandServer(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasSuffix(r.URL.Path, "/myself"):
				fmt.Fprint(w, `{"accountId":"account"}`)
			case strings.HasSuffix(r.URL.Path, "/field"):
				fmt.Fprint(w, `[]`)
			case strings.HasSuffix(r.URL.Path, "/search/jql"):
				fmt.Fprint(w, `{"issues":[{"key":"APP-1"}],"isLast":true}`)
			case strings.HasSuffix(r.URL.Path, "/issue/APP-1"):
				fmt.Fprint(w, `{"key":"APP-1","fields":{"summary":"CLI selection","project":{"id":"1","key":"APP"},"priority":{"name":"High"},"status":{"name":"Open"}}}`)
			default:
				http.Error(w, "excluded enrichment requested", 500)
			}
		})
		command := exec.Command(binary, "--cwd", workDir, "external-issues", "jira", "sync", "APP", "--status", "!Done,!Closed", "--filter", "Priority=High,Medium", "--sync-dir", filepath.Join(workDir, "issues"), "--json", "--no-progress")
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		Expect(command.Run()).To(Succeed(), stderr.String())
		Expect(stdout.String()).To(ContainSubstring(`"key": "APP-1"`))
		data, err := os.ReadFile(filepath.Join(workDir, "issues", syncTestDomain(), "APP", "APP-1.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).NotTo(ContainSubstring("High"))
	})
})
