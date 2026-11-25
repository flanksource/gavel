package runstatus_test

import (
	"context"
	"os"
	"os/exec"
	"sync"
	"time"

	clickyexec "github.com/flanksource/clicky/exec"
	clickytask "github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/runstatus"
	"github.com/flanksource/gavel/sdk/status"
	"github.com/flanksource/gavel/slots"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// tasks is a settable stand-in for clicky's global task groups.
type tasks struct {
	sync.Mutex
	snapshots []clickytask.TaskSnapshot
}

func (t *tasks) set(snapshots ...clickytask.TaskSnapshot) {
	t.Lock()
	defer t.Unlock()
	t.snapshots = snapshots
}

func (t *tasks) read() []clickytask.TaskSnapshot {
	t.Lock()
	defer t.Unlock()
	return append([]clickytask.TaskSnapshot(nil), t.snapshots...)
}

const lintGroup = "lint-run"

var started = time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)

func group(status string) clickytask.TaskSnapshot {
	return clickytask.TaskSnapshot{ID: lintGroup, GroupID: lintGroup, Type: "group", Name: testui.LintTaskGroupName, Status: status, StartedAt: started.Format(time.RFC3339Nano)}
}

func linter(name, status string) clickytask.TaskSnapshot {
	return clickytask.TaskSnapshot{ID: name, GroupID: lintGroup, Type: "task", Name: name, Status: status, StartedAt: started.Format(time.RFC3339Nano)}
}

// shortCache is a cache dir; ginkgo's are short enough for a unix socket under them.
func shortCache() string { return GinkgoT().TempDir() }

func published(cache string) func() (status.Status, error) {
	return func() (status.Status, error) { return status.Read(status.Path(cache, os.Getpid())) }
}

var _ = Describe("Reporter", func() {
	var cache string
	var live *tasks
	start := func(ctx context.Context, socketAfter time.Duration) *runstatus.Reporter {
		reporter, err := runstatus.Start(ctx, runstatus.Options{
			Cache: cache, WorkDir: "/src/acme", Command: "gavel lint", Interval: 10 * time.Millisecond, SocketAfter: socketAfter, Tasks: live.read,
		})
		Expect(err).NotTo(HaveOccurred())
		return reporter
	}
	BeforeEach(func() {
		cache, live = shortCache(), &tasks{}
	})

	It("publishes task group counts the way the live UI counts them, and the running tasks", func(ctx SpecContext) {
		live.set(group("running"), linter("golangci-lint", "success"), linter("eslint", "failed"), linter("ruff", "running"), linter("tsc", "pending"))
		reporter := start(ctx, time.Hour)
		DeferCleanup(reporter.Stop)
		reporter.SetTests([]parsers.Test{{Name: "pkg", Children: []parsers.Test{{Name: "a", Passed: true}, {Name: "b", Failed: true}, {Name: "c", Pending: true}}}})

		Eventually(published(cache)).Should(SatisfyAll(
			HaveField("PID", os.Getpid()), HaveField("Command", "gavel lint"), HaveField("WorkDir", "/src/acme"),
			HaveField("Tests", Equal(&status.Summary{Total: 3, Passed: 1, Failed: 1, Pending: 1})),
			HaveField("Groups", ConsistOf(status.Group{Name: testui.LintTaskGroupName, Kind: status.GroupLint, Started: started,
				Summary: status.Summary{Total: 4, Passed: 1, Failed: 1, Running: 1, Pending: 1}})),
			HaveField("Running", ConsistOf(status.Item{Name: "ruff", Group: testui.LintTaskGroupName, Started: started})),
		))
	})

	It("names the subprocess each running task runs", func(ctx SpecContext) {
		const pid = 4242
		ruff := linter("ruff", "running")
		ruff.Details = clickyexec.ExecTaskDetails{Command: "ruff", Args: []string{"check"}, PID: pid, Status: "running", Started: &started}
		live.set(group("running"), ruff, linter("tsc", "pending"))
		reporter := start(ctx, time.Hour)
		DeferCleanup(reporter.Stop)
		Eventually(published(cache)).Should(HaveField("Running", ConsistOf(SatisfyAll(HaveField("Name", "ruff"), HaveField("PID", pid)))))
	})

	It("rewrites the file only when progress changes", func(ctx SpecContext) {
		live.set(group("running"), linter("ruff", "running"))
		reporter := start(ctx, time.Hour)
		DeferCleanup(reporter.Stop)
		first, err := published(cache)()
		Expect(err).NotTo(HaveOccurred())
		Consistently(published(cache), 100*time.Millisecond).Should(HaveField("Updated", first.Updated))

		live.set(group("success"), linter("ruff", "success"))
		Eventually(published(cache)).Should(HaveField("Groups", ContainElement(HaveField("Summary.Passed", 1))))
	})

	It("lists the slots the process holds", func(ctx SpecContext) {
		lease, err := slots.Acquire(ctx, slots.Request{Limits: slots.Limits{Global: 2}, WorkDir: GinkgoT().TempDir(), CacheDir: cache, Command: "go test ./a"})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(lease.Close)
		reporter := start(ctx, time.Hour)
		DeferCleanup(reporter.Stop)
		Eventually(published(cache)).Should(HaveField("Slots", ConsistOf(SatisfyAll(
			HaveField("Pool", "global"), HaveField("State", status.LeaseHeld), HaveField("Limit", 2), HaveField("Command", "go test ./a"),
		))))
	})

	It("streams over its socket once the run has lasted, and removes both when it stops", func(ctx SpecContext) {
		live.set(group("running"), linter("ruff", "running"))
		reporter := start(ctx, 50*time.Millisecond)
		Eventually(published(cache)).Should(HaveField("Socket", status.SocketPath(cache, os.Getpid())))

		watcher, err := status.Watch(ctx, status.WatchOptions{Cache: cache, PID: os.Getpid(), Interval: time.Hour})
		Expect(err).NotTo(HaveOccurred())
		Eventually(watcher.Updates()).Should(Receive(HaveField("Socket", Not(BeEmpty()))))
		live.set(group("running"), linter("ruff", "success"), linter("tsc", "running"))
		Eventually(watcher.Updates()).Should(Receive(HaveField("Running", ConsistOf(HaveField("Name", "tsc")))))

		Expect(reporter.Stop()).To(Succeed())
		Expect(status.Path(cache, os.Getpid())).NotTo(BeAnExistingFile())
		Expect(status.SocketPath(cache, os.Getpid())).NotTo(BeAnExistingFile())
	})

	It("removes the status files of processes that exited", func(ctx SpecContext) {
		exited := exec.Command("true")
		Expect(exited.Run()).To(Succeed())
		Expect(os.MkdirAll(status.Dir(cache), 0o700)).To(Succeed())
		for _, path := range []string{status.Path(cache, exited.Process.Pid), status.SocketPath(cache, exited.Process.Pid)} {
			Expect(os.WriteFile(path, []byte("{}"), 0o600)).To(Succeed())
		}
		reporter := start(ctx, time.Hour)
		DeferCleanup(reporter.Stop)
		Expect(status.Path(cache, exited.Process.Pid)).NotTo(BeAnExistingFile())
		Expect(status.SocketPath(cache, exited.Process.Pid)).NotTo(BeAnExistingFile())
	})
})
