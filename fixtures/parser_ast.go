package fixtures

import (
	"fmt"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/goccy/go-yaml"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// commandBlockBuilder helps build a command fixture from markdown AST
type commandBlockBuilder struct {
	name        string
	content     string
	language    string
	frontmatter string
	validations []string
	isComplete  bool
}

// parseMarkdownWithGoldmarkTree parses markdown content using goldmark AST parser and returns a tree structure
func parseMarkdownWithGoldmarkTree(content string, frontMatter *FrontMatter, sourceDir string) (*FixtureNode, error) {
	rootNode := &FixtureNode{
		Name:     "Content",
		Type:     SectionNode,
		Level:    0,
		Children: make([]*FixtureNode, 0),
	}

	// Create goldmark parser with table extension
	md := goldmark.New(
		goldmark.WithExtensions(extension.Table),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	source := []byte(content)
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	// State for parsing
	var currentCommand *commandBlockBuilder
	var inCommandBlock bool
	var sectionStack []*FixtureNode
	var currentSection = rootNode
	var standaloneCodeBlock *commandBlockBuilder // For standalone code blocks without "command:" prefix
	var parentHeading string                     // For generating test names from context

	// Walk the AST
	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			headingText := extractNodeText(node, source)
			level := node.Level

			// Complete any pending standalone code block
			if standaloneCodeBlock != nil && !standaloneCodeBlock.isComplete {
				if fixture := buildFixtureFromCommand(standaloneCodeBlock, frontMatter, sourceDir); fixture != nil {
					currentSection.AddChild(&FixtureNode{
						Name:     fixture.Test.Name,
						Type:     TestNode,
						Test:     fixture.Test,
						Children: make([]*FixtureNode, 0),
					})
				}
				standaloneCodeBlock = nil
			}

			// Check if this is a command heading (starts with "command:")
			isCommandHeading := strings.HasPrefix(strings.ToLower(headingText), "command:")

			// Complete previous command block if exists
			if currentCommand != nil && !currentCommand.isComplete {
				if fixture := buildFixtureFromCommand(currentCommand, frontMatter, sourceDir); fixture != nil {
					// Add test to the current section
					currentSection.AddChild(&FixtureNode{
						Name:     fixture.Test.Name,
						Type:     TestNode,
						Test:     fixture.Test,
						Children: make([]*FixtureNode, 0),
					})
				}
				currentCommand = nil
			}

			if isCommandHeading {
				// Start new command block
				commandName := strings.TrimSpace(strings.TrimPrefix(headingText, "command:"))
				if commandName == "" {
					commandName = strings.TrimSpace(strings.TrimPrefix(strings.ToLower(headingText), "command:"))
				}

				currentCommand = &commandBlockBuilder{
					name:        commandName,
					validations: make([]string, 0),
				}
				inCommandBlock = true
				// Don't create a section node for command headings
			} else {
				// Not a command heading - exit command block mode
				inCommandBlock = false

				// Track parent heading for standalone code blocks
				parentHeading = headingText

				// Adjust section stack based on heading level
				adjustSectionStack(&sectionStack, level-1)

				// Create section node for regular headings
				sectionNode := &FixtureNode{
					Name:     headingText,
					Type:     SectionNode,
					Level:    level,
					Children: make([]*FixtureNode, 0),
				}

				// Add to parent (root or parent section)
				parent := rootNode
				if len(sectionStack) > 0 {
					parent = sectionStack[len(sectionStack)-1]
				}
				parent.AddChild(sectionNode)

				// Update stack
				if len(sectionStack) >= level-1 {
					sectionStack = sectionStack[:level-1]
				}
				sectionStack = append(sectionStack, sectionNode)
				currentSection = sectionNode
			}

		case *ast.FencedCodeBlock:
			var infoString string
			if node.Info != nil {
				infoString = string(node.Info.Segment.Value(source))
			}
			lang := strings.ToLower(extractLanguage(infoString))
			codeContent := extractCodeBlockContent(&node.BaseBlock, source)

			if inCommandBlock && currentCommand != nil {
				// Handle code blocks within command blocks (existing behavior)
				// Get allowed code blocks from frontmatter
				allowedBlocks := getCodeBlocksOrDefault(frontMatter)

				// Check if this language should be executed
				if shouldExecuteCodeBlock(lang, allowedBlocks) {
					currentCommand.language = lang
					currentCommand.content = codeContent
				} else if strings.ToLower(lang) == "frontmatter" || strings.ToLower(lang) == "yaml" {
					// Always parse frontmatter/yaml blocks regardless of codeBlocks filter
					currentCommand.frontmatter = codeContent
				} else {
					logger.V(4).Infof("Skipping code block '%s' not in allowed %v", lang, allowedBlocks)
				}
			} else if !inCommandBlock && isExecutableLanguage(lang) {
				// Handle standalone code blocks (new behavior)
				// Complete any pending standalone code block first
				if standaloneCodeBlock != nil && !standaloneCodeBlock.isComplete {
					if fixture := buildFixtureFromCommand(standaloneCodeBlock, frontMatter, sourceDir); fixture != nil {
						currentSection.AddChild(&FixtureNode{
							Name:     fixture.Test.Name,
							Type:     TestNode,
							Test:     fixture.Test,
							Children: make([]*FixtureNode, 0),
						})
					}
					standaloneCodeBlock = nil
				}

				// Generate test name from parent heading or default
				testName := parentHeading
				if testName == "" {
					testName = fmt.Sprintf("%s-block", lang)
				}

				// Create new standalone code block
				standaloneCodeBlock = &commandBlockBuilder{
					name:        testName,
					language:    lang,
					content:     codeContent,
					validations: make([]string, 0),
				}

			} else if !inCommandBlock && standaloneCodeBlock != nil {
				// Handle frontmatter/yaml blocks for standalone code blocks
				if strings.ToLower(lang) == "frontmatter" || strings.ToLower(lang) == "yaml" {
					standaloneCodeBlock.frontmatter = codeContent
				}
			}

		case *ast.List:
			// Check if this is a validation list
			listText := extractNodeText(node, source)
			isValidationList := strings.Contains(strings.ToLower(listText), "validation") ||
				strings.Contains(listText, "cel:") ||
				strings.Contains(listText, "regex:") ||
				strings.Contains(listText, "contains:")

			if inCommandBlock && currentCommand != nil && isValidationList {
				// Handle validations for command blocks (existing behavior)
				validations := extractValidationsFromList(node, source)
				currentCommand.validations = append(currentCommand.validations, validations...)
			} else if !inCommandBlock && standaloneCodeBlock != nil && isValidationList {
				// Handle validations for standalone code blocks (new behavior)
				validations := extractValidationsFromList(node, source)
				standaloneCodeBlock.validations = append(standaloneCodeBlock.validations, validations...)

				// Complete the standalone code block now that we have validations
				if fixture := buildFixtureFromCommand(standaloneCodeBlock, frontMatter, sourceDir); fixture != nil {
					currentSection.AddChild(&FixtureNode{
						Name:     fixture.Test.Name,
						Type:     TestNode,
						Test:     fixture.Test,
						Children: make([]*FixtureNode, 0),
					})
				}
				standaloneCodeBlock = nil
			}

		case *extast.Table:
			// Handle existing table format - add tests to current section
			if !inCommandBlock {
				tableFixtures, err := parseTableFromAST(node, source, frontMatter, sourceDir)
				if err != nil {
					return ast.WalkStop, err
				}
				// Add table fixtures to current section
				for _, fixture := range tableFixtures {
					if fixture.Test != nil {
						testNode := &FixtureNode{
							Name:     fixture.Test.Name,
							Type:     TestNode,
							Test:     fixture.Test,
							Children: make([]*FixtureNode, 0),
						}
						currentSection.AddChild(testNode)
					}
				}
			}
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking AST: %w", err)
	}

	// Complete final standalone code block if exists
	if standaloneCodeBlock != nil && !standaloneCodeBlock.isComplete {
		if fixture := buildFixtureFromCommand(standaloneCodeBlock, frontMatter, sourceDir); fixture != nil {
			currentSection.AddChild(&FixtureNode{
				Name:     fixture.Test.Name,
				Type:     TestNode,
				Test:     fixture.Test,
				Children: make([]*FixtureNode, 0),
			})
		}
	}

	// Complete final command block if exists
	if currentCommand != nil && !currentCommand.isComplete {
		if fixture := buildFixtureFromCommand(currentCommand, frontMatter, sourceDir); fixture != nil {
			// Add test to the current command section
			currentSection.AddChild(&FixtureNode{
				Name:     fixture.Test.Name,
				Type:     TestNode,
				Test:     fixture.Test,
				Children: make([]*FixtureNode, 0),
			})
		}
	}

	return rootNode, nil
}

// parseMarkdownWithGoldmark provides backwards compatibility by converting tree to flat list
func parseMarkdownWithGoldmark(content string, frontMatter *FrontMatter, sourceDir string) ([]FixtureNode, error) {
	tree, err := parseMarkdownWithGoldmarkTree(content, frontMatter, sourceDir)
	if err != nil {
		return nil, err
	}

	var fixtures []FixtureNode
	tree.Walk(func(node *FixtureNode) {
		if node.Test != nil {
			fixtures = append(fixtures, *node)
		}
	})

	return fixtures, nil
}

// extractNodeText extracts plain text content from an AST node
func extractNodeText(node ast.Node, source []byte) string {
	var buf strings.Builder

	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if text, ok := n.(*ast.Text); ok {
				buf.Write(text.Segment.Value(source))
			}
		}
		return ast.WalkContinue, nil
	})

	return buf.String()
}

