package utils

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Suite")
}

func setupGitRepo(root string) {
	os.MkdirAll(filepath.Join(root, ".git", "info"), 0755)
}

func collectPaths(root string, allowList ...string) ([]string, error) {
	var paths []string
	err := WalkGitIgnored(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if rel != "." {
			paths = append(paths, filepath.ToSlash(rel))
		}
		return nil
	}, allowList...)
	return paths, err
}

func collectBoundedPaths(root string, allowList ...string) ([]string, error) {
	var paths []string
	err := WalkGitIgnoredBounded(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if rel != "." {
			paths = append(paths, filepath.ToSlash(rel))
		}
		return nil
	}, allowList...)
	return paths, err
}

var _ = Describe("WalkGitIgnored", func() {
	var root string

	BeforeEach(func() {
		root = GinkgoT().TempDir()
	})

	It("skips gitignored files and directories", func() {
		setupGitRepo(root)
		os.WriteFile(filepath.Join(root, ".gitignore"), []byte("vendor/\n*.log\n"), 0644)
		os.MkdirAll(filepath.Join(root, "vendor", "lib"), 0755)
		os.WriteFile(filepath.Join(root, "vendor", "lib", "dep.go"), nil, 0644)
		os.WriteFile(filepath.Join(root, "main.go"), nil, 0644)
		os.WriteFile(filepath.Join(root, "debug.log"), nil, 0644)

		paths, err := collectPaths(root)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(ContainElement("main.go"))
		Expect(paths).NotTo(ContainElement("vendor"))
		Expect(paths).NotTo(ContainElement("vendor/lib"))
		Expect(paths).NotTo(ContainElement("vendor/lib/dep.go"))
		Expect(paths).NotTo(ContainElement("debug.log"))
	})

	It("handles nested gitignore files", func() {
		setupGitRepo(root)
		os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.log\n"), 0644)
		os.MkdirAll(filepath.Join(root, "sub", "build"), 0755)
		os.WriteFile(filepath.Join(root, "sub", ".gitignore"), []byte("build/\n"), 0644)
		os.WriteFile(filepath.Join(root, "sub", "code.go"), nil, 0644)
		os.WriteFile(filepath.Join(root, "sub", "build", "out.bin"), nil, 0644)

		paths, err := collectPaths(root)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(ContainElement("sub/code.go"))
		Expect(paths).NotTo(ContainElement("sub/build"))
		Expect(paths).NotTo(ContainElement("sub/build/out.bin"))
	})

	It("falls back to walking everything when no .git present", func() {
		os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.log\n"), 0644)
		os.WriteFile(filepath.Join(root, "main.go"), nil, 0644)
		os.WriteFile(filepath.Join(root, "debug.log"), nil, 0644)

		paths, err := collectPaths(root)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(ContainElement("main.go"))
		Expect(paths).To(ContainElement("debug.log"))
	})

	It("respects .git/info/exclude", func() {
		setupGitRepo(root)
		os.WriteFile(filepath.Join(root, ".git", "info", "exclude"), []byte("secret/\n"), 0644)
		os.MkdirAll(filepath.Join(root, "secret"), 0755)
		os.WriteFile(filepath.Join(root, "secret", "key.pem"), nil, 0644)
		os.WriteFile(filepath.Join(root, "main.go"), nil, 0644)

		paths, err := collectPaths(root)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(ContainElement("main.go"))
		Expect(paths).NotTo(ContainElement("secret"))
		Expect(paths).NotTo(ContainElement("secret/key.pem"))
	})

	It("allowList overrides gitignore", func() {
		setupGitRepo(root)
		os.WriteFile(filepath.Join(root, ".gitignore"), []byte(".todos/\n.codex/\n"), 0644)
		os.MkdirAll(filepath.Join(root, ".todos"), 0755)
		os.WriteFile(filepath.Join(root, ".todos", "task.md"), nil, 0644)
		os.MkdirAll(filepath.Join(root, ".codex"), 0755)
		os.WriteFile(filepath.Join(root, ".codex", "data.json"), nil, 0644)

		paths, err := collectPaths(root, ".todos", ".codex")
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(ContainElement(".todos"))
		Expect(paths).To(ContainElement(".todos/task.md"))
		Expect(paths).To(ContainElement(".codex"))
		Expect(paths).To(ContainElement(".codex/data.json"))
	})

	It("always skips .git directory", func() {
		setupGitRepo(root)
		os.WriteFile(filepath.Join(root, "main.go"), nil, 0644)

		paths, err := collectPaths(root)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(ContainElement("main.go"))
		Expect(paths).NotTo(ContainElement(".git"))
	})
})

