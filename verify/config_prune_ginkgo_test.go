package verify

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func savedConfig(cfg GavelConfig) string {
	GinkgoHelper()
	dir := GinkgoT().TempDir()
	Expect(SaveGavelConfig(dir, cfg)).To(Succeed())
	data, err := os.ReadFile(filepath.Join(dir, GavelConfigFileName))
	Expect(err).NotTo(HaveOccurred())
	return string(data)
}

var _ = Describe("saving config", func() {
	It("writes only the sections that carry a setting", func() {
		var cfg GavelConfig
		cfg.Lint.Ignore = []LintIgnoreRule{{Rule: "acme-brand-names", Source: "betterleaks"}}

		Expect(savedConfig(cfg)).To(Equal("lint:\n  ignore:\n  - rule: acme-brand-names\n    source: betterleaks\n"))
	})

	It("writes nothing for a config that configures nothing", func() {
		Expect(savedConfig(GavelConfig{})).To(Equal("{}\n"))
	})

	It("collapses a section whose only children are empty", func() {
		var cfg GavelConfig
		cfg.Commit.Types = []string{"feat"}

		saved := savedConfig(cfg)
		Expect(saved).To(ContainSubstring("types:"))
		// commit.lint, commit.tidy, commit.precommit and the three PromptSpecs are
		// all zero-valued structs that omitempty cannot drop.
		Expect(saved).NotTo(ContainSubstring("{}"))
		Expect(saved).NotTo(ContainSubstring("precommit"))
	})

	DescribeTable("keeps values a user can mean",
		func(mutate func(*GavelConfig), expected string) {
			var cfg GavelConfig
			mutate(&cfg)
			Expect(savedConfig(cfg)).To(ContainSubstring(expected))
		},
		Entry("a false pointer gate", func(c *GavelConfig) {
			disabled := false
			c.Commit.Lint.Secrets = &disabled
		}, "secrets: false"),
		Entry("a true pointer gate", func(c *GavelConfig) {
			enabled := true
			c.Commit.Tidy.Enabled = &enabled
		}, "enabled: true"),
		Entry("a populated list", func(c *GavelConfig) {
			c.Commit.GitIgnore = []string{"node_modules/"}
		}, "node_modules/"),
		Entry("a non-zero integer", func(c *GavelConfig) {
			c.Commit.MaxCommits = 7
		}, "maxCommits: 7"),
	)

	It("round-trips a large integer without reformatting it", func() {
		var cfg GavelConfig
		cfg.Todos.CheckConcurrency = 9007199254740993

		Expect(savedConfig(cfg)).To(ContainSubstring("9007199254740993"))
	})

	It("drops the dead keys an older gavel wrote, keeping the real content", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, GavelConfigFileName)
		legacy := "checks: {}\n" +
			"commit:\n  compatibility: {}\n  messagePrompt: {}\n  precommit: {}\n  summaryPrompt: {}\n" +
			"fixtures: {}\n" +
			"lint:\n  ignore:\n  - rule: acme-brand-names\n    source: betterleaks\n" +
			"todos:\n  runPrompt: {}\n" +
			"verify:\n  model: \"\"\n"
		Expect(os.WriteFile(path, []byte(legacy), 0o600)).To(Succeed())

		cfg, err := LoadSingleGavelConfig(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(SaveGavelConfig(dir, cfg)).To(Succeed())

		data, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal("lint:\n  ignore:\n  - rule: acme-brand-names\n    source: betterleaks\n"))
	})
})
