package types

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/gavel/fixtures"
	"github.com/ghodss/yaml"
	"github.com/samber/lo"
)

type TODOS []*TODO

func (todos TODOS) Sort() {
	// Sort by priority (high → medium → low), then alphabetically by filename
	sort.Slice(todos, func(i, j int) bool {
		// Compare priorities
		pi := priorityOrder(todos[i].Priority)
		pj := priorityOrder(todos[j].Priority)

		if pi != pj {
			return pi < pj
		}

		// Same priority, sort alphabetically by filename
		nameI := filepath.Base(todos[i].FilePath)
		nameJ := filepath.Base(todos[j].FilePath)
		return nameI < nameJ
	})

}

func priorityOrder(p Priority) int {
	switch p {
	case PriorityHigh:
		return 0
	case PriorityMedium:
		return 1
	case PriorityLow:
		return 2
	default:
		return 999
	}
}

type Attempt struct {
	Status     Status
	Timestamp  time.Time
	Duration   time.Duration
	Cost       float64
	Tokens     int
	Model      string
	Commit     string
	Transcript string // relative path to transcript .md
}

// TODO represents a structured TODO item parsed from a markdown file.
// It combines fixture test nodes with TODO-specific metadata for tracking
// implementation tasks including reproduction steps, verification tests, and execution status.
type TODO struct {
	FilePath string                `json:"file_path,omitempty"`
	FileNode *fixtures.FixtureNode `json:"file_node,omitempty"` // Root file node from fixtures parser

	TODOFrontmatter `json:",inline"`

	// Sections are FixtureNodes with Tests
	StepsToReproduce []*fixtures.FixtureNode `json:"steps_to_reproduce,omitempty"` // Section containing reproduction steps
	// Plain text implementation instructions
	Implementation    string                  `json:"implementation,omitempty"`
	Verification      []*fixtures.FixtureNode `json:"verification,omitempty"`       // Section containing verification tests
	CustomValidations []*fixtures.FixtureNode `json:"custom_validations,omitempty"` // Section containing custom validation tests
}

