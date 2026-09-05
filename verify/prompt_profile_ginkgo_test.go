package verify

import (
	"encoding/json"

	"github.com/flanksource/captain/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("prompt runtime profile configuration", func() {
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

	DescribeTable("rejects profile selection outside lifecycle resolution", func(config PromptSpec, source string) {
		_, err := config.Resolve(api.Spec{}, source, nil, "")
		Expect(err).To(MatchError(ContainSubstring("runtime profiles require TODO lifecycle resolution")))
	},
		Entry("configuration", PromptSpec{RuntimeProfile: "reviewer"}, "Review the change"),
		Entry("prompt pin", PromptSpec{}, "---\nruntimeProfile: reviewer\n---\nReview the change"),
	)
})
