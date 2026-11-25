package commit

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/clicky/prompt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("explicit ignored commit paths", func() {
	dashboardSelect := promptSelectFunc
	var repo string
	var prompts int
	BeforeEach(func() {
		repo = initCommitRepo(GinkgoT())
		previousTTY, previousPrompt := stdinIsTerminal, promptSelectFunc
		DeferCleanup(func() { stdinIsTerminal, promptSelectFunc = previousTTY, previousPrompt })
		stdinIsTerminal = func() bool { return true }
		prompts = 0
		promptSelectFunc = func(_ context.Context, choices []promptSelectOption, title string) (promptSelectOption, bool) {
			prompts++
			Expect(title).To(ContainSubstring("build"))
			Expect(choices[0].Label).To(Equal("Force-add and continue"))
			return choices[0], true
		}
	})

	DescribeTable("force-adds only the approved paths", func(ignoredPath string, withDot bool) {
		writeFile(GinkgoT(), repo, ".gitignore", "build\nother-output/\n")
		Expect(os.MkdirAll(filepath.Dir(filepath.Join(repo, ignoredPath)), 0o755)).To(Succeed())
		writeFile(GinkgoT(), repo, ignoredPath, "artifact\n")
		Expect(os.MkdirAll(filepath.Join(repo, "other-output"), 0o755)).To(Succeed())
		writeFile(GinkgoT(), repo, "other-output/ignored.txt", "unrelated\n")
		writeFile(GinkgoT(), repo, "source.txt", "source\n")
		paths := []string{ignoredPath, "source.txt"}
		if withDot {
			paths = []string{"build", "."}
		}
		Expect(stageExplicitFiles(context.Background(), repo, paths)).To(Succeed())
		Expect(prompts).To(Equal(1))
		staged, err := stagedFiles(repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(staged).To(ContainElements(ignoredPath, "source.txt"))
		Expect(staged).NotTo(ContainElement("other-output/ignored.txt"))
	},
		Entry("ignored file", "build", false),
		Entry("file with spaces in ignored directory", "build/output file.txt", false),
		Entry("ignored directory alongside dot", "build/output.txt", true),
	)

	DescribeTable("preserves the index when approval is unavailable", func(tty, accept bool) {
		writeFile(GinkgoT(), repo, ".gitignore", "build\n")
		writeFile(GinkgoT(), repo, "build", "artifact\n")
		writeFile(GinkgoT(), repo, "existing.txt", "staged\n")
		gitRun(GinkgoT(), repo, "add", "existing.txt")
		before := gitOutput(GinkgoT(), repo, "diff", "--cached")
		stdinIsTerminal = func() bool { return tty }
		promptSelectFunc = func(_ context.Context, choices []promptSelectOption, _ string) (promptSelectOption, bool) {
			prompts++
			return choices[1], accept
		}
		err := stageExplicitFiles(context.Background(), repo, []string{"build"})
		Expect(err).To(HaveOccurred())
		Expect(gitOutput(GinkgoT(), repo, "diff", "--cached")).To(Equal(before))
		if !tty {
			Expect(err.Error()).To(ContainSubstring("git add -f"))
			Expect(prompts).To(BeZero())
		}
	},
		Entry("cancel choice", true, true),
		Entry("dismissed prompt", true, false),
		Entry("noninteractive execution", false, false),
	)

	It("honors repository exclude rules", func() {
		writeFile(GinkgoT(), repo, ".git/info/exclude", "build\n")
		writeFile(GinkgoT(), repo, "build", "artifact\n")
		Expect(stageExplicitFiles(context.Background(), repo, []string{"build"})).To(Succeed())
		Expect(prompts).To(Equal(1))
		staged, err := stagedFiles(repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(staged).To(Equal([]string{"build"}))
	})

	It("routes approval to the dashboard when stdin is not a terminal", func() {
		writeFile(GinkgoT(), repo, ".gitignore", "build\n")
		writeFile(GinkgoT(), repo, "build", "artifact\n")
		stdinIsTerminal = func() bool { return false }
		manager := prompt.NewManager(prompt.NewMemory(prompt.MemoryConfig{}))
		prompt.SetDefault(manager)
		DeferCleanup(func() { prompt.SetDefault(nil) })
		promptSelectFunc = dashboardSelect
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		done := make(chan error, 1)
		go func() { done <- stageExplicitFiles(ctx, repo, []string{"build"}) }()
		var id string
		Eventually(func() string {
			if pending := manager.List(prompt.Filter{}); len(pending) == 1 {
				id = pending[0].ID
			}
			return id
		}, 5*time.Second).ShouldNot(BeEmpty())
		Expect(manager.Resolve(id, prompt.Answer{Values: map[string]any{"choice": "0"}})).To(Succeed())
		Eventually(done, 5*time.Second).Should(Receive(BeNil()))
	})

	It("does not prompt for an already tracked ignored file", func() {
		writeFile(GinkgoT(), repo, ".gitignore", "build\n")
		writeFile(GinkgoT(), repo, "build", "first\n")
		gitRun(GinkgoT(), repo, "add", "-f", "build")
		gitRun(GinkgoT(), repo, "commit", "-m", "feat: initial artifact")
		writeFile(GinkgoT(), repo, "build", "changed\n")
		Expect(stageExplicitFiles(context.Background(), repo, []string{"build"})).To(Succeed())
		Expect(prompts).To(BeZero())
	})

	DescribeTable("commits an approved ignored path", func(grouped bool) {
		GinkgoT().Setenv(testEnvVar, "1")
		previous := groupChangesByAIFunc
		DeferCleanup(func() { groupChangesByAIFunc = previous })
		groupChangesByAIFunc = func(_ context.Context, _ Options, source stagedSource) ([]commitGroup, error) {
			return []commitGroup{{Label: "artifact", Changes: source.Changes}}, nil
		}
		writeFile(GinkgoT(), repo, ".gitignore", "build\n")
		writeFile(GinkgoT(), repo, "build", "artifact\n")
		result, err := Run(context.Background(), Options{WorkDir: repo, Files: []string{"build"}, CommitAll: grouped, LintSecretsFlag: "false"})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Commits).To(HaveLen(1))
		Expect(result.Commits[0].Hash).NotTo(BeEmpty())
		Expect(gitOutput(GinkgoT(), repo, "diff-tree", "--no-commit-id", "--name-only", "-r", "HEAD")).To(Equal("build\n"))
	}, Entry("single commit", false), Entry("grouped commit", true))
})
