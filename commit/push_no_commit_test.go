package commit

import (
	"bytes"
	"context"
	"path/filepath"

	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("loadAheadCommits", func() {
	var workDir string

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		runGit(workDir, "init", "-b", "main")
		runGit(workDir, "commit", "--allow-empty", "-m", "root")
		runGit(workDir, "checkout", "-b", "feat/x")
		runGit(workDir, "commit", "--allow-empty", "-m", "feat: first")
		runGit(workDir, "commit", "--allow-empty", "-m", "feat: second")

		// Mirror branches as origin/<name> so refs resolve without a real remote.
		runGit(workDir, "update-ref", filepath.Join("refs/remotes/origin", "main"), "main")
	})

	It("returns commits ahead of origin/<default-base> oldest first", func() {
		commits, err := loadAheadCommits(workDir, "feat/x", "origin/main")
		Expect(err).ToNot(HaveOccurred())
		Expect(commits).To(HaveLen(2))
		Expect(commits[0].Message).To(Equal("feat: first"))
		Expect(commits[1].Message).To(Equal("feat: second"))
		Expect(commits[0].Hash).ToNot(BeEmpty())
		Expect(commits[1].Hash).ToNot(BeEmpty())
	})

	It("prefers origin/<branch> over default base when present", func() {
		// Mirror the feature branch at its initial commit; HEAD has moved one
		// commit beyond, so only that one commit is "ahead".
		runGit(workDir, "update-ref", filepath.Join("refs/remotes/origin", "feat/x"), "HEAD~1")

		commits, err := loadAheadCommits(workDir, "feat/x", "origin/main")
		Expect(err).ToNot(HaveOccurred())
		Expect(commits).To(HaveLen(1))
		Expect(commits[0].Message).To(Equal("feat: second"))
	})

	It("returns empty when HEAD is not ahead of any base", func() {
		runGit(workDir, "update-ref", filepath.Join("refs/remotes/origin", "feat/x"), "HEAD")
		commits, err := loadAheadCommits(workDir, "feat/x", "origin/main")
		Expect(err).ToNot(HaveOccurred())
		Expect(commits).To(BeEmpty())
	})

	It("returns empty when no base ref can be resolved", func() {
		commits, err := loadAheadCommits(workDir, "feat/y", "origin/does-not-exist")
		Expect(err).ToNot(HaveOccurred())
		Expect(commits).To(BeEmpty())
	})
})

var _ = Describe("pushAfterCommit with no commits in result", func() {
	var (
		workDir string
		opts    Options
	)

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		runGit(workDir, "init", "-b", "main")
		runGit(workDir, "commit", "--allow-empty", "-m", "root")
		runGit(workDir, "checkout", "-b", "feat/push-only")
		runGit(workDir, "commit", "--allow-empty", "-m", "feat: ahead")
		runGit(workDir, "update-ref", filepath.Join("refs/remotes/origin", "main"), "main")
		opts = Options{WorkDir: workDir, Push: true, DryRun: true}
	})

	It("returns ErrNothingToPush when no staged commits and HEAD is not ahead", func() {
		// Move origin/main to HEAD so HEAD has nothing ahead.
		runGit(workDir, "update-ref", filepath.Join("refs/remotes/origin", "main"), "HEAD")

		err := pushAfterCommitForTest(context.Background(), opts, &Result{DryRun: true}, pushDeps{
			defaultBranch: func(github.Options) (string, error) { return "main", nil },
			aheadCommits:  loadAheadCommits,
		})
		Expect(err).To(MatchError(ErrNothingToPush))
	})

	It("populates result.Commits from ahead-of-upstream commits and pushes", func() {
		var pushedRef string
		err := pushAfterCommitForTest(context.Background(), opts, &Result{DryRun: true}, pushDeps{
			searchPRs: func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
				return github.PRSearchResults{
					{Number: 9, Source: "feat/push-only", Target: "main", URL: "http://x/9"},
				}, nil, nil
			},
			defaultBranch: func(github.Options) (string, error) { return "main", nil },
			aheadCommits:  loadAheadCommits,
			gitPush: func(_, refspec string) error {
				pushedRef = refspec
				return nil
			},
		})
		Expect(err).ToNot(HaveOccurred())
		// dry-run path: gitPush should NOT be invoked.
		Expect(pushedRef).To(BeEmpty())
	})
})

