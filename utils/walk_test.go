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
