package library_test

import (
	"os"
	"os/exec"
	"strings"

	promptlib "github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/prompts/library"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// gitIn runs git in dir with a fixed identity and fails the spec on error.
func gitIn(dir string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %s: %s", strings.Join(args, " "), out)
	return strings.TrimSpace(string(out))
}

// commitFile writes name and commits it, returning the new HEAD.
func commitFile(dir, name string) string {
	Expect(os.WriteFile(dir+"/"+name, []byte(name), 0o644)).To(Succeed())
	gitIn(dir, "add", name)
	gitIn(dir, "commit", "-m", name)
	return gitIn(dir, "rev-parse", "HEAD")
}

var _ = Describe("pr-rebase.prompt", func() {
	const (
		repo   = "/src/acme"
		branch = "topic's $(echo pwned)"
		head   = "1111111111111111111111111111111111111111"
	)

	render := func(baseSHA string) api.Spec {
		tpl, err := promptlib.LoadFS(library.FS, "pr-rebase.prompt")
		Expect(err).NotTo(HaveOccurred())
		req, _, err := tpl.Render(promptlib.RenderOptions{Declared: true, Data: map[string]any{
			"repo": repo, "remote": "origin", "number": 104, "url": "https://github.com/acme/widgets/pull/104",
			"branch": branch, "base": "main", "baseSha": baseSHA, "head": head,
		}})
		Expect(err).NotTo(HaveOccurred())
		return req
	}

	It("ships at the library root, so it resolves by bare name", func() {
		Expect(library.FS.Open("pr-rebase.prompt")).NotTo(BeNil())
	})

	It("starts the agent in a new worktree of the checkout at the PR head", func() {
		spec := render("2222222222222222222222222222222222222222")
		Expect(spec.Setup).NotTo(BeNil())
		Expect(spec.Setup.Checkout).NotTo(BeNil())
		Expect(spec.Setup.Checkout.Path).To(Equal(repo))
		Expect(*spec.Setup.Checkout.Worktree).To(Equal(shell.Worktree{
			Mode: shell.WorktreeNew, Prefix: "gavel/pr-104-rebase", Base: head,
			Uncommitted: shell.CloneSkip, Ignored: shell.CloneClone,
		}))
	})

	It("denies the agent every way to merge, push or abandon the rebase", func() {
		tools := render("2222222222222222222222222222222222222222").Permissions.Tools
		Expect(tools).To(HaveKeyWithValue("Bash(git rebase:*)", api.ToolPolicy("allow")))
		for _, denied := range []string{"Bash(git merge:*)", "Bash(git pull:*)", "Bash(git push:*)", "Bash(git rebase --abort:*)", "Bash(git rebase --rebase-merges:*)", "Bash(git reset:*)"} {
			Expect(tools).To(HaveKeyWithValue(denied, api.ToolPolicy("deny")), denied)
		}
	})

	It("force-pushes only after verify, leased to the head the rebase started from", func() {
		after := render("2222222222222222222222222222222222222222").Workflow.After
		Expect(after).To(HaveLen(1))
		Expect(after[0].EffectiveWhen()).To(Equal(api.CommitWhen("onVerify")))
		Expect(landingGit(after[0].Commands)).To(Equal([]string{
			"push --force-with-lease=refs/heads/topic's $(echo pwned):" + head + " origin HEAD:refs/heads/topic's $(echo pwned)",
		}))
	})

	DescribeTable("verification fails on merge commits and passes on a rebased branch",
		func(merged bool) {
			dir := GinkgoT().TempDir()
			gitIn(dir, "init", "-q", "-b", "main")
			commitFile(dir, "root")
			gitIn(dir, "switch", "-q", "-c", "topic")
			commitFile(dir, "topic")
			gitIn(dir, "switch", "-q", "main")
			base := commitFile(dir, "base")
			gitIn(dir, "switch", "-q", "topic")
			if merged {
				gitIn(dir, "merge", "-q", "--no-edit", "main")
			} else {
				gitIn(dir, "rebase", "-q", "main")
			}

			var failed []string
			for _, command := range render(base).Workflow.Verify.Commands {
				cmd := exec.Command("sh", "-c", command)
				cmd.Dir = dir
				if out, err := cmd.CombinedOutput(); err != nil {
					failed = append(failed, string(out))
				}
			}
			if merged {
				Expect(failed).To(ConsistOf(ContainSubstring("merge commits on the branch")))
			} else {
				Expect(failed).To(BeEmpty())
			}
		},
		Entry("base merged into the branch", true),
		Entry("branch rebased onto the base", false),
	)
})
