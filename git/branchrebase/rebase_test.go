package branchrebase_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/git/branchrebase"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func TestBranchRebase(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Branch rebase") }

func git(dir string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1", "GIT_EDITOR=true")
	out, err := cmd.CombinedOutput()
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "%s: %s", args, out)
	return strings.TrimSpace(string(out))
}

func write(dir, name, contents string) {
	ExpectWithOffset(1, os.WriteFile(filepath.Join(dir, name), []byte(contents), 0600)).To(Succeed())
}
func HaveContents(want string) types.GomegaMatcher {
	return WithTransform(func(path string) (string, error) { data, err := os.ReadFile(path); return string(data), err }, Equal(want))
}
func commit(dir, name, contents string) string {
	write(dir, name, contents)
	git(dir, "add", name)
	git(dir, "commit", "-qm", "change "+name)
	return git(dir, "rev-parse", "HEAD")
}

var _ = Describe("Rebase and keep", func() {
	var repo, worktree, head string
	var opts branchrebase.Options
	BeforeEach(func() {
		repo = GinkgoT().TempDir()
		git(repo, "init", "-qb", "main")
		git(repo, "config", "user.email", "test@example.com")
		git(repo, "config", "user.name", "Test")
		commit(repo, "shared file.txt", "base\n")
		worktree = filepath.Join(GinkgoT().TempDir(), "topic")
		git(repo, "worktree", "add", "-qb", "topic", worktree)
		head = commit(worktree, "topic.txt", "topic\n")
		commit(repo, "main.txt", "main\n")
		opts = branchrebase.Options{Action: "start", Worktree: worktree, Branch: "topic", Target: "refs/heads/main", ExpectedHead: head}
	})

	It("restores staged, unstaged and untracked edits without landing or deleting the worktree", func() {
		write(worktree, "topic.txt", "staged\n")
		git(worktree, "add", "topic.txt")
		write(worktree, "topic.txt", "unstaged\n")
		write(worktree, "untracked file.txt", "saved\n")
		before := git(worktree, "status", "--porcelain")
		main := git(repo, "rev-parse", "HEAD")
		result, err := branchrebase.Run(context.Background(), opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Phase).To(Equal("done"))
		Expect(git(worktree, "status", "--porcelain")).To(Equal(before))
		Expect(git(worktree, "show", ":topic.txt")).To(Equal("staged"))
		Expect(filepath.Join(worktree, "topic.txt")).To(HaveContents("unstaged\n"))
		Expect(filepath.Join(worktree, "untracked file.txt")).To(HaveContents("saved\n"))
		Expect(git(repo, "rev-parse", "HEAD")).To(Equal(main))
		git(worktree, "merge-base", "--is-ancestor", main, "HEAD")
		Expect(git(worktree, "branch", "--show-current")).To(Equal("topic"))
	})

	It("lists the primary checkout's branch first, then origin/main, then the other local branches", func() {
		git(repo, "branch", "extra")
		state, err := branchrebase.Inspect(context.Background(), worktree)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.Targets).To(Equal([]branchrebase.Target{
			{Ref: "refs/heads/main", Label: "main (main checkout)"},
			{Ref: "refs/remotes/origin/main", Label: "origin/main"},
			{Ref: "refs/heads/extra", Label: "extra"},
		}))
	})

	It("does not offer the primary checkout's branch when inspecting the primary checkout itself", func() {
		state, err := branchrebase.Inspect(context.Background(), repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.Targets).To(Equal([]branchrebase.Target{
			{Ref: "refs/remotes/origin/main", Label: "origin/main"},
			{Ref: "refs/heads/topic", Label: "topic"},
		}))
	})

	It("merges the target into the branch with a merge commit, keeping the branch's commit and saved edits", func() {
		write(worktree, "local.txt", "local\n")
		main := git(repo, "rev-parse", "HEAD")
		opts.Strategy = branchrebase.StrategyMerge
		result, err := branchrebase.Run(context.Background(), opts)
		Expect(err).NotTo(HaveOccurred())
		Expect([]any{result.Phase, result.Strategy}).To(Equal([]any{"done", branchrebase.StrategyMerge}))
		Expect(git(worktree, "rev-parse", "HEAD^1", "HEAD^2")).To(Equal(head + "\n" + main))
		Expect(filepath.Join(worktree, "local.txt")).To(HaveContents("local\n"))
	})

	It("records a merge commit with noFastForward even when the branch could fast-forward", func() {
		git(worktree, "reset", "-q", "--hard", git(repo, "rev-parse", "HEAD~1"))
		opts.ExpectedHead = git(worktree, "rev-parse", "HEAD")
		opts.Strategy, opts.NoFastForward = branchrebase.StrategyMerge, true
		_, err := branchrebase.Run(context.Background(), opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(git(worktree, "rev-list", "--parents", "-n", "1", "HEAD")).To(HaveLen(3*40 + 2))
	})

	It("keeps a conflicting merge recoverable, then aborts it and restores the original commit and saved edits", func() {
		opts.ExpectedHead = commit(worktree, "shared file.txt", "topic conflict\n")
		commit(repo, "shared file.txt", "main conflict\n")
		write(worktree, "local.txt", "local\n")
		opts.Strategy = branchrebase.StrategyMerge
		result, err := branchrebase.Run(context.Background(), opts)
		Expect(err).To(HaveOccurred())
		Expect(result.Conflicts).To(Equal([]string{"shared file.txt"}))
		result, err = branchrebase.Run(context.Background(), branchrebase.Options{Action: "abort", Worktree: worktree, Operation: result.ID})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Phase).To(Equal("aborted"))
		Expect(git(worktree, "rev-parse", "HEAD")).To(Equal(opts.ExpectedHead))
		Expect(filepath.Join(worktree, "local.txt")).To(HaveContents("local\n"))
	})

	DescribeTable("rejects strategy options that do not apply before changing the worktree", func(strategy branchrebase.Strategy, rebaseMerges, noFastForward bool, message string) {
		opts.Strategy, opts.RebaseMerges, opts.NoFastForward = strategy, rebaseMerges, noFastForward
		_, err := branchrebase.Run(context.Background(), opts)
		Expect(err).To(MatchError(ContainSubstring(message)))
		Expect(git(worktree, "rev-parse", "HEAD")).To(Equal(head))
	},
		Entry("unknown strategy", branchrebase.Strategy("squash"), false, false, "strategy must be"),
		Entry("rebaseMerges with merge", branchrebase.StrategyMerge, true, false, "rebaseMerges applies only"),
		Entry("noFastForward with rebase", branchrebase.StrategyRebase, false, true, "noFastForward applies only"),
	)

	It("rejects a stale head before stashing or rebasing", func() {
		commit(worktree, "new.txt", "new\n")
		write(worktree, "local.txt", "local\n")
		_, err := branchrebase.Run(context.Background(), opts)
		Expect(err).To(MatchError(ContainSubstring("HEAD changed")))
		Expect(filepath.Join(worktree, "local.txt")).To(HaveContents("local\n"))
		Expect(git(worktree, "stash", "list")).To(BeEmpty())
	})

	DescribeTable("rejects option injection from persisted recovery metadata before running git", func(field string) {
		state := branchrebase.Operation{ID: "recovery", Worktree: worktree, Branch: "topic", OriginalHead: head, Target: "refs/heads/main", TargetSHA: head, Phase: "rebase"}
		switch field {
		case "target":
			state.TargetSHA = "--help"
		case "stash":
			state.StashSHA = "--help"
		case "original":
			state.OriginalHead = "--help"
		}
		raw, err := json.Marshal(state)
		Expect(err).NotTo(HaveOccurred())
		path := git(worktree, "rev-parse", "--path-format=absolute", "--git-path", "gavel-rebase.json")
		Expect(os.WriteFile(path, raw, 0600)).To(Succeed())
		_, err = branchrebase.Inspect(context.Background(), worktree)
		Expect(err).To(MatchError(ContainSubstring("invalid rebase recovery")))
	}, Entry("target", "target"), Entry("stash", "stash"), Entry("original", "original"))

	DescribeTable("refuses git worktrees that are not real directories", func(kind string) {
		path := ""
		if kind == "file" {
			path = filepath.Join(worktree, "topic.txt")
		}
		_, err := branchrebase.Inspect(context.Background(), path)
		Expect(err).To(HaveOccurred())
	}, Entry("empty", "empty"), Entry("file", "file"))

	DescribeTable("rejects invalid targets in recovery metadata before resuming", func(target string) {
		state := branchrebase.Operation{ID: "recovery", Worktree: worktree, Branch: "topic", OriginalHead: head, Target: target, TargetSHA: head, Phase: "rebase"}
		raw, err := json.Marshal(state)
		Expect(err).NotTo(HaveOccurred())
		path := git(worktree, "rev-parse", "--path-format=absolute", "--git-path", "gavel-rebase.json")
		Expect(os.WriteFile(path, raw, 0600)).To(Succeed())
		_, err = branchrebase.Inspect(context.Background(), worktree)
		Expect(err).To(HaveOccurred())
		Expect(git(worktree, "rev-parse", "HEAD")).To(Equal(head))
	},
		Entry("option", "--exec=unexpected"),
		Entry("revision expression", "refs/heads/main^{commit}"),
		Entry("control character", "refs/heads/main\n--exec=unexpected"),
		Entry("unapproved remote", "refs/remotes/other/main"),
	)

	It("rejects option-like targets before changing the worktree", func() {
		opts.Target = "--exec=unexpected"
		_, err := branchrebase.Run(context.Background(), opts)
		Expect(err).To(MatchError(ContainSubstring("target must be")))
		Expect(git(worktree, "rev-parse", "HEAD")).To(Equal(head))
	})

	It("retains a conflicting rebase and its stash across inspection, then aborts and restores edits", func() {
		opts.ExpectedHead = commit(worktree, "shared file.txt", "topic conflict\n")
		commit(repo, "shared file.txt", "main conflict\n")
		write(worktree, "local.txt", "local\n")
		result, err := branchrebase.Run(context.Background(), opts)
		Expect(err).To(HaveOccurred())
		Expect(result.Conflicts).To(Equal([]string{"shared file.txt"}))
		Expect(result.StashSHA).NotTo(BeEmpty())
		state, err := branchrebase.Inspect(context.Background(), worktree)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.Branch).To(BeEmpty())
		Expect(state.Operation.ID).To(Equal(result.ID))
		abort := branchrebase.Options{Action: "abort", Worktree: worktree, Operation: result.ID}
		result, err = branchrebase.Run(context.Background(), abort)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Phase).To(Equal("aborted"))
		Expect(git(worktree, "rev-parse", "HEAD")).To(Equal(opts.ExpectedHead))
		Expect(filepath.Join(worktree, "local.txt")).To(HaveContents("local\n"))
	})

	It("fetches origin/main before rebasing instead of using stale tracking refs", func() {
		remote := filepath.Join(GinkgoT().TempDir(), "origin.git")
		git(repo, "init", "--bare", remote)
		git(repo, "remote", "add", "origin", remote)
		git(repo, "push", "-qu", "origin", "main")
		fresh := commit(repo, "fresh.txt", "fresh\n")
		git(repo, "push", "-q", "origin", "main")
		opts.Target = "refs/remotes/origin/main"
		result, err := branchrebase.Run(context.Background(), opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.TargetSHA).To(Equal(fresh))
		git(worktree, "merge-base", "--is-ancestor", fresh, "HEAD")
	})

	It("preserves unrelated stashes and makes finalization idempotent", func() {
		write(worktree, "unrelated.txt", "unrelated\n")
		git(worktree, "stash", "push", "-u", "-m", "unrelated")
		unrelated := git(worktree, "rev-parse", "refs/stash")
		write(worktree, "local.txt", "local\n")
		result, err := branchrebase.Run(context.Background(), opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(git(worktree, "rev-parse", "refs/stash")).To(Equal(unrelated))
		_, err = branchrebase.Run(context.Background(), branchrebase.Options{Action: "finish", Worktree: worktree, Operation: result.ID})
		Expect(err).NotTo(HaveOccurred())
		state, err := branchrebase.Inspect(context.Background(), worktree)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.Operation).To(BeNil())
	})

	It("retains saved edits on restoration failure until explicitly resolved", func() {
		write(worktree, "shared file.txt", "saved local edit\n")
		commit(repo, "shared file.txt", "upstream edit\n")
		result, err := branchrebase.Run(context.Background(), opts)
		Expect(err).To(HaveOccurred())
		Expect(result.Phase).To(Equal("restore-conflict"))
		Expect(result.StashSHA).NotTo(BeEmpty())
		finish := branchrebase.Options{Action: "finish", Worktree: worktree, Operation: result.ID}
		_, err = branchrebase.Run(context.Background(), finish)
		Expect(err).To(MatchError(ContainSubstring("--resolved")))
		Expect(git(worktree, "stash", "list")).To(ContainSubstring("gavel-rebase:" + result.ID))
		write(worktree, "shared file.txt", "upstream and saved local edit\n")
		git(worktree, "add", "shared file.txt")
		finish.Resolved = true
		_, err = branchrebase.Run(context.Background(), finish)
		Expect(err).NotTo(HaveOccurred())
		Expect(git(worktree, "diff", "--cached", "--name-only")).To(Equal("shared file.txt"))
	})

	It("resumes a manually resolved rebase using the pinned target", func() {
		opts.ExpectedHead = commit(worktree, "shared file.txt", "topic conflict\n")
		commit(repo, "shared file.txt", "main conflict\n")
		result, err := branchrebase.Run(context.Background(), opts)
		Expect(err).To(HaveOccurred())
		write(worktree, "shared file.txt", "combined\n")
		git(worktree, "add", "shared file.txt")
		git(worktree, "-c", "core.editor=true", "rebase", "--continue")
		commit(repo, "later.txt", "later target\n")
		result, err = branchrebase.Run(context.Background(), branchrebase.Options{Action: "finish", Worktree: worktree, Operation: result.ID})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.TargetSHA).NotTo(Equal(git(repo, "rev-parse", "HEAD")))
		Expect(result.Phase).To(Equal("done"))
	})

	It("requires the correct operation ID and lists merged local targets", func() {
		git(repo, "branch", "already-merged", head+"~1")
		state, err := branchrebase.Inspect(context.Background(), worktree)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.Targets).To(ContainElements(branchrebase.Target{Ref: "refs/heads/main", Label: "main (main checkout)"}, branchrebase.Target{Ref: "refs/heads/already-merged", Label: "already-merged"}))
		opts.Target = "refs/remotes/origin/main"
		result, err := branchrebase.Run(context.Background(), opts)
		Expect(err).To(HaveOccurred())
		Expect(result.Phase).To(Equal("fetch"))
		_, err = branchrebase.Run(context.Background(), branchrebase.Options{Action: "abort", Worktree: worktree, Operation: "wrong"})
		Expect(err).To(HaveOccurred())
		_, err = branchrebase.Run(context.Background(), branchrebase.Options{Action: "abort", Worktree: worktree, Operation: result.ID})
		Expect(err).NotTo(HaveOccurred())
	})

	It("retires an operation without saved edits after its rebase was finished by hand and another branch checked out", func() {
		opts.ExpectedHead = commit(worktree, "shared file.txt", "topic conflict\n")
		commit(repo, "shared file.txt", "main conflict\n")
		result, err := branchrebase.Run(context.Background(), opts)
		Expect(err).To(HaveOccurred())
		Expect(result.StashSHA).To(BeEmpty())
		write(worktree, "shared file.txt", "combined\n")
		git(worktree, "add", "shared file.txt")
		git(worktree, "-c", "core.editor=true", "rebase", "--continue")
		rebased := git(worktree, "rev-parse", "HEAD")
		git(worktree, "checkout", "-qb", "other")
		result, err = branchrebase.Run(context.Background(), branchrebase.Options{Action: "abort", Worktree: worktree, Operation: result.ID})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Phase).To(Equal("aborted"))
		Expect(git(worktree, "branch", "--show-current")).To(Equal("other"))
		Expect(git(worktree, "rev-parse", "topic")).To(Equal(rebased))
		state, err := branchrebase.Inspect(context.Background(), worktree)
		Expect(err).NotTo(HaveOccurred())
		Expect(state.Operation).To(BeNil())
	})

	It("keeps saved edits and refuses to abort once HEAD moved past the original commit", func() {
		opts.ExpectedHead = commit(worktree, "shared file.txt", "topic conflict\n")
		commit(repo, "shared file.txt", "main conflict\n")
		write(worktree, "local.txt", "local\n")
		result, err := branchrebase.Run(context.Background(), opts)
		Expect(err).To(HaveOccurred())
		write(worktree, "shared file.txt", "combined\n")
		git(worktree, "add", "shared file.txt")
		git(worktree, "-c", "core.editor=true", "rebase", "--continue")
		_, err = branchrebase.Run(context.Background(), branchrebase.Options{Action: "abort", Worktree: worktree, Operation: result.ID})
		Expect(err).To(MatchError(And(ContainSubstring(result.StashSHA), ContainSubstring("finish"))))
		Expect(git(worktree, "stash", "list")).To(ContainSubstring("gavel-rebase:" + result.ID))
	})

	It("rebases a worktree whose only edits are submodule checkouts that git stash cannot save", func() {
		sub := GinkgoT().TempDir()
		git(sub, "init", "-qb", "main")
		git(sub, "config", "user.email", "test@example.com")
		git(sub, "config", "user.name", "Test")
		commit(sub, "lib.txt", "v1\n")
		git(repo, "-c", "protocol.file.allow=always", "submodule", "add", "-q", sub, "lib")
		git(repo, "commit", "-qm", "add submodule")
		git(worktree, "rebase", "-q", "main")
		git(worktree, "-c", "protocol.file.allow=always", "submodule", "update", "-q", "--init")
		opts.ExpectedHead = commit(worktree, "topic.txt", "topic again\n")
		lib := filepath.Join(worktree, "lib")
		git(lib, "config", "user.email", "test@example.com")
		git(lib, "config", "user.name", "Test")
		commit(lib, "lib.txt", "v2 checked out only\n")
		commit(repo, "after.txt", "after\n")
		before := git(worktree, "status", "--porcelain")
		Expect(before).To(ContainSubstring("lib"))
		result, err := branchrebase.Run(context.Background(), opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Phase).To(Equal("done"))
		Expect(result.StashSHA).To(BeEmpty())
		Expect(git(worktree, "status", "--porcelain")).To(Equal(before))
		git(worktree, "merge-base", "--is-ancestor", git(repo, "rev-parse", "HEAD"), "HEAD")
	})

	It("refuses to resume after another branch was checked out at the original HEAD", func() {
		opts.ExpectedHead = commit(worktree, "shared file.txt", "topic conflict\n")
		commit(repo, "shared file.txt", "main conflict\n")
		result, err := branchrebase.Run(context.Background(), opts)
		Expect(err).To(HaveOccurred())
		git(worktree, "rebase", "--abort")
		git(worktree, "checkout", "-qb", "other")
		_, err = branchrebase.Run(context.Background(), branchrebase.Options{Action: "start", Worktree: worktree, Operation: result.ID})
		Expect(err).To(HaveOccurred())
		Expect(git(worktree, "rev-parse", "HEAD")).To(Equal(opts.ExpectedHead))
		Expect(git(worktree, "branch", "--show-current")).To(Equal("other"))
	})
})
