package editor

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing/fstest"

	"github.com/flanksource/gavel/projects"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Editor comment projects", func() {
	It("GET /api/comments/projects lists each edited file's project once, with its edited files, skipping files outside every project", func(ctx SpecContext) {
		alpha, beta, outside := GinkgoT().TempDir(), GinkgoT().TempDir(), GinkgoT().TempDir()
		catalogPath := filepath.Join(GinkgoT().TempDir(), "projects.json")
		DeferCleanup(projects.UseFileForTest(catalogPath))
		for name, dir := range map[string]string{"alpha": alpha, "beta": beta} {
			_, err := projects.NewFile(catalogPath).Create(context.Background(), projects.Project{Name: name, Dir: dir})
			Expect(err).NotTo(HaveOccurred())
		}
		paths := []string{filepath.Join(alpha, "a.yaml"), filepath.Join(alpha, "docs", "b.md"), filepath.Join(beta, "c.json"), filepath.Join(outside, "d.yaml")}
		for _, path := range paths {
			Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
			Expect(os.WriteFile(path, []byte("name: x\n"), 0o600)).To(Succeed())
		}
		store, err := NewStore(paths)
		Expect(err).NotTo(HaveOccurred())
		idOf := map[string]string{}
		for _, file := range store.Files() {
			idOf[filepath.Base(file.Path)] = file.ID
		}
		canonical := func(dir string) string {
			resolved, err := projects.CanonicalPath(dir)
			Expect(err).NotTo(HaveOccurred())
			return resolved
		}

		handler := (&Server{Store: store, Context: context.Background(), Assets: fstest.MapFS{"index.html": {Data: []byte("Editor")}}}).Handler()
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "/api/comments/projects", nil))
		Expect(response.Code).To(Equal(200), response.Body.String())
		var got []CommentProject
		Expect(json.Unmarshal(response.Body.Bytes(), &got)).To(Succeed())

		Expect(got).To(Equal([]CommentProject{
			{Scope: "gavel://project/alpha", Project: "alpha", Dir: canonical(alpha), Files: []CommentProjectFile{
				{ID: idOf["a.yaml"], URI: "gavel://project/alpha/files/a.yaml"},
				{ID: idOf["b.md"], URI: "gavel://project/alpha/files/docs/b.md"},
			}},
			{Scope: "gavel://project/beta", Project: "beta", Dir: canonical(beta), Files: []CommentProjectFile{
				{ID: idOf["c.json"], URI: "gavel://project/beta/files/c.json"},
			}},
		}))
	})
})
