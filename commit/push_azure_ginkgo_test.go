package commit

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"os/exec"
	"path/filepath"
)

var _ = Describe("Azure commit push validation", func() {
	It("rejects auto-merge before changing staged files or HEAD", func() {
		repo := GinkgoT().TempDir()
		runGit(repo, "init", "-b", "feature")
		runGit(repo, "commit", "--allow-empty", "-m", "root")
		runGit(repo, "remote", "add", "origin", "https://dev.azure.com/acme/product/_git/service")
		Expect(os.WriteFile(filepath.Join(repo, "feature.txt"), []byte("feature"), 0600)).To(Succeed())
		runGit(repo, "add", "feature.txt")
		before, err := exec.Command("git", "-C", repo, "status", "--porcelain=v1").Output()
		Expect(err).NotTo(HaveOccurred())
		_, err = Run(context.Background(), Options{WorkDir: repo, Push: true, AutoMerge: true, Stage: StageStaged, Message: "feat: feature"})
		Expect(err).To(MatchError(ContainSubstring("--auto-merge is not supported")))
		Expect(os.ReadFile(filepath.Join(repo, "feature.txt"))).To(Equal([]byte("feature")))
		after, err := exec.Command("git", "-C", repo, "status", "--porcelain=v1").Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(Equal(before))
		head, err := exec.Command("git", "-C", repo, "log", "-1", "--format=%s").Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(string(head)).To(Equal("root\n"))
	})
})
