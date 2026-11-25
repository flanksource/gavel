package main

import (
	"bytes"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

// The fixture document schema must accept every runner step the runner itself
// decodes: test/lint steps unmarshal onto testrunner.RunOptions / lint.Options,
// so a framework alias ParseFramework accepts and the `runtime` process options
// are valid there even though neither is a CLI flag.
var _ = Describe("fixture document schema", func() {
	var validate func(string) error
	BeforeEach(func() {
		doc, err := fixturesSchemaDocument()
		Expect(err).NotTo(HaveOccurred())
		raw, err := json.Marshal(doc["document"])
		Expect(err).NotTo(HaveOccurred())
		parsed, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
		Expect(err).NotTo(HaveOccurred())
		compiler := jsonschema.NewCompiler()
		Expect(compiler.AddResource("fixture-document.json", parsed)).To(Succeed())
		schema, err := compiler.Compile("fixture-document.json")
		Expect(err).NotTo(HaveOccurred())
		validate = func(document string) error {
			var value any
			Expect(yaml.Unmarshal([]byte(document), &value)).To(Succeed())
			normalized, err := json.Marshal(value)
			Expect(err).NotTo(HaveOccurred())
			instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(normalized))
			Expect(err).NotTo(HaveOccurred())
			return schema.Validate(instance)
		}
	})

	DescribeTable("accepts runner steps the runner decodes", func(step string) {
		Expect(validate("kind: Fixture\nname: example\nsteps:\n" + step)).To(Succeed())
	},
		Entry("a framework alias", "  - kind: test\n    framework: [go]\n"),
		Entry("the canonical framework name", "  - kind: test\n    framework: [go test]\n"),
		Entry("test runtime env", "  - kind: test\n    runtime:\n      env:\n        GOWORK: 'off'\n"),
		Entry("lint runtime timeout", "  - kind: lint\n    runtime:\n      timeout: 2m\n"),
	)

	DescribeTable("rejects runner steps the runner refuses", func(step string) {
		Expect(validate("kind: Fixture\nname: example\nsteps:\n" + step)).NotTo(Succeed())
	},
		Entry("an unknown framework", "  - kind: test\n    framework: [cobol]\n"),
		Entry("an unknown runtime option", "  - kind: test\n    runtime:\n      priority: high\n"),
	)
})
