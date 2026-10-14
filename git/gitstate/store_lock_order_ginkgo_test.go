package gitstate_test

import (
	"context"
	"fmt"
	"sync"

	"github.com/flanksource/commons-db/dbtest"
	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/git/gitstate"
	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// A commit moves a branch and rewrites its worktree's index at once, so the
// watcher runs a refs scan and that worktree's status scan side by side.
var _ = Describe("store writes from concurrent refs and status scans", func() {
	It("never deadlock, so neither scan is lost", func(ctx SpecContext) {
		handle := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_gitstate"})
		GinkgoT().Setenv(database.EnvDSN, handle.DSN())
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		store := gitstate.NewStore(opened.Gorm())

		const root, feature = "/repo", "/wt-feature"
		id, err := store.EnsureRepo(ctx, root, root+"/.git")
		Expect(err).NotTo(HaveOccurred())
		refs := func(head string) gitstate.RefsUpdate {
			return gitstate.RefsUpdate{
				Fingerprint: head, BaseBranch: "main", BaseSHA: "base", CurrentBranch: "main", BaseCheckedOut: true,
				Worktrees: []gitstate.WorktreeRef{
					{Worktree: gavelgit.Worktree{Path: root, Branch: "main", Head: "base", Primary: true}},
					{Worktree: gavelgit.Worktree{Path: feature, Branch: "feature", Head: head}},
				},
				Branches: []gavelgit.BranchRef{{Name: "main", Head: "base"}, {Name: "feature", Head: head, Worktree: feature}},
			}
		}
		Expect(store.ApplyRefs(ctx, id, refs("head-0"))).To(Succeed())

		const rounds = 100
		for round := range rounds {
			var wg sync.WaitGroup
			start := make(chan struct{})
			errs := make([]error, 2)
			wg.Add(2)
			go func() {
				defer wg.Done()
				<-start
				_, errs[0] = store.Locked(context.Background(), "git_refs:"+id.String(), true, func(tx *gitstate.Store) error {
					return tx.ApplyRefs(context.Background(), id, refs(fmt.Sprintf("head-%d", round+1)))
				})
			}()
			go func() {
				defer wg.Done()
				<-start
				_, errs[1] = store.Locked(context.Background(), "git_status:"+id.String()+":"+feature, true, func(tx *gitstate.Store) error {
					return tx.ApplyStatus(context.Background(), id, feature, gitstate.StatusUpdate{Fingerprint: fmt.Sprint(round)})
				})
			}()
			close(start)
			wg.Wait()
			Expect(errs[0]).NotTo(HaveOccurred(), "refs scan, round %d", round)
			Expect(errs[1]).NotTo(HaveOccurred(), "status scan, round %d", round)
		}
	})
})