func (todo TODO) AsYaml() (string, error) {
	// Serialize frontmatter
	frontmatterYAML, err := yaml.Marshal(&todo.TODOFrontmatter)
	if err != nil {
		return "", fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	// Create file content
	var content strings.Builder
	content.WriteString("---\n")
	content.WriteString(string(frontmatterYAML))
	content.WriteString("---\n")

	// Add markdown content if available
	if todo.Implementation != "" {
		content.WriteString("\n")
		content.WriteString(todo.Implementation)
	}
	return content.String(), nil
}

func (t TODO) PrettyRow(opts interface{}) map[string]api.Text {
	title := t.Title
	if title == "" {
		title = t.Filename()
	}
	row := map[string]api.Text{
		"Title":    clicky.Text(title, "order-1"),
		"Status":   t.Status.Pretty().Styles("order-2"),
		"Priority": t.Priority.Pretty().Styles("order-3"),
	}
	if t.LastRun != nil {
		row["Updated"] = clicky.Text("", "order-4").Append(time.Since(*t.LastRun), "text-muted")
	}
	return row
}

func (t TODO) Filename() string {
	file := filepath.Base(t.FilePath)
	return lo.PascalCase(file)
}

// Pretty returns a formatted text representation of the TODO item
func (t TODO) Pretty() api.Text {
	result := api.Text{}

	// Add title/file path
	if t.FilePath != "" {
		result = result.Add(icons.File).Append(" ", "").Append(t.Filename(), "text-blue-600 font-medium")
	}

	// Add priority if set
	if t.Priority != "" {
		result = result.Add(t.Priority.Pretty())
	}

	// Add status if set
	if t.Status != "" {
		result = result.Append(" ", "").Add(t.Status.Pretty())
	}

	return result
}

// PrettyDetailed returns a comprehensive formatted representation of the TODO with all sections and metadata.
func (t TODO) PrettyDetailed() api.Text {
	result := api.Text{}

	// Header line with filename, priority, status
	header := t.Pretty()
	result = result.Add(header).NewLine().NewLine()

	// File path
	result = result.Append("File: ", "text-gray-500").Append(t.FilePath, "").NewLine()

	// Language
	if t.Language != "" {
		result = result.Append("Language: ", "text-gray-500").Add(t.Language.Pretty()).NewLine()
	}

	// Attempts
	if t.Attempts > 0 {
		result = result.Append("Attempts: ", "text-gray-500").Append(fmt.Sprintf("%d", t.Attempts), "").NewLine()
	}

	// Last run
	if t.LastRun != nil {
		result = result.Append("Last Run: ", "text-gray-500").Append(t.LastRun.Format(time.RFC3339), "").NewLine()
	}

	result = result.NewLine()

	// Show fixture tests from the FileNode tree
	if t.FileNode != nil {
		result = result.Add(formatFixtureTree(t.FileNode, 0))
	}

	// LLM Configuration section
	if t.LLM != nil {
		result = result.Add(icons.Lambda).Append(" LLM Configuration", "text-blue-600 font-bold").NewLine()
		if t.LLM.Model != "" {
			result = result.Append("  Model: ", "text-gray-500").Append(t.LLM.Model, "").NewLine()
		}
		if t.LLM.MaxTokens > 0 {
			result = result.Append("  Max Tokens: ", "text-gray-500").Append(fmt.Sprintf("%d", t.LLM.MaxTokens), "").NewLine()
		}
		if t.LLM.SessionId != "" {
			result = result.Append("  Session ID: ", "text-gray-500").Append(t.LLM.SessionId, "").NewLine()
		}
	}

	return result
}

// formatFixtureTree recursively formats a fixture node tree
func formatFixtureTree(node *fixtures.FixtureNode, depth int) api.Text {
	result := clicky.Text("", "")

	// Skip the root file node itself
	if depth > 0 {
		indent := strings.Repeat("  ", depth-1)

		// Show section header
		if node.Name != "" && node.Type == fixtures.SectionNode {
			result = result.Append(indent, "").Add(icons.Folder).Append(" ", "").
				Append(node.Name, "text-blue-600 font-bold").NewLine()
		}

		// Show test
		if node.Test != nil {
			result = result.Append(indent, "").Append("  \u2514\u2500 ", "text-gray-500").
				Append(node.Test.Name, "").NewLine()
		}
	}

	// Recursively format children
	for _, child := range node.Children {
		result = result.Add(formatFixtureTree(child, depth+1))
	}

	return result
}

// StringOrSlice handles YAML fields that can be either a single string or a list of strings.
type StringOrSlice []string

func (s StringOrSlice) MarshalYAML() (interface{}, error) {
	if len(s) == 1 {
		return s[0], nil
	}
	return []string(s), nil
}

func (s *StringOrSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var single string
	if err := unmarshal(&single); err == nil {
		*s = StringOrSlice{single}
		return nil
	}
	var list []string
	if err := unmarshal(&list); err != nil {
		return err
	}
	*s = list
	return nil
}

func (s *StringOrSlice) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*s = StringOrSlice{single}
		return nil
	}
	var list []string
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	*s = list
	return nil
}

func (s StringOrSlice) MarshalJSON() ([]byte, error) {
	if len(s) == 1 {
		return json.Marshal(s[0])
	}
	return json.Marshal([]string(s))
}

type PR struct {
	Number        int    `yaml:"number,omitempty" json:"number,omitempty"`
	URL           string `yaml:"url,omitempty" json:"url,omitempty"`
	Head          string `yaml:"head,omitempty" json:"head,omitempty"`
	Base          string `yaml:"base,omitempty" json:"base,omitempty"`
	CommentID     int64  `yaml:"comment_id,omitempty" json:"comment_id,omitempty"`
	CommentAuthor string `yaml:"comment_author,omitempty" json:"comment_author,omitempty"`
	CommentURL    string `yaml:"comment_url,omitempty" json:"comment_url,omitempty"`
}

