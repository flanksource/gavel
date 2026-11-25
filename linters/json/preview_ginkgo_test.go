package jsonlint

import (
	"context"
	"net/url"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Preview declarations", func() {
	var dir string
	fileURL := func(path string) string { return (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String() }
	BeforeEach(func() { dir = GinkgoT().TempDir() })

	DescribeTable("resolves the component a data file declares against the file's folder",
		func(name, source, component string) {
			declared, err := PreviewDeclaration([]byte(source), filepath.Join(dir, name))
			Expect(err).NotTo(HaveOccurred())
			if component == "" {
				Expect(declared).To(BeEmpty())
				return
			}
			Expect(declared).To(Equal(fileURL(filepath.Join(dir, component))))
		},
		Entry("a gavel header", "data.yaml", "# yaml-language-server: $schema=./schema.json\n# gavel: preview=./views/Preview.tsx\nname: example\n", "views/Preview.tsx"),
		Entry("a quoted gavel header in a .yml file", "data.yml", "# gavel: preview=\"views/My Preview.tsx\"\nname: example\n", "views/My Preview.tsx"),
		Entry("a header ahead of a $preview key", "data.yaml", "# gavel: preview=Header.tsx\n$preview: Key.tsx\n", "Header.tsx"),
		Entry("a $preview key in YAML", "data.yaml", "$preview: ./Preview.tsx\nname: example\n", "Preview.tsx"),
		Entry("a $preview key in JSON", "data.json", `{"$schema":"./schema.json","$preview":"./Preview.tsx","name":"example"}`, "Preview.tsx"),
		Entry("a header after the document starts, which is not a header", "data.yaml", "name: example\n# gavel: preview=Late.tsx\n", ""),
		Entry("no declaration", "data.yaml", "name: example\n", ""),
	)

	DescribeTable("rejects a malformed declaration",
		func(name, source, message string) {
			_, err := PreviewDeclaration([]byte(source), filepath.Join(dir, name))
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("an empty header", "data.yaml", "# gavel: preview=\nname: example\n", "nonempty preview="),
		Entry("a header without preview=", "data.yaml", "# gavel: prevue=Preview.tsx\n", "nonempty preview="),
		Entry("a non-string $preview", "data.json", `{"$preview":7}`, "$preview must be a nonempty path"),
		Entry("an empty $preview", "data.yaml", "$preview: ''\n", "$preview must be a nonempty path"),
	)

	It("resolves a schema's x-clicky-preview against the schema that declares it", func() {
		Expect(os.MkdirAll(filepath.Join(dir, "schemas"), 0700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "schemas", "data.json"), []byte(`{"type":"object","x-clicky-preview":{"ref":"views/Preview.tsx"}}`), 0600)).To(Succeed())
		bundle, err := EditorSchema(context.Background(), []byte(`{"$schema":"./schemas/data.json"}`), filepath.Join(dir, "data.json"), "")
		Expect(err).NotTo(HaveOccurred())

		component, err := SchemaPreview(bundle)

		Expect(err).NotTo(HaveOccurred())
		Expect(component).To(Equal(fileURL(filepath.Join(dir, "schemas", "views", "Preview.tsx"))))
	})

	DescribeTable("reads the root schema's x-clicky-preview",
		func(root map[string]any, component, message string) {
			declared, err := SchemaPreview(map[string]any{"$ref": "#/$defs/resource0", "$defs": map[string]any{"resource0": root}})
			if message != "" {
				Expect(err).To(MatchError(ContainSubstring(message)))
				return
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(declared).To(Equal(component))
		},
		Entry("none declared", map[string]any{"type": "object"}, "", ""),
		Entry("a ref", map[string]any{"x-clicky-preview": map[string]any{"ref": "file:///p/Preview.tsx"}}, "file:///p/Preview.tsx", ""),
		Entry("a bare string", map[string]any{"x-clicky-preview": "Preview.tsx"}, "", "x-clicky-preview must be { ref }"),
		Entry("a non-string ref", map[string]any{"x-clicky-preview": map[string]any{"ref": 7}}, "", "x-clicky-preview must be { ref }"),
	)

	It("has no schema preview without a schema", func() {
		Expect(SchemaPreview(nil)).To(BeEmpty())
	})
})
