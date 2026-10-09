package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gavelgit "github.com/flanksource/gavel/git"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RangeDiffStat", func() {
	var dir string
	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), "git %v: %s", args, out)
		return strings.TrimSpace(string(out))
	}
	commit := func(file, content, message string) string {
		Expect(os.WriteFile(filepath.Join(dir, file), []byte(content), 0o644)).To(Succeed())
		git("add", file)
		git("commit", "-q", "-m", message)
		return git("rev-parse", "HEAD")
	}
	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		git("init", "-q")
		git("config", "user.email", "test@example.com")
		git("config", "user.name", "Test User")
		git("config", "commit.gpgsign", "false")
	})

	It("counts the commits, files and lines between from and to, excluding from itself", func() {
		commit("base.txt", "base\n", "chore: base")
		setup := commit("wip.txt", "one\ntwo\nthree\n", "chore(setup): snapshot uncommitted changes")
		commit("a.txt", "a1\na2\n", "feat: a")
		head := commit("wip.txt", "one\n", "fix: trim wip")

		stat, err := gavelgit.RangeDiffStat(dir, setup, head)

		Expect(err).NotTo(HaveOccurred())
		Expect(stat).To(Equal(gavelgit.DiffStat{Commits: 2, Files: 2, Adds: 2, Dels: 2}))
	})

	It("is empty when from and to are the same commit", func() {
		head := commit("a.txt", "a\n", "feat: a")
		Expect(gavelgit.RangeDiffStat(dir, head, head)).To(Equal(gavelgit.DiffStat{}))
	})

	It("fails when a sha is not in the repository", func() {
		head := commit("a.txt", "a\n", "feat: a")
		_, err := gavelgit.RangeDiffStat(dir, "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef", head)
		Expect(err).To(MatchError(ContainSubstring("deadbeef")))
	})

	It("rejects a value that is not a commit hash before running git", func() {
		_, err := gavelgit.RangeDiffStat(dir, "--all", "HEAD")
		Expect(err).To(MatchError(ContainSubstring(`invalid commit hash "--all"`)))
	})
})
