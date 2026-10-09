package prcreate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/pr/model"
	"github.com/flanksource/gavel/pr/provider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const testBase = "origin/main"

// recordingDeps returns Deps whose PR call records the request and whose
// content generator returns fixed content, so no LLM or GitHub call is made.
func recordingDeps(content commitpkg.PRContent, prs *[]model.CreatePRInput, contentInputs *[]commitpkg.PRContentInput) Deps {
	return Deps{Preflight: func(context.Context, provider.Options) error { return nil },
		CreatePR: func(_ provider.Options, in model.CreatePRInput) (*model.CreatePRResult, error) {
			*prs = append(*prs, in)
			return &model.CreatePRResult{Number: len(*prs), URL: "https://example/pr", Title: in.Title, Base: in.Base}, nil
		},
		GenerateContent: func(_ context.Context, in commitpkg.PRContentInput) (commitpkg.PRContent, error) {
			if contentInputs != nil {
				*contentInputs = append(*contentInputs, in)
			}
			return content, nil
		},
	}
}

func failingDeps() Deps {
	return Deps{Preflight: func(context.Context, provider.Options) error { return nil },
		CreatePR: func(provider.Options, model.CreatePRInput) (*model.CreatePRResult, error) {
			Fail("CreatePR must not be called")
			return nil, nil
		},
		GenerateContent: func(context.Context, commitpkg.PRContentInput) (commitpkg.PRContent, error) {
			Fail("GenerateContent must not be called")
			return commitpkg.PRContent{}, nil
		},
	}
}

var _ = Describe("splitBaseRef", func() {
	DescribeTable("strips only the origin/ remote and refs/heads/ prefixes for GitHub",
		func(in, branch, ref string) {
			gotBranch, gotRef := splitBaseRef(in)
			Expect([]string{gotBranch, gotRef}).To(Equal([]string{branch, ref}))
		},
		Entry("remote-tracking", "origin/main", "main", "origin/main"),
		Entry("local branch", "main", "main", "main"),
		Entry("full ref", "refs/heads/foo", "foo", "refs/heads/foo"),
	)
})

var _ = Describe("preflight", func() {
	var f *repoFixture
	BeforeEach(func() { f = newRepoFixture() })

	It("resolves an existing SHA to its full form and rejects an unknown one", func() {
		base, got, err := preflight(f.repo, Input{SHAs: []string{f.topicSHA[:10]}, Base: testBase})
		Expect(err).NotTo(HaveOccurred())
		Expect(base).To(Equal(testBase))
		Expect(got).To(Equal([]string{f.topicSHA}))

		_, _, err = preflight(f.repo, Input{SHAs: []string{"deadbeef"}, Base: testBase})
		Expect(err).To(MatchError(ContainSubstring(`commit "deadbeef" not found`)))
	})

	It("rejects an empty SHA list", func() {
		_, _, err := preflight(f.repo, Input{Base: testBase})
		Expect(err).To(MatchError(ContainSubstring("at least one SHA")))
	})

	It("rejects a merge commit unless a mainline parent is chosen", func() {
		git(f.repo, "checkout", "-b", "other")
		commitFile(f.repo, "other.txt", "o\n", "chore: other branch")
		git(f.repo, "checkout", "main")
		git(f.repo, "merge", "--no-ff", "-m", "merge: other", "other")
		mergeSHA := git(f.repo, "rev-parse", "HEAD")

		_, _, err := preflight(f.repo, Input{SHAs: []string{mergeSHA}, Base: testBase})
		Expect(err).To(MatchError(ContainSubstring("--mainline")))

		_, _, err = preflight(f.repo, Input{SHAs: []string{mergeSHA}, Base: testBase, Mainline: 1})
		Expect(err).NotTo(HaveOccurred())
	})

	It("refuses while the source repo is mid-rebase", func() {
		gitDir := git(f.repo, "rev-parse", "--absolute-git-dir")
		Expect(os.WriteFile(filepath.Join(gitDir, "REBASE_HEAD"), []byte(f.topicSHA), 0o644)).To(Succeed())
		_, _, err := preflight(f.repo, Input{SHAs: []string{f.topicSHA}, Base: testBase})
		Expect(err).To(MatchError(ContainSubstring("REBASE_HEAD")))
	})
})

var _ = Describe("assertSafeWorktreePath", func() {
	var f *repoFixture
	BeforeEach(func() { f = newRepoFixture() })

	It("refuses a path outside the scratch dir", func() {
		Expect(assertSafeWorktreePath(f.repo, "/etc")).To(MatchError(ContainSubstring("outside")))
	})

	It("refuses a scratch path without a marker", func() {
		wt := filepath.Join(f.repo, scratchSub, "abc12345-x")
		Expect(os.MkdirAll(wt, 0o755)).To(Succeed())
		Expect(assertSafeWorktreePath(f.repo, wt)).To(MatchError(ContainSubstring("marker missing")))
	})

	It("refuses a marked path outside the scratch dir", func() {
		bad := filepath.Join(f.repo, "not-tmp", "abc")
		Expect(os.MkdirAll(bad, 0o755)).To(Succeed())
		Expect(writeMarker(bad, strings.Repeat("a", 40), f.repo)).To(Succeed())
		Expect(assertSafeWorktreePath(f.repo, bad)).To(MatchError(ContainSubstring("outside")))
	})
})

