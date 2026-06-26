package verify

import (
	"fmt"
	"math"
	"os"

	"github.com/flanksource/commons/logger"
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

	adapter, model := ResolveAdapter(cfg.Model)

	logger.Infof("Verifying %s using %s", scope, model)

	prompt, err := renderPrompt(scope, cfg, opts.Issue)
	if err != nil {
		return nil, fmt.Errorf("failed to render prompt: %w", err)
	}

	schemaFile, err := SchemaFile(cfg.Checks, opts.Issue != nil, criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema file: %w", err)
	}
	defer os.Remove(schemaFile)

	raw, err := Execute(adapter, prompt, model, schemaFile, opts.RepoPath, logger.V(2).Enabled())
	if err != nil {
		return nil, fmt.Errorf("CLI execution failed: %w", err)
	}

	adapter.PostExecute(raw)

	result, err := adapter.ParseResponse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result.Score = ComputeOverallScore(result)
	return &result, nil
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
		if cr.Met {
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
