package runtime

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("renderAttempt workspace", func() {
	const (
		setup = "5e7a0000aaaa1111bbbb2222cccc3333dddd4444"
		head  = "4eadbeef0000111122223333444455556666777a"
		first = "c0ffee00111122223333444455556666777788aa"
	)
	attempt := func(workspace *api.WorkspaceRecord) string {
		return renderAttempt(&types.TODO{TODOFrontmatter: types.TODOFrontmatter{Attempts: 1}},
			&todos.ExecutionResult{Success: true, Workspace: workspace})
	}

	It("prints the branch, the agent's Setup..Head range, a removed worktree and one line per commit", func() {
		body := attempt(&api.WorkspaceRecord{
			Worktree: &api.WorktreeState{Branch: "shell/1a2b", Base: "ba5e", Setup: setup, Head: head, Removed: true},
			Commits: []api.CommitRecord{
				{SHA: first, Message: "feat(todos): first\n\nGavel-Issue-Id: 42"},
				{SHA: head, Message: "fix: second"},
			},
		})

		Expect(body).To(ContainSubstring("- **Branch:** `shell/1a2b`\n" +
			"- **Range:** `5e7a000..4eadbee`\n" +
			"- **Worktree:** removed\n" +
			"- **Commit:** `c0ffee0` feat(todos): first\n" +
			"- **Commit:** `4eadbee` fix: second\n"))
	})

	It("names the path and reason of a kept worktree with its dirty count", func() {
		body := attempt(&api.WorkspaceRecord{Worktree: &api.WorktreeState{
			Branch: "shell/9", Setup: setup, Head: setup, Path: "/repo/.worktrees/shell-9",
			Kept: true, KeptReason: "uncommitted changes", Dirty: []string{"a.go", "b.go"},
		}})

		Expect(body).To(ContainSubstring("- **Worktree:** kept at `/repo/.worktrees/shell-9`: uncommitted changes (2 dirty paths)\n"))
		Expect(body).NotTo(ContainSubstring("**Commit:**"))
	})

	It("says the branch was deleted when teardown dropped an empty branch", func() {
		body := attempt(&api.WorkspaceRecord{Worktree: &api.WorktreeState{
			Branch: "shell/7", Setup: setup, Head: setup, Removed: true, BranchDeleted: true,
		}})

		Expect(body).To(ContainSubstring("- **Worktree:** removed; branch deleted\n"))
	})

	It("prints nothing about a workspace when the run reported none", func() {
		Expect(attempt(nil)).NotTo(SatisfyAny(ContainSubstring("**Branch:**"), ContainSubstring("**Worktree:**")))
	})
})
