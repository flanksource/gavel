package commit

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/status"
	"github.com/flanksource/repomap"
)

// gitIgnoreCall records one invocation of the stubbed appendGitIgnore writer.
type gitIgnoreCall struct {
	gitRoot string
	entries []string
}

// fakeIgnoreWriter returns a writer that records calls and returns the
// configured "written" subset (mimicking dedup-against-existing behavior).
func fakeIgnoreWriter(written []string) (gitIgnoreWriter, *[]gitIgnoreCall) {
	calls := &[]gitIgnoreCall{}
	w := func(gitRoot string, entries []string) ([]string, error) {
		*calls = append(*calls, gitIgnoreCall{gitRoot: gitRoot, entries: append([]string(nil), entries...)})
		return written, nil
	}
	return w, calls
}

// modelWith builds a treeModel pre-wired with the given files, gitRoot, and
// writer. Cursor stays at index 0.
func modelWith(files []status.FileStatus, w gitIgnoreWriter) treeModel {
	m := newTreeModel(files)
	m.gitRoot = "/repo"
	m.appendGitIgnore = w
	return m
}

// moveCursorTo finds the node at the given path and positions the cursor on
// its visible row. Fails the spec if the node isn't visible.
func moveCursorTo(m *treeModel, path string) {
	for i, v := range m.visible {
		if v.Path == path {
			m.cursor = i
			return
		}
	}
	Fail("path not visible in tree: " + path)
}

