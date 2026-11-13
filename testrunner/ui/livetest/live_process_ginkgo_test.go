package livetest

import (
	"testing"
	"time"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/task"
	flanksourceContext "github.com/flanksource/commons/context"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLiveSnapshot(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Live snapshot")
}

var _ = Describe("Live command projection", func() {
	It("includes the running PID and start time without a fabricated exit code", func() {
		process := exec.NewExec("sleep 30").WithProcessGroup()
		group := task.StartGroup[exec.ExecResult](testui.TestTaskGroupName)
		handle := group.Add("live command", func(_ flanksourceContext.Context, t *task.Task) (exec.ExecResult, error) {
			return *process.WithTask(t).Run().Result(), nil
		})
		DeferCleanup(func() { _ = process.KillTree(); handle.WaitFor() })
		Eventually(process.Pid, 5*time.Second).Should(BeNumerically(">", 0))
		tests := testui.NewServer().Snapshot().Tests
		Expect(tests).NotTo(BeEmpty())
		var found bool
		for _, root := range tests {
			for _, child := range root.Children {
				if child.TaskID != handle.ID() {
					continue
				}
				found = true
				Expect(child.Running).To(BeTrue())
				Expect(child.Pending).To(BeFalse())
				Expect(child.Attempts).To(HaveLen(1))
				Expect(child.Attempts[0].PID).To(Equal(process.Pid()))
				Expect(child.Attempts[0].Started.IsZero()).To(BeFalse())
				Expect(child.Attempts[0].ExitCode).To(BeNil())
			}
		}
		Expect(found).To(BeTrue())
	})
})
