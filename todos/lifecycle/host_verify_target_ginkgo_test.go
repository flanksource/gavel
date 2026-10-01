package lifecycle_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// gitIn runs git in dir under a fixed identity and returns its trimmed output.
func gitIn(dir string, args ...string) string {
	GinkgoHelper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=gavel", "-c", "user.email=gavel@example.com", "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %s: %s", strings.Join(args, " "), out)
	return strings.TrimSpace(string(out))
}

// commitFile writes one file and commits it on the checked-out branch.
func commitFile(dir, name, content string) string {
	GinkgoHelper()
	Expect(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)).To(Succeed())
	gitIn(dir, "add", name)
	gitIn(dir, "commit", "-m", "add "+name)
	return gitIn(dir, "rev-parse", "HEAD")
}

var _ = Describe("Host VerifyStep target", func() {
	const (
		runBranch = "shell/0123abcd"
		// The definition of done passes only where the run's file exists: in the
		// run's commit, never in the main checkout it branched from.
		runFile = "feature.txt"
		prURL   = "https://github.com/acme/widgets/pull/7"
	)
	var (
		provider *fakeProvider
		host     *lifecycle.Host
		ctx      context.Context
		repo     string
		base     string
		runHead  string
		todo     *types.TODO
	)

	verifyCheckoutAt := func(sha string) *shell.Checkout {
		return &shell.Checkout{Mode: shell.CheckoutLocal, Worktree: &shell.Worktree{
			Mode: shell.WorktreeNew, Prefix: "gavel-verify", Base: sha,
			Uncommitted: shell.CloneSkip, Ignored: shell.CloneClone,
		}}
	}
	expectNoVerifyWorktreeLeft := func() {
		GinkgoHelper()
		Expect(gitIn(repo, "worktree", "list", "--porcelain")).NotTo(ContainSubstring("gavel-verify"))
		Expect(gitIn(repo, "branch", "--list", "gavel-verify/*")).To(BeEmpty())
	}

	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		provider = &fakeProvider{plan: todos.PlanState{Exists: true, Approved: true}}
		host = newHost(provider)
		ctx = context.Background()

		repo = host.WorkDir
		gitIn(repo, "init", "-b", "main")
		base = commitFile(repo, "README.md", "widgets\n")
		gitIn(repo, "checkout", "-b", runBranch)
		runHead = commitFile(repo, runFile, "the run's work\n")
		gitIn(repo, "checkout", "main")

		provider.runWorktree = &native.RunWorktree{PromptRunID: uuid.New(), Worktree: api.WorktreeState{
			Repo: repo, Path: filepath.Join(repo, ".shell", "worktrees", "run"), Branch: runBranch,
			Base: base, Setup: base, Head: runHead, Removed: true,
		}}
		todo = hostTodo()
		todo.VerificationMarkdown = "### command: the run's file exists\n\n```bash\ntest -f " + runFile + "\n```"
	})

	It("verify step checks out the last run's head in a fresh worktree", func() {
		resolution, err := host.Resolve(ctx, todo, stepNamed(host.Def, lifecycle.StepVerify), lifecycle.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Setup.Checkout).To(Equal(verifyCheckoutAt(runHead)))

		result, err := host.VerifyStep(ctx, todo, api.Spec{})

		Expect(err).NotTo(HaveOccurred())
		Expect(result.AllPassed).To(BeTrue(), "the fixture ran against the run's commit, which has %s: %s", runFile, result.TestResult.Output)
		Expect(result.TestResult.CWD).To(ContainSubstring("gavel-verify"), "the check ran in the verify worktree")
		expectNoVerifyWorktreeLeft()
	})

	It("reports the verified sha and branch, not the main checkout's", func() {
		result, err := host.VerifyStep(ctx, todo, api.Spec{})

		Expect(err).NotTo(HaveOccurred())
		Expect(result.Target).To(Equal(&types.VerifyTarget{Branch: runBranch, SHA: runHead, Source: types.VerifyTargetRun}))
		Expect([]string{result.TestResult.GitBranch, result.TestResult.GitCommit}).To(Equal([]string{runBranch, runHead[:7]}))
		Expect(result.Pretty().String()).To(ContainSubstring("verified run " + runBranch + " @ " + runHead[:7]))
	})

	It("a PR-landed todo verifies the PR topic head", func() {
		gitIn(repo, "checkout", "-b", "feat/widgets", base)
		topicHead := commitFile(repo, runFile, "the run's work, cherry-picked\n")
		gitIn(repo, "checkout", "main")
		prNumber := 7
		provider.runWorktree.Landing = &native.RunLanding{
			PromptRunID: provider.runWorktree.PromptRunID, Via: native.LandingPR, TargetBranch: "main",
			LandedSHA: topicHead, PRNumber: &prNumber, PRURL: prURL,
		}

		resolution, err := host.Resolve(ctx, todo, stepNamed(host.Def, lifecycle.StepVerify), lifecycle.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Setup.Checkout).To(Equal(verifyCheckoutAt(topicHead)))

		result, err := host.VerifyStep(ctx, todo, api.Spec{})

		Expect(err).NotTo(HaveOccurred())
		Expect(result.AllPassed).To(BeTrue())
		Expect(result.Target).To(Equal(&types.VerifyTarget{SHA: topicHead, Source: types.VerifyTargetPR, PR: prURL}))
		Expect(result.Pretty().String()).To(ContainSubstring("verified pr " + prURL + " @ " + topicHead[:7]))
		expectNoVerifyWorktreeLeft()
	})

	It("a merge-landed todo verifies the checkout", func() {
		gitIn(repo, "merge", "--ff-only", runBranch)
		provider.runWorktree.Landing = &native.RunLanding{
			PromptRunID: provider.runWorktree.PromptRunID, Via: native.LandingMerge, TargetBranch: "main", LandedSHA: runHead,
		}

		resolution, err := host.Resolve(ctx, todo, stepNamed(host.Def, lifecycle.StepVerify), lifecycle.RunOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Setup.Checkout).To(BeNil())

		result, err := host.VerifyStep(ctx, todo, api.Spec{})

		Expect(err).NotTo(HaveOccurred())
		Expect(result.AllPassed).To(BeTrue())
		Expect(result.Target).To(Equal(&types.VerifyTarget{Branch: "main", SHA: gitIn(repo, "rev-parse", "--short", "HEAD"), Source: types.VerifyTargetCheckout}))
	})

	It("verifies the checkout of a todo no run recorded a worktree for", func() {
		provider.runWorktree = nil

		resolution, err := host.Resolve(ctx, todo, stepNamed(host.Def, lifecycle.StepVerify), lifecycle.RunOptions{})

		Expect(err).NotTo(HaveOccurred())
		Expect(resolution.Spec.Setup.Checkout).To(BeNil())
		Expect(resolution.VerifyTarget).To(Equal(&types.VerifyTarget{Source: types.VerifyTargetCheckout}))
	})

	// Verifying the head of a worktree that still holds uncommitted edits judges
	// work without them: a verdict on something the run did not produce.
	It("refuses a run whose worktree was kept with uncommitted edits", func() {
		provider.runWorktree.Worktree.Removed = false
		provider.runWorktree.Worktree.Kept = true
		provider.runWorktree.Worktree.KeptReason = "uncommitted changes"
		provider.runWorktree.Worktree.Dirty = []string{"main.go"}

		_, err := host.Resolve(ctx, todo, stepNamed(host.Def, lifecycle.StepVerify), lifecycle.RunOptions{})

		Expect(err).To(MatchError(And(ContainSubstring(runBranch), ContainSubstring("main.go"))))
	})

	It("rejects a verify step spec that declares setup.checkout", func() {
		request := api.Spec{Setup: &shell.Setup{Checkout: &shell.Checkout{Mode: shell.CheckoutLocal}}}

		_, err := host.Resolve(ctx, todo, stepNamed(host.Def, lifecycle.StepVerify), lifecycle.RunOptions{Request: request})

		Expect(err).To(MatchError(ContainSubstring("setup.checkout")))
	})
})
