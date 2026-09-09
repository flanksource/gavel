package betterleaks

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hairyhenderson/toml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBetterleaksGitIgnore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Betterleaks Git Ignore Suite")
}

var _ = Describe("Betterleaks Git ignore passthrough", func() {
	It("adds ignored paths to a wrapper scan config", func() {
		workDir := GinkgoT().TempDir()
		base := filepath.Join(workDir, "base.toml")
		Expect(os.WriteFile(base, []byte("title = \"base\"\n"), 0o644)).To(Succeed())

		out, err := ResolveConfig(ResolveConfigOptions{
			WorkDir:      workDir,
			Configs:      []string{base},
			IgnoredPaths: []string{".agents/", "debug.log"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal(filepath.Join(workDir, ".tmp", "betterleaks-scan.toml")))

		data, err := os.ReadFile(out)
		Expect(err).NotTo(HaveOccurred())
		var cfg tomlConfig
		Expect(toml.Unmarshal(data, &cfg)).To(Succeed())
		Expect(cfg.Extend.Path).To(Equal(base))
		Expect(cfg.Allowlists).To(Equal([]tomlAllowlist{{
			Description: "Paths ignored by Git",
			Paths:       []string{`^\.agents(?:/|$)`, `^debug\.log$`},
		}}))
	})

	It("discovers directories ignored by repository and global Git excludes", func() {
		home := GinkgoT().TempDir()
		xdg := filepath.Join(home, "xdg")
		GinkgoT().Setenv("HOME", home)
		GinkgoT().Setenv("XDG_CONFIG_HOME", xdg)
		Expect(os.MkdirAll(filepath.Join(xdg, "git"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(xdg, "git", "ignore"), []byte(".agents/\n"), 0o644)).To(Succeed())

		repo := GinkgoT().TempDir()
		cmd := exec.Command("git", "init", "-q")
		cmd.Dir = repo
		Expect(cmd.Run()).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repo, ".gitignore"), []byte(".tmp/\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(repo, ".agents"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repo, ".agents", "memory.md"), nil, 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(repo, ".tmp"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repo, ".tmp", "report.json"), nil, 0o644)).To(Succeed())

		paths, err := gitIgnoredScanPaths(repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(ConsistOf(".agents/", ".tmp/"))
	})
})
