package git

import (
	"context"
	_ "embed"
	"fmt"
	"sort"
	"strings"

	"github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/internal/prompting"
	"github.com/flanksource/gavel/models"
	"github.com/ghodss/yaml"
)

//go:embed ai-summary-group.prompt
var summaryGroupPrompt string

type AISummaryOutput struct {
	Name        string `yaml:"name,omitempty" json:"name,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

func renderSummaryPrompt(scope models.ScopeType, window string, commits models.CommitAnalyses) (string, error) {
	if summaryGroupPrompt == "" {
		return "", fmt.Errorf("AI summary group prompt template is empty")
	}

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

	req, _, err := prompt.Load(summaryGroupPrompt).Render(map[string]any{
		"window":  window,
		"scope":   scope,
		"commits": commitMaps,
		"files":   files,
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to render AI prompt template: %w", err)
	}
	return req.Prompt, nil
}

func GenerateGroupSummary(ctx context.Context, scope models.ScopeType, window string, commits models.CommitAnalyses, agent ai.Agent) (string, string, error) {
	promptText, err := renderSummaryPrompt(scope, window, commits)
	if err != nil {
		return "", "", err
	}

	prompting.Prepare()
	resp, err := agent.ExecutePrompt(ctx, ai.PromptRequest{
		Name:   fmt.Sprintf("Summary: %s - %s", scope, window),
		Prompt: promptText,
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
