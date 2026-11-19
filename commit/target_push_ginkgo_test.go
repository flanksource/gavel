package commit

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("commit push target", func() {
	var (
		workDir                       string
		deps                          pushDeps
		created                       []model.CreatePRInput
		rebased, pushed, repositories []string
	)

	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
		runGit(workDir, "init", "-b", "main")
		runGit(workDir, "commit", "--allow-empty", "-m", "root")
		runGit(workDir, "remote", "add", "origin", "https://github.com/acme/service.git")
		runGit(workDir, "remote", "add", "zim", "https://github.com/acme-fork/service.git")
		runGit(workDir, "update-ref", "refs/remotes/zim/stable", "HEAD")
		runGit(workDir, "checkout", "-b", "feature")
		runGit(workDir, "commit", "--allow-empty", "-m", "feat: first")
		runGit(workDir, "update-ref", "refs/remotes/origin/feature", "HEAD")
		created, rebased, pushed, repositories = nil, nil, nil, nil
		deps = pushDeps{
			searchPRs: func(opts github.Options, _ github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
				repositories = append(repositories, opts.Repo)
				return github.PRSearchResults{}, nil, nil
			},
			defaultBranch: func(github.Options) (string, error) { return "main", nil },
			isAncestor:    func(_, _, _ string) bool { return false },
			generatePRPrompt: func(_ context.Context, in PRContentInput) (PRContent, error) {
				Expect(in.Commits).To(Equal([]PRCommitInput{{Message: "feat: first"}}))
				return PRContent{Title: "feat: first", Branch: "feat/generated"}, nil
			},
			rebaseOnto: func(_, branch string) error { rebased = append(rebased, branch); return nil },
			gitPush:    func(_, refspec string) error { pushed = append(pushed, refspec); return nil },
			createPR: func(opts github.Options, in model.CreatePRInput) (*model.CreatePRResult, error) {
				repositories = append(repositories, opts.Repo)
				created = append(created, in)
				return &model.CreatePRResult{Number: 7, Base: in.Base}, nil
			},
		}
	})

	It("opens a PR on the selected repository and base using the explicit topic, even when already pushed elsewhere", func() {
		opts := Options{WorkDir: workDir, Push: true, Target: "zim/stable:feat/selected"}
		Expect(pushAfterCommitForTest(context.Background(), opts, &Result{}, deps)).To(Succeed())
		Expect(created).To(Equal([]model.CreatePRInput{{Title: "feat: first", Head: "feat/selected", Base: "stable"}}))
		Expect(repositories).To(Equal([]string{"acme-fork/service", "acme-fork/service"}))
		Expect(rebased).To(Equal([]string{"stable"}))
		Expect(pushed).To(Equal([]string{"HEAD:feat/selected"}))
	})

	It("keeps the current topic when no topic override is given", func() {
		Expect(pushAfterCommitForTest(context.Background(), Options{WorkDir: workDir, Push: true, Target: "zim/stable"}, &Result{}, deps)).To(Succeed())
		Expect(created[0].Head).To(Equal("feature"))
		Expect(created[0].Base).To(Equal("stable"))
	})

	It("previews a staged commit even when HEAD has no commits ahead of the selected base", func() {
		var output bytes.Buffer
		previousOutput := dryRunOutput
		dryRunOutput = &output
		DeferCleanup(func() { dryRunOutput = previousOutput })
		runGit(workDir, "update-ref", "refs/remotes/zim/stable", "HEAD")
		deps.generatePRPrompt = func(_ context.Context, in PRContentInput) (PRContent, error) {
			Expect(in.Commits).To(Equal([]PRCommitInput{{Message: "fix: staged", Files: []string{"feature.go"}}}))
			return PRContent{Title: "fix: staged", Branch: "fix/staged"}, nil
		}
		opts := Options{WorkDir: workDir, Push: true, DryRun: true, Target: "zim/stable:fix/staged"}
		result := &Result{DryRun: true, Commits: []CommitResult{{Message: "fix: staged", Files: []string{"feature.go"}}}}
		Expect(pushAfterCommitForTest(context.Background(), opts, result, deps)).To(Succeed())
		Expect(pushed).To(BeEmpty())
		Expect(created).To(BeEmpty())
		Expect(output.String()).To(ContainSubstring("would push HEAD:fix/staged and open PR against stable (remote zim)"))
	})

	It("does not select a PR against a different base or with a different explicit topic", func() {
		deps.searchPRs = func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
			return github.PRSearchResults{
				{Number: 1, Source: "feat/selected", Target: "main"},
				{Number: 2, Source: "other", Target: "stable"},
			}, nil, nil
		}
		deps.isAncestor = func(_, _, _ string) bool { return true }
		Expect(pushAfterCommitForTest(context.Background(), Options{WorkDir: workDir, Push: true, Target: "zim/stable:feat/selected"}, &Result{}, deps)).To(Succeed())
		Expect(created).To(HaveLen(1))
	})

	It("updates a PR matching both the explicit topic and base", func() {
		deps.searchPRs = func(github.Options, github.PRSearchOptions) (github.PRSearchResults, *github.RateLimit, error) {
			return github.PRSearchResults{{Number: 3, Source: "feat/selected", Target: "stable"}}, nil, nil
		}
		Expect(pushAfterCommitForTest(context.Background(), Options{WorkDir: workDir, Push: true, Target: "zim/stable:feat/selected"}, &Result{}, deps)).To(Succeed())
		Expect(created).To(BeEmpty())
		Expect(pushed).To(Equal([]string{"HEAD:feat/selected"}))
	})

	It("does not push an AI-generated topic directly onto the selected base", func() {
		runGit(workDir, "branch", "-m", "stable")
		deps.generatePRPrompt = func(context.Context, PRContentInput) (PRContent, error) {
			return PRContent{Title: "feat: first", Branch: "stable"}, nil
		}
		err := pushAfterCommitForTest(context.Background(), Options{WorkDir: workDir, Push: true, Target: "zim/stable"}, &Result{}, deps)
		Expect(err).To(MatchError(ContainSubstring("must differ from base")))
		Expect(pushed).To(BeEmpty())
	})

	It("fails before changing the index or HEAD when the target is invalid or push is disabled", func() {
		Expect(os.WriteFile(filepath.Join(workDir, "staged.txt"), []byte("staged"), 0600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, "other.txt"), []byte("other"), 0600)).To(Succeed())
		runGit(workDir, "add", "staged.txt")
		for _, opts := range []Options{
			{WorkDir: workDir, Target: "main", Message: "feat: change", Files: []string{"other.txt"}},
			{WorkDir: workDir, Push: true, Target: "missing/main", Message: "feat: change", Files: []string{"other.txt"}},
		} {
			indexBefore, err := exec.Command("git", "-C", workDir, "status", "--porcelain=v1").Output()
			Expect(err).NotTo(HaveOccurred())
			before, err := exec.Command("git", "-C", workDir, "rev-parse", "HEAD").Output()
			Expect(err).NotTo(HaveOccurred())
			_, err = Run(context.Background(), opts)
			Expect(err).To(MatchError(ContainSubstring("--target")))
			after, err := exec.Command("git", "-C", workDir, "rev-parse", "HEAD").Output()
			Expect(err).NotTo(HaveOccurred())
			Expect(after).To(Equal(before))
			indexAfter, err := exec.Command("git", "-C", workDir, "status", "--porcelain=v1").Output()
			Expect(err).NotTo(HaveOccurred())
			Expect(indexAfter).To(Equal(indexBefore))
		}
	})

	It("validates the provider of the selected remote before fetching or committing", func() {
		runGit(workDir, "remote", "set-url", "zim", "https://dev.azure.com/acme/product/_git/service")
		mockAzurePush(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.URL.Path).To(Equal("/acme/product/_apis/git/repositories/service"))
			w.WriteHeader(http.StatusForbidden)
		})
		_, err := Run(context.Background(), Options{WorkDir: workDir, Push: true, Target: "zim/stable", AutoMerge: true})
		Expect(err).To(MatchError(ContainSubstring("HTTP 403")))
	})

	It("fetches, rebases and pushes only the selected local remote", func() {
		remote := filepath.Join(GinkgoT().TempDir(), "remote.git")
		runGit(workDir, "init", "--bare", remote)
		runGit(workDir, "remote", "set-url", "zim", remote)
		runGit(workDir, "checkout", "main")
		Expect(os.WriteFile(filepath.Join(workDir, "base.txt"), []byte("advanced remote base"), 0600)).To(Succeed())
		runGit(workDir, "add", "base.txt")
		runGit(workDir, "commit", "-m", "chore: advance base")
		runGit(workDir, "push", "zim", "main:stable")
		runGit(workDir, "checkout", "feature")
		opts := Options{WorkDir: workDir, Push: true, Target: "zim/stable:feat/selected"}
		var err error
		opts.target, err = parseCommitTarget(workDir, opts.Target)
		Expect(err).NotTo(HaveOccurred())
		actual := defaultPushDeps(GinkgoT().Context(), opts)
		Expect(actual.rebaseOnto(workDir, "stable")).To(Succeed())
		Expect(os.ReadFile(filepath.Join(workDir, "base.txt"))).To(Equal([]byte("advanced remote base")))
		Expect(actual.gitPush(workDir, "HEAD:feat/selected")).To(Succeed())
		out, err := exec.Command("git", "--git-dir", remote, "rev-parse", "refs/heads/feat/selected").Output()
		Expect(err).NotTo(HaveOccurred())
		head, err := exec.Command("git", "-C", workDir, "rev-parse", "HEAD").Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal(head))
	})
})
