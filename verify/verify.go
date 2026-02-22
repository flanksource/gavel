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
}

func RunVerify(opts RunOptions) (*VerifyResult, error) {
	scope, err := ResolveScope(opts.Args, opts.CommitRange, opts.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve scope: %w", err)
	}
	adapter, model := ResolveAdapter(opts.Config.Model)

	logger.Infof("Verifying %s using %s", scope, model)

	prompt, err := renderPrompt(scope, opts.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to render prompt: %w", err)
	}

	schemaFile, err := SchemaFile(opts.Config.Checks)
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

func ComputeOverallScore(r VerifyResult) int {
	var total, passed int
	for _, cr := range r.Checks {
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