var _ = Describe("WalkGitIgnoredBounded", func() {
	var root string

	BeforeEach(func() {
		root = GinkgoT().TempDir()
		setupGitRepo(root)
	})

	It("still traverses the starting root when it contains go.mod and .git", func() {
		Expect(os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/root\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "pkg"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "pkg", "pkg_test.go"), nil, 0o644)).To(Succeed())

		paths, err := collectBoundedPaths(root)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(ContainElement("go.mod"))
		Expect(paths).To(ContainElement("pkg"))
		Expect(paths).To(ContainElement("pkg/pkg_test.go"))
	})

	It("skips nested module subtrees", func() {
		Expect(os.MkdirAll(filepath.Join(root, "visible"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "visible", "keep_test.go"), nil, 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "nested-module", "pkg"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "nested-module", "go.mod"), []byte("module example.com/nested\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "nested-module", "pkg", "skip_test.go"), nil, 0o644)).To(Succeed())

		paths, err := collectBoundedPaths(root)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(ContainElement("visible"))
		Expect(paths).To(ContainElement("visible/keep_test.go"))
		Expect(paths).NotTo(ContainElement("nested-module"))
		Expect(paths).NotTo(ContainElement("nested-module/pkg"))
		Expect(paths).NotTo(ContainElement("nested-module/pkg/skip_test.go"))
	})

	It("skips nested git repo subtrees", func() {
		Expect(os.MkdirAll(filepath.Join(root, "visible"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "visible", "keep_test.go"), nil, 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "nested-repo", ".git"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "nested-repo", "pkg"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "nested-repo", "pkg", "skip_test.go"), nil, 0o644)).To(Succeed())

		paths, err := collectBoundedPaths(root)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(ContainElement("visible"))
		Expect(paths).To(ContainElement("visible/keep_test.go"))
		Expect(paths).NotTo(ContainElement("nested-repo"))
		Expect(paths).NotTo(ContainElement("nested-repo/pkg"))
		Expect(paths).NotTo(ContainElement("nested-repo/pkg/skip_test.go"))
	})

	// `git worktree add` writes .git as a *file* holding a `gitdir:` pointer, so
	// a boundary check that requires a directory walks straight into the copy.
	It("skips linked worktrees whose .git is a file", func() {
		Expect(os.MkdirAll(filepath.Join(root, "visible"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "visible", "keep_test.go"), nil, 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "worktree", "pkg"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "worktree", ".git"), []byte("gitdir: /elsewhere/.git/worktrees/wt\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "worktree", "pkg", "skip_test.go"), nil, 0o644)).To(Succeed())

		paths, err := collectBoundedPaths(root)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(ContainElement("visible/keep_test.go"))
		Expect(paths).NotTo(ContainElement("worktree"))
		Expect(paths).NotTo(ContainElement("worktree/pkg/skip_test.go"))
	})

	It("skips descendant directories that contain both go.mod and .git", func() {
		Expect(os.MkdirAll(filepath.Join(root, "nested-both", ".git"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "nested-both", "pkg"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "nested-both", "go.mod"), []byte("module example.com/both\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "nested-both", "pkg", "skip_test.go"), nil, 0o644)).To(Succeed())

		paths, err := collectBoundedPaths(root)
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).NotTo(ContainElement("nested-both"))
		Expect(paths).NotTo(ContainElement("nested-both/pkg"))
		Expect(paths).NotTo(ContainElement("nested-both/pkg/skip_test.go"))
	})
})

var _ = Describe("FindNearestProjectRoot", func() {
	var root string

	BeforeEach(func() {
		root = GinkgoT().TempDir()
		setupGitRepo(root)
	})

	It("finds the nearest marker inside the git root", func() {
		Expect(os.MkdirAll(filepath.Join(root, "frontend", "src"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "frontend", "package.json"), []byte("{}"), 0o644)).To(Succeed())

		got := FindNearestProjectRoot(filepath.Join(root, "frontend", "src"), []string{"package.json"})
		Expect(got).To(Equal(filepath.Join(root, "frontend")))
	})

	It("returns empty when no marker is found within the git root", func() {
		Expect(os.MkdirAll(filepath.Join(root, "src"), 0o755)).To(Succeed())
		got := FindNearestProjectRoot(filepath.Join(root, "src"), []string{"go.mod"})
		Expect(got).To(BeEmpty())
	})

	It("stops at the git root boundary", func() {
		parent := filepath.Dir(root)
		Expect(os.WriteFile(filepath.Join(parent, "go.mod"), []byte("module parent\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "pkg"), 0o755)).To(Succeed())

		got := FindNearestProjectRoot(filepath.Join(root, "pkg"), []string{"go.mod"})
		Expect(got).To(BeEmpty())
	})

	It("matches any of the supplied markers (first ancestor wins)", func() {
		Expect(os.MkdirAll(filepath.Join(root, "app", "src"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "app", "pyproject.toml"), []byte(""), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "ruff.toml"), []byte(""), 0o644)).To(Succeed())

		got := FindNearestProjectRoot(filepath.Join(root, "app", "src"), []string{"ruff.toml", "pyproject.toml"})
		Expect(got).To(Equal(filepath.Join(root, "app")))
	})

	It("treats the git root itself as a candidate", func() {
		Expect(os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/x\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "pkg"), 0o755)).To(Succeed())

		got := FindNearestProjectRoot(filepath.Join(root, "pkg"), []string{"go.mod"})
		Expect(got).To(Equal(root))
	})

	It("returns empty when markers list is empty", func() {
		Expect(FindNearestProjectRoot(root, nil)).To(BeEmpty())
		Expect(FindNearestProjectRoot(root, []string{})).To(BeEmpty())
	})
})

var _ = Describe("IsWithin", func() {
	It("returns true when path equals root", func() {
		Expect(IsWithin("/repo", "/repo")).To(BeTrue())
	})
	It("returns true for a nested path", func() {
		Expect(IsWithin("/repo/pkg/a", "/repo")).To(BeTrue())
	})
	It("returns false for a sibling path", func() {
		Expect(IsWithin("/other/pkg", "/repo")).To(BeFalse())
	})
	It("returns false when a ../ traversal escapes root", func() {
		Expect(IsWithin("/repo/../outside", "/repo")).To(BeFalse())
	})
	It("returns false for a parent of root", func() {
		Expect(IsWithin("/", "/repo")).To(BeFalse())
	})
})

var _ = Describe("FindGitRoot", func() {
	var root string

	BeforeEach(func() {
		root = GinkgoT().TempDir()
	})

	It("finds the root when .git is a directory", func() {
		Expect(os.MkdirAll(filepath.Join(root, ".git"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "pkg"), 0o755)).To(Succeed())
		Expect(FindGitRoot(filepath.Join(root, "pkg"))).To(Equal(root))
	})

	It("finds the root when .git is a file (worktree / submodule gitdir pointer)", func() {
		// Linked worktrees and submodules record .git as a file containing a
		// `gitdir:` line, not a directory. FindGitRoot must still resolve it.
		Expect(os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: /elsewhere/.git/worktrees/wt\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "pkg"), 0o755)).To(Succeed())
		Expect(FindGitRoot(filepath.Join(root, "pkg"))).To(Equal(root))
	})

	It("returns empty when no .git is present", func() {
		Expect(FindGitRoot(root)).To(BeEmpty())
	})
})

var _ = Describe("GitRoot", func() {
	var root string

	BeforeEach(func() {
		root = GinkgoT().TempDir()
	})

	It("lifts a subdirectory to the root of its working tree", func() {
		// An agent run anchored on a subdirectory records its edits in a
		// namespace git never uses, so nothing it changed can be attributed.
		Expect(os.MkdirAll(filepath.Join(root, ".git"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "apps", "storybook"), 0o755)).To(Succeed())
		Expect(GitRoot(filepath.Join(root, "apps", "storybook"))).To(Equal(root))
	})

	It("keeps a directory that is outside any repository", func() {
		Expect(GitRoot(root)).To(Equal(root))
	})
})

var _ = Describe("FindAllProjectRoots", func() {
	var root string

	BeforeEach(func() {
		root = GinkgoT().TempDir()
		setupGitRepo(root)
	})

	It("finds every outermost package.json under the tree", func() {
		Expect(os.MkdirAll(filepath.Join(root, "frontend"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "frontend", "package.json"), []byte("{}"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "tools"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "tools", "package.json"), []byte("{}"), 0o644)).To(Succeed())

		got := FindAllProjectRoots(root, []string{"package.json"})
		Expect(got).To(ConsistOf(
			filepath.Join(root, "frontend"),
			filepath.Join(root, "tools"),
		))
	})

	It("finds nested project roots while pruning dependencies", func() {
		Expect(os.MkdirAll(filepath.Join(root, "app", "packages", "inner"), 0o755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "app", "node_modules", "dependency"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "app", "package.json"), []byte("{}"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "app", "packages", "inner", "package.json"), []byte("{}"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "app", "node_modules", "dependency", "package.json"), []byte("{}"), 0o644)).To(Succeed())

		got := FindAllProjectRoots(root, []string{"package.json"})
		Expect(got).To(Equal([]string{
			filepath.Join(root, "app"),
			filepath.Join(root, "app", "packages", "inner"),
		}))
	})

	It("does not cross into nested git repositories", func() {
		Expect(os.MkdirAll(filepath.Join(root, "nested", ".git"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "nested", "package.json"), []byte("{}"), 0o644)).To(Succeed())

		got := FindAllProjectRoots(root, []string{"package.json"})
		Expect(got).To(BeEmpty())
	})

	It("returns the root itself when it carries the marker", func() {
		Expect(os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "pkg"), 0o755)).To(Succeed())

		got := FindAllProjectRoots(root, []string{"go.mod"})
		Expect(got).To(ConsistOf(root))
	})

	It("returns empty slice when no marker present", func() {
		Expect(os.MkdirAll(filepath.Join(root, "src"), 0o755)).To(Succeed())
		got := FindAllProjectRoots(root, []string{"go.mod"})
		Expect(got).To(BeEmpty())
	})

	It("returns nil when markers list is empty", func() {
		Expect(FindAllProjectRoots(root, nil)).To(BeNil())
	})
})

