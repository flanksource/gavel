package verify

import (
	"encoding/json"

	"github.com/flanksource/captain/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("prompt runtime selection configuration", func() {
	It("preserves ordered preset selection beside the flat spec and file across JSON", func() {
		original := PromptSpec{
			Spec: api.Spec{Model: api.Model{Name: "cli:sonnet"}}, File: "review.prompt",
			Presets: []string{"organization", "reviewer"}, PresetsSet: true,
		}
		encoded, err := json.Marshal(original)
		Expect(err).NotTo(HaveOccurred())
		var restored PromptSpec
		Expect(json.Unmarshal(encoded, &restored)).To(Succeed())
		Expect(restored).To(Equal(original))
		Expect(string(encoded)).To(ContainSubstring(`"presets":["organization","reviewer"]`))
	})

	It("preserves an explicit empty preset selection", func() {
		original := PromptSpec{Presets: []string{}, PresetsSet: true}
		encoded, err := json.Marshal(original)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).To(ContainSubstring(`"presets":[]`))
		var restored PromptSpec
		Expect(json.Unmarshal(encoded, &restored)).To(Succeed())
		Expect(restored.PresetsSet).To(BeTrue())
		Expect(restored.Presets).To(BeEmpty())
	})

	It("preserves a preset pin when adopting an inline prompt document", func() {
		var spec PromptSpec
		encoded, err := json.Marshal("---\npresets: [organization, reviewer]\n---\nReview the change")
		Expect(err).NotTo(HaveOccurred())
		Expect(json.Unmarshal(encoded, &spec)).To(Succeed())
		Expect(spec.Presets).To(Equal([]string{"organization", "reviewer"}))
		Expect(spec.PresetsSet).To(BeTrue())
	})

	It("preserves profile selection beside the flat spec and file across JSON", func() {
		original := PromptSpec{Spec: api.Spec{Model: api.Model{Name: "cli:sonnet"}}, File: "review.prompt", RuntimeProfile: "reviewer"}
		encoded, err := json.Marshal(original)
		Expect(err).NotTo(HaveOccurred())
		var restored PromptSpec
		Expect(json.Unmarshal(encoded, &restored)).To(Succeed())
		Expect(restored).To(Equal(original))
		Expect(string(encoded)).To(ContainSubstring(`"runtimeProfile":"reviewer"`))
	})

	It("treats a profile-only configuration as nonempty and merges explicit selections", func() {
		base := PromptSpec{RuntimeProfile: "reviewer"}
		Expect(base.IsEmpty()).To(BeFalse())
		Expect(base.Merge(PromptSpec{}).RuntimeProfile).To(Equal("reviewer"))
		Expect(base.Merge(PromptSpec{RuntimeProfile: "implementer"}).RuntimeProfile).To(Equal("implementer"))
	})

	It("preserves a frontmatter pin when adopting an inline prompt document", func() {
		var spec PromptSpec
		encoded, err := json.Marshal("---\nruntimeProfile: reviewer\n---\nReview the change")
		Expect(err).NotTo(HaveOccurred())
		Expect(json.Unmarshal(encoded, &spec)).To(Succeed())
		Expect(spec.RuntimeProfile).To(Equal("reviewer"))
	})

	DescribeTable("rejects preset selection outside lifecycle resolution", func(config PromptSpec, source string) {
		_, err := config.Resolve(PromptResolveOptions{DefaultPrompt: source})
		Expect(err).To(MatchError(ContainSubstring("runtime presets require TODO lifecycle resolution")))
	},
		Entry("configuration", PromptSpec{Presets: []string{"reviewer"}, PresetsSet: true}, "Review the change"),
		Entry("prompt pin", PromptSpec{}, "---\npresets: [reviewer]\n---\nReview the change"),
	)

	It("warns and ignores a deprecated profile outside lifecycle resolution", func() {
		resolved, err := (PromptSpec{RuntimeProfile: "reviewer"}).Resolve(PromptResolveOptions{DefaultPrompt: "Review the change"})
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.Warnings).To(ContainElement(api.RuntimeProfileDeprecationWarning))
	})
})
