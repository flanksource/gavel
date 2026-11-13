package verify

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("External issues configuration", func() {
	It("accumulates unique field patterns across home, repository and nested layers", func() {
		home, repo := GinkgoT().TempDir(), GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		target := filepath.Join(repo, "nested")
		Expect(os.Mkdir(filepath.Join(repo, ".git"), 0o700)).To(Succeed())
		Expect(os.Mkdir(target, 0o700)).To(Succeed())
		for dir, config := range map[string]string{
			home:   "externalIssues: {jira: {ignoreFields: [Internal*]}}\n",
			repo:   "externalIssues: {jira: {ignoreFields: [customfield_10053, Internal*]}}\n",
			target: "externalIssues: {jira: {ignoreFields: [Description]}}\n",
		} {
			Expect(os.WriteFile(filepath.Join(dir, GavelConfigFileName), []byte(config), 0o600)).To(Succeed())
		}
		cfg, err := LoadGavelConfig(target)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.ExternalIssues.Jira.IgnoreFields).To(Equal([]string{"Internal*", "customfield_10053", "Description"}))
		trace, err := LoadGavelConfigTrace(target)
		Expect(err).NotTo(HaveOccurred())
		Expect(trace.Merged.ExternalIssues).To(Equal(cfg.ExternalIssues))
		cfg.ExternalIssues.Jira.IgnoreFields[0] = "Changed"
		cached, err := LoadGavelConfig(target)
		Expect(err).NotTo(HaveOccurred())
		Expect(cached.ExternalIssues.Jira.IgnoreFields[0]).To(Equal("Internal*"))
	})

	It("documents ignoreFields as a string array in the configuration schema", func() {
		Expect(gavelConfigSchema()["properties"]).To(HaveKeyWithValue("externalIssues",
			HaveKeyWithValue("properties", HaveKeyWithValue("jira",
				HaveKeyWithValue("properties", HaveKeyWithValue("ignoreFields", SatisfyAll(
					HaveKeyWithValue("type", "array"), HaveKeyWithValue("items", HaveKeyWithValue("type", "string")),
				)))))))
	})
})
