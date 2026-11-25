package blame

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PR reviewer recommendations", func() {
	var dir, base, head, commitDate string
	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if commitDate != "" {
			cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+commitDate, "GIT_COMMITTER_DATE="+commitDate)
		}
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), "%s", out)
		return strings.TrimSpace(string(out))
	}
	write := func(path, content string) {
		Expect(os.MkdirAll(filepath.Dir(filepath.Join(dir, path)), 0755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, path), []byte(content), 0644)).To(Succeed())
	}
	commit := func(author string) string {
		git("add", "--", "pkg", ".github", "CODEOWNERS")
		git("-c", "user.name="+author, "-c", "user.email="+strings.ToLower(author)+"@example.com", "commit", "-qm", author)
		return git("rev-parse", "HEAD")
	}
	analyze := func() *Result {
		report, err := Analyze(context.Background(), Options{Dir: dir, Base: base, Head: head, HistoryLimit: 100})
		Expect(err).NotTo(HaveOccurred())
		return report
	}
	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		git("init", "-q")
		git("config", "commit.gpgsign", "false")
		write("pkg/source.txt", "one\ntwo\nthree\nfour\nfive\n")
		write(".github/CODEOWNERS", "* @general\npkg/ @acme/team\npkg/unowned.txt\n")
		write("CODEOWNERS", "* @wrong-location\n")
		commitDate = "2020-01-01T00:00:00Z"
		commit("Alice")
		write("pkg/source.txt", "one\nbob\nthree\nfour\nfive\n")
		commitDate = "2026-01-01T00:00:00Z"
		base = commit("Bob")
		write("pkg/source.txt", "one\nchanged\nthree\nfour\nfive\n")
		head = commit("Author")
	})

	It("ranks effective owners first and attributes changed lines and file history independently", func() {
		write("pkg/source.txt", "uncommitted changes must not affect blame\n")
		report := analyze()
		Expect(report.CodeownersFile).To(Equal(".github/CODEOWNERS"))
		Expect(report.Files).To(Equal([]string{"pkg/source.txt"}))
		Expect(report.Reviewers).To(Equal([]Reviewer{
			{Identity: Identity{Login: "@acme/team"}, CodeOwnerFiles: 1, Files: []Evidence{{Path: "pkg/source.txt", CodeOwner: true, Rule: "pkg/", RuleLine: 2}}},
			{Identity: Identity{Name: "Bob", Email: "bob@example.com"}, Lines: 1, Commits: 1, Files: []Evidence{{Path: "pkg/source.txt", Lines: 1, Commits: 1}}},
			{Identity: Identity{Name: "Alice", Email: "alice@example.com"}, Commits: 1, Files: []Evidence{{Path: "pkg/source.txt", Commits: 1}}},
		}))
	})

	It("keeps insertion context separate from changed-line blame", func() {
		git("reset", "--hard", base)
		write("pkg/source.txt", "one\nbob\ninserted\nthree\nfour\nfive\n")
		head = commit("Author")
		report := analyze()
		Expect(report.Reviewers[1].Lines).To(BeZero())
		Expect(report.Reviewers[1].AdjacentLines).To(Equal(1))
		Expect(report.Reviewers[2].AdjacentLines).To(Equal(1))
	})

	It("reads ownership at the base tip rather than the merge base or PR head", func() {
		git("reset", "--hard", base)
		write(".github/CODEOWNERS", "* @new-owner\n")
		base = commit("Maintainer")
		report := analyze()
		Expect(report.Reviewers[0].Login).To(Equal("@new-owner"))
		Expect(report.MergeBase).NotTo(Equal(base))
	})

	It("handles new, renamed, deleted and binary files, and ownerless overrides", func() {
		git("reset", "--hard", base)
		git("mv", "pkg/source.txt", "pkg/renamed file.txt")
		write("pkg/new.txt", "new\n")
		write("pkg/unowned.txt", "no owner\n")
		write("pkg/image.bin", "\x00\x01")
		head = commit("Author")
		report := analyze()
		Expect(report.Files).To(ConsistOf("pkg/renamed file.txt", "pkg/new.txt", "pkg/unowned.txt", "pkg/image.bin"))
		Expect(report.Reviewers[0].CodeOwnerFiles).To(Equal(3))
		Expect(report.Reviewers[1].Commits).To(Equal(1))
		git("reset", "--hard", base)
		git("rm", "pkg/source.txt")
		write("pkg/new.txt", "new\n")
		head = commit("Author")
		report = analyze()
		Expect(report.Reviewers[1].Lines + report.Reviewers[2].Lines).To(Equal(5))
	})

	It("excludes resolved GitHub identities matching the PR author", func() {
		report, err := Analyze(context.Background(), Options{Dir: dir, Base: base, Head: head, HistoryLimit: 100,
			Exclude: []string{"bob"}, ResolveAuthor: func(_ context.Context, _ string) (string, error) { return "bob", nil },
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Reviewers).To(HaveLen(1))
		Expect(report.Reviewers[0].Login).To(Equal("@acme/team"))
	})

	It("excludes bot contributions from both line blame and file history", func() {
		git("reset", "--hard", base)
		write("pkg/source.txt", "one\nbot\nthree\nfour\nfive\n")
		base = commit("Automation[bot]")
		write("pkg/source.txt", "one\nchanged\nthree\nfour\nfive\n")
		head = commit("Author")
		report := analyze()
		Expect(report.Reviewers).To(HaveLen(3))
		Expect(report.Reviewers[1].Lines + report.Reviewers[2].Lines).To(BeZero())
		for _, reviewer := range report.Reviewers {
			Expect(reviewer.Name).NotTo(ContainSubstring("[bot]"))
		}
	})

	It("removes email ownership aliases when the linked GitHub account is excluded", func() {
		git("reset", "--hard", base)
		write(".github/CODEOWNERS", "* bob@example.com\n")
		base = commit("Maintainer")
		write("pkg/source.txt", "one\nchanged\nthree\nfour\nfive\n")
		head = commit("Author")
		report, err := Analyze(context.Background(), Options{Dir: dir, Base: base, Head: head, HistoryLimit: 100,
			Exclude: []string{"bob"}, ResolveAuthor: func(_ context.Context, _ string) (string, error) { return "bob", nil },
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Reviewers).To(BeEmpty())
	})

	It("merges email owners and multiple author emails into a single GitHub reviewer", func() {
		git("reset", "--hard", base)
		write(".github/CODEOWNERS", "* alice@example.com @alice\n")
		base = commit("Maintainer")
		write("pkg/source.txt", "changed\nbob\nthree\nfour\nfive\n")
		head = commit("Author")
		report, err := Analyze(context.Background(), Options{Dir: dir, Base: base, Head: head, HistoryLimit: 100,
			ResolveAuthor: func(_ context.Context, _ string) (string, error) { return "alice", nil },
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Reviewers).To(HaveLen(1))
		Expect(report.Reviewers[0].Login).To(Equal("@alice"))
		Expect(report.Reviewers[0].CodeOwnerFiles).To(Equal(1))
		Expect(report.Reviewers[0].Commits).To(Equal(2))
		Expect(report.Reviewers[0].Lines).To(Equal(1))
	})

	It("limits history per file without dropping changed-line authors", func() {
		report, err := Analyze(context.Background(), Options{Dir: dir, Base: base, Head: head, HistoryLimit: 1})
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Reviewers).To(HaveLen(2))
		Expect(report.Reviewers[1].Name).To(Equal("Bob"))
		Expect(report.Reviewers[1].Commits).To(Equal(1))
	})

	It("filters file history and changed-line blame by an inclusive commit date while keeping owners", func() {
		git("reset", "--hard", base)
		write("pkg/source.txt", "changed\nchanged\nthree\nfour\nfive\n")
		head = commit("Author")
		cutoff := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
		report, err := Analyze(context.Background(), Options{Dir: dir, Base: base, Head: head, HistoryLimit: 100, Since: cutoff})
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Since).To(Equal(&cutoff))
		Expect(report.Reviewers).To(HaveLen(2))
		Expect(report.Reviewers[0].Login).To(Equal("@acme/team"))
		Expect(report.Reviewers[1].Name).To(Equal("Bob"))
		Expect(report.Reviewers[1].Lines).To(Equal(1))
		Expect(report.Reviewers[1].Commits).To(Equal(1))
		report, err = Analyze(context.Background(), Options{Dir: dir, Base: base, Head: head, HistoryLimit: 100, Since: cutoff.Add(time.Second)})
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Reviewers).To(HaveLen(1))
		Expect(report.Reviewers[0].Login).To(Equal("@acme/team"))
	})

	It("filters insertion context by date without removing current ADO reviewer policies", func() {
		git("reset", "--hard", base)
		write("pkg/source.txt", "one\nbob\ninserted\nthree\nfour\nfive\n")
		head = commit("Author")
		report, err := Analyze(context.Background(), Options{Dir: dir, Base: base, Head: head, HistoryLimit: 100,
			Since:           time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			PolicyReviewers: []PolicyReviewer{{Identity: Identity{Email: "alice@example.com"}, Policy: Policy{ID: 17, Name: "Required reviewers", Required: true}}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Reviewers).To(HaveLen(3))
		Expect(report.Reviewers[0].Email).To(Equal("alice@example.com"))
		Expect(report.Reviewers[0].Files).To(BeEmpty())
		Expect(report.Reviewers[2].Name).To(Equal("Bob"))
		Expect(report.Reviewers[2].AdjacentLines).To(Equal(1))
	})

	It("continues past older descendants and applies the history limit after the date filter", func() {
		git("reset", "--hard", base)
		commitDate = "2028-01-01T00:00:00Z"
		write("pkg/source.txt", "recent\nbob\nthree\nfour\nfive\n")
		commit("Recent")
		commitDate = "2020-01-01T00:00:00Z"
		write("pkg/source.txt", "recent\nold\nthree\nfour\nfive\n")
		base = commit("Old")
		write("pkg/source.txt", "recent\nold\nchanged\nfour\nfive\n")
		head = commit("Author")
		report, err := Analyze(context.Background(), Options{Dir: dir, Base: base, Head: head, HistoryLimit: 1,
			Since: time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)})
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Reviewers).To(HaveLen(2))
		Expect(report.Reviewers[1].Name).To(Equal("Recent"))
		Expect(report.Reviewers[1].Commits).To(Equal(1))
		Expect(report.Reviewers[1].Lines).To(BeZero())
	})

	It("returns an empty recommendation set for identical endpoints", func() {
		head = base
		report := analyze()
		Expect(report.Files).To(BeEmpty())
		Expect(report.Reviewers).To(BeEmpty())
	})

	It("uses docs ownership when neither higher priority location exists", func() {
		git("reset", "--hard", base)
		git("rm", ".github/CODEOWNERS", "CODEOWNERS")
		write("docs/CODEOWNERS", "* @docs-owner\n")
		git("add", "--", "docs/CODEOWNERS")
		git("-c", "user.name=Maintainer", "-c", "user.email=maintainer@example.com", "commit", "-qm", "ownership")
		base = git("rev-parse", "HEAD")
		write("pkg/source.txt", "one\nchanged\nthree\nfour\nfive\n")
		git("add", "--", "pkg/source.txt")
		git("-c", "user.name=Author", "-c", "user.email=author@example.com", "commit", "-qm", "change")
		head = git("rev-parse", "HEAD")
		report := analyze()
		Expect(report.CodeownersFile).To(Equal("docs/CODEOWNERS"))
		Expect(report.Reviewers[0].Login).To(Equal("@docs-owner"))
	})

	It("rejects invalid revisions and incomplete history instead of guessing", func() {
		_, err := Analyze(context.Background(), Options{Dir: dir, Base: "missing", Head: head, HistoryLimit: 100})
		Expect(err).To(HaveOccurred())
		git("rev-parse", "--git-dir")
		Expect(os.WriteFile(filepath.Join(dir, ".git", "shallow"), []byte(base+"\n"), 0644)).To(Succeed())
		_, err = Analyze(context.Background(), Options{Dir: dir, Base: base, Head: head, HistoryLimit: 100})
		Expect(err).To(MatchError(ContainSubstring("shallow")))
	})
})
