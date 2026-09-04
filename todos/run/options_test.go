package run_test

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The dashboard badges and the CLI's refusals read the resolved spec through
// these predicates, so each one is pinned to the spec shape that means it.
var _ = Describe("run option predicates", func() {
	worktree := func(uncommitted shell.CloneMode) api.Spec {
		return api.Spec{Setup: &shell.Setup{Checkout: &shell.Checkout{Worktree: &shell.Worktree{
			Mode: shell.WorktreeNew, Uncommitted: uncommitted,
		}}}}
	}

	Describe("Dirty", func() {
		It("is the worktree carrying uncommitted work across", func() {
			Expect(run.Dirty(worktree(shell.CloneClone))).To(BeTrue())
		})

		It("is false for a worktree that leaves uncommitted work behind", func() {
			Expect(run.Dirty(worktree(""))).To(BeFalse())
		})

		It("is false without a worktree: the run already happens in the dirty tree", func() {
			Expect(run.Dirty(api.Spec{})).To(BeFalse())
			Expect(run.Dirty(api.Spec{Setup: &shell.Setup{}})).To(BeFalse())
			Expect(run.Dirty(api.Spec{Setup: &shell.Setup{Checkout: &shell.Checkout{}}})).To(BeFalse())
		})
	})

	Describe("DryRun", func() {
		commits := func(dry ...bool) api.Spec {
			stanzas := make([]api.Commit, 0, len(dry))
			for _, d := range dry {
				stanzas = append(stanzas, api.Commit{On: api.CommitOnRun, DryRun: d})
			}
			return api.Spec{Workflow: &api.Workflow{Commits: stanzas}}
		}

		It("is true only when every commit stanza is dry", func() {
			Expect(run.DryRun(commits(true, true))).To(BeTrue())
		})

		It("is false for a run that mixes dry and live stanzas", func() {
			Expect(run.DryRun(commits(true, false))).To(BeFalse())
		})

		It("is false for a run that commits nothing at all", func() {
			Expect(run.DryRun(api.Spec{})).To(BeFalse())
			Expect(run.DryRun(commits())).To(BeFalse())
		})
	})

	Describe("SessionIDFor", func() {
		const prior = "session-prior"
		withPrior := &types.TODO{TODOFrontmatter: types.TODOFrontmatter{LLM: &types.LLM{SessionId: prior}}}

		It("reuses the todo's prior session on resume", func() {
			Expect(run.SessionIDFor(api.Spec{Model: api.Model{Mode: api.ModeCmux}}, withPrior, true)).To(Equal(prior))
		})

		It("mints a fresh id for a cmux run that is not a resume", func() {
			id := run.SessionIDFor(api.Spec{Model: api.Model{Mode: api.ModeCmux}}, withPrior, false)

			Expect(id).ToNot(BeEmpty())
			Expect(id).ToNot(Equal(prior))
		})

		It("reuses the known session for every other runtime, and none when there is none", func() {
			Expect(run.SessionIDFor(api.Spec{}, withPrior, false)).To(Equal(prior))
			Expect(run.SessionIDFor(api.Spec{}, &types.TODO{}, false)).To(BeEmpty())
			Expect(run.SessionIDFor(api.Spec{}, nil, true)).To(BeEmpty())
		})
	})

	Describe("IsStoppable", func() {
		It("excludes cmux, whose surface outlives the request", func() {
			Expect(run.IsStoppable(api.Spec{Model: api.Model{Mode: api.ModeCmux}})).To(BeFalse())
		})

		It("includes every runtime that owns its agent process", func() {
			Expect(run.IsStoppable(api.Spec{})).To(BeTrue())
			Expect(run.IsStoppable(api.Spec{Model: api.Model{Mode: api.ModeCLI}})).To(BeTrue())
		})
	})
})
