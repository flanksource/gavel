package github

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BeginConflictMerge", func() {
	var (
		repo   *conflictRepo
		report *MergeConflictReport
	)

	BeforeEach(func() {
		repo = newConflictRepo("https://github.com/" + conflictRepoSlug + ".git")
		repo.write("go.mod", "require xz v0.5.12\n")
		repo.write("shared.md", "base\n")
		repo.commit("base")

		repo.git("switch", "-c", "feature")
		repo.write("go.mod", "require xz v0.5.14\n")
		repo.commit("bump on the branch")

		repo.git("switch", "main")
		repo.write("go.mod", "require xz v0.5.20\n")
		repo.write("main-only.md", "added on main\n")
		repo.commit("bump on main")

		report = DetectMergeConflicts(repo.options(), repo.conflictingPR())
		Expect(report.Files).To(ConsistOf(MergeConflict{Path: "go.mod", Kind: "content"}))
		repo.git("switch", "feature")
	})

	It("starts the base merge and leaves the conflicted paths for the agent", func() {
		files, err := BeginConflictMerge(repo.dir, report)

		Expect(err).ToNot(HaveOccurred())
		Expect(files).To(Equal([]MergeConflict{{Path: "go.mod", Kind: "content"}}))
		content, err := os.ReadFile(filepath.Join(repo.dir, "go.mod"))
		Expect(err).ToNot(HaveOccurred())
		Expect(string(content)).To(ContainSubstring("<<<<<<<"))
		Expect(repo.git("diff", "--cached", "--name-only")).To(ContainSubstring("main-only.md"),
			"the cleanly merged side of the base must already be staged")

		unresolved, inProgress, err := MergeInProgress(repo.dir)
		Expect(err).ToNot(HaveOccurred())
		Expect(inProgress).To(BeTrue())
		Expect(unresolved).To(Equal([]string{"go.mod"}))
	})

	It("starts the merge from a scratch branch built on the PR head", func() {
		repo.git("switch", "-c", "gavel/pr-7-fix")

		_, err := BeginConflictMerge(repo.dir, report)

		Expect(err).ToNot(HaveOccurred())
	})

	It("still reports a staged file with leftover conflict markers as unresolved", func() {
		_, err := BeginConflictMerge(repo.dir, report)
		Expect(err).ToNot(HaveOccurred())
		repo.git("add", "go.mod")

		unresolved, inProgress, err := MergeInProgress(repo.dir)

		Expect(err).ToNot(HaveOccurred())
		Expect(inProgress).To(BeTrue())
		Expect(unresolved).To(Equal([]string{"go.mod"}))
	})

	It("reports a resolved merge as in progress with nothing unresolved", func() {
		_, err := BeginConflictMerge(repo.dir, report)
		Expect(err).ToNot(HaveOccurred())
		repo.write("go.mod", "require xz v0.5.20\n")
		repo.git("add", "go.mod")

		unresolved, inProgress, err := MergeInProgress(repo.dir)

		Expect(err).ToNot(HaveOccurred())
		Expect(inProgress).To(BeTrue())
		Expect(unresolved).To(BeEmpty())
	})

	It("refuses a checkout with uncommitted tracked changes", func() {
		repo.write("shared.md", "local edit\n")

		_, err := BeginConflictMerge(repo.dir, report)

		Expect(err).To(MatchError(ContainSubstring("uncommitted")))
		Expect(err).To(MatchError(ContainSubstring("--worktree")))
	})

	It("refuses a checkout that does not contain the PR head", func() {
		repo.git("switch", "main")

		_, err := BeginConflictMerge(repo.dir, report)

		Expect(err).To(MatchError(ContainSubstring(shortOID(report.HeadOID))))
		Expect(err).To(MatchError(ContainSubstring("--worktree")))
	})

	It("refuses a report with no conflicted files to merge", func() {
		_, err := BeginConflictMerge(repo.dir, &MergeConflictReport{Unavailable: "stale verdict"})

		Expect(err).To(MatchError(ContainSubstring("stale verdict")))
	})
})

var _ = Describe("MergeInProgress", func() {
	It("reports no merge on an ordinary checkout", func() {
		repo := newConflictRepo("https://github.com/" + conflictRepoSlug + ".git")
		repo.write("a.md", strings.Repeat("x\n", 3))
		repo.commit("base")

		unresolved, inProgress, err := MergeInProgress(repo.dir)

		Expect(err).ToNot(HaveOccurred())
		Expect(inProgress).To(BeFalse())
		Expect(unresolved).To(BeEmpty())
	})
})
