package prcreate

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ResolveBase", func() {
	const configuredBase = "origin/develop"
	var dir string

	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		dir = GinkgoT().TempDir()
	})

	writeConfig := func(base string) {
		Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("pr:\n  base: "+base+"\n"), 0o600)).To(Succeed())
	}

	It("prefers an explicit override over the configured base", func() {
		writeConfig(configuredBase)
		Expect(ResolveBase(dir, "release")).To(Equal("release"))
	})

	It("uses the configured pr.base when no override is given", func() {
		writeConfig(configuredBase)
		Expect(ResolveBase(dir, "")).To(Equal(configuredBase))
	})

	It("reports a directory that is not a repository", func() {
		_, err := ResolveBase(dir, "")
		Expect(err).To(HaveOccurred())
	})

	DescribeTable("uses the remote HEAD instead of guessing from local branch names", func(branch string) {
		f := newRepoFixture()
		git(f.repo, "branch", branch, f.baseSHA)
		git(f.repo, "update-ref", "refs/remotes/origin/"+branch, f.baseSHA)
		git(f.repo, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/"+branch)
		Expect(ResolveBase(f.repo, "")).To(Equal("origin/" + branch))
	},
		Entry("master even when main exists", "master"),
		Entry("a branch containing a slash", "release/stable"),
	)

	It("queries the remote's default branch when origin/HEAD is missing", func() {
		f := newRepoFixture()
		git(f.repo, "push", "origin", "main:master")
		git(f.bare, "symbolic-ref", "HEAD", "refs/heads/master")
		Expect(ResolveBase(f.repo, "")).To(Equal("origin/master"))
	})

	DescribeTable("uses an existing local base when there is no origin", func(branch string) {
		f := newRepoFixture()
		git(f.repo, "remote", "remove", "origin")
		if branch != "main" {
			git(f.repo, "branch", "-m", "main", branch)
		}
		Expect(ResolveBase(f.repo, "")).To(Equal(branch))
	},
		Entry("main", "main"),
		Entry("master", "master"),
	)

	It("reports a remote failure instead of guessing main", func() {
		f := newRepoFixture()
		git(f.repo, "remote", "set-url", "origin", filepath.Join(dir, "missing.git"))
		_, err := ResolveBase(f.repo, "")
		Expect(err).To(MatchError(ContainSubstring("default branch")))
	})

	It("reports a remote without an advertised default branch", func() {
		f := newRepoFixture()
		git(f.bare, "symbolic-ref", "HEAD", "refs/heads/missing")
		_, err := ResolveBase(f.repo, "")
		Expect(err).To(MatchError(ContainSubstring("default branch")))
	})

	It("fails on an unreadable configuration", func() {
		Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte("pr: [unterminated\n"), 0o600)).To(Succeed())
		_, err := ResolveBase(dir, "")
		Expect(err).To(MatchError(ContainSubstring("load PR configuration")))
	})
})
