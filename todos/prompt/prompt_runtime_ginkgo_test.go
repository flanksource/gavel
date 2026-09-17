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

	// The dashboard's prompt editor is seeded from a preview that already leads
	// with the directive, so an edited prompt sent back as the override must not
	// gain a second one — nor keep a stale one after the effort changes.
	DescribeTable("leads an overridden prompt with exactly one directive for the resolved effort",
		func(effort api.Effort, override string) {
			const body = "## parser\n\nReview the parser"
			spec := api.Spec{
				Model:  api.Model{Name: "claude-sonnet-5", Mode: api.ModeAgent, Effort: effort},
				Prompt: api.Prompt{User: override + body},
			}
			request, _, err := Render([]*types.TODO{newTestTODO("other", "Unrelated todo")}, Options{Mode: types.ModePlan, Spec: spec})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(request.Prompt.User).To(gomega.Equal(EffortDirective(string(effort)) + "\n\n" + body))
		},
		Entry("no directive in the override", api.EffortXHigh, ""),
		Entry("the preview's directive kept in the override", api.EffortXHigh, EffortDirective("xhigh")+"\n\n"),
		Entry("directives stacked by earlier round-trips", api.EffortXHigh, EffortDirective("xhigh")+"\n\n"+EffortDirective("xhigh")+"\n\n"),
		Entry("a directive for an effort changed after editing", api.EffortXHigh, EffortDirective("high")+"\n\n"),
		Entry("the medium directive under a low effort", api.EffortLow, EffortDirective("medium")+"\n\n"),
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
