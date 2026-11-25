package commit

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky/prompt"
	"github.com/flanksource/clicky/task"
	commonscontext "github.com/flanksource/commons/context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Push rebase prompt", func() {
	It("appears on the running task and expires when the task is stopped", func() {
		t := GinkgoT()
		repo := initCommitRepo(t)
		fork := strings.TrimSpace(gitOutput(t, repo, "rev-parse", "HEAD"))
		const base = "rebase-base"
		const title = "Rebase onto origin/rebase-base conflicted. Retry with which strategy?"
		remote := filepath.Join(t.TempDir(), "remote.git")
		gitRun(t, repo, "init", "--bare", remote)
		gitRun(t, repo, "remote", "add", "origin", remote)
		writeFile(t, repo, "README.md", "# remote change\n")
		gitRun(t, repo, "commit", "-am", "Update remote")
		gitRun(t, repo, "push", "origin", "HEAD:refs/heads/"+base)
		gitRun(t, repo, "checkout", "-b", "topic", fork)
		writeFile(t, repo, "README.md", "# local change\n")
		gitRun(t, repo, "commit", "-am", "Update topic")

		previous := prompt.GlobalManager()
		manager := prompt.NewManager(prompt.NewMemory(prompt.MemoryConfig{}))
		prompt.SetDefault(manager)
		DeferCleanup(prompt.SetDefault, previous)
		group := task.StartGroup[struct{}]("Open PR", task.WithConcurrency(1))
		run := group.Add("Open PR", func(ctx commonscontext.Context, _ *task.Task) (struct{}, error) {
			deps, err := providerPushDeps(ctx, Options{WorkDir: repo, PushBranch: base})
			if err != nil {
				return struct{}{}, err
			}
			return struct{}{}, deps.rebaseOnto(repo, base)
		}, task.WithSlots(0), task.WithCancellationDrain())
		DeferCleanup(func() {
			run.Cancel()
			for _, pending := range manager.List(prompt.Filter{State: string(prompt.StatePending)}) {
				_ = manager.Resolve(pending.ID, prompt.Answer{Cancelled: true})
			}
			_, _ = run.GetResult()
		})

		var snapshot task.TaskSnapshot
		Eventually(func(g Gomega) {
			g.Expect(task.SnapshotByID(group.ID())).To(ContainElement(HaveField("ID", run.ID()), &snapshot))
			g.Expect(snapshot.Prompts).To(ConsistOf(HaveField("Title", title)))
		}, 5*time.Second).Should(Succeed())

		run.Cancel()
		_, err := run.GetResult()
		Expect(err).To(MatchError(context.Canceled))
		Expect(manager.List(prompt.Filter{Labels: map[string]string{prompt.TaskLabel: run.ID()}})).
			To(ConsistOf(And(HaveField("ID", snapshot.Prompts[0].ID), HaveField("State", string(prompt.StateExpired)))))
	})
})
