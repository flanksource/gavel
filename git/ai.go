package git

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/flanksource/clicky/ai"
	"github.com/flanksource/commons/logger"
	. "github.com/flanksource/gavel/models"
	"github.com/flanksource/gomplate/v3"
	"github.com/ghodss/yaml"
)

//go:embed ai-commit-analyze.md
var commitAnalysisPrompt string

func AnalyzeWithAI(ctx context.Context, commit CommitAnalysis, agent ai.Agent, opts AnalyzeOptions) (CommitAnalysis, error) {

	if opts.MinScore > 0 && commit.QualityScore >= opts.MinScore {
		return commit, nil
	}

	if commitAnalysisPrompt == "" {
		return commit, fmt.Errorf("AI commit analysis prompt template is empty")
	}

	prompt, err := gomplate.RunTemplate(commit.AsMap(), gomplate.Template{
		Template: commitAnalysisPrompt,
	})
	if err != nil {
		return commit, fmt.Errorf("failed to render AI prompt template: %w", err)
	}

	resp, err := agent.ExecutePrompt(ctx, ai.PromptRequest{
		Name:   commit.PrettySubject().String(),
		Prompt: prompt,
	})
	if err != nil {
		return commit, fmt.Errorf("AI prompt execution failed: %w", err)
	}

	aiAnalysis := AIAnalysisOutput{}

	if strings.HasPrefix(resp.Result, "```yaml") {
		// Trim code block markers if present
		resp.Result = strings.TrimPrefix(resp.Result, "```yaml")

	} else if strings.HasPrefix(resp.Result, "```") {
		resp.Result = strings.TrimPrefix(resp.Result, "```")
	}

	resp.Result = strings.TrimSuffix(resp.Result, "```")
	if err := yaml.Unmarshal([]byte(resp.Result), &aiAnalysis); err == nil {
		if aiAnalysis.Type != "" {
			commit.CommitType = aiAnalysis.Type
		}
		if aiAnalysis.Scope != "" {
			commit.Scope = aiAnalysis.Scope
		}
		if aiAnalysis.Subject != "" {
			commit.Subject = strings.TrimSpace(aiAnalysis.Subject)
		}
		if aiAnalysis.Body != "" {
			commit.Body = strings.TrimSpace(aiAnalysis.Body)
		}
	} else {
		logger.Warnf("Failed to parse AI analysis output as YAML: %v", resp.Result)

		lines := strings.Split(resp.Result, "\n")
		for i, line := range lines {
			if i == 0 {
				commit.CommitType, commit.Scope, commit.Subject = parseCommitTypeAndScope(strings.TrimSpace(line))
			} else {
				commit.Body += strings.TrimSpace(line) + "\n"
			}
		}
	}

	commit.Trailers["AI-Analyzed"] = "true"

	return commit, nil
}
