package main

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/flanksource/clicky/prompt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("dashboard commit command", func() {
	It("inherits prompt mode from configuration without changing the CLI exit code", func() {
		repo := GinkgoT().TempDir()
		for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "developer@example.com"}, {"config", "user.name", "Developer"}, {"commit", "--allow-empty", "-qm", "initial"}} {
			cmd := exec.Command("git", args...)
			cmd.Dir = repo
			out, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(out))
		}
		Expect(os.WriteFile(filepath.Join(repo, ".gavel.yaml"), []byte("commit:\n  precommit:\n    mode: prompt\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repo, "change.txt"), []byte("pending change\n"), 0o600)).To(Succeed())
		lock := filepath.Join(repo, ".git", "index.lock")
		Expect(os.WriteFile(lock, nil, 0o600)).To(Succeed())
		aged := time.Now().Add(-time.Minute)
		Expect(os.Chtimes(lock, aged, aged)).To(Succeed())
		original := prompt.GlobalManager()
		manager := prompt.NewManager(prompt.NewMemory(prompt.MemoryConfig{}))
		prompt.SetDefault(manager)
		DeferCleanup(func() { prompt.SetDefault(original) })
		ctx, cancel := context.WithCancel(prompt.WithScope(GinkgoT().Context(), prompt.Scope{Owner: "dashboard-commit-test"}))
		DeferCleanup(cancel)
		before := exitCode
		finished := make(chan error, 1)
		go func() {
			finished <- runDashboardCommit(ctx, []string{"commit", "--work-dir", repo, "--message=fix: example", "--dry-run", "--tidy=false", "--lint=false", "--lint-secrets=false", "change.txt"}, io.Discard)
		}()
		var pending prompt.PromptSnapshot
		Eventually(func(g Gomega) {
			questions := manager.List(prompt.Filter{Owner: "dashboard-commit-test", State: "pending"})
			g.Expect(questions).To(HaveLen(1))
			pending = questions[0]
		}, 15*time.Second).Should(Succeed())
		Expect(manager.Resolve(pending.ID, prompt.Answer{Values: map[string]any{"choice": "0"}})).To(Succeed())
		var err error
		Eventually(finished, 5*time.Second).Should(Receive(&err))
		Expect(err).NotTo(HaveOccurred())
		Expect(lock).NotTo(BeAnExistingFile())
		Expect(exitCode).To(Equal(before))
	})
})
