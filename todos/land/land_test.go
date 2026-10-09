package land_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/git/branchmerge"
	prcreate "github.com/flanksource/gavel/pr/create"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	"github.com/flanksource/gavel/todos/land"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Land", func() {
	var f *landFixture

	BeforeEach(func(ctx SpecContext) { f = newLandFixture(ctx) })

	landings := func(ctx context.Context) []native.RunLanding {
		GinkgoHelper()
		issueID, err := uuid.Parse(f.todo.ID)
		Expect(err).NotTo(HaveOccurred())
		recorded, err := f.provider.Repository().ListLandings(ctx, issueID)
		Expect(err).NotTo(HaveOccurred())
		return recorded
	}

	Describe("merge", func() {
		It("merge cherry-picks setup..head onto the current branch and deletes the shell branch", func(ctx SpecContext) {
			f.runWorktree(false)
			runID := f.recordRun(ctx)

			landing, err := land.Land(ctx, f.provider, f.todo, land.Options{Via: native.LandingMerge})
			Expect(err).NotTo(HaveOccurred())

			head := git(f.repo, "rev-parse", "HEAD")
			Expect(git(f.repo, "log", "-1", "--format=%s")).To(Equal(agentMessage))
			Expect(filepath.Join(f.repo, "feature.txt")).To(BeARegularFile())
			Expect(branchExists(f.repo, shellBranch)).To(BeFalse())
			Expect(landing.PromptRunID).To(Equal(runID))
			Expect(landing.Via).To(Equal(native.LandingMerge))
			Expect(landing.TargetBranch).To(Equal("main"))
			Expect(landing.LandedSHA).To(Equal(head))
			Expect(landing.BranchDeletedAt).NotTo(BeNil())
			Expect(landing.CommitCount).To(Equal(1))
			Expect(landings(ctx)).To(Equal([]native.RunLanding{*landing}))
		})

		It("records every commit of setup..head it landed, not only the ones captain recorded", func(ctx SpecContext) {
			f.runWorktree(true)
			// The agent committed a second time itself; captain's workspace still
			// records only the hook's commit.
			f.state.Head = commitFile(f.worktree, "extra.txt", "extra\n", "feat: agent's own commit")
			f.recordRun(ctx)

			landing, err := land.Land(ctx, f.provider, f.todo, land.Options{Via: native.LandingMerge})
			Expect(err).NotTo(HaveOccurred())

			Expect(landing.CommitCount).To(Equal(2))
			Expect(git(f.repo, "rev-parse", landing.LandedSHA+"~2")).To(Equal(f.state.Base))
		})

		It("lands Setup..Head, excluding the setup snapshot", func(ctx SpecContext) {
			f.runWorktree(false)
			f.recordRun(ctx)

			_, err := land.Land(ctx, f.provider, f.todo, land.Options{Via: native.LandingMerge})
			Expect(err).NotTo(HaveOccurred())

			Expect(git(f.repo, "log", "--format=%s", f.state.Base+"..HEAD")).To(Equal(agentMessage))
			Expect(filepath.Join(f.repo, "wip.txt")).NotTo(BeAnExistingFile())
		})

		It("aborts and reports conflicts without touching the checkout", func(ctx SpecContext) {
			f.runWorktree(false)
			f.recordRun(ctx)
			before := commitFile(f.repo, "feature.txt", "main's own feature\n", "chore: conflicting feature")

			_, err := land.Land(ctx, f.provider, f.todo, land.Options{Via: native.LandingMerge})

			var conflict *branchmerge.ConflictError
			Expect(errors.As(err, &conflict)).To(BeTrue(), "expected *branchmerge.ConflictError, got %v", err)
			Expect(conflict.Paths).To(Equal([]string{"feature.txt"}))
			Expect(git(f.repo, "rev-parse", "HEAD")).To(Equal(before))
			Expect(git(f.repo, "status", "--porcelain")).To(BeEmpty())
			Expect(filepath.Join(f.repo, ".git", "CHERRY_PICK_HEAD")).NotTo(BeAnExistingFile())
			Expect(branchExists(f.repo, shellBranch)).To(BeTrue())
			Expect(landings(ctx)).To(BeEmpty())
		})

		It("refuses a staged index in the checkout", func(ctx SpecContext) {
			f.runWorktree(false)
			f.recordRun(ctx)
			Expect(os.WriteFile(filepath.Join(f.repo, "staged.txt"), []byte("x\n"), 0o644)).To(Succeed())
			git(f.repo, "add", "staged.txt")

			_, err := land.Land(ctx, f.provider, f.todo, land.Options{Via: native.LandingMerge})
			Expect(err).To(MatchError(branchmerge.ErrCheckoutNotReady))
			Expect(git(f.repo, "log", "-1", "--format=%s")).To(Equal("chore: initial"))
		})
	})

	Describe("refusals", func() {
		It("refuses a kept dirty worktree", func(ctx SpecContext) {
			f.runWorktree(true)
			Expect(os.WriteFile(filepath.Join(f.worktree, "notes.txt"), []byte("unsaved\n"), 0o644)).To(Succeed())
			f.state.Kept, f.state.KeptReason, f.state.Dirty = true, "uncommitted changes", []string{"notes.txt"}
			f.recordRun(ctx)

			_, err := land.Land(ctx, f.provider, f.todo, land.Options{Via: native.LandingMerge})
			Expect(err).To(MatchError(land.ErrWorktreeDirty))
			Expect(err).To(MatchError(ContainSubstring("notes.txt")))
			Expect(git(f.repo, "log", "-1", "--format=%s")).To(Equal("chore: initial"))
			Expect(filepath.Join(f.worktree, "notes.txt")).To(BeARegularFile())
			Expect(branchExists(f.repo, shellBranch)).To(BeTrue())
		})

		It("refuses a run with no commits past its setup snapshot", func(ctx SpecContext) {
			f.runWorktree(false)
			f.state.Head = f.state.Setup
			f.recordRun(ctx)

			_, err := land.Land(ctx, f.provider, f.todo, land.Options{Via: native.LandingMerge})
			Expect(err).To(MatchError(land.ErrNoCommits))
		})

		It("refuses a run that already landed", func(ctx SpecContext) {
			f.runWorktree(false)
			f.recordRun(ctx)
			_, err := land.Land(ctx, f.provider, f.todo, land.Options{Via: native.LandingMerge})
			Expect(err).NotTo(HaveOccurred())

			_, err = land.Land(ctx, f.provider, f.todo, land.Options{Via: native.LandingMerge})
			Expect(err).To(MatchError(native.ErrAlreadyLanded))
			Expect(landings(ctx)).To(HaveLen(1))
		})

		It("refuses a todo with no run step workspace", func(ctx SpecContext) {
			_, err := land.Land(ctx, f.provider, f.todo, land.Options{Via: native.LandingMerge})
			Expect(err).To(MatchError(land.ErrNoRunWorkspace))
		})
	})

	Describe("pr", func() {
		It("pr uses prcreate and deletes the branch only after the PR is created", func(ctx SpecContext) {
			f.runWorktree(false)
			runID := f.recordRun(ctx)
			var created []model.CreatePRInput
			branchAtCreate := false
			deps := prcreate.Deps{Preflight: func(context.Context, provider.Options) error { return nil },
				GenerateContent: func(context.Context, commitpkg.PRContentInput) (commitpkg.PRContent, error) {
					return commitpkg.PRContent{Title: "feat: add feature", Body: "Adds feature.txt", Branch: "feat/feature"}, nil
				},
				CreatePR: func(_ provider.Options, in model.CreatePRInput) (*model.CreatePRResult, error) {
					created = append(created, in)
					branchAtCreate = branchExists(f.repo, shellBranch)
					return &model.CreatePRResult{Number: 7, URL: "https://github.com/acme/land/pull/7", Base: in.Base}, nil
				},
			}

			landing, err := land.Land(ctx, f.provider, f.todo, land.Options{Via: native.LandingPR, Base: "origin/main", Deps: deps})
			Expect(err).NotTo(HaveOccurred())

			Expect(created).To(HaveLen(1))
			Expect(created[0].Base).To(Equal("main"))
			Expect(branchAtCreate).To(BeTrue(), "the shell branch must survive until the PR exists")
			Expect(branchExists(f.repo, shellBranch)).To(BeFalse())
			topicHead := git(f.bare, "rev-parse", "refs/heads/"+created[0].Head)
			Expect(git(f.bare, "log", "--format=%s", "main.."+topicHead)).To(Equal(agentMessage))
			Expect(git(f.repo, "rev-parse", "HEAD")).To(Equal(f.state.Base), "a PR landing leaves the checkout alone")
			prNumber := 7
			Expect(*landing).To(MatchFields(IgnoreExtras, Fields{
				"PromptRunID": Equal(runID), "Via": Equal(native.LandingPR), "TargetBranch": Equal("main"),
				"LandedSHA": Equal(topicHead), "PRNumber": Equal(&prNumber), "CommitCount": Equal(1),
				"PRURL": Equal("https://github.com/acme/land/pull/7"), "BranchDeletedAt": Not(BeNil()),
			}))
		})
	})
})
