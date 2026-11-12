package commit

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("index-patch commits", func() {
	var repo string

	BeforeEach(func() {
		repo = initCommitRepo(GinkgoT())
		GinkgoT().Setenv(testEnvVar, "1")
		writeFile(GinkgoT(), repo, "lines.txt", "a\nb\nc\nd\ne\n")
		writeFile(GinkgoT(), repo, "one.txt", "one\n")
		writeFile(GinkgoT(), repo, "two.txt", "two\n")
		gitRun(GinkgoT(), repo, "add", ".")
		gitRun(GinkgoT(), repo, "commit", "-qm", "files")
	})

	snapshot := func(id string, paths ...string) string {
		GinkgoHelper()
		patch, err := IndexPatchFile(repo, id)
		Expect(err).NotTo(HaveOccurred())
		Expect(SnapshotStagedPatch(context.Background(), SnapshotOptions{WorkDir: repo, Paths: paths, PatchFile: patch})).To(Succeed())
		return patch
	}
	staged := func() string { return gitOutput(GinkgoT(), repo, "diff", "--cached", "--name-only") }

	It("commits only the staged lines through a private index and leaves the rest unstaged", func() {
		writeFile(GinkgoT(), repo, "lines.txt", "a\nB\nc\nd\ne\n")
		gitRun(GinkgoT(), repo, "add", "lines.txt")
		writeFile(GinkgoT(), repo, "lines.txt", "a\nB\nc\nD\ne\n")
		writeFile(GinkgoT(), repo, "other.txt", "other\n")
		gitRun(GinkgoT(), repo, "add", "other.txt")

		patch := snapshot("lines", "lines.txt")
		Expect(staged()).To(Equal("other.txt\n"), "the snapshot leaves the real index")

		hookOutput := filepath.Join(repo, ".git", "hook-staged")
		_, err := Run(context.Background(), Options{WorkDir: repo, IndexPatch: patch, Config: verify.CommitConfig{
			Hooks: []verify.CommitHook{{Name: "list staged", Run: "git diff --cached --name-only > " + hookOutput}},
		}})
		Expect(err).NotTo(HaveOccurred())

		Expect(gitOutput(GinkgoT(), repo, "show", "HEAD:lines.txt")).To(Equal("a\nB\nc\nd\ne\n"))
		Expect(gitOutput(GinkgoT(), repo, "show", "--name-only", "--pretty=", "HEAD")).To(Equal("lines.txt\n"))
		Expect(os.ReadFile(hookOutput)).To(Equal([]byte("lines.txt\n")), "hooks see the private index")
		Expect(staged()).To(Equal("other.txt\n"), "the real index keeps its own staging and is in sync for committed paths")
		Expect(gitOutput(GinkgoT(), repo, "diff", "--", "lines.txt")).To(ContainSubstring("-d\n+D\n"))
		Expect(patch).NotTo(BeAnExistingFile())
		Expect(patch + ".index").NotTo(BeAnExistingFile())
	})

	It("lands two snapshots of disjoint files in order without reverting each other", func() {
		writeFile(GinkgoT(), repo, "one.txt", "one changed\n")
		writeFile(GinkgoT(), repo, "two.txt", "two changed\n")
		gitRun(GinkgoT(), repo, "add", "one.txt", "two.txt")
		first := snapshot("first", "one.txt")
		second := snapshot("second", "two.txt")
		Expect(staged()).To(BeEmpty())

		_, err := Run(context.Background(), Options{WorkDir: repo, IndexPatch: first})
		Expect(err).NotTo(HaveOccurred())
		_, err = Run(context.Background(), Options{WorkDir: repo, IndexPatch: second})
		Expect(err).NotTo(HaveOccurred())

		Expect(gitOutput(GinkgoT(), repo, "show", "--name-only", "--pretty=", "HEAD~1")).To(Equal("one.txt\n"))
		Expect(gitOutput(GinkgoT(), repo, "show", "--name-only", "--pretty=", "HEAD")).To(Equal("two.txt\n"))
		Expect(gitOutput(GinkgoT(), repo, "show", "HEAD:one.txt")).To(Equal("one changed\n"))
		Expect(gitOutput(GinkgoT(), repo, "status", "--porcelain")).To(BeEmpty())
	})

	It("keeps the patch and the real index when the commit fails, so a retry lands it", func() {
		writeFile(GinkgoT(), repo, "one.txt", "one changed\n")
		gitRun(GinkgoT(), repo, "add", "one.txt")
		patch := snapshot("retry", "one.txt")
		head := gitOutput(GinkgoT(), repo, "rev-parse", "HEAD")

		_, err := Run(context.Background(), Options{WorkDir: repo, IndexPatch: patch, Config: verify.CommitConfig{
			Hooks: []verify.CommitHook{{Name: "fail", Run: "exit 3"}},
		}})
		Expect(err).To(HaveOccurred())
		Expect(patch).To(BeAnExistingFile())
		Expect(gitOutput(GinkgoT(), repo, "rev-parse", "HEAD")).To(Equal(head))
		Expect(staged()).To(BeEmpty())

		_, err = Run(context.Background(), Options{WorkDir: repo, IndexPatch: patch})
		Expect(err).NotTo(HaveOccurred())
		Expect(gitOutput(GinkgoT(), repo, "show", "HEAD:one.txt")).To(Equal("one changed\n"))
		Expect(gitOutput(GinkgoT(), repo, "status", "--porcelain")).To(BeEmpty())
	})

	It("refuses to snapshot a path that is not staged and leaves the index alone", func() {
		writeFile(GinkgoT(), repo, "one.txt", "one changed\n")
		writeFile(GinkgoT(), repo, "two.txt", "two changed\n")
		gitRun(GinkgoT(), repo, "add", "one.txt")
		patch, err := IndexPatchFile(repo, "partial")
		Expect(err).NotTo(HaveOccurred())

		err = SnapshotStagedPatch(context.Background(), SnapshotOptions{WorkDir: repo, Paths: []string{"one.txt", "two.txt"}, PatchFile: patch})

		Expect(err).To(MatchError(ContainSubstring("two.txt")))
		Expect(staged()).To(Equal("one.txt\n"))
		Expect(patch).NotTo(BeAnExistingFile())
	})

	DescribeTable("rejects options that pick their own staging",
		func(opts Options, message string) {
			opts.WorkDir, opts.IndexPatch = repo, filepath.Join(repo, ".git", "missing.patch")
			_, err := Run(context.Background(), opts)
			Expect(err).To(MatchError(ErrIndexPatchExclusive))
			Expect(err.Error()).To(ContainSubstring(message))
		},
		Entry("file arguments", Options{Files: []string{"one.txt"}}, "file arguments"),
		Entry("interactive", Options{Interactive: true}, "--interactive"),
		Entry("fixup", Options{Fixup: FixupAuto}, "--fixup"),
		Entry("since", Options{Since: "HEAD~1"}, "--since"),
	)
})
