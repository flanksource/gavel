package jsonlint

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("YAML schemas", func() {
	DescribeTable("loads local YAML schemas and relative references", func(extension string) {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "schema"+extension), []byte("$ref: types"+extension+"#/$defs/name\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "types"+extension), []byte("$defs:\n  name:\n    type: object\n    required: [name]\n"), 0o600)).To(Succeed())
		schema, err := newValidator(context.Background()).compile("schema"+extension, dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(schema.Validate(map[string]any{"name": "example"})).To(Succeed())
		Expect(schema.Validate(map[string]any{})).NotTo(Succeed())
	}, Entry("yaml", ".yaml"), Entry("yml", ".yml"))

	It("loads remote YAML references and preserves exact numbers", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/yaml")
			if r.URL.Path == "/schema.yaml" {
				_, err := w.Write([]byte("$ref: types.yaml\n"))
				Expect(err).NotTo(HaveOccurred())
				return
			}
			_, err := w.Write([]byte("type: object\nrequired: [name]\nproperties:\n  name:\n    type: string\n"))
			Expect(err).NotTo(HaveOccurred())
		}))
		DeferCleanup(server.Close)
		schema, err := newValidator(context.Background()).compile(server.URL+"/schema.yaml", ".")
		Expect(err).NotTo(HaveOccurred())
		Expect(schema.Validate(map[string]any{"name": "example"})).To(Succeed())
		Expect(schema.Validate(map[string]any{})).NotTo(Succeed())
	})

	DescribeTable("rejects invalid YAML schemas", func(contents string) {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "schema.yaml"), []byte(contents), 0o600)).To(Succeed())
		_, err := newValidator(context.Background()).compile("schema.yaml", dir)
		Expect(err).To(HaveOccurred())
	},
		Entry("duplicate keys", "type: string\ntype: number\n"),
		Entry("multiple documents", "type: string\n---\ntype: number\n"),
		Entry("malformed input", "type: [\n"),
	)
})
