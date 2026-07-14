package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/ai"
)

type RunOptions struct {
	Config      VerifyConfig
	RepoPath    string
	Args        []string
	CommitRange string
	// Issue, when set, makes the run issue-aware: the reviewer scores the
	// issue's commits against its description, comments, and stored acceptance
	// criteria, and the result carries Implemented + AcceptanceCriteria.
	Issue *IssueContext
	// PromptOverride, when non-empty, is used verbatim instead of the rendered
	// verify prompt — the dashboard's editable prompt. The JSON output schema is
	// unchanged (criteria/checks/ratings still drive it).
	PromptOverride string
	// AgentConfig, when set, runs the review through a captain agentic backend
	// (claude-code / claude-agent): the agent fetches the diff via its own tools
	// and is told to emit JSON matching the schema, which is parsed here. Fixture
	// AI steps and TODO definition-of-done checks set it when they need an
	// explicit backend; other library callers can leave it nil to use defaults.
	AgentConfig *ai.AgentConfig
	// SchemaCEL, when non-empty, overrides the built-in schema.cel expression that
	// generates the JSON output schema from the review's checks/criteria (the
	// fixture AI-step `schema.cel` frontmatter). Empty uses DefaultSchemaCEL.
	SchemaCEL string
}

func RunVerify(opts RunOptions) (*VerifyResult, error) {
	cfg := opts.Config
	var criteria []string

	scope, err := resolveRunScope(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve scope: %w", err)
	}
	if opts.Issue != nil {
		cfg.Checks = issueChecksConfig(opts.Issue.CheckIDs)
		criteria = opts.Issue.Criteria
	}

	prompt, err := resolveVerifyPrompt(opts, scope, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to render prompt: %w", err)
	}

	// Every review runs through a captain backend (native structured output).
	// When the caller supplies no agent config, synthesize the default from the
	// verify model.
	if opts.AgentConfig == nil {
		ac := ai.DefaultConfig()
		if cfg.Model != "" {
			ac.Model = cfg.Model
		}
		opts.AgentConfig = &ac
	}

	result, err := runAgentic(opts, scope, cfg, prompt, criteria)
	if err != nil {
		return nil, err
	}

	result.Score = ComputeOverallScore(result)
	return &result, nil
}

// runAgentic executes the review with a captain agentic backend and parses the
// JSON reply. The schema is built from the run's checks + criteria and embedded
// in the prompt (the agentic backends have no native structured-output mode).
func runAgentic(opts RunOptions, scope ReviewScope, cfg VerifyConfig, prompt string, criteria []string) (VerifyResult, error) {
	schema, err := EvalSchema(opts.SchemaCEL, EnabledChecks(cfg.Checks), opts.Issue != nil, criteria)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("failed to build schema: %w", err)
	}
	logger.Infof("Verifying %s using %s", scope, opts.AgentConfig.Model)

	provider, err := ai.NewProvider(*opts.AgentConfig)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("build verify agent: %w", err)
	}
	req := captainai.Request{Prompt: api.Prompt{
		User:       prompt,
		SchemaJSON: json.RawMessage(schema),
		Source:     "verify",
	}}
	req.SetCwd(opts.RepoPath)
	resp, err := provider.Execute(context.Background(), req)
	if err != nil {
		return VerifyResult{}, fmt.Errorf("verify agent execution failed: %w", err)
	}

	// Providers that honor the schema natively return the JSON on StructuredData;
	// cmux/text backends leave it on Text. Prefer the clean structured payload.
	raw := resp.Text
	if sd, ok := resp.StructuredData.(json.RawMessage); ok && len(sd) > 0 {
		raw = string(sd)
	}
	result, err := captainai.ParseStructured(raw, validateVerifyResult)
	if err != nil {
		return VerifyResult{}, err
	}
	return *result, nil
}

// resolveVerifyPrompt returns the verbatim PromptOverride when set, otherwise the
// rendered verify prompt. Keeping it separate makes the override path unit-testable.
func resolveVerifyPrompt(opts RunOptions, scope ReviewScope, cfg VerifyConfig) (string, error) {
	if opts.PromptOverride != "" {
		return opts.PromptOverride, nil
	}
	template, err := cfg.PromptTemplate.Resolve(opts.RepoPath, verifyPromptTemplate)
	if err != nil {
		return "", err
	}
	return renderPrompt(template, scope, cfg, opts.Issue, opts.RepoPath)
}

// PreviewPrompt renders the verify prompt for a run without executing it, so the
// dashboard can seed an editable prompt with exactly what RunVerify would send.
func PreviewPrompt(opts RunOptions) (string, error) {
	cfg := opts.Config
	scope, err := resolveRunScope(opts)
	if err != nil {
		return "", fmt.Errorf("failed to resolve scope: %w", err)
	}
	if opts.Issue != nil {
		cfg.Checks = issueChecksConfig(opts.Issue.CheckIDs)
	}
	template, err := cfg.PromptTemplate.Resolve(opts.RepoPath, verifyPromptTemplate)
	if err != nil {
		return "", err
	}
	return renderPrompt(template, scope, cfg, opts.Issue, opts.RepoPath)
}

// resolveRunScope targets the issue's commits when the run is issue-aware,
// otherwise falls back to the generic arg/commit-range scope resolution.
func resolveRunScope(opts RunOptions) (ReviewScope, error) {
	if opts.Issue != nil && len(opts.Issue.CommitSHAs) > 0 {
		return ReviewScope{Type: "commits", Commits: opts.Issue.CommitSHAs}, nil
	}
	return ResolveScope(opts.Args, opts.CommitRange, opts.RepoPath)
}

// issueChecksConfig narrows the static checks to the issue's selected applicable
// set, always keeping definition-of-done so an issue-aware run scores at least
// one check (the parse layer requires it).
func issueChecksConfig(selected []string) ChecksConfig {
	keep := map[string]bool{"definition-of-done": true}
	for _, id := range selected {
		keep[id] = true
	}
	var disabled []string
	for _, c := range AllChecks {
		if !keep[c.ID] {
			disabled = append(disabled, c.ID)
		}
	}
	return ChecksConfig{Disabled: disabled}
}

func ComputeOverallScore(r VerifyResult) int {
	var total, passed int
	for _, cr := range r.Checks {
		total++
		if cr.Pass {
			passed++
		}
	}
	// Stored acceptance criteria count like checks toward the pass rate.
	for _, cr := range r.AcceptanceCriteria {
		total++
		if cr.Pass {
			passed++
		}
	}

	checkScore := 0.0
	if total > 0 {
		checkScore = float64(passed) / float64(total) * 100
	}

	ratingSum := 0.0
	ratingCount := 0
	for _, rr := range r.Ratings {
		ratingSum += float64(rr.Score)
		ratingCount++
	}
	ratingScore := 0.0
	if ratingCount > 0 {
		ratingScore = ratingSum / float64(ratingCount)
	}

	completenessScore := 0.0
	if r.Completeness.Pass {
		completenessScore = 100
	}

	// Weighted: checks 50%, ratings 35%, completeness 15%
	combined := checkScore*0.50 + ratingScore*0.35 + completenessScore*0.15
	return int(math.Round(combined))
}
