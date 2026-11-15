package workspaceaccess

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAccess(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Workspace access") }

type memory struct{ roots []Root }

func (m *memory) List(context.Context) ([]Root, error) { return m.roots, nil }
func (m *memory) Grant(_ context.Context, path string) (Root, error) {
	r := Root{ID: "saved", Path: path, Source: "saved"}
	m.roots = append(m.roots, r)
	return r, nil
}
func (m *memory) Revoke(context.Context, string) error { m.roots = nil; return nil }

var _ = Describe("workspace access", func() {
	It("shares grants between services and rejects revoked and escaped paths", func() {
		root := GinkgoT().TempDir()
		root, err := Canonical(root)
		Expect(err).NotTo(HaveOccurred())
		child := filepath.Join(root, "docs")
		Expect(os.Mkdir(child, 0700)).To(Succeed())
		repo := &memory{}
		one := &Service{Repository: repo}
		two := &Service{Repository: repo}
		_, err = two.Authorize(context.Background(), child)
		Expect(err).To(HaveOccurred())
		_, err = one.Grant(context.Background(), root, root)
		Expect(err).NotTo(HaveOccurred())
		Expect(two.Authorize(context.Background(), child)).To(Equal(child))
		outside := GinkgoT().TempDir()
		Expect(os.Symlink(outside, filepath.Join(root, "escape"))).To(Succeed())
		_, err = two.Authorize(context.Background(), filepath.Join(root, "escape"))
		Expect(err).To(HaveOccurred())
		Expect(repo.Revoke(context.Background(), "saved")).To(Succeed())
		_, err = two.Authorize(context.Background(), child)
		Expect(err).To(HaveOccurred())
	})
	It("refuses management requests from remote callers and cross-origin browsers", func() {
		for _, caller := range []struct {
			peer, host, origin string
			status             int
		}{
			{"192.0.2.1:1234", "localhost", "", 403},
			{"127.0.0.1:1234", "localhost", "http://evil.example", 403},
			{"127.0.0.1:1234", "evil.example", "", 403},
			{"127.0.0.1:1234", "localhost", "http://localhost", 200},
			{"[::1]:1234", "[::1]:8080", "http://[::1]:8080", 200},
		} {
			req := httptest.NewRequest("GET", "http://"+caller.host+"/api/workspace/access", nil)
			req.RemoteAddr = caller.peer
			req.Header.Set("Origin", caller.origin)
			res := httptest.NewRecorder()
			(&Service{Repository: &memory{}}).Handler().ServeHTTP(res, req)
			Expect(res.Code).To(Equal(caller.status))
		}
	})
	It("requires the reviewed canonical path and keeps command-line roots fixed", func() {
		root, err := Canonical(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		repo := &memory{}
		service := &Service{Repository: repo, Pinned: []string{root}}
		_, err = service.Grant(context.Background(), root, root+"/changed")
		Expect(err).To(MatchError("folder changed since preview; review access again"))
		_, err = service.Grant(context.Background(), root, root)
		Expect(err).To(MatchError("folder is already enabled by the command line"))
		Expect(repo.roots).To(BeEmpty())
		Expect(service.Authorize(context.Background(), root)).To(Equal(root))
	})
})