// TODOFrontmatter contains metadata for a TODO item parsed from YAML front-matter.
// It extends the standard fixtures.FrontMatter with TODO-specific fields like priority,
// status, and execution tracking.
type TODOFrontmatter struct {
	fixtures.FrontMatter `yaml:",inline" json:",inline"` // Embed standard fixture frontmatter

	// TODO-specific fields
	Title         string            `yaml:"title,omitempty" json:"title,omitempty"`
	Priority      Priority          `yaml:"priority,omitempty" json:"priority,omitempty"`
	Status        Status            `yaml:"status,omitempty" json:"status,omitempty"`
	LastRun       *time.Time        `yaml:"last_run,omitempty" json:"last_run,omitempty"`
	Attempts      int               `yaml:"attempts,omitempty" json:"attempts,omitempty"`
	Language      Language          `yaml:"language,omitempty" json:"language,omitempty"`
	WorkingCommit string            `yaml:"working_commit,omitempty" json:"working_commit,omitempty"`
	Branch        string            `yaml:"branch,omitempty" json:"branch,omitempty"`
	Path          StringOrSlice     `yaml:"path,omitempty" json:"path,omitempty"`
	LLM           *LLM              `yaml:"llm,omitempty" json:"llm,omitempty"`
	Verify        *TODOVerifyConfig `yaml:"verify,omitempty" json:"verify,omitempty"`
	PR            *PR               `yaml:"pr,omitempty" json:"pr,omitempty"`
	Prompt        string            `yaml:"prompt,omitempty" json:"prompt,omitempty"`
}

// CleanMetadata removes keys from Metadata that match struct field yaml tags.
// This fixes a goccy/go-yaml bug where inline maps capture ALL fields.
func (f *TODOFrontmatter) CleanMetadata() {
	// Clean embedded FrontMatter fields first
	f.FrontMatter.CleanMetadata()

	if f.Metadata == nil {
		return
	}
	// Keys from TODOFrontmatter
	delete(f.Metadata, "title")
	delete(f.Metadata, "priority")
	delete(f.Metadata, "status")
	delete(f.Metadata, "last_run")
	delete(f.Metadata, "attempts")
	delete(f.Metadata, "language")
	delete(f.Metadata, "path")
	delete(f.Metadata, "llm")
	delete(f.Metadata, "verify")
	delete(f.Metadata, "working_commit")
	delete(f.Metadata, "branch")
	delete(f.Metadata, "max_turns")
	delete(f.Metadata, "pr")
	delete(f.Metadata, "prompt")
}

// Pretty returns a formatted text representation of the TODOFrontmatter
func (f TODOFrontmatter) Pretty() api.Text {
	result := api.Text{}

	// Start with a simple metadata display
	result = result.Append("📋 TODO Metadata", "text-blue-600 font-bold")

	// Add metadata line
	var metadata []api.Text
	if f.Priority != "" {
		metadata = append(metadata, clicky.Text("Priority: ", "text-gray-500").Add(f.Priority.Pretty()))
	}
	if f.Status != "" {
		metadata = append(metadata, clicky.Text("Status: ", "text-gray-500").Add(f.Status.Pretty()))
	}
	if f.Language != "" {
		metadata = append(metadata, clicky.Text("Language: ", "text-gray-500").Add(f.Language.Pretty()))
	}
	if f.Attempts > 0 {
		metadata = append(metadata, clicky.Text("Attempts: ", "text-gray-500").Append(fmt.Sprintf("%d", f.Attempts), "text-orange-600"))
	}
	if f.LastRun != nil {
		metadata = append(metadata, clicky.Text("Last Run: ", "text-gray-500").Append(f.LastRun.Format("2006-01-02 15:04"), "text-purple-600"))
	}

	if len(metadata) > 0 {
		result = result.Append("\n", "")
		for i, meta := range metadata {
			if i > 0 {
				result = result.Append(" | ", "text-gray-400")
			}
			result = result.Add(meta)
		}
	}

	return result
}

