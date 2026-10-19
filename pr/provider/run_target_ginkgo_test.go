package provider

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
)

var _ = Describe("PR source branch pushes", func() {
	It("pushes only the explicit source repository and rejects an advanced head", func() {
		ctx := context.Background()
		root := GinkgoT().TempDir()
		remote, work := filepath.Join(root, "source.git"), filepath.Join(root, "work")
		_, err := targetGit(ctx, root, "init", "--bare", remote)
		Expect(err).NotTo(HaveOccurred())
		_, err = targetGit(ctx, root, "clone", remote, work)
		Expect(err).NotTo(HaveOccurred())
		_, err = targetGit(ctx, work, "config", "user.name", "Fixture User")
		Expect(err).NotTo(HaveOccurred())
		_, err = targetGit(ctx, work, "config", "user.email", "fixture@example.com")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(work, "file"), []byte("first"), 0600)).To(Succeed())
		_, err = targetGit(ctx, work, "add", "file")
		Expect(err).NotTo(HaveOccurred())
		_, err = targetGit(ctx, work, "commit", "-m", "test: initial fixture")
		Expect(err).NotTo(HaveOccurred())
		_, err = targetGit(ctx, work, "push", "origin", "HEAD:refs/heads/topic")
		Expect(err).NotTo(HaveOccurred())
		first, err := Revision(ctx, work, "HEAD")
		Expect(err).NotTo(HaveOccurred())
		baseRemote := filepath.Join(root, "base.git")
		_, err = targetGit(ctx, root, "init", "--bare", baseRemote)
		Expect(err).NotTo(HaveOccurred())
		_, err = targetGit(ctx, work, "remote", "set-url", "origin", baseRemote)
		Expect(err).NotTo(HaveOccurred())
		target := &RunTarget{RepositoryURL: remote, Branch: "topic", HeadSHA: first}
		Expect(os.WriteFile(filepath.Join(work, "file"), []byte("second"), 0600)).To(Succeed())
		_, err = targetGit(ctx, work, "commit", "-am", "test: repair fixture")
		Expect(err).NotTo(HaveOccurred())
		pushed, err := target.PushHead(ctx, work)
		Expect(err).NotTo(HaveOccurred())
		remoteHead, err := targetGit(ctx, remote, "rev-parse", "refs/heads/topic")
		Expect(err).NotTo(HaveOccurred())
		Expect(pushed).To(Equal(remoteHead))
		Expect(pushed).NotTo(Equal(first))
		_, err = targetGit(ctx, baseRemote, "rev-parse", "--verify", "refs/heads/topic")
		Expect(err).To(HaveOccurred())
		target.HeadSHA = first
		_, err = target.PushHead(ctx, work)
		Expect(err).To(MatchError(ContainSubstring("advanced")))
	})

	It("rewrites the source branch only for a rebasing run", func() {
		ctx := context.Background()
		root := GinkgoT().TempDir()
		remote, work := filepath.Join(root, "source.git"), filepath.Join(root, "work")
		git := func(dir string, args ...string) string {
			GinkgoHelper()
			out, err := targetGit(ctx, dir, args...)
			Expect(err).NotTo(HaveOccurred())
			return out
		}
		commit := func(content, message string) {
			GinkgoHelper()
			Expect(os.WriteFile(filepath.Join(work, "file"), []byte(content), 0600)).To(Succeed())
			git(work, "commit", "-am", message)
		}
		git(root, "init", "--bare", remote)
		git(root, "clone", remote, work)
		git(work, "config", "user.name", "Fixture User")
		git(work, "config", "user.email", "fixture@example.com")
		Expect(os.WriteFile(filepath.Join(work, "file"), []byte("base"), 0600)).To(Succeed())
		git(work, "add", "file")
		git(work, "commit", "-m", "test: base")
		base := git(work, "rev-parse", "HEAD")
		commit("branch", "test: branch work")
		git(work, "push", "origin", "HEAD:refs/heads/topic")
		published := git(work, "rev-parse", "HEAD")

		// Replay the branch commit onto a new base: the new head no longer
		// descends from the published one.
		git(work, "reset", "--hard", base)
		commit("upstream", "test: base moved on")
		commit("branch on upstream", "test: branch work, rebased")
		rebased := git(work, "rev-parse", "HEAD")

		target := &RunTarget{RepositoryURL: remote, Branch: "topic", HeadSHA: published}
		_, err := target.PushHead(ctx, work)
		Expect(err).To(MatchError(ContainSubstring("fast-forward")))
		Expect(git(remote, "rev-parse", "refs/heads/topic")).To(Equal(published))

		target.Rebase = true
		pushed, err := target.PushHead(ctx, work)
		Expect(err).NotTo(HaveOccurred())
		Expect(pushed).To(Equal(rebased))
		Expect(git(remote, "rev-parse", "refs/heads/topic")).To(Equal(rebased))
		Expect(target.HeadSHA).To(Equal(rebased))
	})
})
