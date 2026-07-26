package commit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/status"
)

var _ = Describe("interactive batch validation", func() {
	It("requires interactive mode", func() {
		err := validateInteractiveOptions(Options{Batch: true})
		Expect(errors.Is(err, ErrBatchRequiresInteractive)).To(BeTrue())
	})

	It("rejects AI summaries during batch collection", func() {
		previous := stdinIsTerminal
		stdinIsTerminal = func() bool { return true }
		DeferCleanup(func() { stdinIsTerminal = previous })

		err := validateInteractiveOptions(Options{Interactive: true, Batch: true, Summary: true})
		Expect(errors.Is(err, ErrBatchWithSummary)).To(BeTrue())
	})
})

var _ = Describe("collectInteractiveBatches", func() {
	var restore func()

	BeforeEach(func() {
		restore = installOrchestratorStubs()
	})

	AfterEach(func() {
		restore()
	})

	It("queues picker selections in order without staging or starting AI", func() {
		files := []status.FileStatus{
			{Path: "api/a.go", State: status.StateUnstaged},
			{Path: "docs/readme.md", State: status.StateUntracked},
			{Path: "web/app.tsx", State: status.StateStaged},
		}
		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: files}, nil
		}
		pickerCalls := 0
		runTreePickerFunc = func(candidates []status.FileStatus, _ string, _ <-chan status.AISummaryUpdate) (treePickerResult, error) {
			pickerCalls++
			paths := make([]string, len(candidates))
			for i := range candidates {
				paths[i] = candidates[i].Path
			}
			sort.Strings(paths)
			switch pickerCalls {
			case 1:
				Expect(paths).To(Equal([]string{"api/a.go", "docs/readme.md", "web/app.tsx"}))
				return treePickerResult{Selected: []string{"api/a.go", "web/app.tsx"}}, nil
			case 2:
				Expect(paths).To(Equal([]string{"docs/readme.md"}))
				return treePickerResult{Selected: []string{"docs/readme.md"}}, nil
			default:
				Fail("picker reopened after every candidate was queued")
				return treePickerResult{}, nil
			}
		}
		resetAllStagedFn = func(string) error {
			Fail("batch collection must not reset the index")
			return nil
		}
		addFilesFunc = func(string, []string) error {
			Fail("batch collection must not stage files")
			return nil
		}
		startCandidateSummariesFunc = func(context.Context, Options, []status.FileStatus) (*candidateSummarySession, error) {
			Fail("batch collection must not start AI summaries")
			return nil, nil
		}

		batches, err := collectInteractiveBatches(context.Background(), Options{WorkDir: "/repo", Interactive: true, Batch: true})

		Expect(err).NotTo(HaveOccurred())
		Expect(batches).To(Equal([][]string{
			{"api/a.go", "web/app.tsx"},
			{"docs/readme.md"},
		}))
		Expect(pickerCalls).To(Equal(2))
	})

	It("uses escape to finish after a batch and cancel before the first", func() {
		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{
				{Path: "one.go", State: status.StateUnstaged},
				{Path: "two.go", State: status.StateUnstaged},
			}}, nil
		}
		pickerCalls := 0
		runTreePickerFunc = func([]status.FileStatus, string, <-chan status.AISummaryUpdate) (treePickerResult, error) {
			pickerCalls++
			if pickerCalls == 1 {
				return treePickerResult{Selected: []string{"one.go"}}, nil
			}
			return treePickerResult{}, ErrInteractiveCancelled
		}

		batches, err := collectInteractiveBatches(context.Background(), Options{WorkDir: "/repo", Interactive: true, Batch: true})

		Expect(err).NotTo(HaveOccurred())
		Expect(batches).To(Equal([][]string{{"one.go"}}))

		pickerCalls = 0
		runTreePickerFunc = func([]status.FileStatus, string, <-chan status.AISummaryUpdate) (treePickerResult, error) {
			return treePickerResult{}, ErrInteractiveCancelled
		}
		_, err = collectInteractiveBatches(context.Background(), Options{WorkDir: "/repo", Interactive: true, Batch: true})
		Expect(errors.Is(err, ErrInteractiveCancelled)).To(BeTrue())
	})
})

