package parsers

import (
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Go test batch metadata", func() {
	It("retains cached Ginkgo context when filtering results", func() {
		original := Test{Cached: true, Context: GinkgoContext{Labels: []string{"database"}, NumAttempts: 2}, Children: Tests{{Name: "spec", Passed: true}}}
		filtered := original.Filter(TestFilter{})
		Expect(filtered.Cached).To(BeTrue())
		Expect(filtered.Context).To(Equal(original.Context))
	})

	It("fails when a Ginkgo report is missing a frame", func() {
		lines := []goTestEvent{
			{Action: "output", Package: "example.com/suite", Output: "GAVEL_GINKGO_V1:0/2:e30=\n"},
			{Action: "pass", Package: "example.com/suite"},
		}
		var input strings.Builder
		for _, event := range lines {
			data, err := json.Marshal(event)
			Expect(err).NotTo(HaveOccurred())
			input.Write(data)
			input.WriteByte('\n')
		}
		_, err := ParseGoTestBatch(strings.NewReader(input.String()), "", map[string]bool{"example.com/suite": true})
		Expect(err).To(MatchError(ContainSubstring("incomplete Ginkgo metadata")))
	})
})
