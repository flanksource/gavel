package todos

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	captainapi "github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

const providerPersistenceTimeout = 30 * time.Second

var ErrExecutionCancelled = errors.New("todo run stopped by user")

// ApprovalBroker builds the callback a run answers tool-permission requests
// with. Only a host that serves an approval surface supplies one — the CLI
// leaves it nil, because a run that asked a terminal for a decision would block
// until its timeout.
//
// It is a factory rather than a callback because the durable approval rows are
// keyed on the session and prompt run Captain admits, neither of which exists
// until the run is under way.
type ApprovalBroker func(ctx *ExecutorContext) (captainapi.PermissionFunc, error)

// ExecutionResult is the record one lifecycle step run leaves behind: what it
// cost, what it said, and what it decided. The lifecycle host produces it and
// the provider persists it as the todo's attempt.
type ExecutionResult struct {
	Success          bool
	Skipped          bool
	Cancelled        bool
	ExecutorName     string        // Which runtime ran it (e.g. "cli-claude")
	TokensUsed       int           // Total tokens consumed
	CostUSD          float64       // Cost in USD
	Duration         time.Duration // Total execution time
	NumTurns         int           // Number of interaction rounds
	ActionsPerformed []string      // List of actions taken (tool uses, etc.)
	ErrorMessage     string
	CommitSHA        string
	Runtime          RunStartMetadata
	Transcript       *ExecutionTranscript
	// Envelope fields — the agent's structured final result. EndStatus is empty
	// when no envelope was captured.
	Summary   string
	EndStatus types.EndStatus
	Questions []types.AgentQuestion
	Plan      *types.PlanResult
	// Triage is a triage run's verdict and the edits it wants applied. The agent
	// is read-only, so this is a request, not a record of something that
	// happened; the host's OnOutcome performs the writes.
	Triage *types.TriageEnvelope
	// DoD is the definition-of-done verdict: nil when the step declared no
	// verifiers, else Ran is true and Passed reports whether every verifier
	// passed within the iteration budget.
	DoD *DoDOutcome
}

// DoDOutcome is the terminal verdict of a run's definition of done: Passed is
// true only when the agent loop stopped because every verifier passed (captain
// "condition-met"), false when the iteration budget ran out with a check still
// red.
//
// Report is the last iteration's typed verification report — captain's
// api.VerifyReport, on the wire exactly as captain spells it, so the dashboard
// renders the same document the agent loop judged.
type DoDOutcome struct {
	Ran    bool                     `json:"ran"`
	Passed bool                     `json:"passed"`
	Report *captainapi.VerifyReport `json:"report,omitempty"`
}

func (e ExecutionResult) Pretty() api.Text {
	result := clicky.Text(" Executed with ", "text-gray-500").Append(e.ExecutorName, "text-blue-600 font-bold")

	if e.Success {
		result = result.Add(icons.Pass)
	} else if e.Skipped {
		result = result.Add(icons.Skip)
	} else {
		result = result.Add(icons.Fail)
	}

	if e.TokensUsed > 0 {
		result = result.Append(fmt.Sprintf("   Tokens: %d", e.TokensUsed), "text-gray-500")
	}

	if e.CostUSD > 0 {
		result = result.Append(fmt.Sprintf("   Cost: $%.4f", e.CostUSD), "text-gray-500")
	}

	if e.Duration > 0 {
		result = result.Append(fmt.Sprintf("   Duration: %s", e.Duration.String()), "text-gray-500")
	}

	if e.NumTurns > 0 {
		result = result.Append(fmt.Sprintf("   Turns: %d", e.NumTurns), "text-gray-500")
	}

	if len(e.ActionsPerformed) > 0 {
		result = result.Append("   Actions: ", "text-gray-500").Append(fmt.Sprintf("%v", e.ActionsPerformed), "text-gray-500")
	}

	return result
}

// RenderRunStartComment is the comment a run leaves on its todo when it starts:
// the session and the runtime it resolved.
func RenderRunStartComment(meta RunStartMetadata) string {
	var b strings.Builder
	b.WriteString("**Todo run started**\n\n")
	b.WriteString("- **Session ID:** `" + commentValue(meta.SessionID, "unknown") + "`\n")
	b.WriteString("- **Mode:** `" + commentValue(meta.Mode, "run") + "`\n")
	b.WriteString("- **Driver:** `" + commentValue(meta.Driver, "unknown") + "`\n")
	b.WriteString("- **Agent:** `" + commentValue(meta.Agent, "unknown") + "`\n")
	b.WriteString("- **Provider:** `" + commentValue(meta.Provider, "unknown") + "`\n")
	b.WriteString("- **Mode:** `" + commentValue(meta.RuntimeMode, "default") + "`\n")
	b.WriteString("- **Resolved Model:** `" + commentValue(meta.ResolvedModel, "default") + "`\n")
	b.WriteString("- **Effort:** `" + commentValue(meta.Effort, "default") + "`")
	return b.String()
}

func commentValue(value, fallback string) string {
	if v := strings.TrimSpace(value); v != "" {
		return v
	}
	return fallback
}

// PersistenceContext is the context a provider write runs under once the run's
// own context may be gone: the caller's while it is live, otherwise a bounded
// detached one, so a cancelled run still records how it ended.
func PersistenceContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		return context.WithTimeout(context.Background(), providerPersistenceTimeout)
	}
	if ctx.Err() == nil {
		return ctx, func() {}
	}
	return context.WithTimeout(context.WithoutCancel(ctx), providerPersistenceTimeout)
}

// RunNodes runs fixture nodes through the fixture engine's node runner,
// returning one result per test node. A walk that could not finish is reported
// as an extra errored result rather than as a short list that reads like
// everything passed.
func RunNodes(ctx context.Context, workDir string, nodes []*fixtures.FixtureNode) []fixtures.FixtureResult {
	results, _, err := fixtures.RunNodes(ctx, nodes, fixtures.RunOptions{WorkDir: workDir})
	if err != nil {
		return append(results, fixtures.FixtureResult{
			Name: "verification progress", Status: task.StatusERR,
			Test: fixtures.FixtureTest{Name: "verification progress"}, Error: err.Error(),
		})
	}
	return results
}

// AllPassed checks if all fixture results passed.
func AllPassed(results []fixtures.FixtureResult) bool {
	for _, r := range results {
		if !r.IsOK() {
			return false
		}
	}
	return true
}
