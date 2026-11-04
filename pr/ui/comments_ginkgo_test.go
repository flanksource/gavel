package ui

import (
	"context"
	"encoding/json"
	"path/filepath"

	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gavel comments authorization", func() {
	It("uses registered project scope and rejects foreign repositories and changed captured revisions", func() {
		previous := projectsPath
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
		DeferCleanup(func() { projectsPath = previous })
		Expect(SaveProjects([]Project{{Name: "demo", Dir: GinkgoT().TempDir(), Repos: []string{"acme/demo"}}})).To(Succeed())
		access := commentAccess{}
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
})
