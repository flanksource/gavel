package editor

import (
	"net/http/httptest"
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
