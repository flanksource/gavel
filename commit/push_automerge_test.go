package commit

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/github"
)

var _ = Describe("commit -p --auto-merge", func() {
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
		runGit(workDir, "commit", "--allow-empty", "-m", "feat: work")
		opts = Options{WorkDir: workDir, Push: true, AutoMerge: true, MergeType: "rebase"}

		previousAgent := newAgentFunc
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) { return nil, nil }
		DeferCleanup(func() { newAgentFunc = previousAgent })
	})

	newPRDeps := func(autoMerge func(github.Options, string, string) error) pushDeps {
		return pushDeps{
			searchPRs: func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
				return github.PRSearchResults{}, nil, nil
			},
			defaultBranch: func(github.Options) (string, error) { return "main", nil },
			isAncestor:    func(_, _, _ string) bool { return false },
			generatePRPrompt: func(context.Context, clickyai.Agent, PRContentInput) (PRContent, error) {
				return PRContent{Title: "feat: work", Body: "body", Branch: "feat/topic"}, nil
			},
			rebaseOnto: func(_, _ string) error { return nil },
			gitPush:    func(_, _ string) error { return nil },
			createPR: func(github.Options, github.CreatePRInput) (*github.CreatePRResult, error) {
				return &github.CreatePRResult{Number: 7, URL: "http://x/7", Title: "feat: work", NodeID: "PR_node_7", Base: "main"}, nil
			},
			enableAutoMerge: autoMerge,
		}
	}

	It("enables auto-merge on the newly opened PR with the chosen node ID and method", func() {
		var (
			gotNodeID string
			gotMethod string
			calls     int
		)
		err := pushAfterCommitForTest(context.Background(), opts, &Result{
			Commits: []CommitResult{{Message: "feat: work", Hash: "abc123"}},
		}, newPRDeps(func(_ github.Options, nodeID, method string) error {
			calls++
			gotNodeID = nodeID
			gotMethod = method
			return nil
		}))
		Expect(err).ToNot(HaveOccurred())
		Expect(calls).To(Equal(1))
		Expect(gotNodeID).To(Equal("PR_node_7"))
		Expect(gotMethod).To(Equal("rebase"))
	})

	It("propagates an error when enabling auto-merge fails", func() {
		err := pushAfterCommitForTest(context.Background(), opts, &Result{
			Commits: []CommitResult{{Message: "feat: work", Hash: "abc123"}},
		}, newPRDeps(func(github.Options, string, string) error {
			return errors.New("auto merge is not allowed for this repository")
		}))
		Expect(err).To(MatchError(ContainSubstring("enable auto-merge on PR #7")))
		Expect(err).To(MatchError(ContainSubstring("auto merge is not allowed")))
	})

	It("does not enable auto-merge in dry-run mode", func() {
		dryOpts := opts
		dryOpts.DryRun = true
		err := pushAfterCommitForTest(context.Background(), dryOpts, &Result{
			Commits: []CommitResult{{Message: "feat: work", Hash: "abc123"}},
		}, newPRDeps(func(github.Options, string, string) error {
			Fail("enableAutoMerge must not be called in dry-run")
			return nil
		}))
		Expect(err).ToNot(HaveOccurred())
	})

	It("does not enable auto-merge when pushing to an existing matching PR", func() {
		deps := pushDeps{
			searchPRs: func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
				return github.PRSearchResults{{Number: 3, Source: "feat/topic", URL: "http://x/3"}}, nil, nil
			},
			rebaseOnto: func(_, _ string) error { return nil },
			gitPush:    func(_, _ string) error { return nil },
			enableAutoMerge: func(github.Options, string, string) error {
				Fail("enableAutoMerge must not be called for an existing PR")
				return nil
			},
		}
		err := pushAfterCommitForTest(context.Background(), opts, &Result{
			Commits: []CommitResult{{Message: "feat: work", Hash: "abc123"}},
		}, deps)
		Expect(err).ToNot(HaveOccurred())
	})
})
