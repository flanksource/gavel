package fixtures

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gomplate/v3"
)

// FixtureTest represents a single test case parsed from a markdown table.
// It contains the test definition including commands, queries, expected results,
// and environment configuration from both the table and file front-matter.
type FixtureTest struct {
	FrontMatter     `json:"frontmatter,omitempty"`
	ExecFixtureBase `json:",inline"`

	// Name of the test to be displayed in reports
	Name string `json:"name,omitempty"`
	// The working directory for executing the test
	SourceDir string       `json:"source_dir,omitempty"`
	Query     string       `json:"query,omitempty"`
	Expected  Expectations `json:"expected,omitempty"`

	TemplateVars map[string]any           `json:"template_vars,omitempty"` // Template variables (.file, .filename, .dir)
	TempFiles    map[string]*TempFileInfo `json:"temp_files,omitempty"`
}

func NewFixtureTest(opts RunOptions) (*FixtureTest, error) {

	f := FixtureTest{
		TemplateVars: map[string]any{
			"executablePath": opts.ExecutablePath,
			"workDir":        opts.WorkDir,
		},
		TempFiles: map[string]*TempFileInfo{},
	}

	// Create temp files if needed

	if tempFileData, ok := opts.ExtraArgs["temp_files"].(map[string]interface{}); ok {
		for name, content := range tempFileData {
			tempFile, err := createTempFile(name, fmt.Sprint(content))
			if err != nil {
				return nil, fmt.Errorf("failed to create temp file for fixture: %w", err)
			}

			f.TempFiles[name] = tempFile

		}
	}
	return &f, nil
}

func (f *FixtureTest) Cleanup() {
	for _, tempFile := range f.TempFiles {
		defer os.Remove(tempFile.Path)
	}
}

func (fixture FixtureTest) AsMap() map[string]any {
	m := map[string]any{
		"name":         fixture.Name,
		"sourceDir":    fixture.SourceDir,
		"query":        fixture.Query,
		"expectations": fixture.Expected,
	}
	for k, v := range fixture.ExecBase().AsMap() {
		m[k] = v
	}
	for k, v := range fixture.TemplateVars {
		m[k] = v
	}
	return m
}

func (fixture FixtureTest) String() string {
	return fixture.Pretty().String()
}

func (fixture FixtureTest) ExecBase() ExecFixtureBase {
	return fixture.FrontMatter.MergeInto(fixture.ExecFixtureBase)
}

func (fixture FixtureTest) Pretty() api.Text {
	return clicky.Text(fixture.Name, "italic text-orange-500")
}

