package commit

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/status"
	"github.com/flanksource/repomap"
)

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

func timeMinus(d time.Duration) time.Time {
	return time.Now().Add(-d)
}

// --- Validation -----------------------------------------------------------

var _ = Describe("validateInteractiveOptions", func() {
	It("rejects -i with -A", func() {
		err := validateInteractiveOptions(Options{Interactive: true, CommitAll: true})
		Expect(errors.Is(err, ErrInteractiveWithCommitAll)).To(BeTrue())
	})

	It("rejects -i with -m", func() {
		err := validateInteractiveOptions(Options{Interactive: true, Message: "chore: x"})
		Expect(errors.Is(err, ErrInteractiveWithMessage)).To(BeTrue())
	})

	It("rejects -i in a non-TTY environment", func() {
		previous := stdinIsTerminal
		stdinIsTerminal = func() bool { return false }
		defer func() { stdinIsTerminal = previous }()

		err := validateInteractiveOptions(Options{Interactive: true})
		Expect(errors.Is(err, ErrInteractiveNonTTY)).To(BeTrue())
	})

	It("ignores --summary when -i is not set", func() {
		err := validateInteractiveOptions(Options{Summary: true})
		Expect(err).ToNot(HaveOccurred())
	})

	It("accepts -i in a TTY", func() {
		previous := stdinIsTerminal
		stdinIsTerminal = func() bool { return true }
		defer func() { stdinIsTerminal = previous }()

		err := validateInteractiveOptions(Options{Interactive: true})
		Expect(err).ToNot(HaveOccurred())
	})
})

// --- Tree model construction ---------------------------------------------

var _ = Describe("treeModel construction", func() {
	files := []status.FileStatus{
		{Path: "cmd/gavel/main.go", Adds: 5, FileMap: &repomap.FileMap{Language: "Go"}},
		{Path: "cmd/gavel/commit.go", Adds: 10, FileMap: &repomap.FileMap{Language: "Go"}},
		{Path: "ui/src/App.tsx", Adds: 20, FileMap: &repomap.FileMap{Language: "TypeScript"}},
		{Path: "ui/src/App_test.tsx", Adds: 7, FileMap: &repomap.FileMap{Language: "TypeScript", Scopes: []repomap.ScopeType{repomap.ScopeTypeTest}}},
		{Path: "docs/readme.md", Adds: 1, FileMap: &repomap.FileMap{Language: "Markdown"}},
		{Path: "newfile", FileMap: nil},
	}

	It("builds a folder tree with directories sorted before files", func() {
		m := newTreeModel(files)
		// root.Children: cmd/, docs/, ui/ (dirs alphabetical) then newfile (file).
		Expect(rootChildNames(m)).To(Equal([]string{"cmd", "docs", "ui", "newfile"}))
	})

	It("renders all visible nodes when expanded by default", func() {
		m := newTreeModel(files)
		// 3 dirs + 1 file at depth 1; all dirs expanded so children show.
		visiblePaths := visiblePathList(m)
		Expect(visiblePaths).To(ContainElements(
			"cmd", "cmd/gavel", "cmd/gavel/main.go", "cmd/gavel/commit.go",
			"ui", "ui/src", "ui/src/App.tsx", "ui/src/App_test.tsx",
			"docs", "docs/readme.md", "newfile",
		))
	})

	It("counts leaves correctly", func() {
		m := newTreeModel(files)
		Expect(countLeaves(m.root)).To(Equal(len(files)))
	})
})

// --- Selection propagation ------------------------------------------------

