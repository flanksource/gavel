package git

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/clicky/ai"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/internal/prompting"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gomplate/v3"
)

//go:embed ai-commit-message.md
var commitMessagePrompt string

//go:embed ai-commit-functionality-removed.md
var functionalityRemovedPrompt string

//go:embed ai-commit-compatibility-issues.md
var compatibilityIssuesPrompt string

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

type functionalityRemovedSchema struct {
	FunctionalityRemoved []string `json:"functionalityRemoved" description:"User-visible functionality removed by this diff; empty when nothing is removed"`
}

type compatibilityIssuesSchema struct {
	CompatibilityIssues []string `json:"compatibilityIssues" description:"Backward compatibility issues or breaking changes introduced by this diff; empty when there are none"`
}

type CommitPromptAnalysis struct {
	Commit               models.CommitAnalysis
	FunctionalityRemoved []string
	CompatibilityIssues  []string
}

func AnalyzeWithAI(ctx context.Context, commit models.CommitAnalysis, agent ai.Agent, opts AnalyzeOptions) (models.CommitAnalysis, error) {
	if opts.MinScore > 0 && commit.QualityScore >= opts.MinScore {
		return commit, nil
	}

	analyzed, err := analyzeCommitMessageWithAI(ctx, commit, agent)
	if err != nil {
		return commit, err
	}
	if analyzed.Trailers == nil {
		analyzed.Trailers = map[string]string{}
	}
	analyzed.Trailers["AI-Analyzed"] = "true"
	return analyzed, nil
}

func AnalyzeCommitPromptsWithAI(ctx context.Context, commit models.CommitAnalysis, agent ai.Agent, includeMessage bool, opts AnalyzeOptions) (CommitPromptAnalysis, error) {
	out := CommitPromptAnalysis{Commit: commit}
	if opts.MinScore > 0 && commit.QualityScore >= opts.MinScore {
		return out, nil
	}

	if includeMessage {
		analyzed, err := analyzeCommitMessageWithAI(ctx, out.Commit, agent)
		if err != nil {
			return out, err
		}
		out.Commit = analyzed
	}

	compatibilityAnalysis, err := AnalyzeCompatibilityPromptsWithAI(ctx, out.Commit, agent, opts)
	if err != nil {
		return out, err
	}
	out.FunctionalityRemoved = compatibilityAnalysis.FunctionalityRemoved
	out.CompatibilityIssues = compatibilityAnalysis.CompatibilityIssues

	if out.Commit.Trailers == nil {
		out.Commit.Trailers = map[string]string{}
	}
	out.Commit.Trailers["AI-Analyzed"] = "true"

	return out, nil
}

func AnalyzeCompatibilityPromptsWithAI(ctx context.Context, commit models.CommitAnalysis, agent ai.Agent, opts AnalyzeOptions) (CommitPromptAnalysis, error) {
	out := CommitPromptAnalysis{Commit: commit}
	if opts.MinScore > 0 && commit.QualityScore >= opts.MinScore {
		return out, nil
	}

	functionalityRemoved, err := analyzeFunctionalityRemovedWithAI(ctx, out.Commit, agent)
	if err != nil {
		return out, err
	}
	out.FunctionalityRemoved = functionalityRemoved

	compatibilityIssues, err := analyzeCompatibilityIssuesWithAI(ctx, out.Commit, agent)
	if err != nil {
		return out, err
	}
	out.CompatibilityIssues = compatibilityIssues

	return out, nil
}

