package gitstate_test

import (
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/git/gitstate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Project git summary", func() {
	DescribeTable("counts branches with uncommitted files once per branch", func(changes gitstate.Changes) {
		state := gitstate.State{
			Base: "main",
			Branches: []gavelgit.BranchInfo{
				{Name: "feature", Ahead: 2, Diff: gavelgit.DiffStat{Adds: 4, Dels: 1}},
				{Name: "unchecked", Ahead: 1},
			},
			Worktrees: []gitstate.Worktree{
				{Worktree: gavelgit.Worktree{Branch: "main", Primary: true}, Changes: changes},
				{Worktree: gavelgit.Worktree{Branch: "feature"}, Changes: changes},
				{Worktree: gavelgit.Worktree{Branch: "feature"}, Changes: changes},
				{Worktree: gavelgit.Worktree{Branch: "clean"}},
			},
		}

		Expect(state.Summary("acme")).To(Equal(gitstate.Summary{
			Name: "acme", Base: "main", Branches: 2, Worktrees: 3, UncommittedBranches: 2,
		}))
	},
		Entry("staged files", gitstate.Changes{Staged: 3}),
		Entry("unstaged files", gitstate.Changes{Unstaged: 3}),
		Entry("staged and unstaged files", gitstate.Changes{Both: 3}),
		Entry("untracked files", gitstate.Changes{Untracked: 3}),
		Entry("conflicted files", gitstate.Changes{Conflict: 3}),
	)

	It("does not count clean, detached or prunable worktrees as uncommitted branches", func() {
		state := gitstate.State{Base: "main", Worktrees: []gitstate.Worktree{
			{Worktree: gavelgit.Worktree{Branch: "main", Primary: true}},
			{Worktree: gavelgit.Worktree{Detached: true}, Changes: gitstate.Changes{Untracked: 1}},
			{Worktree: gavelgit.Worktree{Branch: "removed", Prunable: true}, Changes: gitstate.Changes{Unstaged: 1}},
		}}

		Expect(state.Summary("acme").UncommittedBranches).To(BeZero())
	})

	It("counts lines of the primary checkout's uncommitted changes only, not of linked worktrees or unmerged branches", func() {
		state := gitstate.State{
			Base:     "main",
			Branches: []gavelgit.BranchInfo{{Name: "feature", Ahead: 2, Diff: gavelgit.DiffStat{Adds: 40, Dels: 10}}},
			Worktrees: []gitstate.Worktree{
				{Worktree: gavelgit.Worktree{Branch: "main", Primary: true}, Changes: gitstate.Changes{Unstaged: 1, Adds: 7, Dels: 2}},
				{Worktree: gavelgit.Worktree{Branch: "feature"}, Changes: gitstate.Changes{Unstaged: 1, Adds: 300, Dels: 100}},
			},
		}

		Expect(state.Summary("acme")).To(Equal(gitstate.Summary{
			Name: "acme", Base: "main", Adds: 7, Dels: 2, Worktrees: 1, Branches: 1, UncommittedBranches: 2,
		}))
	})

	It("preserves line totals while counting a dirty branch once", func() {
		state := gitstate.State{Base: "main", Worktrees: []gitstate.Worktree{
			{Worktree: gavelgit.Worktree{Branch: "main", Primary: true}, Changes: gitstate.Changes{Both: 2, Adds: 5, Dels: 3}},
		}}

		Expect(state.Summary("acme")).To(Equal(gitstate.Summary{
			Name: "acme", Base: "main", Adds: 5, Dels: 3, UncommittedBranches: 1,
		}))
	})

	It("surfaces failed worktree scans instead of presenting stale counts", func() {
		state := gitstate.State{Base: "main", Worktrees: []gitstate.Worktree{
			{Worktree: gavelgit.Worktree{Path: "/work/feature", Branch: "feature"},
				Changes: gitstate.Changes{Unstaged: 1}, StatusError: "index is unreadable"},
		}}

		Expect(state.Summary("acme")).To(Equal(gitstate.Summary{
			Name: "acme", Base: "main", Error: "worktree /work/feature: index is unreadable",
		}))
	})

	It("surfaces a failed ref scan without presenting stale counts", func() {
		state := gitstate.State{Base: "main", Error: "refs are unreadable", Branches: []gavelgit.BranchInfo{{Name: "feature"}}}

		Expect(state.Summary("acme")).To(Equal(gitstate.Summary{Name: "acme", Base: "main", Error: state.Error}))
	})
})
