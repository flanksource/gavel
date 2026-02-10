package todos

import (
	"fmt"
	"os"
	"strings"

	"github.com/flanksource/gavel/todos/types"
	"github.com/goccy/go-yaml"
)

// FrontmatterResult contains parsed frontmatter and remaining markdown content
type FrontmatterResult struct {
	Frontmatter     types.TODOFrontmatter
	MarkdownContent string
}

// ParseFrontmatter extracts and parses YAML frontmatter from markdown content.
// Returns the parsed frontmatter and the remaining markdown content.
func ParseFrontmatter(content string) (*FrontmatterResult, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return nil, fmt.Errorf("invalid TODO file: missing frontmatter")
	}

	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return nil, fmt.Errorf("invalid TODO file: frontmatter not closed")
	}

	frontmatterYAML := strings.Join(lines[1:endIdx], "\n")
	var frontmatter types.TODOFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &frontmatter); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Clean up Metadata to fix goccy/go-yaml inline map behavior
	frontmatter.CleanMetadata()

	markdownContent := strings.Join(lines[endIdx+1:], "\n")

	return &FrontmatterResult{
		Frontmatter:     frontmatter,
		MarkdownContent: markdownContent,
	}, nil
}

// ParseFrontmatterFromFile reads a file and parses its frontmatter.
func ParseFrontmatterFromFile(filePath string) (*FrontmatterResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return ParseFrontmatter(string(content))
}

// WriteFrontmatter serializes frontmatter to YAML and combines with markdown content.
func WriteFrontmatter(frontmatter *types.TODOFrontmatter, markdownContent string) (string, error) {
	yamlBytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(yamlBytes)
	sb.WriteString("---\n")
	sb.WriteString(markdownContent)

	return sb.String(), nil
}
