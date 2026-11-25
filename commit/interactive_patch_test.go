package commit

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/gavel/status"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("interactive file patch", func() {
	var repo string

	BeforeEach(func() {
		repo = GinkgoT().TempDir()
		previewGit(repo, "init")
		previewGit(repo, "config", "user.email", "test@example.com")
		previewGit(repo, "config", "user.name", "Test User")
	})

	It("shows the net staged and unstaged change for one literal path without touching the index", func(ctx SpecContext) {
		path := "src[1].go"
		writeRepoFile(repo, path, "before\n")
		writeRepoFile(repo, "src1.go", "unrelated before\n")
		previewGit(repo, "--literal-pathspecs", "add", "--", path, "src1.go")
		previewGit(repo, "-c", "commit.gpgsign=false", "commit", "-m", "initial")
		writeRepoFile(repo, path, "intermediate\n")
		previewGit(repo, "--literal-pathspecs", "add", "--", path)
		writeRepoFile(repo, path, "after\n")
		writeRepoFile(repo, "src1.go", "unrelated after\n")
		indexPath := filepath.Join(repo, ".git", "index")
		staged, err := os.ReadFile(indexPath)
		Expect(err).NotTo(HaveOccurred())

		patch, err := readTreePatch(ctx, repo, status.FileStatus{Path: path, State: status.StateBoth})
		Expect(err).NotTo(HaveOccurred())
		Expect(patch).To(And(ContainSubstring("\n-before\n+after\n"), Not(ContainSubstring("intermediate")), Not(ContainSubstring("unrelated"))))
		Expect(os.ReadFile(indexPath)).To(Equal(staged))
	})

	It("shows an untracked file and staged additions in a repository without HEAD", func(ctx SpecContext) {
		writeRepoFile(repo, "first.go", "staged\n")
		previewGit(repo, "add", "first.go")
		writeRepoFile(repo, "first.go", "working\n")
		writeRepoFile(repo, "untracked.go", "new\n")

		first, err := readTreePatch(ctx, repo, status.FileStatus{Path: "first.go", State: status.StateBoth})
		Expect(err).NotTo(HaveOccurred())
		Expect(first).To(And(ContainSubstring("--- /dev/null"), ContainSubstring("\n+working\n"), Not(ContainSubstring("+staged"))))
		untracked, err := readTreePatch(ctx, repo, status.FileStatus{Path: "untracked.go", State: status.StateUntracked})
		Expect(err).NotTo(HaveOccurred())
		Expect(untracked).To(And(ContainSubstring("--- /dev/null"), ContainSubstring("\n+new\n")))
	})

	It("includes rename metadata, deletions, and a binary file summary", func(ctx SpecContext) {
		writeRepoFile(repo, "old.go", "renamed content\n")
		writeRepoFile(repo, "gone.go", "deleted content\n")
		previewGit(repo, "add", "old.go", "gone.go")
		previewGit(repo, "-c", "commit.gpgsign=false", "commit", "-m", "initial")
		previewGit(repo, "mv", "old.go", "new.go")
		Expect(os.Remove(filepath.Join(repo, "gone.go"))).To(Succeed())
		writeRepoFile(repo, "new.bin", "\x00binary\x00")

		rename, err := readTreePatch(ctx, repo, status.FileStatus{Path: "new.go", PreviousPath: "old.go", State: status.StateStaged})
		Expect(err).NotTo(HaveOccurred())
		Expect(rename).To(ContainSubstring("rename from old.go\nrename to new.go\n"))
		deleted, err := readTreePatch(ctx, repo, status.FileStatus{Path: "gone.go", State: status.StateUnstaged})
		Expect(err).NotTo(HaveOccurred())
		Expect(deleted).To(ContainSubstring("\n-deleted content\n"))
		binary, err := readTreePatch(ctx, repo, status.FileStatus{Path: "new.bin", State: status.StateUntracked})
		Expect(err).NotTo(HaveOccurred())
		Expect(binary).To(ContainSubstring("Binary files /dev/null and b/new.bin differ"))
	})
})

func previewGit(repo string, args ...string) string {
	cmd := exec.CommandContext(context.Background(), "git", args...)
	cmd.Dir = repo
	output, err := cmd.CombinedOutput()
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "git %v: %s", args, output)
	return string(output)
}