var _ = Describe("treeModel gitignore action", func() {
	Describe("ignore as file (i then f)", func() {
		It("writes the file path and removes it from the tree", func() {
			files := []status.FileStatus{
				{Path: "a/x.go", State: status.StateUntracked, FileMap: &repomap.FileMap{Language: "Go"}},
				{Path: "b/y.go", State: status.StateUntracked, FileMap: &repomap.FileMap{Language: "Go"}},
			}
			w, calls := fakeIgnoreWriter([]string{"a/x.go"})
			m := modelWith(files, w)
			moveCursorTo(&m, "a/x.go")

			m, _ = updateKey(m, "i")
			Expect(m.ignorePrompt).ToNot(BeNil())
			m, _ = updateKey(m, "f")

			Expect(m.ignorePrompt).To(BeNil())
			Expect(*calls).To(HaveLen(1))
			Expect((*calls)[0].entries).To(Equal([]string{"a/x.go"}))
			Expect((*calls)[0].gitRoot).To(Equal("/repo"))
			Expect(visiblePathList(m)).ToNot(ContainElement("a/x.go"))
			Expect(visiblePathList(m)).To(ContainElement("b/y.go"))
		})

		It("is disabled when the cursor is on a directory", func() {
			files := []status.FileStatus{{Path: "a/x.go", State: status.StateUntracked}}
			w, calls := fakeIgnoreWriter(nil)
			m := modelWith(files, w)
			moveCursorTo(&m, "a") // directory

			m, _ = updateKey(m, "i")
			m, _ = updateKey(m, "f")

			Expect(*calls).To(BeEmpty())
			Expect(visiblePathList(m)).To(ContainElement("a/x.go"))
		})
	})

	Describe("ignore as folder (i then d)", func() {
		It("writes 'a/' and removes every file under it", func() {
			files := []status.FileStatus{
				{Path: "a/x.go", State: status.StateUntracked},
				{Path: "a/sub/y.go", State: status.StateUntracked},
				{Path: "b/z.go", State: status.StateUntracked},
			}
			w, calls := fakeIgnoreWriter([]string{"a/"})
			m := modelWith(files, w)
			moveCursorTo(&m, "a/x.go")

			m, _ = updateKey(m, "i")
			m, _ = updateKey(m, "d")

			Expect((*calls)[0].entries).To(Equal([]string{"a/"}))
			Expect(visiblePathList(m)).ToNot(ContainElement("a/x.go"))
			Expect(visiblePathList(m)).ToNot(ContainElement("a/sub/y.go"))
			Expect(visiblePathList(m)).To(ContainElement("b/z.go"))
		})

		It("is disabled for top-level files (no enclosing folder)", func() {
			files := []status.FileStatus{{Path: "README.md", State: status.StateUntracked}}
			w, calls := fakeIgnoreWriter(nil)
			m := modelWith(files, w)
			moveCursorTo(&m, "README.md")

			m, _ = updateKey(m, "i")
			m, _ = updateKey(m, "d")

			Expect(*calls).To(BeEmpty())
			Expect(visiblePathList(m)).To(ContainElement("README.md"))
		})

		It("ignores the directory itself when the cursor is on a folder", func() {
			files := []status.FileStatus{
				{Path: "junk/a", State: status.StateUntracked},
				{Path: "junk/b", State: status.StateUntracked},
			}
			w, calls := fakeIgnoreWriter([]string{"junk/"})
			m := modelWith(files, w)
			moveCursorTo(&m, "junk")

			m, _ = updateKey(m, "i")
			m, _ = updateKey(m, "d")

			Expect((*calls)[0].entries).To(Equal([]string{"junk/"}))
			Expect(visiblePathList(m)).ToNot(ContainElement("junk"))
			Expect(visiblePathList(m)).ToNot(ContainElement("junk/a"))
		})
	})

	Describe("ignore as extension (i then e)", func() {
		It("writes '*.log' and removes every file with that extension", func() {
			files := []status.FileStatus{
				{Path: "a/foo.log", State: status.StateUntracked},
				{Path: "b/bar.log", State: status.StateUntracked},
				{Path: "c/keep.go", State: status.StateUntracked},
			}
			w, calls := fakeIgnoreWriter([]string{"*.log"})
			m := modelWith(files, w)
			moveCursorTo(&m, "a/foo.log")

			m, _ = updateKey(m, "i")
			m, _ = updateKey(m, "e")

			Expect((*calls)[0].entries).To(Equal([]string{"*.log"}))
			Expect(visiblePathList(m)).ToNot(ContainElement("a/foo.log"))
			Expect(visiblePathList(m)).ToNot(ContainElement("b/bar.log"))
			Expect(visiblePathList(m)).To(ContainElement("c/keep.go"))
		})

		It("is disabled for files with no extension", func() {
			files := []status.FileStatus{
				{Path: "Makefile", State: status.StateUntracked},
				{Path: "src/main.go", State: status.StateUntracked},
			}
			w, calls := fakeIgnoreWriter(nil)
			m := modelWith(files, w)
			moveCursorTo(&m, "Makefile")

			m, _ = updateKey(m, "i")
			m, _ = updateKey(m, "e")

			Expect(*calls).To(BeEmpty())
			Expect(visiblePathList(m)).To(ContainElement("Makefile"))
		})
	})

	Describe("rm --cached scheduling", func() {
		It("queues already-staged files for git rm --cached", func() {
			files := []status.FileStatus{
				{Path: "a/x.go", State: status.StateStaged},
				{Path: "b/y.go", State: status.StateUntracked},
			}
			w, _ := fakeIgnoreWriter([]string{"a/x.go"})
			m := modelWith(files, w)
			moveCursorTo(&m, "a/x.go")

			m, _ = updateKey(m, "i")
			m, _ = updateKey(m, "f")

			Expect(m.pendingRmCached).To(Equal([]string{"a/x.go"}))
		})

		It("does not queue purely-untracked files", func() {
			files := []status.FileStatus{{Path: "a/x.go", State: status.StateUntracked}}
			w, _ := fakeIgnoreWriter([]string{"a/x.go"})
			m := modelWith(files, w)
			moveCursorTo(&m, "a/x.go")

			m, _ = updateKey(m, "i")
			m, _ = updateKey(m, "f")

			Expect(m.pendingRmCached).To(BeEmpty())
		})

		It("queues unstaged-but-tracked files (state == unstaged)", func() {
			files := []status.FileStatus{{Path: "a/x.go", State: status.StateUnstaged}}
			w, _ := fakeIgnoreWriter([]string{"*.go"})
			m := modelWith(files, w)
			moveCursorTo(&m, "a/x.go")

			m, _ = updateKey(m, "i")
			m, _ = updateKey(m, "e")

			Expect(m.pendingRmCached).To(Equal([]string{"a/x.go"}))
		})
	})

	Describe("submenu cancel", func() {
		It("esc closes the submenu without writing or pruning", func() {
			files := []status.FileStatus{{Path: "a/x.go", State: status.StateUntracked}}
			w, calls := fakeIgnoreWriter(nil)
			m := modelWith(files, w)
			moveCursorTo(&m, "a/x.go")

			m, _ = updateKey(m, "i")
			Expect(m.ignorePrompt).ToNot(BeNil())
			m, _ = updateKey(m, "esc")

			Expect(m.ignorePrompt).To(BeNil())
			Expect(*calls).To(BeEmpty())
			Expect(visiblePathList(m)).To(ContainElement("a/x.go"))
			Expect(m.cancelled).To(BeFalse())
		})
	})

	Describe("status line feedback", func() {
		It("reports 'ignored' when new lines were written", func() {
			files := []status.FileStatus{{Path: "a/x.log", State: status.StateUntracked}}
			w, _ := fakeIgnoreWriter([]string{"*.log"})
			m := modelWith(files, w)
			moveCursorTo(&m, "a/x.log")

			m, _ = updateKey(m, "i")
			m, _ = updateKey(m, "e")

			Expect(stripANSI(m.statusLine)).To(ContainSubstring("ignored: *.log"))
		})

		It("reports 'already ignored' when the writer returned no new lines", func() {
			files := []status.FileStatus{{Path: "a/x.log", State: status.StateUntracked}}
			w, _ := fakeIgnoreWriter(nil) // entry already present → no new lines
			m := modelWith(files, w)
			moveCursorTo(&m, "a/x.log")

			m, _ = updateKey(m, "i")
			m, _ = updateKey(m, "e")

			Expect(stripANSI(m.statusLine)).To(ContainSubstring("already ignored: *.log"))
		})
	})

	Describe("rendering with submenu open", func() {
		It("replaces the help footer with the ignore options", func() {
			files := []status.FileStatus{{Path: "a/x.log", State: status.StateUntracked}}
			w, _ := fakeIgnoreWriter(nil)
			m := modelWith(files, w)
			m.height = 30
			moveCursorTo(&m, "a/x.log")

			m, _ = updateKey(m, "i")
			plain := stripANSI(m.View())

			Expect(plain).To(ContainSubstring("ignore:"))
			Expect(plain).To(ContainSubstring("f=file"))
			Expect(plain).To(ContainSubstring("d=folder"))
			Expect(plain).To(ContainSubstring("e=ext (*.log)"))
			Expect(plain).ToNot(ContainSubstring("space=toggle"))
		})
	})
})
