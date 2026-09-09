package commit

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/status"
)

// --- Orchestrator --------------------------------------------------------

var _ = Describe("runInteractiveStaging", func() {
	var (
		restore func()
		opts    Options
	)

	BeforeEach(func() {
		opts = Options{WorkDir: "/repo", Interactive: true}
		restore = installOrchestratorStubs()
	})

	AfterEach(func() {
		restore()
	})

	It("stages exactly the user-selected paths", func() {
		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{
				{Path: "a.go", State: status.StateUnstaged},
				{Path: "b.tsx", State: status.StateUntracked},
				{Path: "c.md", State: status.StateStaged},
			}}, nil
		}
		runTreePickerFunc = func([]status.FileStatus, string, <-chan status.AISummaryUpdate) (treePickerResult, error) {
			return treePickerResult{Selected: []string{"a.go", "c.md"}}, nil
		}
		var resetCalled bool
		var added []string
		resetAllStagedFn = func(string) error { resetCalled = true; return nil }
		addFilesFunc = func(_ string, files []string) error { added = files; return nil }

		paths, err := runInteractiveStaging(context.TODO(), opts)
		Expect(err).ToNot(HaveOccurred())
		Expect(resetCalled).To(BeTrue())
		sort.Strings(added)
		Expect(added).To(Equal([]string{"a.go", "c.md"}))
		sort.Strings(paths)
		Expect(paths).To(Equal([]string{"a.go", "c.md"}))
	})

	It("runs git rm --cached for tracked-but-now-ignored paths before staging", func() {
		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{
				{Path: "a.go", State: status.StateUnstaged},
				{Path: "old.log", State: status.StateStaged},
			}}, nil
		}
		var pickerGitRoot string
		runTreePickerFunc = func(_ []status.FileStatus, gitRoot string, _ <-chan status.AISummaryUpdate) (treePickerResult, error) {
			pickerGitRoot = gitRoot
			return treePickerResult{
				Selected: []string{"a.go"},
				RmCached: []string{"old.log"},
			}, nil
		}
		var rmPaths []string
		var rmCalledBeforeAdd bool
		var addCalled bool
		gitRmCachedFunc = func(_ string, paths []string) error {
			rmPaths = paths
			rmCalledBeforeAdd = !addCalled
			return nil
		}
		addFilesFunc = func(_ string, _ []string) error { addCalled = true; return nil }

		_, err := runInteractiveStaging(context.TODO(), opts)
		Expect(err).ToNot(HaveOccurred())
		Expect(pickerGitRoot).To(Equal(opts.WorkDir))
		Expect(rmPaths).To(Equal([]string{"old.log"}))
		Expect(rmCalledBeforeAdd).To(BeTrue())
	})

	It("returns ErrInteractiveEmpty when no candidates exist", func() {
		gatherStatusFunc = func(string) (*status.Result, error) { return &status.Result{}, nil }
		_, err := runInteractiveStaging(context.TODO(), opts)
		Expect(errors.Is(err, ErrNothingStaged)).To(BeTrue())
	})

	It("returns ErrInteractiveEmpty when picker returns no selection", func() {
		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{{Path: "x"}}}, nil
		}
		runTreePickerFunc = func([]status.FileStatus, string, <-chan status.AISummaryUpdate) (treePickerResult, error) {
			return treePickerResult{}, nil
		}
		_, err := runInteractiveStaging(context.TODO(), opts)
		Expect(errors.Is(err, ErrInteractiveEmpty)).To(BeTrue())
	})

	It("propagates ErrInteractiveCancelled from the picker", func() {
		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{{Path: "x"}}}, nil
		}
		runTreePickerFunc = func([]status.FileStatus, string, <-chan status.AISummaryUpdate) (treePickerResult, error) {
			return treePickerResult{}, ErrInteractiveCancelled
		}
		_, err := runInteractiveStaging(context.TODO(), opts)
		Expect(errors.Is(err, ErrInteractiveCancelled)).To(BeTrue())
	})

	It("filters out conflict files from the candidate list", func() {
		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{
				{Path: "ok.go", State: status.StateUnstaged},
				{Path: "bad.go", State: status.StateConflict},
			}}, nil
		}
		var candidatesSeen []status.FileStatus
		runTreePickerFunc = func(c []status.FileStatus, _ string, _ <-chan status.AISummaryUpdate) (treePickerResult, error) {
			candidatesSeen = c
			return treePickerResult{Selected: []string{"ok.go"}}, nil
		}
		_, err := runInteractiveStaging(context.TODO(), opts)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(candidatesSeen)).To(Equal(1))
		Expect(candidatesSeen[0].Path).To(Equal("ok.go"))
	})

	It("streams --summary updates for non-conflicting candidates into the picker", func() {
		opts.Summary = true
		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{
				{Path: "kept.go", State: status.StateUnstaged},
				{Path: "ignored.md", State: status.StateConflict},
			}}, nil
		}
		stopped := false
		startCandidateSummariesFunc = func(_ context.Context, _ Options, files []status.FileStatus) (*candidateSummarySession, error) {
			Expect(files).To(HaveLen(1))
			Expect(files[0].Path).To(Equal("kept.go"))
			prepared := &status.Result{Files: append([]status.FileStatus(nil), files...)}
			prepared.PrepareAISummaries()
			updates := make(chan status.AISummaryUpdate, 1)
			updates <- status.AISummaryUpdate{Index: 0, Status: status.AISummaryStatusDone, Summary: "tighten commit flow"}
			close(updates)
			return &candidateSummarySession{
				Files:   prepared.Files,
				Updates: updates,
				Stop:    func() { stopped = true },
			}, nil
		}
		runTreePickerFunc = func(files []status.FileStatus, _ string, updates <-chan status.AISummaryUpdate) (treePickerResult, error) {
			Expect(files).To(HaveLen(1))
			Expect(files[0].AIStatus).To(Equal(status.AISummaryStatusPending))
			update := <-updates
			Expect(update.Summary).To(Equal("tighten commit flow"))
			return treePickerResult{Selected: []string{"kept.go"}}, nil
		}

		_, err := runInteractiveStaging(context.TODO(), opts)
		Expect(err).ToNot(HaveOccurred())
		Expect(stopped).To(BeTrue())
	})

	It("returns a summary setup error before opening the picker", func() {
		opts.Summary = true
		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{{Path: "kept.go", State: status.StateUnstaged}}}, nil
		}
		setupErr := errors.New("no AI provider")
		startCandidateSummariesFunc = func(context.Context, Options, []status.FileStatus) (*candidateSummarySession, error) {
			return nil, setupErr
		}
		pickerCalled := false
		runTreePickerFunc = func([]status.FileStatus, string, <-chan status.AISummaryUpdate) (treePickerResult, error) {
			pickerCalled = true
			return treePickerResult{}, nil
		}

		_, err := runInteractiveStaging(context.TODO(), opts)
		Expect(err).To(MatchError(ContainSubstring("start candidate summaries")))
		Expect(errors.Is(err, setupErr)).To(BeTrue())
		Expect(pickerCalled).To(BeFalse())
	})
})

