package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/clicky"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

func TestMain(m *testing.M) {
	if filepath.Base(os.Args[0]) == "twg" {
		data, err := json.Marshal(os.Args[1:])
		if err != nil || os.WriteFile(os.Getenv("GAVEL_TEST_TWG_ARGS"), data, 0o600) != nil {
			os.Exit(1)
		}
		response := os.Getenv("GAVEL_TEST_TWG_RESPONSE")
		if response == "" {
			response = `{"data":[{"key":"EXAMPLE-1","summary":"Example","description":{"type":"doc","version":1,"content":[]},"comments":[]}]}`
		}
		_, _ = os.Stdout.WriteString(response)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestCommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "External Jira Commands")
}

var _ = Describe("Jira commands", func() {
	BeforeEach(func() {
		clicky.Flags.FormatOptions = clicky.FormatOptions{}
		clicky.UseFormatter(clicky.FormatOptions{})
	})
	It("registers get, list, comments and attachment operations", func() {
		root := &cobra.Command{Use: "gavel", SilenceErrors: true, SilenceUsage: true}
		Register(root)
		for _, args := range [][]string{
			{"external-issues", "jira", "get"}, {"external-issues", "jira", "list"},
			{"external-issues", "jira", "comment", "list"}, {"external-issues", "jira", "comment", "create"},
			{"external-issues", "jira", "attachment", "list"}, {"external-issues", "jira", "attachment", "upload"},
			{"external-issues", "jira", "attachment", "download"}, {"external-issues", "jira", "attachment", "delete"},
		} {
			cmd, _, err := root.Find(args)
			Expect(err).NotTo(HaveOccurred())
			Expect(cmd.Name()).To(Equal(args[len(args)-1]))
		}
	})

	It("passes get options to twg without a shell", func() {
		dir := GinkgoT().TempDir()
		binary, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Symlink(binary, filepath.Join(dir, "twg"))).To(Succeed())
		GinkgoT().Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
		path := filepath.Join(dir, "args.json")
		GinkgoT().Setenv("GAVEL_TEST_TWG_ARGS", path)
		root := &cobra.Command{Use: "gavel", SilenceErrors: true, SilenceUsage: true}
		clicky.BindAllFlagsToCommand(root, "format", "tasks")
		Register(root)
		root.SetArgs([]string{"external-issues", "jira", "--site", "example", "get", "EXAMPLE-1", "--json"})
		Expect(root.Execute()).To(Succeed())
		data, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		var args []string
		Expect(json.Unmarshal(data, &args)).To(Succeed())
		Expect(args).To(ContainElements("jira", "workitem", "get", "EXAMPLE-1", "--full", "--site", "example", "--output-summary", "none"))
	})

	It("writes complete issue Markdown to a Clicky output sink", func() {
		dir := GinkgoT().TempDir()
		binary, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Symlink(binary, filepath.Join(dir, "twg"))).To(Succeed())
		GinkgoT().Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
		GinkgoT().Setenv("GAVEL_TEST_TWG_ARGS", filepath.Join(dir, "args.json"))
		GinkgoT().Setenv("GAVEL_TEST_TWG_RESPONSE", `{"data":[{"key":"EXAMPLE-1","summary":"Example","url":"https://example.com/EXAMPLE-1","description":{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Description text"}]}]},"comments":[{"author":{"displayName":"Alex"},"created":"2026-01-01","body":{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Comment text"}]}]}}],"attachment":[{"filename":"report.pdf","content":"https://example.com/report.pdf"}]}]}`)
		root := &cobra.Command{Use: "gavel", SilenceErrors: true, SilenceUsage: true}
		clicky.BindAllFlagsToCommand(root, "format", "tasks")
		root.PersistentPreRun = func(*cobra.Command, []string) { clicky.Flags.UseFlags() }
		Register(root)
		output := filepath.Join(dir, "issue.md")
		root.SetArgs([]string{"external-issues", "jira", "get", "EXAMPLE-1", "--format", "markdown=" + output})
		Expect(root.Execute()).To(Succeed())
		data, err := os.ReadFile(output)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring("Description text"))
		Expect(string(data)).To(ContainSubstring("Comment text"))
		Expect(string(data)).To(ContainSubstring("report.pdf"))
	})

	It("requires explicit confirmation before deleting an attachment", func() {
		root := &cobra.Command{Use: "gavel", SilenceErrors: true, SilenceUsage: true}
		Register(root)
		root.SetArgs([]string{"external-issues", "jira", "attachment", "delete", "123"})
		Expect(root.Execute()).To(MatchError(ContainSubstring("--yes")))
	})

	DescribeTable("routes the requested Jira operation", func(command []string, response string, expected []string) {
		dir := GinkgoT().TempDir()
		binary, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Symlink(binary, filepath.Join(dir, "twg"))).To(Succeed())
		GinkgoT().Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
		path := filepath.Join(dir, "args.json")
		GinkgoT().Setenv("GAVEL_TEST_TWG_ARGS", path)
		GinkgoT().Setenv("GAVEL_TEST_TWG_RESPONSE", response)
		root := &cobra.Command{Use: "gavel", SilenceErrors: true, SilenceUsage: true}
		clicky.BindAllFlagsToCommand(root, "format", "tasks")
		Register(root)
		root.SetArgs(append([]string{"external-issues", "jira"}, command...))
		Expect(root.Execute()).To(Succeed())
		data, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		var args []string
		Expect(json.Unmarshal(data, &args)).To(Succeed())
		Expect(args).To(ContainElements(expected))
	},
		Entry("list", []string{"list", "--jql", "project = EXAMPLE", "--limit", "3"}, `{"data":{"issues":[]}}`, []string{"query", "--jql", "project = EXAMPLE", "--limit", "3"}),
		Entry("comments", []string{"comment", "list", "EXAMPLE-1"}, `{"data":[{"key":"EXAMPLE-1","comments":[]}]}`, []string{"get", "EXAMPLE-1", "--full"}),
		Entry("create comment", []string{"comment", "create", "EXAMPLE-1", "--body", "Hello"}, `{"data":{"id":"1"}}`, []string{"comment", "create", "--issue-id", "EXAMPLE-1", "--body", "Hello", "--body-format", "markdown"}),
		Entry("list attachments", []string{"attachment", "list", "EXAMPLE-1"}, `{"data":[]}`, []string{"attachment", "query", "--issue-id", "EXAMPLE-1"}),
		Entry("upload", []string{"attachment", "upload", "EXAMPLE-1", "--file", "report.txt"}, `{"data":[]}`, []string{"attachment", "upload", "--issue-id", "EXAMPLE-1", "--file", "report.txt"}),
		Entry("download", []string{"attachment", "download", "123", "--output", "report.txt"}, `{"data":{}}`, []string{"attachment", "download", "--attachment-id", "123", "--output-path", "report.txt"}),
		Entry("delete", []string{"attachment", "delete", "123", "--yes"}, `{"data":{}}`, []string{"attachment", "delete", "--attachment-id", "123", "--yes"}),
	)
})
