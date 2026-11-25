package jsonlint

import (
	"context"
	"os"
	"path/filepath"

	commonsContext "github.com/flanksource/commons/context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("YAML source diagnostics", func() {
	DescribeTable("includes an excerpt for schema and syntax failures", func(body string, schemaText string, line int) {
		root := GinkgoT().TempDir()
		path, schema := filepath.Join(root, "data.yaml"), filepath.Join(root, "schema.json")
		Expect(os.WriteFile(path, []byte(body), 0600)).To(Succeed())
		Expect(os.WriteFile(schema, []byte(schemaText), 0600)).To(Succeed())
		adapter := New(root)
		adapter.Schema, adapter.Files = schema, []string{path}
		violations, err := adapter.Run(commonsContext.NewContext(context.Background()), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(HaveLen(1))
		Expect(violations[0].Line).To(Equal(line))
		Expect(violations[0].Excerpt).NotTo(BeNil())
		Expect(violations[0].Excerpt.Pretty().String()).To(ContainSubstring(" | "))
	},
		Entry("typed value", "name: 123\n", `{"properties":{"name":{"type":"string"}}}`, 1),
		Entry("missing required property", "other: value\n", `{"required":["name"]}`, 1),
		Entry("malformed YAML", "broken: [\n", `true`, 1),
	)
	It("locates each unknown key in the correct document and nested mapping", func() {
		root := GinkgoT().TempDir()
		schema := filepath.Join(root, "schema.json")
		Expect(os.WriteFile(schema, []byte(`{"type":"object","properties":{"nested":{"type":"object","additionalProperties":false}},"additionalProperties":false}`), 0600)).To(Succeed())
		path := filepath.Join(root, "data.yaml")
		Expect(os.WriteFile(path, []byte("nested: {}\n---\nnested:\n  cleanup: true\n  typo: true\n"), 0600)).To(Succeed())
		adapter := New(root)
		adapter.Schema, adapter.Files = schema, []string{path}
		violations, err := adapter.Run(commonsContext.NewContext(context.Background()), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(HaveLen(2))
		Expect(violations[0].Line).To(Equal(4))
		Expect(violations[0].Column).To(Equal(3))
		Expect(violations[1].Line).To(Equal(5))
		Expect(violations[0].Excerpt).NotTo(BeNil())
		Expect(violations[0].Excerpt.Pretty().String()).To(ContainSubstring("> 4 |   cleanup: true"))
	})
})