// TODOVerifyConfig specifies gavel verify gate settings for a TODO.
type TODOVerifyConfig struct {
	Categories     []string `yaml:"categories,omitempty" json:"categories,omitempty"`
	ScoreThreshold int      `yaml:"score_threshold,omitempty" json:"score_threshold,omitempty"`
}

// LLM contains configuration and tracking for LLM usage when executing a TODO.
// It specifies model selection, cost/token limits, and records actual usage.
type LLM struct {
	// Model specifies which LLM model to use for running the todo (e.g. sonnet, haiku, gpt-4)
	Model string `yaml:"model" json:"model,omitempty"`
	// MaxTokens is the maximum number of tokens that can be used to complete this todo
	MaxTokens int `yaml:"max_tokens" json:"max_tokens,omitempty"`
	// MaxCost is the maximum cost in USD cents that can be incurred when running this todo
	MaxCost float64 `yaml:"max_cost" json:"max_cost,omitempty"`
	// TokensUsed records the actual tokens consumed, populated after running the todo
	TokensUsed int `yaml:"tokens_used,omitempty" json:"tokens_used,omitempty"`
	// CostIncurred records the actual cost in USD cents, populated after running the todo
	CostIncurred float64 `yaml:"cost_incurred,omitempty" json:"cost_incurred,omitempty"`
	// MaxTurns is the maximum number of conversation turns allowed
	MaxTurns int `yaml:"max_turns,omitempty" json:"max_turns,omitempty"`
	// Existing session ID for continuing conversations with the LLM
	SessionId string `yaml:"session_id,omitempty" json:"session_id,omitempty"`
}

// Pretty returns a formatted text representation of the LLM configuration
func (l LLM) Pretty() api.Text {
	result := clicky.Text("").Add(icons.Lambda).Append(" ", "").Append(l.Model, "text-blue-600 font-bold")

	var details []api.Text

	// Add token information
	if l.MaxTokens > 0 {
		tokenInfo := clicky.Text("Tokens: ", "text-gray-500")
		if l.TokensUsed > 0 {
			tokenInfo = tokenInfo.Append(fmt.Sprintf("%d/%d", l.TokensUsed, l.MaxTokens), "text-orange-600")
		} else {
			tokenInfo = tokenInfo.Append(fmt.Sprintf("max %d", l.MaxTokens), "text-blue-500")
		}
		details = append(details, tokenInfo)
	}

	// Add cost information
	if l.MaxCost > 0 {
		costInfo := clicky.Text("Cost: ", "text-gray-500")
		if l.CostIncurred > 0 {
			costInfo = costInfo.Append(fmt.Sprintf("$%.4f/$%.4f", l.CostIncurred, l.MaxCost), "text-red-600")
		} else {
			costInfo = costInfo.Append(fmt.Sprintf("max $%.4f", l.MaxCost), "text-green-600")
		}
		details = append(details, costInfo)
	}

	// Add session ID if present
	if l.SessionId != "" {
		details = append(details, clicky.Text("Session: ", "text-gray-500").Append(l.SessionId[:8]+"...", "text-purple-600"))
	}

	if len(details) > 0 {
		result = result.Append(" (", "text-gray-400")
		for i, detail := range details {
			if i > 0 {
				result = result.Append(", ", "text-gray-400")
			}
			result = result.Add(detail)
		}
		result = result.Append(")", "text-gray-400")
	}

	return result
}

// Priority indicates the urgency level of a TODO item.
type Priority string

const (
	// PriorityHigh indicates a critical or urgent TODO item.
	PriorityHigh Priority = "high"
	// PriorityMedium indicates a moderately important TODO item.
	PriorityMedium Priority = "medium"
	// PriorityLow indicates a TODO item with lower urgency.
	PriorityLow Priority = "low"
)