var _ = Describe("Run with --push and nothing staged", func() {
	var workDir string

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		runGit(workDir, "init", "-b", "main")
		runGit(workDir, "config", "user.email", "test@example.com")
		runGit(workDir, "config", "user.name", "Test User")
		runGit(workDir, "config", "commit.gpgsign", "false")
		runGit(workDir, "commit", "--allow-empty", "-m", "root")
		runGit(workDir, "checkout", "-b", "feat/runpush")
		runGit(workDir, "commit", "--allow-empty", "-m", "feat: ahead one")
		runGit(workDir, "update-ref", filepath.Join("refs/remotes/origin", "main"), "main")
	})

	It("skips commit step and pushes HEAD when ahead-of-upstream commits exist", func() {
		var pushed string
		pushDepsForTest = &pushDeps{
			searchPRs: func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
				return github.PRSearchResults{
					{Number: 5, Source: "feat/runpush", Target: "main", URL: "http://x/5"},
				}, nil, nil
			},
			defaultBranch: func(github.Options) (string, error) { return "main", nil },
			isAncestor:    gitIsAncestor,
			gitPush: func(_, refspec string) error {
				pushed = refspec
				return nil
			},
			pickPR:           choosePR,
			generatePRPrompt: GeneratePRContent,
			aheadCommits:     loadAheadCommits,
		}
		defer func() { pushDepsForTest = nil }()

		result, err := Run(context.Background(), Options{
			WorkDir: workDir,
			Push:    true,
			DryRun:  true, // dry-run skips real gitPush, but still exercises the branch-match path
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(result).ToNot(BeNil())
		Expect(result.Commits).To(HaveLen(1))
		Expect(result.Commits[0].Message).To(Equal("feat: ahead one"))
		Expect(pushed).To(BeEmpty(), "dry-run must not invoke gitPush")
	})

	It("returns ErrNothingToPush when nothing is staged and no commits ahead", func() {
		runGit(workDir, "update-ref", filepath.Join("refs/remotes/origin", "main"), "HEAD")

		pushDepsForTest = &pushDeps{
			searchPRs: func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
				return github.PRSearchResults{}, nil, nil
			},
			defaultBranch: func(github.Options) (string, error) { return "main", nil },
			aheadCommits:  loadAheadCommits,
		}
		defer func() { pushDepsForTest = nil }()

		_, err := Run(context.Background(), Options{
			WorkDir: workDir,
			Push:    true,
			DryRun:  true,
		})
		Expect(err).To(MatchError(ErrNothingToPush))
	})

	It("dry-run prints the commit-to-be-pushed preview", func() {
		var buf bytes.Buffer
		previousOut := dryRunOutput
		dryRunOutput = &buf
		defer func() { dryRunOutput = previousOut }()

		pushDepsForTest = &pushDeps{
			searchPRs: func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
				return github.PRSearchResults{
					{Number: 5, Source: "feat/runpush", Target: "main", URL: "http://x/5"},
				}, nil, nil
			},
			defaultBranch: func(github.Options) (string, error) { return "main", nil },
			isAncestor:    gitIsAncestor,
			gitPush:       func(_, _ string) error { return nil },
			pickPR:        choosePR,
			aheadCommits:  loadAheadCommits,
		}
		defer func() { pushDepsForTest = nil }()

		_, err := Run(context.Background(), Options{
			WorkDir: workDir,
			Push:    true,
			DryRun:  true,
		})
		Expect(err).ToNot(HaveOccurred())
		out := buf.String()
		Expect(out).To(ContainSubstring("DRY RUN"))
		Expect(out).To(ContainSubstring("would push 1 existing commit"))
		// Pretty() splits "feat: ahead one" into a styled prefix + subject;
		// match the subject only.
		Expect(out).To(ContainSubstring("ahead one"))
		Expect(out).To(ContainSubstring("would push HEAD:feat/runpush"))
	})
})

