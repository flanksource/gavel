package commit

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sync/atomic"
	"time"

	clickytask "github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clickyai "github.com/flanksource/gavel/ai"
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

type candidateSummaryAgent struct {
	closed     atomic.Bool
	closeCalls atomic.Int32
}

func (a *candidateSummaryAgent) ExecutePrompt(context.Context, clickyai.PromptRequest) (*clickyai.PromptResponse, error) {
	return &clickyai.PromptResponse{StructuredData: json.RawMessage(`{"summary":"describe streamed chooser change"}`)}, nil
}

func (a *candidateSummaryAgent) ExecuteBatch(context.Context, []clickyai.PromptRequest) (map[string]*clickyai.PromptResponse, error) {
	return nil, errors.New("unexpected batch execution")
}

func (a *candidateSummaryAgent) GetCosts() clickyai.Costs { return nil }
func (a *candidateSummaryAgent) Close() error {
	a.closed.Store(true)
	a.closeCalls.Add(1)
	return nil
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

var _ = Describe("startCandidateSummaries", func() {
	It("streams summaries and restores renderer state when stopped", func() {
		workDir := initCommitRepoForGinkgo()
		defer os.RemoveAll(workDir)
		Expect(os.WriteFile(filepath.Join(workDir, "new.go"), []byte("package new\n"), 0o644)).To(Succeed())

		agent := &candidateSummaryAgent{}
		previousNewAgent := newAgentFunc
		newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) { return agent, nil }
		defer func() { newAgentFunc = previousNewAgent }()

		previousNoRender := clickytask.IsNoRender()
		defer clickytask.SetNoRender(previousNoRender)

		opts := promptTestOptions()
		opts.WorkDir = workDir
		session, err := startCandidateSummaries(context.Background(), opts, []status.FileStatus{{
			Path:  "new.go",
			State: status.StateUntracked,
		}})
		Expect(err).ToNot(HaveOccurred())
		Expect(session.Files).To(HaveLen(1))
		Expect(session.Files[0].AIStatus).To(Equal(status.AISummaryStatusPending))
		Expect(clickytask.IsNoRender()).To(BeTrue())

		var updates []status.AISummaryUpdate
		for update := range session.Updates {
			updates = append(updates, update)
		}
		session.Stop()
		session.Stop()

		Expect(updates).To(ContainElement(status.AISummaryUpdate{
			Index:   0,
			Status:  status.AISummaryStatusDone,
			Summary: "describe streamed chooser change",
		}))
		Expect(agent.closed.Load()).To(BeTrue())
		Expect(agent.closeCalls.Load()).To(Equal(int32(1)))
		Expect(clickytask.IsNoRender()).To(Equal(previousNoRender))
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

	It("streams AI summary states into the matching file row", func() {
		prepared := &status.Result{Files: []status.FileStatus{
			{Path: "cmd/gavel/main.go", State: status.StateUnstaged},
			{Path: "ui/App.tsx", State: status.StateUntracked},
		}}
		prepared.PrepareAISummaries()
		updates := make(chan status.AISummaryUpdate, 2)
		m := newTreeModel(prepared.Files)
		m.summaryUpdates = updates
		nodeAtPath(m, "cmd/gavel/main.go").Selected = true

		Expect(stripANSI(m.View())).To(ContainSubstring("⏳ ai"))
		updates <- status.AISummaryUpdate{Index: 0, Status: status.AISummaryStatusRunning}
		model, next := m.Update(m.Init()())
		m = model.(treeModel)
		Expect(stripANSI(m.View())).To(ContainSubstring("⟳ ai"))

		updates <- status.AISummaryUpdate{Index: 0, Status: status.AISummaryStatusDone, Summary: "stream chooser summaries"}
		model, next = m.Update(next())
		m = model.(treeModel)
		plain := stripANSI(m.View())
		Expect(plain).To(ContainSubstring("main.go  unstaged · stream chooser summaries"))
		Expect(nodeAtPath(m, "cmd/gavel/main.go").Selected).To(BeTrue())

		close(updates)
		model, next = m.Update(next())
		m = model.(treeModel)
		Expect(next).To(BeNil())
		Expect(m.summaryUpdates).To(BeNil())
	})

	It("renders per-file AI summary failures without blocking the picker", func() {
		prepared := &status.Result{Files: []status.FileStatus{{Path: "failed.go", State: status.StateUnstaged}}}
		prepared.PrepareAISummaries()
		m := newTreeModel(prepared.Files)

		model, cmd := m.Update(aiSummaryUpdateMsg{
			open: true,
			update: status.AISummaryUpdate{
				Index:  0,
				Status: status.AISummaryStatusFailed,
				Error:  "provider unavailable",
			},
		})
		m = model.(treeModel)
		Expect(cmd).To(BeNil())
		Expect(stripANSI(m.View())).To(ContainSubstring("⚠ ai summary failed"))

		m, _ = updateKey(m, " ")
		Expect(m.selectedPaths()).To(ConsistOf("failed.go"))
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
