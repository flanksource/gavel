package github

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// conflictRepo is a throwaway git repository whose `main` and `feature`
// branches diverge exactly the way a stale PR does, so DetectMergeConflicts can
// replay a real merge without reaching GitHub. The remote URL is what
// conflictRemote matches on; nothing is ever fetched from it.
type conflictRepo struct {
	dir string
}

const conflictRepoSlug = "acme/widget"

func newConflictRepo(remoteURL string) *conflictRepo {
	repo := &conflictRepo{dir: GinkgoT().TempDir()}
	repo.git("init", "-b", "main")
	repo.git("config", "user.email", "gavel@example.com")
	repo.git("config", "user.name", "Gavel Test")
	repo.git("remote", "add", "origin", remoteURL)
	return repo
}

func (r *conflictRepo) git(args ...string) string {
	GinkgoHelper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	out, err := cmd.CombinedOutput()
	Expect(err).ToNot(HaveOccurred(), "git %v: %s", args, out)
	return string(out)
}

func (r *conflictRepo) write(path, content string) {
	GinkgoHelper()
	Expect(os.WriteFile(filepath.Join(r.dir, path), []byte(content), 0o600)).To(Succeed())
}

func (r *conflictRepo) commit(message string) {
	GinkgoHelper()
	r.git("add", "-A")
	r.git("commit", "-m", message)
}

func (r *conflictRepo) oid(rev string) string {
	GinkgoHelper()
	cmd := exec.Command("git", "rev-parse", rev)
	cmd.Dir = r.dir
	out, err := cmd.Output()
	Expect(err).ToNot(HaveOccurred())
	return string(out[:40])
}

// conflictingPR builds the PRInfo shape FetchPR produces for a PR GitHub has
// marked CONFLICTING, pointing at this repo's two branch tips.
func (r *conflictRepo) conflictingPR() *PRInfo {
	return &PRInfo{
		Number:      7,
		State:       "OPEN",
		Mergeable:   "CONFLICTING",
		MergeState:  "DIRTY",
		BaseRefName: "main",
		HeadRefName: "feature",
		BaseRefOID:  r.oid("main"),
		HeadRefOID:  r.oid("feature"),
	}
}

func (r *conflictRepo) options() Options {
	return Options{WorkDir: r.dir, Repo: conflictRepoSlug}
}

var _ = Describe("DetectMergeConflicts", func() {
	Describe("a PR whose branches really do conflict", func() {
		var report *MergeConflictReport

		BeforeEach(func() {
			repo := newConflictRepo("https://github.com/" + conflictRepoSlug + ".git")
			repo.write("go.mod", "require xz v0.5.12\n")
			repo.write("doc.md", "shared\n")
			repo.commit("base")

			repo.git("switch", "-c", "feature")
			repo.write("go.mod", "require xz v0.5.14\n")
			Expect(os.Remove(filepath.Join(repo.dir, "doc.md"))).To(Succeed())
			repo.commit("bump on the branch and drop the doc")

			repo.git("switch", "main")
			repo.write("go.mod", "require xz v0.5.20\n")
			repo.write("doc.md", "shared, and edited on main\n")
			repo.commit("bump on main")

			report = DetectMergeConflicts(repo.options(), repo.conflictingPR())
		})

		It("lists every conflicting path with the kind of conflict git reported", func() {
			Expect(report).ToNot(BeNil())
			Expect(report.Unavailable).To(BeEmpty())
			Expect(report.Files).To(ConsistOf(
				MergeConflict{Path: "go.mod", Kind: "content"},
				MergeConflict{Path: "doc.md", Kind: "modify/delete"},
			))
		})

		It("records the merge endpoints it replayed", func() {
			Expect(report.BaseRefName).To(Equal("main"))
			Expect(report.HeadRefName).To(Equal("feature"))
			Expect(report.MergeBase).ToNot(BeEmpty())
			Expect(report.MergeBase).ToNot(Equal(report.BaseOID))
		})

		It("offers the commands that resolve it in the user's checkout", func() {
			Expect(report.ResolveCommands()).To(Equal([]string{
				"git fetch origin main",
				"git switch feature",
				"git merge origin/main",
			}))
		})
	})

	It("says the verdict looks stale when the same two commits merge cleanly", func() {
		repo := newConflictRepo("https://github.com/" + conflictRepoSlug + ".git")
		repo.write("shared.md", "base\n")
		repo.commit("base")

		repo.git("switch", "-c", "feature")
		repo.write("branch-only.md", "added on the branch\n")
		repo.commit("branch adds its own file")

		repo.git("switch", "main")
		repo.write("main-only.md", "added on main\n")
		repo.commit("main adds its own file")

		report := DetectMergeConflicts(repo.options(), repo.conflictingPR())

		Expect(report).ToNot(BeNil())
		Expect(report.Files).To(BeEmpty())
		Expect(report.Unavailable).To(ContainSubstring("cleanly"))
	})

	It("refuses to guess from a checkout of a different repository", func() {
		repo := newConflictRepo("https://github.com/other/project.git")
		repo.write("go.mod", "base\n")
		repo.commit("base")
		repo.git("switch", "-c", "feature")
		repo.write("go.mod", "branch\n")
		repo.commit("branch")
		repo.git("switch", "main")
		repo.write("go.mod", "main\n")
		repo.commit("main")

		report := DetectMergeConflicts(repo.options(), repo.conflictingPR())

		Expect(report).ToNot(BeNil())
		Expect(report.Files).To(BeEmpty())
		Expect(report.Unavailable).To(ContainSubstring(conflictRepoSlug))
	})

	It("reports a missing endpoint rather than an empty conflict list", func() {
		// GitHub nulls baseRef once the base branch is deleted, leaving the
		// merge undefined rather than clean.
		pr := &PRInfo{
			Number: 7, State: "OPEN", Mergeable: "CONFLICTING",
			BaseRefName: "main", HeadRefName: "feature",
			HeadRefOID: "0123456789abcdef0123456789abcdef01234567",
		}

		report := DetectMergeConflicts(Options{Repo: conflictRepoSlug}, pr)

		Expect(report).ToNot(BeNil())
		Expect(report.Unavailable).To(ContainSubstring("merge endpoints"))
	})

	DescribeTable("returns nothing for a PR GitHub is not blocking on conflicts",
		func(pr *PRInfo) {
			Expect(DetectMergeConflicts(Options{Repo: conflictRepoSlug}, pr)).To(BeNil())
		},
		Entry("no PR at all", nil),
		Entry("cleanly mergeable", &PRInfo{State: "OPEN", Mergeable: "MERGEABLE"}),
		Entry("mergeability not yet computed", &PRInfo{State: "OPEN", Mergeable: "UNKNOWN"}),
		Entry("already merged, where GitHub leaves a stale verdict", &PRInfo{State: "MERGED", Mergeable: "CONFLICTING"}),
	)
})