var _ = Describe("FilterGitIgnored", func() {
	var root string

	BeforeEach(func() {
		root = GinkgoT().TempDir()
	})

	It("filters paths matching gitignore patterns", func() {
		setupGitRepo(root)
		os.WriteFile(filepath.Join(root, ".gitignore"), []byte("vendor/\n*.log\n"), 0644)

		paths := []string{
			filepath.Join(root, "main.go"),
			filepath.Join(root, "vendor", "dep.go"),
			filepath.Join(root, "debug.log"),
		}
		result := FilterGitIgnored(paths, root)
		Expect(result).To(ConsistOf(filepath.Join(root, "main.go")))
	})

	It("returns all paths when no git root exists", func() {
		paths := []string{
			filepath.Join(root, "a.go"),
			filepath.Join(root, "b.go"),
		}
		result := FilterGitIgnored(paths, root)
		Expect(result).To(Equal(paths))
	})

	It("handles nested gitignore files", func() {
		setupGitRepo(root)
		os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.log\n"), 0644)
		os.MkdirAll(filepath.Join(root, "sub"), 0755)
		os.WriteFile(filepath.Join(root, "sub", ".gitignore"), []byte("build/\n"), 0644)

		paths := []string{
			filepath.Join(root, "main.go"),
			filepath.Join(root, "app.log"),
			filepath.Join(root, "sub", "code.go"),
			filepath.Join(root, "sub", "build", "out.bin"),
		}
		result := FilterGitIgnored(paths, root)
		Expect(result).To(ConsistOf(
			filepath.Join(root, "main.go"),
			filepath.Join(root, "sub", "code.go"),
		))
	})

	It("respects .git/info/exclude", func() {
		setupGitRepo(root)
		os.WriteFile(filepath.Join(root, ".git", "info", "exclude"), []byte("secret/\n"), 0644)

		paths := []string{
			filepath.Join(root, "main.go"),
			filepath.Join(root, "secret", "key.pem"),
		}
		result := FilterGitIgnored(paths, root)
		Expect(result).To(ConsistOf(filepath.Join(root, "main.go")))
	})

	It("returns empty slice for all-ignored input", func() {
		setupGitRepo(root)
		os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.log\n"), 0644)

		paths := []string{
			filepath.Join(root, "debug.log"),
			filepath.Join(root, "error.log"),
		}
		result := FilterGitIgnored(paths, root)
		Expect(result).To(BeEmpty())
	})
})

