package editor

import (
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Preview project sources", func() {
	It("uses the nearest package boundary for an explicitly selected nested page", func() {
		root := GinkgoT().TempDir()
		Expect(os.Mkdir(filepath.Join(root, "nested"), 0700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"dependencies":{}}`), 0600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "title.jsx"), []byte("export default () => null"), 0600)).To(Succeed())
		name := filepath.Join(root, "nested", "page.jsx")
		Expect(os.WriteFile(name, []byte("export default () => null"), 0600)).To(Succeed())
		store, err := NewStore([]string{name})
		Expect(err).NotTo(HaveOccurred())
		result, err := store.Preview(store.Files()[0].ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Page).To(Equal("nested/page.jsx"))
		Expect(result.Paths).To(ContainElement("title.jsx"))
	})
	It("indexes project imports without adding support files to editable files", func() {
		store, err := NewStore([]string{"testdata/page-preview/page.jsx"})
		Expect(err).NotTo(HaveOccurred())
		result, err := store.Preview(store.Files()[0].ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Page).To(Equal("page.jsx"))
		Expect(result.Paths).To(ContainElements("title.jsx", "page.css", "page.mdx"))
		Expect(result.Dependencies).To(HaveKeyWithValue("@flanksource/facet", "^0.1.74"))
		Expect(store.Files()).To(HaveLen(1))
	})
	It("serves a schema's form extension refs as paths in the document's preview root", func() {
		root := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(root, "schemas", "extensions"), 0700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "schemas", "extensions", "pickers.tsx"), []byte("export default {}"), 0600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "product.yaml"), []byte("name: example\n"), 0600)).To(Succeed())
		store, err := NewStore([]string{filepath.Join(root, "product.yaml")})
		Expect(err).NotTo(HaveOccurred())
		fileURL := func(path string) string { return (&url.URL{Scheme: "file", Path: filepath.ToSlash(path)}).String() }
		schema := map[string]any{"$defs": map[string]any{"resource0": map[string]any{"x-clicky-extensions": []any{
			map[string]any{"ref": fileURL(filepath.Join(root, "schemas", "extensions", "pickers.tsx"))},
			map[string]any{"tsx": "export default {}"},
		}}}}
		Expect(store.PreviewExtensionRefs(store.Files()[0].ID, schema)).To(Succeed())
		Expect(schema).To(Equal(map[string]any{"$defs": map[string]any{"resource0": map[string]any{"x-clicky-extensions": []any{
			map[string]any{"ref": "schemas/extensions/pickers.tsx"},
			map[string]any{"tsx": "export default {}"},
		}}}}))

		outside := filepath.Join(GinkgoT().TempDir(), "pickers.tsx")
		for ref, message := range map[string]string{
			fileURL(outside):                         "is outside the preview root",
			"https://example.com/pickers.tsx":        "must be a file in the project",
			fileURL(filepath.Join(root, "gone.tsx")): "is not a previewable project source",
		} {
			err := store.PreviewExtensionRefs(store.Files()[0].ID, map[string]any{"x-clicky-extensions": []any{map[string]any{"ref": ref}}})
			Expect(err).To(MatchError(ContainSubstring(message)), ref)
		}
	})
	It("rejects source traversal and symlinks outside the project", func() {
		root := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(root, "page.jsx"), []byte("export default () => null"), 0600)).To(Succeed())
		outside := filepath.Join(GinkgoT().TempDir(), "secret.json")
		Expect(os.WriteFile(outside, []byte("private"), 0600)).To(Succeed())
		Expect(os.Symlink(outside, filepath.Join(root, "linked.json"))).To(Succeed())
		store, err := NewStore([]string{filepath.Join(root, "page.jsx")})
		Expect(err).NotTo(HaveOccurred())
		for _, path := range []string{"../secret.json", "linked.json"} {
			request := httptest.NewRequest("GET", "/source", nil)
			request.SetPathValue("id", store.Files()[0].ID)
			request.SetPathValue("path", path)
			response := httptest.NewRecorder()
			store.previewSource(response, request)
			Expect(response.Code).To(BeNumerically(">=", 400))
			Expect(response.Body.String()).NotTo(ContainSubstring("private"))
		}
	})
})
