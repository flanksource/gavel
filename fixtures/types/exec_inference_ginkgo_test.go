package types

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var _ = Describe("inferred exec documents", func() {
	DescribeTable("keeps parsing and the document schema consistent", func(step string, valid bool) {
		body := fmt.Sprintf(`{"kind":"Fixture","steps":[%s]}`, step)
		_, parseErr := fixtures.ParseFixtureDocuments("inference.fixture.json", []byte(body))
		schema, err := fixtures.DocumentSchema(nil)
		Expect(err).NotTo(HaveOccurred())
		compiler := jsonschema.NewCompiler()
		encoded, err := json.Marshal(schema)
		Expect(err).NotTo(HaveOccurred())
		resource, err := jsonschema.UnmarshalJSON(bytes.NewReader(encoded))
		Expect(err).NotTo(HaveOccurred())
		Expect(compiler.AddResource("fixture.json", resource)).To(Succeed())
		compiled, err := compiler.Compile("fixture.json")
		Expect(err).NotTo(HaveOccurred())
		instance, err := jsonschema.UnmarshalJSON(bytes.NewBufferString(body))
		Expect(err).NotTo(HaveOccurred())
		Expect([]bool{parseErr == nil, compiled.Validate(instance) == nil}).To(Equal([]bool{valid, valid}))
	},
		Entry("implicit", `{"exec":"printf","args":["ready"]}`, true),
		Entry("explicit", `{"kind":"exec","exec":"printf"}`, true),
		Entry("script with kind", `{"kind":"exec","script":"echo ready"}`, true),
		Entry("script without kind", `{"script":"echo ready"}`, false),
		Entry("invalid executable", `{"exec":false}`, false),
		Entry("empty kind", `{"kind":"","exec":"printf"}`, false),
		Entry("null kind", `{"kind":null,"exec":"printf"}`, false),
		Entry("explicit unknown kind", `{"kind":"missing","exec":"printf"}`, false),
	)

	It("runs inferred YAML exec steps with exports and CEL references", func() {
		nodes := parseYAMLDocument(`kind: Fixture
steps:
  - id: mint
    exec: printf
    args: ['{"token":"fixture-token"}']
    export: {TOKEN: '{{.outputs.json.token}}'}
  - exec: printf
    args: ['%s', '$TOKEN']
    cel: stdout == steps.mint.json.token
`)
		results, _, err := fixtures.RunNodes(context.Background(), nodes, fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})
		Expect(err).NotTo(HaveOccurred())
		Expect(results).To(HaveLen(2))
		for _, result := range results {
			Expect(result.Status).To(Equal(task.StatusPASS), result.Error)
			Expect(result.Type).To(Equal("exec"))
		}
		Expect(json.Valid([]byte(results[0].Stdout))).To(BeTrue())
	})
})
