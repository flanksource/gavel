package outline

import (
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/testrunner/parsers"
)

var _ = Describe("Markdown fixture outline", func() {
	It("adds executable fixtures without running setup or commands", func() {
		workDir := GinkgoT().TempDir()
		sentinel := filepath.Join(workDir, "executed")
		fixturePath := filepath.Join(workDir, "smoke.fixture.md")
		content := `---
build: "touch ` + sentinel + `"
---

# Smoke

## Run tests

` + "```yaml test" + `
paths: [./pkg]
` + "```" + `

## Command

` + "```bash" + `
touch ` + sentinel + `
` + "```" + `
`
		Expect(os.WriteFile(fixturePath, []byte(content), 0o600)).To(Succeed())

		entries, err := collectFixtureTests(workDir, nil, []string{"**/*.fixture.md"})
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(2))
		Expect(entries[0].Framework).To(Equal(parsers.Fixture))
		Expect(entries[0].File).To(Equal("smoke.fixture.md"))
		Expect(entries[0].Labels).NotTo(BeEmpty())
		_, err = os.Stat(sentinel)
		Expect(os.IsNotExist(err)).To(BeTrue())
	})

	It("honors positional path filters", func() {
		workDir := GinkgoT().TempDir()
		for _, dir := range []string{"included", "excluded"} {
			Expect(os.MkdirAll(filepath.Join(workDir, dir), 0o700)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(workDir, dir, "test.fixture.md"), []byte("# Suite\n\n```bash\necho ok\n```\n"), 0o600)).To(Succeed())
		}

		entries, err := collectFixtureTests(workDir, []string{"included"}, []string{"**/*.fixture.md"})
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(1))
		Expect(entries[0].File).To(Equal("included/test.fixture.md"))
	})

	// Scratch worktrees hold a full copy of the repo, so a blind glob reports
	// every fixture twice. They are only covered by the user's global gitignore,
	// so pruning has to key off the nested checkout, not the repo .gitignore.
	It("skips fixtures under gitignored dirs and nested checkouts", func() {
		workDir := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(workDir, ".git"), 0o700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, ".gitignore"), []byte("node_modules/\n"), 0o600)).To(Succeed())

		body := []byte("# Suite\n\n```bash\necho ok\n```\n")
		for _, dir := range []string{"examples", ".runtime/worktrees/copy/examples", "node_modules/pkg"} {
			Expect(os.MkdirAll(filepath.Join(workDir, dir), 0o700)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(workDir, dir, "a.fixture.md"), body, 0o600)).To(Succeed())
		}
		// The worktree is a checkout in its own right, not a gitignored path.
		Expect(os.MkdirAll(filepath.Join(workDir, ".runtime/worktrees/copy/.git"), 0o700)).To(Succeed())

		files, err := discoverFixtureFiles(workDir, nil, []string{"**/*.fixture.md"})
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(HaveLen(1))
		Expect(relativeTo(files[0], workDir)).To(Equal("examples/a.fixture.md"))
	})

	// .gavel.yaml globs are hand-written, so "./pkg/x/*.md" is as common as
	// "pkg/x/*.md". Matching against workDir-relative paths only works if the
	// pattern is normalized the way filepath.Join used to normalize it.
	It("treats equivalent relative patterns as the same glob", func() {
		workDir := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(workDir, "pkg", "formula"), 0o700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workDir, "pkg", "formula", "rates.md"), []byte("# Suite\n\n```bash\necho ok\n```\n"), 0o600)).To(Succeed())

		for _, pattern := range []string{"pkg/formula/**/*.md", "./pkg/formula/**/*.md", "pkg/./formula/**/*.md"} {
			files, err := discoverFixtureFiles(workDir, nil, []string{pattern})
			Expect(err).NotTo(HaveOccurred(), pattern)
			Expect(files).To(HaveLen(1), pattern)
			Expect(relativeTo(files[0], workDir)).To(Equal("pkg/formula/rates.md"), pattern)
		}
	})

	It("resolves relative fixture patterns that escape the workdir", func() {
		root := GinkgoT().TempDir()
		workDir := filepath.Join(root, "repo")
		Expect(os.MkdirAll(filepath.Join(root, "shared"), 0o700)).To(Succeed())
		Expect(os.MkdirAll(workDir, 0o700)).To(Succeed())
		external := filepath.Join(root, "shared", "a.fixture.md")
		Expect(os.WriteFile(external, []byte("# Suite\n\n```bash\necho ok\n```\n"), 0o600)).To(Succeed())

		files, err := discoverFixtureFiles(workDir, nil, []string{"../shared/*.fixture.md"})
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(ConsistOf(external))
	})

	It("resolves absolute fixture patterns outside the workdir", func() {
		workDir := GinkgoT().TempDir()
		outside := GinkgoT().TempDir()
		external := filepath.Join(outside, "external.fixture.md")
		Expect(os.WriteFile(external, []byte("# Suite\n\n```bash\necho ok\n```\n"), 0o600)).To(Succeed())

		files, err := discoverFixtureFiles(workDir, nil, []string{filepath.Join(outside, "*.fixture.md")})
		Expect(err).NotTo(HaveOccurred())
		Expect(files).To(ConsistOf(external))
	})
})
