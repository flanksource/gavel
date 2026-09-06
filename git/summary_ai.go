package git

import (
	"context"
	_ "embed"
	"fmt"
	"sort"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/prompts"
	"github.com/ghodss/yaml"
)

//go:embed ai-summary-group.prompt
var summaryGroupPrompt string

type AISummaryOutput struct {
	Name        string `yaml:"name,omitempty" json:"name,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

func summaryPromptData(scope models.ScopeType, window string, commits models.CommitAnalyses) map[string]any {
	filesSet := make(map[string]struct{})
	for _, commit := range commits {
		for _, change := range commit.Changes {
			filesSet[change.File] = struct{}{}
		}
	}

	files := make([]string, 0, len(filesSet))
	for file := range filesSet {
		files = append(files, file)
	}
	sort.Strings(files)

	commitMaps := make([]map[string]any, 0, len(commits))
	for _, commit := range commits {
		commitMaps = append(commitMaps, commit.AsMap())
	}

	return map[string]any{
		"window":  window,
		"scope":   scope,
		"commits": commitMaps,
		"files":   files,
	}
}

func prepareGroupSummary(scope models.ScopeType, window string, commits models.CommitAnalyses, opts SummaryOptions) (captaincli.AIRuntimeResolved, error) {
	options := opts.PromptOptions
	options.DefaultPrompt = summaryGroupPrompt
	options.Name = prompts.CommitSummary
	options.RequireModel = true
	options.Saved = &opts.Saved.AI
	options.Normalize = func(spec api.Spec) (api.SpecNormalization, error) {
		return (captaincli.AIRuntimeOptions{}).Normalize(captaincli.AIRuntimeNormalizeOptions{Spec: spec, Saved: opts.Saved, Cwd: options.Dir})
	}
	options.Data = summaryPromptData(scope, window, commits)
	resolved, err := opts.Prompt.Resolve(options)
	if err != nil {
		return captaincli.AIRuntimeResolved{}, err
	}
	if resolved.Spec.Prompt.Source == "" {
		resolved.Spec.Prompt.Source = "ai-summary-group.prompt"
	}
	return (captaincli.AIRuntimeOptions{}).Project(captaincli.AIRuntimeProjectOptions{Resolved: resolved, Saved: opts.Saved})
}

func GenerateGroupSummary(ctx context.Context, scope models.ScopeType, window string, commits models.CommitAnalyses, options SummaryOptions) (string, string, error) {
	prepared, err := prepareGroupSummary(scope, window, commits, options)
	if err != nil {
		return "", "", err
	}

	resp, err := executePreparedPrompt(ctx, preparedPromptOptions{
		Name: fmt.Sprintf("Summary: %s - %s", scope, window), Runtime: prepared, Agent: options.Agent, AgentFactory: options.AgentFactory,
	})
	if err != nil {
		logger.Warnf("AI prompt execution failed: %v", err)
		return "", "", err
	}

	aiOutput := AISummaryOutput{}

	// Handle code block markers
	result := resp.Result
	if strings.HasPrefix(result, "```yaml") {
		result = strings.TrimPrefix(result, "```yaml")
	} else if strings.HasPrefix(result, "```") {
		result = strings.TrimPrefix(result, "```")
	}
	result = strings.TrimSuffix(result, "```")
	result = strings.TrimSpace(result)

	if err := yaml.Unmarshal([]byte(result), &aiOutput); err != nil {
		logger.Warnf("Failed to parse AI summary output as YAML: %v, raw output: %s", err, result)
		return "", "", fmt.Errorf("failed to parse AI output: %w", err)
	}

	return aiOutput.Name, aiOutput.Description, nil
}
