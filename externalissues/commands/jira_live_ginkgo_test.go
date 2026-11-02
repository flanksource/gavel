//go:build jira_live

package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/externalissues/jira"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("live Jira command", func() {
	It("syncs issue keys and native IDs expanded from the second Excel worksheet", func() {
		key := os.Getenv("GAVEL_JIRA_LIVE_ISSUE")
		Expect(key).NotTo(BeEmpty(), "GAVEL_JIRA_LIVE_ISSUE is required")
		GinkgoT().Setenv("GAVEL_JIRA_CACHE_DIR", GinkgoT().TempDir())
		data, err := commandOutput([]string{"get", key}, "json")
		Expect(err).NotTo(HaveOccurred())
		var issue struct {
			ID string `json:"id"`
		}
		Expect(json.Unmarshal([]byte(data), &issue)).To(Succeed())
		Expect(issue.ID).NotTo(BeEmpty())
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "issues.xlsx")
		writeSyncWorkbook(path, []string{key, issue.ID})
		data, err = commandOutput([]string{"sync", "@" + path + ":'Sprint Backlog'!Issue Key", "--sync-dir", dir, "--attachments", "--no-cache"}, "json")
		Expect(err).NotTo(HaveOccurred())
		var results jira.SyncResults
		Expect(json.Unmarshal([]byte(data), &results)).To(Succeed())
		Expect(results).To(HaveLen(2))
		for _, result := range results {
			Expect(result.Key).To(Equal(key))
			Expect(result.Path).To(BeAnExistingFile())
			for _, attachment := range result.Attachments {
				Expect(attachment).To(BeAnExistingFile())
			}
		}
	})

	It("exports a full issue as Markdown", func() {
		id := os.Getenv("GAVEL_JIRA_LIVE_ISSUE")
		Expect(id).NotTo(BeEmpty(), "GAVEL_JIRA_LIVE_ISSUE is required")
		cacheDir := GinkgoT().TempDir()
		GinkgoT().Setenv("GAVEL_JIRA_CACHE_DIR", cacheDir)
		root := &cobra.Command{Use: "gavel", SilenceErrors: true, SilenceUsage: true}
		clicky.BindAllFlagsToCommand(root, "format", "tasks")
		root.PersistentPreRun = func(*cobra.Command, []string) { clicky.Flags.UseFlags() }
		Register(root)
		output := filepath.Join(GinkgoT().TempDir(), "issue.md")
		root.SetArgs([]string{"external-issues", "jira", "get", id, "--full", "--format", "markdown=" + output})
		Expect(root.Execute()).To(Succeed())
		data, err := os.ReadFile(output)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(ContainSubstring(id))
		Expect(string(data)).To(ContainSubstring("Description"))
		Expect(string(data)).To(ContainSubstring("Attachments"))
		entries, err := os.ReadDir(cacheDir)
		Expect(err).NotTo(HaveOccurred())
		var cachePath string
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".json") {
				cachePath = filepath.Join(cacheDir, entry.Name())
			}
		}
		Expect(cachePath).NotTo(BeEmpty())
		before, err := os.ReadFile(cachePath)
		Expect(err).NotTo(HaveOccurred())
		clicky.Flags.FormatOptions = clicky.FormatOptions{}
		clicky.UseFormatter(clicky.FormatOptions{})
		root = &cobra.Command{Use: "gavel", SilenceErrors: true, SilenceUsage: true}
		clicky.BindAllFlagsToCommand(root, "format", "tasks")
		root.PersistentPreRun = func(*cobra.Command, []string) { clicky.Flags.UseFlags() }
		Register(root)
		jsonPath := filepath.Join(GinkgoT().TempDir(), "issue.json")
		root.SetArgs([]string{"external-issues", "jira", "get", id, "--format", "json=" + jsonPath})
		Expect(root.Execute()).To(Succeed())
		data, err = os.ReadFile(jsonPath)
		Expect(err).NotTo(HaveOccurred())
		var result struct {
			Fields map[string]json.RawMessage `json:"fields"`
		}
		Expect(json.Unmarshal(data, &result)).To(Succeed())
		Expect(result.Fields).NotTo(BeEmpty())
		for key := range result.Fields {
			Expect(key).NotTo(HavePrefix("customfield_"))
		}
		var rootFields map[string]json.RawMessage
		Expect(json.Unmarshal(data, &rootFields)).To(Succeed())
		for _, key := range []string{"native", "schema", "source", "scheme"} {
			Expect(rootFields).NotTo(HaveKey(key))
		}
		GinkgoWriter.Printf("exported %d named fields without unresolved custom IDs\n", len(result.Fields))
		after, err := os.ReadFile(cachePath)
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(Equal(before))
	})

	It("maps fields in a Jira list result", func() {
		id := os.Getenv("GAVEL_JIRA_LIVE_ISSUE")
		Expect(id).NotTo(BeEmpty(), "GAVEL_JIRA_LIVE_ISSUE is required")
		GinkgoT().Setenv("GAVEL_JIRA_CACHE_DIR", GinkgoT().TempDir())
		clicky.Flags.FormatOptions = clicky.FormatOptions{}
		clicky.UseFormatter(clicky.FormatOptions{})
		root := &cobra.Command{Use: "gavel", SilenceErrors: true, SilenceUsage: true}
		clicky.BindAllFlagsToCommand(root, "format", "tasks")
		root.PersistentPreRun = func(*cobra.Command, []string) { clicky.Flags.UseFlags() }
		Register(root)
		output := filepath.Join(GinkgoT().TempDir(), "list.json")
		root.SetArgs([]string{"external-issues", "jira", "list", "--jql", "key = " + id, "--limit", "1", "--fields", "customfield_10053", "--format", "json=" + output})
		Expect(root.Execute()).To(Succeed())
		data, err := os.ReadFile(output)
		Expect(err).NotTo(HaveOccurred())
		var rows []struct {
			Key    string                     `json:"key"`
			Fields map[string]json.RawMessage `json:"fields"`
		}
		Expect(json.Unmarshal(data, &rows)).To(Succeed())
		Expect(rows).To(HaveLen(1))
		Expect(rows[0].Key).To(Equal(id))
		Expect(rows[0].Fields).NotTo(BeEmpty())
		Expect(rows[0].Fields).To(HaveLen(1))
		for key := range rows[0].Fields {
			Expect(key).NotTo(HavePrefix("customfield_"))
		}
	})
})
