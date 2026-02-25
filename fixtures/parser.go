package fixtures

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/goccy/go-yaml"
)

// ParseMarkdownFixtures parses markdown files containing test fixtures
func ParseMarkdownFixtures(fixtureFilePath string) ([]FixtureNode, error) {
	file, err := os.Open(fixtureFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open fixture file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Get the directory containing the fixture file
	sourceDir := filepath.Dir(fixtureFilePath)

	// Parse front-matter if present
	frontMatter, content, err := parseFrontMatter(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse front-matter: %w", err)
	}

	// If we have content after front-matter, parse that
	if content != "" {
		nodes, err := parseMarkdownContentWithSourceDir(content, frontMatter, sourceDir)
		if err != nil {
			return nil, err
		}
		// Expand fixtures for file patterns if specified
		nodes, err = expandFixturesForFiles(nodes, frontMatter, sourceDir)
		if err != nil {
			return nil, fmt.Errorf("failed to expand fixtures for files: %w", err)
		}
		// Set SourceDir on all fixture tests
		setSourceDirOnNodes(nodes, sourceDir)
		return nodes, nil
	}

	// No front-matter found, read the entire file
	_, _ = file.Seek(0, 0)
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading fixture file: %w", err)
	}

	fullContent := strings.Join(lines, "\n")
	nodes, err := parseMarkdownContentWithSourceDir(fullContent, nil, sourceDir)
	if err != nil {
		return nil, err
	}
	// No file expansion without frontmatter
	// Set SourceDir on all fixture tests
	setSourceDirOnNodes(nodes, sourceDir)
	return nodes, nil
}

// parseTableRow converts a table row into a FixtureTest
func parseTableRow(headers, values []string) *FixtureNode {
	if len(headers) != len(values) {
		return nil
	}

	fixture := FixtureTest{
		Expected: Expectations{},
	}

	for i, header := range headers {
		value := values[i]
		header = strings.ToLower(strings.TrimSpace(header))

		switch header {
		case "test name", "name":
			fixture.Name = value
		case "cwd", "working directory", "dir":
			fixture.CWD = value
		case "query":
			fixture.Query = value
		case "cli", "command", "exec":
			fixture.Exec = value
		case "cli args", "args", "arguments":
			fixture.Args = strings.Split(value, " ")
		case "exit code", "exitcode", "expected exit code":
			if value != "" && value != "-" {
				code, err := strconv.Atoi(value)
				if err == nil {
					fixture.Expected.ExitCode = &code
				}
			}
		case "expected count", "count":
			if value != "" && value != "-" {
				count, err := strconv.Atoi(value)
				if err == nil {
					fixture.Expected.Count = &count
				}
			}
		case "expected output", "output":
			fixture.Expected.Stdout = value
		case "expected format", "format":
			fixture.Expected.Format = value
		case "expected error", "error":
			fixture.Expected.Error = value
		case "expected matches", "matches":
			fixture.Expected.Output = value
		case "expected results", "results":
			fixture.Expected.Output = value
		case "expected files", "files":
			fixture.Expected.Output = value
		case "template output":
			fixture.Expected.Output = value
		case "cel validation", "cel", "validation", "expr":
			fixture.Expected.CEL = value
		default:
			if fixture.Expected.Properties == nil {
				fixture.Expected.Properties = make(map[string]any)
			}
			fixture.Expected.Properties[header] = value
		}
	}

	// Don't return fixtures without names
	if fixture.Name == "" {
		return nil
	}

	return &FixtureNode{
		Type: TestNode,
		Test: &fixture,
	}
}

// parseFrontMatter extracts YAML front-matter from a markdown file
func parseFrontMatter(file *os.File) (*FrontMatter, string, error) {
	scanner := bufio.NewScanner(file)

	// Check for front-matter delimiter
	if !scanner.Scan() {
		return nil, "", nil
	}

	firstLine := strings.TrimSpace(scanner.Text())
	if firstLine != "---" {
		// No front-matter
		_, _ = file.Seek(0, 0)
		return nil, "", nil
	}

	// Collect front-matter lines
	var frontMatterLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			// End of front-matter
			break
		}
		frontMatterLines = append(frontMatterLines, line)
	}

	// Parse YAML front-matter
	frontMatterYAML := strings.Join(frontMatterLines, "\n")
	var frontMatter FrontMatter
	if err := yaml.Unmarshal([]byte(frontMatterYAML), &frontMatter); err != nil {
		return nil, "", fmt.Errorf("failed to parse YAML front-matter: %w", err)
	}

	// Collect remaining content
	var contentLines []string
	for scanner.Scan() {
		contentLines = append(contentLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, "", err
	}

	content := strings.Join(contentLines, "\n")
	return &frontMatter, content, nil
}

// parseMarkdownContentWithSourceDir parses markdown content with source directory context
func parseMarkdownContentWithSourceDir(content string, frontMatter *FrontMatter, sourceDir string) ([]FixtureNode, error) {
	// Try goldmark parser first (supports command: format)
	return parseMarkdownWithGoldmark(content, frontMatter, sourceDir)
}

// ParseAllFixtures parses all markdown files in a directory
func ParseAllFixtures(dir string) ([]FixtureNode, error) {
	var allFixtures []FixtureNode

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixtures directory: %w", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".md") {
			filepath := dir + "/" + file.Name()
			fixtures, err := ParseMarkdownFixtures(filepath)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", file.Name(), err)
			}
			allFixtures = append(allFixtures, fixtures...)
		}
	}

	return allFixtures, nil
}

// GroupFixturesByCategory groups fixtures by their markdown file or section
func GroupFixturesByCategory(fixtures []FixtureTest) map[string][]FixtureTest {
	grouped := make(map[string][]FixtureTest)

	for _, fixture := range fixtures {
		// Use CWD as a simple categorization for now
		category := fixture.CWD
		if category == "" {
			category = "default"
		}
		grouped[category] = append(grouped[category], fixture)
	}

	return grouped
}

