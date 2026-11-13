package lifecycle

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/prompts/library"
	"github.com/flanksource/gavel/verify"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
)

var _ = g.Describe("captured checklist templates", func() {
	g.It("preserves the template through grader sanitization and layers runtime overrides", func() {
		dir := g.GinkgoT().TempDir()
		source := "---\nmodel: gpt-5\nbudget: {cost: 1}\n---\nReview {{{criteria}}}"
		o.Expect(os.WriteFile(filepath.Join(dir, "check.prompt"), []byte(source), 0o600)).To(o.Succeed())
		cfg := verify.GavelConfig{AI: api.Spec{Model: api.Model{Name: "base-model"}}, Todos: verify.TodosConfig{Verify: verify.PromptSpec{File: "check.prompt", Spec: api.Spec{Budget: api.Budget{Cost: 3}}}}}
		h := &Host{WorkDir: dir}
		layers, err := h.profileLayers(context.Background(), LayerInput{Config: cfg, Step: StepVerify})
		o.Expect(err).NotTo(o.HaveOccurred())
		composed, err := api.ComposeSpecLayers(api.ResolveSpecOptions{Layers: layers.Layers})
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(composed.Spec.Name).To(o.Equal("gpt-5"))
		o.Expect(composed.Spec.Budget.Cost).To(o.Equal(3.0))
		grader := fixtures.GraderSpec(composed.Spec)
		o.Expect(grader.Prompt.User).To(o.BeEmpty())
		o.Expect(grader.Prompt.Metadata[library.ChecklistTemplateKey]).To(o.Equal(source))
	})
	g.It("refuses a missing configured checklist template", func() {
		h := &Host{WorkDir: g.GinkgoT().TempDir()}
		_, err := h.profileLayers(context.Background(), LayerInput{Step: StepVerify, Config: verify.GavelConfig{Todos: verify.TodosConfig{Verify: verify.PromptSpec{File: "missing.prompt"}}}})
		o.Expect(err).To(o.HaveOccurred())
	})
})