var _ = Describe("PartitionGitIgnored", func() {
	var root string

	BeforeEach(func() {
		root = GinkgoT().TempDir()
	})

	It("splits kept and ignored, honoring !-negation", func() {
		setupGitRepo(root)
		os.WriteFile(filepath.Join(root, ".gitignore"), []byte("dist/*\n!dist/keep.js\n"), 0644)

		paths := []string{
			filepath.Join(root, "src", "main.go"),
			filepath.Join(root, "dist", "bundle.js"),
			filepath.Join(root, "dist", "keep.js"),
		}
		kept, ignored := PartitionGitIgnored(paths, root)
		Expect(kept).To(ConsistOf(
			filepath.Join(root, "src", "main.go"),
			filepath.Join(root, "dist", "keep.js"),
		))
		Expect(ignored).To(ConsistOf(filepath.Join(root, "dist", "bundle.js")))
	})

	It("keeps everything and reports nothing ignored without a git root", func() {
		paths := []string{filepath.Join(root, "a.go"), filepath.Join(root, "b.go")}
		kept, ignored := PartitionGitIgnored(paths, root)
		Expect(kept).To(Equal(paths))
		Expect(ignored).To(BeEmpty())
	})
})

var _ = Describe("GlobWalkGitIgnored", func() {
	var root string

	BeforeEach(func() {
		root = GinkgoT().TempDir()
		setupGitRepo(root)
	})

	collect := func(patterns, extraIgnore []string) []string {
		var matched []string
		err := GlobWalkGitIgnored(root, patterns, extraIgnore, func(rel string, d fs.DirEntry) error {
			matched = append(matched, rel)
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
		return matched
	}

	It("matches files at top level and any nested depth", func() {
		Expect(os.WriteFile(filepath.Join(root, "main.go"), nil, 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "pkg", "inner"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "pkg", "inner", "deep.go"), nil, 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "pkg", "notes.md"), nil, 0o644)).To(Succeed())

		Expect(collect([]string{"**/*.go"}, nil)).To(ConsistOf("main.go", "pkg/inner/deep.go"))
	})

	It("does not descend into gitignored directories", func() {
		Expect(os.WriteFile(filepath.Join(root, ".gitignore"), []byte("node_modules/\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "node_modules", "dep"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "node_modules", "dep", "vendored.py"), nil, 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "app.py"), nil, 0o644)).To(Succeed())

		Expect(collect([]string{"**/*.py"}, nil)).To(ConsistOf("app.py"))
	})

	It("prunes directories matched by extraIgnore (.gavel gitignore)", func() {
		Expect(os.MkdirAll(filepath.Join(root, "dist"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "dist", "bundle.js"), nil, 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "index.js"), nil, 0o644)).To(Succeed())

		Expect(collect([]string{"**/*.js"}, []string{"dist/"})).To(ConsistOf("index.js"))
	})
})

var _ = Describe("GlobFilesBounded", func() {
	var root string

	BeforeEach(func() {
		root = GinkgoT().TempDir()
		setupGitRepo(root)
	})

	rel := func(files []string) []string {
		out := make([]string, 0, len(files))
		for _, f := range files {
			r, err := filepath.Rel(root, f)
			Expect(err).NotTo(HaveOccurred())
			out = append(out, filepath.ToSlash(r))
		}
		return out
	}

	// Hand-written .gavel.yaml globs use "./pkg/x/*.md" as often as "pkg/x/*.md".
	// Matching against root-relative paths only works if the pattern is cleaned
	// the way filepath.Join used to clean it.
	It("treats equivalent relative patterns as the same glob", func() {
		Expect(os.MkdirAll(filepath.Join(root, "pkg", "formula"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "pkg", "formula", "rates.md"), nil, 0o644)).To(Succeed())

		for _, pattern := range []string{"pkg/formula/**/*.md", "./pkg/formula/**/*.md", "pkg/./formula/**/*.md"} {
			files, err := GlobFilesBounded(root, []string{pattern})
			Expect(err).NotTo(HaveOccurred(), pattern)
			Expect(rel(files)).To(ConsistOf("pkg/formula/rates.md"), pattern)
		}
	})

	// Scratch worktrees hold a full copy of the repo, so a blind glob reports
	// every match twice. They are often covered only by the user's global
	// gitignore, so pruning has to key off the nested checkout.
	It("skips gitignored dirs and nested checkouts", func() {
		Expect(os.WriteFile(filepath.Join(root, ".gitignore"), []byte("node_modules/\n"), 0o644)).To(Succeed())
		for _, dir := range []string{"examples", ".runtime/worktrees/copy/examples", "node_modules/pkg"} {
			Expect(os.MkdirAll(filepath.Join(root, dir), 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(root, dir, "a.fixture.md"), nil, 0o644)).To(Succeed())
		}
		Expect(os.MkdirAll(filepath.Join(root, ".runtime/worktrees/copy/.git"), 0o755)).To(Succeed())

		files, err := GlobFilesBounded(root, []string{"**/*.fixture.md"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rel(files)).To(ConsistOf("examples/a.fixture.md"))
	})

	It("resolves patterns that escape root, which the bounded walk cannot reach", func() {
		outer := filepath.Dir(root)
		shared := filepath.Join(outer, "shared-"+filepath.Base(root))
		Expect(os.MkdirAll(shared, 0o755)).To(Succeed())
		DeferCleanup(os.RemoveAll, shared)
		external := filepath.Join(shared, "a.fixture.md")
		Expect(os.WriteFile(external, nil, 0o644)).To(Succeed())

		relative, err := GlobFilesBounded(root, []string{"../" + filepath.Base(shared) + "/*.fixture.md"})
		Expect(err).NotTo(HaveOccurred())
		Expect(relative).To(ConsistOf(external))

		absolute, err := GlobFilesBounded(root, []string{filepath.Join(shared, "*.fixture.md")})
		Expect(err).NotTo(HaveOccurred())
		Expect(absolute).To(ConsistOf(external))
	})

	It("deduplicates files matched by more than one pattern", func() {
		Expect(os.WriteFile(filepath.Join(root, "a.fixture.md"), nil, 0o644)).To(Succeed())

		files, err := GlobFilesBounded(root, []string{"**/*.md", "*.fixture.md"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rel(files)).To(ConsistOf("a.fixture.md"))
	})

	// A silently-skipped bad glob turns a config typo into "no tests found".
	It("fails loudly on an invalid glob", func() {
		_, err := GlobFilesBounded(root, []string{"pkg/[unclosed"})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("pkg/[unclosed"))
	})

	It("returns nothing when given no patterns", func() {
		Expect(os.WriteFile(filepath.Join(root, "a.fixture.md"), nil, 0o644)).To(Succeed())

		files, err := GlobFilesBounded(root, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(BeEmpty())
	})
})

var _ = Describe("AnyGlobMatchGitIgnored", func() {
	var root string

	BeforeEach(func() {
		root = GinkgoT().TempDir()
		setupGitRepo(root)
	})

	It("returns false when the only match is inside a gitignored directory", func() {
		Expect(os.WriteFile(filepath.Join(root, ".gitignore"), []byte("node_modules/\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "node_modules", "dep"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "node_modules", "dep", "vendored.py"), nil, 0o644)).To(Succeed())

		Expect(AnyGlobMatchGitIgnored(root, []string{"**/*.py"}, nil)).To(BeFalse())
	})

	It("returns true when a match exists outside ignored directories", func() {
		Expect(os.MkdirAll(filepath.Join(root, "src"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "src", "app.py"), nil, 0o644)).To(Succeed())

		Expect(AnyGlobMatchGitIgnored(root, []string{"**/*.py"}, nil)).To(BeTrue())
	})

	It("returns false when extraIgnore hides the only match", func() {
		Expect(os.MkdirAll(filepath.Join(root, "dist"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "dist", "bundle.js"), nil, 0o644)).To(Succeed())

		Expect(AnyGlobMatchGitIgnored(root, []string{"**/*.js"}, []string{"dist/"})).To(BeFalse())
	})

	It("returns false when no file matches the patterns", func() {
		Expect(os.WriteFile(filepath.Join(root, "README.md"), nil, 0o644)).To(Succeed())
		Expect(AnyGlobMatchGitIgnored(root, []string{"**/*.go"}, nil)).To(BeFalse())
	})

	It("returns false for empty patterns", func() {
		Expect(os.WriteFile(filepath.Join(root, "main.go"), nil, 0o644)).To(Succeed())
		Expect(AnyGlobMatchGitIgnored(root, nil, nil)).To(BeFalse())
	})
})
