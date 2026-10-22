package changeset_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/flanksource/gavel/git/changeset"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestChangeset(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Changeset")
}

func git(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %v: %s", args, out)
	return string(out)
}

func write(dir, name, content string) {
	Expect(os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600)).To(Succeed())
}

func repository() string {
	dir := GinkgoT().TempDir()
	git(dir, "init", "-q", "-b", "main")
	git(dir, "config", "user.name", "Test")
	git(dir, "config", "user.email", "test@example.com")
	git(dir, "config", "commit.gpgsign", "false")
	write(dir, "tracked.txt", "original\n")
	git(dir, "add", "tracked.txt")
	git(dir, "commit", "-qm", "initial")
	return dir
}

func localEnvironment() changeset.Environment {
	return changeset.Environment{Lookup: func(string) string { return "" }, ReadFile: os.ReadFile}
}

var _ = Describe("Shared changesets", func() {
	It("accepts known files without Git and resolves aliases of repository paths", func() {
		dir := GinkgoT().TempDir()
		selected, err := changeset.Resolve(context.Background(), changeset.Options{Dir: dir, Files: []string{"known.go"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(selected.Paths(dir)).To(Equal([]string{"known.go"}))
		repo := repository()
		alias := filepath.Join(dir, "alias")
		Expect(os.Symlink(repo, alias)).To(Succeed())
		selected, err = changeset.Resolve(context.Background(), changeset.Options{Dir: alias, Files: []string{"tracked.txt"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(selected.Paths(alias)).To(Equal([]string{"tracked.txt"}))
	})
	It("shares one resolution per repository and refreshes on a new run", func() {
		dir := repository()
		selection := changeset.NewSelection(changeset.Options{Dir: dir, Changed: true, Environment: localEnvironment()})
		first, err := selection.Resolve(context.Background(), dir)
		Expect(err).NotTo(HaveOccurred())
		write(dir, "later.txt", "new\n")
		same, err := selection.Resolve(context.Background(), dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(same).To(BeIdenticalTo(first))
		Expect(same.Changes).To(BeEmpty())
		fresh, err := selection.Fresh().Resolve(context.Background(), dir)
		Expect(err).NotTo(HaveOccurred())
		Expect(fresh.Paths(dir)).To(Equal([]string{"later.txt"}))
	})
	It("distinguishes unrestricted selection from a valid empty selection", func() {
		dir := repository()
		all, err := changeset.Resolve(context.Background(), changeset.Options{Dir: dir, Environment: localEnvironment()})
		Expect(err).NotTo(HaveOccurred())
		Expect(all).To(BeNil())
		empty, err := changeset.Resolve(context.Background(), changeset.Options{Dir: dir, Changed: true, Environment: localEnvironment()})
		Expect(err).NotTo(HaveOccurred())
		Expect(empty).NotTo(BeNil())
		Expect(empty.Changes).To(BeEmpty())
	})
	It("selects local edits without including committed branch changes", func() {
		dir := repository()
		write(dir, "committed.txt", "committed\n")
		git(dir, "add", "committed.txt")
		git(dir, "commit", "-qm", "branch")
		write(dir, "tracked.txt", "staged\n")
		git(dir, "add", "tracked.txt")
		write(dir, "tracked.txt", "unstaged\n")
		write(dir, "untracked.txt", "new\n")
		write(dir, ".gitignore", "ignored.txt\n")
		write(dir, "ignored.txt", "ignored\n")
		selected, err := changeset.Resolve(context.Background(), changeset.Options{Dir: dir, Changed: true, Environment: localEnvironment()})
		Expect(err).NotTo(HaveOccurred())
		Expect(selected.Paths(dir)).To(Equal([]string{".gitignore", "tracked.txt", "untracked.txt"}))
	})
	It("selects staged files and leaves their working-tree contents intact", func() {
		dir := repository()
		write(dir, "tracked.txt", "staged\n")
		git(dir, "add", "tracked.txt")
		write(dir, "tracked.txt", "unstaged\n")
		write(dir, "untracked.txt", "new\n")
		selected, err := changeset.Resolve(context.Background(), changeset.Options{Dir: dir, Since: "staged", Environment: localEnvironment()})
		Expect(err).NotTo(HaveOccurred())
		Expect(selected.Paths(dir)).To(Equal([]string{"tracked.txt"}))
		Expect(os.ReadFile(filepath.Join(dir, "tracked.txt"))).To(Equal([]byte("unstaged\n")))
	})
	It("uses the merge-base of a diverged ref and unions explicit files", func() {
		dir := repository()
		git(dir, "checkout", "-qb", "feature")
		write(dir, "feature.txt", "feature\n")
		git(dir, "add", "feature.txt")
		git(dir, "commit", "-qm", "feature")
		git(dir, "checkout", "-q", "main")
		write(dir, "target.txt", "target\n")
		git(dir, "add", "target.txt")
		git(dir, "commit", "-qm", "target")
		git(dir, "checkout", "-q", "feature")
		selected, err := changeset.Resolve(context.Background(), changeset.Options{Dir: dir, Since: "main", Files: []string{"explicit.txt"}, Environment: localEnvironment()})
		Expect(err).NotTo(HaveOccurred())
		Expect(selected.Paths(dir)).To(Equal([]string{"explicit.txt", "feature.txt"}))
	})
	It("preserves rename paths, deleted files and literal whitespace", func() {
		dir := repository()
		name := " renamed\tfile\n.txt "
		git(dir, "mv", "tracked.txt", name)
		selected, err := changeset.Resolve(context.Background(), changeset.Options{Dir: dir, Changed: true, Environment: localEnvironment()})
		Expect(err).NotTo(HaveOccurred())
		Expect(selected.Changes).To(Equal([]changeset.Change{{Path: name, PreviousPath: "tracked.txt", Status: "renamed"}}))
		Expect(selected.CurrentPaths()).To(Equal([]string{name}))
		git(dir, "commit", "-qm", "rename")
		Expect(os.Remove(filepath.Join(dir, name))).To(Succeed())
		selected, err = changeset.Resolve(context.Background(), changeset.Options{Dir: dir, Changed: true, Environment: localEnvironment()})
		Expect(err).NotTo(HaveOccurred())
		Expect(selected.Changes).To(Equal([]changeset.Change{{Path: name, Status: "deleted"}}))
		Expect(selected.CurrentPaths()).To(BeEmpty())
	})
})
