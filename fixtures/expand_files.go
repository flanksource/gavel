package fixtures

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

func expandFixtureTreeForFiles(root *FixtureNode, frontMatter *FrontMatter, sourceDir string) error {
	if root == nil || frontMatter == nil || frontMatter.Files == "" {
		return nil
	}

	matches, err := matchFixtureFiles(frontMatter.Files, sourceDir)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return nil
	}

	expandFixtureTreeChildren(root, matches, sourceDir)
	return nil
}

func expandFixtureTreeChildren(node *FixtureNode, matches []string, sourceDir string) {
	expanded := make([]*FixtureNode, 0, len(node.Children)*len(matches))
	for _, child := range node.Children {
		if child.Test != nil {
			for _, matchedFile := range matches {
				clone := *child
				clone.Test = expandFixtureTestForFile(child.Test, matchedFile, sourceDir)
				clone.Name = clone.Test.Name
				expanded = append(expanded, &clone)
			}
			continue
		}

		expandFixtureTreeChildren(child, matches, sourceDir)
		expanded = append(expanded, child)
	}
	node.Children = expanded
}

func expandFixtureTestForFile(test *FixtureTest, matchedFile, sourceDir string) *FixtureTest {
	clone := *test
	clone.TemplateVars = globTemplateVars(matchedFile, sourceDir)
	if clone.Name != "" {
		clone.Name = fmt.Sprintf("%s [%s]", clone.Name, clone.TemplateVars["file"])
	}
	return &clone
}

func matchFixtureFiles(pattern, sourceDir string) ([]string, error) {
	requestedPattern := pattern
	if !filepath.IsAbs(pattern) {
		pattern = filepath.Join(sourceDir, pattern)
	}

	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern '%s': %w", requestedPattern, err)
	}

	files := make([]string, 0, len(matches))
	for _, match := range matches {
		absFile, err := filepath.Abs(match)
		if err != nil {
			continue
		}
		info, err := os.Stat(absFile)
		if err == nil && !info.IsDir() {
			files = append(files, absFile)
		}
	}
	return files, nil
}
