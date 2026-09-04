package fixtures

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/goccy/go-yaml"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"

	"github.com/flanksource/gavel/fixtures/record"
)

// commandBlockBuilder helps build a command fixture from markdown AST
type commandBlockBuilder struct {
	name        string
	content     string
	language    string
	frontmatter string
	validations []string
	isComplete  bool
	origin      *FixtureOrigin
	config      *executableFenceConfig
}

type executableFenceConfig struct {
	Content    string                 `yaml:"content"`
	CWD        string                 `yaml:"cwd"`
	Env        map[string]any         `yaml:"env"`
	Terminal   string                 `yaml:"terminal"`
	Record     *record.Spec           `yaml:"record"`
	OS         string                 `yaml:"os"`
	Arch       string                 `yaml:"arch"`
	Skip       string                 `yaml:"skip"`
	ExitCode   *int                   `yaml:"exitCode"`
	Stdout     string                 `yaml:"stdout"`
	Stderr     string                 `yaml:"stderr"`
	Error      string                 `yaml:"error"`
	Format     string                 `yaml:"format"`
	Count      *int                   `yaml:"count"`
	Output     string                 `yaml:"output"`
	Timeout    string                 `yaml:"timeout"`
	CEL        string                 `yaml:"cel"`
	Properties map[string]interface{} `yaml:"properties"`
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
	var tableIndex int

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
				if err := appendCommandFixture(currentSection, standaloneCodeBlock, frontMatter, sourceDir); err != nil {
					return ast.WalkStop, err
				}
				standaloneCodeBlock = nil
			}

			// Check if this is a command heading (starts with "command:")
			isCommandHeading := strings.HasPrefix(strings.ToLower(headingText), "command:")

			// Complete previous command block if exists
			if currentCommand != nil && !currentCommand.isComplete {
				if err := appendCommandFixture(currentSection, currentCommand, frontMatter, sourceDir); err != nil {
					return ast.WalkStop, err
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
					origin: &FixtureOrigin{
						Kind:        "command",
						SectionPath: currentSection.GetSectionPath(),
						Line:        nodeStartLine(node, source),
					},
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
					Name:  headingText,
					Type:  SectionNode,
					Level: level,
					Origin: &FixtureOrigin{
						Kind:        "section",
						SectionPath: currentSection.GetSectionPath(),
						Line:        nodeStartLine(node, source),
					},
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

			if kind := runnerStepKind(infoString); kind != "" {
				// `yaml test` / `yaml lint` (or bare `test`/`lint`): the body is
				// engine options, run by the TestStepRunner/LintStepRunner hook.
				name := parentHeading
				if name == "" {
					name = kind + "-step"
				}
				test := &FixtureTest{
					Name:       name,
					SourceDir:  sourceDir,
					RunnerStep: &RunnerStepSpec{Kind: kind, Config: codeContent},
				}
				if frontMatter != nil {
					test.FrontMatter = *frontMatter
				}
				currentSection.AddChild(&FixtureNode{
					Name: name,
					Type: TestNode,
					Test: test,
					Origin: &FixtureOrigin{
						Kind:        "runner-step",
						SectionPath: currentSection.GetSectionPath(),
						Line:        nodeStartLine(node, source),
					},
					Children: make([]*FixtureNode, 0),
				})
			} else if inCommandBlock && currentCommand != nil {
				// Handle code blocks within command blocks (existing behavior)
				// Get allowed code blocks from frontmatter
				allowedBlocks := getCodeBlocksOrDefault(frontMatter)

				// Check if this language should be executed
				if shouldExecuteCodeBlock(lang, allowedBlocks) {
					currentCommand.language = lang
					currentCommand.content = codeContent
					if cfg, ok := parseExecutableFenceConfig(codeContent); ok {
						currentCommand.content = cfg.Content
						currentCommand.config = &cfg
					}
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
					if err := appendCommandFixture(currentSection, standaloneCodeBlock, frontMatter, sourceDir); err != nil {
						return ast.WalkStop, err
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
					origin: &FixtureOrigin{
						Kind:        "standalone-code",
						SectionPath: currentSection.GetSectionPath(),
						Line:        nodeStartLine(node, source),
					},
				}
				if cfg, ok := parseExecutableFenceConfig(codeContent); ok {
					standaloneCodeBlock.content = cfg.Content
					standaloneCodeBlock.config = &cfg
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
				if err := appendCommandFixture(currentSection, standaloneCodeBlock, frontMatter, sourceDir); err != nil {
					return ast.WalkStop, err
				}
				standaloneCodeBlock = nil
			}

		case *extast.Table:
			// Handle existing table format - add tests to current section
			if !inCommandBlock {
				tableIndex++
				tableNode, err := parseTableFromAST(node, source, frontMatter, sourceDir, tableIndex, currentSection.GetSectionPath())
				if err != nil {
					return ast.WalkStop, err
				}
				if tableNode != nil {
					currentSection.AddChild(tableNode)
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
		if err := appendCommandFixture(currentSection, standaloneCodeBlock, frontMatter, sourceDir); err != nil {
			return nil, err
		}
	}

	// Complete final command block if exists
	if currentCommand != nil && !currentCommand.isComplete {
		if err := appendCommandFixture(currentSection, currentCommand, frontMatter, sourceDir); err != nil {
			return nil, err
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

// appendCommandFixture builds a fixture from a completed command block and
// attaches it to section. Every completion site funnels through here so a
// rejected per-test frontmatter key fails the parse in one place.
func appendCommandFixture(section *FixtureNode, cmd *commandBlockBuilder, frontMatter *FrontMatter, sourceDir string) error {
	fixture, err := buildFixtureFromCommand(cmd, frontMatter, sourceDir)
	if err != nil || fixture == nil {
		return err
	}
	section.AddChild(&FixtureNode{
		Name:     fixture.Test.Name,
		Type:     TestNode,
		Test:     fixture.Test,
		Origin:   fixture.Origin,
		Children: make([]*FixtureNode, 0),
	})
	return nil
}

// buildFixtureFromCommand converts a commandBlockBuilder to a FixtureTest
func buildFixtureFromCommand(cmd *commandBlockBuilder, frontMatter *FrontMatter, sourceDir string) (*FixtureNode, error) {
	if cmd.name == "" || cmd.content == "" {
		return nil, nil
	}
	exec := ExecFixtureBase{
		Exec: cmd.language,
		Args: []string{cmd.content},
	}

	switch cmd.language {
	case "exec":
		exec.Exec = "bash"
		exec.Args = []string{"-c", cmd.content}
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
			Terminal string         `yaml:"terminal"`
			OS       string         `yaml:"os"`
			Arch     string         `yaml:"arch"`
			Skip     string         `yaml:"skip"`
			// Setup is bound only so it can be rejected. A setup is prepared once
			// per file and shared by every test in it, so a per-test block cannot be
			// honoured — and unknown keys here are otherwise dropped silently.
			Setup map[string]any `yaml:"setup"`
		}

		// `record:` is decoded on its own, ahead of the tolerant decode below that
		// drops every key when any one of them fails: a mistyped recorder must be a
		// red parse, not a fixture that silently records nothing.
		var recordOnly struct {
			Record *record.Spec `yaml:"record"`
		}
		if err := yaml.Unmarshal([]byte(cmd.frontmatter), &recordOnly); err != nil {
			return nil, fmt.Errorf("%s: %w", cmd.name, err)
		}
		if recordOnly.Record != nil {
			fixture.Record = recordOnly.Record
		}

		if err := yaml.Unmarshal([]byte(cmd.frontmatter), &cmdFrontMatter); err == nil {
			if cmdFrontMatter.Setup != nil {
				return nil, fmt.Errorf("%s: setup: is file-level frontmatter only, it cannot be set per-test", cmd.name)
			}
			if cmdFrontMatter.CWD != "" {
				fixture.CWD = cmdFrontMatter.CWD
			}
			if cmdFrontMatter.ExitCode != nil {
				fixture.Expected.ExitCode = cmdFrontMatter.ExitCode
			}
			if cmdFrontMatter.Env != nil {
				fixture.Env = cmdFrontMatter.Env
			}
			if cmdFrontMatter.Terminal != "" {
				fixture.Terminal = cmdFrontMatter.Terminal
			}
			if cmdFrontMatter.OS != "" {
				fixture.TestOS = cmdFrontMatter.OS
			}
			if cmdFrontMatter.Arch != "" {
				fixture.TestArch = cmdFrontMatter.Arch
			}
			if cmdFrontMatter.Skip != "" {
				fixture.TestSkip = cmdFrontMatter.Skip
			}
		}
	}

	if cmd.config != nil {
		applyExecutableFenceConfig(&fixture, *cmd.config)
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
		existingCEL := fixture.Expected.CEL
		nextCEL := ""
		if len(cmd.validations) == 1 {
			nextCEL = cmd.validations[0]
		} else {
			nextCEL = strings.Join(cmd.validations, " && ")
		}
		if existingCEL != "" {
			fixture.Expected.CEL = existingCEL + " && " + nextCEL
		} else {
			fixture.Expected.CEL = nextCEL
		}
	}

	cmd.isComplete = true

	return &FixtureNode{
		Type:   TestNode,
		Test:   &fixture,
		Origin: cmd.origin,
	}, nil
}

func parseExecutableFenceConfig(content string) (executableFenceConfig, bool) {
	var cfg executableFenceConfig
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return executableFenceConfig{}, false
	}
	if strings.TrimSpace(cfg.Content) == "" {
		return executableFenceConfig{}, false
	}
	return cfg, true
}

func applyExecutableFenceConfig(fixture *FixtureTest, cfg executableFenceConfig) {
	if cfg.CWD != "" {
		fixture.CWD = cfg.CWD
	}
	if cfg.Env != nil {
		fixture.Env = cfg.Env
	}
	if cfg.Terminal != "" {
		fixture.Terminal = cfg.Terminal
	}
	if cfg.Record != nil {
		fixture.Record = cfg.Record
	}
	if cfg.OS != "" {
		fixture.TestOS = cfg.OS
	}
	if cfg.Arch != "" {
		fixture.TestArch = cfg.Arch
	}
	if cfg.Skip != "" {
		fixture.TestSkip = cfg.Skip
	}
	if cfg.ExitCode != nil {
		fixture.Expected.ExitCode = cfg.ExitCode
	}
	if cfg.Stdout != "" {
		fixture.Expected.Stdout = cfg.Stdout
	}
	if cfg.Stderr != "" {
		fixture.Expected.Stderr = cfg.Stderr
	}
	if cfg.Error != "" {
		fixture.Expected.Error = cfg.Error
	}
	if cfg.Format != "" {
		fixture.Expected.Format = cfg.Format
	}
	if cfg.Count != nil {
		fixture.Expected.Count = cfg.Count
	}
	if cfg.Output != "" {
		fixture.Expected.Output = cfg.Output
	}
	if cfg.CEL != "" {
		fixture.Expected.CEL = cfg.CEL
	}
	if cfg.Timeout != "" {
		if timeout, err := time.ParseDuration(cfg.Timeout); err == nil {
			fixture.Expected.Timeout = &timeout
		}
	}
	if cfg.Properties != nil {
		fixture.Expected.Properties = cfg.Properties
	}
}

// parseTableFromAST parses table-based fixtures from AST (existing functionality)
func parseTableFromAST(tableAST *extast.Table, source []byte, frontMatter *FrontMatter, sourceDir string, tableIndex int, sectionPath string) (*FixtureNode, error) {
	tableFixtureNode := &FixtureNode{
		Name:     fmt.Sprintf("Table %d", tableIndex),
		Type:     TableNode,
		Children: make([]*FixtureNode, 0),
		Origin: &FixtureOrigin{
			Kind:        "table",
			SectionPath: sectionPath,
			TableIndex:  tableIndex,
			Line:        nodeStartLine(tableAST, source),
		},
	}
	var headers []string
	rowIndex := 0

	// Walk through table rows
	for child := tableAST.FirstChild(); child != nil; child = child.NextSibling() {
		if tableHead, ok := child.(*extast.TableHeader); ok {
			// Extract headers
			for headerChild := tableHead.FirstChild(); headerChild != nil; headerChild = headerChild.NextSibling() {
				if cell, ok := headerChild.(*extast.TableCell); ok {
					headerText := extractNodeText(cell, source)
					headers = append(headers, strings.TrimSpace(headerText))
				}
			}
		} else if tableRow, ok := child.(*extast.TableRow); ok {
			rowIndex++
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
				fixtureNode, err := parseTableRow(headers, values)
				if err != nil {
					return nil, err
				}
				if fixtureNode != nil {
					// Apply frontmatter and source directory
					if fixtureNode.Test != nil {
						applyFrontMatterToFixture(fixtureNode.Test, frontMatter)
						fixtureNode.Test.SourceDir = sourceDir
					}
					fixtureNode.Name = fixtureNode.Test.Name
					fixtureNode.Origin = &FixtureOrigin{
						Kind:        "table-row",
						SectionPath: sectionPath,
						TableIndex:  tableIndex,
						RowIndex:    rowIndex,
						Line:        nodeStartLine(tableRow, source),
					}
					tableFixtureNode.AddChild(fixtureNode)
				}
			}
		}
	}

	if len(tableFixtureNode.Children) == 0 {
		return nil, nil
	}

	return tableFixtureNode, nil
}

func nodeStartLine(node ast.Node, source []byte) int {
	lines := node.Lines()
	if lines == nil || lines.Len() == 0 {
		return 0
	}
	return lineNumberForOffset(source, lines.At(0).Start)
}

func lineNumberForOffset(source []byte, offset int) int {
	if offset < 0 {
		return 0
	}
	line := 1
	for i, b := range source {
		if i >= offset {
			return line
		}
		if b == '\n' {
			line++
		}
	}
	return line
}

// applyFrontMatterToFixture applies frontmatter settings to a fixture
func applyFrontMatterToFixture(fixture *FixtureTest, frontMatter *FrontMatter) {
	if frontMatter == nil {
		return
	}

	fixture.FrontMatter = *frontMatter

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