// GetFixtureByName finds a specific fixture by name
func GetFixtureByName(fixtures []FixtureTest, name string) *FixtureTest {
	for _, fixture := range fixtures {
		if fixture.Name == name {
			return &fixture
		}
	}
	return nil
}

// ParseFixtureForTest parses a fixture file and returns the parsed structure
// without executing any commands. Used for validating parsing logic in tests.
func ParseFixtureForTest(filePath string) (*FixtureNode, error) {
	return ParseMarkdownFixturesWithTree(filePath)
}

// ParseMarkdownFixturesWithTree parses markdown files into a hierarchical tree structure
func ParseMarkdownFixturesWithTree(filePath string) (*FixtureNode, error) {
	fileTree := &FixtureNode{
		Name:     filepath.Base(filePath),
		Type:     FileNode,
		Children: make([]*FixtureNode, 0),
	}

	// Get the directory containing the fixture file
	sourceDir := filepath.Dir(filePath)

	// Parse the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Parse front-matter and get content
	frontMatter, content, err := parseFrontMatter(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse front-matter: %w", err)
	}

	// If no content after front-matter, read the entire file
	if content == "" {
		_, _ = file.Seek(0, 0)
		scanner := bufio.NewScanner(file)
		var lines []string
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		content = strings.Join(lines, "\n")
	}

	// Use the new AST-based parser to build the tree directly
	contentTree, err := parseMarkdownWithGoldmarkTree(content, frontMatter, sourceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse markdown content: %w", err)
	}

	// Move children from content tree to file tree
	for _, child := range contentTree.Children {
		fileTree.AddChild(child)
	}

	return fileTree, nil
}

// buildSectionPath constructs a section path from the stack

// adjustSectionStack adjusts the section stack to the correct level
func adjustSectionStack(stack *[]*FixtureNode, targetLevel int) {
	if targetLevel < 0 {
		*stack = []*FixtureNode{}
		return
	}

	if len(*stack) > targetLevel {
		*stack = (*stack)[:targetLevel]
	}
}

// setSourceDirOnNodes recursively sets the SourceDir on all fixture tests in the node tree
func setSourceDirOnNodes(nodes []FixtureNode, sourceDir string) {
	for i := range nodes {
		setSourceDirOnNode(&nodes[i], sourceDir)
	}
}

// setSourceDirOnNode recursively sets SourceDir on a node and its children
func setSourceDirOnNode(node *FixtureNode, sourceDir string) {
	if node.Test != nil {
		node.Test.SourceDir = sourceDir
	}
	for _, child := range node.Children {
		setSourceDirOnNode(child, sourceDir)
	}
}

// expandFixturesForFiles expands a fixture into multiple fixtures based on file glob pattern
func expandFixturesForFiles(fixtures []FixtureNode, frontMatter *FrontMatter, sourceDir string) ([]FixtureNode, error) {
	// If no files pattern specified, return fixtures as-is
	if frontMatter == nil || frontMatter.Files == "" {
		return fixtures, nil
	}

	var expandedFixtures []FixtureNode

	// Find files matching the glob pattern
	// Start search from the source directory (where the fixture file is located)
	pattern := frontMatter.Files

	// If pattern is absolute, use it directly
	if !filepath.IsAbs(pattern) {
		// Make pattern relative to search path
		pattern = filepath.Join(sourceDir, pattern)
	}

	// Use doublestar to find matching files
	matches, err := doublestar.FilepathGlob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern '%s': %w", frontMatter.Files, err)
	}

	// If no matches found, return original fixtures
	if len(matches) == 0 {
		return fixtures, nil
	}

	// For each matched file, create a copy of each fixture with template variables
	for _, matchedFile := range matches {
		// Get file info
		absFile, err := filepath.Abs(matchedFile)
		if err != nil {
			continue
		}

		fileInfo, err := os.Stat(absFile)
		if err != nil || fileInfo.IsDir() {
			continue // Skip directories
		}

		// Calculate template variables
		fileDir := filepath.Dir(absFile)
		fileName := filepath.Base(absFile)
		fileNameNoExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))

		// Make paths relative to source directory if possible
		relFile, _ := filepath.Rel(sourceDir, absFile)
		relDir, _ := filepath.Rel(sourceDir, fileDir)

		// Create template variables map
		templateVars := map[string]string{
			"file":     relFile,                // Relative path to file
			"filename": fileNameNoExt,          // Filename without extension
			"dir":      relDir,                 // Directory containing the file
			"absfile":  absFile,                // Absolute path to file
			"absdir":   fileDir,                // Absolute directory
			"basename": fileName,               // Full filename with extension
			"ext":      filepath.Ext(fileName), // File extension
		}

		// Create a copy of each fixture with the template variables
		for _, fixture := range fixtures {
			// Deep copy the fixture
			expandedFixture := fixture
			if expandedFixture.Test != nil {
				// Create a new test copy
				testCopy := *expandedFixture.Test

				// Update the test name to include the file
				if testCopy.Name != "" {
					testCopy.Name = fmt.Sprintf("%s [%s]", testCopy.Name, relFile)
				}

				// Set template variables
				testCopy.TemplateVars = make(map[string]any)
				for k, v := range templateVars {
					testCopy.TemplateVars[k] = v
				}

				expandedFixture.Test = &testCopy
			}

			expandedFixtures = append(expandedFixtures, expandedFixture)
		}
	}

	// If we expanded fixtures, return the expanded list; otherwise return original
	if len(expandedFixtures) > 0 {
		return expandedFixtures, nil
	}

	return fixtures, nil
}
