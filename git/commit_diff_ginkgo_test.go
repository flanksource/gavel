package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gavelgit "github.com/flanksource/gavel/git"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// missingSHA is well-formed but names no object in any test repository.
const missingSHA = "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"

var _ = Describe("CommitDiff and CommitFiles", func() {
	var repo, setup, first, second string
	git := func(args ...string) string {
		GinkgoHelper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), "git %v: %s", args, out)
		return strings.TrimSpace(string(out))
	}
	commit := func(message string, files map[string]string) string {
		GinkgoHelper()
		for name, body := range files {
			Expect(os.MkdirAll(filepath.Join(repo, filepath.Dir(name)), 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(repo, name), []byte(body), 0o644)).To(Succeed())
			git("add", name)
		}
		git("commit", "-q", "-m", message)
		return git("rev-parse", "HEAD")
	}
	paths := func(files []gavelgit.CommitFile) []string {
		var out []string
		for _, f := range files {
			out = append(out, f.Path)
		}
		return out
	}

	BeforeEach(func() {
		repo = GinkgoT().TempDir()
		git("init", "-q")
		git("config", "user.email", "test@example.com")
		git("config", "user.name", "Test User")
		git("config", "commit.gpgsign", "false")
		// Force colour on so a missing --no-color would leak escapes into the diff.
		git("config", "color.ui", "always")
		commit("chore: base", map[string]string{"base.txt": "base\n"})
		setup = commit("chore(setup): snapshot", map[string]string{"wip.txt": "wip\n"})
		first = commit("feat: a", map[string]string{"pkg/a.go": "package pkg\n"})
		second = commit("docs: b", map[string]string{"docs/b.md": "# b\n"})
	})

	It("renders a single commit as a plain unified diff without colour or diffstat", func() {
		result, err := gavelgit.CommitDiff(repo, gavelgit.CommitDiffOptions{Head: second})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Diff).To(HavePrefix("diff --git a/docs/b.md b/docs/b.md\n"))
		Expect(result.Diff).To(ContainSubstring("+# b"))
		Expect(result.Diff).NotTo(ContainSubstring("\x1b["))
		Expect(result.Diff).NotTo(ContainSubstring("file changed"))
		Expect(result.Diff).NotTo(ContainSubstring("pkg/a.go"))
		Expect(result.Truncated).To(BeFalse())
		Expect(result.Binary).To(BeFalse())
	})

	It("diffs every commit in a setup..head range", func() {
		result, err := gavelgit.CommitDiff(repo, gavelgit.CommitDiffOptions{Base: setup, Head: second})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Diff).To(ContainSubstring("diff --git a/pkg/a.go b/pkg/a.go"))
		Expect(result.Diff).To(ContainSubstring("diff --git a/docs/b.md b/docs/b.md"))
		Expect(result.Diff).NotTo(ContainSubstring("wip.txt"))
		Expect(result.Diff).NotTo(ContainSubstring("\x1b["))

		files, err := gavelgit.CommitFiles(repo, gavelgit.CommitDiffOptions{Base: setup, Head: second})
		Expect(err).NotTo(HaveOccurred())
		Expect(paths(files)).To(ConsistOf("pkg/a.go", "docs/b.md"))
	})

	It("narrows a range and a single commit to a directory pathspec", func() {
		ranged, err := gavelgit.CommitDiff(repo, gavelgit.CommitDiffOptions{Base: setup, Head: second, File: "pkg"})
		Expect(err).NotTo(HaveOccurred())
		Expect(ranged.Diff).To(ContainSubstring("pkg/a.go"))
		Expect(ranged.Diff).NotTo(ContainSubstring("docs/b.md"))

		single, err := gavelgit.CommitDiff(repo, gavelgit.CommitDiffOptions{Head: first, File: "pkg/"})
		Expect(err).NotTo(HaveOccurred())
		Expect(single.Diff).To(ContainSubstring("+package pkg"))

		files, err := gavelgit.CommitFiles(repo, gavelgit.CommitDiffOptions{Base: setup, Head: second, File: "docs"})
		Expect(err).NotTo(HaveOccurred())
		Expect(paths(files)).To(ConsistOf("docs/b.md"))
	})

	It("flags a diff as binary only when every file in it is binary", func() {
		head := commit("feat: logo", map[string]string{"logo.png": "\x89PNG\x00\x01\x02", "notes.txt": "notes\n"})
		image, err := gavelgit.CommitDiff(repo, gavelgit.CommitDiffOptions{Head: head, File: "logo.png"})
		Expect(err).NotTo(HaveOccurred())
		Expect(image.Binary).To(BeTrue())

		whole, err := gavelgit.CommitDiff(repo, gavelgit.CommitDiffOptions{Head: head})
		Expect(err).NotTo(HaveOccurred())
		Expect(whole.Binary).To(BeFalse(), "a text file in the diff must stay viewable")
	})

	DescribeTable("rejects malformed input before git runs",
		func(opts func() gavelgit.CommitDiffOptions, want string) {
			_, err := gavelgit.CommitDiff(repo, opts())
			Expect(err).To(MatchError(ContainSubstring(want)))
			_, err = gavelgit.CommitFiles(repo, opts())
			Expect(err).To(MatchError(ContainSubstring(want)))
			Expect(opts().Validate()).To(MatchError(ContainSubstring(want)))
		},
		Entry("empty head", func() gavelgit.CommitDiffOptions { return gavelgit.CommitDiffOptions{} }, `invalid commit hash ""`),
		Entry("invalid head", func() gavelgit.CommitDiffOptions { return gavelgit.CommitDiffOptions{Head: "not-a-hash"} }, `invalid commit hash "not-a-hash"`),
		Entry("invalid base", func() gavelgit.CommitDiffOptions {
			return gavelgit.CommitDiffOptions{Base: "--all", Head: second}
		}, `invalid base commit hash "--all"`),
		Entry("option-shaped path", func() gavelgit.CommitDiffOptions {
			return gavelgit.CommitDiffOptions{Head: second, File: "--output=x"}
		}, "must not begin with"),
	)

	DescribeTable("reports a well-formed sha missing from the repository as ErrCommitNotFound",
		func(opts func() gavelgit.CommitDiffOptions) {
			_, err := gavelgit.CommitDiff(repo, opts())
			Expect(err).To(MatchError(gavelgit.ErrCommitNotFound))
			Expect(err).To(MatchError(ContainSubstring(missingSHA[:8])))
			_, err = gavelgit.CommitFiles(repo, opts())
			Expect(err).To(MatchError(gavelgit.ErrCommitNotFound))
		},
		Entry("missing head", func() gavelgit.CommitDiffOptions { return gavelgit.CommitDiffOptions{Head: missingSHA} }),
		Entry("missing abbreviated head", func() gavelgit.CommitDiffOptions { return gavelgit.CommitDiffOptions{Head: missingSHA[:8]} }),
		Entry("missing base", func() gavelgit.CommitDiffOptions {
			return gavelgit.CommitDiffOptions{Base: missingSHA, Head: second}
		}),
	)

	It("surfaces a directory that is not a repository as an ordinary error", func() {
		_, err := gavelgit.CommitDiff(GinkgoT().TempDir(), gavelgit.CommitDiffOptions{Head: second})
		Expect(err).To(HaveOccurred())
		Expect(err).NotTo(MatchError(gavelgit.ErrCommitNotFound))
	})
})

var _ = DescribeTable("IsBinaryDiff",
	func(diff string, want bool) {
		Expect(gavelgit.IsBinaryDiff(diff)).To(Equal(want))
	},
	Entry("empty diff", "", false),
	Entry("one binary file", "diff --git a/x.png b/x.png\nnew file mode 100644\nBinary files /dev/null and b/x.png differ\n", true),
	Entry("one binary patch", "diff --git a/x.png b/x.png\nindex 1..2 100644\nGIT binary patch\nliteral 3\n", true),
	Entry("one text file", "diff --git a/x.txt b/x.txt\n--- a/x.txt\n+++ b/x.txt\n@@ -1 +1 @@\n-a\n+b\n", false),
	Entry("binary beside text", "diff --git a/x.png b/x.png\nBinary files a/x.png and b/x.png differ\n"+
		"diff --git a/x.txt b/x.txt\n--- a/x.txt\n+++ b/x.txt\n@@ -1 +1 @@\n-a\n+b\n", false),
)
