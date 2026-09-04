package fixtures

import (
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Fixture file expansion", func() {
	ginkgo.It("expands each table row for every frontmatter file match", func() {
		root := ginkgo.GinkgoT().TempDir()
		fixtureDir := filepath.Join(root, "fixtures")
		gomega.Expect(os.MkdirAll(fixtureDir, 0o750)).To(gomega.Succeed())

		for _, name := range []string{"alpha.tsx", "beta.tsx"} {
			gomega.Expect(os.WriteFile(filepath.Join(root, name), []byte(name), 0o600)).To(gomega.Succeed())
		}

		fixturePath := filepath.Join(fixtureDir, "html.md")
		gomega.Expect(os.WriteFile(fixturePath, []byte(`---
files: ../*.tsx
exec: facet
args: [html, "{{.absfile}}"]
---

# HTML rendering

| Name | Exit Code |
|------|-----------|
| Render TSX | 0 |
`), 0o600)).To(gomega.Succeed())

		tree, err := ParseMarkdownFixturesWithTree(fixturePath)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		type expandedFixture struct {
			NodeName string
			Name     string
			File     any
			Filename any
			AbsFile  any
		}
		actual := struct {
			Fixtures []expandedFixture
			Count    int
		}{}
		tree.Walk(func(node *FixtureNode) {
			if node.Test == nil {
				return
			}
			actual.Fixtures = append(actual.Fixtures, expandedFixture{
				NodeName: node.Name,
				Name:     node.Test.Name,
				File:     node.Test.TemplateVars["file"],
				Filename: node.Test.TemplateVars["filename"],
				AbsFile:  node.Test.TemplateVars["absfile"],
			})
		})
		actual.Count = BuildOutline(tree, root).Fixtures

		gomega.Expect(actual).To(gomega.Equal(struct {
			Fixtures []expandedFixture
			Count    int
		}{
			Fixtures: []expandedFixture{
				{
					NodeName: "Render TSX [../alpha.tsx]",
					Name:     "Render TSX [../alpha.tsx]",
					File:     "../alpha.tsx",
					Filename: "alpha",
					AbsFile:  filepath.Join(root, "alpha.tsx"),
				},
				{
					NodeName: "Render TSX [../beta.tsx]",
					Name:     "Render TSX [../beta.tsx]",
					File:     "../beta.tsx",
					Filename: "beta",
					AbsFile:  filepath.Join(root, "beta.tsx"),
				},
			},
			Count: 2,
		}))
	})
})
