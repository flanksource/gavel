package schemas

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSchemas(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Embedded Schemas")
}

var _ = Describe("Embedded action documents", func() {
	DescribeTable("loads an embedded schema and rejects unknown fields", func(action string) {
		document, err := Load(action)
		Expect(err).NotTo(HaveOccurred())
		Expect(document.Schema["$id"]).To(Equal("urn:gavel:project-action:" + action + ":v2"))
		Expect(document.Validate(map[string]any{"unexpected": true})).To(HaveOccurred())
	}, Entry("commit", "commit"), Entry("lint", "lint"), Entry("test", "test"))

	It("rejects unknown schema names", func() {
		_, err := Load("../commit")
		Expect(err).To(MatchError(ContainSubstring("unsupported action schema")))
	})

	DescribeTable("rejects incomplete or ambiguous serialization order", func(document string) {
		_, err := decodeDocument([]byte(document))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("x-order"))
	},
		Entry("missing", `{"type":"object","properties":{"name":{"type":"string"}}}`),
		Entry("duplicate", `{"type":"object","properties":{"a":{"type":"string"},"b":{"type":"string"}},"x-order":["a","a"]}`),
		Entry("unknown", `{"type":"object","properties":{"a":{"type":"string"}},"x-order":["b"]}`),
	)

	DescribeTable("rejects property shapes the command serializer cannot consume", func(document, message string) {
		_, err := decodeDocument([]byte(document))
		Expect(err).To(MatchError(ContainSubstring(message)))
	},
		Entry("boolean schema", `{"type":"object","properties":{"name":true},"x-order":["name"]}`, `property "name" must be an object schema`),
		Entry("missing items", `{"type":"object","properties":{"names":{"type":"array"}},"x-order":["names"]}`, `array property "names" must declare its item schema`),
	)
})
