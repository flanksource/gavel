package tasksource_test

import (
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
})