// extractCodeBlockContent extracts the content from a fenced code block
func extractCodeBlockContent(node *ast.BaseBlock, source []byte) string {
	var buf strings.Builder

	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		buf.Write(line.Value(source))
	}

	return strings.TrimSpace(buf.String())
}

// extractValidationsFromList extracts validation expressions from a list node
func extractValidationsFromList(listNode *ast.List, source []byte) []string {
	var validations []string

	_ = ast.Walk(listNode, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if listItem, ok := n.(*ast.ListItem); ok {
				itemText := extractNodeText(listItem, source)
				itemText = strings.TrimSpace(itemText)

				// Skip empty items
				if itemText == "" {
					return ast.WalkSkipChildren, nil
				}

				// Process different validation types
				if strings.HasPrefix(itemText, "cel:") {
					validations = append(validations, strings.TrimSpace(strings.TrimPrefix(itemText, "cel:")))
				} else if strings.HasPrefix(itemText, "contains:") {
					containsText := strings.TrimSpace(strings.TrimPrefix(itemText, "contains:"))
					// Remove quotes if present
					containsText = strings.Trim(containsText, `"'`)
					validations = append(validations, fmt.Sprintf(`stdout.contains("%s")`, containsText))
				} else if strings.HasPrefix(itemText, "regex:") {
					regexText := strings.TrimSpace(strings.TrimPrefix(itemText, "regex:"))
					// Remove quotes if present
					regexText = strings.Trim(regexText, `"'`)
					// Escape quotes in the regex pattern for CEL string
					regexText = strings.ReplaceAll(regexText, `"`, `\"`)
					validations = append(validations, fmt.Sprintf(`stdout.matches("%s")`, regexText))
				} else if strings.HasPrefix(itemText, "not:") {
					notText := strings.TrimSpace(strings.TrimPrefix(itemText, "not:"))
					if strings.HasPrefix(notText, "contains:") {
						containsText := strings.TrimSpace(strings.TrimPrefix(notText, "contains:"))
						containsText = strings.Trim(containsText, `"'`)
						validations = append(validations, fmt.Sprintf(`!stdout.contains("%s")`, containsText))
					} else {
						validations = append(validations, fmt.Sprintf("!(%s)", notText))
					}
				} else if strings.Contains(itemText, ":") {
					// Generic validation format - assume it's a CEL expression
					validations = append(validations, itemText)
				}

				return ast.WalkSkipChildren, nil
			}
		}
		return ast.WalkContinue, nil
	})

	return validations
}

