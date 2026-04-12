package changegraph_test

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/flanksource/gavel/internal/changegraph"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// initRepo creates a fresh git repo at dir with an initial commit. It sets
// user.name/user.email locally so CI environments without global git config
// still work.
func initRepo(dir string) {
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred(), "git %v failed: %s", args, out)
	}
	run("init", "--quiet", "-b", "main")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	run("config", "commit.gpgsign", "false")

	// Seed with an initial commit so diff and merge-base have something to
	// compare against.
	Expect(os.WriteFile(filepath.Join(dir, "seed.txt"), []byte("seed\n"), 0o644)).To(Succeed())
	run("add", "seed.txt")
	run("commit", "--quiet", "-m", "seed")
}

var _ = Describe("ComputeFileSet", func() {
	var repo string

	BeforeEach(func() {
		repo = GinkgoT().TempDir()
		initRepo(repo)
	})

	It("includes unstaged modifications", func() {
		Expect(os.WriteFile(filepath.Join(repo, "seed.txt"), []byte("changed\n"), 0o644)).To(Succeed())

		fs, err := changegraph.ComputeFileSet(repo, changegraph.DiffOptions{IncludeUnstaged: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(fs.Sorted()).To(ConsistOf("seed.txt"))
	})

	It("includes staged changes when requested", func() {
		Expect(os.WriteFile(filepath.Join(repo, "new.txt"), []byte("x\n"), 0o644)).To(Succeed())
		cmd := exec.Command("git", "add", "new.txt")
		cmd.Dir = repo
		Expect(cmd.Run()).To(Succeed())

		fs, err := changegraph.ComputeFileSet(repo, changegraph.DiffOptions{IncludeStaged: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(fs.Sorted()).To(ConsistOf("new.txt"))
	})

	It("includes untracked files when requested", func() {
		Expect(os.WriteFile(filepath.Join(repo, "untracked.txt"), []byte("x\n"), 0o644)).To(Succeed())

		fs, err := changegraph.ComputeFileSet(repo, changegraph.DiffOptions{IncludeUntracked: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(fs.Sorted()).To(ConsistOf("untracked.txt"))
	})

	It("unions multiple sources", func() {
		// Staged
		Expect(os.WriteFile(filepath.Join(repo, "a.txt"), []byte("a\n"), 0o644)).To(Succeed())
		addCmd := exec.Command("git", "add", "a.txt")
		addCmd.Dir = repo
		Expect(addCmd.Run()).To(Succeed())
		// Unstaged modification on seed.txt
		Expect(os.WriteFile(filepath.Join(repo, "seed.txt"), []byte("mod\n"), 0o644)).To(Succeed())
		// Untracked
		Expect(os.WriteFile(filepath.Join(repo, "c.txt"), []byte("c\n"), 0o644)).To(Succeed())

		fs, err := changegraph.ComputeFileSet(repo, changegraph.DiffOptions{
			IncludeStaged:    true,
			IncludeUnstaged:  true,
			IncludeUntracked: true,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(fs.Sorted()).To(ConsistOf("a.txt", "seed.txt", "c.txt"))
	})

	It("computes since a ref", func() {
		run := func(args ...string) {
			cmd := exec.Command("git", args...)
			cmd.Dir = repo
			Expect(cmd.Run()).To(Succeed())
		}
		// Make a second commit so HEAD~1 is meaningful.
		Expect(os.WriteFile(filepath.Join(repo, "second.txt"), []byte("s\n"), 0o644)).To(Succeed())
		run("add", "second.txt")
		run("commit", "--quiet", "-m", "second")

		fs, err := changegraph.ComputeFileSet(repo, changegraph.DiffOptions{Since: "HEAD~1"})
		Expect(err).NotTo(HaveOccurred())
		Expect(fs.Sorted()).To(ConsistOf("second.txt"))
	})
})

var _ = Describe("FileSet", func() {
	It("Add normalizes separators", func() {
		fs := changegraph.NewFileSet()
		fs.Add("a\\b\\c.go")
		Expect(fs.Has("a/b/c.go")).To(BeTrue())
	})

	It("Union merges sets", func() {
		a := changegraph.NewFileSet()
		a.Add("x")
		b := changegraph.NewFileSet()
		b.Add("y")
		a.Union(b)
		Expect(a.Sorted()).To(Equal([]string{"x", "y"}))
	})

	It("ignores empty paths", func() {
		fs := changegraph.NewFileSet()
		fs.Add("")
		fs.Add("   ")
		Expect(fs.Len()).To(Equal(0))
	})
})
