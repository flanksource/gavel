package types

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/exec"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/repomap"
)

func init() {
	fixtures.AIStepRunner = RunAIStep
}

// checklistResponse is the ai-step's structured output: one verdict per
// acceptance-criterion checklist item.
type checklistResponse struct {
	Items []fixtures.ChecklistResult `json:"items" description:"One entry per acceptance criterion, in the same order they were listed, each with a pass/fail verdict and a one-line justification."`
}

// RunAIStep runs an AI acceptance-criteria checklist fixture. It builds a captain
// agent directly from the fixture's `ai:` front matter, has the agent evaluate
// each checklist item against the change in the repository, and emits one
// pass/fail Test per item. The per-item {item, passed, message} verdicts are also
// stored under Metadata["checklist"] so the definition-of-done predicate can read
// them as results.checklist. Pass/fail uses the step's CEL expectation over the
// JSON output when present, otherwise the implicit "every criterion passed" rule.
func RunAIStep(fixture fixtures.FixtureTest, opts fixtures.RunOptions) fixtures.FixtureResult {
	now := time.Now()
	result := fixtures.FixtureResult{
		Name:     fixture.Name,
		Type:     "verify",
		Start:    &now,
		Test:     fixture,
		Metadata: map[string]interface{}{},
	}

	repoPath := fixtureRepoPath(fixture, opts)
	result.CWD = repoPath

	items := checklistItems(fixture)
	if len(items) == 0 {
		result.Status = task.StatusSKIP
		result.Error = "no acceptance criteria to verify"
		result.Duration = time.Since(now)
		return result
	}

	var schema checklistResponse
	spec, config := resolveAIStepSpec(fixture, opts, &schema)
	provider, err := ai.NewProvider(config)
	if err != nil {
		return result.Errorf(err, "build ai provider")
	}
	defer ai.CloseProvider(provider) //nolint:errcheck

	runContext := opts.Context
	if runContext == nil {
		runContext = context.Background()
	}
	resp, err := provider.Execute(runContext, spec)
	if err != nil {
		return result.Errorf(err, "checklist prompt")
	}
	if derr := decodeChecklistResponse(resp, &schema); derr != nil && len(schema.Items) == 0 {
		return result.Errorf(derr, "decode checklist response")
	}

	return scoreChecklist(fixture, result, items, schema.Items, now)
}

// resolveAIStepSpec layers the fixture's own `ai:` front matter over the caller's
// spec and returns both the spec to execute and the provider config built from
// it. The fixture wins because it is the most specific statement about how it
// wants to be graded; a todo run's generated fixture declares none, so the
// caller's resolved verification spec stands.
//
// WithoutSession is the invariant: a grader inherits how to run, never what was
// already said. Resuming the session it is judging would be the candidate
// marking its own exam.
func resolveAIStepSpec(fixture fixtures.FixtureTest, opts fixtures.RunOptions, schema *checklistResponse) (api.Spec, ai.AgentConfig) {
	var spec api.Spec
	if opts.Spec != nil {
		spec = opts.Spec.WithoutSession()
	}
	config := fixture.AI.ToAgentConfig()
	spec = spec.Merge(api.Spec{Model: config.Model, Budget: config.Budget})

	spec.Prompt.User = buildChecklistPrompt(fixture, fixtureRepoPath(fixture, opts), checklistItems(fixture), opts.Changed)
	spec.Prompt.Source = "fixtures.ai-step"
	spec.Prompt.Schema = schema
	spec.Prompt.SchemaJSON = nil

	config.Model = spec.Model
	config.Budget = spec.Budget
	return spec, config
}

func decodeChecklistResponse(resp *api.Response, target *checklistResponse) error {
	if len(target.Items) > 0 {
		return nil
	}
	if resp == nil {
		return fmt.Errorf("checklist prompt returned no response")
	}
	if resp.StructuredData != nil {
		var raw []byte
		switch value := resp.StructuredData.(type) {
		case json.RawMessage:
			raw = value
		default:
			var err error
			raw, err = json.Marshal(value)
			if err != nil {
				return err
			}
		}
		if err := json.Unmarshal(raw, target); err == nil {
			return nil
		}
	}
	if strings.TrimSpace(resp.Text) == "" {
		return fmt.Errorf("checklist prompt returned no structured output")
	}
	return json.Unmarshal([]byte(resp.Text), target)
}

func fixtureRepoPath(fixture fixtures.FixtureTest, opts fixtures.RunOptions) string {
	if repoPath := repomap.FindGitRoot(fixture.SourceDir); repoPath != "" {
		return repoPath
	}
	if fixture.SourceDir != "" {
		return fixture.SourceDir
	}
	return opts.WorkDir
}

// checklistItems returns the fixture's acceptance-criteria checklist.
func checklistItems(fixture fixtures.FixtureTest) []fixtures.ChecklistItem {
	if fixture.AIStep == nil {
		return nil
	}
	return fixture.AIStep.Criteria
}

