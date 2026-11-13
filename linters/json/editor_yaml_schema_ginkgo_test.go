package jsonlint

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Editor YAML schemas", func() {
	It("bundles YAML schemas and relative YML references without losing numeric precision", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "schema.yaml"), []byte("x-help-display: hover\ntype: object\nproperties:\n  count:\n    $ref: types.yml#/$defs/count\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "types.yml"), []byte("$defs:\n  count:\n    type: integer\n    minimum: 9007199254740993\n"), 0o600)).To(Succeed())
		schema, err := EditorSchema(context.Background(), []byte("# yaml-language-server: $schema=./schema.yaml\ncount: 9007199254740993\n"), filepath.Join(dir, "config.yaml"), "")
		Expect(err).NotTo(HaveOccurred())
		Expect(schema["$defs"]).To(HaveLen(2))
		encoded, err := json.Marshal(schema)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).To(ContainSubstring("9007199254740993"))
		Expect(string(encoded)).NotTo(ContainSubstring("types.yml"))
	})

	It("loads extensionless remote YAML schemas by content type", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
			_, err := w.Write([]byte("type: object\nproperties:\n  name:\n    type: string\n"))
			Expect(err).NotTo(HaveOccurred())
		}))
		DeferCleanup(server.Close)
		_, err := EditorSchema(context.Background(), []byte("name: example"), "config.yaml", server.URL+"/schema")
		Expect(err).NotTo(HaveOccurred())
	})
})