var _ = Describe("treeModel selection propagation", func() {
	makeModel := func() treeModel {
		return newTreeModel([]status.FileStatus{
			{Path: "cmd/gavel/main.go", FileMap: &repomap.FileMap{Language: "Go"}},
			{Path: "cmd/gavel/commit.go", FileMap: &repomap.FileMap{Language: "Go"}},
			{Path: "ui/App.tsx", FileMap: &repomap.FileMap{Language: "TypeScript"}},
			{Path: "ui/App_test.tsx", FileMap: &repomap.FileMap{Language: "TypeScript", Scopes: []repomap.ScopeType{repomap.ScopeTypeTest}}},
		})
	}

	It("toggling a folder selects all leaf descendants", func() {
		m := makeModel()
		cmdNode := nodeAtPath(m, "cmd")
		Expect(cmdNode).ToNot(BeNil())
		toggleNode(cmdNode)
		Expect(m.selectedPaths()).To(ConsistOf("cmd/gavel/main.go", "cmd/gavel/commit.go"))
		// Toggle again deselects all.
		toggleNode(cmdNode)
		Expect(m.selectedPaths()).To(BeEmpty())
	})

	It("checkbox shows tri-state on partially-selected folders", func() {
		m := makeModel()
		nodeAtPath(m, "cmd/gavel/main.go").Selected = true
		Expect(stripANSI(checkboxFor(nodeAtPath(m, "cmd")))).To(Equal("[~]"))
		nodeAtPath(m, "cmd/gavel/commit.go").Selected = true
		Expect(stripANSI(checkboxFor(nodeAtPath(m, "cmd")))).To(Equal("[x]"))
	})

	It("toggleByLanguage flips all matching files", func() {
		m := makeModel()
		m.toggleByLanguage("Go")
		Expect(m.selectedPaths()).To(ConsistOf("cmd/gavel/main.go", "cmd/gavel/commit.go"))
		m.toggleByLanguage("Go")
		Expect(m.selectedPaths()).To(BeEmpty())
	})

	It("toggleByScope flips all files in that scope", func() {
		m := makeModel()
		m.toggleByScope(repomap.ScopeTypeTest)
		Expect(m.selectedPaths()).To(ConsistOf("ui/App_test.tsx"))
	})
})

// --- Key handling ---------------------------------------------------------

var _ = Describe("treeModel key handling", func() {
	files := []status.FileStatus{
		{Path: "a/x.go", FileMap: &repomap.FileMap{Language: "Go"}},
		{Path: "b/y.go", FileMap: &repomap.FileMap{Language: "Go"}},
	}

	It("space toggles the file under the cursor", func() {
		m := newTreeModel(files)
		// Move cursor down to a file (skip the 'a' dir at index 0).
		m, _ = updateKey(m, "down")
		m, _ = updateKey(m, " ")
		Expect(m.selectedPaths()).To(ConsistOf("a/x.go"))
	})

	It("g toggles all Go files at once", func() {
		m := newTreeModel(files)
		m, _ = updateKey(m, "g")
		Expect(m.selectedPaths()).To(ConsistOf("a/x.go", "b/y.go"))
	})

	It("enter sets submitted; esc sets cancelled", func() {
		m := newTreeModel(files)
		m, _ = updateKey(m, "enter")
		Expect(m.submitted).To(BeTrue())
		Expect(m.cancelled).To(BeFalse())

		m2 := newTreeModel(files)
		m2, _ = updateKey(m2, "esc")
		Expect(m2.cancelled).To(BeTrue())
	})

	It("left collapses an expanded directory", func() {
		m := newTreeModel(files)
		// cursor is on 'a' dir; ensure expanded then collapse with left.
		dir := m.currentNode()
		Expect(dir.IsDir).To(BeTrue())
		Expect(dir.Expanded).To(BeTrue())
		m, _ = updateKey(m, "left")
		Expect(dir.Expanded).To(BeFalse())
		Expect(visiblePathList(m)).To(Equal([]string{"a", "b", "b/y.go"}))
	})

	It("slash opens a live filter over file paths", func() {
		m := newTreeModel([]status.FileStatus{
			{Path: "cmd/gavel/main.go", FileMap: &repomap.FileMap{Language: "Go"}},
			{Path: "ui/App.tsx", FileMap: &repomap.FileMap{Language: "TypeScript"}},
			{Path: "docs/readme.md", FileMap: &repomap.FileMap{Language: "Markdown"}},
		})

		m, _ = updateKey(m, "/")
		Expect(m.filtering).To(BeTrue())
		m, _ = updateKey(m, "tsx")

		Expect(m.filterQuery).To(Equal("tsx"))
		Expect(visiblePathList(m)).To(Equal([]string{"ui", "ui/App.tsx"}))

		m, _ = updateKey(m, "enter")
		Expect(m.filtering).To(BeFalse())
		Expect(m.filterQuery).To(Equal("tsx"))
	})

	It("filter matches status language and scope chips", func() {
		m := newTreeModel([]status.FileStatus{
			{Path: "cmd/gavel/main.go", State: status.StateStaged, FileMap: &repomap.FileMap{Language: "Go"}},
			{Path: "ui/App_test.tsx", State: status.StateUntracked, FileMap: &repomap.FileMap{Language: "TypeScript", Scopes: []repomap.ScopeType{repomap.ScopeTypeTest}}},
			{Path: "docs/readme.md", State: status.StateUnstaged, FileMap: &repomap.FileMap{Language: "Markdown"}},
		})

		m, _ = updateKey(m, "/")
		m, _ = updateKey(m, "test")
		Expect(visiblePathList(m)).To(Equal([]string{"ui", "ui/App_test.tsx"}))

		m, _ = updateKey(m, "ctrl+u")
		m, _ = updateKey(m, "markdown")
		Expect(visiblePathList(m)).To(Equal([]string{"docs", "docs/readme.md"}))

		m, _ = updateKey(m, "ctrl+u")
		m, _ = updateKey(m, "untracked")
		Expect(visiblePathList(m)).To(Equal([]string{"ui", "ui/App_test.tsx"}))
	})

	It("esc clears the active filter and restores the full tree", func() {
		m := newTreeModel(files)
		allVisible := visiblePathList(m)

		m, _ = updateKey(m, "/")
		m, _ = updateKey(m, "a/x")
		Expect(visiblePathList(m)).To(Equal([]string{"a", "a/x.go"}))

		m, _ = updateKey(m, "esc")
		Expect(m.filtering).To(BeFalse())
		Expect(m.filterQuery).To(BeEmpty())
		Expect(visiblePathList(m)).To(Equal(allVisible))
	})
})

