package github

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("BeginConflictRebase", func() {
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
		repo.write("branch-only.md", "added on the branch\n")
		repo.commit("branch adds its own file")
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

	It("replays the branch onto the base and stops on the conflicting commit", func() {
		files, err := BeginConflictRebase(repo.dir, report)

		Expect(err).ToNot(HaveOccurred())
		Expect(files).To(Equal([]MergeConflict{{Path: "go.mod", Kind: "content"}}))
		content, err := os.ReadFile(filepath.Join(repo.dir, "go.mod"))
		Expect(err).ToNot(HaveOccurred())
		Expect(string(content)).To(ContainSubstring("<<<<<<<"))
		Expect(strings.TrimSpace(repo.git("log", "--format=%s", report.BaseOID+"..HEAD"))).To(Equal("branch adds its own file"),
			"the clean commit before the conflict must already sit on top of the base")

		unresolved, inProgress, err := RebaseInProgress(repo.dir)
		Expect(err).ToNot(HaveOccurred())
		Expect(inProgress).To(BeTrue())
		Expect(unresolved).To(Equal([]string{"go.mod"}))
	})

	It("starts the rebase from a scratch branch built on the PR head", func() {
		repo.git("switch", "-c", "gavel/pr-7-fix")

		_, err := BeginConflictRebase(repo.dir, report)

		Expect(err).ToNot(HaveOccurred())
	})

	It("still reports a staged file with leftover conflict markers as unresolved", func() {
		_, err := BeginConflictRebase(repo.dir, report)
		Expect(err).ToNot(HaveOccurred())
		repo.git("add", "go.mod")

		unresolved, inProgress, err := RebaseInProgress(repo.dir)

		Expect(err).ToNot(HaveOccurred())
		Expect(inProgress).To(BeTrue())
		Expect(unresolved).To(Equal([]string{"go.mod"}))
	})

	It("reports a resolved stop as in progress with nothing unresolved", func() {
		_, err := BeginConflictRebase(repo.dir, report)
		Expect(err).ToNot(HaveOccurred())
		repo.write("go.mod", "require xz v0.5.20\n")
		repo.git("add", "go.mod")

		unresolved, inProgress, err := RebaseInProgress(repo.dir)

		Expect(err).ToNot(HaveOccurred())
		Expect(inProgress).To(BeTrue())
		Expect(unresolved).To(BeEmpty())
	})

	It("refuses to start while a rebase is already stopped", func() {
		_, err := BeginConflictRebase(repo.dir, report)
		Expect(err).ToNot(HaveOccurred())

		_, err = BeginConflictRebase(repo.dir, report)

		Expect(err).To(MatchError(ContainSubstring("already in progress")))
	})

	It("refuses a checkout with uncommitted tracked changes", func() {
		repo.write("shared.md", "local edit\n")

		_, err := BeginConflictRebase(repo.dir, report)

		Expect(err).To(MatchError(ContainSubstring("uncommitted")))
		Expect(err).To(MatchError(ContainSubstring("--worktree")))
	})

	It("refuses a checkout that does not contain the PR head", func() {
		repo.git("switch", "main")

		_, err := BeginConflictRebase(repo.dir, report)

		Expect(err).To(MatchError(ContainSubstring(shortOID(report.HeadOID))))
		Expect(err).To(MatchError(ContainSubstring("--worktree")))
	})

	It("refuses a report with no conflicted files to rebase over", func() {
		_, err := BeginConflictRebase(repo.dir, &MergeConflictReport{Unavailable: "stale verdict"})

		Expect(err).To(MatchError(ContainSubstring("stale verdict")))
	})

	It("puts the branch back when the replay turns out clean", func() {
		head := repo.oid("HEAD")
		stale := *report
		stale.BaseOID = repo.oid("main~1")

		_, err := BeginConflictRebase(repo.dir, &stale)

		Expect(err).To(MatchError(ContainSubstring("verdict looks stale")))
		Expect(repo.oid("HEAD")).To(Equal(head))
		_, inProgress, err := RebaseInProgress(repo.dir)
		Expect(err).ToNot(HaveOccurred())
		Expect(inProgress).To(BeFalse())
	})
})

var _ = Describe("RebaseInProgress", func() {
	It("reports no rebase on an ordinary checkout", func() {
		repo := newConflictRepo("https://github.com/" + conflictRepoSlug + ".git")
		repo.write("a.md", strings.Repeat("x\n", 3))
		repo.commit("base")

		unresolved, inProgress, err := RebaseInProgress(repo.dir)

		Expect(err).ToNot(HaveOccurred())
		Expect(inProgress).To(BeFalse())
		Expect(unresolved).To(BeEmpty())
	})
})