// Pretty returns a formatted text representation of the Priority with appropriate styling
func (p Priority) Pretty() api.Text {
	switch p {
	case PriorityHigh:
		return clicky.Text("").Add(icons.Error).Append(" HIGH", "text-red-600 font-bold")
	case PriorityMedium:
		return clicky.Text("").Add(icons.Warning).Append(" MEDIUM", "text-yellow-600 font-medium")
	case PriorityLow:
		return clicky.Text("").Add(icons.Pass).Append(" LOW", "text-green-600")
	default:
		return clicky.Text(string(p), "text-gray-500")
	}
}

// Status represents the current execution state of a TODO item.
type Status string

const (
	// StatusPending indicates the TODO has not been started.
	StatusPending Status = "pending"
	// StatusInProgress indicates the TODO is currently being worked on.
	StatusInProgress Status = "in_progress"
	// StatusCompleted indicates the TODO has been successfully completed.
	StatusCompleted Status = "completed"
	// StatusFailed indicates the TODO execution failed.
	StatusFailed Status = "failed"
	// StatusSkipped indicates the TODO was skipped because reproduction steps already pass.
	StatusSkipped Status = "skipped"
)

// Pretty returns a formatted text representation of the Status with color coding
func (s Status) Pretty() api.Text {
	switch s {
	case StatusPending:
		return clicky.Text("").Add(icons.Info).Append(" PENDING", "text-gray-600")
	case StatusInProgress:
		return clicky.Text("").Add(icons.ArrowRight).Append(" IN PROGRESS", "text-blue-600 font-medium")
	case StatusCompleted:
		return clicky.Text("").Add(icons.Pass).Append(" COMPLETED", "text-green-600 font-bold")
	case StatusFailed:
		return clicky.Text("").Add(icons.Fail).Append(" FAILED", "text-red-600 font-bold")
	case StatusSkipped:
		return clicky.Text("").Add(icons.Skip).Append(" SKIPPED", "text-yellow-600")
	default:
		return clicky.Text(string(s), "text-gray-500")
	}
}

// Language specifies the programming language for a TODO implementation.
type Language string

const (
	// LanguageGo indicates a Go implementation.
	LanguageGo Language = "go"
	// LanguageTypeScript indicates a TypeScript implementation.
	LanguageTypeScript Language = "typescript"
	// LanguagePython indicates a Python implementation.
	LanguagePython Language = "python"
)

// Pretty returns a formatted text representation of the Language with styling
func (l Language) Pretty() api.Text {
	switch l {
	case LanguageGo:
		return clicky.Text("").Add(icons.Package).Append(" Go", "text-blue-600 font-medium")
	case LanguageTypeScript:
		return clicky.Text("").Add(icons.File).Append(" TypeScript", "text-blue-500 font-medium")
	case LanguagePython:
		return clicky.Text("").Add(icons.Lambda).Append(" Python", "text-yellow-600 font-medium")
	default:
		return clicky.Text(string(l), "text-gray-500")
	}
}

// TestResultInfo captures comprehensive test execution context for appending to todo files.
type TestResultInfo struct {
	Command   string        // Full command that was run
	CWD       string        // Working directory
	GitBranch string        // Current git branch
	GitCommit string        // Current commit SHA (short)
	GitDirty  bool          // Whether there are uncommitted changes
	Timestamp time.Time     // When the test was run
	Passed    bool          // Overall pass/fail
	Output    string        // Test output (stdout/stderr combined, truncated)
	Duration  time.Duration // How long the test took
}