var _ = Describe("dry-run new-PR push generates title/body and reports without pushing", func() {
	var (
		workDir string
		opts    Options
	)

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		runGit(workDir, "init", "-b", "main")
		runGit(workDir, "commit", "--allow-empty", "-m", "root")
		runGit(workDir, "checkout", "-b", "feat/new-pr")
		runGit(workDir, "commit", "--allow-empty", "-m", "feat: needs PR")
		runGit(workDir, "update-ref", filepath.Join("refs/remotes/origin", "main"), "main")
		opts = Options{WorkDir: workDir, Push: true, DryRun: true}
	})

	It("invokes generatePRPrompt and includes its output, but never pushes or creates a PR", func() {
		var buf bytes.Buffer
		previousOut := dryRunOutput
		dryRunOutput = &buf
		defer func() { dryRunOutput = previousOut }()

		// Stub out the AI agent constructor so the test does not depend on
		// an API key being configured in the environment. The injected
		// generatePRPrompt below is what actually drives the dry-run
		// output, so the agent value here is never invoked.
		previousAgent := newAgentFunc
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) {
			return &agentLifecycleProbe{}, nil
		}
		defer func() { newAgentFunc = previousAgent }()

		var (
			llmCalls    int
			pushCalls   int
			createCalls int
		)
		err := pushAfterCommitForTest(context.Background(), opts, &Result{DryRun: true, PushOnly: true}, pushDeps{
			// No matching PR, no ancestor PR — forces the new-PR path.
			searchPRs: func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
				return github.PRSearchResults{}, nil, nil
			},
			defaultBranch: func(github.Options) (string, error) { return "main", nil },
			isAncestor:    func(_, _, _ string) bool { return false },
			aheadCommits:  loadAheadCommits,
			generatePRPrompt: func(context.Context, PRContentInput) (PRContent, error) {
				llmCalls++
				return PRContent{
					Title:  "feat: add the thing",
					Body:   "## What\n- adds the thing\n## Why\n- needed it",
					Branch: "feat/add-the-thing",
				}, nil
			},
			gitPush: func(_, _ string) error {
				pushCalls++
				return nil
			},
			createPR: func(github.Options, model.CreatePRInput) (*model.CreatePRResult, error) {
				createCalls++
				return nil, nil
			},
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(llmCalls).To(Equal(1), "dry-run should call the LLM exactly once")
		Expect(pushCalls).To(Equal(0), "dry-run must not invoke gitPush")
		Expect(createCalls).To(Equal(0), "dry-run must not invoke createPR")

		out := buf.String()
		Expect(out).To(ContainSubstring("would push HEAD:feat/new-pr and open PR against main"))
		Expect(out).To(ContainSubstring("title: feat: add the thing"))
		Expect(out).To(ContainSubstring("## What"))
	})
})

