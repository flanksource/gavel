package lifecycle_test

import (
	"context"
	"encoding/json"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The result schema is built per step from its declared events, so a prompt
// can never report an event its step's outcomes do not read.
var _ = Describe("a resolved step's result schema", func() {
	DescribeTable("is {success, event, comment} with the step's events as the event enum",
		func(name string) {
			GinkgoT().Setenv("HOME", GinkgoT().TempDir())
			host := newHost(&fakeProvider{plan: todos.PlanState{Exists: true, Approved: true, Content: "# Plan"}})
			step := stepNamed(host.Def, name)

			resolved, err := host.Resolve(context.Background(), hostTodo(), step, lifecycle.RunOptions{
				Request: api.Spec{Model: api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent}},
			})

			Expect(err).NotTo(HaveOccurred())
			var schema struct {
				Required   []string `json:"required"`
				Properties map[string]struct {
					Enum []string `json:"enum"`
				} `json:"properties"`
			}
			Expect(json.Unmarshal(resolved.Spec.Prompt.SchemaJSON, &schema)).To(Succeed())
			Expect(schema.Required).To(ConsistOf("success", "event", "comment"))
			Expect(schema.Properties).To(HaveLen(3))
			Expect(schema.Properties["event"].Enum).To(Equal(step.Events))
		},
		Entry("run", "run"), Entry("plan", "plan"), Entry("triage", "triage"), Entry("triage.new", "triage.new"),
	)
})
