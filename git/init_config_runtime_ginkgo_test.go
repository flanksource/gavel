package git

import (
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("repository configuration prompt runtime", func() {
	It("requires an editing agent while inheriting the selected model", func() {
		layers, err := (verify.PromptSpec{}).Layers(verify.PromptResolveOptions{
			Base:          api.Spec{Model: api.Model{Name: "configured-model", Mode: api.ModeAPI}},
			DefaultPrompt: initConfigPrompt,
			Data:          map[string]any{"workDir": "/repo", "configPath": "/repo/.gitanalyze.yaml"},
		})
		Expect(err).NotTo(HaveOccurred())
		resolved, err := api.ComposeSpecLayers(api.ResolveSpecOptions{Layers: layers})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Spec.Model).To(SatisfyAll(HaveField("Name", "configured-model"), HaveField("Mode", api.ModeAgent)))
	})
})
