package tasksource_test

import (
	"encoding/json"
	"time"

	"github.com/flanksource/clicky/task"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("embedded task projection", func() {
	It("preserves active phases without contributing authored test counts", func() {
		tests := testui.TaskTests([]task.TaskSnapshot{
			{Type: "group", ID: "Before", GroupID: "before", Name: "Before", Status: "running"},
			{Type: "task", ID: "schema", GroupID: "before", Name: "Initialize runtime", Status: "running"},
		})
		Expect(tests).To(HaveLen(1))
		Expect(tests[0].Running).To(BeTrue())
		Expect(tests[0].Children[0].Name).To(Equal("Initialize runtime"))
		Expect(tests[0].Sum().Total).To(BeZero())
	})
	It("gives a running task its start time and a finished task its exact duration, never a live elapsed time", func() {
		started := time.Date(2026, 1, 2, 3, 4, 5, 6000, time.UTC)
		elapsed := 1500 * time.Millisecond
		tests := testui.TaskTests([]task.TaskSnapshot{
			{Type: "group", ID: "Before", GroupID: "before", Name: "Before", Status: "running", StartedAt: started.Format(time.RFC3339Nano)},
			{Type: "task", ID: "running", GroupID: "before", Name: "Running", Status: "running", StartedAt: started.Format(time.RFC3339Nano)},
			{Type: "task", ID: "done", GroupID: "before", Name: "Done", Status: "success",
				StartedAt: started.Format(time.RFC3339Nano), FinishedAt: started.Add(elapsed).Format(time.RFC3339Nano)},
			{Type: "task", ID: "queued", GroupID: "before", Name: "Queued", Status: "pending"},
		})
		running, done, queued := tests[0].Children[0], tests[0].Children[1], tests[0].Children[2]
		Expect(running.Started).To(HaveValue(BeTemporally("==", started)))
		Expect(running.Duration).To(BeZero())
		Expect(done.Duration).To(Equal(elapsed))
		Expect(queued.Started).To(BeNil())
		Expect(queued.Duration).To(BeZero())

		encoded, err := json.Marshal(running)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).To(ContainSubstring(`"started":"` + started.Format(time.RFC3339Nano) + `"`))
		Expect(string(encoded)).NotTo(ContainSubstring(`"duration"`))
		encoded, err = json.Marshal(queued)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).NotTo(ContainSubstring(`"started"`))
	})
})
