//go:build cli_smoke

package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira CLI task progress", func() {
	It("keeps JSON on stdout and flushes task progress to stderr on success and failure", func() {
		binary := os.Getenv("GAVEL_TEST_BINARY")
		Expect(binary).NotTo(BeEmpty(), "set GAVEL_TEST_BINARY to the binary produced by make build")
		commandServer(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/rest/api/3/issue/EXAMPLE-1":
				fmt.Fprint(w, `{"key":"EXAMPLE-1","fields":{"summary":"Example","project":{"id":"10001"}}}`)
			case "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case "/rest/api/3/field", "/rest/api/3/issue/EXAMPLE-1/remotelink":
				fmt.Fprint(w, `[]`)
			case "/rest/api/3/issue/EXAMPLE-1/comment":
				fmt.Fprint(w, `{"total":0,"comments":[]}`)
			default:
				http.NotFound(w, r)
			}
		})
		for _, keys := range [][]string{{"EXAMPLE-1"}, {"EXAMPLE-1", "EXAMPLE-2"}} {
			dir := GinkgoT().TempDir()
			args := append([]string{"external-issues", "jira", "sync"}, keys...)
			args = append(args, "--sync-dir", dir, "--json", "--no-progress")
			command := exec.Command(binary, args...)
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			err := command.Run()
			if len(keys) == 1 {
				Expect(err).NotTo(HaveOccurred(), stderr.String())
			} else {
				Expect(err).To(HaveOccurred())
				Expect(stderr.String()).To(ContainSubstring("EXAMPLE-2 (2/2)"))
			}
			var results []struct {
				Key  string `json:"key"`
				Path string `json:"path"`
			}
			Expect(json.Unmarshal(stdout.Bytes(), &results)).To(Succeed(), stdout.String())
			Expect(results).To(HaveLen(1))
			Expect(results[0].Path).To(Equal(filepath.Join(dir, "10001", "EXAMPLE-1.md")))
			Expect(stderr.String()).To(ContainSubstring("Sync Jira issues"))
			Expect(stderr.String()).To(ContainSubstring("Sync EXAMPLE-1"))
			Expect(strings.Contains(stdout.String(), "Sync Jira issues")).To(BeFalse())
		}
	})
})
