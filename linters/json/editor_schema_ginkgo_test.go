package jsonlint

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Editor schemas", func() {
	It("refuses remote document declarations without sending a request", func() {
		var requests atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requests.Add(1)
			_, err := w.Write([]byte(`true`))
			Expect(err).NotTo(HaveOccurred())
		}))
		DeferCleanup(server.Close)
		_, err := EditorSchema(context.Background(), EditorSchemaOptions{Data: []byte(`{"$schema":"` + server.URL + `/schema"}`), Path: "config.json", Override: ""})
		Expect(err).To(MatchError(ContainSubstring("not fetched")))
		Expect(requests.Load()).To(BeZero())
	})
	It("bundles relative schema references with internal pointers", func() {
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "schema.json"), []byte(`{"type":"object","properties":{"name":{"$ref":"types.json#/$defs/name"}}}`), 0600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "types.json"), []byte(`{"$defs":{"name":{"type":"string"}}}`), 0600)).To(Succeed())
		schema, err := EditorSchema(context.Background(), EditorSchemaOptions{Data: []byte("# yaml-language-server: $schema=./schema.json\nname: example\n"), Path: filepath.Join(dir, "config.yaml"), Override: ""})
		Expect(err).NotTo(HaveOccurred())
		Expect(schema["$defs"]).To(HaveLen(2))
		encoded, err := json.Marshal(schema)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(encoded)).NotTo(ContainSubstring("types.json"))
	})
	It("does not invent schemas and reports missing declared schemas", func() {
		path := filepath.Join(GinkgoT().TempDir(), "config.yaml")
		schema, err := EditorSchema(context.Background(), EditorSchemaOptions{Data: []byte("name: example\n"), Path: path, Override: ""})
		Expect(err).NotTo(HaveOccurred())
		Expect(schema).To(BeNil())
		_, err = EditorSchema(context.Background(), EditorSchemaOptions{Data: []byte("$schema: ./missing.json\n"), Path: path, Override: ""})
		Expect(err).To(HaveOccurred())
	})
})
