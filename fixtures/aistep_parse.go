package fixtures

import (
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/ai"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// AIStepRunner executes an AI verification fixture and returns its result. It is
// a package-level hook so the fixtures package needn't import verify (which would
// cycle via todos/types). The fixtures/types package registers the
// implementation in its init().
var AIStepRunner func(fixture FixtureTest, opts RunOptions) FixtureResult

// TestStepRunner / LintStepRunner execute a `yaml test` / `yaml lint` fixture
// step: the raw YAML body is unmarshalled onto testrunner.RunOptions /
// lint.Options and the engine is run. They are package-level hooks so the
// fixtures package needn't import testrunner or the lint package (which would
// cycle: testrunner imports fixtures). The fixtures/types package registers the
// implementations in its init().
var (
	TestStepRunner func(fixture FixtureTest, opts RunOptions) FixtureResult
	LintStepRunner func(fixture FixtureTest, opts RunOptions) FixtureResult
)

// FixtureAIConfig is the `ai:` front-matter block. Its presence marks a fixture
// file as an AI verification step; the fields map onto ai.AgentConfig (resolved
// in aistep.go to avoid coupling this package to the ai package).
type FixtureAIConfig struct {
	Model         string        `yaml:"model,omitempty" json:"model,omitempty"`
	Temperature   float64       `yaml:"temperature,omitempty" json:"temperature,omitempty"`
	MaxTokens     int           `yaml:"maxTokens,omitempty" json:"maxTokens,omitempty"`
	MaxConcurrent int           `yaml:"maxConcurrent,omitempty" json:"maxConcurrent,omitempty"`
	CacheTTL      time.Duration `yaml:"cacheTTL,omitempty" json:"cacheTTL,omitempty"`
	NoCache       bool          `yaml:"noCache,omitempty" json:"noCache,omitempty"`
}

// ToAgentConfig maps the `ai:` front matter onto an ai.AgentConfig, falling back
// to defaultModel when no model is set. A nil receiver yields the default config
// (so callers can pass an absent block directly).
// The model stays a string on the wire — that is the fixture file's format — but
// it lands in a structured api.Model, so `model: agent:sol:high` now carries its
// backend and effort through instead of being flattened to a name and re-guessed.
// The selector is resolved by ai.NewProvider.
func (c *FixtureAIConfig) ToAgentConfig(defaultModel string) ai.AgentConfig {
	cfg := ai.AgentConfig{
		Model:         api.Model{Name: defaultModel},
		Budget:        api.Budget{MaxTokens: 10000},
		MaxConcurrent: 4,
	}
	if c == nil {
		return cfg
	}
	if c.Model != "" {
		cfg.Model.Name = c.Model
	}
	if c.Temperature != 0 {
		t := c.Temperature
		cfg.Model.Temperature = &t
	}
	if c.MaxTokens > 0 {
		cfg.Budget.MaxTokens = c.MaxTokens
	}
	if c.MaxConcurrent > 0 {
		cfg.MaxConcurrent = c.MaxConcurrent
	}
	if c.CacheTTL > 0 {
		cfg.CacheTTL = c.CacheTTL
	}
	cfg.NoCache = c.NoCache
	return cfg
}

// FixtureVerifyConfig is the `verify:` front-matter block: the review scope
// (default: working-tree diff), the pass threshold (default 80), and check IDs
// to disable for this step.
type FixtureVerifyConfig struct {
	Scope     string   `yaml:"scope,omitempty" json:"scope,omitempty"`
	Threshold int      `yaml:"threshold,omitempty" json:"threshold,omitempty"`
	Disabled  []string `yaml:"disabled,omitempty" json:"disabled,omitempty"`
}

// ChecklistItem is one GitHub-style task-list entry from the document body. Each
// item becomes one scored acceptance criterion.
type ChecklistItem struct {
	Text    string `json:"text"`
	Checked bool   `json:"checked"`
}

// ChecklistResult is the per-criterion verdict an AI step returns: whether the
// change satisfies the criterion, with a one-line justification. It is the
// ai-step's structured output — one entry per ChecklistItem — and is surfaced to
// the definition-of-done predicate as the `results.checklist` list.
type ChecklistResult struct {
	Item    string `json:"item" description:"The acceptance-criterion text, echoed verbatim."`
	Passed  bool   `json:"passed" description:"True only when the change clearly satisfies this criterion."`
	Message string `json:"message,omitempty" description:"One-line justification: the evidence for a pass or the gap for a fail."`
}

// AIStepSpec carries the parsed inputs for an AI verification step: the optional
// custom reviewer instructions (the ```prompt block body), the document prose
// description, and the acceptance-criteria checklist.
type AIStepSpec struct {
	Prompt      string          `json:"prompt,omitempty"`
	Description string          `json:"description,omitempty"`
	Criteria    []ChecklistItem `json:"criteria,omitempty"`
}

// ExtractChecklist returns the GitHub-style task-list items in a markdown body,
// regardless of which section or list they appear in. The checkbox state is
// captured; the item text excludes the box itself.
func ExtractChecklist(content string) []ChecklistItem {
	md := goldmark.New(goldmark.WithExtensions(extension.TaskList))
	source := []byte(content)
	doc := md.Parser().Parse(text.NewReader(source))

	var items []ChecklistItem
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		li, ok := n.(*ast.ListItem)
		if !ok {
			return ast.WalkContinue, nil
		}
		cb := findTaskCheckBox(li)
		if cb == nil {
			return ast.WalkContinue, nil // a plain bullet, not a task item
		}
		itemText := strings.TrimSpace(extractNodeText(li, source))
		if itemText == "" {
			return ast.WalkSkipChildren, nil
		}
		items = append(items, ChecklistItem{Text: itemText, Checked: cb.IsChecked})
		return ast.WalkSkipChildren, nil
	})
	return items
}

