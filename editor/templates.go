package editor

import (
	"embed"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// The sources carry a .tmpl suffix so fixture discovery and linters never
// treat them as the gavel files they seed.
//
//go:embed templates/*.tmpl
var templateFS embed.FS

// Template seeds a new file. Suffix is the ending every file created from it
// must have, which keeps the file the kind (and schema) the template is for.
type Template struct {
	ID          string `json:"id"`
	Group       string `json:"group"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Suffix      string `json:"suffix"`
	Source      string `json:"source"`
	file        string
}

var templateCatalog = []Template{
	{ID: "markdown", Group: "Markdown", Label: "Markdown document", Description: "A heading and a paragraph", Name: "untitled.md", Suffix: ".md", file: "markdown.md.tmpl"},
	{ID: "markdown-frontmatter", Group: "Markdown", Label: "Markdown with frontmatter", Description: "YAML frontmatter, edited as a form when a schema applies", Name: "untitled.md", Suffix: ".md", file: "markdown-frontmatter.md.tmpl"},
	{ID: "fixture-markdown", Group: "Markdown", Label: "Gavel fixture", Description: "Command tests with CEL assertions, run by gavel fixtures", Name: "untitled.fixture.md", Suffix: ".fixture.md", file: "fixture.fixture.md.tmpl"},
	{ID: "facet-mdx", Group: "Facet", Label: "Facet MDX report", Description: "Markdown prose with facet StatCard, ProgressBar and Badge components", Name: "report.mdx", Suffix: ".mdx", file: "facet-report.mdx.tmpl"},
	{ID: "facet-tsx", Group: "Facet", Label: "Facet TSX page", Description: "A typed React page built from facet components", Name: "page.tsx", Suffix: ".tsx", file: "facet-page.tsx.tmpl"},
	{ID: "facet-jsx", Group: "Facet", Label: "Facet JSX page", Description: "A React page built from facet components", Name: "page.jsx", Suffix: ".jsx", file: "facet-page.jsx.tmpl"},
	{ID: "json", Group: "JSON & YAML", Label: "JSON document", Description: "A JSON object", Name: "untitled.json", Suffix: ".json", file: "document.json.tmpl"},
	{ID: "json-schema", Group: "JSON & YAML", Label: "JSON Schema", Description: "An object schema to validate documents against", Name: "untitled.schema.json", Suffix: ".schema.json", file: "schema.schema.json.tmpl"},
	{ID: "yaml", Group: "JSON & YAML", Label: "YAML document", Description: "A YAML mapping", Name: "untitled.yaml", Suffix: ".yaml", file: "document.yaml.tmpl"},
	{ID: "fixture-yaml", Group: "JSON & YAML", Label: "Gavel fixture (YAML)", Description: "A kind: Fixture document of exec steps with CEL assertions", Name: "untitled.fixture.yaml", Suffix: ".fixture.yaml", file: "fixture.fixture.yaml.tmpl"},
}

// Templates is the catalog of new-file templates with their sources.
var Templates = sync.OnceValues(func() ([]Template, error) {
	templates := make([]Template, 0, len(templateCatalog))
	for _, template := range templateCatalog {
		data, err := templateFS.ReadFile("templates/" + template.file)
		if err != nil {
			return nil, fmt.Errorf("template %q: %w", template.ID, err)
		}
		template.Source = string(data)
		templates = append(templates, template)
	}
	return templates, nil
})

// sourceFor is the source a new file at path starts with: the template's,
// checked against its suffix, or empty without a template.
func sourceFor(templateID, path string) (string, error) {
	if templateID == "" {
		return "", nil
	}
	templates, err := Templates()
	if err != nil {
		return "", err
	}
	for _, template := range templates {
		if template.ID != templateID {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(path), template.Suffix) {
			return "", fmt.Errorf("a %s must be named *%s, not %s", template.Label, template.Suffix, path)
		}
		return template.Source, nil
	}
	return "", fmt.Errorf("unknown template %q", templateID)
}

func (s *Server) templates(w http.ResponseWriter, _ *http.Request) {
	templates, err := Templates()
	respond(w, templates, err)
}
