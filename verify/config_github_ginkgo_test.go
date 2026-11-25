package verify

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var _ = Describe("GitHub linter configuration", func() {
	It("accepts repository queries alongside literal names", func() {
		entries := []string{"acme/service", "org:acme topic:mission-control"}
		cfg, err := DecodeGitHubLintConfig(map[string]any{"repos": entries})
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Repos).To(Equal(entries))
	})
	DescribeTable("rejects empty repository selectors", func(entry string) {
		_, err := DecodeGitHubLintConfig(map[string]any{"repos": []string{entry}})
		Expect(err).To(HaveOccurred())
	}, Entry("empty", ""), Entry("whitespace", "  "), Entry("empty qualifier", "org:"))
	It("decodes typed fields and defaults", func() {
		cfg, err := DecodeGitHubLintConfig(map[string]any{"orgs": []string{"acme"}, "repos": []string{"personal/service"}, "excludeRepos": []string{"acme/legacy-*"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg).To(Equal(GitHubLintConfig{Orgs: []string{"acme"}, Repos: []string{"personal/service"}, ExcludeRepos: []string{"acme/legacy-*"}, StaleAfter: "720h"}))
	})
	DescribeTable("rejects invalid configuration", func(raw map[string]any) {
		_, err := DecodeGitHubLintConfig(raw)
		Expect(err).To(HaveOccurred())
	}, Entry("unknown field", map[string]any{"org": "acme"}), Entry("bad repo", map[string]any{"repos": []string{"service"}}), Entry("bad glob", map[string]any{"excludeRepos": []string{"acme/["}}), Entry("zero duration", map[string]any{"staleAfter": "0s"}), Entry("negative duration", map[string]any{"staleAfter": "-1h"}), Entry("wrong type", map[string]any{"orgs": "acme"}))
	It("allows GitHub config at the runtime validation boundary", func() {
		cfg := GavelConfig{Lint: LintConfig{Linters: map[string]LintLinterConfig{"github": {Config: map[string]any{"orgs": []string{"acme"}}}}}}
		Expect(cfg.validateRuntime()).To(Succeed())
	})
	It("replaces lists and inherits omitted fields across layers", func() {
		base := GavelConfig{Lint: LintConfig{Linters: map[string]LintLinterConfig{"github": {Config: map[string]any{"orgs": []any{"acme"}, "excludeRepos": []any{"acme/old-*"}, "staleAfter": "240h"}}}}}
		over := GavelConfig{Lint: LintConfig{Linters: map[string]LintLinterConfig{"github": {Config: map[string]any{"orgs": []any{}, "repos": []any{"personal/service"}}}}}}
		cfg, err := DecodeGitHubLintConfig(MergeGavelConfig(base, over).Lint.Linters["github"].Config)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Orgs).To(BeEmpty())
		Expect(cfg.Repos).To(Equal([]string{"personal/service"}))
		Expect(cfg.ExcludeRepos).To(Equal([]string{"acme/old-*"}))
		Expect(cfg.StaleAfter).To(Equal("240h"))
	})
	It("publishes typed GitHub settings while retaining opaque native UIR config", func() {
		raw, err := ConfigJSONSchema()
		Expect(err).NotTo(HaveOccurred())
		document, err := jsonschema.UnmarshalJSON(strings.NewReader(raw))
		Expect(err).NotTo(HaveOccurred())
		compiler := jsonschema.NewCompiler()
		Expect(compiler.AddResource(ConfigSchemaID, document)).To(Succeed())
		schema, err := compiler.Compile(ConfigSchemaID)
		Expect(err).NotTo(HaveOccurred())
		for _, native := range []map[string]any{
			{"repos": []any{"org:acme topic:mission-control", "acme/service"}},
			{"orgs": []any{"acme"}, "repos": []any{"personal/service"}, "excludeRepos": []any{"acme/*"}, "staleAfter": "240h"},
			{},
		} {
			Expect(schema.Validate(map[string]any{"lint": map[string]any{"linters": map[string]any{"github": map[string]any{"config": native}, "uir": map[string]any{"config": map[string]any{"nativeField": true}}}}})).To(Succeed())
		}
		for _, native := range []map[string]any{{"org": "acme"}, {"orgs": "acme"}, {"repos": []any{"missing-owner"}}, {"staleAfter": 24}} {
			Expect(schema.Validate(map[string]any{"lint": map[string]any{"linters": map[string]any{"github": map[string]any{"config": native}}}})).NotTo(Succeed())
		}
	})
})