func findTaskCheckBox(n ast.Node) *extast.TaskCheckBox {
	var found *extast.TaskCheckBox
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if cb, ok := c.(*extast.TaskCheckBox); ok {
			found = cb
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return found
}

// parseAIFixtureTree builds the fixture tree for an AI verification file: one AI
// step whose criteria are the document checklist, with the first ```prompt block
// (optional) as custom reviewer instructions and the top-level prose as the
// description. Validation bullets (cel:/contains:/regex:) become the step's
// expectations so CEL can assert over the JSON verify result.
func parseAIFixtureTree(content string, frontMatter *FrontMatter, sourceDir string) (*FixtureNode, error) {
	root := &FixtureNode{Name: "Content", Type: SectionNode, Children: make([]*FixtureNode, 0)}

	md := goldmark.New(
		goldmark.WithExtensions(extension.Table, extension.TaskList),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)
	source := []byte(content)
	doc := md.Parser().Parse(text.NewReader(source))

	var name, promptBody string
	var descParts, validations []string

	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch node := n.(type) {
		case *ast.Heading:
			if name == "" {
				name = strings.TrimSpace(extractNodeText(node, source))
			}
		case *ast.Paragraph:
			if _, topLevel := node.Parent().(*ast.Document); topLevel {
				if txt := strings.TrimSpace(extractNodeText(node, source)); txt != "" {
					descParts = append(descParts, txt)
				}
			}
		case *ast.FencedCodeBlock:
			var info string
			if node.Info != nil {
				info = string(node.Info.Segment.Value(source))
			}
			switch strings.ToLower(extractLanguage(info)) {
			case "prompt", "ai":
				if promptBody == "" {
					promptBody = extractCodeBlockContent(&node.BaseBlock, source)
				}
			}
		case *ast.List:
			listText := extractNodeText(node, source)
			if strings.Contains(listText, "cel:") || strings.Contains(listText, "contains:") || strings.Contains(listText, "regex:") {
				validations = append(validations, extractValidationsFromList(node, source)...)
			}
		}
		return ast.WalkContinue, nil
	})

	if name == "" {
		name = "verify"
	}

	fixture := FixtureTest{
		Name:      name,
		SourceDir: sourceDir,
		Expected:  Expectations{Properties: make(map[string]interface{})},
		AIStep: &AIStepSpec{
			Prompt:      promptBody,
			Description: strings.Join(descParts, "\n\n"),
			Criteria:    ExtractChecklist(content),
		},
	}
	if frontMatter != nil {
		fixture.FrontMatter = *frontMatter
	}
	switch len(validations) {
	case 0:
	case 1:
		fixture.Expected.CEL = validations[0]
	default:
		fixture.Expected.CEL = strings.Join(validations, " && ")
	}

	root.AddChild(&FixtureNode{
		Name:     fixture.Name,
		Type:     TestNode,
		Test:     &fixture,
		Children: make([]*FixtureNode, 0),
		Origin: &FixtureOrigin{
			Kind:        "ai-step",
			SectionPath: "",
		},
	})
	return root, nil
}
