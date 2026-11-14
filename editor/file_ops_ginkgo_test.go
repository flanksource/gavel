package editor

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing/fstest"

	"github.com/flanksource/gavel/fixtures"
	_ "github.com/flanksource/gavel/fixtures/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("File operations", func() {
	var dir string
	var store *Store
	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Readme\n"), 0o600)).To(Succeed())
		var err error
		store, err = NewStore([]string{dir})
		Expect(err).NotTo(HaveOccurred())
	})
	readme := func() Document {
		doc, err := store.Read(store.Files()[0].ID)
		Expect(err).NotTo(HaveOccurred())
		return doc
	}

	Describe("templates", func() {
		templates, err := Templates()
		It("loads every template source", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(templates).To(HaveLen(len(templateCatalog)))
		})
		for _, template := range templates {
			It("creates a valid "+template.Label+" as "+template.Name, func() {
				path := filepath.Join(dir, "new", template.Name)
				doc, err := store.Create(path, template.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(doc.File).To(Equal(File{ID: doc.ID, Path: path, Kind: KindDocument}))
				Expect(doc.Source).To(Equal(template.Source))
				Expect(store.Files()).To(ContainElement(doc.File))
				switch filepath.Ext(template.Name) {
				case ".json":
					Expect(json.Valid([]byte(template.Source))).To(BeTrue(), template.Source)
				case ".yaml":
					var value map[string]any
					Expect(yaml.Unmarshal([]byte(template.Source), &value)).To(Succeed())
				}
			})
		}
		It("seeds a Markdown fixture gavel fixtures parses into a test", func() {
			doc, err := store.Create(filepath.Join(dir, "smoke.fixture.md"), "fixture-markdown")
			Expect(err).NotTo(HaveOccurred())
			tree, err := fixtures.ParseMarkdownFixturesWithTree(doc.Path)
			Expect(err).NotTo(HaveOccurred())
			Expect(tree.GetAllTests()).To(HaveLen(1))
		})
		It("seeds a YAML fixture gavel fixtures parses into its exec step", func() {
			doc, err := store.Create(filepath.Join(dir, "smoke.fixture.yaml"), "fixture-yaml")
			Expect(err).NotTo(HaveOccurred())
			node, err := fixtures.ParseFixtureFile(doc.Path)
			Expect(err).NotTo(HaveOccurred())
			Expect(node.GetAllTests()).To(HaveLen(1))
		})
	})

	Describe("Create", func() {
		It("creates an empty document without a template", func() {
			doc, err := store.Create(filepath.Join(dir, "blank.yaml"), "")
			Expect(err).NotTo(HaveOccurred())
			Expect(doc.Source).To(BeEmpty())
		})
		DescribeTable("refuses", func(name, template string, want string) {
			_, err := store.Create(filepath.Join(dir, name), template)
			Expect(err).To(MatchError(ContainSubstring(want)))
		},
			Entry("an existing file", "readme.md", "", "already exists"),
			Entry("a path outside the edited folders", "../escape.md", "", "outside the edited folders"),
			Entry("an unsupported extension", "notes.txt", "", "new files must end in one of"),
			Entry("a name without the template's suffix", "plain.md", "fixture-markdown", "must be named *.fixture.md"),
			Entry("an unknown template", "a.md", "nope", `unknown template "nope"`),
		)
		It("refuses a relative path", func() {
			_, err := store.Create("relative.md", "")
			Expect(err).To(MatchError(ContainSubstring("path must be absolute")))
		})
		AfterEach(func() {
			Expect(filepath.Join(filepath.Dir(dir), "escape.md")).NotTo(BeAnExistingFile())
		})
	})

	Describe("Rename", func() {
		It("moves the file into a new folder under a new id", func() {
			old := readme()
			moved, err := store.Rename(old.ID, filepath.Join(dir, "docs", "intro.md"), old.Revision)
			Expect(err).NotTo(HaveOccurred())
			Expect(moved).To(Equal(File{ID: moved.ID, Path: filepath.Join(dir, "docs", "intro.md"), Kind: KindDocument}))
			Expect(moved.ID).NotTo(Equal(old.ID))
			Expect(store.Files()).To(Equal([]File{moved}))
			Expect(filepath.Join(dir, "readme.md")).NotTo(BeAnExistingFile())
			Expect(os.ReadFile(moved.Path)).To(Equal([]byte("# Readme\n")))
		})
		It("refuses a document changed on disk since it was read", func() {
			old := readme()
			Expect(os.WriteFile(old.Path, []byte("changed"), 0o600)).To(Succeed())
			_, err := store.Rename(old.ID, filepath.Join(dir, "intro.md"), old.Revision)
			Expect(err).To(MatchError(ErrConflict))
		})
		It("never replaces an existing file", func() {
			Expect(os.WriteFile(filepath.Join(dir, "other.md"), []byte("keep"), 0o600)).To(Succeed())
			old := readme()
			_, err := store.Rename(old.ID, filepath.Join(dir, "other.md"), old.Revision)
			Expect(err).To(MatchError(ErrExists))
			Expect(os.ReadFile(filepath.Join(dir, "other.md"))).To(Equal([]byte("keep")))
		})
		It("refuses a name that changes the file's kind", func() {
			old := readme()
			_, err := store.Rename(old.ID, filepath.Join(dir, "readme.png"), old.Revision)
			Expect(err).To(MatchError(ContainSubstring("must keep it a supported document")))
		})
	})

	Describe("Delete", func() {
		It("removes the file from disk and from the edited files", func() {
			old := readme()
			deleted, err := store.Delete(old.ID, old.Revision)
			Expect(err).NotTo(HaveOccurred())
			Expect(deleted).To(Equal(old.File))
			Expect(store.Files()).To(BeEmpty())
			Expect(old.Path).NotTo(BeAnExistingFile())
		})
		It("keeps a document changed on disk since it was read", func() {
			old := readme()
			Expect(os.WriteFile(old.Path, []byte("changed"), 0o600)).To(Succeed())
			_, err := store.Delete(old.ID, old.Revision)
			Expect(err).To(MatchError(ErrConflict))
			Expect(old.Path).To(BeAnExistingFile())
		})
	})

	Describe("HTTP API", func() {
		var handler http.Handler
		BeforeEach(func() {
			handler = (&Server{Store: store, Context: context.Background(), Assets: fstest.MapFS{}}).Handler()
		})
		call := func(method, path string, body any) *httptest.ResponseRecorder {
			data, err := json.Marshal(body)
			Expect(err).NotTo(HaveOccurred())
			request := httptest.NewRequest(method, path, bytes.NewReader(data))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			return response
		}
		It("GET /api/templates lists the catalog with sources", func() {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest("GET", "/api/templates", nil))
			Expect(response.Code).To(Equal(http.StatusOK))
			var templates []Template
			Expect(json.Unmarshal(response.Body.Bytes(), &templates)).To(Succeed())
			Expect(templates).To(HaveLen(len(templateCatalog)))
			Expect(templates[0].Source).NotTo(BeEmpty())
		})
		It("creates, renames and deletes a file", func() {
			created := call("POST", "/api/files", map[string]string{"path": filepath.Join(dir, "page.tsx"), "template": "facet-tsx"})
			Expect(created.Code).To(Equal(http.StatusCreated), created.Body.String())
			var doc Document
			Expect(json.Unmarshal(created.Body.Bytes(), &doc)).To(Succeed())

			renamed := call("POST", "/api/file/"+doc.ID+"/rename", map[string]string{"path": filepath.Join(dir, "home.tsx"), "revision": doc.Revision})
			Expect(renamed.Code).To(Equal(http.StatusOK), renamed.Body.String())
			var moved File
			Expect(json.Unmarshal(renamed.Body.Bytes(), &moved)).To(Succeed())
			Expect(moved.Path).To(Equal(filepath.Join(dir, "home.tsx")))

			deleted := call("DELETE", "/api/file/"+moved.ID, map[string]string{"revision": doc.Revision})
			Expect(deleted.Code).To(Equal(http.StatusOK), deleted.Body.String())
			Expect(store.Files()).To(HaveLen(1))
		})
		It("answers 409 for a file that already exists", func() {
			response := call("POST", "/api/files", map[string]string{"path": filepath.Join(dir, "readme.md")})
			Expect(response.Code).To(Equal(http.StatusConflict), response.Body.String())
		})
	})
})
