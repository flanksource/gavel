package workspace_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/workspace"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWorkspace(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workspace Suite")
}

// tree creates dirs under a fresh, symlink-resolved temp root.
func tree(dirs ...string) string {
	root, err := filepath.EvalSymlinks(GinkgoT().TempDir())
	Expect(err).NotTo(HaveOccurred())
	for _, dir := range dirs {
		Expect(os.MkdirAll(filepath.Join(root, dir), 0o755)).To(Succeed())
	}
	return root
}

var _ = Describe("Resolve", func() {
	DescribeTable("picks the state root",
		func(layout []string, invocation, wantGitRoot, wantStateRoot string) {
			root := tree(layout...)
			dirs, err := workspace.Resolve(filepath.Join(root, invocation))
			Expect(err).NotTo(HaveOccurred())
			Expect(dirs.Invocation).To(Equal(filepath.Join(root, invocation)))
			if wantGitRoot == "-" {
				Expect(dirs.GitRoot).To(BeEmpty())
			} else {
				Expect(dirs.GitRoot).To(Equal(filepath.Join(root, wantGitRoot)))
			}
			Expect(dirs.StateRoot).To(Equal(filepath.Join(root, wantStateRoot)))
		},
		Entry("nested .gavel in a subdirectory wins over the git root",
			[]string{"repo/.git", "repo/.gavel", "repo/svc/.gavel", "repo/svc/pkg"}, "repo/svc/pkg", "repo", "repo/svc"),
		Entry("no .gavel falls back to the git root",
			[]string{"repo/.git", "repo/pkg/sub"}, "repo/pkg/sub", "repo", "repo"),
		Entry("a .gavel above the git root (like ~/.gavel) is ignored",
			[]string{".gavel", "repo/.git", "repo/pkg"}, "repo/pkg", "repo", "repo"),
		Entry("outside a working tree the invocation dir is the state root",
			[]string{".gavel", "plain/sub"}, "plain/sub", "-", "plain/sub"),
	)

	It("treats a .git file (linked worktree) as the git root", func() {
		root := tree("wt/pkg")
		Expect(os.WriteFile(filepath.Join(root, "wt", ".git"), []byte("gitdir: /elsewhere\n"), 0o644)).To(Succeed())
		dirs, err := workspace.Resolve(filepath.Join(root, "wt", "pkg"))
		Expect(err).NotTo(HaveOccurred())
		Expect(dirs.GitRoot).To(Equal(filepath.Join(root, "wt")))
		Expect(dirs.State("run-1.json")).To(Equal(filepath.Join(root, "wt", ".gavel", "run-1.json")))
	})

	It("fails on a missing directory", func() {
		_, err := workspace.Resolve(filepath.Join(tree(), "missing"))
		Expect(err).To(MatchError(ContainSubstring("missing")))
	})

	It("rejects repository state that escapes through a symlink", func() {
		root := tree("repo/.git", "repo/pkg", "outside")
		Expect(os.Symlink(filepath.Join(root, "outside"), filepath.Join(root, "repo", ".gavel"))).To(Succeed())
		_, err := workspace.Resolve(filepath.Join(root, "repo", "pkg"))
		Expect(err).To(MatchError(ContainSubstring("path escapes")))
	})

	It("accepts a state directory linked within the repository", func() {
		root := tree("repo/.git", "repo/pkg", "repo/state")
		Expect(os.Symlink("state", filepath.Join(root, "repo", ".gavel"))).To(Succeed())
		dirs, err := workspace.Resolve(filepath.Join(root, "repo", "pkg"))
		Expect(err).NotTo(HaveOccurred())
		Expect(dirs.StateRoot).To(Equal(filepath.Join(root, "repo")))
	})
})

var _ = Describe("From", func() {
	It("returns ErrNoDirs for a context that was never given Dirs", func() {
		_, err := workspace.From(context.Background())
		Expect(err).To(MatchError(workspace.ErrNoDirs))
	})

	It("returns the Dirs stored by With", func() {
		want := workspace.Dirs{Invocation: "/a", GitRoot: "/a", StateRoot: "/a"}
		got, err := workspace.From(workspace.With(context.Background(), want))
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(want))
	})
})
