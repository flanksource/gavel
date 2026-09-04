package utils

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GitIgnoredPaths", func() {
	It("honors repository, info, and global Git exclude sources", func() {
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
		Expect(os.WriteFile(filepath.Join(repo, ".gitignore"), []byte("*.log\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repo, ".git", "info", "exclude"), []byte("private/\n"), 0o644)).To(Succeed())

		paths := []string{
			filepath.Join(repo, ".agents", "memory.md"),
			filepath.Join(repo, "debug.log"),
			filepath.Join(repo, "private", "key.txt"),
			filepath.Join(repo, "main.go"),
		}
		ignored, err := GitIgnoredPaths(paths, repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(ignored).To(Equal(map[string]struct{}{
			paths[0]: {},
			paths[1]: {},
			paths[2]: {},
		}))
	})
})
