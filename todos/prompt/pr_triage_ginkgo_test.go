package prompt

import (
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = Describe("triaging PR feedback titles", func() {
	DescribeTable("requires concrete titles while preserving accurate titles",
		func(name string) {
			req, _, err := renderResolvedForTest([]*types.TODO{newTestTODO("feedback", "Address feedback on acme/widgets#12")}, Options{Prompt: name, RawFixture: true})
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(req.Prompt.System).To(gomega.ContainSubstring("Address feedback on"))
			gomega.Expect(req.Prompt.System).To(gomega.ContainSubstring("concrete defect, change, or outcome"))
			gomega.Expect(req.Prompt.System).To(gomega.ContainSubstring("Never reword a title that is already accurate"))
		},
		Entry("existing TODO", "triage"),
		Entry("new TODO", "triage.new"),
	)
})
