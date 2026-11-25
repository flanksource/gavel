package main

import (
	"fmt"

	jsonlint "github.com/flanksource/gavel/linters/json"
	"github.com/flanksource/gavel/verify"
)

// The fixture schemas `gavel fixtures --schema` generates are the schemas of
// the fixture files that declare none, for the json linter and `gavel edit`.
func init() {
	jsonlint.RegisterBuiltinSchema(jsonlint.BuiltinSchema{
		Name: "fixture-frontmatter",
		Matches: func(path string, cfg verify.GavelConfig) (bool, error) {
			return jsonlint.MatchesFixtureFiles(path, cfg, ".md", ".mdx")
		},
		Load: func() (map[string]any, error) { return fixtureFrontmatterSchema(), nil },
	})
	jsonlint.RegisterBuiltinSchema(jsonlint.BuiltinSchema{
		Name: "fixture-document",
		Matches: func(path string, cfg verify.GavelConfig) (bool, error) {
			return jsonlint.MatchesFixtureFiles(path, cfg, ".yaml", ".yml", ".json")
		},
		Load: func() (map[string]any, error) {
			doc, err := fixturesSchemaDocument()
			if err != nil {
				return nil, err
			}
			document, ok := doc["document"].(map[string]any)
			if !ok {
				return nil, fmt.Errorf("fixture schema bundle has no document schema (got %T)", doc["document"])
			}
			return document, nil
		},
	})
}
