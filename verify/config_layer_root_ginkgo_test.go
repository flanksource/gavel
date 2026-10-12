package verify

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoadGavelConfig layer confinement", func() {
	var repo string

	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		repo = GinkgoT().TempDir()
		Expect(os.Mkdir(filepath.Join(repo, ".git"), 0o755)).To(Succeed())
	})

	It("refuses a .gavel.yaml symlink that points outside its directory", func() {
		outside := filepath.Join(GinkgoT().TempDir(), "elsewhere.yaml")
		Expect(os.WriteFile(outside, []byte("procfile: {profile: outside}\n"), 0o600)).To(Succeed())
		Expect(os.Symlink(outside, filepath.Join(repo, GavelConfigFileName))).To(Succeed())

		_, err := LoadGavelConfig(repo)

		Expect(err).To(MatchError(ContainSubstring(filepath.Join(repo, GavelConfigFileName))))
	})

	It("reads a .gavel.yaml symlink that stays inside its directory", func() {
		Expect(os.Mkdir(filepath.Join(repo, "config"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repo, "config", "gavel.yaml"), []byte("procfile: {profile: inside}\n"), 0o600)).To(Succeed())
		Expect(os.Symlink(filepath.Join("config", "gavel.yaml"), filepath.Join(repo, GavelConfigFileName))).To(Succeed())

		cfg, err := LoadGavelConfig(repo)

		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Procfile.Profile).To(Equal("inside"))
	})
})