// buildFixtureFromCommand converts a commandBlockBuilder to a FixtureTest
func buildFixtureFromCommand(cmd *commandBlockBuilder, frontMatter *FrontMatter, sourceDir string) *FixtureNode {
	if cmd.name == "" || cmd.content == "" {
		return nil
	}
	exec := ExecFixtureBase{
		Exec: cmd.language,
		Args: []string{cmd.content},
	}

	switch cmd.language {
	case "bash", "sh":
		exec.Args = []string{"-c", cmd.content}
	case "shell":
		exec.Exec = "bash"
		exec.Args = []string{"-c", cmd.content}
	case "pwsh", "powershell":
		exec.Args = []string{"-Command", cmd.content}

	case "python", "python3":
		exec.Args = []string{"-c", cmd.content}

	case "typescript", "ts":
		exec.Exec = "ts-node"
		exec.Args = []string{"-e", cmd.content}
	case "javascript", "js":
		exec.Exec = "node"
		exec.Args = []string{"-e", cmd.content}

	// Add more languages as needed
	default:
		// For unrecognized languages, use as-is
	}
	// No special handling needed for Go

	fixture := FixtureTest{
		Name:            cmd.name,
		ExecFixtureBase: exec,
		SourceDir:       sourceDir,
		Expected: Expectations{
			Properties: make(map[string]interface{}),
		},
	}

	// Apply frontmatter from command block if present
	if cmd.frontmatter != "" {
		var cmdFrontMatter struct {
			CWD      string         `yaml:"cwd"`
			ExitCode *int           `yaml:"exitCode"`
			Env      map[string]any `yaml:"env"`
			Timeout  string         `yaml:"timeout"`
		}

		if err := yaml.Unmarshal([]byte(cmd.frontmatter), &cmdFrontMatter); err == nil {
			if cmdFrontMatter.CWD != "" {
				fixture.CWD = cmdFrontMatter.CWD
			}
			if cmdFrontMatter.ExitCode != nil {
				fixture.Expected.ExitCode = cmdFrontMatter.ExitCode
			}
			if cmdFrontMatter.Env != nil {
				fixture.Env = cmdFrontMatter.Env
			}
		}
	}

	// Apply file-level frontmatter if present
	if frontMatter != nil {
		// Assign the entire frontmatter to preserve metadata
		fixture.FrontMatter = *frontMatter

		if frontMatter.Exec != "" {
			fixture.Exec = frontMatter.Exec
		}
		if frontMatter.Build != "" {
			fixture.Build = frontMatter.Build
		}
		if frontMatter.Env != nil && fixture.Env == nil {
			fixture.Env = frontMatter.Env
		}
	}

	// Combine validations into CEL expression
	if len(cmd.validations) > 0 {
		if len(cmd.validations) == 1 {
			fixture.Expected.CEL = cmd.validations[0]
		} else {
			fixture.Expected.CEL = strings.Join(cmd.validations, " && ")
		}
	}

	cmd.isComplete = true

	return &FixtureNode{
		Type: TestNode,
		Test: &fixture,
	}
}

