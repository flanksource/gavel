package git

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/internal/prompting"
	"github.com/flanksource/gavel/models"
)

//go:embed ai-commit-message.prompt
var commitMessagePrompt string

//go:embed ai-commit-functionality-removed.prompt
var functionalityRemovedPrompt string

//go:embed ai-commit-compatibility-issues.prompt
var compatibilityIssuesPrompt string

// Prompt source names, reported as PromptRequest.Source so the AI logging
// middleware identifies which template produced each request under -v.
const (
	commitMessagePromptFile        = "ai-commit-message.prompt"
	functionalityRemovedPromptFile = "ai-commit-functionality-removed.prompt"
	compatibilityIssuesPromptFile  = "ai-commit-compatibility-issues.prompt"
)

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

type CommitPromptAnalysis struct {
	Commit               models.CommitAnalysis
	FunctionalityRemoved []string
	CompatibilityIssues  []string
}

func AnalyzeWithAI(ctx context.Context, commit models.CommitAnalysis, agent ai.Agent, opts AnalyzeOptions) (models.CommitAnalysis, error) {
	if opts.MinScore > 0 && commit.QualityScore >= opts.MinScore {
		return commit, nil
	}

	analyzed, err := analyzeCommitMessageWithAI(ctx, commit, agent, opts.MaxBodyLines, opts.MessagePrompt)
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
		analyzed, err := analyzeCommitMessageWithAI(ctx, out.Commit, agent, opts.MaxBodyLines, opts.MessagePrompt)
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

	functionalityRemoved, err := analyzeFunctionalityRemovedWithAI(ctx, out.Commit, agent, opts.FunctionalityRemovedPrompt)
	if err != nil {
		return out, err
	}
	out.FunctionalityRemoved = functionalityRemoved

	compatibilityIssues, err := analyzeCompatibilityIssuesWithAI(ctx, out.Commit, agent, opts.CompatibilityPrompt)
	if err != nil {
		return out, err
	}
	out.CompatibilityIssues = compatibilityIssues

	return out, nil
}

func analyzeCommitMessageWithAI(ctx context.Context, commit models.CommitAnalysis, agent ai.Agent, maxBodyLines int, override string) (models.CommitAnalysis, error) {
	template := promptOrDefault(override, commitMessagePrompt)
	if template == "" {
		return commit, fmt.Errorf("AI commit message prompt template is empty")
	}

	promptText, schemaJSON, err := renderCommitPrompt(commit, template, map[string]any{"maxBodyLines": maxBodyLines})
	if err != nil {
		return commit, err
	}

	prompting.Prepare()
	resp, err := agent.ExecutePrompt(ctx, ai.PromptRequest{
		Name:       promptName(commit, "commit message"),
		Source:     commitMessagePromptFile,
		Prompt:     promptText,
		SchemaJSON: schemaJSON,
	})
	if err != nil {
		return commit, fmt.Errorf("execute AI commit message prompt: %w", err)
	}
	if resp.Error != "" {
		return commit, fmt.Errorf("AI commit message prompt returned error: %s", resp.Error)
	}

	var schema commitMessageSchema
	if err := ai.DecodeStructured(resp, &schema); err != nil {
		return commit, fmt.Errorf("decode AI commit message response: %w", err)
	}

	if strings.TrimSpace(schema.Subject) == "" {
		return commit, fmt.Errorf("AI analysis returned empty subject (raw text: %q)", truncate(resp.Result, 400))
	}

	if schema.Type != "" {
		commit.CommitType = models.CommitType(schema.Type)
	}
	if schema.Scope != "" {
		commit.Scope = models.ScopeType(schema.Scope)
	}
	// The model sometimes echoes the conventional-commit prefix into the
	// subject (e.g. type=chore, subject="chore: update deps"). Composing the
	// message would then double it ("chore: chore: ..."). Strip a redundant
	// leading prefix, recovering type/scope from it when the model left them
	// blank.
	commit.CommitType, commit.Scope, commit.Subject =
		dedupeCommitPrefix(commit.CommitType, commit.Scope, strings.TrimSpace(schema.Subject))
	if schema.Body != "" {
		commit.Body = strings.TrimSpace(schema.Body)
	}

	return commit, nil
}

func analyzeFunctionalityRemovedWithAI(ctx context.Context, commit models.CommitAnalysis, agent ai.Agent, override string) ([]string, error) {
	template := promptOrDefault(override, functionalityRemovedPrompt)
	if template == "" {
		return nil, fmt.Errorf("AI functionality-removed prompt template is empty")
	}

	promptText, schemaJSON, err := renderCommitPrompt(commit, template, nil)
	if err != nil {
		return nil, err
	}

	prompting.Prepare()
	resp, err := agent.ExecutePrompt(ctx, ai.PromptRequest{
		Name:       promptName(commit, "removed functionality"),
		Source:     functionalityRemovedPromptFile,
		Prompt:     promptText,
		SchemaJSON: schemaJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("execute AI functionality-removed prompt: %w", err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("AI functionality-removed prompt returned error: %s", resp.Error)
	}

	// parseStringArrayResult recovers the array from the raw JSON reply (a bare
	// array or a {"functionalityRemoved": [...]} object), so no separate decode.
	return parseStringArrayResult(nil, resp.Result, "functionalityRemoved"), nil
}

func analyzeCompatibilityIssuesWithAI(ctx context.Context, commit models.CommitAnalysis, agent ai.Agent, override string) ([]string, error) {
	template := promptOrDefault(override, compatibilityIssuesPrompt)
	if template == "" {
		return nil, fmt.Errorf("AI compatibility-issues prompt template is empty")
	}

	promptText, schemaJSON, err := renderCommitPrompt(commit, template, nil)
	if err != nil {
		return nil, err
	}

	prompting.Prepare()
	resp, err := agent.ExecutePrompt(ctx, ai.PromptRequest{
		Name:       promptName(commit, "compatibility issues"),
		Source:     compatibilityIssuesPromptFile,
		Prompt:     promptText,
		SchemaJSON: schemaJSON,
	})
	if err != nil {
		return nil, fmt.Errorf("execute AI compatibility-issues prompt: %w", err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("AI compatibility-issues prompt returned error: %s", resp.Error)
	}

	// parseStringArrayResult recovers the array from the raw JSON reply.
	return parseStringArrayResult(nil, resp.Result, "compatibilityIssues"), nil
}

// promptOrDefault returns the .gavel.yaml override when set, otherwise the
// embedded default template.
func promptOrDefault(override, embedded string) string {
	if strings.TrimSpace(override) != "" {
		return override
	}
	return embedded
}

// renderCommitPrompt renders the template body and returns the user prompt plus
// the JSON output schema declared in the .prompt frontmatter (empty when none).
func renderCommitPrompt(commit models.CommitAnalysis, template string, extra map[string]any) (string, json.RawMessage, error) {
	data := commit.AsMap()
	for k, v := range extra {
		data[k] = v
	}
	req, _, err := prompt.Load(template).Render(data, nil)
	if err != nil {
		return "", nil, fmt.Errorf("render AI prompt template: %w", err)
	}
	return req.Prompt.User, req.Prompt.SchemaJSON, nil
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
