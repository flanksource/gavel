package branchmerge_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const featureBranch = "feature/work"

// mergeFixture is a main checkout on main plus a linked worktree on
// featureBranch carrying two commits (a.txt, b.txt) branched off main.
type mergeFixture struct {
	repo, worktree string
	base           string
	commits        []string
}

func git(dir string, args ...string) string {
	GinkgoHelper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_SYSTEM="+os.DevNull)
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %s:\n%s", strings.Join(args, " "), out)
	return strings.TrimSpace(string(out))
}

func commitFile(dir, name, content, message string) string {
	GinkgoHelper()
	Expect(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)).To(Succeed())
	git(dir, "add", name)
	git(dir, "commit", "-q", "-m", message)
	return git(dir, "rev-parse", "HEAD")
}

func branchExists(repo, branch string) bool {
	return exec.Command("git", "-C", repo, "rev-parse", "--verify", "--quiet", "refs/heads/"+branch).Run() == nil
}

func newMergeFixture() *mergeFixture {
	GinkgoHelper()
	root, err := filepath.EvalSymlinks(GinkgoT().TempDir())
	Expect(err).NotTo(HaveOccurred())
	f := &mergeFixture{repo: filepath.Join(root, "repo"), worktree: filepath.Join(root, "wt")}
	Expect(os.MkdirAll(f.repo, 0o755)).To(Succeed())
	git(f.repo, "init", "-q", "-b", "main")
	git(f.repo, "config", "user.email", "test@example.com")
	git(f.repo, "config", "user.name", "test")
	f.base = commitFile(f.repo, "README.md", "hello\n", "chore: initial")
	git(f.repo, "worktree", "add", "-q", "-b", featureBranch, f.worktree, f.base)
	f.commits = []string{
		commitFile(f.worktree, "a.txt", "a\n", "feat: add a"),
		commitFile(f.worktree, "b.txt", "b\n", "feat: add b"),
	}
	return f
}
