package commit

import (
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clickyai "github.com/flanksource/clicky/ai"
	"github.com/flanksource/gavel/github"
)

var _ = Describe("isProtectedBranch", func() {
	DescribeTable("recognises common protected branch names",
		func(name string, expected bool) {
			Expect(isProtectedBranch(name)).To(Equal(expected))
		},
		Entry("main", "main", true),
		Entry("master", "master", true),
		Entry("develop", "develop", true),
		Entry("trunk", "trunk", true),
		Entry("MAIN (case-insensitive)", "MAIN", true),
		Entry("trimmed whitespace", "  main  ", true),
		Entry("topic branch", "feat/something", false),
		Entry("empty", "", false),
	)
})

var _ = Describe("sanitizeBranchName", func() {
	DescribeTable("normalises AI-suggested branch names",
		func(in, expected string) {
			Expect(sanitizeBranchName(in)).To(Equal(expected))
		},
		Entry("already clean", "feat/foo-bar", "feat/foo-bar"),
		Entry("uppercase to lower", "Feat/Foo-Bar", "feat/foo-bar"),
		Entry("spaces to dashes", "feat foo bar", "feat-foo-bar"),
		Entry("underscore to dash", "feat_foo", "feat-foo"),
		Entry("collapse double slash", "feat//foo", "feat/foo"),
		Entry("collapse double dash", "feat--foo", "feat-foo"),
		Entry("trim leading slash", "/feat/foo", "feat/foo"),
		Entry("strip illegal chars", "feat/foo!@#bar", "feat/foobar"),
		Entry("empty input", "", ""),
		Entry("only illegal chars", "!!!", ""),
	)
})

var _ = Describe("executeNewPRPush on a protected current branch", func() {
	var (
		workDir string
		opts    Options
	)

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		runGit(workDir, "init", "-b", "main")
		runGit(workDir, "config", "user.email", "test@example.com")
		runGit(workDir, "config", "user.name", "Test User")
		runGit(workDir, "config", "commit.gpgsign", "false")
		runGit(workDir, "commit", "--allow-empty", "-m", "root")
		runGit(workDir, "commit", "--allow-empty", "-m", "feat: on main")
		// HEAD is on main, with no origin/main → "ahead" set is HEAD~ + HEAD.
		opts = Options{WorkDir: workDir, Push: true}

		// Stub the LLM constructor so generatePRContent isn't actually called.
		previousAgent := newAgentFunc
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) {
			return nil, nil
		}
		DeferCleanup(func() { newAgentFunc = previousAgent })
	})

	It("uses the AI-suggested branch as the head ref and never pushes HEAD:main", func() {
		var (
			pushedRef    string
			createdHead  string
			confirmCalls int
		)
		err := pushAfterCommitForTest(context.Background(), opts, &Result{
			Commits: []CommitResult{{Message: "feat: on main", Hash: "abc123"}},
		}, pushDeps{
			searchPRs: func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
				return github.PRSearchResults{}, nil, nil
			},
			defaultBranch: func(github.Options) (string, error) { return "main", nil },
			isAncestor:    func(_, _, _ string) bool { return false },
			generatePRPrompt: func(context.Context, clickyai.Agent, prContentInput) (prContent, error) {
				return prContent{
					Title:  "feat: on main",
					Body:   "body",
					Branch: "feat/topic-from-main",
				}, nil
			},
			rebaseOnto: func(_, _ string) error { return nil },
			gitPush: func(_, refspec string) error {
				pushedRef = refspec
				return nil
			},
			createPR: func(_ github.Options, in github.CreatePRInput) (*github.CreatePRResult, error) {
				createdHead = in.Head
				return &github.CreatePRResult{Number: 7, URL: "http://x/7", Title: in.Title, Base: "main"}, nil
			},
			confirmProtectedRef: func(string) bool {
				confirmCalls++
				return true // not exercised on this path
			},
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(pushedRef).To(Equal("HEAD:feat/topic-from-main"))
		Expect(createdHead).To(Equal("feat/topic-from-main"))
		Expect(confirmCalls).To(Equal(0), "confirmation only fires when target is itself protected")
	})

	It("fails when AI returns no branch suggestion", func() {
		err := pushAfterCommitForTest(context.Background(), opts, &Result{
			Commits: []CommitResult{{Message: "feat: x"}},
		}, pushDeps{
			searchPRs: func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
				return github.PRSearchResults{}, nil, nil
			},
			defaultBranch: func(github.Options) (string, error) { return "main", nil },
			isAncestor:    func(_, _, _ string) bool { return false },
			generatePRPrompt: func(context.Context, clickyai.Agent, prContentInput) (prContent, error) {
				return prContent{Title: "feat: x", Branch: ""}, nil
			},
		})
		Expect(err).To(MatchError(ContainSubstring(`AI did not suggest a branch name`)))
	})
})

var _ = Describe("executeExistingPRPush guards protected source branches", func() {
	var (
		workDir string
		opts    Options
	)

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		runGit(workDir, "init", "-b", "main")
		runGit(workDir, "config", "user.email", "test@example.com")
		runGit(workDir, "config", "user.name", "Test User")
		runGit(workDir, "config", "commit.gpgsign", "false")
		runGit(workDir, "commit", "--allow-empty", "-m", "root")
		runGit(workDir, "checkout", "-b", "feat/topic")
		runGit(workDir, "commit", "--allow-empty", "-m", "feat: x")
		runGit(workDir, "update-ref", filepath.Join("refs/remotes/origin", "main"), "main")
		opts = Options{WorkDir: workDir, Push: true}
	})

	It("cancels when the user declines the protected-branch confirmation", func() {
		var pushCalls int
		err := pushAfterCommitForTest(context.Background(), opts, &Result{
			Commits: []CommitResult{{Message: "feat: x"}},
		}, pushDeps{
			searchPRs: func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
				// PR head is "main" — would push HEAD:main without a guard.
				return github.PRSearchResults{
					{Number: 1, Source: "main", Target: "release", URL: "http://x/1"},
				}, nil, nil
			},
			isAncestor: func(_, ref, _ string) bool { return ref == "origin/main" },
			gitPush: func(_, _ string) error {
				pushCalls++
				return nil
			},
			rebaseOnto: func(_, _ string) error { return nil },
			confirmProtectedRef: func(name string) bool {
				Expect(name).To(Equal("main"))
				return false
			},
		})
		Expect(err).To(MatchError(ContainSubstring(`push to protected branch "main" cancelled`)))
		Expect(pushCalls).To(Equal(0))
	})

	It("proceeds when the user confirms the protected-branch push", func() {
		var pushedRef string
		err := pushAfterCommitForTest(context.Background(), opts, &Result{
			Commits: []CommitResult{{Message: "feat: x"}},
		}, pushDeps{
			searchPRs: func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
				return github.PRSearchResults{
					{Number: 1, Source: "main", Target: "release", URL: "http://x/1"},
				}, nil, nil
			},
			isAncestor: func(_, ref, _ string) bool { return ref == "origin/main" },
			gitPush: func(_, refspec string) error {
				pushedRef = refspec
				return nil
			},
			rebaseOnto:          func(_, _ string) error { return nil },
			confirmProtectedRef: func(string) bool { return true },
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(pushedRef).To(Equal("HEAD:main"))
	})
})