func analyzeCommitMessageWithAI(ctx context.Context, commit models.CommitAnalysis, agent ai.Agent) (models.CommitAnalysis, error) {
	if commitMessagePrompt == "" {
		return commit, fmt.Errorf("AI commit message prompt template is empty")
	}

	prompt, err := renderCommitPrompt(commit, commitMessagePrompt)
	if err != nil {
		return commit, err
	}

	schema := &commitMessageSchema{}
	prompting.Prepare()
	resp, err := agent.ExecutePrompt(ctx, ai.PromptRequest{
		Name:             promptName(commit, "commit message"),
		Prompt:           prompt,
		StructuredOutput: schema,
	})
	if err != nil {
		return commit, fmt.Errorf("execute AI commit message prompt: %w", err)
	}
	if resp.Error != "" {
		return commit, fmt.Errorf("AI commit message prompt returned error: %s", resp.Error)
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

	return commit, nil
}

func analyzeFunctionalityRemovedWithAI(ctx context.Context, commit models.CommitAnalysis, agent ai.Agent) ([]string, error) {
	if functionalityRemovedPrompt == "" {
		return nil, fmt.Errorf("AI functionality-removed prompt template is empty")
	}

	prompt, err := renderCommitPrompt(commit, functionalityRemovedPrompt)
	if err != nil {
		return nil, err
	}

	schema := &functionalityRemovedSchema{}
	prompting.Prepare()
	resp, err := agent.ExecutePrompt(ctx, ai.PromptRequest{
		Name:             promptName(commit, "removed functionality"),
		Prompt:           prompt,
		StructuredOutput: schema,
	})
	if err != nil {
		return nil, fmt.Errorf("execute AI functionality-removed prompt: %w", err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("AI functionality-removed prompt returned error: %s", resp.Error)
	}

	return parseStringArrayResult(schema.FunctionalityRemoved, resp.Result, "functionalityRemoved"), nil
}

func analyzeCompatibilityIssuesWithAI(ctx context.Context, commit models.CommitAnalysis, agent ai.Agent) ([]string, error) {
	if compatibilityIssuesPrompt == "" {
		return nil, fmt.Errorf("AI compatibility-issues prompt template is empty")
	}

	prompt, err := renderCommitPrompt(commit, compatibilityIssuesPrompt)
	if err != nil {
		return nil, err
	}

	schema := &compatibilityIssuesSchema{}
	prompting.Prepare()
	resp, err := agent.ExecutePrompt(ctx, ai.PromptRequest{
		Name:             promptName(commit, "compatibility issues"),
		Prompt:           prompt,
		StructuredOutput: schema,
	})
	if err != nil {
		return nil, fmt.Errorf("execute AI compatibility-issues prompt: %w", err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("AI compatibility-issues prompt returned error: %s", resp.Error)
	}

	return parseStringArrayResult(schema.CompatibilityIssues, resp.Result, "compatibilityIssues"), nil
}

func renderCommitPrompt(commit models.CommitAnalysis, template string) (string, error) {
	prompt, err := gomplate.RunTemplate(commit.AsMap(), gomplate.Template{
		Template: template,
	})
	if err != nil {
		return "", fmt.Errorf("render AI prompt template: %w", err)
	}
	return prompt, nil
}

func promptName(commit models.CommitAnalysis, suffix string) string {
	name := strings.TrimSpace(commit.PrettySubject().String())
	if name == "" {
		name = "commit diff"
	}
	return fmt.Sprintf("%s: %s", suffix, name)
}

func normalizeItems(items []string) []string {
	if len(items) == 0 {
		return nil
	}

	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func parseStringArrayResult(schemaItems []string, raw, fieldName string) []string {
	items := normalizeItems(schemaItems)
	if len(items) > 0 {
		return items
	}

	raw = unwrapJSONCodeFence(strings.TrimSpace(raw))
	if raw == "" {
		return nil
	}

	var parsed []string
	if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
		return normalizeItems(parsed)
	}

	var wrapped map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &wrapped); err != nil {
		return nil
	}

	field, ok := wrapped[fieldName]
	if !ok {
		return nil
	}
	if err := json.Unmarshal(field, &parsed); err != nil {
		return nil
	}
	return normalizeItems(parsed)
}

func unwrapJSONCodeFence(raw string) string {
	if !strings.HasPrefix(raw, "```") {
		return raw
	}

	lines := strings.Split(raw, "\n")
	if len(lines) == 0 {
		return raw
	}
	if strings.HasPrefix(lines[0], "```") {
		lines = lines[1:]
	}
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "```" {
		lines = lines[:len(lines)-1]
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
