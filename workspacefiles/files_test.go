package workspacefiles

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWorkspaceFiles(t *testing.T) { RegisterFailHandler(Fail); RunSpecs(t, "Workspace files") }

var _ = Describe("Authorized working tree files", func() {
	var service *Service
	var directory string
	BeforeEach(func() {
		directory = GinkgoT().TempDir()
		Expect(exec.Command("git", "init", "-q", directory).Run()).To(Succeed())
		Expect(os.WriteFile(filepath.Join(directory, "guide.md"), []byte("original"), 0o644)).To(Succeed())
		var err error
		service, err = New(Options{Roots: []string{directory}, Authorize: func(path string) (string, error) { return filepath.EvalSymlinks(path) }, IdleTimeout: time.Minute})
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() { Expect(service.Close()).To(Succeed()) })
	It("rejects stale saves and escaping paths without touching disk", func(ctx SpecContext) {
		file, err := service.Read(ctx, FileRequest{Root: directory, Path: "guide.md"})
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(directory, "guide.md"), []byte("external"), 0o644)).To(Succeed())
		_, err = service.Save(ctx, SaveRequest{FileRequest: FileRequest{Root: directory, Path: "guide.md"}, Revision: file.Revision, Source: "draft"})
		Expect(err).To(MatchError(ErrConflict))
		contents, err := os.ReadFile(filepath.Join(directory, "guide.md"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal("external"))
		_, err = service.Read(ctx, FileRequest{Root: directory, Path: "../outside.md"})
		Expect(err).To(HaveOccurred())
	})
	It("rejects symlinks outside the selected repository", func(ctx SpecContext) {
		outside := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(outside, "secret.md"), []byte("secret"), 0o600)).To(Succeed())
		Expect(os.Symlink(outside, filepath.Join(directory, "link"))).To(Succeed())
		_, err := service.Read(ctx, FileRequest{Root: directory, Path: "link/secret.md"})
		Expect(err).To(HaveOccurred())
	})
	It("shares physical watches, filters siblings and expires viewers independently", func(ctx SpecContext) {
		input := FileRequest{Root: directory, Path: "guide.md"}
		first, err := service.Acquire(ctx, input, "one")
		Expect(err).NotTo(HaveOccurred())
		second, err := service.Acquire(ctx, input, "two")
		Expect(err).NotTo(HaveOccurred())
		Expect(service.watchCount()).To(Equal(1))
		events, err := service.Subscribe(first.ID, "one")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(directory, "other.md"), []byte("sibling"), 0o644)).To(Succeed())
		Consistently(events, 20*time.Millisecond).ShouldNot(Receive())
		Expect(os.WriteFile(filepath.Join(directory, "guide.md"), []byte("changed"), 0o644)).To(Succeed())
		Eventually(events).Should(Receive())
		Expect(service.Release(first.ID, "one")).To(Succeed())
		Expect(service.watchCount()).To(Equal(1))
		service.mu.Lock()
		service.leases[second.ID].Expires = time.Now().Add(-time.Second)
		service.mu.Unlock()
		Eventually(func() int { service.expire(time.Now()); return service.watchCount() }).Should(Equal(0))
		_, err = service.Renew(second.ID, "two")
		Expect(err).To(MatchError(ErrExpired))
	})
	It("keeps the selected filename live across an atomic save", func(ctx SpecContext) {
		input := FileRequest{Root: directory, Path: "guide.md"}
		lease, err := service.Acquire(ctx, input, "viewer")
		Expect(err).NotTo(HaveOccurred())
		events, err := service.Subscribe(lease.ID, "viewer")
		Expect(err).NotTo(HaveOccurred())
		replacement := filepath.Join(directory, "replacement.md")
		Expect(os.WriteFile(replacement, []byte("atomic replacement"), 0o644)).To(Succeed())
		Expect(os.Rename(replacement, filepath.Join(directory, "guide.md"))).To(Succeed())
		Eventually(events).Should(Receive())
		file, err := service.Read(ctx, input)
		Expect(err).NotTo(HaveOccurred())
		Expect(file.Source).To(Equal("atomic replacement"))
		service.mu.Lock()
		expires := service.leases[lease.ID].Expires
		service.mu.Unlock()
		Expect(expires).To(Equal(lease.Expires))
	})
	It("does not accept a nested directory as a repository root", func(ctx SpecContext) {
		nested := filepath.Join(directory, "nested")
		Expect(os.Mkdir(nested, 0o755)).To(Succeed())
		_, err := service.Repository(context.Background(), nested)
		Expect(err).To(HaveOccurred())
	})
})