// --- Interactive loop ----------------------------------------------------

var _ = Describe("runInteractiveLoop", func() {
	var (
		repo    string
		restore func()
		prevTTY func() bool
		prevEnv string
	)

	BeforeEach(func() {
		repo = initCommitRepoForGinkgo()
		restore = installOrchestratorStubs()
		prevTTY = stdinIsTerminal
		stdinIsTerminal = func() bool { return true }
		prevEnv = os.Getenv(testEnvVar)
		Expect(os.Setenv(testEnvVar, "1")).To(Succeed())
	})

	AfterEach(func() {
		restore()
		stdinIsTerminal = prevTTY
		if prevEnv == "" {
			_ = os.Unsetenv(testEnvVar)
		} else {
			_ = os.Setenv(testEnvVar, prevEnv)
		}
		_ = os.RemoveAll(repo)
	})

	It("commits each user-picked subset and exits when no candidates remain", func() {
		writeRepoFile(repo, "a.go", "package a\n")
		writeRepoFile(repo, "b.go", "package b\n")

		// First iteration: gather sees both, user picks a.go.
		// Second iteration: gather sees b.go (after a.go committed), user picks b.go.
		// Third iteration: gather returns nothing → loop exits cleanly.
		callsGather := 0
		gatherStatusFunc = func(string) (*status.Result, error) {
			callsGather++
			switch callsGather {
			case 1:
				return &status.Result{Files: []status.FileStatus{
					{Path: "a.go", State: status.StateUntracked},
					{Path: "b.go", State: status.StateUntracked},
				}}, nil
			case 2:
				return &status.Result{Files: []status.FileStatus{
					{Path: "b.go", State: status.StateUntracked},
				}}, nil
			default:
				return &status.Result{}, nil
			}
		}
		picks := [][]string{{"a.go"}, {"b.go"}}
		callsPicker := 0
		runTreePickerFunc = func([]status.FileStatus, string, <-chan status.AISummaryUpdate) (treePickerResult, error) {
			out := picks[callsPicker]
			callsPicker++
			return treePickerResult{Selected: out}, nil
		}
		// The orchestrator stubs replace addFiles/resetAllStaged with no-ops,
		// so we provide implementations that touch the real repo.
		resetAllStagedFn = resetAllStaged
		addFilesFunc = addFiles

		result, err := Run(context.Background(), Options{
			WorkDir:     repo,
			Interactive: true,
			Force:       true,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(result).ToNot(BeNil())
		Expect(result.Commits).To(HaveLen(2))
		Expect(callsPicker).To(Equal(2))
		Expect(callsGather).To(Equal(3))
	})

	It("exits cleanly when the user cancels after committing at least once", func() {
		writeRepoFile(repo, "a.go", "package a\n")
		writeRepoFile(repo, "b.go", "package b\n")

		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{
				{Path: "a.go", State: status.StateUntracked},
				{Path: "b.go", State: status.StateUntracked},
			}}, nil
		}
		callsPicker := 0
		runTreePickerFunc = func([]status.FileStatus, string, <-chan status.AISummaryUpdate) (treePickerResult, error) {
			callsPicker++
			if callsPicker == 1 {
				return treePickerResult{Selected: []string{"a.go"}}, nil
			}
			return treePickerResult{}, ErrInteractiveCancelled
		}
		resetAllStagedFn = resetAllStaged
		addFilesFunc = addFiles

		result, err := Run(context.Background(), Options{
			WorkDir:     repo,
			Interactive: true,
			Force:       true,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(result.Commits).To(HaveLen(1))
		Expect(callsPicker).To(Equal(2))
	})

	It("surfaces ErrInteractiveCancelled if the user cancels on the first iteration", func() {
		writeRepoFile(repo, "a.go", "package a\n")

		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{
				{Path: "a.go", State: status.StateUntracked},
			}}, nil
		}
		runTreePickerFunc = func([]status.FileStatus, string, <-chan status.AISummaryUpdate) (treePickerResult, error) {
			return treePickerResult{}, ErrInteractiveCancelled
		}

		_, err := Run(context.Background(), Options{
			WorkDir:     repo,
			Interactive: true,
			Force:       true,
		})
		Expect(errors.Is(err, ErrInteractiveCancelled)).To(BeTrue())
	})
})

