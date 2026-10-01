package prcreate

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type repoFixture struct {
	repo     string // working copy
	bare     string // bare repo acting as `origin`
	baseSHA  string // commit on main in working copy
	topicSHA string // feature commit on a branch in working copy
}

func git(dir string, args ...string) string {
	GinkgoHelper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@example.com",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
	)
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %s:\n%s", strings.Join(args, " "), out)
	return strings.TrimSpace(string(out))
}

func commitFile(repo, name, content, message string) string {
	GinkgoHelper()
	Expect(os.WriteFile(filepath.Join(repo, name), []byte(content), 0o644)).To(Succeed())
	git(repo, "add", name)
	git(repo, "commit", "-m", message)
	return git(repo, "rev-parse", "HEAD")
}

// newRepoFixture builds a working copy with a local bare `origin`, a pushed
// main commit, and one unpushed feature commit on branch `feature`.
func newRepoFixture() *repoFixture {
	GinkgoHelper()
	// The PR content prompt loads .gavel.yaml, which layers ~/.gavel.yaml under
	// the repo's; without an isolated HOME these assert against the developer's.
	GinkgoT().Setenv("HOME", GinkgoT().TempDir())
	root := GinkgoT().TempDir()
	bare := filepath.Join(root, "origin.git")
	repo := filepath.Join(root, "work")

	git(root, "init", "--bare", "-b", "main", bare)
	git(root, "init", "-b", "main", repo)
	// Persist an identity in the repo's local config so git operations spawned
	// by production code (which doesn't inject GIT_AUTHOR_*) can author commits
	// even on a CI runner with no global git identity.
	git(repo, "config", "user.email", "test@example.com")
	git(repo, "config", "user.name", "test")
	git(repo, "remote", "add", "origin", bare)

	baseSHA := commitFile(repo, "README.md", "hello\n", "chore: initial")
	git(repo, "push", "-u", "origin", "main")

	git(repo, "checkout", "-b", "feature")
	topicSHA := commitFile(repo, "feature.txt", "new\n", "feat: add feature.txt")
	git(repo, "checkout", "main")

	return &repoFixture{repo: repo, bare: bare, baseSHA: baseSHA, topicSHA: topicSHA}
}

func (f *repoFixture) scratchEntries() []os.DirEntry {
	GinkgoHelper()
	entries, err := os.ReadDir(filepath.Join(f.repo, scratchSub))
	if os.IsNotExist(err) {
		return nil
	}
	Expect(err).NotTo(HaveOccurred())
	return entries
}
