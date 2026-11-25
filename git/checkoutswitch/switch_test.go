package checkoutswitch_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/git/checkoutswitch"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCheckoutSwitch(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Checkout switch") }

func git(dir string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "%s: %s", args, out)
	return strings.TrimSpace(string(out))
}

func commit(dir, name, contents string) string {
	ExpectWithOffset(1, os.WriteFile(filepath.Join(dir, name), []byte(contents), 0600)).To(Succeed())
	git(dir, "add", name)
	git(dir, "commit", "-qm", "change "+name)
	return git(dir, "rev-parse", "HEAD")
}

var _ = Describe("Switch", func() {
	ctx := context.Background()
	var upstream, repo string

	head := func(ref string) string { return git(repo, "rev-parse", ref) }
	switchTo := func(sync checkoutswitch.Sync) (checkoutswitch.Result, error) {
		return checkoutswitch.Switch(ctx, checkoutswitch.Options{Repo: repo, Branch: "main", ExpectedHead: head("HEAD"), Sync: sync})
	}

	BeforeEach(func() {
		GinkgoT().Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
		GinkgoT().Setenv("GIT_CONFIG_NOSYSTEM", "1")
		GinkgoT().Setenv("GIT_AUTHOR_NAME", "Test")
		GinkgoT().Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
		GinkgoT().Setenv("GIT_COMMITTER_NAME", "Test")
		GinkgoT().Setenv("GIT_COMMITTER_EMAIL", "test@example.com")
		upstream = GinkgoT().TempDir()
		git(upstream, "init", "-qb", "main")
		commit(upstream, "base.txt", "base\n")
		repo = filepath.Join(GinkgoT().TempDir(), "clone")
		git(upstream, "clone", "-q", upstream, repo)
		git(repo, "switch", "-qc", "topic")
		commit(repo, "topic.txt", "topic\n")
	})

	It("fast-forwards a branch that is only behind its upstream and checks it out", func() {
		landed := commit(upstream, "upstream.txt", "new\n")
		previous := head("HEAD")
		result, err := switchTo(checkoutswitch.SyncFastForward)
		Expect(err).NotTo(HaveOccurred())
		Expect(git(repo, "branch", "--show-current")).To(Equal("main"))
		Expect(result).To(Equal(checkoutswitch.Result{
			Branch: "main", Head: landed, Sync: checkoutswitch.SyncFastForward,
			PreviousBranch: "topic", PreviousHead: previous, PreviousBranchHead: git(upstream, "rev-parse", "HEAD~1"),
		}))
	})

	It("refuses to fast-forward a diverged branch and leaves the checkout alone", func() {
		git(repo, "switch", "-q", "main")
		commit(repo, "local.txt", "local\n")
		local := head("main")
		git(repo, "switch", "-q", "topic")
		commit(upstream, "upstream.txt", "new\n")
		_, err := switchTo(checkoutswitch.SyncFastForward)
		Expect(err).To(MatchError(checkoutswitch.ErrRefused))
		Expect(err.Error()).To(ContainSubstring("diverged from origin/main (1 ahead, 1 behind)"))
		Expect(head("main")).To(Equal(local))
		Expect(git(repo, "branch", "--show-current")).To(Equal("topic"))
	})

	It("resets a diverged branch whose local commits another branch keeps", func() {
		git(repo, "switch", "-q", "main")
		local := commit(repo, "local.txt", "local\n")
		git(repo, "branch", "keep")
		git(repo, "switch", "-q", "topic")
		landed := commit(upstream, "upstream.txt", "new\n")
		result, err := switchTo(checkoutswitch.SyncReset)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Head).To(Equal(landed))
		Expect(head("keep")).To(Equal(local))
	})

	It("counts a local commit the upstream landed as a cherry-pick as preserved", func() {
		git(repo, "switch", "-q", "main")
		commit(repo, "local.txt", "local\n")
		git(repo, "switch", "-q", "topic")
		commit(upstream, "upstream.txt", "new\n")
		commit(upstream, "local.txt", "local\n")
		git(repo, "fetch", "-q")
		plan, err := checkoutswitch.Inspect(ctx, repo, "main")
		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Unpreserved).To(BeEmpty())
		Expect([]int{plan.Ahead, plan.Behind}).To(Equal([]int{1, 2}))
		_, err = switchTo(checkoutswitch.SyncReset)
		Expect(err).NotTo(HaveOccurred())
	})

	It("refuses a reset that would orphan a commit only the branch holds and lists it", func() {
		git(repo, "switch", "-q", "main")
		local := commit(repo, "local.txt", "local\n")
		git(repo, "switch", "-q", "topic")
		commit(upstream, "upstream.txt", "new\n")
		plan, err := checkoutswitch.Inspect(ctx, repo, "main")
		Expect(err).NotTo(HaveOccurred())
		Expect(plan.Unpreserved).To(Equal([]checkoutswitch.Commit{{SHA: local, Subject: "change local.txt"}}))
		_, err = switchTo(checkoutswitch.SyncReset)
		Expect(err).To(MatchError(checkoutswitch.ErrRefused))
		Expect(err.Error()).To(ContainSubstring(local[:8] + " change local.txt"))
		Expect(head("main")).To(Equal(local))
	})

	It("syncs the branch already checked out", func() {
		git(repo, "switch", "-q", "main")
		landed := commit(upstream, "upstream.txt", "new\n")
		result, err := switchTo(checkoutswitch.SyncFastForward)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Head).To(Equal(landed))
		Expect(filepath.Join(repo, "upstream.txt")).To(BeAnExistingFile())
	})

	It("undoes the branch move when switching would overwrite an untracked file", func() {
		before := head("main")
		commit(upstream, "clash.txt", "upstream\n")
		Expect(os.WriteFile(filepath.Join(repo, "clash.txt"), []byte("mine\n"), 0600)).To(Succeed())
		_, err := switchTo(checkoutswitch.SyncFastForward)
		Expect(err).To(HaveOccurred())
		Expect(head("main")).To(Equal(before))
		Expect(git(repo, "branch", "--show-current")).To(Equal("topic"))
	})

	It("switches without syncing when asked to", func() {
		before := head("main")
		commit(upstream, "upstream.txt", "new\n")
		result, err := switchTo(checkoutswitch.SyncNone)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Head).To(Equal(before))
	})

	DescribeTable("refuses an unsafe switch",
		func(prepare func() string, want string) {
			expected := prepare()
			_, err := checkoutswitch.Switch(ctx, checkoutswitch.Options{Repo: repo, Branch: "main", ExpectedHead: expected, Sync: checkoutswitch.SyncNone})
			Expect(err).To(MatchError(checkoutswitch.ErrRefused))
			Expect(err.Error()).To(ContainSubstring(want))
			Expect(git(repo, "branch", "--show-current")).To(Equal("topic"))
		},
		Entry("with uncommitted tracked changes", func() string {
			Expect(os.WriteFile(filepath.Join(repo, "topic.txt"), []byte("edited\n"), 0600)).To(Succeed())
			return head("HEAD")
		}, "1 tracked files"),
		Entry("when HEAD moved since the plan", func() string {
			stale := head("HEAD")
			commit(repo, "later.txt", "later\n")
			return stale
		}, "refresh and try again"),
		Entry("when another worktree has the branch checked out", func() string {
			git(repo, "worktree", "add", "-q", filepath.Join(GinkgoT().TempDir(), "main-wt"), "main")
			return head("HEAD")
		}, "is checked out in the worktree"),
	)
})
