package fixtures

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("exec kind inference", func() {
	DescribeTable("infers exec only when kind is absent and exec is present",
		func(body, wantKind, wantError string) {
			spec, err := splitStep(json.RawMessage(body))
			if wantError != "" {
				Expect(err).To(MatchError(ContainSubstring(wantError)))
				return
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(spec.Kind).To(Equal(wantKind))
		},
		Entry("executable", `{"exec":"printf","args":["ready"]}`, "exec", ""),
		Entry("explicit kind", `{"kind":"probe","exec":"printf"}`, "probe", ""),
		Entry("empty explicit kind", `{"kind":"","exec":"printf"}`, "", "kind is required"),
		Entry("script alone", `{"script":"echo ready"}`, "", "kind is required"),
		Entry("no executable", `{"name":"missing"}`, "", "kind is required"),
	)
})
