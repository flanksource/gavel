package editor

import (
	"context"
	"github.com/flanksource/gavel/workspaceaccess"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
)

type editorGrants struct{ roots []workspaceaccess.Root }

func (g *editorGrants) List(context.Context) ([]workspaceaccess.Root, error) { return g.roots, nil }
func (g *editorGrants) Grant(context.Context, string) (workspaceaccess.Root, error) {
	panic("unexpected grant")
}
func (g *editorGrants) Revoke(context.Context, string) error { g.roots = nil; return nil }

var _ = Describe("editor workspace access", func() {
	It("opens empty folders, preserves existing IDs, and enforces revocation on reads and writes", func() {
		initial := GinkgoT().TempDir()
		opened := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(initial, "first.md"), []byte("first"), 0600)).To(Succeed())
		store, err := NewStore([]string{initial})
		Expect(err).NotTo(HaveOccurred())
		store.EnableAccess()
		canonical, err := workspaceaccess.Canonical(opened)
		Expect(err).NotTo(HaveOccurred())
		grants := &editorGrants{roots: []workspaceaccess.Root{{ID: "saved", Path: canonical}}}
		store.Access.Repository = grants
		original := store.Files()[0]
		Expect(store.OpenFolder(context.Background(), opened)).To(Succeed())
		Expect(store.Files()).To(ConsistOf(original))
		Expect(os.WriteFile(filepath.Join(opened, "second.mdx"), []byte("second"), 0600)).To(Succeed())
		Expect(store.OpenFolder(context.Background(), opened)).To(Succeed())
		Expect(store.Files()).To(HaveLen(2))
		var added File
		for _, f := range store.Files() {
			if f.ID != original.ID {
				added = f
			}
		}
		doc, err := store.Read(added.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(doc.Source).To(Equal("second"))
		Expect(grants.Revoke(context.Background(), "saved")).To(Succeed())
		_, err = store.Read(added.ID)
		Expect(err).To(BeAssignableToTypeOf(&workspaceaccess.Denied{}))
		_, err = store.Write(doc)
		Expect(err).To(BeAssignableToTypeOf(&workspaceaccess.Denied{}))
		Expect(store.Read(original.ID)).To(HaveField("Source", "first"))
		server := &Server{Store: store}
		Expect(server.Reload()).To(Succeed())
		Expect(server.lintState().Error).To(ContainSubstring("Workspace access was revoked"))
	})
})
