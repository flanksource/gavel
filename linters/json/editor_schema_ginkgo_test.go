package jsonlint

import (
	"context"
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/url"
	"os"
	"path/filepath"
)

var _ = Describe("Editor schemas", func() {
	It("bundles relative schema references with internal pointers", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "schema.json"), []byte(`{"type":"object","properties":{"name":{"$ref":"types.json#/$defs/name"}}}`), 0600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "types.json"), []byte(`{"$defs":{"name":{"type":"string"}}}`), 0600)).To(Succeed())
		schema, err := EditorSchema(context.Background(), []byte("# yaml-language-server: $schema=./schema.json\nname: example\n"), filepath.Join(dir, "config.yaml"), "")
		Expect(err).NotTo(HaveOccurred())
		Expect(schema["$defs"]).To(HaveLen(2))
		encoded, err := json.Marshal(schema)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).NotTo(ContainSubstring("types.json"))
	})
	It("resolves form extension refs against the schema file that declares them", func() {
		dir := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(dir, "schemas"), 0700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "schemas", "form.json"), []byte(`{"type":"object","x-clicky-extensions":[{"ref":"extensions/pickers.tsx"},{"tsx":"export default {}","name":"inline"}]}`), 0600)).To(Succeed())
		schema, err := EditorSchema(context.Background(), []byte("# yaml-language-server: $schema=schemas/form.json\nname: example\n"), filepath.Join(dir, "config.yaml"), "")
		Expect(err).NotTo(HaveOccurred())
		root := (&url.URL{Scheme: "file", Path: filepath.ToSlash(dir)}).String()
		Expect(schema["$defs"]).To(Equal(map[string]any{"resource0": map[string]any{"type": "object", "x-clicky-extensions": []any{
			map[string]any{"ref": root + "/schemas/extensions/pickers.tsx"},
			map[string]any{"tsx": "export default {}", "name": "inline"},
		}}}))
	})
	It("rejects form extensions that are not a list of sources", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "form.json"), []byte(`{"x-clicky-extensions":[{"ref":7}]}`), 0600)).To(Succeed())
		_, err := EditorSchema(context.Background(), []byte("# yaml-language-server: $schema=form.json\n"), filepath.Join(dir, "config.yaml"), "")
		Expect(err).To(MatchError(ContainSubstring("x-clicky-extensions[0].ref must be a string")))
	})
	It("does not invent schemas and reports missing declared schemas", func() {
		path := filepath.Join(GinkgoT().TempDir(), "config.yaml")
		schema, err := EditorSchema(context.Background(), []byte("name: example\n"), path, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(schema).To(BeNil())
		_, err = EditorSchema(context.Background(), []byte("$schema: ./missing.json\n"), path, "")
		Expect(err).To(HaveOccurred())
	})
	It("names the file, the step and the offending line when the source does not parse", func() {
		dir := GinkgoT().TempDir()
		source := []byte("name: example\ndescription: Design diagrams: the house way\n")
		_, err := EditorSchema(context.Background(), source, filepath.Join(dir, "config.yaml"), "")
		Expect(err).To(MatchError(`read the $schema declaration of config.yaml: yaml: line 2: mapping values are not allowed in this context: "description: Design diagrams: the house way"`))
		_, err = EditorSchema(context.Background(), source, filepath.Join(dir, "SKILL.md"), "")
		Expect(err).To(MatchError(`read the $schema declaration of SKILL.md: yaml: frontmatter line 2 (file line 3): mapping values are not allowed in this context: "description: Design diagrams: the house way"`))
	})
})
