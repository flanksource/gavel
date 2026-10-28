//go:build jira_live

package commands

import (
	"os"
	"path/filepath"

	"github.com/flanksource/clicky"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("live Jira command", func() {
	It("exports a full issue as Markdown", func() {
		id := os.Getenv("GAVEL_JIRA_LIVE_ISSUE")
		Expect(id).NotTo(BeEmpty(), "GAVEL_JIRA_LIVE_ISSUE is required")
		root := &cobra.Command{Use: "gavel", SilenceErrors: true, SilenceUsage: true}
		clicky.BindAllFlagsToCommand(root, "format", "tasks")
		root.PersistentPreRun = func(*cobra.Command, []string) { clicky.Flags.UseFlags() }
		Register(root)
		output := filepath.Join(GinkgoT().TempDir(), "issue.md")
		root.SetArgs([]string{"external-issues", "jira", "get", id, "--format", "markdown=" + output})
		Expect(root.Execute()).To(Succeed())
		data, err := os.ReadFile(output)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring(id))
		Expect(string(data)).To(ContainSubstring("Description"))
		Expect(string(data)).To(ContainSubstring("Comments"))
		Expect(string(data)).To(ContainSubstring("Attachments"))
	})
})
