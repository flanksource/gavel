// Package todos provides functionality for parsing, discovering, and executing
// TODO markdown files with YAML frontmatter and executable test fixtures.
package todos

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
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
	// Parse as fixture - this captures ALL frontmatter in metadata
	fileNode, err := fixtures.ParseMarkdownFixturesWithTree(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse markdown: %w", err)
	}

	// Extract TODO-specific fields from fixture metadata
	todoFrontmatter := extractTODOFrontmatter(fileNode)

	// Set default cwd to git root if not specified
	if todoFrontmatter.CWD == "" {
		todoFrontmatter.CWD = findGitRoot(filePath)
	}

	// Validate required TODO fields
	if err := validateFrontmatter(&todoFrontmatter); err != nil {
		return nil, err
	}

	// Extract sections for backward compatibility
	stepsToReproduce := extractSection(fileNode, "Steps to Reproduce")
	implementation := extractImplementationText(fileNode, "Implementation")
	verification := extractSection(fileNode, "Verification")
	customValidations := extractSection(fileNode, "Custom Validations")

	return &types.TODO{
		FilePath:          filePath,
		FileNode:          fileNode,
		TODOFrontmatter:   todoFrontmatter,
		StepsToReproduce:  stepsToReproduce,
		Implementation:    implementation,
		Verification:      verification,
		CustomValidations: customValidations,
	}, nil
}

// extractTODOFrontmatter extracts TODO-specific fields from fixture metadata.
// Since the fixture parser captures all frontmatter fields in metadata,
// we just need to extract and type-convert the TODO fields.
// We walk the tree to find the first test node which contains the frontmatter.
func extractTODOFrontmatter(fileNode *fixtures.FixtureNode) types.TODOFrontmatter {
	todoFM := types.TODOFrontmatter{}

	// Walk the tree to find first test node with frontmatter
	var firstTest *fixtures.FixtureTest
	var walk func(*fixtures.FixtureNode)
	walk = func(node *fixtures.FixtureNode) {
		if firstTest != nil {
			return // Already found
		}
		if node.Test != nil {
			firstTest = node.Test
			return
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(fileNode)

	// If no test found, return empty frontmatter
	if firstTest == nil {
		return todoFM
	}

	// Copy the fixture frontmatter
	todoFM.FrontMatter = firstTest.FrontMatter

	// Extract TODO-specific fields from metadata
	if firstTest.Metadata != nil {
		metadata := firstTest.Metadata

		// Priority
		if val, ok := metadata["priority"].(string); ok {
			todoFM.Priority = types.Priority(val)
		}

		// Status
		if val, ok := metadata["status"].(string); ok {
			todoFM.Status = types.Status(val)
		}

		// Language
		if val, ok := metadata["language"].(string); ok {
			todoFM.Language = types.Language(val)
		}

		// Attempts
		if val, ok := metadata["attempts"].(int); ok {
			todoFM.Attempts = val
		}

		// LastRun (parse time string)
		if val, ok := metadata["last_run"].(string); ok {
			if t, err := time.Parse(time.RFC3339, val); err == nil {
				todoFM.LastRun = &t
			}
		}

		// LLM configuration
		if llmData, ok := metadata["llm"].(map[string]interface{}); ok {
			llm := &types.LLM{}
			if model, ok := llmData["model"].(string); ok {
				llm.Model = model
			}
			if maxTokens, ok := llmData["max_tokens"].(int); ok {
				llm.MaxTokens = maxTokens
			}
			if maxCost, ok := llmData["max_cost"].(float64); ok {
				llm.MaxCost = maxCost
			}
			if tokensUsed, ok := llmData["tokens_used"].(int); ok {
				llm.TokensUsed = tokensUsed
			}
			if costIncurred, ok := llmData["cost_incurred"].(float64); ok {
				llm.CostIncurred = costIncurred
			}
			if sessionId, ok := llmData["session_id"].(string); ok {
				llm.SessionId = sessionId
			}
			todoFM.LLM = llm
		}
	}

	return todoFM
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
