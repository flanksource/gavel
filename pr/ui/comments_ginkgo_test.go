package ui

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"

	"github.com/flanksource/gavel/projects"

	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gavel comments authorization", func() {
	It("uses registered project scope and rejects foreign repositories and changed captured revisions", func() {
		useProjectsFile(filepath.Join(GinkgoT().TempDir(), "projects.json"))
		Expect(seedProjects([]projects.Project{{Name: "demo", Dir: GinkgoT().TempDir(), Repos: []string{"acme/demo"}}})).To(Succeed())
		access := CommentAccess{}
		ctx := context.Background()
		p, err := access.Principal(ctx)
		Expect(err).NotTo(HaveOccurred())
		target := comments.Target{URI: "https://github.com/acme/demo/pull/1", ScopeURI: "gavel://project/demo", Kind: "pull-request"}
		Expect(access.Authorize(ctx, p, "publish", target)).To(Succeed())
		target.URI = "https://github.com/another/repo/pull/1"
		Expect(access.Authorize(ctx, p, "publish", target)).To(MatchError(comments.ErrForbidden))
		target = comments.Target{URI: "gavel://project/demo/files/a.go", ScopeURI: "gavel://project/demo", Kind: "code", Revision: "new", Selector: json.RawMessage(`{"kind":"code-range","version":1,"path":"a.go","revision":"old","startLine":1,"endLine":1}`)}
		_, err = access.Resolve(ctx, p, target)
		Expect(err).To(MatchError(comments.ErrConflict))
		p.TenantID = "another-tenant"
		Expect(access.Authorize(ctx, p, "read", target)).To(MatchError(comments.ErrForbidden))
	})
	It("resolves a project scope by name without listing the whole catalog", func() {
		path := filepath.Join(GinkgoT().TempDir(), "projects.json")
		dir := GinkgoT().TempDir()
		file := projects.NewFile(path)
		_, err := file.Create(context.Background(), projects.Project{Name: "demo", Dir: dir, Repos: []string{"acme/demo"}})
		Expect(err).NotTo(HaveOccurred())
		_, err = file.Create(context.Background(), projects.Project{Name: "site", Kind: projects.KindWeb})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(projects.UseForTest(noListCatalog{file}, path))
		ctx := context.Background()
		workdir, err := CommentWorkdir(ctx, comments.Target{ScopeURI: "gavel://project/demo"})
		Expect(err).NotTo(HaveOccurred())
		expected, err := projects.CanonicalPath(dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(workdir).To(Equal(expected))
		for _, scope := range []string{"gavel://project/missing", "gavel://project/site"} {
			_, err = CommentWorkdir(ctx, comments.Target{ScopeURI: scope})
			Expect(err).To(MatchError(comments.ErrForbidden), scope)
		}
	})
})

// noListCatalog fails every List, so a test proves a lookup reads one project.
type noListCatalog struct{ projects.Service }

func (noListCatalog) List(context.Context) ([]projects.Project, error) {
	return nil, errors.New("comment scope lookups must not list the whole catalog")
}
