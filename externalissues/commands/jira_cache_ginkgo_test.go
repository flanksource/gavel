package commands

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Jira cached get command", func() {
	It("defaults to full output and reuses synced data across output formats", func() {
		reads := 0
		commandServer(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/rest/api/3/myself":
				fmt.Fprint(w, `{"accountId":"account"}`)
			case r.URL.Path == "/rest/api/3/field", strings.HasSuffix(r.URL.Path, "/remotelink"):
				fmt.Fprint(w, `[]`)
			case strings.HasSuffix(r.URL.Path, "/comment"):
				fmt.Fprint(w, `{"total":1,"comments":[{"id":"c1","body":"Cached comment body"}]}`)
			default:
				reads++
				fmt.Fprint(w, `{"id":"123","key":"EXAMPLE-1","fields":{"summary":"Cached issue","project":{"id":"10001","key":"EXAMPLE"},"description":"Complete description"}}`)
			}
		})
		run := func(args []string, format string) string {
			if binary := os.Getenv("GAVEL_TEST_BINARY"); binary != "" {
				output, err := exec.Command(binary, append(append([]string{"external-issues", "jira"}, args...), "--format", format, "--no-progress")...).Output()
				Expect(err).NotTo(HaveOccurred())
				return string(output)
			}
			output, err := commandOutput(args, format)
			Expect(err).NotTo(HaveOccurred())
			return output
		}
		Expect(run([]string{"get", "EXAMPLE-1"}, "markdown")).To(ContainSubstring("Cached comment body"))
		Expect(run([]string{"get", "123", "--full=false"}, "markdown")).NotTo(ContainSubstring("Cached comment body"))
		Expect(run([]string{"get", "123"}, "json")).To(ContainSubstring(`"key": "EXAMPLE-1"`))
		Expect(reads).To(Equal(1))
		Expect(filepath.Join(os.Getenv("XDG_CACHE_HOME"), ".gavel", "jira", syncTestDomain(), "EXAMPLE", "EXAMPLE-1.md")).To(BeAnExistingFile())
		run([]string{"get", "EXAMPLE-1", "--ttl=0"}, "json")
		Expect(reads).To(Equal(2))
	})
})
