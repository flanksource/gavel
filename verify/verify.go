package verify

import (
	"fmt"
	"math"

	"github.com/flanksource/commons/logger"
)

type RunOptions struct {
	Config      VerifyConfig
	RepoPath    string
	Args        []string
	CommitRange string
}

func RunVerify(opts RunOptions) (*VerifyResult, error) {
	scope := ResolveScope(opts.Args, opts.CommitRange)
	tool, model := ResolveCLI(opts.Config.Model)

	logger.Infof("Verifying %s using %s", scope, model)

	prompt, err := renderPrompt(scope, opts.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to render prompt: %w", err)
	}

	raw, err := Execute(tool, prompt, model, opts.RepoPath, logger.V(2).Enabled())
	if err != nil {
		return nil, fmt.Errorf("CLI execution failed: %w", err)
	}

	result, err := parseVerifyResponse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	result.Score = ComputeOverallScore(result.Sections, opts.Config.Weights)
	return &result, nil
}

func ComputeOverallScore(sections []SectionResult, weights map[string]float64) int {
	if len(sections) == 0 {
		return 0
	}

	var totalWeight, weightedSum float64
	for _, s := range sections {
		w := 1.0
		if weights != nil {
			if ww, ok := weights[s.Name]; ok {
				w = ww
			}
		}
		totalWeight += w
		weightedSum += float64(s.Score) * w
	}

	if totalWeight == 0 {
		return 0
	}
	return int(math.Round(weightedSum / totalWeight))
}