// --- Render output -------------------------------------------------------

var _ = Describe("treeModel View rendering", func() {
	files := []status.FileStatus{
		{Path: "cmd/gavel/main.go", State: status.StateStaged, Adds: 5, Dels: 1, FileMap: &repomap.FileMap{Language: "Go"}},
		{Path: "ui/App_test.tsx", State: status.StateUntracked, Adds: 7, FileMap: &repomap.FileMap{Language: "TypeScript", Scopes: []repomap.ScopeType{repomap.ScopeTypeTest}}},
		{Path: "docs/readme.md", State: status.StateUnstaged, Adds: 1, Dels: 1, FileMap: &repomap.FileMap{Language: "Markdown"}},
	}

	It("emits ANSI escapes for chips, checkboxes, and the cursor marker", func() {
		m := newTreeModel(files)
		m.height = 30
		// Cursor starts at the first dir; move to the Go file row (cmd/gavel/main.go)
		// and toggle so a green [x] is in the output.
		m, _ = updateKey(m, "down")
		m, _ = updateKey(m, "down")
		m, _ = updateKey(m, " ")

		out := m.View()
		Expect(out).To(ContainSubstring("\x1b["), "expected ANSI SGR escapes in tree picker output")
		Expect(out).To(ContainSubstring("▶ "), "cursor row must use the ▶ marker")

		// The ▶ cursor marker must be wrapped in an ANSI styling sequence,
		// not emitted as plain text.
		Expect(out).To(MatchRegexp(`\x1b\[[0-9;]*m▶ `),
			"cursor marker should be preceded by an ANSI SGR escape")

		plain := stripANSI(out)
		// Tree renders leaf names only (parents are implied by indent).
		Expect(plain).To(ContainSubstring("main.go"))
		Expect(plain).To(ContainSubstring("App_test.tsx"))
		Expect(plain).To(ContainSubstring("readme.md"))
		Expect(plain).To(ContainSubstring("cmd/"))
		Expect(plain).To(ContainSubstring("[x]"))
		Expect(plain).To(ContainSubstring("[ ]"))
		Expect(plain).To(ContainSubstring("+5"))
		Expect(plain).To(ContainSubstring("-1"))
		Expect(plain).To(ContainSubstring("staged"))
		Expect(plain).To(ContainSubstring("Go"))
	})

	It("renders the help line and header counts in plain semantics", func() {
		m := newTreeModel(files)
		m.height = 30
		plain := stripANSI(m.View())
		Expect(plain).To(ContainSubstring("Select files to commit"))
		Expect(plain).To(ContainSubstring("(0 / 3 selected)"))
		Expect(plain).To(ContainSubstring("/=filter"))
		Expect(plain).To(ContainSubstring("space=toggle"))
	})

	It("renders the active filter prompt and match count", func() {
		m := newTreeModel(files)
		m.height = 30
		m, _ = updateKey(m, "/")
		m, _ = updateKey(m, "test")

		plain := stripANSI(m.View())
		Expect(plain).To(ContainSubstring(`filter="test" (1 files)`))
		Expect(plain).To(ContainSubstring("filter: test"))
		Expect(plain).To(ContainSubstring("enter=keep"))
		Expect(plain).To(ContainSubstring("App_test.tsx"))
		Expect(plain).ToNot(ContainSubstring("main.go"))
	})

	It("renders a relative age chip for files with a known mtime", func() {
		aged := []status.FileStatus{{
			Path:       "cmd/gavel/main.go",
			State:      status.StateUnstaged,
			Adds:       2,
			ModifiedAt: timeMinus(3 * time.Hour),
		}}
		m := newTreeModel(aged)
		m.height = 20
		plain := stripANSI(m.View())
		Expect(plain).To(ContainSubstring("3h ago"),
			"chip row should include the file's relative mtime")
	})

	It("omits the age chip when ModifiedAt is unknown", func() {
		ghost := []status.FileStatus{{
			Path:     "cmd/gavel/main.go",
			State:    status.StateStaged,
			Adds:     2,
			WorkKind: status.KindDeleted,
		}}
		m := newTreeModel(ghost)
		m.height = 20
		plain := stripANSI(m.View())
		Expect(plain).ToNot(ContainSubstring(" ago"),
			"a deleted/unknown-mtime row must not render an age chip")
	})
})

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
		runTreePickerFunc = func([]status.FileStatus, string) (treePickerResult, error) {
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
		runTreePickerFunc = func(_ []status.FileStatus, gitRoot string) (treePickerResult, error) {
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
		runTreePickerFunc = func([]status.FileStatus, string) (treePickerResult, error) {
			return treePickerResult{}, nil
		}
		_, err := runInteractiveStaging(context.TODO(), opts)
		Expect(errors.Is(err, ErrInteractiveEmpty)).To(BeTrue())
	})

	It("propagates ErrInteractiveCancelled from the picker", func() {
		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{{Path: "x"}}}, nil
		}
		runTreePickerFunc = func([]status.FileStatus, string) (treePickerResult, error) {
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
		runTreePickerFunc = func(c []status.FileStatus, _ string) (treePickerResult, error) {
			candidatesSeen = c
			return treePickerResult{Selected: []string{"ok.go"}}, nil
		}
		_, err := runInteractiveStaging(context.TODO(), opts)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(candidatesSeen)).To(Equal(1))
		Expect(candidatesSeen[0].Path).To(Equal("ok.go"))
	})

	It("prints --summary output filtered to candidates before the picker", func() {
		opts.Summary = true
		gatherStatusFunc = func(string) (*status.Result, error) {
			return &status.Result{Files: []status.FileStatus{
				{Path: "kept.go", State: status.StateUnstaged},
				{Path: "ignored.md", State: status.StateConflict},
			}}, nil
		}
		runTreePickerFunc = func([]status.FileStatus, string) (treePickerResult, error) {
			return treePickerResult{Selected: []string{"kept.go"}}, nil
		}
		writer, drain := newCapturedStdout()
		previousOut := interactiveStdout
		interactiveStdout = writer
		defer func() { interactiveStdout = previousOut }()

		_, err := runInteractiveStaging(context.TODO(), opts)
		Expect(err).ToNot(HaveOccurred())
		out := drain()
		Expect(out).To(ContainSubstring("kept.go"))
		Expect(out).ToNot(ContainSubstring("ignored.md"))
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
		runTreePickerFunc = func([]status.FileStatus, string) (treePickerResult, error) {
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
		runTreePickerFunc = func([]status.FileStatus, string) (treePickerResult, error) {
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
		runTreePickerFunc = func([]status.FileStatus, string) (treePickerResult, error) {
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
	gatherStatusFunc = func(string) (*status.Result, error) { return &status.Result{}, nil }
	resetAllStagedFn = func(string) error { return nil }
	addFilesFunc = func(string, []string) error { return nil }
	gitRmCachedFunc = func(string, []string) error { return nil }
	runTreePickerFunc = func([]status.FileStatus, string) (treePickerResult, error) { return treePickerResult{}, nil }
	return func() {
		gatherStatusFunc = prevGather
		resetAllStagedFn = prevReset
		addFilesFunc = prevAdd
		gitRmCachedFunc = prevRm
		runTreePickerFunc = prevPicker
	}
}

// newCapturedStdout returns an *os.File that interactiveStdout can be set to,
// plus a drain() that closes the writer, waits for the reader goroutine, and
// returns the captured bytes. We need an *os.File because interactiveStdout is
// typed that way to match os.Stdout.
func newCapturedStdout() (*os.File, func() string) {
	r, w, err := os.Pipe()
	Expect(err).ToNot(HaveOccurred())
	buf := &bytes.Buffer{}
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = io.Copy(buf, r)
	}()
	drain := func() string {
		_ = w.Close()
		<-done
		_ = r.Close()
		return buf.String()
	}
	return w, drain
}
