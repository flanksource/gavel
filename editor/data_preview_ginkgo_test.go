package editor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Data file preview components", func() {
	var root string
	write := func(name, content string) string {
		path := filepath.Join(root, filepath.FromSlash(name))
		Expect(os.MkdirAll(filepath.Dir(path), 0700)).To(Succeed())
		Expect(os.WriteFile(path, []byte(content), 0600)).To(Succeed())
		return path
	}
	serverFor := func(name string) (*Server, string) {
		store, err := NewStore(StoreOptions{Paths: []string{filepath.Join(root, name)}})
		Expect(err).NotTo(HaveOccurred())
		return &Server{Store: store}, store.Files()[0].ID
	}
	BeforeEach(func() {
		root = GinkgoT().TempDir()
		write("package.json", `{"dependencies":{}}`)
		write("views/Header.tsx", "export default () => null")
		write("schemas/views/Schema.tsx", "export default () => null")
		write("schemas/data.json", `{"type":"object","x-clicky-preview":{"ref":"views/Schema.tsx"}}`)
		write("page.css", "body {}")
	})

	DescribeTable("names the component as a path in the preview root",
		func(name, source, component string) {
			write(name, source)
			server, id := serverFor(name)
			Expect(server.PreviewComponent(context.Background(), id, source)).To(Equal(component))
		},
		Entry("from the YAML header ahead of the schema", "data.yaml", "# yaml-language-server: $schema=./schemas/data.json\n# gavel: preview=./views/Header.tsx\nname: example\n", "views/Header.tsx"),
		Entry("from a JSON $preview key", "data.json", `{"$preview":"./views/Header.tsx"}`, "views/Header.tsx"),
		Entry("from the schema's x-clicky-preview", "data.yaml", "# yaml-language-server: $schema=./schemas/data.json\nname: example\n", "schemas/views/Schema.tsx"),
		Entry("as none when nothing declares one", "data.yaml", "name: example\n", ""),
	)

	It("reads the unsaved source, not the file on disk", func() {
		write("data.yaml", "name: example\n")
		server, id := serverFor("data.yaml")
		Expect(server.PreviewComponent(context.Background(), id, "# gavel: preview=views/Header.tsx\n")).To(Equal("views/Header.tsx"))
	})

	DescribeTable("rejects a component it cannot render",
		func(source, message string) {
			write("data.yaml", source)
			server, id := serverFor("data.yaml")
			_, err := server.PreviewComponent(context.Background(), id, source)
			Expect(err).To(MatchError(ContainSubstring(message)))
		},
		Entry("outside the preview root", "# gavel: preview=../elsewhere/Preview.tsx\n", "is outside the preview root"),
		Entry("over http", "# gavel: preview=https://example.com/Preview.tsx\n", "must be a file in the project"),
		Entry("missing", "# gavel: preview=views/Gone.tsx\n", "is not a previewable project source"),
		Entry("not a page", "# gavel: preview=page.css\n", "must be one of .tsx, .jsx, .mdx"),
		Entry("an empty header", "# gavel: preview=\n", "nonempty preview="),
	)

	It("answers a document that is not JSON or YAML with a bad request", func() {
		write("notes.md", "# notes\n")
		server, id := serverFor("notes.md")
		request := httptest.NewRequest(http.MethodPost, "/api/file/"+id+"/preview/component", strings.NewReader(`{"source":"# notes\n"}`))
		request.SetPathValue("id", id)
		response := httptest.NewRecorder()

		server.previewComponent(response, request)

		Expect([]any{response.Code, strings.TrimSpace(response.Body.String())}).To(Equal([]any{http.StatusBadRequest, filepath.Join(root, "notes.md") + " is not a JSON or YAML document; only data files preview through a component"}))
	})
})
