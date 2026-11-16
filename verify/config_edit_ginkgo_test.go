package verify

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("edit config", func() {
	DescribeTable("rejects extensions and patterns gavel edit cannot apply",
		func(cfg EditConfig, message string) {
			_, err := cfg.Normalize()
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("an extension without a dot", EditConfig{Extensions: []string{"md"}}, `"md" must look like .md`),
		Entry("an extension holding a path", EditConfig{Extensions: []string{".d/x"}}, `must look like .md`),
		Entry("a pattern leaving the folder", EditConfig{IncludeIgnored: []string{"../secrets/**"}}, "must be a pattern inside the folder"),
		Entry("an absolute pattern", EditConfig{IncludeIgnored: []string{"/etc/**"}}, "must be a pattern inside the folder"),
		Entry("an empty pattern", EditConfig{IncludeIgnored: []string{" "}}, "must be a pattern inside the folder"),
		Entry("an invalid glob", EditConfig{IncludeIgnored: []string{"dist/[x"}}, "invalid glob"),
	)

	It("lower-cases extensions and cleans patterns", func() {
		Expect(EditConfig{Extensions: []string{" .GO "}, IncludeIgnored: []string{"./.tmp/"}}.Normalize()).
			To(Equal(EditConfig{Extensions: []string{".go"}, IncludeIgnored: []string{".tmp"}}))
	})

	It("saves the edit section and keeps the directory's other sections", func() {
		dir := GinkgoT().TempDir()
		file := filepath.Join(dir, GavelConfigFileName)
		Expect(os.WriteFile(file, []byte("pr:\n  base: origin/main\n"), 0o644)).To(Succeed())

		Expect(SaveEditConfig(dir, EditConfig{Extensions: []string{".go", ".md"}, IncludeIgnored: []string{".tmp"}})).To(Succeed())

		saved, err := LoadSingleGavelConfig(file)
		Expect(err).NotTo(HaveOccurred())
		Expect(saved.Edit).To(Equal(EditConfig{Extensions: []string{".go", ".md"}, IncludeIgnored: []string{".tmp"}}))
		Expect(saved.PR.Base).To(Equal("origin/main"))
	})

	It("refuses to load a .gavel.yaml with an invalid edit section", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, GavelConfigFileName), []byte("edit:\n  extensions: [md]\n"), 0o644)).To(Succeed())

		_, err := LoadSingleGavelConfig(filepath.Join(dir, GavelConfigFileName))
		Expect(err).To(MatchError(ContainSubstring(`edit.extensions: "md"`)))
	})
})
