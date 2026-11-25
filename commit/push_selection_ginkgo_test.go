package commit

import (
	"bytes"
	"context"
	"errors"
	"strings"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Multiple PR push selection", func() {
	var (
		opts    Options
		deps    pushDeps
		result  *Result
		choice  string
		cancel  bool
		calls   []string
		created []model.CreatePRInput
		output  bytes.Buffer
	)

	BeforeEach(func() {
		opts = Options{WorkDir: GinkgoT().TempDir(), Push: true}
		runGit(opts.WorkDir, "init", "-b", "feat/topic")
		runGit(opts.WorkDir, "commit", "--allow-empty", "-m", "feat: work")
		result = &Result{Commits: []CommitResult{{Message: "feat: work", Hash: "abc123"}}}
		choice, cancel, calls, created = "Create new PR", false, nil, nil
		output.Reset()
		previousOutput, previousSelect := dryRunOutput, promptSelectFunc
		dryRunOutput = &output
		DeferCleanup(func() { dryRunOutput, promptSelectFunc = previousOutput, previousSelect })
		promptSelectFunc = func(_ context.Context, options []promptSelectOption, title string) (promptSelectOption, bool) {
			Expect(title).To(Equal("Multiple open PRs could accept this push — select one:"))
			Expect(options).To(Equal([]promptSelectOption{
				{Index: 0, Label: "#17  First  (feat/a → main)"},
				{Index: 1, Label: "#23  Second  (feat/b → main)"},
				{Index: 2, Label: "Create new PR"},
			}))
			if cancel {
				return promptSelectOption{}, false
			}
			for _, option := range options {
				if strings.Contains(option.Label, choice) {
					return option, true
				}
			}
			Fail("requested choice missing")
			return promptSelectOption{}, false
		}
		deps = defaultPushDeps(GinkgoT().Context(), opts)
		deps.searchPRs = func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
			return github.PRSearchResults{
				{Number: 17, Title: "First", Source: "feat/a", Target: "main"},
				{Number: 23, Title: "Second", Source: "feat/b", Target: "main"},
			}, nil, nil
		}
		deps.isAncestor = func(_, _, _ string) bool { return true }
		deps.defaultBranch = func(github.Options) (string, error) { return "main", nil }
		deps.generatePRPrompt = func(_ context.Context, in PRContentInput) (PRContent, error) {
			calls = append(calls, "generate")
			Expect(in.Commits).To(Equal([]PRCommitInput{{Message: "feat: work"}}))
			return PRContent{Title: "feat: generated", Body: "Generated body", Branch: "feat/generated"}, nil
		}
		deps.rebaseOnto = func(_, branch string) error { calls = append(calls, "rebase:"+branch); return nil }
		deps.gitPush = func(_, refspec string) error { calls = append(calls, "push:"+refspec); return nil }
		deps.createPR = func(_ github.Options, in model.CreatePRInput) (*model.CreatePRResult, error) {
			calls = append(calls, "create")
			created = append(created, in)
			return &model.CreatePRResult{Number: 31, NodeID: "new-node", Base: in.Base}, nil
		}
		deps.enableAutoMerge = func(_ github.Options, pr *model.CreatePRResult, method string) error {
			calls = append(calls, "merge:"+pr.NodeID+":"+method)
			return nil
		}
	})

	It("creates a PR with AI-generated title and body", func() {
		Expect(pushWithDeps(context.Background(), opts, result, deps)).To(Succeed())
		Expect(calls).To(Equal([]string{"generate", "rebase:main", "push:HEAD:feat/topic", "create"}))
		Expect(created).To(Equal([]model.CreatePRInput{{Title: "feat: generated", Body: "Generated body", Head: "feat/topic", Base: "main"}}))
	})

	It("pushes to the selected existing PR without generating content", func() {
		choice = "#23"
		Expect(pushWithDeps(context.Background(), opts, result, deps)).To(Succeed())
		Expect(calls).To(Equal([]string{"rebase:feat/b", "push:HEAD:feat/b"}))
	})

	It("cancels without generating, pushing, or creating", func() {
		cancel = true
		Expect(pushWithDeps(context.Background(), opts, result, deps)).To(MatchError("push cancelled"))
		Expect(calls).To(BeEmpty())
	})

	It("stops before pushing when AI generation fails", func() {
		deps.generatePRPrompt = func(context.Context, PRContentInput) (PRContent, error) {
			return PRContent{}, errors.New("generation failed")
		}
		Expect(pushWithDeps(context.Background(), opts, result, deps)).To(MatchError("generate PR title/body: generation failed"))
		Expect(calls).To(BeEmpty())
	})

	It("stops before side effects when the picker fails", func() {
		deps.pickPR = func(context.Context, string, []github.PRListItem) (pushTarget, error) {
			return pushTarget{}, errors.New("picker failed")
		}
		Expect(pushWithDeps(context.Background(), opts, result, deps)).To(MatchError("pick PR: picker failed"))
		Expect(calls).To(BeEmpty())
	})

	It("previews AI-generated content without publishing in dry-run", func() {
		opts.DryRun = true
		Expect(pushWithDeps(context.Background(), opts, result, deps)).To(Succeed())
		Expect(calls).To(Equal([]string{"generate"}))
		Expect(output.String()).To(ContainSubstring("title: feat: generated\nbody:\nGenerated body"))
	})

	It("enables auto-merge on the newly selected PR", func() {
		opts.AutoMerge, opts.MergeType = true, "squash"
		Expect(pushWithDeps(context.Background(), opts, result, deps)).To(Succeed())
		Expect(calls).To(Equal([]string{"generate", "rebase:main", "push:HEAD:feat/topic", "create", "merge:new-node:squash"}))
	})
})
