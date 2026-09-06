package prompt

import (
	"encoding/json"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = Describe("resolved TODO prompt runtime", func() {
	It("rejects cleared generation text even when the runtime has verification commands", func() {
		spec := api.Spec{
			Model:    api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent, Effort: api.EffortHigh},
			Workflow: &api.Workflow{Verify: &api.Verify{Commands: []string{"true"}}},
		}.WithExplicit("/prompt/user")
		_, _, err := Render([]*types.TODO{newTestTODO("parser", "Review the parser")}, Options{Mode: types.ModeRun, Spec: spec})
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	DescribeTable("rejects an explicit empty conversation before adding an effort directive",
		func(prompt string) {
			var spec api.Spec
			gomega.Expect(json.Unmarshal([]byte(`{"model":"claude-sonnet-5","mode":"agent","prompt":`+prompt+`}`), &spec)).To(gomega.Succeed())
			_, _, err := Render([]*types.TODO{newTestTODO("parser", "Review the parser")}, Options{Mode: types.ModeRun, Spec: spec})
			gomega.Expect(err).To(gomega.HaveOccurred())
		},
		Entry("empty user", `{"user":""}`),
		Entry("empty prompt", `{}`),
	)

	It("renders conversation without reapplying template runtime fields", func() {
		spec := api.Spec{Model: api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent}}
		request, config, err := Render([]*types.TODO{newTestTODO("parser", "Review the parser")}, Options{
			Mode: types.ModeRun, Spec: spec,
			Template: "---\nbudget:\n  cost: 99\npermissions:\n  mode: bypassPermissions\nmemory:\n  skipUser: true\n---\nReview {{count}} item",
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(request.Budget).To(gomega.Equal(spec.Budget))
		gomega.Expect(request.Permissions).To(gomega.Equal(spec.Permissions))
		gomega.Expect(request.Memory).To(gomega.Equal(spec.Memory))
		gomega.Expect(request.Model).To(gomega.Equal(spec.Model))
		gomega.Expect(config.Model).To(gomega.Equal(spec.Model))
		gomega.Expect(config.Budget).To(gomega.Equal(spec.Budget))
		gomega.Expect(request.Prompt.User).To(gomega.ContainSubstring("Review 1 item"))
		gomega.Expect(request.Prompt.SchemaJSON).NotTo(gomega.BeEmpty())
	})
})
