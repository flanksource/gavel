package library_test

import (
	"io/fs"
	"strings"

	promptlib "github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts/library"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("branch-rebase.prompt", func() {
	It("uses the failed worktree operation and finalizes without landing or removing it", func() {
		tpl, err := promptlib.LoadFS(library.FS, "branch-rebase.prompt")
		Expect(err).NotTo(HaveOccurred())
		spec, _, err := tpl.Render(promptlib.RenderOptions{Declared: true, Data: map[string]any{
			"gavelExecutable": "/bin/gavel",
			"worktree":        "/src/acme's tree", "operation": "op-123", "branch": "topic", "target": "refs/heads/main", "targetSha": "abc", "phase": "rebase", "error": "conflict", "conflicts": []string{"shared file.txt"}, "stashSha": "saved",
		}})
		Expect(err).NotTo(HaveOccurred())
		Expect(spec.Setup.Cwd).To(Equal("/src/acme's tree"))
		Expect(spec.Model.Mode).To(Equal(api.ModeAgent))
		Expect(spec.Workflow.Verify.Commands).To(Equal([]string{`'/bin/gavel' git rebase check --worktree '/src/acme'\''s tree' --operation 'op-123'`}))
		Expect(spec.Workflow.After).To(HaveLen(1))
		Expect(spec.Workflow.After[0].Commands).To(Equal([]string{`'/bin/gavel' git rebase finish --worktree '/src/acme'\''s tree' --operation 'op-123'`}))
		Expect(spec.Workflow.After[0].EffectiveWhen()).To(Equal(api.CommitWhen("onVerify")))
		Expect(spec.Prompt.User).To(ContainSubstring("shared file.txt"))
		for _, denied := range []string{"Bash(git push:*)", "Bash(git merge:*)", "Bash(git reset:*)", "Bash(git worktree:*)", "Bash(git branch:*)", "Bash(git commit:*)"} {
			Expect(spec.Permissions.Tools).To(HaveKeyWithValue(denied, api.ToolPolicy("deny")))
		}
	})

	It("is discoverable with valid unrendered frontmatter", func() {
		raw, err := fs.ReadFile(library.FS, "branch-rebase.prompt")
		Expect(err).NotTo(HaveOccurred())
		var frontmatter map[string]any
		Expect(yaml.Unmarshal([]byte(strings.SplitN(string(raw), "---\n", 3)[1]), &frontmatter)).To(Succeed())
		Expect(frontmatter).To(HaveKeyWithValue("name", "branch-rebase"))
	})
})
