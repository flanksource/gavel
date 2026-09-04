package types

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"
	"time"

	captainapi "github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/labels"
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

		// Same priority, sort alphabetically by provider display name
		nameI := todos[i].sortName()
		nameJ := todos[j].sortName()
		return nameI < nameJ
	})

}

func (t TODO) sortName() string {
	if t.Title != "" {
		return strings.ToLower(t.Title)
	}
	if t.FilePath != "" {
		return strings.ToLower(filepath.Base(t.FilePath))
	}
	return strings.ToLower(t.ID)
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

type ProviderEvent struct {
	ID        string    `json:"id,omitempty"`
	ShortID   string    `json:"short_id,omitempty"`
	Kind      string    `json:"kind,omitempty"`
	Actor     string    `json:"actor,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Title     string    `json:"title,omitempty"`
	Body      string    `json:"body,omitempty"`
	Label     string    `json:"label,omitempty"`
	// OldLabel/NewLabel are set for a "LabelChanged" event, which collapses an
	// adjacent LabelRemoved/LabelAdded pair within the same label namespace.
	OldLabel string `json:"old_label,omitempty"`
	NewLabel string `json:"new_label,omitempty"`
}

// TODO represents a structured TODO item parsed from a markdown file.
// It combines fixture test nodes with TODO-specific metadata for tracking
// implementation tasks including reproduction steps, verification tests, and execution status.
type TODO struct {
	FilePath       string                `json:"file_path,omitempty"`
	FileNode       *fixtures.FixtureNode `json:"file_node,omitempty"` // Root file node from fixtures parser
	ID             string                `json:"id,omitempty"`
	ShortID        string                `json:"short_id,omitempty"`
	Version        int64                 `json:"version,omitempty"`
	WorkspaceID    string                `json:"workspace_id,omitempty"`
	ExecutionState string                `json:"execution_state,omitempty"`
	Provider       string                `json:"provider,omitempty"`
	Workspace      string                `json:"workspace,omitempty"`
	ProviderState  string                `json:"provider_state,omitempty"`
	Labels         []string              `json:"labels,omitempty"`
	// LabelDefinitions are the resolved presentations of Labels, in the same
	// order. The runtime provider populates them from one definition read per
	// request; a TODO from a source with no definition store leaves them empty
	// and the renderers fall back to the hashed palette colour.
	LabelDefinitions labels.Definitions `json:"label_definitions,omitempty"`
	// PhaseRuns is the latest run per lifecycle phase (plan/triage/run/verify).
	// The runtime provider populates it from one workspace-wide read per
	// request; a TODO from a source without run history leaves it empty and the
	// renderers show every phase as never run.
	PhaseRuns      PhaseRuns       `json:"phase_runs,omitempty"`
	ProviderEvents []ProviderEvent `json:"provider_events,omitempty"`
	// ExternalIssue is the external tracker issue this TODO is linked to, or
	// nil when it has never been pushed to one.
	ExternalIssue *ExternalIssue `json:"external_issue,omitempty"`

	TODOFrontmatter `json:",inline"`

	// Sections are FixtureNodes with Tests
	StepsToReproduce []*fixtures.FixtureNode `json:"steps_to_reproduce,omitempty"` // Section containing reproduction steps
	// Plain text implementation instructions
	Implementation    string                  `json:"implementation,omitempty"`
	Verification      []*fixtures.FixtureNode `json:"verification,omitempty"`       // Section containing verification tests
	CustomValidations []*fixtures.FixtureNode `json:"custom_validations,omitempty"` // Section containing custom validation tests
	MarkdownBody      string                  `json:"markdown_body,omitempty"`
	// VerificationMarkdown is the raw fixture source stored separately from MarkdownBody.
	VerificationMarkdown string `json:"verification_markdown,omitempty"`
	// AcceptanceCriteria are the editable done-ness criteria parsed from the
	// "## Acceptance Criteria" section, scored by issue-aware verification.
	AcceptanceCriteria []AcceptanceCriterion `json:"acceptance_criteria,omitempty"`
}

// ExternalIssue is the tracker issue a TODO has been pushed to. State carries
// the upstream issue's own status once something fetches it; it is empty until
// then and must be read as "unknown", never as open.
type ExternalIssue struct {
	Kind   string `json:"kind"`
	Repo   string `json:"repo"`
	Number int    `json:"number"`
	URL    string `json:"url"`
	State  string `json:"state,omitempty"`
}

// AcceptanceCriterion is one editable done-ness criterion for a TODO. CheckID is
// set when the line maps to a static verify.AllChecks id (rendered as
// "<id>: <text>"); empty CheckID marks a custom, functionality-specific
// criterion. Done reflects the checklist box (`- [x]`).
type AcceptanceCriterion struct {
	Text    string `json:"text"`
	CheckID string `json:"check_id,omitempty"`
	Done    bool   `json:"done,omitempty"`
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
	Title         string             `yaml:"title,omitempty" json:"title,omitempty"`
	Priority      Priority           `yaml:"priority,omitempty" json:"priority,omitempty"`
	Status        Status             `yaml:"status,omitempty" json:"status,omitempty"`
	Created       *time.Time         `yaml:"created,omitempty" json:"created,omitempty"`
	LastRun       *time.Time         `yaml:"last_run,omitempty" json:"last_run,omitempty"`
	Attempts      int                `yaml:"attempts,omitempty" json:"attempts,omitempty"`
	Language      Language           `yaml:"language,omitempty" json:"language,omitempty"`
	WorkingCommit string             `yaml:"working_commit,omitempty" json:"working_commit,omitempty"`
	Branch        string             `yaml:"branch,omitempty" json:"branch,omitempty"`
	CWD           string             `yaml:"cwd,omitempty" json:"cwd,omitempty"`
	Path          StringOrSlice      `yaml:"path,omitempty" json:"path,omitempty"`
	LLM           *LLM               `yaml:"llm,omitempty" json:"llm,omitempty"`
	Verify        *TODOVerifyConfig  `yaml:"verify,omitempty" json:"verify,omitempty"`
	Checks        *AgentChecksConfig `yaml:"checks,omitempty" json:"checks,omitempty"`
	PR            *PR                `yaml:"pr,omitempty" json:"pr,omitempty"`
	Prompt        string             `yaml:"prompt,omitempty" json:"prompt,omitempty"`
	// PlanPath is the agent's native plan-mode file from the last plan run
	// (reported in the run envelope); the plan is read from there, never copied.
	PlanPath   string     `yaml:"plan_path,omitempty" json:"plan_path,omitempty"`
	PlanStatus PlanStatus `yaml:"plan_status,omitempty" json:"plan_status,omitempty"`
	// RunMode records the mode of the last agent run so an answer-resume
	// continues in the same mode.
	RunMode RunMode `yaml:"run_mode,omitempty" json:"run_mode,omitempty"`
	// LastRunSummary is the final summary from the most recent run's envelope.
	LastRunSummary string `yaml:"last_run_summary,omitempty" json:"last_run_summary,omitempty"`
	// Questions are the agent's blocking questions while Status is ask.
	Questions []AgentQuestion `yaml:"questions,omitempty" json:"questions,omitempty"`
}

// CleanMetadata removes keys from Metadata that match struct field yaml tags.
// This fixes a goccy/go-yaml bug where inline maps capture ALL fields.
func (f *TODOFrontmatter) CleanMetadata() {
	// Clean embedded FrontMatter fields first
	f.FrontMatter.CleanMetadata()

	if f.Metadata == nil {
		return
	}
	if nested, ok := f.Metadata["metadata"]; ok {
		switch values := nested.(type) {
		case map[string]any:
			for k, v := range values {
				f.Metadata[k] = v
			}
		case map[interface{}]interface{}:
			for k, v := range values {
				if key, ok := k.(string); ok {
					f.Metadata[key] = v
				}
			}
		}
		delete(f.Metadata, "metadata")
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
	delete(f.Metadata, "checks")
	delete(f.Metadata, "working_commit")
	delete(f.Metadata, "branch")
	delete(f.Metadata, "cwd")
	delete(f.Metadata, "max_turns")
	delete(f.Metadata, "pr")
	delete(f.Metadata, "prompt")
	delete(f.Metadata, "plan_path")
	delete(f.Metadata, "plan_status")
	delete(f.Metadata, "run_mode")
	delete(f.Metadata, "last_run_summary")
	delete(f.Metadata, "questions")
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

// Status represents the current execution state of a TODO item.
type Status string

const (
	// StatusDraft indicates the TODO is being drafted and is not ready to run.
	StatusDraft Status = "draft"
	// StatusPending indicates the TODO has not been started.
	StatusPending Status = "pending"
	// StatusInProgress indicates the TODO is currently being worked on.
	StatusInProgress Status = "in_progress"
	// StatusReview indicates a proposed plan awaits human approval.
	StatusReview Status = "review"
	// StatusAsk indicates the agent is blocked on questions a human must answer.
	StatusAsk Status = "ask"
	// StatusVerified indicates the TODO's implementation met its definition of
	// done (the run's fixture verifiers passed).
	StatusVerified Status = "verified"
	// StatusUnverified indicates the TODO was implemented but its definition of
	// done was still failing when the run's iteration budget ran out — it needs
	// another look or another iteration, distinct from an outright failed run.
	StatusUnverified Status = "unverified"
	// StatusCompleted indicates the TODO has been successfully completed.
	StatusCompleted Status = "completed"
	// StatusFailed indicates the TODO execution failed.
	StatusFailed Status = "failed"
	// StatusSkipped indicates the TODO was skipped because reproduction steps already pass.
	StatusSkipped Status = "skipped"
)

// KnownStatuses returns the TODO statuses accepted by parsers and APIs.
func KnownStatuses() []Status {
	return []Status{
		StatusDraft,
		StatusPending,
		StatusInProgress,
		StatusReview,
		StatusAsk,
		StatusFailed,
		StatusUnverified,
		StatusVerified,
		StatusCompleted,
		StatusSkipped,
	}
}

func IsKnownStatus(status Status) bool {
	for _, known := range KnownStatuses() {
		if status == known {
			return true
		}
	}
	return false
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

// CheckResult contains the result of checking a single TODO by running its
// definition of done. Report is captain's typed verification report — the same
// document an in-loop check produces — so every surface renders one shape.
type CheckResult struct {
	TODO       *TODO                    `json:"todo,omitempty"`
	AllPassed  bool                     `json:"allPassed"`
	Duration   time.Duration            `json:"duration"`
	Error      error                    `json:"-"`
	ErrorText  string                   `json:"error,omitempty"`
	Report     *captainapi.VerifyReport `json:"report,omitempty"`
	TestResult *TestResultInfo          `json:"testResult,omitempty"` // Comprehensive test result info for updating todo file
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
