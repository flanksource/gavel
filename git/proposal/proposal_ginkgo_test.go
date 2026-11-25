package proposal_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/git/proposal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	ref     = "refs/gavel/comments/thread-1/reply-1"
	message = "gavel: proposal for comment thread-1"
	patch   = `diff --git a/greeting.txt b/greeting.txt
--- a/greeting.txt
+++ b/greeting.txt
@@ -1,2 +1,2 @@
 hello
-world
+gavel
diff --git a/added.txt b/added.txt
new file mode 100644
--- /dev/null
+++ b/added.txt
@@ -0,0 +1,2 @@
+first
+second
`
)

func run(dir string, args ...string) string {
	GinkgoHelper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=test", "-c", "user.email=test@example.invalid"}, args...)...)
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %s: %s", strings.Join(args, " "), out)
	return strings.TrimSpace(string(out))
}

func write(dir, name, content string) {
	GinkgoHelper()
	Expect(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)).To(Succeed())
}

func newRepo() (string, string) {
	GinkgoHelper()
	dir := GinkgoT().TempDir()
	run(dir, "init", "-q")
	write(dir, "greeting.txt", "hello\nworld\n")
	write(dir, "notes.txt", "one\n")
	run(dir, "add", ".")
	run(dir, "commit", "-qm", "initial")
	return dir, run(dir, "rev-parse", "HEAD")
}

func create(dir string) proposal.Result {
	GinkgoHelper()
	result, err := proposal.Create(context.Background(), proposal.CreateOptions{Dir: dir, Patch: []byte(patch), Message: message, Ref: ref})
	Expect(err).NotTo(HaveOccurred())
	return result
}

var _ = Describe("Proposal", func() {
	var dir, head string
	ctx := context.Background()

	BeforeEach(func() {
		dir, head = newRepo()
	})

	It("commits the patch on HEAD at the ref without touching the working tree or index", func() {
		write(dir, "notes.txt", "one\ndirty\n")
		statusBefore := run(dir, "status", "--porcelain")

		result := create(dir)

		Expect(result).To(Equal(proposal.Result{Base: head, Commit: run(dir, "rev-parse", ref), Files: []proposal.FileStat{
			{Path: "added.txt", Additions: 2},
			{Path: "greeting.txt", Additions: 1, Deletions: 1},
		}}))
		Expect(run(dir, "rev-parse", result.Commit+"^")).To(Equal(head))
		Expect(run(dir, "log", "-1", "--format=%an <%ae> %s", result.Commit)).To(Equal("gavel <gavel@gavel.invalid> " + message))
		Expect(run(dir, "status", "--porcelain")).To(Equal(statusBefore))
		Expect(os.ReadFile(filepath.Join(dir, "notes.txt"))).To(Equal([]byte("one\ndirty\n")))
		Expect(run(dir, "rev-parse", "HEAD")).To(Equal(head))
	})

	DescribeTable("rejects invalid creates",
		func(opts proposal.CreateOptions, wantErr string) {
			opts.Dir = dir
			_, err := proposal.Create(ctx, opts)
			Expect(err).To(MatchError(ContainSubstring(wantErr)))
			Expect(run(dir, "for-each-ref", "refs/gavel")).To(BeEmpty())
		},
		Entry("empty patch", proposal.CreateOptions{Message: message, Ref: ref}, "empty patch"),
		Entry("patch that does not apply", proposal.CreateOptions{Patch: []byte(strings.ReplaceAll(patch, "-world", "-planet")), Message: message, Ref: ref}, "patch does not apply"),
		Entry("malformed ref", proposal.CreateOptions{Patch: []byte(patch), Message: message, Ref: "refs/gavel/bad..ref"}, "invalid ref"),
		Entry("unknown base", proposal.CreateOptions{Base: "no-such-branch", Patch: []byte(patch), Message: message, Ref: ref}, "no-such-branch"),
	)

	It("refuses to overwrite an existing ref", func() {
		create(dir)
		_, err := proposal.Create(ctx, proposal.CreateOptions{Dir: dir, Patch: []byte(patch), Message: message, Ref: ref})
		Expect(err).To(MatchError(ContainSubstring(ref)))
	})

	It("applies to the working tree only", func() {
		result := create(dir)
		Expect(proposal.Apply(ctx, dir, result.Base, result.Commit, false)).To(Succeed())
		Expect(run(dir, "diff", "--name-only")).To(Equal("greeting.txt"))
		Expect(run(dir, "status", "--porcelain", "added.txt")).To(Equal("?? added.txt"))
		Expect(run(dir, "diff", "--cached")).To(BeEmpty())
	})

	It("stages the patch when asked", func() {
		result := create(dir)
		Expect(proposal.Apply(ctx, dir, result.Base, result.Commit, true)).To(Succeed())
		Expect(run(dir, "diff", "--cached", "--name-only")).To(Equal("added.txt\ngreeting.txt"))
	})

	It("reports a conflicting apply as ErrConflict and leaves the working tree unchanged", func() {
		result := create(dir)
		write(dir, "greeting.txt", "hello\nplanet\n")
		statusBefore := run(dir, "status", "--porcelain")

		for _, stage := range []bool{false, true} {
			err := proposal.Apply(ctx, dir, result.Base, result.Commit, stage)
			Expect(err).To(MatchError(proposal.ErrConflict))
			Expect(err).To(MatchError(ContainSubstring("greeting.txt")))
			Expect(run(dir, "status", "--porcelain")).To(Equal(statusBefore))
			Expect(os.ReadFile(filepath.Join(dir, "greeting.txt"))).To(Equal([]byte("hello\nplanet\n")))
		}
	})

	It("deletes the ref", func() {
		create(dir)
		Expect(proposal.Delete(ctx, dir, ref)).To(Succeed())
		Expect(run(dir, "for-each-ref", "refs/gavel")).To(BeEmpty())
	})
})
