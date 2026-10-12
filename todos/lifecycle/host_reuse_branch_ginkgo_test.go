package lifecycle_test

import (
	"context"
	"path/filepath"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Host Resolve with a reused branch", func() {
	const (
		runBranch = "shell/0123abcd"
		setupSHA  = "1111111aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		headSHA   = "2222222bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	)
	var (
		provider *fakeProvider
		host     *lifecycle.Host
		ctx      context.Context
		run      lifecycle.Step
		// request sets the worktree keep and ignored settings a reuse preserves.
		request api.Spec
	)

	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		provider = &fakeProvider{plan: todos.PlanState{Exists: true, Approved: true, Content: "# plan", Revision: 2}}
		host = newHost(provider)
		ctx = context.Background()
		run = stepNamed(host.Def, "run")
		request = api.Spec{Setup: &shell.Setup{Checkout: &shell.Checkout{Worktree: &shell.Worktree{
			Keep: true, Ignored: shell.CloneSkip,
		}}}}
		provider.runWorktree = &native.RunWorktree{PromptRunID: uuid.New(), Worktree: api.WorktreeState{
			Repo: host.WorkDir, Path: filepath.Join(host.WorkDir, ".shell", "worktrees", "run"), Branch: runBranch,
			Base: setupSHA, Setup: setupSHA, Head: headSHA, Removed: true,
		}}
	})

	resolve := func() (*lifecycle.Resolution, error) {
		return host.Resolve(ctx, hostTodo(), run, lifecycle.RunOptions{ReuseBranch: true, Request: request})
	}
	checkoutOf := func(worktree *shell.Worktree) *shell.Checkout {
		return &shell.Checkout{Mode: shell.CheckoutLocal, Worktree: worktree}
	}

	It("checks the removed run's branch out as it is, keeping the step's keep and ignored settings", func() {
		resolution, err := resolve()

		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Setup.Checkout).To(Equal(checkoutOf(&shell.Worktree{
			Mode: shell.WorktreeBranch, Branch: runBranch, Keep: true, Ignored: shell.CloneSkip,
		})))
		Expect(resolution.Spec.Prompt.AppendSystem).To(Equal(
			"This run continues on branch `" + runBranch + "`, which already holds the previous attempt's commits (1111111..2222222). " +
				"Build on them, and address every unresolved review comment listed in the task."))
		Expect(resolution.Prompt).NotTo(ContainSubstring(runBranch), "the branch notice is run framing, not task text")
	})

	It("appends the branch notice after a system prompt the step already appends", func() {
		const stepRule = "Never edit generated files."
		request.Prompt.AppendSystem = stepRule

		resolution, err := resolve()

		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Prompt.AppendSystem).To(HavePrefix(stepRule + "\n\nThis run continues on branch `" + runBranch + "`"))
	})

	It("continues in the worktree a run kept, when its directory still exists", func() {
		kept := GinkgoT().TempDir()
		provider.runWorktree.Worktree.Removed = false
		provider.runWorktree.Worktree.Kept = true
		provider.runWorktree.Worktree.Path = kept

		resolution, err := resolve()

		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Setup.Checkout).To(Equal(checkoutOf(&shell.Worktree{
			Mode: shell.WorktreeExisting, Path: kept, Keep: true, Ignored: shell.CloneSkip,
		})))
	})

	It("checks the branch out again when the kept worktree's directory is gone", func() {
		provider.runWorktree.Worktree.Removed = false
		provider.runWorktree.Worktree.Kept = true
		provider.runWorktree.Worktree.Path = filepath.Join(GinkgoT().TempDir(), "deleted")

		resolution, err := resolve()

		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Setup.Checkout).To(Equal(checkoutOf(&shell.Worktree{
			Mode: shell.WorktreeBranch, Branch: runBranch, Keep: true, Ignored: shell.CloneSkip,
		})))
	})

	DescribeTable("refuses a todo with no reusable branch",
		func(mutate func()) {
			mutate()

			_, err := resolve()

			Expect(err).To(MatchError(ContainSubstring("todo " + hostIssueID + " has no reusable branch")))
		},
		Entry("no run recorded a worktree", func() { provider.runWorktree = nil }),
		Entry("the run's branch was deleted", func() { provider.runWorktree.Worktree.BranchDeleted = true }),
	)

	It("refuses to reuse a branch for a step that is not a run", func() {
		_, err := host.Resolve(ctx, hostTodo(), stepNamed(host.Def, "plan"), lifecycle.RunOptions{ReuseBranch: true})

		Expect(err).To(MatchError(ContainSubstring("reusing a run's branch applies to a run step")))
	})

	It("leaves a run that does not reuse a branch on a fresh worktree", func() {
		resolution, err := host.Resolve(ctx, hostTodo(), run, lifecycle.RunOptions{Request: request})

		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Setup.Checkout.Worktree.Mode).To(Equal(shell.WorktreeNew))
		Expect(resolution.Spec.Prompt.AppendSystem).To(BeEmpty())
	})
})
