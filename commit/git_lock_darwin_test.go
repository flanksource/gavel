package commit

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("macOS git lock watchers", func() {
	It("retries Git successfully while a kqueue watcher has the stale lock open", func() {
		repo := initCommitRepo(GinkgoT())
		lock := filepath.Join(repo, ".git", "index.lock")
		Expect(os.WriteFile(lock, nil, 0o644)).To(Succeed())
		old := time.Now().Add(-2 * gitStaleLockAge)
		Expect(os.Chtimes(lock, old, old)).To(Succeed())
		watcher, err := fsnotify.NewWatcher()
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(watcher.Close()).To(Succeed()) })
		Expect(watcher.Add(filepath.Dir(lock))).To(Succeed())
		out, err := exec.Command("lsof", "-t", "--", lock).Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Fields(string(out))).To(ContainElement(strconv.Itoa(os.Getpid())), "the watcher must actually have the lock file open")

		ctx := withGitLockPolicy(context.Background(), gitLockPolicy{Mode: CheckModePrompt, AssumeYes: true})
		_, err = runGitRetryingLocks(ctx, repo, "reset", "--mixed", "HEAD")
		Expect(err).NotTo(HaveOccurred())
		Expect(lock).NotTo(BeAnExistingFile())
	})
})
