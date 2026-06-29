package types

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/repomap"
)

const (
	// defaultAIStepModel is an agentic captain backend (claude-code) so the agent
	// can fetch the diff via its own tools. Overridden by `ai.model` front matter.
	defaultAIStepModel     = "claude-code-sonnet"
	defaultAIStepThreshold = 80
)

func init() {
	fixtures.AIStepRunner = RunAIStep
}

// RunAIStep runs an AI verification fixture: it reviews the scope diff against
// the document's checklist criteria via a captain agent and maps the structured
// verify result onto a FixtureResult. Pass/fail uses the step's CEL/expectations
// over the JSON output when present, otherwise the implicit
// implemented && score >= threshold rule.
func RunAIStep(fixture fixtures.FixtureTest, opts fixtures.RunOptions) fixtures.FixtureResult {
	now := time.Now()
	result := fixtures.FixtureResult{
		Name:     fixture.Name,
		Type:     "verify",
		Start:    &now,
		Test:     fixture,
		Metadata: map[string]interface{}{},
	}

	repoPath := repomap.FindGitRoot(fixture.SourceDir)
	if repoPath == "" {
		repoPath = fixture.SourceDir
	}
	if repoPath == "" {
		repoPath = opts.WorkDir
	}
	result.CWD = repoPath

	fm := fixture.FrontMatter
	agentCfg := fm.AI.ToAgentConfig(defaultAIStepModel)

	cfg := verify.VerifyConfig{Model: agentCfg.Model}
	if fm.Verify != nil && len(fm.Verify.Disabled) > 0 {
		cfg.Checks.Disabled = fm.Verify.Disabled
	}

	runOpts := verify.RunOptions{
		Config:         cfg,
		RepoPath:       repoPath,
		AgentConfig:    &agentCfg,
		PromptOverride: fixture.AIStep.Prompt,
		Issue:          buildAIStepIssue(fixture),
	}
	if fm.Verify != nil && fm.Verify.Scope != "" {
		runOpts.Args = []string{fm.Verify.Scope}
	}

	vr, err := verify.RunVerify(runOpts)
	if err != nil {
		return result.Errorf(err, "verify failed")
	}

	jsonBytes, err := json.Marshal(vr)
	if err != nil {
		return result.Errorf(err, "marshal verify result")
	}
	result.Actual = vr
	result.Metadata["verify"] = vr
	result.Stdout = string(jsonBytes)

	// CEL/expectations over the JSON output govern when present; otherwise the
	// implicit threshold.
	if fixture.Expected.CEL != "" {
		return fixture.Expected.Evaluate(result,
			exec.ExecResult{Stdout: string(jsonBytes), ExitCode: 0},
			fixtures.EvaluateOptions{SourceDir: fixture.SourceDir})
	}

	threshold := defaultAIStepThreshold
	if fm.Verify != nil && fm.Verify.Threshold > 0 {
		threshold = fm.Verify.Threshold
	}
	implemented := vr.Implemented != nil && *vr.Implemented
	if implemented && vr.Score >= threshold {
		result.Status = task.StatusPASS
		result.Duration = time.Since(now)
		return result
	}
	return result.Failf("verify score %d/%d (implemented=%v)", vr.Score, threshold, implemented)
}

// buildAIStepIssue turns the fixture's checklist into a verify issue context:
// items naming a known check ID enable that static check, the rest become custom
// acceptance criteria scored one verdict each.
func buildAIStepIssue(fixture fixtures.FixtureTest) *verify.IssueContext {
	issue := &verify.IssueContext{Title: fixture.Name}
	if fixture.AIStep == nil {
		return issue
	}
	issue.Description = fixture.AIStep.Description
	for _, item := range fixture.AIStep.Criteria {
		if id := matchCheckID(item.Text); id != "" {
			issue.CheckIDs = append(issue.CheckIDs, id)
		} else {
			issue.Criteria = append(issue.Criteria, item.Text)
		}
	}
	return issue
}

func matchCheckID(text string) string {
	t := strings.TrimSpace(text)
	for _, c := range verify.AllChecks {
		if c.ID == t {
			return c.ID
		}
	}
	return ""
}
