package ui

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	cexec "github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/prompt"
	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("dashboard commit prompts", func() {
	DescribeTable("handles a stale lock through the running task", func(action string) {
		repo := createDiffRepository()
		writeDiffFile(repo, "change.txt", "pending change\n")
		lock := filepath.Join(repo, ".git", "index.lock")
		Expect(os.WriteFile(lock, nil, 0o600)).To(Succeed())
		aged := time.Now().Add(-time.Minute)
		Expect(os.Chtimes(lock, aged, aged)).To(Succeed())
		server := &Server{}
		finished := make(chan error, 1)
		server.SetProjectCommitRunner(func(ctx context.Context, _ []string, _ io.Writer) error {
			_, err := commit.Run(ctx, commit.Options{
				WorkDir: repo, Files: []string{"change.txt"}, Message: "fix: example", DryRun: true,
				LintFlag: "false", LintSecretsFlag: "false", TidyFlag: "false",
				Config: verify.CommitConfig{Precommit: verify.PrecommitConfig{Mode: "prompt"}},
			})
			finished <- err
			return err
		})
		group := clickytask.StartGroup[cexec.ExecResult]("Commit prompt test", clickytask.WithConcurrency(1))
		DeferCleanup(group.Cancel)
		task := server.runProjectCommit(group.Context(), repo, []string{"commit"}, group.Group)
		var pending prompt.PromptSnapshot
		Eventually(func(g Gomega) {
			prompts := ensurePromptManager().List(prompt.Filter{State: "pending", Labels: map[string]string{prompt.TaskLabel: task.ID()}})
			g.Expect(prompts).To(HaveLen(1))
			pending = prompts[0]
			g.Expect(pending.Title).To(ContainSubstring("Stale git lock"))
			g.Expect(clickytask.SnapshotByID(group.ID())).To(ContainElement(WithTransform(func(s clickytask.TaskSnapshot) int { return len(s.Prompts) }, Equal(1))))
		}, 15*time.Second).Should(Succeed())
		switch action {
		case "remove":
			Expect(ensurePromptManager().Resolve(pending.ID, prompt.Answer{Values: map[string]any{"choice": "0"}})).To(Succeed())
		case "decline":
			Expect(ensurePromptManager().Resolve(pending.ID, prompt.Answer{Values: map[string]any{"choice": "1"}})).To(Succeed())
		case "stop":
			Expect(clickytask.ControlTask(GinkgoT().Context(), group.ID(), task.ID(), clickytask.ControlStop)).To(Succeed())
		}
		var err error
		Eventually(finished).Should(Receive(&err))
		_, _ = task.GetResult()
		if action == "remove" {
			Expect(err).NotTo(HaveOccurred())
			Expect(lock).NotTo(BeAnExistingFile())
		} else {
			Expect(err).To(HaveOccurred())
			Expect(lock).To(BeAnExistingFile())
		}
		if action == "stop" {
			Expect(task.Status()).To(Equal(clickytask.StatusCancelled))
		}
		Expect(ensurePromptManager().List(prompt.Filter{State: "pending", Labels: map[string]string{prompt.TaskLabel: task.ID()}})).To(BeEmpty())
	}, Entry("remove and resume", "remove"), Entry("decline", "decline"), Entry("stop", "stop"))
})
