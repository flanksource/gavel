package gitstate_test

import (
	"path/filepath"
	"time"

	"github.com/flanksource/gavel/git/gitstate"
	"github.com/flanksource/gavel/internal/safefs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Missing locked worktrees", func() {
	It("loads healthy branches and records a missing detached checkout's error separately", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		missing := filepath.Join(filepath.Dir(r.root), "wt-initializing")
		git(r.root, "worktree", "add", "--detach", missing, r.base)
		head := commit(missing, "detached.txt", "historical snapshot\n", "test: detached snapshot")
		at, err := time.Parse(time.RFC3339, git(r.root, "log", "-1", "--format=%cI", head))
		Expect(err).NotTo(HaveOccurred())
		git(r.root, "worktree", "lock", "--reason", "initializing", missing)
		Expect(safefs.SafeRemove(missing, safefs.RemoveContext{AllowedRoot: filepath.Dir(missing)})).To(Succeed())

		state, err := r.tracker.State(ctx, r.root)

		Expect(err).NotTo(HaveOccurred())
		Expect(state.Error).To(BeEmpty())
		Expect(state.Branches).To(HaveLen(1))
		Expect(state.Branches[0].Name).To(Equal("feature"))
		Expect(state.Worktrees).To(HaveLen(3))
		Expect(state.Worktrees[0].StatusError).To(BeEmpty())
		Expect(state.Worktrees).To(ContainElement(And(
			HaveField("Path", missing), HaveField("Head", head),
			HaveField("LastCommitAt", at.UTC()),
			HaveField("StatusError", ContainSubstring("no such file or directory")),
		)))
	})
})
