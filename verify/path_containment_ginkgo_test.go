package verify

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoadSingleGavelConfig", func() {
	var dir string

	BeforeEach(func() {
		dir = GinkgoT().TempDir()
	})

	It("reads the .gavel.yaml in the requested directory", func() {
		Expect(os.WriteFile(filepath.Join(dir, GavelConfigFileName), []byte("ai:\n  model: claude\n"), 0o600)).To(Succeed())

		cfg, err := LoadSingleGavelConfig(filepath.Join(dir, GavelConfigFileName))

		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.AI.Model.Name).To(Equal("claude"))
	})

	It("refuses to read a file that is not a .gavel.yaml", func() {
		secret := filepath.Join(dir, "secrets.yaml")
		Expect(os.WriteFile(secret, []byte("ai:\n  model: claude\n"), 0o600)).To(Succeed())

		_, err := LoadSingleGavelConfig(secret)

		Expect(err).To(MatchError(ContainSubstring(GavelConfigFileName)))
		Expect(err).To(MatchError(ContainSubstring(secret)))
		Expect(os.IsNotExist(err)).To(BeFalse(), "an unreadable name is a caller error, not a missing file")
	})

	It("refuses a traversal that ends in a different file name", func() {
		_, err := LoadSingleGavelConfig(filepath.Join(dir, "..", "..", "etc", "passwd"))

		Expect(err).To(MatchError(ContainSubstring(GavelConfigFileName)))
	})
})

var _ = Describe("PromptSpec.TemplateSource", func() {
	const (
		body     = "---\nmodel: claude\n---\nBody {{diff}}"
		fallback = "builtin body"
		relative = "prompts/grouping.prompt"
	)
	var dir string

	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(dir, filepath.Dir(relative)), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, relative), []byte(body), 0o600)).To(Succeed())
	})

	It("reads a relative file against the directory that declared it", func() {
		Expect(PromptSpec{File: relative}.TemplateSource(dir, fallback)).To(Equal(body))
	})

	It("refuses a relative file that walks out of the declaring directory", func() {
		escaping := filepath.Join("..", filepath.Base(dir), relative)

		_, err := PromptSpec{File: escaping}.TemplateSource(dir, fallback)

		Expect(err).To(MatchError(ContainSubstring(escaping)))
		Expect(err).To(MatchError(ContainSubstring(dir)))
	})

	It("reads an absolute file inside the declaring directory", func() {
		Expect(PromptSpec{File: filepath.Join(dir, relative)}.TemplateSource(dir, fallback)).To(Equal(body))
	})

	// A hosted run reads its repository's own .gavel.yaml on serve, so a prompt file it names
	// must not lead to any other file serve can read.
	It("refuses an absolute file outside the declaring directory", func() {
		outside := filepath.Join(GinkgoT().TempDir(), "secret.prompt")
		Expect(os.WriteFile(outside, []byte(body), 0o600)).To(Succeed())

		_, err := PromptSpec{File: outside}.TemplateSource(dir, fallback)

		Expect(err).To(MatchError(ContainSubstring(outside)))
		Expect(err).To(MatchError(ContainSubstring(dir)))
	})

	It("refuses a file that is a symlink out of the declaring directory", func() {
		outside := filepath.Join(GinkgoT().TempDir(), "secret.prompt")
		Expect(os.WriteFile(outside, []byte(body), 0o600)).To(Succeed())
		Expect(os.Symlink(outside, filepath.Join(dir, "link.prompt"))).To(Succeed())

		source, err := PromptSpec{File: "link.prompt"}.TemplateSource(dir, fallback)

		Expect(err).To(MatchError(ContainSubstring("escapes")))
		Expect(source).To(BeEmpty())
	})

	It("falls back to the builtin body when no file and no prompt are set", func() {
		Expect(PromptSpec{}.TemplateSource(dir, fallback)).To(Equal(fallback))
	})
})
