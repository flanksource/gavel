package library_test

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	promptlib "github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts/library"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

// landingGit runs rendered after-commands through sh with git stubbed out and
// returns the git invocations they made (args joined by spaces), or "exit N"
// for a command that failed — so the shell branching and quoting are exercised
// as captain will run them, without touching a repository.
func landingGit(commands []string) []string {
	log := filepath.Join(GinkgoT().TempDir(), "git.log")
	var calls []string
	for _, command := range commands {
		cmd := exec.Command("sh", "-c", `git() { printf '%s\n' "$*" >> "$GIT_LOG"; }; `+command)
		cmd.Env = append(os.Environ(), "GIT_LOG="+log)
		if err := cmd.Run(); err != nil {
			var exit *exec.ExitError
			Expect(errors.As(err, &exit)).To(BeTrue(), "%s: %v", command, err)
			return append(calls, fmt.Sprintf("exit %d", exit.ExitCode()))
		}
		calls = nil
		if raw, err := os.ReadFile(log); err == nil {
			calls = strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
		}
	}
	return calls
}

var _ = Describe("branch-merge.prompt", func() {
	const (
		repo     = "/src/acme"
		worktree = "/src/acme/.worktrees/topic"
		branch   = "topic's $(echo pwned)"
		base     = "main"
	)

	render := func(vars map[string]any) api.Spec {
		tpl, err := promptlib.LoadFS(library.FS, "branch-merge.prompt")
		Expect(err).NotTo(HaveOccurred())
		data := map[string]any{"repo": repo, "worktree": worktree, "branch": branch, "base": base, "conflicts": []string{"go.mod", "cmd/main.go"}}
		for k, v := range vars {
			data[k] = v
		}
		req, _, err := tpl.Render(promptlib.RenderOptions{Data: data, Declared: true})
		Expect(err).NotTo(HaveOccurred())
		return req
	}

	It("ships at the library root, so it resolves by bare name", func() {
		names, err := fs.Glob(library.FS, "*.prompt")
		Expect(err).NotTo(HaveOccurred())
		Expect(names).To(ContainElement("branch-merge.prompt"))
	})

	It("starts the agent in the worktree with edits allowed and landing commands denied", func() {
		spec := render(map[string]any{"mode": "squash"})
		Expect(spec.Setup).NotTo(BeNil())
		Expect(spec.Setup.Cwd).To(Equal(worktree))
		Expect(string(spec.Permissions.Mode)).To(Equal("acceptEdits"))
		Expect(spec.Permissions.Tools).To(HaveKeyWithValue("Bash(git rebase:*)", api.ToolPolicy("allow")))
		for _, denied := range []string{"Bash(git push:*)", "Bash(git merge:*)", "Bash(git rebase --abort:*)", "Bash(git reset:*)", "Bash(git worktree:*)"} {
			Expect(spec.Permissions.Tools).To(HaveKeyWithValue(denied, api.ToolPolicy("deny")), denied)
		}
	})

	It("verifies a finished, clean rebase onto the quoted base", func() {
		spec := render(map[string]any{"mode": "squash"})
		Expect(spec.Workflow.Verify.Commands).To(ContainElements(
			"git merge-base --is-ancestor 'main' HEAD",
			"git diff --check 'main'...HEAD",
		))
	})

	It("keeps its frontmatter valid YAML before rendering, so listings parse it", func() {
		raw, err := fs.ReadFile(library.FS, "branch-merge.prompt")
		Expect(err).NotTo(HaveOccurred())
		parts := strings.SplitN(string(raw), "---\n", 3)
		Expect(parts).To(HaveLen(3))
		var frontmatter map[string]any
		Expect(yaml.Unmarshal([]byte(parts[1]), &frontmatter)).To(Succeed())
		Expect(frontmatter).To(HaveKey("workflow"))
	})

	It("lands from the main checkout only after verify, with every user value shell-quoted", func() {
		after := render(map[string]any{"mode": "squash", "message": "fix: it's # done"}).Workflow.After
		Expect(after).To(HaveLen(1))
		Expect(after[0].EffectiveWhen()).To(Equal(api.CommitWhen("onVerify")))
		Expect(after[0].Cwd).To(Equal(repo))
		Expect(after[0].Commands).To(HaveLen(4))
		Expect(after[0].Commands[0]).To(ContainSubstring(`squash) git merge --squash 'topic'\''s $(echo pwned)' ;;`))
		Expect(after[0].Commands[1]).To(ContainSubstring(`git commit -m 'fix: it'\''s # done'`))
		Expect(after[0].Commands[2:]).To(Equal([]string{
			`git worktree remove '/src/acme/.worktrees/topic'`,
			`git branch -D 'topic'\''s $(echo pwned)'`,
		}))
	})

	DescribeTable("lands each mode with the matching git commands",
		func(vars map[string]any, expected []string) {
			Expect(landingGit(render(vars).Workflow.After[0].Commands[:2])).To(Equal(expected))
		},
		Entry("squash with a message", map[string]any{"mode": "squash", "message": "feat: x"},
			[]string{"merge --squash topic's $(echo pwned)", "commit -m feat: x"}),
		Entry("squash without a message keeps git's squash message", map[string]any{"mode": "squash"},
			[]string{"merge --squash topic's $(echo pwned)", "commit --no-edit"}),
		Entry("incremental fast-forwards the rebased commits", map[string]any{"mode": "incremental", "message": "ignored"},
			[]string{"merge --ff-only topic's $(echo pwned)"}),
	)

	It("refuses an unknown mode instead of landing", func() {
		Expect(landingGit(render(map[string]any{"mode": "rebase"}).Workflow.After[0].Commands[:1])).To(Equal([]string{"exit 1"}))
	})

	It("lists the files the failed land conflicted on", func() {
		tpl, err := promptlib.LoadFS(library.FS, "branch-merge.prompt")
		Expect(err).NotTo(HaveOccurred())
		req, _, err := tpl.Render(promptlib.RenderOptions{Declared: true, Data: map[string]any{
			"repo": repo, "worktree": worktree, "branch": "topic", "base": base, "mode": "squash", "conflicts": []string{"go.mod"},
		}})
		Expect(err).NotTo(HaveOccurred())
		Expect(req.Prompt.User).To(ContainSubstring("- `go.mod`"))
		Expect(req.Prompt.User).To(ContainSubstring("git rebase main"))
	})
})