type ExecFixtureBase struct {
	// Build command to run before tests
	Build string `yaml:"build,omitempty" json:"build,omitempty"`
	// Exec command to run for each test, can be overridden per test, defaults to running bash -c
	Exec string `yaml:"exec,omitempty" json:"exec,omitempty"`
	// Args are the base arguments to pass to the command
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`
	// Env defines environment variables for the tests
	Env map[string]any `yaml:"env,omitempty" json:"env,omitempty"`
	// Working directory for executing tests
	CWD string `yaml:"cwd,omitempty" json:"cwd,omitempty"`
}

func relativePath(path string) string {
	wd, _ := os.Getwd()
	rel, _ := filepath.Rel(wd, path)
	if rel == "." {
		return ""
	}
	return rel
}

func (e ExecFixtureBase) Pretty() api.Text {
	t := clicky.Text(e.Exec, "font-bold").Append(strings.Join(e.Args, " "), "wrap-space")
	if len(e.Env) > 0 {
		t = t.Append(" ").Append(clicky.Map(e.Env))
	}
	rel := relativePath(e.CWD)
	if rel != "" {
		t = t.Append(fmt.Sprintf(" (cwd: %s)", rel), "text-yellow-600")
	}

	return t
}

func (e ExecFixtureBase) IsEmpty() bool {
	return e.Exec == "" && e.Build == "" && len(e.Args) == 0
}

func (e ExecFixtureBase) Template(data map[string]any) (ExecFixtureBase, error) {
	var err error
	if e.Exec, err = gomplate.RunTemplate(data, gomplate.Template{
		Template: e.Exec,
	}); err != nil {
		return ExecFixtureBase{}, err
	}

	if e.Build, err = gomplate.RunTemplate(data, gomplate.Template{
		Template: e.Build,
	}); err != nil {
		return ExecFixtureBase{}, err
	}

	for i := range e.Args {
		e.Args[i], err = gomplate.RunTemplate(data, gomplate.Template{
			Template: e.Args[i],
		})
		if err != nil {
			return ExecFixtureBase{}, err
		}
	}
	return e, nil
}

func (e ExecFixtureBase) AsMap() map[string]string {
	envMap := map[string]string{}
	if e.CWD != "" {
		envMap["CWD"] = e.CWD
	}
	if e.Build != "" {
		envMap["BUILD"] = e.Build
	}
	if e.Exec != "" {
		envMap["EXEC"] = e.Exec
	}
	for k, v := range e.Env {
		envMap[k] = fmt.Sprintf("%v", v)
	}
	return envMap
}

func (e ExecFixtureBase) String() string {
	return e.Pretty().String()
}

func (e ExecFixtureBase) MergeInto(other ExecFixtureBase) ExecFixtureBase {
	merged := e
	if other.Build != "" {
		merged.Build = other.Build
	}
	if other.Exec != "" {
		merged.Exec = other.Exec
	}
	if other.CWD != "" {
		merged.CWD = other.CWD
	}
	if len(other.Args) > 0 {
		merged.Args = other.Args
	}
	if merged.Env == nil {
		merged.Env = make(map[string]any)
	}
	for k, v := range other.Env {
		merged.Env[k] = v
	}

	if merged.CWD == "" {
		merged.CWD, _ = os.Getwd()
	}

	if merged.Exec == "" {
		merged.Exec = "bash"
	}

	return merged
}

// FrontMatter represents the YAML front-matter section in markdown fixture files.
// It defines global configuration for all tests in the file including build commands,
// environment variables, and execution settings.
type FrontMatter struct {
	ExecFixtureBase `yaml:",inline" json:",inline"`

	Files string `yaml:"files,omitempty" json:"files,omitempty"` // Glob pattern to match files

	// CodeBlocks specifies which code block languages to execute (defaults to ["bash"])
	CodeBlocks []string `yaml:"codeBlocks,omitempty" json:"codeBlocks,omitempty"`

	// Total timeout for test execution
	Timeout *time.Duration `yaml:"timeout,omitempty" json:"timeout,omitempty"`

	Metadata map[string]interface{} `yaml:",inline" json:"metadata,omitempty"`
}

// CleanMetadata removes keys from Metadata that match struct field yaml tags.
// This fixes a goccy/go-yaml bug where inline maps capture ALL fields.
func (f *FrontMatter) CleanMetadata() {
	if f.Metadata == nil {
		return
	}
	// Keys from ExecFixtureBase
	delete(f.Metadata, "build")
	delete(f.Metadata, "exec")
	delete(f.Metadata, "args")
	delete(f.Metadata, "env")
	delete(f.Metadata, "cwd")
	// Keys from FrontMatter itself
	delete(f.Metadata, "files")
	delete(f.Metadata, "codeBlocks")
	delete(f.Metadata, "timeout")
}

// TODO: Register custom renderer for status icons when clicky supports it
// For now, the pretty tags with color mapping should handle this

// Text is a type alias for api.Text used for rendering formatted text output.
type Text api.Text

// NodeType represents the type of fixture node in the hierarchical tree structure.
// It distinguishes between file-level, section-level, and individual test nodes.
type NodeType int

const (
	// FileNode represents a markdown file containing fixture tests.
	FileNode NodeType = iota
	// SectionNode represents a section or subsection within a fixture file.
	SectionNode
	// TestNode represents an individual test case within a section.
	TestNode
)

func (nt NodeType) Pretty() api.Text {
	return clicky.Text(nt.String(), "text-gray-500")
}

// FixtureResult represents the outcome of executing a single fixture test.
// It contains core information, execution results, and metadata about the test run.
type FixtureResult struct {
	// Core fields
	Name     string        `json:"name" pretty:"label=Test Name,style=text-blue-600"`
	Type     string        `json:"type,omitempty" pretty:"label=Type,style=text-gray-500"`
	Status   task.Status   `json:"status,omitempty" `
	Duration time.Duration `json:"duration,omitempty" pretty:"label=Duration,style=text-yellow-600,omitempty"`
	Test     FixtureTest   `json:"-"` // Only populated for Test nodes

	// Result data
	Error     string      `json:"error,omitempty" pretty:"label=Error,style=text-red-600,omitempty"`
	Expected  interface{} `json:"expected,omitempty" pretty:"label=Expected,omitempty"`
	Actual    interface{} `json:"actual,omitempty" pretty:"label=Actual,omitempty"`
	CELResult bool        `json:"cel_result,omitempty" pretty:"label=CEL Result,omitempty"`

	// Execution metadata
	Command  string                 `json:"command,omitempty" pretty:"label=Command,style=text-cyan-600,omitempty"`
	CWD      string                 `json:"cwd,omitempty" pretty:"label=Working Dir,style=text-purple-500,omitempty"`
	Stdout   string                 `json:"stdout,omitempty" pretty:"label=Stdout,omitempty"`
	Stderr   string                 `json:"stderr,omitempty" pretty:"label=Stderr,omitempty"`
	ExitCode int                    `json:"exit_code,omitempty" pretty:"label=Exit Code,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty" pretty:"label=Metadata,omitempty"`
	Start    *time.Time             `json:"start,omitempty" pretty:"label=Start Time,omitempty"`
}

func (f FixtureResult) Failf(format string, args ...interface{}) FixtureResult {
	if f.Start != nil {
		f.Duration = time.Since(*f.Start)
	}
	f.Status = task.StatusFAIL
	f.Error = fmt.Sprintf(format, args...)
	return f
}

func (f FixtureResult) Errorf(err error, format string, args ...interface{}) FixtureResult {
	if f.Start != nil {
		f.Duration = time.Since(*f.Start)
	}
	f.Status = task.StatusERR
	f.Error = err.Error() + ": " + fmt.Sprintf(format, args...)
	return f
}

func (f FixtureResult) Stats() Stats {
	switch f.Status {
	case task.StatusFailed:
		return Stats{Failed: 1, Total: 1}
	case task.StatusPASS, task.StatusSuccess:
		return Stats{Passed: 1, Total: 1}
	case task.StatusSKIP:
		return Stats{Skipped: 1, Total: 1}
	case task.StatusERR, task.StatusCancelled:
		return Stats{Error: 1, Total: 1}
	default:
		return Stats{}
	}
}

func (f FixtureResult) String() string {
	return fmt.Sprintf("%s - %s", f.Test.Name, f.Status.String())
}

func (f FixtureResult) Pretty() api.Text {
	t := f.Status.Pretty().Append(" ").Add(f.Test.Pretty())

	if f.Error != "" {
		t = t.Space().Append(f.Error, "text-red-600")
	}
	t = t.Space()

	// For failures, show full output including stdout/stderr
	isFailed := f.Status == task.StatusFAIL || f.Status == task.StatusERR || f.Status == task.StatusFailed
	if f.Actual != nil {
		if isFailed {
			if full, ok := f.Actual.(api.PrettyFull); ok {
				t = t.Append(full.PrettyFull())
			} else {
				t = t.Append(f.Actual)
			}
		} else {
			t = t.Append(f.Actual)
		}
	}

	return t
}

// Out returns the combined stderr and stdout output
func (f FixtureResult) Out() string {
	return f.Stderr + f.Stdout
}

func (f FixtureResult) IsOK() bool {
	return f.Status.Health() == task.HealthOK
}

// Visitor defines an interface for visiting fixture test results.
type Visitor interface {
	Visit(test *FixtureResult)
}

// Stats provides summary statistics for fixture test execution.
// It tracks the total number of tests and their outcomes.
type Stats struct {
	Total   int `json:"total,omitempty"`
	Passed  int `json:"passed,omitempty"`
	Failed  int `json:"failed,omitempty"`
	Skipped int `json:"skipped,omitempty"`
	Error   int `json:"error,omitempty"`
}

func (s Stats) Merge(o Stats) Stats {
	return Stats{
		Total:   s.Total + o.Total,
		Passed:  s.Passed + o.Passed,
		Failed:  s.Failed + o.Failed,
		Skipped: s.Skipped + o.Skipped,
		Error:   s.Error + o.Error,
	}
}

func (s Stats) Add(result *FixtureResult) Stats {
	if result == nil {
		return s
	}
	s.Total++
	switch result.Status {
	case task.StatusFailed:
		s.Failed++
	case task.StatusPASS, task.StatusSuccess:
		s.Passed++
	case task.StatusSKIP:
		s.Skipped++
	case task.StatusERR, task.StatusCancelled:
		s.Error++
	}
	return s
}

func (f FixtureNode) GetStats() Stats {
	s := Stats{}

	s = s.Add(f.Results)
	for _, child := range f.Children {
		s = s.Merge(child.GetStats())
	}
	return s
}

// UpdateStats calculates and updates the Stats field for this node
func (fn *FixtureNode) UpdateStats() {
	stats := fn.GetStats()
	fn.Stats = &stats
}

func (s Stats) IsOK() bool {
	return s.Failed == 0 && s.Error == 0
}

func (s Stats) HasFailures() bool {
	return s.Failed > 0 || s.Error > 0
}

func (f *FixtureNode) AddFileNode(path string) *FixtureNode {
	node := &FixtureNode{
		Name:   path,
		Type:   FileNode,
		Parent: f,
	}
	f.Children = append(f.Children, node)
	return node
}

// Pretty prints status, with green for passed red for failed and yellow for skipped
func (s Stats) Pretty() api.Text {

	t := api.Text{}
	if s.Passed > 0 {
		t = t.Append(strconv.Itoa(s.Passed), "text-green-500")
	}
	if s.Failed > 0 {
		if !t.IsEmpty() {
			t = t.Append("/", "text-gray-500")
		}
		t = t.Append(strconv.Itoa(s.Failed), "text-red-500")

	}
	if s.Skipped > 0 {
		t = t.Append(fmt.Sprintf(" %d skipped", s.Skipped), "text-yellow-500")
	}
	if s.Error > 0 {
		t = t.Append(fmt.Sprintf("%d errors", s.Error), "text-red-500")
	}
	return t
}

func (s Stats) String() string {
	if s.Total == 0 {
		return "-"
	}
	str := fmt.Sprintf("%d/%d", s.Passed, s.Failed+s.Passed)

	if s.Skipped > 0 {
		str += fmt.Sprintf(" %d skipped", s.Skipped)
	}

	if s.Error > 0 {
		str += fmt.Sprintf(" %d error", s.Error)
	}

	return str
}

func (s *Stats) Visit(node *FixtureNode) {
	test := node.Results
	if test == nil {
		return
	}
	s.Total++

	switch test.Status {
	case task.StatusFailed, task.StatusERR, task.StatusCancelled:
		s.Failed++
	case task.StatusPASS, task.StatusSuccess:
		s.Passed++
	case task.StatusSKIP:
		s.Skipped++
	}
}

func (s *Stats) Health() task.Health {
	if s.Failed+s.Error > 0 {
		return task.HealthError
	}
	if s.Total == 0 || s.Skipped > 0 {
		return task.HealthWarning
	}
	return task.HealthOK
}

// String returns a string representation of NodeType
func (nt NodeType) String() string {
	switch nt {
	case FileNode:
		return "file"
	case SectionNode:
		return "section"
	case TestNode:
		return "test"
	default:
		return "unknown"
	}
}

// FixtureNode represents a node in the hierarchical fixture tree structure.
// Nodes can represent files, sections, or individual test cases and form a tree hierarchy.
type FixtureNode struct {
	Name     string         `json:"name" pretty:"label"` // Node name (file, section, or test)
	Type     NodeType       `json:"type" pretty:"type"`  // File, Section, or Test
	Level    int            `json:"level,omitempty"`     // Nesting level (0=file, 1=section, 2=subsection, etc.)
	Children []*FixtureNode `json:"children,omitempty" `
	Parent   *FixtureNode   `json:"-"`                 // Parent node reference
	Test     *FixtureTest   `json:"test,omitempty" `   // Only populated for Test nodes
	Results  *FixtureResult `json:"results,omitempty"` // Only populated after execution
	Stats    *Stats         `json:"stats,omitempty"`   // Aggregated statistics for sections/files
}

func (f FixtureNode) Pretty() api.Text {
	if f.Results != nil {
		return f.Results.Pretty()
	}

	s := clicky.Text("").Add(f.Type.Pretty()).Append(" ").Append(f.Name)

	if f.Stats != nil {
		s = s.Append(" (").Add(f.Stats.Pretty()).Append(")")
	}

	return s
}

func (f FixtureNode) GetChildren() []api.TreeNode {
	nodes := make([]api.TreeNode, len(f.Children))
	for i, child := range f.Children {
		nodes[i] = child.Tree()
	}
	return nodes
}

// FixtureTree represents the complete hierarchical structure of parsed fixtures.
// It contains root nodes for each file and aggregated statistics across all tests.
type FixtureTree struct {
	Root     []*FixtureNode   `json:"root"`
	AllTests []*FixtureResult `json:"all_tests"`
	Stats    *Stats           `json:"stats"`
}

// NewFixtureTree creates a new fixture tree
func NewFixtureTree() *FixtureTree {
	return &FixtureTree{
		Root:     make([]*FixtureNode, 0),
		AllTests: make([]*FixtureResult, 0),
		Stats:    &Stats{},
	}
}

// AddFileNode adds a file node to the tree root
func (ft *FixtureTree) AddFileNode(name, path string) *FixtureNode {
	node := &FixtureNode{
		Name: name,
		Type: FileNode,
	}
	ft.Root = append(ft.Root, node)
	return node
}

// BuildAllTestsList builds the AllTests slice from the tree structure
func (ft *FixtureTree) BuildAllTestsList() {
	ft.AllTests = make([]*FixtureResult, 0)
	for _, root := range ft.Root {
		ft.AllTests = append(ft.AllTests, root.GetAllTests()...)
	}
}

// AddChild adds a child node to this node
func (fn *FixtureNode) AddChild(child *FixtureNode) {
	child.Parent = fn
	fn.Children = append(fn.Children, child)
}

// GetSectionPath returns the full section path (e.g., "File > Section > Subsection")
func (fn *FixtureNode) GetSectionPath() string {
	if fn.Parent == nil {
		return fn.Name
	}
	if fn.Parent.Type == FileNode {
		return fn.Name
	}
	return fn.Parent.GetSectionPath() + " > " + fn.Name
}

// TreeMixin interface implementation
// Tree returns a TreeNode representation of this FixtureNode
func (fn *FixtureNode) Tree() api.TreeNode {
	return &FixtureTreeNode{fixture: fn}
}

// FixtureTreeNode is an internal TreeNode implementation for FixtureNode
type FixtureTreeNode struct {
	fixture *FixtureNode
}

func (ftn FixtureTreeNode) Pretty() api.Text {
	if ftn.fixture.Results != nil {
		return ftn.fixture.Results.Pretty()
	}

	// Get icon based on node type and status
	var icon string
	switch ftn.fixture.Type {
	case FileNode:
		icon = "📁"
	case SectionNode:
		icon = "📂"
	case TestNode:
		if ftn.fixture.Results != nil {
			icon = ftn.fixture.Results.Status.Icon()
		} else {
			icon = "📄"
		}
	default:
		icon = "📄"
	}

	content := ftn.fixture.Name

	// Add icon to content
	if icon != "" {
		content = fmt.Sprintf("%s %s", icon, content)
	}

	// Add stats for section/file nodes (not individual tests)
	if ftn.fixture.Type != TestNode && ftn.fixture.Stats != nil && ftn.fixture.Stats.Total > 0 {
		content = fmt.Sprintf("%s (%d/%d passed)",
			content, ftn.fixture.Stats.Passed, ftn.fixture.Stats.Total)
	}

	// Add test details if available
	if ftn.fixture.Type == TestNode && ftn.fixture.Results != nil {
		if ftn.fixture.Results.Duration > 0 {
			content = fmt.Sprintf("%s (%s)", content, ftn.fixture.Results.Duration)
		}
	}

	// Get style
	var style string
	if ftn.fixture.Results != nil {
		style = ftn.fixture.Results.Status.Style()
	} else if ftn.fixture.Stats != nil {
		style = ftn.fixture.Stats.Health().Style()
	} else {
		// Default styles by type
		switch ftn.fixture.Type {
		case FileNode:
			style = "text-blue-600 font-bold"
		case SectionNode:
			style = "text-blue-500"
		default:
			style = ""
		}
	}

	return api.Text{
		Content: content,
		Style:   style,
	}
}

func (ftn FixtureTreeNode) GetChildren() []api.TreeNode {
	if len(ftn.fixture.Children) == 0 {
		return nil
	}
	nodes := make([]api.TreeNode, len(ftn.fixture.Children))
	for i, child := range ftn.fixture.Children {
		nodes[i] = child.Tree()
	}
	return nodes
}

// Walk visits all nodes in the tree, calling visitor for test nodes
func (fn *FixtureNode) Walk(vistor func(f *FixtureNode)) {
	// Visit test nodes (nodes that have a Test to execute)
	if fn.Test != nil {
		vistor(fn)
	}
	// Recursively visit children
	for _, child := range fn.Children {
		child.Walk(vistor)
	}
}

// GetAllTests returns all test nodes in this subtree
func (fn *FixtureNode) GetAllTests() []*FixtureResult {
	var tests []*FixtureResult
	fn.Walk(func(f *FixtureNode) {
		if f.Type == TestNode && f.Test != nil {
			// Create a FixtureResult from the FixtureTest
			result := &FixtureResult{
				Name: f.Test.Name,
				Type: "query", // Default type for tests
				Test: *f.Test,
			}
			tests = append(tests, result)
		}
	})
	return tests
}

// PruneEmptySections removes SectionNodes that contain no tests.
// This method modifies the tree in-place, removing empty sections recursively.
// A section is considered empty if it has no TestNode descendants (Stats.Total == 0).
func (fn *FixtureNode) PruneEmptySections() {
	if fn == nil || fn.Children == nil {
		return
	}

	// First, recursively prune all children
	for _, child := range fn.Children {
		child.PruneEmptySections()
	}

	// Then filter out empty sections from this node's children
	filtered := make([]*FixtureNode, 0, len(fn.Children))
	for _, child := range fn.Children {
		// Keep if:
		// 1. It's a TestNode (actual test)
		// 2. It's a FileNode (file container)
		// 3. It's a SectionNode with tests (calculated stats Total > 0)
		shouldKeep := child.Type == TestNode ||
			child.Type == FileNode ||
			(child.Type == SectionNode && child.GetStats().Total > 0)

		if shouldKeep {
			filtered = append(filtered, child)
		}
	}

	fn.Children = filtered
}