// buildChecklistPrompt renders the acceptance-criteria review prompt: the
// document prose and any custom instructions, then the numbered criteria, asking
// the agent to inspect the change at repoPath and return a per-item verdict.
// When the verification is scoped to particular files, the grader is told which,
// so it judges the change under review rather than everything dirty in the tree.
func buildChecklistPrompt(fixture fixtures.FixtureTest, repoPath string, items []fixtures.ChecklistItem, changed []string) string {
	var b strings.Builder
	b.WriteString("You are verifying whether a code change satisfies its acceptance criteria.\n")
	fmt.Fprintf(&b, "Inspect the current change in the git repository at %s — the working-tree diff, staged changes, and the most recent commits — using your tools.\n\n", repoPath)
	if len(changed) > 0 {
		b.WriteString("The change under review is limited to these files (relative to the repository root); judge the criteria against them:\n")
		for _, path := range changed {
			fmt.Fprintf(&b, "- %s\n", path)
		}
		b.WriteString("\n")
	}
	if fixture.AIStep != nil {
		if desc := strings.TrimSpace(fixture.AIStep.Description); desc != "" {
			b.WriteString(desc)
			b.WriteString("\n\n")
		}
		if custom := strings.TrimSpace(fixture.AIStep.Prompt); custom != "" {
			b.WriteString(custom)
			b.WriteString("\n\n")
		}
	}
	b.WriteString("Acceptance criteria — return exactly one verdict per item, in this order:\n")
	for i, item := range items {
		fmt.Fprintf(&b, "%d. %s\n", i+1, item.Text)
	}
	b.WriteString("\nFor each criterion set `passed` true only when the change clearly satisfies it, and give a one-line `message` citing the evidence for a pass or the gap for a fail.")
	return b.String()
}

// scoreChecklist aligns the agent's verdicts to the criteria (one result per
// item, order-preserving), records them for the predicate and as child Tests,
// and decides the step status. A CEL expectation governs when present; otherwise
// the step passes only when every criterion passed.
func scoreChecklist(fixture fixtures.FixtureTest, result fixtures.FixtureResult, items []fixtures.ChecklistItem, verdicts []fixtures.ChecklistResult, start time.Time) fixtures.FixtureResult {
	checklist := alignChecklist(items, verdicts)
	if result.Metadata == nil {
		result.Metadata = map[string]interface{}{}
	}
	result.Metadata["checklist"] = checklist
	result.Actual = checklist
	result.Children = checklistChildren(checklist)

	jsonBytes, err := json.Marshal(checklistResponse{Items: checklist})
	if err != nil {
		return result.Errorf(err, "marshal checklist result")
	}
	result.Stdout = string(jsonBytes)

	if fixture.Expected.CEL != "" {
		return fixture.Expected.Evaluate(result,
			exec.ExecResult{Stdout: string(jsonBytes), ExitCode: 0},
			fixtures.EvaluateOptions{SourceDir: fixture.SourceDir})
	}

	failed := 0
	for _, c := range checklist {
		if !c.Passed {
			failed++
		}
	}
	if failed == 0 {
		result.Status = task.StatusPASS
		result.Duration = time.Since(start)
		return result
	}
	return result.Failf("%d/%d acceptance criteria not met", failed, len(checklist))
}

// alignChecklist maps the agent's verdicts back onto the criteria, one entry per
// item in the original order. It matches by criterion text first, falls back to
// positional order when the counts match, and marks any criterion the agent did
// not answer as failed — so a missing verdict never silently passes.
func alignChecklist(items []fixtures.ChecklistItem, verdicts []fixtures.ChecklistResult) []fixtures.ChecklistResult {
	byItem := make(map[string]fixtures.ChecklistResult, len(verdicts))
	for _, v := range verdicts {
		byItem[normalizeCriterion(v.Item)] = v
	}
	out := make([]fixtures.ChecklistResult, len(items))
	for i, item := range items {
		v, ok := byItem[normalizeCriterion(item.Text)]
		if !ok && len(verdicts) == len(items) {
			v, ok = verdicts[i], true
		}
		if !ok {
			out[i] = fixtures.ChecklistResult{Item: item.Text, Passed: false, Message: "no verdict returned for this criterion"}
			continue
		}
		out[i] = fixtures.ChecklistResult{Item: item.Text, Passed: v.Passed, Message: v.Message}
	}
	return out
}

func normalizeCriterion(text string) string {
	return strings.ToLower(strings.TrimSpace(text))
}

// checklistChildren renders each criterion verdict as a pass/fail child Test so
// the fixture engine's stats/display pipeline rolls the checklist up like any
// other step. A failed item carries its justification as the node error.
func checklistChildren(checklist []fixtures.ChecklistResult) []*fixtures.FixtureNode {
	children := make([]*fixtures.FixtureNode, 0, len(checklist))
	for _, c := range checklist {
		status := task.StatusPASS
		if !c.Passed {
			status = task.StatusFAIL
		}
		res := &fixtures.FixtureResult{Name: c.Item, Type: "verify", Status: status}
		if !c.Passed {
			res.Error = c.Message
		}
		children = append(children, &fixtures.FixtureNode{
			Name:    c.Item,
			Type:    fixtures.TestNode,
			Results: res,
		})
	}
	return children
}
