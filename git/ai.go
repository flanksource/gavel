package git

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/flanksource/clicky/ai"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gomplate/v3"
)

//go:embed ai-commit-analyze.md
var commitAnalysisPrompt string

// commitMessageSchema is the structured-output schema handed to the LLM.
// Fields are ordered to match the expected conventional-commit layout.
// Subject is required (no json omitempty) so the provider's schema
// generator marks it required, causing the model to always emit it.
type commitMessageSchema struct {
	Type    string `json:"type" description:"Conventional commit type: feat|fix|perf|refactor|test|docs|build|ci|chore|revert"`
	Scope   string `json:"scope,omitempty" description:"Optional scope, e.g. db, api, fe, kubernetes"`
	Subject string `json:"subject" description:"Imperative subject line, max 100 chars, no trailing period"`
	Body    string `json:"body,omitempty" description:"Optional body explaining why and impact"`
}

func AnalyzeWithAI(ctx context.Context, commit models.CommitAnalysis, agent ai.Agent, opts AnalyzeOptions) (models.CommitAnalysis, error) {
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
		return commit, fmt.Errorf("render AI prompt template: %w", err)
	}

	schema := &commitMessageSchema{}
	resp, err := agent.ExecutePrompt(ctx, ai.PromptRequest{
		Name:             commit.PrettySubject().String(),
		Prompt:           prompt,
		StructuredOutput: schema,
	})
	if err != nil {
		return commit, fmt.Errorf("execute AI prompt: %w", err)
	}
	if resp.Error != "" {
		return commit, fmt.Errorf("AI prompt returned error: %s", resp.Error)
	}

	logger.V(2).Infof("AI commit analysis structured response: type=%q scope=%q subject=%q body.len=%d",
		schema.Type, schema.Scope, schema.Subject, len(schema.Body))

	if strings.TrimSpace(schema.Subject) == "" {
		return commit, fmt.Errorf("AI analysis returned empty subject (raw text: %q)", truncate(resp.Result, 400))
	}

	if schema.Type != "" {
		commit.CommitType = models.CommitType(schema.Type)
	}
	if schema.Scope != "" {
		commit.Scope = models.ScopeType(schema.Scope)
	}
	commit.Subject = strings.TrimSpace(schema.Subject)
	if schema.Body != "" {
		commit.Body = strings.TrimSpace(schema.Body)
	}

	if commit.Trailers == nil {
		commit.Trailers = map[string]string{}
	}
	commit.Trailers["AI-Analyzed"] = "true"

	return commit, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