var _ = Describe("runInteractiveBatch", func() {
	DescribeTable("collects every batch before generating messages or committing",
		func(dryRun bool, expectedCommitCount string) {
			repo := initCommitRepoForGinkgo()
			DeferCleanup(func() { Expect(os.RemoveAll(repo)).To(Succeed()) })
			writeRepoFile(repo, "one.go", "package one\n")
			writeRepoFile(repo, "two.go", "package two\n")

			restore := installOrchestratorStubs()
			DeferCleanup(restore)
			gatherStatusFunc = func(workDir string) (*status.Result, error) {
				return status.GatherBase(workDir, status.Options{NoRepomap: true})
			}
			resetAllStagedFn = resetAllStaged
			addFilesFunc = addFiles

			initialHead := gitOutputForGinkgo(repo, "rev-parse", "HEAD")
			pickerCalls := 0
			runTreePickerFunc = func([]status.FileStatus, string, <-chan status.AISummaryUpdate) (treePickerResult, error) {
				pickerCalls++
				Expect(gitOutputForGinkgo(repo, "rev-parse", "HEAD")).To(Equal(initialHead), "selection must finish before the first commit")
				if pickerCalls == 1 {
					return treePickerResult{Selected: []string{"one.go"}}, nil
				}
				return treePickerResult{Selected: []string{"two.go"}}, nil
			}

			previousTTY := stdinIsTerminal
			stdinIsTerminal = func() bool { return true }
			DeferCleanup(func() { stdinIsTerminal = previousTTY })
			previousEnv := os.Getenv(testEnvVar)
			Expect(os.Setenv(testEnvVar, "1")).To(Succeed())
			DeferCleanup(func() { restoreEnv(testEnvVar, previousEnv) })
			previousDryRunOutput := dryRunOutput
			dryRunOutput = io.Discard
			DeferCleanup(func() { dryRunOutput = previousDryRunOutput })

			result, err := Run(context.Background(), Options{
				WorkDir:         repo,
				Interactive:     true,
				Batch:           true,
				DryRun:          dryRun,
				Force:           true,
				PrecommitMode:   "skip",
				LintFlag:        "false",
				LintSecretsFlag: "false",
				TidyFlag:        "false",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Commits).To(HaveLen(2))
			Expect(result.Commits[0].Files).To(Equal([]string{"one.go"}))
			Expect(result.Commits[1].Files).To(Equal([]string{"two.go"}))
			Expect(gitOutputForGinkgo(repo, "rev-list", "--count", initialHead+"..HEAD")).To(Equal(expectedCommitCount))
		},
		Entry("live", false, "2"),
		Entry("dry-run", true, "0"),
	)

	It("does not push or hide a later empty batch after an earlier commit", func() {
		repo := initCommitRepoForGinkgo()
		DeferCleanup(func() { Expect(os.RemoveAll(repo)).To(Succeed()) })
		writeRepoFile(repo, "one.go", "package one\n")

		restore := installOrchestratorStubs()
		DeferCleanup(restore)
		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{
				{Path: "one.go", State: status.StateUntracked},
				{Path: "README.md", State: status.StateUnstaged},
			}}, nil
		}
		resetAllStagedFn = resetAllStaged
		addFilesFunc = addFiles
		pickerCalls := 0
		runTreePickerFunc = func([]status.FileStatus, string, <-chan status.AISummaryUpdate) (treePickerResult, error) {
			pickerCalls++
			if pickerCalls == 1 {
				return treePickerResult{Selected: []string{"one.go"}}, nil
			}
			return treePickerResult{Selected: []string{"README.md"}}, nil
		}

		previousTTY := stdinIsTerminal
		stdinIsTerminal = func() bool { return true }
		DeferCleanup(func() { stdinIsTerminal = previousTTY })
		previousEnv := os.Getenv(testEnvVar)
		Expect(os.Setenv(testEnvVar, "1")).To(Succeed())
		DeferCleanup(func() { restoreEnv(testEnvVar, previousEnv) })

		result, err := Run(context.Background(), Options{
			WorkDir:         repo,
			Interactive:     true,
			Batch:           true,
			Push:            true,
			Force:           true,
			PrecommitMode:   "skip",
			LintFlag:        "false",
			LintSecretsFlag: "false",
			TidyFlag:        "false",
		})

		Expect(errors.Is(err, ErrNothingStaged)).To(BeTrue())
		Expect(result.Commits).To(HaveLen(1))
		Expect(result.PushOnly).To(BeFalse())
	})

	It("fails rather than AI-regrouping an authoritative batch", func() {
		repo := initCommitRepoForGinkgo()
		DeferCleanup(func() { Expect(os.RemoveAll(repo)).To(Succeed()) })
		writeRepoFile(repo, "one.go", "package one\n")
		writeRepoFile(repo, "two.go", "package two\n")
		gitOutputForGinkgo(repo, "add", "one.go", "two.go")

		previousAgent := newAgentFunc
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) { return &agentLifecycleProbe{}, nil }
		DeferCleanup(func() { newAgentFunc = previousAgent })
		previousAnalyze := analyzeCommitMessageWithAIFunc
		analyzeCommitMessageWithAIFunc = func(context.Context, models.CommitAnalysis, clickyai.Agent, git.AnalyzeOptions) (models.CommitAnalysis, error) {
			return models.CommitAnalysis{}, fmt.Errorf("request has too many tokens for this model")
		}
		DeferCleanup(func() { analyzeCommitMessageWithAIFunc = previousAnalyze })
		previousGroup := groupChangesByAIFunc
		groupChangesByAIFunc = func(context.Context, Options, stagedSource) ([]commitGroup, error) {
			Fail("an authoritative batch must not be AI-regrouped")
			return nil, nil
		}
		DeferCleanup(func() { groupChangesByAIFunc = previousGroup })

		_, err := runSingleCommit(context.Background(), Options{
			WorkDir:         repo,
			Stage:           StageStaged,
			Batch:           true,
			Force:           true,
			PrecommitMode:   "skip",
			LintSecretsFlag: "false",
			TidyFlag:        "false",
		})

		Expect(err).To(MatchError(ContainSubstring("too many tokens")))
	})
})

func restoreEnv(name, value string) {
	if value == "" {
		Expect(os.Unsetenv(name)).To(Succeed())
		return
	}
	Expect(os.Setenv(name, value)).To(Succeed())
}

func gitOutputForGinkgo(workDir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %v: %s", args, out)
	return strings.TrimSpace(string(out))
}