// Pretty returns a formatted text representation of the TestResultInfo for markdown output.
func (r TestResultInfo) Pretty() api.Text {
	t := api.Text{Content: "## Latest Failure"}.NewLine().NewLine()

	t = t.Add(api.KeyValuePair{Key: "Run", Value: r.Timestamp.Format(time.RFC3339)}).NewLine()

	if r.Command != "" {
		t = t.Add(api.KeyValuePair{Key: "Command", Value: "`" + r.Command + "`"}).NewLine()
	}

	t = t.Add(api.KeyValuePair{Key: "CWD", Value: "`" + r.CWD + "`"}).NewLine()

	if r.GitBranch != "" {
		t = t.Add(api.KeyValuePair{Key: "Branch", Value: "`" + r.GitBranch + "`"}).NewLine()
	}

	if r.GitCommit != "" {
		commitVal := "`" + r.GitCommit + "`"
		if r.GitDirty {
			commitVal += " (dirty)"
		}
		t = t.Add(api.KeyValuePair{Key: "Commit", Value: commitVal}).NewLine()
	}

	resultStr := "PASSED"
	if !r.Passed {
		resultStr = "FAILED"
	}
	t = t.Add(api.KeyValuePair{
		Key:   "Result",
		Value: fmt.Sprintf("%s (%s)", resultStr, r.Duration.Round(time.Millisecond)),
	}).NewLine()

	if r.Output != "" {
		t = t.NewLine().Add(api.Code{Content: r.Output})
	}

	return t
}

// BuildTestResultInfoOptions contains options for building TestResultInfo.
type BuildTestResultInfoOptions struct {
	CWD       string
	GitBranch string
	GitCommit string
	GitDirty  bool
	Timestamp time.Time
	Passed    bool
	Duration  time.Duration
}

// CheckResult contains the result of checking a single TODO by running its verification tests.
type CheckResult struct {
	TODO       *TODO
	Results    []fixtures.FixtureResult
	AllPassed  bool
	Duration   time.Duration
	Error      error
	TestResult *TestResultInfo // Comprehensive test result info for updating todo file
}

// Pretty returns a formatted text representation of the CheckResult
func (c CheckResult) Pretty() api.Text {
	result := c.TODO.Pretty()

	if c.Error != nil {
		return result.Append(" ", "").Add(icons.Fail).Append(fmt.Sprintf(" Error: %v", c.Error), "text-red-600")
	}

	testsRun := len(c.Results)
	passed := 0
	for _, r := range c.Results {
		if r.IsOK() {
			passed++
		}
	}

	if c.AllPassed {
		result = result.Append(" ", "").Add(icons.Pass).Append(fmt.Sprintf(" %d/%d tests passed", passed, testsRun), "text-green-600")
	} else {
		result = result.Append(" ", "").Add(icons.Fail).Append(fmt.Sprintf(" %d/%d tests passed", passed, testsRun), "text-red-600")
	}

	if c.Duration > 0 {
		result = result.Append(fmt.Sprintf(" (%s)", c.Duration.Round(time.Millisecond)), "text-gray-500")
	}

	// Add detailed test results
	result = result.NewLine()
	for _, testResult := range c.Results {
		result = result.Append("  ", "").Add(testResult.Pretty()).NewLine()
	}

	return result
}

// CountTests returns the total number of test nodes in a fixture node tree.
func CountTests(nodes []*fixtures.FixtureNode) int {
	count := 0
	for _, node := range nodes {
		count += countTestsRecursive(node)
	}
	return count
}

// countTestsRecursive recursively counts TestNode types in a fixture tree.
func countTestsRecursive(node *fixtures.FixtureNode) int {
	count := 0
	if node.Type == fixtures.TestNode {
		count = 1
	}
	for _, child := range node.Children {
		count += countTestsRecursive(child)
	}
	return count
}

// CollectTests recursively collects all test nodes from a fixture tree.
// Returns a flat slice of all FixtureNode pointers that have Test populated.
func CollectTests(node *fixtures.FixtureNode) []*fixtures.FixtureNode {
	var tests []*fixtures.FixtureNode
	collectTestsRecursive(node, &tests)
	return tests
}

// collectTestsRecursive helper for CollectTests.
func collectTestsRecursive(node *fixtures.FixtureNode, tests *[]*fixtures.FixtureNode) {
	if node.Test != nil {
		*tests = append(*tests, node)
	}
	for _, child := range node.Children {
		collectTestsRecursive(child, tests)
	}
}
