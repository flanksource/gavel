package types

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/commons/merge"
	"github.com/flanksource/gavel/fixtures"
)

// TODOVerifyConfig specifies legacy AI gate settings carried by a TODO.
type TODOVerifyConfig struct {
	Categories     []string `yaml:"categories,omitempty" json:"categories,omitempty"`
	ScoreThreshold int      `yaml:"score_threshold,omitempty" json:"score_threshold,omitempty"`
}

// AgentChecksConfig configures the post-completion check loop: the gavel test
// and lint options gavel runs once an agent reports done, feeding any failures
// back to the agent. It is read from .gavel.yaml (project default) and from a
// TODO's frontmatter (per-issue override). The loop is opt-in: it runs only
// when one of those enables it; the loop's round budget is the lifecycle
// step's `workflow.verify.maxIterations`.
type AgentChecksConfig struct {
	// Enabled gates the loop. nil means "inherit" (off unless a higher layer
	// turns it on); a non-nil value is authoritative.
	Enabled *bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	// Retry is the CEL definition-of-done predicate: while it evaluates true the
	// agent is re-run with the failing/warned nodes as feedback. It reads the
	// verification report under the variable `verify` and is stamped into the
	// generated definition-of-done document's front matter. Empty resolves to
	// fixtures.DefaultRetryExpr.
	Retry string `yaml:"retry,omitempty" json:"retry,omitempty"`
	// Test, when non-nil, runs `gavel test` with these options. nil skips tests.
	Test *AgentTestConfig `yaml:"test,omitempty" json:"test,omitempty"`
	// Lint, when non-nil, runs `gavel lint` with these options. nil skips lint.
	Lint *AgentLintConfig `yaml:"lint,omitempty" json:"lint,omitempty"`
}

// AgentTestConfig is the subset of gavel test options the check loop exposes.
type AgentTestConfig struct {
	Paths   []string `yaml:"paths,omitempty" json:"paths,omitempty"`     // package paths to test (empty = discover)
	Changed bool     `yaml:"changed,omitempty" json:"changed,omitempty"` // only packages affected by changes
	Timeout string   `yaml:"timeout,omitempty" json:"timeout,omitempty"` // global wall-clock deadline (e.g. "5m")
}

// AgentLintConfig is the subset of gavel lint options the check loop exposes.
type AgentLintConfig struct {
	Linters []string `yaml:"linters,omitempty" json:"linters,omitempty"` // linters to run (empty = all detected)
	Changed bool     `yaml:"changed,omitempty" json:"changed,omitempty"` // only report new issues vs the base ref
	Timeout string   `yaml:"timeout,omitempty" json:"timeout,omitempty"` // per-linter deadline (e.g. "5m")
}

// IsEnabled reports whether the loop should run for this resolved config.
func (c AgentChecksConfig) IsEnabled() bool {
	return c.Enabled != nil && *c.Enabled
}

// HasChecks reports whether at least one of test/lint is configured to run.
func (c AgentChecksConfig) HasChecks() bool {
	return c.Test != nil || c.Lint != nil
}

// Overlay returns base with override's set fields layered on top: the override
// wins field by field, and its unset fields leave base intact — so a TODO whose
// frontmatter names only `retry` still inherits the project's test and lint
// options. A nil override returns base unchanged. The merge is structural, so a
// field added above needs no code here.
func (base AgentChecksConfig) Overlay(override *AgentChecksConfig) AgentChecksConfig {
	if override == nil {
		return base
	}
	return merge.Apply(base, *override, merge.Policy{Replace: []any{(*bool)(nil)}})
}

// ResolveAgentChecks produces the effective config for a run by overlaying a
// TODO's frontmatter onto the project default (frontmatter wins field-by-field).
// One default is then applied: an enabled config with neither test nor lint
// set gets the sensible default of running both against changed files.
func ResolveAgentChecks(project AgentChecksConfig, frontmatter *AgentChecksConfig) AgentChecksConfig {
	resolved := project.Overlay(frontmatter)
	if resolved.Retry == "" {
		resolved.Retry = fixtures.DefaultRetryExpr
	}
	if resolved.IsEnabled() && !resolved.HasChecks() {
		resolved.Test = &AgentTestConfig{Changed: true}
		resolved.Lint = &AgentLintConfig{Changed: true}
	}
	return resolved
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
