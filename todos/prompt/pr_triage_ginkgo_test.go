package prompt

import (
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = Describe("triaging PR feedback titles", func() {
	DescribeTable("requires concrete titles while preserving accurate titles",
		func(name string, envelope EnvelopeKind) {
			req, _, err := renderResolvedForTest([]*types.TODO{newTestTODO("feedback", "Address feedback on acme/widgets#12")}, Options{Prompt: name, Envelope: envelope})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(req.Prompt.System).To(gomega.ContainSubstring("Address feedback on"))
			gomega.Expect(req.Prompt.System).To(gomega.ContainSubstring("must replace"))
			gomega.Expect(req.Prompt.System).To(gomega.ContainSubstring("Never reword a title that is already accurate"))
		},
		Entry("existing TODO", "triage", EnvelopeTriage),
		Entry("new TODO", "triage.new", EnvelopeTriageNew),
	)
})
