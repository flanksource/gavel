package betterleaks

import (
	"context"
	"os"
	"path/filepath"

	commonsContext "github.com/flanksource/commons/context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Secrets config confinement", func() {
	DescribeTable("rejects escaping config symlinks before scanning", func(extra bool) {
		root, outside := GinkgoT().TempDir(), GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		external := filepath.Join(outside, "config.toml")
		Expect(os.WriteFile(external, []byte("title = \"outside\"\n"), 0600)).To(Succeed())
		name := ".betterleaks.toml"
		if extra {
			name = "extra.toml"
			Expect(os.WriteFile(filepath.Join(root, ".gavel.yaml"), []byte("secrets:\n  configs: [extra.toml]\n"), 0600)).To(Succeed())
		}
		Expect(os.Symlink(external, filepath.Join(root, name))).To(Succeed())
		_, err := NewBetterleaks(root).Run(commonsContext.NewContext(context.Background()), nil)
		Expect(err).To(MatchError(ContainSubstring("discover betterleaks config")))
	}, Entry("native", false), Entry("declared", true))

	DescribeTable("rejects escaping config symlinks in direct resolution", func(count int) {
		root, outside := GinkgoT().TempDir(), GinkgoT().TempDir()
		external := filepath.Join(outside, "config.toml")
		Expect(os.WriteFile(external, []byte("title = \"outside\"\n"), 0600)).To(Succeed())
		alias := filepath.Join(root, "alias.toml")
		Expect(os.Symlink(external, alias)).To(Succeed())
		configs := []string{alias}
		if count == 2 {
			local := filepath.Join(root, "local.toml")
			Expect(os.WriteFile(local, []byte("title = \"inside\"\n"), 0600)).To(Succeed())
			configs = append(configs, local)
		}
		_, err := ResolveConfig(ResolveConfigOptions{WorkDir: root, Configs: configs})
		Expect(err).To(HaveOccurred())
	}, Entry("one", 1), Entry("merged", 2))
})
