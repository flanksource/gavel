package prfix

import (
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPRPromptLayers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PR prompt layers")
}

var _ = Describe("lifecycle runtime precedence", func() {
	DescribeTable("applies the lifecycle spec after prompt defaults and before project overrides",
		func(override verify.PromptSpec, expected int) {
			layers, err := Layers(ResolveOptions{
				Reference: "pr.review", Dir: "/repo", PR: prContext(),
				Context: &api.Spec{Budget: api.Budget{MaxTurns: 7}}, Prompt: override,
			})
			Expect(err).NotTo(HaveOccurred())
			resolved, err := api.ResolveSpecLayers(api.ResolveSpecOptions{Layers: layers, RequireModel: true})
			Expect(err).NotTo(HaveOccurred())
			Expect(resolved.Spec.Budget.MaxTurns).To(Equal(expected))
		},
		Entry("lifecycle override", verify.PromptSpec{}, 7),
		Entry("project override", verify.PromptSpec{Spec: api.Spec{Budget: api.Budget{MaxTurns: 3}}}, 3),
	)
})