// parseTableFromAST parses table-based fixtures from AST (existing functionality)
func parseTableFromAST(tableNode *extast.Table, source []byte, frontMatter *FrontMatter, sourceDir string) ([]FixtureNode, error) {
	var fixtures []FixtureNode
	var headers []string

	// Walk through table rows
	for child := tableNode.FirstChild(); child != nil; child = child.NextSibling() {
		if tableHead, ok := child.(*extast.TableHeader); ok {
			// Extract headers
			for headerChild := tableHead.FirstChild(); headerChild != nil; headerChild = headerChild.NextSibling() {
				if cell, ok := headerChild.(*extast.TableCell); ok {
					headerText := extractNodeText(cell, source)
					headers = append(headers, strings.TrimSpace(headerText))
				}
			}
		} else if tableRow, ok := child.(*extast.TableRow); ok {
			// Extract row data
			var values []string
			for cellChild := tableRow.FirstChild(); cellChild != nil; cellChild = cellChild.NextSibling() {
				if cell, ok := cellChild.(*extast.TableCell); ok {
					cellText := extractNodeText(cell, source)
					values = append(values, strings.TrimSpace(cellText))
				}
			}

			// Create fixture from row
			if len(headers) > 0 && len(values) == len(headers) {
				if fixtureNode := parseTableRow(headers, values); fixtureNode != nil {
					// Apply frontmatter and source directory
					if fixtureNode.Test != nil {
						applyFrontMatterToFixture(fixtureNode.Test, frontMatter)
						fixtureNode.Test.SourceDir = sourceDir
					}
					fixtures = append(fixtures, *fixtureNode)
				}
			}
		}
	}

	return fixtures, nil
}

// applyFrontMatterToFixture applies frontmatter settings to a fixture
func applyFrontMatterToFixture(fixture *FixtureTest, frontMatter *FrontMatter) {
	if frontMatter == nil {
		return
	}

	if frontMatter.Build != "" {
		fixture.Build = frontMatter.Build
	}
	if frontMatter.Exec != "" {
		fixture.Exec = frontMatter.Exec
	}
	if frontMatter.Env != nil && fixture.Env == nil {
		fixture.Env = frontMatter.Env
	}
}