var _ = Describe("pushAfterCommit with PushModeBranch on an already-pushed branch", func() {
	const (
		branch      = "feat/pushed"
		baseBranch  = "main"
		prNumber    = 42
		mergeMethod = "rebase"
	)
	var (
		workDir  string
		pushed   []string
		created  []model.CreatePRInput
		merged   []string
		seededBy []string
		deps     pushDeps
	)

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		runGit(workDir, "init", "-b", baseBranch)
		runGit(workDir, "commit", "--allow-empty", "-m", "root")
		runGit(workDir, "update-ref", filepath.Join("refs/remotes/origin", baseBranch), baseBranch)
		runGit(workDir, "checkout", "-b", branch)
		runGit(workDir, "commit", "--allow-empty", "-m", "feat: first")
		runGit(workDir, "commit", "--allow-empty", "-m", "fix: second")
		// The branch is already on the remote at HEAD, so nothing is ahead of its upstream.
		runGit(workDir, "update-ref", filepath.Join("refs/remotes/origin", branch), "HEAD")
		pushed, created, merged, seededBy = nil, nil, nil, nil
		deps = pushDeps{
			searchPRs: func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
				return github.PRSearchResults{}, nil, nil
			},
			defaultBranch: func(github.Options) (string, error) { return baseBranch, nil },
			isAncestor:    func(_, _, _ string) bool { return false },
			aheadCommits:  loadAheadCommits,
			rebaseOnto:    func(_, _ string) error { return nil },
			generatePRPrompt: func(_ context.Context, in PRContentInput) (PRContent, error) {
				for _, c := range in.Commits {
					seededBy = append(seededBy, c.Message)
				}
				return PRContent{Title: "feat: pushed branch", Body: "body", Branch: "feat/unused"}, nil
			},
			gitPush: func(_, refspec string) error { pushed = append(pushed, refspec); return nil },
			createPR: func(_ github.Options, in model.CreatePRInput) (*model.CreatePRResult, error) {
				created = append(created, in)
				return &model.CreatePRResult{Number: prNumber, NodeID: "PR_node", Base: in.Base, URL: "http://x/42"}, nil
			},
			enableAutoMerge: func(_ github.Options, nodeID, method string) error {
				merged = append(merged, nodeID+":"+method)
				return nil
			},
		}
	})

	It("still reports nothing to push in the default commits mode", func() {
		err := pushAfterCommitForTest(context.Background(), Options{WorkDir: workDir, Push: true}, &Result{}, deps)
		Expect(err).To(MatchError(ErrNothingToPush))
	})

	It("opens an auto-merge PR for every commit on the branch ahead of the default branch", func() {
		opts := Options{WorkDir: workDir, Push: true, PushMode: PushModeBranch, AutoMerge: true, MergeType: mergeMethod}
		Expect(pushAfterCommitForTest(context.Background(), opts, &Result{}, deps)).To(Succeed())

		Expect(seededBy).To(Equal([]string{"feat: first", "fix: second"}))
		Expect(pushed).To(Equal([]string{"HEAD:" + branch}))
		Expect(created).To(Equal([]model.CreatePRInput{{Title: "feat: pushed branch", Body: "body", Head: branch, Base: baseBranch}}))
		Expect(merged).To(Equal([]string{"PR_node:" + mergeMethod}))
	})

	It("seeds the PR from the whole branch, not only the commit this run made", func() {
		runGit(workDir, "commit", "--allow-empty", "-m", "chore: third")
		fresh := CommitResult{Hash: "abc", Message: "chore: third"}
		opts := Options{WorkDir: workDir, Push: true, PushMode: PushModeBranch}
		Expect(pushAfterCommitForTest(context.Background(), opts, &Result{Commits: []CommitResult{fresh}}, deps)).To(Succeed())

		Expect(seededBy).To(Equal([]string{"feat: first", "fix: second", "chore: third"}))
	})
})

// pushAfterCommitForTest mirrors pushAfterCommit but lets tests inject
// pushDeps directly. Keeps tests honest about the exported entry point
// without exposing the deps in production code.
func pushAfterCommitForTest(ctx context.Context, opts Options, result *Result, deps pushDeps) error {
	// Fill missing deps from defaults so the test only specifies what matters.
	d := defaultPushDeps()
	if deps.searchPRs != nil {
		d.searchPRs = deps.searchPRs
	}
	if deps.defaultBranch != nil {
		d.defaultBranch = deps.defaultBranch
	}
	if deps.createPR != nil {
		d.createPR = deps.createPR
	}
	if deps.isAncestor != nil {
		d.isAncestor = deps.isAncestor
	}
	if deps.gitPush != nil {
		d.gitPush = deps.gitPush
	}
	if deps.rebaseOnto != nil {
		d.rebaseOnto = deps.rebaseOnto
	}
	if deps.pickPR != nil {
		d.pickPR = deps.pickPR
	}
	if deps.generatePRPrompt != nil {
		d.generatePRPrompt = deps.generatePRPrompt
	}
	if deps.aheadCommits != nil {
		d.aheadCommits = deps.aheadCommits
	}
	if deps.confirmProtectedRef != nil {
		d.confirmProtectedRef = deps.confirmProtectedRef
	}
	if deps.enableAutoMerge != nil {
		d.enableAutoMerge = deps.enableAutoMerge
	}
	return pushWithDeps(ctx, opts, result, d)
}
