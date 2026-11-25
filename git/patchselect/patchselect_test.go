package patchselect_test

import (
	"strings"
	"testing"

	"github.com/flanksource/gavel/git/patchselect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPatchSelect(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "git patchselect")
}

func lines(text ...string) string { return strings.Join(text, "\n") + "\n" }

func old(n int) patchselect.Line   { return patchselect.Line{Side: patchselect.Old, Line: n} }
func added(n int) patchselect.Line { return patchselect.Line{Side: patchselect.New, Line: n} }

const modifiedHeader = "diff --git a/f.txt b/f.txt\nindex 1111111..2222222 100644\n--- a/f.txt\n+++ b/f.txt\n"

var modified = modifiedHeader + lines(
	"@@ -1,5 +1,6 @@",
	" a",
	"-b",
	"+B",
	" c",
	"-d",
	"+D",
	" e",
	"+f",
)

var multiHunk = modifiedHeader + lines(
	"@@ -1,4 +1,5 @@",
	" l1",
	"+x1",
	"+x2",
	" l2",
	"-l3",
	" l4",
	"@@ -10,3 +11,4 @@",
	" l10",
	"+y",
	" l11",
	" l12",
)

var noNewline = modifiedHeader + lines(
	"@@ -1,2 +1,3 @@",
	" a",
	"-b",
	`\ No newline at end of file`,
	"+b",
	"+c",
)

const newFileHeader = "diff --git a/n.txt b/n.txt\nnew file mode 100644\nindex 0000000..3333333\n--- /dev/null\n+++ b/n.txt\n"

var newFile = newFileHeader + lines(
	"@@ -0,0 +1,3 @@",
	"+one",
	"+two",
	"+three",
	`\ No newline at end of file`,
)

const deletedHeader = "diff --git a/d.txt b/d.txt\ndeleted file mode 100644\nindex 4444444..0000000\n--- a/d.txt\n+++ /dev/null\n"

var deleted = deletedHeader + lines(
	"@@ -1,3 +0,0 @@",
	"-one",
	"-two",
	"-three",
)

var _ = Describe("Filter", func() {
	DescribeTable("keeps exactly the selected changed lines",
		func(patch string, reverse bool, selected []patchselect.Line, expected string) {
			filtered, err := patchselect.Filter(patch, patchselect.Options{Lines: selected, Reverse: reverse})
			Expect(err).NotTo(HaveOccurred())
			Expect(filtered).To(Equal(expected))
		},
		Entry("forward: a replacement pair; the other removal becomes context and other additions drop",
			modified, false, []patchselect.Line{old(2), added(2)},
			modifiedHeader+lines("@@ -1,5 +1,5 @@", " a", "-b", "+B", " c", " d", " e")),
		Entry("forward: one addition only",
			modified, false, []patchselect.Line{added(6)},
			modifiedHeader+lines("@@ -1,5 +1,6 @@", " a", " b", " c", " d", " e", "+f")),
		Entry("reverse: unselected additions become context and unselected removals drop",
			modified, true, []patchselect.Line{old(2), added(2)},
			modifiedHeader+lines("@@ -1,6 +1,6 @@", " a", "-b", "+B", " c", " D", " e", " f")),
		Entry("forward multi-hunk: a hunk without selected lines drops and later hunks shift",
			multiHunk, false, []patchselect.Line{added(12)},
			modifiedHeader+lines("@@ -10,3 +10,4 @@", " l10", "+y", " l11", " l12")),
		Entry("forward multi-hunk: recomputed counts carry the offset into the next hunk",
			multiHunk, false, []patchselect.Line{added(2), old(3), added(12)},
			modifiedHeader+lines("@@ -1,4 +1,4 @@", " l1", "+x1", " l2", "-l3", " l4", "@@ -10,3 +10,4 @@", " l10", "+y", " l11", " l12")),
		Entry("reverse multi-hunk: the new side start stays, the old side shifts",
			multiHunk, true, []patchselect.Line{added(12)},
			modifiedHeader+lines("@@ -11,3 +11,4 @@", " l10", "+y", " l11", " l12")),
		Entry("no-newline marker stays with its kept line",
			noNewline, false, []patchselect.Line{old(2), added(2)},
			modifiedHeader+lines("@@ -1,2 +1,2 @@", " a", "-b", `\ No newline at end of file`, "+b")),
		Entry("new file: a dropped last line takes its no-newline marker with it",
			newFile, false, []patchselect.Line{added(1), added(2)},
			newFileHeader+lines("@@ -0,0 +1,2 @@", "+one", "+two")),
		Entry("new file: the marker follows the selected last line",
			newFile, false, []patchselect.Line{added(3)},
			newFileHeader+lines("@@ -0,0 +1,1 @@", "+three", `\ No newline at end of file`)),
		Entry("reverse new file partially unstaged is no longer a creation",
			newFile, true, []patchselect.Line{added(2)},
			"diff --git a/n.txt b/n.txt\n--- a/n.txt\n+++ b/n.txt\n"+lines("@@ -1,2 +1,3 @@", " one", "+two", " three", `\ No newline at end of file`)),
		Entry("deleted file partially staged is no longer a deletion",
			deleted, false, []patchselect.Line{old(1)},
			"diff --git a/d.txt b/d.txt\n--- a/d.txt\n+++ b/d.txt\n"+lines("@@ -1,3 +1,2 @@", "-one", " two", " three")),
		Entry("deleted file fully selected stays a deletion",
			deleted, false, []patchselect.Line{old(1), old(2), old(3)},
			deleted),
		Entry("reverse deleted file partially unstaged stays a deletion of the kept lines",
			deleted, true, []patchselect.Line{old(2)},
			deletedHeader+lines("@@ -1,1 +0,0 @@", "-two")),
	)

	DescribeTable("rejects selections it cannot honour",
		func(patch string, selected []patchselect.Line, expected error, message string) {
			_, err := patchselect.Filter(patch, patchselect.Options{Lines: selected})
			Expect(err).To(HaveOccurred())
			if expected != nil {
				Expect(err).To(MatchError(expected))
			}
			Expect(err.Error()).To(ContainSubstring(message))
		},
		Entry("empty selection", modified, nil, patchselect.ErrNoSelection, "no changed lines"),
		Entry("binary diff", "diff --git a/x.bin b/x.bin\nindex 1..2 100644\nBinary files a/x.bin and b/x.bin differ\n",
			[]patchselect.Line{added(1)}, patchselect.ErrBinary, "binary"),
		Entry("context line", modified, []patchselect.Line{old(1)}, nil, "old:1"),
		Entry("line outside the diff", modified, []patchselect.Line{added(99)}, nil, "new:99"),
		Entry("unknown side", modified, []patchselect.Line{{Side: "both", Line: 2}}, nil, `"both"`),
		Entry("empty patch", "", []patchselect.Line{added(1)}, nil, "no hunks"),
	)
})
