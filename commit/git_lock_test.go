package commit

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("git lock retry", func() {
	var retries []int

	// stubLockRetryDelay records each retry ordinal, runs onRetry (standing in
	// for the other git process finishing) and skips the real backoff sleep.
	stubLockRetryDelay := func(onRetry func(retry int)) {
		retries = nil
		orig := gitLockRetryDelay
		DeferCleanup(func() { gitLockRetryDelay = orig })
		gitLockRetryDelay = func(retry int) time.Duration {
			retries = append(retries, retry)
			onRetry(retry)
			return 0
		}
	}

	holdIndexLock := func(repo string) string {
		lock := filepath.Join(repo, ".git", "index.lock")
		Expect(os.WriteFile(lock, nil, 0o644)).To(Succeed())
		return lock
	}

	releaseOnRetry := func(lock string, at int) func(int) {
		return func(retry int) {
			if retry == at {
				Expect(os.Remove(lock)).To(Succeed())
			}
		}
	}

	stagedRepo := func() string {
		repo := initCommitRepo(GinkgoT())
		writeFile(GinkgoT(), repo, "a.txt", "one\n")
		gitRun(GinkgoT(), repo, "add", "a.txt")
		return repo
	}

	DescribeTable("heldGitLock extracts the lock file git could not create",
		func(output, expectedLock string) {
			lock, held := heldGitLock(output)
			Expect(held).To(Equal(expectedLock != ""))
			Expect(lock).To(Equal(expectedLock))
		},
		Entry("index lock",
			"fatal: Unable to create '/repo/.git/index.lock': File exists.\n\nAnother git process seems to be running in this repository",
			"/repo/.git/index.lock"),
		Entry("HEAD ref lock",
			"fatal: cannot lock ref 'HEAD': Unable to create '/repo/.git/HEAD.lock': File exists.",
			"/repo/.git/HEAD.lock"),
		Entry("nested branch ref lock",
			"error: cannot lock ref 'refs/heads/feature/x': Unable to create '/repo/.git/refs/heads/feature/x.lock': File exists.",
			"/repo/.git/refs/heads/feature/x.lock"),
		Entry("linked worktree index lock",
			"fatal: Unable to create '/repo/.git/worktrees/wt/index.lock': File exists.",
			"/repo/.git/worktrees/wt/index.lock"),
		Entry("nothing to commit", "nothing to commit, working tree clean", ""),
		Entry("ref race without a held lock file", "error: cannot lock ref 'HEAD': is at abc but expected def", ""),
	)

	It("commits once another process releases the index lock", func() {
		repo := stagedRepo()
		stubLockRetryDelay(releaseOnRetry(holdIndexLock(repo), 2))

		hash, err := commitWithMessage(repo, "feat: add a")
		Expect(err).ToNot(HaveOccurred())
		Expect(retries).To(Equal([]int{1, 2}))
		Expect(hash).To(Equal(strings.TrimSpace(gitOutput(GinkgoT(), repo, "rev-parse", "HEAD"))))
		Expect(gitOutput(GinkgoT(), repo, "log", "-1", "--pretty=%s")).To(Equal("feat: add a\n"))
	})

	It("fails with ErrGitLockHeld naming the lock after exactly gitLockRetries retries", func() {
		repo := stagedRepo()
		head := gitOutput(GinkgoT(), repo, "rev-parse", "HEAD")
		lock := holdIndexLock(repo)
		stubLockRetryDelay(func(int) {})

		_, err := commitWithMessage(repo, "feat: add a")
		Expect(errors.Is(err, ErrGitLockHeld)).To(BeTrue(), "expected wrapped ErrGitLockHeld, got %v", err)
		Expect(err.Error()).To(ContainSubstring(filepath.Join(".git", "index.lock")))
		Expect(retries).To(Equal([]int{1, 2, 3}))
		Expect(gitOutput(GinkgoT(), repo, "rev-parse", "HEAD")).To(Equal(head), "no commit may land while the lock is held")
		Expect(lock).To(BeAnExistingFile(), "the lock may belong to a live process and must never be removed")
	})

	It("does not retry failures other than a held lock", func() {
		repo := initCommitRepo(GinkgoT())
		stubLockRetryDelay(func(int) {})

		_, err := commitWithMessage(repo, "feat: nothing staged")
		Expect(err).To(HaveOccurred())
		Expect(errors.Is(err, ErrGitLockHeld)).To(BeFalse())
		Expect(retries).To(BeEmpty())
	})

	It("retries the staging step of `gavel commit <files>` until the lock is released", func() {
		repo := initCommitRepo(GinkgoT())
		writeFile(GinkgoT(), repo, "a.txt", "one\n")
		GinkgoT().Setenv(testEnvVar, "1")
		stubLockRetryDelay(releaseOnRetry(holdIndexLock(repo), 1))

		result, err := Run(context.Background(), Options{WorkDir: repo, Files: []string{"a.txt"}})
		Expect(err).ToNot(HaveOccurred())
		Expect(retries).To(Equal([]int{1}))
		Expect(result.Staged).To(Equal([]string{"a.txt"}))
		Expect(gitOutput(GinkgoT(), repo, "show", "--name-only", "--pretty=", "HEAD")).To(Equal("a.txt\n"))
	})

	It("backs off exponentially with jitter below double each step", func() {
		const samples = 50
		for retry := 1; retry <= gitLockRetries; retry++ {
			step := gitLockRetryBase << (retry - 1)
			distinct := map[time.Duration]struct{}{}
			for range samples {
				delay := gitLockRetryDelay(retry)
				Expect(delay).To(And(BeNumerically(">=", step), BeNumerically("<", 2*step)), "retry %d", retry)
				distinct[delay] = struct{}{}
			}
			Expect(len(distinct)).To(BeNumerically(">", 1), "retry %d delays carry no jitter", retry)
		}
	})
})
