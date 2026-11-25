package proposal_test

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/git/proposal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func text(s string) *string { return &s }

var _ = Describe("Proposal edits", func() {
	var dir, head string
	ctx := context.Background()

	BeforeEach(func() {
		dir, head = newRepo()
		write(dir, "run.sh", "echo hello\necho hello\n")
		Expect(os.Chmod(filepath.Join(dir, "run.sh"), 0o755)).To(Succeed())
		run(dir, "add", "run.sh")
		run(dir, "commit", "-qm", "script")
		head = run(dir, "rev-parse", "HEAD")
	})

	propose := func(edits ...proposal.Edit) (proposal.Result, error) {
		return proposal.Create(ctx, proposal.CreateOptions{Dir: dir, Edits: edits, Message: message, Ref: ref})
	}
	show := func(result proposal.Result, path string) string {
		GinkgoHelper()
		return run(dir, "show", result.Commit+":"+path)
	}

	DescribeTable("commits the edited files at the ref",
		func(edits []proposal.Edit, want map[string]string, files []proposal.FileStat) {
			result, err := propose(edits...)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Files).To(Equal(files))
			for path, content := range want {
				Expect(show(result, path)).To(Equal(content), path)
			}
			Expect(run(dir, "rev-parse", ref)).To(Equal(result.Commit))
		},
		Entry("an exact replacement matching once",
			[]proposal.Edit{{Path: "greeting.txt", OldString: "world", NewString: "gavel"}},
			map[string]string{"greeting.txt": "hello\ngavel"},
			[]proposal.FileStat{{Path: "greeting.txt", Additions: 1, Deletions: 1}}),
		Entry("replaceAll over every match",
			[]proposal.Edit{{Path: "run.sh", OldString: "hello", NewString: "bye", ReplaceAll: true}},
			map[string]string{"run.sh": "echo bye\necho bye"},
			[]proposal.FileStat{{Path: "run.sh", Additions: 2, Deletions: 2}}),
		Entry("a new file from content",
			[]proposal.Edit{{Path: "docs/new.md", Content: text("# New\n")}},
			map[string]string{"docs/new.md": "# New"},
			[]proposal.FileStat{{Path: "docs/new.md", Additions: 1}}),
		Entry("two edits to one file in order",
			[]proposal.Edit{{Path: "greeting.txt", OldString: "world", NewString: "there"}, {Path: "greeting.txt", OldString: "hello\nthere", NewString: "hi there"}},
			map[string]string{"greeting.txt": "hi there"},
			[]proposal.FileStat{{Path: "greeting.txt", Additions: 1, Deletions: 2}}),
	)

	It("deletes a file", func() {
		result, err := propose(proposal.Edit{Path: "notes.txt", Delete: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Files).To(Equal([]proposal.FileStat{{Path: "notes.txt", Deletions: 1}}))
		Expect(run(dir, "ls-tree", "--name-only", result.Commit)).To(Equal("greeting.txt\nrun.sh"))
	})

	It("keeps an executable file's mode and leaves the working tree and index untouched", func() {
		write(dir, "notes.txt", "one\ndirty\n")
		run(dir, "add", "notes.txt")
		statusBefore := run(dir, "status", "--porcelain")

		result, err := propose(proposal.Edit{Path: "run.sh", OldString: "echo hello\necho hello", NewString: "echo hi"})
		Expect(err).NotTo(HaveOccurred())

		Expect(run(dir, "ls-tree", result.Commit, "run.sh")).To(HavePrefix("100755 blob "))
		Expect(run(dir, "status", "--porcelain")).To(Equal(statusBefore))
		Expect(os.ReadFile(filepath.Join(dir, "run.sh"))).To(Equal([]byte("echo hello\necho hello\n")))
		Expect(run(dir, "rev-parse", "HEAD")).To(Equal(head))
	})

	DescribeTable("rejects edits it cannot apply exactly, creating no ref",
		func(edits []proposal.Edit, wantErr string) {
			_, err := propose(edits...)
			Expect(err).To(MatchError(ContainSubstring(wantErr)))
			Expect(run(dir, "for-each-ref", "refs/gavel")).To(BeEmpty())
		},
		Entry("ambiguous match", []proposal.Edit{{Path: "run.sh", OldString: "hello", NewString: "bye"}}, "run.sh: oldString matches 2 times"),
		Entry("no match", []proposal.Edit{{Path: "greeting.txt", OldString: "planet", NewString: "gavel"}}, "greeting.txt: oldString matches 0 times"),
		Entry("replacement of a missing file", []proposal.Edit{{Path: "missing.txt", OldString: "a", NewString: "b"}}, "missing.txt: no such file"),
		Entry("path traversal", []proposal.Edit{{Path: "../outside.txt", Content: text("x")}}, `invalid path "../outside.txt"`),
		Entry("absolute path", []proposal.Edit{{Path: "/etc/passwd", Content: text("x")}}, `invalid path "/etc/passwd"`),
		Entry("unclean path", []proposal.Edit{{Path: "docs/./a.md", Content: text("x")}}, `invalid path "docs/./a.md"`),
		Entry("git dir path", []proposal.Edit{{Path: ".git/config", Content: text("x")}}, `invalid path ".git/config"`),
		Entry("edit with two kinds", []proposal.Edit{{Path: "notes.txt", Delete: true, Content: text("x")}}, "exactly one of"),
		Entry("edits that change nothing", []proposal.Edit{{Path: "notes.txt", Content: text("one\n")}}, "proposed edits change nothing"),
	)

	It("refuses a patch and edits together", func() {
		_, err := proposal.Create(ctx, proposal.CreateOptions{Dir: dir, Patch: []byte(patch), Edits: []proposal.Edit{{Path: "notes.txt", Delete: true}}, Message: message, Ref: ref})
		Expect(err).To(MatchError(ContainSubstring("patch or edits, not both")))
	})
})
