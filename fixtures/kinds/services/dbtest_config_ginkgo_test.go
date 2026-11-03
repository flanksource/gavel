package services_test

import (
	"encoding/json"
	"fmt"

	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var _ = Describe("services kind: database configuration", func() {
	DescribeTable("validates database selection before execution", func(steps, message string) {
		_, err := fixtures.ParseFixtureDocuments("database.fixture.yaml", []byte("kind: Fixture\nsteps:\n"+steps))
		if message == "" {
			Expect(err).NotTo(HaveOccurred())
		} else {
			Expect(err).To(MatchError(ContainSubstring(message)))
		}
	},
		Entry("single database", "  - id: stack\n    kind: services\n    services:\n      db: {dbtest: {}}\n", ""),
		Entry("list form", "  - id: stack\n    kind: services\n    services: [{name: db, dbtest: {}}]\n", ""),
		Entry("ambiguous databases", "  - id: stack\n    kind: services\n    services:\n      one: {dbtest: {}}\n      two: {dbtest: {}}\n", "exactly one dbtest.default"),
		Entry("selected database", "  - id: stack\n    kind: services\n    services:\n      one: {dbtest: {default: true}}\n      two: {dbtest: {}}\n", ""),
		Entry("duplicate defaults", "  - id: stack\n    kind: services\n    services:\n      one: {dbtest: {default: true}}\n      two: {dbtest: {default: true}}\n", "exactly one dbtest.default"),
		Entry("ambiguity across steps", "  - id: first\n    kind: services\n    services: {one: {dbtest: {}}}\n  - id: second\n    kind: services\n    services: {two: {dbtest: {}}}\n", "exactly one dbtest.default"),
		Entry("conflicting service types", "  - id: stack\n    kind: services\n    services: {db: {dbtest: {}, deps: postgres}}\n", "oneOf"),
		Entry("unknown database option", "  - id: stack\n    kind: services\n    services: {db: {dbtest: {typo: true}}}\n", "additional properties"),
	)

	It("selects the sole database across services steps without changing entry order", func() {
		data := `kind: Fixture
steps:
  - id: files
    kind: services
    services: {repo: {git: {files: {README.md: fixture}}}}
  - id: database
    kind: services
    services: {db: {dbtest: {}}}
`
		tree, err := fixtures.ParseFixtureDocuments("database.fixture.yaml", []byte(data))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(tree.Children[1].Test.Step.Config)).To(ContainSubstring(`"default":true`))
	})

	It("rejects invalid database-specific service settings before opening a database", func() {
		for _, field := range []string{`"env":{"A":"b"}`, `"ready":{"tcp":"localhost:1"}`, `"runtime":"docker"`, `"wait":{"timeout":"1s"}`} {
			step := newStep("stack", fmt.Sprintf(`{"services":{"db":{"dbtest":{},%s}}}`, field))
			_, err := execute(step)
			Expect(err).To(MatchError(ContainSubstring("dbtest entry")))
		}
	})

	It("exports a self-contained document schema with database settings", func() {
		schema, err := fixtures.DocumentSchema(nil)
		Expect(err).NotTo(HaveOccurred())
		encoded, err := json.Marshal(schema)
		Expect(err).NotTo(HaveOccurred())
		var document any
		Expect(json.Unmarshal(encoded, &document)).To(Succeed())
		compiler := jsonschema.NewCompiler()
		Expect(compiler.AddResource("fixture://document.json", document)).To(Succeed())
		compiled, err := compiler.Compile("fixture://document.json")
		Expect(err).NotTo(HaveOccurred())
		var example any
		Expect(json.Unmarshal([]byte(`{"kind":"Fixture","steps":[{"id":"stack","kind":"services","services":{"db":{"dbtest":{"migrations":{"dir":"schema"}}}}}]}`), &example)).To(Succeed())
		Expect(compiled.Validate(example)).To(Succeed())
		Expect(json.Unmarshal([]byte(`{"kind":"Fixture","steps":[{"id":"stack","kind":"services","services":{"db":{"dbtest":{"migrations":{"dir":1}}}}}]}`), &example)).To(Succeed())
		Expect(compiled.Validate(example)).NotTo(Succeed())
	})
})
