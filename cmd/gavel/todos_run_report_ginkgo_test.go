package main

import (
	captainapi "github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo run report workspace", func() {
	const (
		todoRef = "ae033218"
		setup   = "5e7a0000aaaa1111bbbb2222cccc3333dddd4444"
		head    = "d0b5c966aaaa1111bbbb2222cccc3333dddd4444"
	)
	report := func(workspace *captainapi.WorkspaceRecord) string {
		execution := &todos.ExecutionResult{ExecutorName: "cli-claude", Workspace: workspace}
		outcome := &lifecycle.StepOutcome{Step: lifecycle.Step{Name: "run"}, Execution: execution}
		return stripANSI(renderRunResult(todoRef, "run", outcome, string(types.StatusCompleted), nil))
	}

	It("prints the branch, Setup..Head, each commit and the land hint for a branch with commits", func() {
		out := report(&captainapi.WorkspaceRecord{
			Worktree: &captainapi.WorktreeState{Branch: "shell/289107d9", Setup: setup, Head: head, Removed: true},
			Commits:  []captainapi.CommitRecord{{SHA: head, Message: "feat(todos): record the workspace\n\nbody"}},
		})

		Expect(out).To(ContainSubstring("workspace: shell/289107d9  5e7a000..d0b5c96  worktree removed"))
		Expect(out).To(ContainSubstring("  d0b5c96 feat(todos): record the workspace"))
		Expect(out).To(ContainSubstring("land it: gavel todos land ae033218 --merge|--pr"))
	})

	It("names a kept worktree's path and reason", func() {
		out := report(&captainapi.WorkspaceRecord{Worktree: &captainapi.WorktreeState{
			Branch: "shell/1", Setup: setup, Head: setup, Path: "/repo/.worktrees/shell-1",
			Kept: true, KeptReason: "commit failed", Dirty: []string{"go.mod"},
		}})

		Expect(out).To(ContainSubstring("worktree kept at `/repo/.worktrees/shell-1`: commit failed (1 dirty paths)"))
		Expect(out).NotTo(ContainSubstring("land it:"))
	})

	It("offers no land hint once teardown deleted the branch", func() {
		out := report(&captainapi.WorkspaceRecord{
			Worktree: &captainapi.WorktreeState{Branch: "shell/2", Setup: setup, Head: head, Removed: true, BranchDeleted: true},
			Commits:  []captainapi.CommitRecord{{SHA: head, Message: "fix: x"}},
		})

		Expect(out).NotTo(ContainSubstring("land it:"))
	})

	It("prints no workspace block for a run that reported none", func() {
		Expect(report(nil)).NotTo(ContainSubstring("workspace:"))
	})
})