var _ = Describe("Create", func() {
	var f *repoFixture
	BeforeEach(func() { f = newRepoFixture() })

	It("opens a PR from one SHA, pushes the topic branch and cleans up the worktree", func() {
		var prs []model.CreatePRInput
		content := commitpkg.PRContent{Title: "feat: add feature.txt", Body: "body", Branch: "feat/add-feature"}
		res, err := Create(context.Background(), f.repo, Input{
			SHAs: []string{f.topicSHA}, Base: testBase, Draft: true,
			Deps: recordingDeps(content, &prs, nil),
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(prs).To(HaveLen(1))
		Expect(prs[0]).To(Equal(model.CreatePRInput{
			Title: content.Title, Body: content.Body, Head: res.TopicBranch, Base: "main", Draft: true,
		}))
		Expect(res.TopicBranch).To(HavePrefix("feat/add-feature-"))
		Expect(res.PR.Number).To(Equal(1))
		Expect(res.Content).To(Equal(content))
		Expect(res.TopicHead).To(Equal(git(f.bare, "rev-parse", "refs/heads/"+res.TopicBranch)))

		Expect(f.scratchEntries()).To(BeEmpty())
		Expect(git(f.repo, "worktree", "list", "--porcelain")).NotTo(ContainSubstring("/.tmp/pr-create/"))
	})

	It("cherry-picks multiple SHAs in order and reports TopicHead", func() {
		git(f.repo, "checkout", "feature")
		secondSHA := commitFile(f.repo, "second.txt", "two\n", "feat: add second.txt")
		git(f.repo, "checkout", "main")

		var prs []model.CreatePRInput
		var contentInputs []commitpkg.PRContentInput
		res, err := Create(context.Background(), f.repo, Input{
			SHAs: []string{secondSHA, f.topicSHA}, Base: testBase,
			Deps: recordingDeps(commitpkg.PRContent{Title: "feat: two", Branch: "feat/two"}, &prs, &contentInputs),
		})
		Expect(err).NotTo(HaveOccurred())

		topicRef := "refs/heads/" + res.TopicBranch
		Expect(res.TopicHead).To(Equal(git(f.bare, "rev-parse", topicRef)))
		Expect(git(f.bare, "log", "--reverse", "--format=%s", "main.."+topicRef)).
			To(Equal("feat: add second.txt\nfeat: add feature.txt"))

		Expect(contentInputs).To(HaveLen(1))
		Expect(contentInputs[0].Commits).To(Equal([]commitpkg.PRCommitInput{
			{Message: "feat: add second.txt", Files: []string{"second.txt"}},
			{Message: "feat: add feature.txt", Files: []string{"feature.txt"}},
		}))
		Expect(f.scratchEntries()).To(BeEmpty())
	})

	It("keeps the worktree and reports a ConflictError on a cherry-pick conflict", func() {
		commitFile(f.repo, "feature.txt", "conflict\n", "chore: conflicting")
		git(f.repo, "push", "origin", "main")

		_, err := Create(context.Background(), f.repo, Input{SHAs: []string{f.topicSHA}, Base: testBase, Deps: failingDeps()})
		var conflict *ConflictError
		Expect(errors.As(err, &conflict)).To(BeTrue(), "expected *ConflictError, got %v", err)
		Expect(err).To(MatchError(ContainSubstring("git cherry-pick " + f.topicSHA[:8])))

		entries := f.scratchEntries()
		Expect(entries).To(HaveLen(1))
		wt := filepath.Join(f.repo, scratchSub, entries[0].Name())
		Expect(conflict.Worktree).To(Equal(wt))
		Expect(conflict.Branch).To(HavePrefix(tmpBranchNS))
		Expect(filepath.Join(wt, markerFile)).To(BeAnExistingFile())
		DeferCleanup(func() { Expect(removeWorktree(f.repo, wt)).To(Succeed()) })
	})

	It("reports an already-landed SHA without keeping a worktree", func() {
		_, err := Create(context.Background(), f.repo, Input{SHAs: []string{f.baseSHA}, Base: testBase, Deps: failingDeps()})
		Expect(err).To(MatchError(ContainSubstring("is already in main; nothing to PR")))
		Expect(f.scratchEntries()).To(BeEmpty())
	})

	It("gives each rerun a fresh topic branch", func() {
		var prs []model.CreatePRInput
		deps := recordingDeps(commitpkg.PRContent{Title: "feat: x", Branch: "feat/x"}, &prs, nil)
		for range 2 {
			_, err := Create(context.Background(), f.repo, Input{SHAs: []string{f.topicSHA}, Base: testBase, Deps: deps})
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(prs).To(HaveLen(2))
		Expect(prs[0].Head).NotTo(Equal(prs[1].Head))
	})

	It("pushes nothing when PR content generation fails", func() {
		deps := failingDeps()
		deps.GenerateContent = func(context.Context, commitpkg.PRContentInput) (commitpkg.PRContent, error) {
			return commitpkg.PRContent{}, commitpkg.ErrLLMUnavailable
		}
		_, err := Create(context.Background(), f.repo, Input{SHAs: []string{f.topicSHA}, Base: testBase, Deps: deps})
		Expect(errors.Is(err, commitpkg.ErrLLMUnavailable)).To(BeTrue(), "got %v", err)
		Expect(git(f.bare, "branch", "--list")).To(Equal("* main"))
	})

	It("rejects missing dependencies before touching git", func() {
		_, err := Create(context.Background(), f.repo, Input{SHAs: []string{f.topicSHA}, Base: testBase})
		Expect(err).To(MatchError(ContainSubstring("Deps.CreatePR is nil")))
		Expect(f.scratchEntries()).To(BeEmpty())
	})
})
