package agents_test

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/comments/agents"
	"github.com/flanksource/gavel/projects"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProjectPathTarget", func() {
	var catalog projects.Service
	var root string

	BeforeEach(func(ctx SpecContext) {
		var err error
		root, err = projects.CanonicalPath(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		Expect(os.MkdirAll(filepath.Join(root, "pkg", "sub"), 0o700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "pkg", "sub", "a.go"), []byte("package sub\n"), 0o600)).To(Succeed())
		file := projects.NewFile(filepath.Join(GinkgoT().TempDir(), "projects.json"))
		_, err = file.Create(ctx, projects.Project{Name: "acme corp", Kind: projects.KindDir, Dir: root})
		Expect(err).NotTo(HaveOccurred())
		catalog = file
	})

	DescribeTable("names the project scope and file URI a path lives at",
		func(ctx SpecContext, rel string, dir bool, uri string) {
			target, err := agents.ProjectPathTarget(ctx, catalog, filepath.Join(root, rel))
			Expect(err).NotTo(HaveOccurred())
			Expect(target).To(Equal(agents.PathTarget{Project: "acme corp", Dir: root, Scope: "gavel://project/acme%20corp", URI: uri, IsDir: dir}))
		},
		Entry("a file", "pkg/sub/a.go", false, "gavel://project/acme%20corp/files/pkg/sub/a.go"),
		Entry("a nested directory", "pkg/sub", true, "gavel://project/acme%20corp/files/pkg/sub"),
		Entry("the project root", ".", true, ""),
		Entry("a file that does not exist yet", "pkg/new.go", false, "gavel://project/acme%20corp/files/pkg/new.go"),
	)

	It("fails with projects.ErrNotFound outside every project", func(ctx SpecContext) {
		_, err := agents.ProjectPathTarget(ctx, catalog, GinkgoT().TempDir())
		Expect(err).To(MatchError(projects.ErrNotFound))
	})
})
