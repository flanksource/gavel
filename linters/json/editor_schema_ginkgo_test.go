package jsonlint

import (
	"context"
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	It("does not invent schemas and reports missing declared schemas", func() {
		path := filepath.Join(GinkgoT().TempDir(), "config.yaml")
		schema, err := EditorSchema(context.Background(), []byte("name: example\n"), path, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(schema).To(BeNil())
		_, err = EditorSchema(context.Background(), []byte("$schema: ./missing.json\n"), path, "")
		Expect(err).To(HaveOccurred())
	})
})
