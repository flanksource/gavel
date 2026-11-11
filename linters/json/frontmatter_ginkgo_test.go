package jsonlint

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const titleSchema = `{"type":"object","properties":{"title":{"type":"string"}},"required":["title"]}`

var _ = func() bool {
	RegisterBuiltinSchema(BuiltinSchema{
		Name: "test-recipe",
		Matches: func(path string, _ verify.GavelConfig) (bool, error) {
			return strings.HasSuffix(path, ".recipe.md"), nil
		},
		Load: func() (map[string]any, error) {
			return map[string]any{"type": "object", "properties": map[string]any{"serves": map[string]any{"type": "integer"}}}, nil
		},
	})
	return true
}()

var _ = Describe("Markdown frontmatter schemas", func() {
	var root string
	write := func(name, contents string) string {
		path := filepath.Join(root, name)
		Expect(os.MkdirAll(filepath.Dir(path), 0o700)).To(Succeed())
		Expect(os.WriteFile(path, []byte(contents), 0o600)).To(Succeed())
		return path
	}
	lint := func(files ...string) ([]models.Violation, error) {
		adapter := New(root)
		adapter.Files = files
		return adapter.Run(commonsContext.NewContext(context.Background()), nil)
	}
	lines := func(violations []models.Violation) []int {
		var out []int
		for _, v := range violations {
			out = append(out, v.Line)
		}
		return out
	}

	BeforeEach(func() {
		dir, err := filepath.EvalSymlinks(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		root = dir
		Expect(os.Mkdir(filepath.Join(root, ".git"), 0o700)).To(Succeed())
		write("schemas/title.json", titleSchema)
		write(".gavel.yaml", "lint:\n  schemas:\n    \"docs/**/*.md\": schemas/title.json\n")
	})

	It("validates frontmatter against the lint.schemas glob, at its line in the Markdown file", func() {
		page := write("docs/guide.md", "---\nsidebar: true\ntitle: 5\n---\n# Guide\n")

		violations, err := lint(page)

		Expect(err).NotTo(HaveOccurred())
		Expect(lines(violations)).To(Equal([]int{3}))
		Expect(violations[0].File).To(Equal(page))
		Expect(violations[0].Excerpt.Line).To(Equal(3))
		Expect(violations[0].Excerpt.Content).To(ContainSubstring("title: 5"))
	})

	It("reports a missing required field on frontmatter-less pages only when they have frontmatter", func() {
		empty := write("docs/empty.md", "---\n---\n# Empty\n")
		none := write("docs/none.md", "# No frontmatter\n")

		violations, err := lint(empty, none)

		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(ConsistOf(HaveField("File", empty)))
	})

	It("prefers a $schema the frontmatter declares over the glob", func() {
		write("schemas/any.json", `{"type":"object"}`)
		page := write("docs/declared.md", "---\n$schema: ../schemas/any.json\ntitle: 5\n---\n")

		violations, err := lint(page)

		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(BeEmpty())
	})

	It("reports unclosed frontmatter as a syntax error", func() {
		page := write("docs/open.md", "---\ntitle: Open\n# never closed\n")

		violations, err := lint(page)

		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(ConsistOf(SatisfyAll(HaveField("Line", 1), HaveField("Rule.Method", "syntax"))))
	})

	It("anchors a nested .gavel.yaml layer's globs to its own directory", func() {
		write("site/schemas/serves.json", `{"type":"object","required":["serves"]}`)
		write("site/.gavel.yaml", "lint:\n  schemas:\n    \"pages/*.md\": schemas/serves.json\n")
		nested := write("site/pages/menu.md", "---\ntitle: Menu\n---\n")
		outside := write("pages/menu.md", "---\ntitle: Menu\n---\n")

		nestedSchema, err := PathSchema(nested)
		Expect(err).NotTo(HaveOccurred())
		outsideSchema, err := PathSchema(outside)
		Expect(err).NotTo(HaveOccurred())

		Expect(nestedSchema).To(Equal(filepath.Join(root, "site/schemas/serves.json")))
		Expect(outsideSchema).To(BeEmpty())
	})

	It("refuses a file that several globs name", func() {
		write(".gavel.yaml", "lint:\n  schemas:\n    \"docs/**/*.md\": schemas/title.json\n    \"**/guide.md\": schemas/title.json\n")
		page := write("docs/guide.md", "---\ntitle: Guide\n---\n")

		_, err := PathSchema(page)

		Expect(err).To(MatchError(ContainSubstring("matches several lint.schemas globs")))
	})

	It("falls back to the builtin schema of a gavel file kind", func() {
		page := write("menus/soup.recipe.md", "---\nserves: many\n---\n")

		reference, err := PathSchema(page)
		Expect(err).NotTo(HaveOccurred())
		violations, err := lint(page)

		Expect(err).NotTo(HaveOccurred())
		Expect(reference).To(Equal("gavel:test-recipe"))
		Expect(lines(violations)).To(Equal([]int{2}))
	})

	It("gives the editor the glob's schema for a page without frontmatter", func() {
		page := write("docs/new.md", "# New\n")

		schema, err := EditorSchema(context.Background(), nil, page, "")

		Expect(err).NotTo(HaveOccurred())
		Expect(schema).To(HaveKeyWithValue("$defs", HaveKeyWithValue("resource0", HaveKeyWithValue("required", ConsistOf("title")))))
	})

	It("gives the editor a builtin schema twice, unchanged by bundling", func() {
		page := write("menus/stew.recipe.md", "---\nserves: 2\n---\n")

		first, err := EditorSchema(context.Background(), []byte("serves: 2\n"), page, "")
		Expect(err).NotTo(HaveOccurred())
		second, err := EditorSchema(context.Background(), []byte("serves: 2\n"), page, "")

		Expect(err).NotTo(HaveOccurred())
		Expect(second).To(Equal(first))
	})
})
