package git

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/flanksource/clicky/ai"
	. "github.com/flanksource/gavel/models"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gomplate/v3"
	"github.com/ghodss/yaml"
)

//go:embed ai-summary-group.md
var summaryGroupPrompt string

type AISummaryOutput struct {
	Name        string `yaml:"name,omitempty" json:"name,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

func GenerateGroupSummary(ctx context.Context, scope ScopeType, window string, commits CommitAnalyses, agent ai.Agent) (string, string, error) {
	if summaryGroupPrompt == "" {
		return "", "", fmt.Errorf("AI summary group prompt template is empty")
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

	templateData := map[string]any{
		"window":  window,
		"scope":   scope,
		"commits": commits,
		"files":   files,
	}

	prompt, err := gomplate.RunTemplate(templateData, gomplate.Template{
		Template: summaryGroupPrompt,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to render AI prompt template: %w", err)
	}

	resp, err := agent.ExecutePrompt(ctx, ai.PromptRequest{
		Name:   fmt.Sprintf("Summary: %s - %s", scope, window),
		Prompt: prompt,
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
