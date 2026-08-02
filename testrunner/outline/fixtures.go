package outline

import (
	"fmt"
	"path/filepath"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
)

func configuredFixturePatterns(workDir string, overrides []string) ([]string, error) {
	if len(overrides) > 0 {
		return append([]string(nil), overrides...), nil
	}
	config, err := verify.LoadGavelConfig(workDir)
	if err != nil {
		return nil, fmt.Errorf("load fixture config: %w", err)
	}
	return config.Fixtures.ResolvedFiles(), nil
}

func collectFixtureTests(workDir string, filters, patterns []string) ([]*Entry, error) {
	files, err := discoverFixtureFiles(workDir, filters, patterns)
	if err != nil || len(files) == 0 {
		return nil, err
	}
	report, err := fixtures.Outline(fixtures.OutlineOptions{Paths: files, WorkDir: workDir})
	if err != nil {
		return nil, err
	}

	var entries []*Entry
	var walk func([]*fixtures.OutlineNode, string)
	walk = func(nodes []*fixtures.OutlineNode, fixtureFile string) {
		for _, node := range nodes {
			currentFile := fixtureFile
			if node.Type == "file" && node.File != "" {
				currentFile = node.File
			}
			if node.Kind != "" && node.Type != "criterion" {
				file := node.File
				if !filepath.IsAbs(file) && currentFile != "" {
					file = currentFile
				}
				entry := &Entry{
					Framework:   parsers.Fixture,
					File:        relativeTo(file, workDir),
					Line:        node.Line,
					Name:        node.Name,
					Description: node.Summary,
					Labels:      []string{node.Kind},
				}
				entries = append(entries, entry)
			}
			walk(node.Children, currentFile)
		}
	}
	walk(report.Tree, "")
	return entries, nil
}

// discoverFixtureFiles resolves fixture patterns against workDir with the same
// bounded, gitignore-aware walk the gotest/ginkgo/node runners use, then applies
// the outline's positional path filters.
func discoverFixtureFiles(workDir string, filters, patterns []string) ([]string, error) {
	matches, err := utils.GlobFilesBounded(workDir, patterns)
	if err != nil {
		return nil, fmt.Errorf("discover fixtures: %w", err)
	}
	if len(filters) == 0 {
		return matches, nil
	}
	var files []string
	for _, file := range matches {
		if matchesFilters(relativeTo(file, workDir), filters) {
			files = append(files, file)
		}
	}
	return files, nil
}
