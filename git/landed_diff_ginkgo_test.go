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

var _ = Describe("LandedDiffStat", func() {
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

	It("diffs the last n commits ending at the landed tip", func() {
		commit("base.txt", "base\n", "chore: base")
		commit("a.txt", "a1\na2\n", "feat: a")
		tip := commit("a.txt", "a1\n", "fix: trim a")

		stat, err := gavelgit.LandedDiffStat(dir, tip, 2)

		Expect(err).NotTo(HaveOccurred())
		Expect(stat).To(Equal(gavelgit.DiffStat{Commits: 2, Files: 1, Adds: 1, Dels: 0}))
	})

	It("reports ErrCommitNotFound when the tip is not in the repository", func() {
		commit("base.txt", "base\n", "chore: base")
		_, err := gavelgit.LandedDiffStat(dir, "feedfacefeedfacefeedfacefeedfacefeedface", 1)
		Expect(err).To(MatchError(gavelgit.ErrCommitNotFound))
	})

	It("surfaces a git failure when the tip has fewer ancestors than n", func() {
		tip := commit("a.txt", "a\n", "feat: a")
		_, err := gavelgit.LandedDiffStat(dir, tip, 3)
		Expect(err).To(HaveOccurred())
		Expect(err).NotTo(MatchError(gavelgit.ErrCommitNotFound))
		Expect(err).To(MatchError(ContainSubstring(tip + "~3")))
	})

	DescribeTable("rejects input before running git", func(tip string, n int, message string) {
		_, err := gavelgit.LandedDiffStat(dir, tip, n)
		Expect(err).To(MatchError(ContainSubstring(message)))
	},
		Entry("a value that is not a commit hash", "--all", 1, `invalid commit hash "--all"`),
		Entry("no commits to count", "abc1234", 0, "at least one commit"),
	)
})