// --- Helpers --------------------------------------------------------------

func initCommitRepoForGinkgo() string {
	dir, err := os.MkdirTemp("", "gavel-commit-loop-")
	Expect(err).ToNot(HaveOccurred())
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test User"},
		{"config", "commit.gpgsign", "false"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		Expect(err).ToNot(HaveOccurred(), "git %v: %s", args, out)
	}
	Expect(os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test\n"), 0o644)).To(Succeed())
	for _, args := range [][]string{
		{"add", "README.md"},
		{"commit", "-m", "initial"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		Expect(err).ToNot(HaveOccurred(), "git %v: %s", args, out)
	}
	return dir
}

func writeRepoFile(dir, name, content string) {
	Expect(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)).To(Succeed())
}

func updateKey(m treeModel, key string) (treeModel, tea.Cmd) {
	msg := tea.KeyMsg{Runes: []rune(key)}
	switch key {
	case "up":
		msg.Type = tea.KeyUp
		msg.Runes = nil
	case "down":
		msg.Type = tea.KeyDown
		msg.Runes = nil
	case "left":
		msg.Type = tea.KeyLeft
		msg.Runes = nil
	case "right":
		msg.Type = tea.KeyRight
		msg.Runes = nil
	case "enter":
		msg.Type = tea.KeyEnter
		msg.Runes = nil
	case "esc":
		msg.Type = tea.KeyEsc
		msg.Runes = nil
	case "backspace":
		msg.Type = tea.KeyBackspace
		msg.Runes = nil
	case "ctrl+u":
		msg.Type = tea.KeyCtrlU
		msg.Runes = nil
	case "ctrl+c":
		msg.Type = tea.KeyCtrlC
		msg.Runes = nil
	case " ":
		msg.Type = tea.KeySpace
	default:
		msg.Type = tea.KeyRunes
	}
	out, cmd := m.handleKey(msg)
	return out.(treeModel), cmd
}

func rootChildNames(m treeModel) []string {
	out := make([]string, len(m.root.Children))
	for i, c := range m.root.Children {
		out[i] = c.Name
	}
	return out
}

func visiblePathList(m treeModel) []string {
	out := make([]string, len(m.visible))
	for i, n := range m.visible {
		out[i] = n.Path
	}
	return out
}

func nodeAtPath(m treeModel, path string) *treeNode {
	var found *treeNode
	var walk func(n *treeNode)
	walk = func(n *treeNode) {
		if n.Path == path {
			found = n
			return
		}
		for _, c := range n.Children {
			if found != nil {
				return
			}
			walk(c)
		}
	}
	walk(m.root)
	return found
}

func installOrchestratorStubs() func() {
	prevGather := gatherStatusFunc
	prevReset := resetAllStagedFn
	prevAdd := addFilesFunc
	prevRm := gitRmCachedFunc
	prevPicker := runTreePickerFunc
	prevSummaries := startCandidateSummariesFunc
	gatherStatusFunc = func(string) (*status.Result, error) { return &status.Result{}, nil }
	resetAllStagedFn = func(string) error { return nil }
	addFilesFunc = func(string, []string) error { return nil }
	gitRmCachedFunc = func(string, []string) error { return nil }
	runTreePickerFunc = func([]status.FileStatus, string, <-chan status.AISummaryUpdate) (treePickerResult, error) {
		return treePickerResult{}, nil
	}
	startCandidateSummariesFunc = func(_ context.Context, _ Options, files []status.FileStatus) (*candidateSummarySession, error) {
		return &candidateSummarySession{Files: files, Stop: func() {}}, nil
	}
	return func() {
		gatherStatusFunc = prevGather
		resetAllStagedFn = prevReset
		addFilesFunc = prevAdd
		gitRmCachedFunc = prevRm
		runTreePickerFunc = prevPicker
		startCandidateSummariesFunc = prevSummaries
	}
}
