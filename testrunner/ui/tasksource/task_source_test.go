package tasksource_test

import (
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestTaskSource(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Run scoped tasks") }

var _ = Describe("run scoped virtual tasks", func() {
	It("resets replay and frozen state for a new run and signals stream completion", func() {
		srv := testui.NewServer()
		srv.LoadSnapshot(testui.Snapshot{})
		srv.BeginRun("next")
		srv.SetTaskSource(func() []task.TaskSnapshot {
			return []task.TaskSnapshot{{ID: "prepare", GroupID: "next", Type: "group", Name: "Prepare", Status: "running"}}
		})
		Expect(srv.Snapshot().Tests).To(HaveLen(1))
		updates := make(chan []parsers.Test)
		done := srv.StreamFrom(updates)
		close(updates)
		Eventually(done).Should(BeClosed())
		Expect(srv.Snapshot().Status.Running).To(BeFalse())
	})
	It("keeps Before and its dependency states alongside authored tests until replay", func() {
		srv := testui.NewServer()
		srv.BeginRun("fixtures")
		source := []task.TaskSnapshot{
			{ID: "Before", GroupID: "before-run-a", Type: "group", Name: "Before", Status: "running"},
			{ID: "db", GroupID: "before-run-a", Type: "task", Name: "Connect database", Status: "running", Description: "Opening write pool"},
			{ID: "schemas", GroupID: "before-run-a", Type: "task", Name: "Resolve schemas", Status: "pending", DependencyIDs: []string{"db"}},
		}
		srv.SetTaskSource(func() []task.TaskSnapshot { return source })
		before := srv.Snapshot().Tests[0]
		Expect(before.Name).To(Equal("Before"))
		Expect(before.Children[0].Running).To(BeTrue())
		Expect(before.Children[0].Message).To(Equal("Opening write pool"))
		Expect(before.Children[1].Message).To(Equal("Waiting for Connect database"))
		source[0].Status, source[1].Status, source[2].Status = "success", "success", "success"
		srv.SetResults([]parsers.Test{{Name: "authored step", Passed: true}})
		srv.MarkDone()
		finished := srv.Snapshot()
		source[0].Status = "running"
		Expect(srv.Snapshot().Tests).To(Equal(finished.Tests))
		Expect(finished.Status.Running).To(BeFalse())
		Expect(finished.Tests).To(HaveLen(2))
		Expect(finished.Tests[0].Passed).To(BeTrue())
		replay := testui.NewServer()
		replay.LoadSnapshot(finished)
		Expect(replay.Snapshot().Tests).To(Equal(finished.Tests))
	})
	It("isolates simultaneous runs even when group names are identical", func() {
		first, second := testui.NewServer(), testui.NewServer()
		first.BeginRun("first")
		second.BeginRun("second")
		first.SetTaskSource(func() []task.TaskSnapshot {
			return []task.TaskSnapshot{{ID: "Before", GroupID: "env-a:run-before", Type: "group", Name: "Before", Status: "running"}}
		})
		second.SetTaskSource(func() []task.TaskSnapshot { return nil })
		second.MarkDone()
		Expect(first.Snapshot().Status.Running).To(BeTrue())
		Expect(second.Snapshot().Status.Running).To(BeFalse())
		Expect(second.Snapshot().Tests).To(BeEmpty())
	})
})
