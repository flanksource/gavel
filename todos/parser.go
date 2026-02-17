// Package todos provides functionality for parsing, discovering, and executing
// TODO markdown files with YAML frontmatter and executable test fixtures.
package todos

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
	"github.com/goccy/go-yaml"
)

// ParseTODO parses a TODO markdown file from the given path and returns a structured TODO object.
// The file is parsed as a fixture, with all frontmatter fields captured in metadata.
// TODO-specific fields (priority, status, language, etc.) are extracted from the fixture metadata.
//
//	---
//	priority: high|medium|low
//	status: pending|in_progress|completed|failed|skipped
//	language: go|typescript|python
//	cwd: /path/to/working/dir
//	---
//	## Any Section Name
//	```bash
//	command to run
//	```
//
// Returns an error if the file cannot be read or parsing fails.
func ParseTODO(filePath string) (*types.TODO, error) {
	// Parse YAML frontmatter directly into TODOFrontmatter
	todoFrontmatter, err := parseTODOFrontmatter(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	if todoFrontmatter.CWD == "" {
		todoFrontmatter.CWD = findGitRoot(filePath)
	}

	if err := validateFrontmatter(&todoFrontmatter); err != nil {
		return nil, err
	}

	// Parse fixture tree for sections (Steps to Reproduce, Verification, etc.)
	fileNode, err := fixtures.ParseMarkdownFixturesWithTree(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse markdown: %w", err)
	}

	return &types.TODO{
		FilePath:          filePath,
		FileNode:          fileNode,
		TODOFrontmatter:   todoFrontmatter,
		StepsToReproduce:  extractSection(fileNode, "Steps to Reproduce"),
		Implementation:    extractImplementationText(fileNode, "Implementation"),
		Verification:      extractSection(fileNode, "Verification"),
		CustomValidations: extractSection(fileNode, "Custom Validations"),
	}, nil
}

// parseTODOFrontmatter reads YAML frontmatter directly from a markdown file
// and unmarshals it into TODOFrontmatter. This works regardless of whether
// the file contains executable code blocks.
func parseTODOFrontmatter(filePath string) (types.TODOFrontmatter, error) {
	var fm types.TODOFrontmatter

	rawYAML, err := extractRawFrontmatter(filePath)
	if err != nil {
		return fm, err
	}
	if rawYAML == "" {
		return fm, nil
	}

	if err := yaml.Unmarshal([]byte(rawYAML), &fm); err != nil {
		return fm, fmt.Errorf("failed to unmarshal frontmatter: %w", err)
	}
	fm.CleanMetadata()
	return fm, nil
}

// extractRawFrontmatter reads the YAML frontmatter string from a markdown file
// delimited by --- markers.
func extractRawFrontmatter(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return "", nil
	}
	if strings.TrimSpace(scanner.Text()) != "---" {
		return "", nil
	}

	var lines []string
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "---" {
			return strings.Join(lines, "\n"), nil
		}
		lines = append(lines, scanner.Text())
	}
	return "", scanner.Err()
}

// validateFrontmatter validates that required fields in the TODO frontmatter are present
// and contain valid values. Required fields include priority, status, and optionally language.
// Returns an error if validation fails.
func validateFrontmatter(fm *types.TODOFrontmatter) error {
	// Validate required fields
	if fm.Priority == "" {
		return fmt.Errorf("missing required field: priority")
	}

	// Validate priority value
	if fm.Priority != types.PriorityHigh && fm.Priority != types.PriorityMedium && fm.Priority != types.PriorityLow {
		return fmt.Errorf("invalid priority: must be high, medium, or low")
	}

	// Validate status value
	validStatuses := []types.Status{types.StatusPending, types.StatusInProgress, types.StatusCompleted, types.StatusFailed, types.StatusSkipped}
	validStatus := false
	for _, s := range validStatuses {
		if fm.Status == s {
			validStatus = true
			break
		}
	}
	if !validStatus {
		return fmt.Errorf("invalid status: must be pending, in_progress, completed, failed, or skipped")
	}

	// Validate language value
	if fm.Language != "" {
		validLanguages := []types.Language{types.LanguageGo, types.LanguageTypeScript, types.LanguagePython}
		validLanguage := false
		for _, l := range validLanguages {
			if fm.Language == l {
				validLanguage = true
				break
			}
		}
		if !validLanguage {
			return fmt.Errorf("invalid language: must be go, typescript, or python")
		}
	}

	return nil
}

func findGitRoot(path string) string {
	dir := path
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		dir = filepath.Dir(path)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			if abs, err := filepath.Abs(dir); err == nil {
				return abs
			}
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// extractSection finds and returns all fixture nodes matching the specified section name
// by walking the fixture tree. Section names typically include "Steps to Reproduce",
// "Verification", and "Custom Validations". Returns an empty slice if no matches found.
func extractSection(root *fixtures.FixtureNode, sectionName string) []*fixtures.FixtureNode {
	var nodes []*fixtures.FixtureNode

	// Walk the tree and find section nodes matching the name
	var walk func(*fixtures.FixtureNode)
	walk = func(node *fixtures.FixtureNode) {
		if node.Type == fixtures.SectionNode && node.Name == sectionName {
			nodes = append(nodes, node)
			return
		}
		for _, child := range node.Children {
			walk(child)
		}
	}

	walk(root)
	return nodes
}

// extractImplementationText extracts the plain text content from the Implementation section.
// The Implementation section contains free-form instructions for Claude Code to execute.
// Returns an empty string if the section is not found.
func extractImplementationText(root *fixtures.FixtureNode, sectionName string) string {
	sections := extractSection(root, sectionName)
	if len(sections) == 0 {
		return ""
	}

	// For now, we'll extract text from the first matching section
	// In a real implementation, we might parse markdown content more thoroughly
	var texts []string
	var collectText func(*fixtures.FixtureNode)
	collectText = func(node *fixtures.FixtureNode) {
		if node.Test != nil && node.Test.Query != "" {
			texts = append(texts, node.Test.Query)
		}
		for _, child := range node.Children {
			collectText(child)
		}
	}

	collectText(sections[0])
	if len(texts) > 0 {
		return strings.Join(texts, "\n")
	}

	// Return section name as placeholder for now
	return sectionName + " content"
}
