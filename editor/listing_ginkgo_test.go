package editor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/gavel/workspaceaccess"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// writeTree writes each root-relative file under root.
func writeTree(root string, files map[string]string) {
	GinkgoHelper()
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
		Expect(os.WriteFile(path, []byte(content), 0o600)).To(Succeed())
	}
}

// listedPaths are the store's files relative to root.
func listedPaths(store *Store, root string) []string {
	GinkgoHelper()
	paths := []string{}
	for _, file := range store.Files() {
		rel, err := filepath.Rel(root, file.Path)
		Expect(err).NotTo(HaveOccurred())
		paths = append(paths, filepath.ToSlash(rel))
	}
	return paths
}

// localRequest is a same-origin request from a loopback caller, which the
// editor's folder-changing routes require.
func localRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Host, req.RemoteAddr = "127.0.0.1:9999", "127.0.0.1:40000"
	req.Header.Set("Content-Type", "application/json")
	return req
}

var _ = Describe("Editor file listing", func() {
	var root string
	BeforeEach(func() {
		root = GinkgoT().TempDir()
		// A .git directory makes root a work tree, so its .gitignore applies.
		Expect(os.Mkdir(filepath.Join(root, ".git"), 0o755)).To(Succeed())
		writeTree(root, map[string]string{
			".gitignore":       ".tmp/\ndist/\n",
			"README.md":        "# readme",
			"main.go":          "package main",
			"logo.png":         "png",
			".tmp/scratch.md":  "scratch",
			".tmp/notes.go":    "package notes",
			"dist/guide/a.md":  "a",
			"dist/guide/b.txt": "b",
		})
	})

	DescribeTable("lists the files the listing names",
		func(listing verify.EditConfig, expected []string) {
			store, err := NewStore(StoreOptions{Paths: []string{root}, Listing: listing})
			Expect(err).NotTo(HaveOccurred())
			Expect(listedPaths(store, root)).To(ConsistOf(expected))
		},
		Entry("the built-in documents and images by default, skipping gitignored files",
			verify.EditConfig{}, []string{"README.md", "logo.png"}),
		Entry("exactly the configured extensions, any of them as documents",
			verify.EditConfig{Extensions: []string{".go"}}, []string{"main.go"}),
		Entry("everything below a gitignored directory pattern, by listed extension",
			verify.EditConfig{IncludeIgnored: []string{".tmp"}}, []string{"README.md", "logo.png", ".tmp/scratch.md"}),
		Entry("a gitignored glob, combined with configured extensions",
			verify.EditConfig{Extensions: []string{".md", ".txt"}, IncludeIgnored: []string{"dist/**/*.txt"}}, []string{"README.md", "dist/guide/b.txt"}),
	)

	It("classifies a configured non-image extension as an editable document", func() {
		store, err := NewStore(StoreOptions{Paths: []string{root}, Listing: verify.EditConfig{Extensions: []string{".go"}}})
		Expect(err).NotTo(HaveOccurred())
		doc, err := store.Read(store.Files()[0].ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(doc.Kind).To(Equal(KindDocument))
		Expect(doc.Source).To(Equal("package main"))
	})

	It("opens a file named outright even when the listing omits its extension", func() {
		store, err := NewStore(StoreOptions{Paths: []string{filepath.Join(root, "README.md")}, Listing: verify.EditConfig{Extensions: []string{".go"}}})
		Expect(err).NotTo(HaveOccurred())
		Expect(listedPaths(store, root)).To(ConsistOf("README.md"))
	})

	It("rejects an include pattern that leaves the folder", func() {
		_, err := NewStore(StoreOptions{Paths: []string{root}, Listing: verify.EditConfig{IncludeIgnored: []string{"../other"}}})
		Expect(err).To(MatchError(ContainSubstring("must be a pattern inside the folder")))
	})

	It("relists the selected folders by a new listing", func() {
		store, err := NewStore(StoreOptions{Paths: []string{root}})
		Expect(err).NotTo(HaveOccurred())
		Expect(store.Relist(verify.EditConfig{Extensions: []string{".go"}, IncludeIgnored: []string{".tmp"}})).To(Succeed())
		Expect(listedPaths(store, root)).To(ConsistOf("main.go", ".tmp/notes.go"))
	})

	Describe("over HTTP", func() {
		var (
			server *Server
			other  string
			grants *editorGrants
		)
		BeforeEach(func() {
			other = GinkgoT().TempDir()
			writeTree(other, map[string]string{"other.md": "other", "tool.go": "package tool", ".gavel.yaml": "edit:\n  extensions: [.go]\n"})
			useProject(other)
			store, err := NewStore(StoreOptions{Paths: []string{root}})
			Expect(err).NotTo(HaveOccurred())
			store.EnableAccess()
			grants = &editorGrants{}
			store.Access.Repository = grants
			server = &Server{Store: store}
		})
		serve := func(req *http.Request) *httptest.ResponseRecorder {
			recorder := httptest.NewRecorder()
			server.Handler().ServeHTTP(recorder, req)
			return recorder
		}
		files := func() []string {
			GinkgoHelper()
			recorder := serve(httptest.NewRequest("GET", "/api/files", nil))
			Expect(recorder.Code).To(Equal(http.StatusOK))
			var listed []File
			Expect(json.Unmarshal(recorder.Body.Bytes(), &listed)).To(Succeed())
			names := []string{}
			for _, file := range listed {
				names = append(names, filepath.Base(file.Path))
			}
			return names
		}

		It("lists the registered projects with the open folders", func() {
			recorder := serve(httptest.NewRequest("GET", "/api/edit/projects", nil))
			Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
			var body struct {
				Projects []EditProject `json:"projects"`
				Roots    []string      `json:"roots"`
			}
			Expect(json.Unmarshal(recorder.Body.Bytes(), &body)).To(Succeed())
			Expect(body.Projects).To(Equal([]EditProject{{Name: fixtureProject, Dir: other}}))
			Expect(body.Roots).To(Equal([]string{root}))
		})

		It("refuses to switch to a project folder without workspace access", func() {
			recorder := serve(localRequest("POST", "/api/edit/project", `{"dir":`+jsonString(other)+`}`))
			Expect(recorder.Code).To(Equal(http.StatusForbidden), recorder.Body.String())
			Expect(files()).To(ConsistOf("README.md", "logo.png"))
		})

		It("switches every edited file to the project, listed by its .gavel.yaml", func() {
			canonical, err := workspaceaccess.Canonical(other)
			Expect(err).NotTo(HaveOccurred())
			grants.roots = []workspaceaccess.Root{{ID: "other", Path: canonical}}

			recorder := serve(localRequest("POST", "/api/edit/project", `{"dir":`+jsonString(other)+`}`))
			Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
			Expect(files()).To(ConsistOf("tool.go"))
			Expect(server.Store.Roots()).To(Equal([]string{other}))
		})

		It("lets a flag override the switched-to project's configured extensions", func() {
			canonical, err := workspaceaccess.Canonical(other)
			Expect(err).NotTo(HaveOccurred())
			grants.roots = []workspaceaccess.Root{{ID: "other", Path: canonical}}
			server.ListingFlags = verify.EditConfig{Extensions: []string{".md"}}

			Expect(serve(localRequest("POST", "/api/edit/project", `{"dir":`+jsonString(other)+`}`)).Code).To(Equal(http.StatusOK))
			Expect(files()).To(ConsistOf("other.md"))
		})

		It("applies a listing without saving it", func() {
			recorder := serve(localRequest("PUT", "/api/edit/options", `{"extensions":[".go"],"save":false}`))
			Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
			Expect(files()).To(ConsistOf("main.go"))
			Expect(filepath.Join(root, verify.GavelConfigFileName)).NotTo(BeAnExistingFile())
		})

		It("saves a listing to the first folder's .gavel.yaml and reports where each value came from", func() {
			recorder := serve(localRequest("PUT", "/api/edit/options", `{"extensions":[".md"],"includeIgnored":[".tmp"],"save":true}`))
			Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
			Expect(files()).To(ConsistOf("README.md", "scratch.md"))

			saved, err := verify.LoadSingleGavelConfig(filepath.Join(root, verify.GavelConfigFileName))
			Expect(err).NotTo(HaveOccurred())
			Expect(saved.Edit).To(Equal(verify.EditConfig{Extensions: []string{".md"}, IncludeIgnored: []string{".tmp"}}))

			recorder = serve(httptest.NewRequest("GET", "/api/edit/options", nil))
			var listing EditListing
			Expect(json.Unmarshal(recorder.Body.Bytes(), &listing)).To(Succeed())
			Expect(listing.Defaults).To(ContainElements(".md", ".yaml", ".png"))
			listing.Defaults = nil
			Expect(listing).To(Equal(EditListing{EditConfig: saved.Edit, Configured: saved.Edit, Dir: root}))
		})

		It("rejects an invalid listing without changing the files", func() {
			recorder := serve(localRequest("PUT", "/api/edit/options", `{"extensions":["go"]}`))
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(ContainSubstring(`"go" must look like .md`))
			Expect(files()).To(ConsistOf("README.md", "logo.png"))
		})
	})
})
