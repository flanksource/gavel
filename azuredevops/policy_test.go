package azuredevops

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("policy response decoding", func() {
	It("preserves explicitly disabled optional policy configuration", func() {
		var policy policyEvaluation
		err := json.Unmarshal([]byte(`{"status":"approved","configuration":{"id":3,"isEnabled":false,"isBlocking":false,"type":{"id":"type-id"}}}`), &policy)
		Expect(err).NotTo(HaveOccurred())
		Expect(policy.Status).To(Equal("approved"))
		Expect(policy.Configuration.ID).To(Equal(3))
		Expect(policy.Configuration.Enabled).To(BeFalse())
		Expect(policy.Configuration.Blocking).To(BeFalse())
	})

	DescribeTable("rejects incomplete configuration", func(configuration string) {
		var policy policyEvaluation
		err := json.Unmarshal([]byte(`{"status":"approved","configuration":`+configuration+`}`), &policy)
		Expect(err).To(MatchError(ContainSubstring("policy configuration")))
	},
		Entry("null", `null`),
		Entry("missing blocking flag", `{"id":3,"isEnabled":true,"type":{"id":"type-id"}}`),
		Entry("missing enabled flag", `{"id":3,"isBlocking":true,"type":{"id":"type-id"}}`),
		Entry("missing identity", `{"isEnabled":true,"isBlocking":true,"type":{"id":"type-id"}}`),
	)
})
