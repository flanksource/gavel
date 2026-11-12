package projects

import (
	"context"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// checkoutWithOrigin is a git checkout whose origin is github.com/<repo>.
func checkoutWithOrigin(repo string) (string, *git.Repository) {
	dir := GinkgoT().TempDir()
	checkout, err := git.PlainInit(dir, false)
	Expect(err).NotTo(HaveOccurred())
	setOrigin(checkout, repo)
	return dir, checkout
}

func setOrigin(checkout *git.Repository, repo string) {
	if _, err := checkout.Remote("origin"); err == nil {
		Expect(checkout.DeleteRemote("origin")).To(Succeed())
	}
	_, err := checkout.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{"https://github.com/" + repo + ".git"}})
	Expect(err).NotTo(HaveOccurred())
}

// catalogContract is the behaviour both modes share; newService returns an
// empty catalog.
func catalogContract(newService func() Service) {
	var (
		ctx     context.Context
		service Service
	)
	BeforeEach(func() {
		ctx, service = context.Background(), newService()
	})

	It("resolves a nested directory to the most specific project", func() {
		parent := GinkgoT().TempDir()
		child := filepath.Join(parent, "services", "config-db")
		Expect(os.MkdirAll(filepath.Join(child, "internal"), 0o755)).To(Succeed())
		_, err := service.Create(ctx, Project{Name: "workspace", Dir: parent})
		Expect(err).NotTo(HaveOccurred())
		_, err = service.Create(ctx, Project{Name: "config-db", Dir: child, Repos: []string{"acme/config-db"}})
		Expect(err).NotTo(HaveOccurred())

		Expect(service.ForDir(ctx, filepath.Join(child, "internal"))).To(HaveField("Name", "config-db"))
		Expect(service.ForDir(ctx, parent)).To(HaveField("Name", "workspace"))
		_, err = service.ForDir(ctx, GinkgoT().TempDir())
		Expect(err).To(MatchError(ErrNotFound))
	})

	It("rejects a duplicate name and a project without a dir", func() {
		_, err := service.Create(ctx, Project{Name: "alpha", Dir: "/srv/alpha"})
		Expect(err).NotTo(HaveOccurred())

		_, err = service.Create(ctx, Project{Name: "alpha", Dir: "/srv/other"})
		Expect(err).To(MatchError(ErrExists))
		_, err = service.Create(ctx, Project{Name: "beta"})
		Expect(err).To(MatchError(ErrInvalid))
		Expect(service.Get(ctx, "alpha")).To(HaveField("Dir", "/srv/alpha"))
	})

	It("updates in place by name and deletes", func() {
		created, err := service.Create(ctx, Project{Name: "alpha", Dir: "/srv/alpha", Repos: []string{"acme/alpha"}})
		Expect(err).NotTo(HaveOccurred())

		updated, err := service.Update(ctx, "alpha", Project{Name: "renamed", Dir: "/srv/new", Repos: []string{"acme/new"}})

		Expect(err).NotTo(HaveOccurred())
		Expect(updated).To(Equal(Project{ID: created.ID, TenantID: created.TenantID, Name: "alpha", Kind: KindDir, Dir: "/srv/new", Repos: []string{"acme/new"}}))
		Expect(service.List(ctx)).To(HaveLen(1))
		Expect(service.Delete(ctx, "alpha")).To(Succeed())
		_, err = service.Get(ctx, "alpha")
		Expect(err).To(MatchError(ErrNotFound))
		Expect(service.Delete(ctx, "alpha")).To(MatchError(ErrNotFound))
	})

	It("ensures a web project once and reuses it", func() {
		first, err := service.EnsureWeb(ctx, "docs-site")
		Expect(err).NotTo(HaveOccurred())

		again, err := service.EnsureWeb(ctx, "docs-site")

		Expect(err).NotTo(HaveOccurred())
		Expect(again).To(Equal(first))
		Expect(first).To(And(HaveField("Kind", KindWeb), HaveField("Dir", ""), HaveField("Repos", BeEmpty())))
		Expect(service.List(ctx)).To(HaveLen(1))
	})

	It("keeps web projects out of ForDir and their names unique across kinds", func() {
		dir := GinkgoT().TempDir()
		_, err := service.EnsureWeb(ctx, "site")
		Expect(err).NotTo(HaveOccurred())
		_, err = service.ForDir(ctx, dir)
		Expect(err).To(MatchError(ErrNotFound))

		_, err = service.Create(ctx, Project{Name: "site", Dir: dir})
		Expect(err).To(MatchError(ErrExists))
		_, err = service.Create(ctx, Project{Name: "checkout", Dir: dir})
		Expect(err).NotTo(HaveOccurred())
		_, err = service.EnsureWeb(ctx, "checkout")
		Expect(err).To(MatchError(ErrExists))
		_, err = service.Create(ctx, Project{Name: "web-with-dir", Kind: KindWeb, Dir: dir})
		Expect(err).To(MatchError(ErrInvalid))
	})

	It("infers repos from the checkout's remotes without storing them", func() {
		dir, checkout := checkoutWithOrigin("acme/widget")
		_, err := service.Create(ctx, Project{Name: "widget", Dir: dir})
		Expect(err).NotTo(HaveOccurred())
		Expect(service.Get(ctx, "widget")).To(And(HaveField("Repos", []string{"acme/widget"}), HaveField("ReposInferred", true)))

		// An update echoing the inferred repos keeps them inferred, so they keep
		// following the remote.
		_, err = service.Update(ctx, "widget", Project{Dir: dir, Repos: []string{"acme/widget"}})
		Expect(err).NotTo(HaveOccurred())
		setOrigin(checkout, "acme/moved")

		Expect(service.Get(ctx, "widget")).To(And(HaveField("Repos", []string{"acme/moved"}), HaveField("ReposInferred", true)))
	})
}

var _ = Describe("File catalog", func() {
	var path string
	catalogContract(func() Service {
		path = filepath.Join(GinkgoT().TempDir(), "projects.json")
		return NewFile(path)
	})

	It("ids a project by its name and lists the catalog sorted case-insensitively", func() {
		Expect(os.WriteFile(path, []byte(`[{"name":"infra","dir":"/i"},{"name":"Beta","dir":"/b"},{"name":"alpha","dir":"/a","repos":null}]`), 0o600)).To(Succeed())

		Expect(NewFile(path).List(context.Background())).To(Equal([]Project{
			{ID: FileID("alpha"), TenantID: localTenant, Name: "alpha", Kind: KindDir, Dir: "/a", Repos: []string{}},
			{ID: FileID("Beta"), TenantID: localTenant, Name: "Beta", Kind: KindDir, Dir: "/b", Repos: []string{}},
			{ID: FileID("infra"), TenantID: localTenant, Name: "infra", Kind: KindDir, Dir: "/i", Repos: []string{}},
		}))
	})

	It("names the catalog file when it cannot be parsed", func() {
		Expect(os.WriteFile(path, []byte("{"), 0o600)).To(Succeed())

		_, err := NewFile(path).List(context.Background())

		Expect(err).To(MatchError(ContainSubstring(path)))
	})
})
