package lifecycle_test

import (
	"context"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/lifecycle"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Host defaults structural composition", func() {
	It("keeps a partial runtime visible until the request selects a supported model", func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		host := newHost(&fakeProvider{plan: todos.PlanState{Exists: true, Approved: true, Content: "# Plan"}})
		host.Config.Todos.Run.Spec = api.Spec{
			Model:       api.Model{Name: "gpt-5.6-sol", Mode: api.ModeAgent},
			Permissions: api.Permissions{Tools: api.ToolsFromLists([]string{"Read"}, nil)},
		}
		step := stepNamed(host.Def, "run")
		defaults, err := host.StepDefaults(context.Background(), step)
		Expect(err).NotTo(HaveOccurred())
		Expect(defaults.Spec.Name).To(Equal("gpt-5.6-sol"))
		Expect(defaults.Spec.Permissions.Tools).To(Equal(api.ToolsFromLists([]string{"Read"}, nil)))

		_, err = host.Resolve(context.Background(), hostTodo(), step, lifecycle.RunOptions{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("tool"))

		resolved, err := host.Resolve(context.Background(), hostTodo(), step, lifecycle.RunOptions{
			Request: api.Spec{Model: api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Spec.Name).To(Equal("claude-sonnet-5"))
		Expect(resolved.Spec.Permissions.Tools).To(Equal(defaults.Spec.Permissions.Tools))
	})
})
