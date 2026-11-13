package jsonlint

import (
	"bytes"
	"errors"
	"path/filepath"
	"regexp"

	"github.com/flanksource/gavel/models"
	"gopkg.in/yaml.v3"
)

var (
	frontmatterOpening = regexp.MustCompile(`^(?:\x{FEFF})?---\r?\n`)
	frontmatterClosing = regexp.MustCompile(`(?m)^---[ \t]*\r?$`)
)

var errUnclosedFrontmatter = errors.New("unclosed Markdown frontmatter")

// markdownFrontmatter returns the YAML between a Markdown file's opening and
// closing --- lines and how many lines precede it; ok is false for a file
// without frontmatter.
func markdownFrontmatter(data []byte) (frontmatter []byte, offset int, ok bool, err error) {
	opening := frontmatterOpening.Find(data)
	if opening == nil {
		return nil, 0, false, nil
	}
	rest := data[len(opening):]
	closing := frontmatterClosing.FindIndex(rest)
	if closing == nil {
		return nil, 0, true, errUnclosedFrontmatter
	}
	return rest[:closing[0]], bytes.Count(opening, []byte("\n")), true, nil
}

// lintMarkdown validates a Markdown file's frontmatter as one YAML document,
// reporting violations at their lines in the Markdown file.
func (v *validator) lintMarkdown(input fileInput) ([]models.Violation, error) {
	frontmatter, offset, ok, err := markdownFrontmatter(input.data)
	if errors.Is(err, errUnclosedFrontmatter) {
		return []models.Violation{violation(input.path, 1, 1, "syntax", err.Error())}, nil
	}
	if !ok {
		return nil, nil
	}
	if input.schema == "" {
		if input.schema, err = yamlHeader(frontmatter); err != nil {
			return nil, err
		}
		if input.schema != "" {
			if input.schema, err = ResolveSchema(input.schema, filepath.Dir(input.path)); err != nil {
				return nil, err
			}
		}
	}
	doc := document{offset: offset}
	if err := yaml.Unmarshal(frontmatter, &doc.node); err != nil {
		return []models.Violation{syntaxViolation(input.path, frontmatter, err, offset)}, nil
	}
	if doc.node.Kind == 0 {
		// Empty frontmatter validates as an empty object.
		doc.node = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}}
	}
	input.data = frontmatter
	return v.validateYAML(input, doc)
}
