package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/commons/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type issueOutputOptions struct {
	Response string
	Args     []string
	Format   string
}

func issueCommandOutput(options issueOutputOptions) (string, error) {
	commandServer(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/rest/api/3/myself":
			fmt.Fprint(w, `{"accountId":"account"}`)
		case r.URL.Path == "/rest/api/3/field":
			fmt.Fprint(w, `[]`)
		case strings.HasSuffix(r.URL.Path, "/comment"):
			fmt.Fprint(w, `{"total":0,"comments":[]}`)
		case strings.HasSuffix(r.URL.Path, "/remotelink"):
			fmt.Fprint(w, `[]`)
		default:
			fmt.Fprint(w, options.Response)
		}
	})
	return commandOutput(options.Args, options.Format)
}

var _ = Describe("Jira output boundary", func() {
	const unsupported = `{"key":"EXAMPLE-1","fields":{"summary":"Future document","description":{"type":"doc","version":1,"content":[{"type":"unknown"}]}}}`
	DescribeTable("keeps description exports independent of rich rendering", func(format string) {
		output, err := issueCommandOutput(issueOutputOptions{Response: unsupported, Args: []string{"get", "EXAMPLE-1", "--full"}, Format: format})
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring("Future document"))
	}, Entry("JSON", "json"), Entry("YAML", "yaml"))
	It("writes Markdown and logs contextual ADF warnings", func() {
		var warnings bytes.Buffer
		previous := logger.GetOutput()
		logger.SetOutput(&warnings)
		DeferCleanup(func() { logger.SetOutput(previous) })
		output, err := issueCommandOutput(issueOutputOptions{Response: unsupported, Args: []string{"get", "EXAMPLE-1", "--full"}, Format: "markdown"})
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(SatisfyAll(ContainSubstring("Future document"), ContainSubstring(`"type":"unknown"`)))
		Expect(warnings.String()).To(SatisfyAll(ContainSubstring("WRN"), ContainSubstring("EXAMPLE-1"), ContainSubstring("description"), ContainSubstring("unsupported node")))
	})
	It("exports custom-field documents as Markdown strings even without catalog metadata", func() {
		response := `{"key":"EXAMPLE-1","fields":{"summary":"Instructions","customfield_unknown":{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Read this"}]}]}}}`
		output, err := issueCommandOutput(issueOutputOptions{Response: response, Args: []string{"get", "EXAMPLE-1", "--no-cache"}, Format: "json"})
		Expect(err).NotTo(HaveOccurred())
		var result struct {
			Fields struct {
				Unknown string `json:"customfield_unknown"`
			} `json:"fields"`
		}
		Expect(json.Unmarshal([]byte(output), &result)).To(Succeed())
		Expect(result.Fields.Unknown).To(Equal("Read this"))
	})
	It("writes a compact Markdown document through the output sink", func() {
		response := `{"key":"EXAMPLE-42","fields":{"summary":"Retry imports","project":{"name":"Platform"},"status":{"name":"In Progress"},"assignee":{"displayName":"Alex"},"priority":{"name":"High"},"labels":["imports"],"description":{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Retry transient failures."}]}]},"attachment":[]}}`
		output, err := issueCommandOutput(issueOutputOptions{Response: response, Args: []string{"get", "EXAMPLE-42"}, Format: "markdown"})
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(SatisfyAll(ContainSubstring("EXAMPLE-42 — Retry imports"), ContainSubstring("In Progress · High · Alex · Platform"), ContainSubstring("Retry transient failures."), ContainSubstring("0 comments · 0 attachments")))
	})
	It("keeps list bodies exportable and propagates the resolved site", func() {
		response := `{"issues":[` + unsupported + `]}`
		output, err := issueCommandOutput(issueOutputOptions{Response: response, Args: []string{"list", "--jql", "project = EXAMPLE"}, Format: "json"})
		Expect(err).NotTo(HaveOccurred())
		var rows []map[string]any
		Expect(json.Unmarshal([]byte(output), &rows)).To(Succeed())
		Expect(rows).To(HaveLen(1))
		Expect(rows[0]["site"]).To(HavePrefix("http://127.0.0.1:"))
	})
})
