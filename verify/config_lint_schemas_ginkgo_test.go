package verify

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("lint.schemas", func() {
	It("anchors each layer's globs and schema paths to its directory, the deeper layer winning a shared glob", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		repo, err := filepath.EvalSymlinks(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		site := filepath.Join(repo, "site")
		Expect(os.Mkdir(filepath.Join(repo, ".git"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(site, 0o755)).To(Succeed())
		write := func(dir, text string) {
			Expect(os.WriteFile(filepath.Join(dir, GavelConfigFileName), []byte(text), 0o600)).To(Succeed())
		}
		write(repo, "lint:\n  schemas:\n    \"docs/*.md\": schemas/docs.json\n    \"/abs/*.yaml\": https://example.com/abs.json\n")
		write(site, "lint:\n  schemas:\n    \"docs/*.md\": schemas/site-docs.json\n    \"/abs/*.yaml\": https://example.com/site.json\n")

		cfg, err := LoadGavelConfig(site)

		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Lint.Schemas).To(Equal(map[string]string{
			filepath.Join(repo, "docs/*.md"): filepath.Join(repo, "schemas/docs.json"),
			filepath.Join(site, "docs/*.md"): filepath.Join(site, "schemas/site-docs.json"),
			"/abs/*.yaml":                    "https://example.com/site.json",
		}))
	})
})
