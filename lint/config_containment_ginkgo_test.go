package lint

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/linters/betterleaks"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Invalid secrets config discovery", func() {
	It("fails selection, execution and preview instead of skipping the scanner", func() {
		root, outside := GinkgoT().TempDir(), GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		Expect(os.Mkdir(filepath.Join(root, ".git"), 0700)).To(Succeed())
		external := filepath.Join(outside, "secrets.toml")
		Expect(os.WriteFile(external, []byte("title = \"outside\"\n"), 0600)).To(Succeed())
		Expect(os.Symlink(external, filepath.Join(root, ".betterleaks.toml"))).To(Succeed())
		_, _, err := shouldSelectLinter(root, verify.GavelConfig{}, betterleaks.NewBetterleaks(root), false)
		Expect(err).To(MatchError(ContainSubstring("discover betterleaks config")))
		_, err = Execute(Options{WorkDir: root})
		Expect(err).To(MatchError(ContainSubstring("discover betterleaks config")))
		Expect(DryRun(Options{WorkDir: root})).To(MatchError(ContainSubstring("discover betterleaks config")))
	})
})
