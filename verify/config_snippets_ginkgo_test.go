package verify

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("config snippets", func() {
	write := func(dir, name, body string) string {
		GinkgoHelper()
		path := filepath.Join(dir, name)
		Expect(os.WriteFile(path, []byte(body), 0o644)).To(Succeed())
		return path
	}

	DescribeTable("names a layer's .gavel.yaml or a .gavel.<name>.yaml snippet as a config file",
		func(name string, want bool) {
			Expect(IsGavelConfigFile(name)).To(Equal(want))
		},
		Entry("the layer file", ".gavel.yaml", true),
		Entry("a snippet", ".gavel.agents.yaml", true),
		Entry("a dotted snippet name", ".gavel.ci.lint.yaml", true),
		Entry("an empty snippet name", ".gavel..yaml", false),
		Entry("a yml extension", ".gavel.agents.yml", false),
		Entry("the schema", "gavel.schema.json", false),
		Entry("an example without the leading dot", "gavel.yaml.example", false),
	)

	It("merges a directory's snippets over its .gavel.yaml in name order and traces which file set each field", func() {
		dir := GinkgoT().TempDir()
		main := write(dir, GavelConfigFileName, "commit:\n  types:\n  - feat\npr:\n  base: origin/main\n")
		zeta := write(dir, ".gavel.zeta.yaml", "pr:\n  base: origin/zeta\n")
		agents := write(dir, ".gavel.agents.yaml", "pr:\n  base: origin/agents\n  draft: true\n")
		write(dir, ".gavel.ignored.yml", "pr:\n  base: origin/ignored\n")

		trace, err := LoadGavelConfigTrace(dir)
		Expect(err).NotTo(HaveOccurred())
		var local []string
		for _, source := range trace.Sources {
			if filepath.Dir(source.Path) == dir {
				local = append(local, source.Path)
			}
		}
		Expect([]any{local, trace.Merged.PR.Base, trace.Merged.PR.Draft, trace.Merged.Commit.Types, trace.Provenance["/pr/base"], trace.Provenance["/pr/draft"], trace.Provenance["/commit/types"]}).
			To(Equal([]any{[]string{main, agents, zeta}, "origin/zeta", true, []string{"feat"}, zeta, agents, main}))
	})

	It("reloads the cached config when a snippet appears beside an unchanged .gavel.yaml", func() {
		dir := GinkgoT().TempDir()
		write(dir, GavelConfigFileName, "pr:\n  base: origin/main\n")
		before, err := LoadGavelConfig(dir)
		Expect(err).NotTo(HaveOccurred())

		write(dir, ".gavel.base.yaml", "pr:\n  base: origin/snippet\n")
		after, err := LoadGavelConfig(dir)
		Expect(err).NotTo(HaveOccurred())
		Expect([]string{before.PR.Base, after.PR.Base}).To(Equal([]string{"origin/main", "origin/snippet"}))
	})

	It("reads a layer made only of snippets, with no .gavel.yaml", func() {
		dir := GinkgoT().TempDir()
		write(dir, ".gavel.pr.yaml", "pr:\n  draft: true\n")

		cfg, err := LoadGavelConfig(dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.PR.Draft).To(BeTrue())
	})

	It("fails the load on an invalid snippet rather than skipping it", func() {
		dir := GinkgoT().TempDir()
		write(dir, ".gavel.bad.yaml", "pr: [unclosed\n")

		_, err := LoadGavelConfig(dir)
		Expect(err).To(MatchError(ContainSubstring(".gavel.bad.yaml")))
	})
})
