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
})
