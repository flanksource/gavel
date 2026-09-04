package git

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/flanksource/captain/pkg/ai/prompt"
	"github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/models"
)

//go:embed ai-commit-message.prompt
var commitMessagePrompt string

// commitMessagePromptFile is reported as PromptRequest.Source so the AI logging
// middleware identifies which template produced each request under -v.
const commitMessagePromptFile = "ai-commit-message.prompt"

// commitMessageSchema is the structured-output schema handed to the LLM.
// Fields are ordered to match the expected conventional-commit layout.
// Subject is required (no json omitempty) so the provider's schema
// generator marks it required, causing the model to always emit it.
type commitMessageSchema struct {
	Type    string `json:"type" description:"Conventional commit type, one of the values enumerated in the schema"`
	Scope   string `json:"scope,omitempty" description:"Optional scope, e.g. db, api, fe, kubernetes"`
	Subject string `json:"subject" description:"Imperative subject line, max 100 chars, no trailing period"`
	Body    string `json:"body,omitempty" description:"Optional body explaining why and impact"`
}

func AnalyzeWithAI(ctx context.Context, commit models.CommitAnalysis, agent ai.Agent, opts AnalyzeOptions) (models.CommitAnalysis, error) {
	if opts.MinScore > 0 && commit.QualityScore >= opts.MinScore {
		return commit, nil
	}

	analyzed, err := analyzeCommitMessageWithAI(ctx, commit, agent, opts.MaxBodyLines, opts.MessagePrompt, opts.AllowedCommitTypes)
	if err != nil {
		return commit, err
	}
	if analyzed.Trailers == nil {
		analyzed.Trailers = map[string]string{}
	}
	analyzed.Trailers["AI-Analyzed"] = "true"
	return analyzed, nil
}

func analyzeCommitMessageWithAI(ctx context.Context, commit models.CommitAnalysis, agent ai.Agent, maxBodyLines int, override string, configuredTypes []string) (models.CommitAnalysis, error) {
	template := promptOrDefault(override, commitMessagePrompt)
	if template == "" {
		return commit, fmt.Errorf("AI commit message prompt template is empty")
	}

	allowedTypes, err := allowedCommitTypes(configuredTypes)
	if err != nil {
		return commit, err
	}

	promptText, schemaJSON, err := renderCommitPrompt(commit, template, maxBodyLines, allowedTypes)
	if err != nil {
		return commit, err
	}

	// No prompting.Prepare() here. It waits for the global clicky task manager to
	// drain, which is correct when a caller owns the terminal (gavel commit) but a
	// deadlock on the git analyze path: AnalyzeCommitHistory calls this from inside
	// a clicky batch item, so the item would be waiting for a task set that
	// includes itself. Callers that own the terminal Prepare before calling in.
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
		if !slices.Contains(allowedTypes, schema.Type) {
			return commit, fmt.Errorf("AI analysis returned commit type %q, which is not one of %s",
				truncate(schema.Type, 60), strings.Join(allowedTypes, "|"))
		}
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

// commitPromptData is the template data for the commit-message prompt's body:
// the {{maxBodyLines}} branch, plus the type vocabulary as the prose
// requirement lists it ({{commitTypesList}}). The schema's `type` enum is not
// template data — see enumerateCommitType.
func commitPromptData(maxBodyLines int, allowedTypes []string) map[string]any {
	return map[string]any{
		"maxBodyLines":    maxBodyLines,
		"commitTypesList": strings.Join(allowedTypes, "|"),
	}
}

// allowedCommitTypes resolves the conventional-commit types the model may pick
// from: the `.gavel.yaml` commit.types list when a project names one, otherwise
// gavel's defaults. They become the prompt schema's `type` enum, so the model
// picks from the project's vocabulary instead of inventing one. An unrecognised
// configured type is an error rather than a silent drop — quietly ignoring it
// would generate messages the project never asked for.
func allowedCommitTypes(configured []string) ([]string, error) {
	known := commitTypeNames(models.SelectableCommitTypes())
	if len(configured) == 0 {
		return known, nil
	}
	out := make([]string, 0, len(configured))
	for _, name := range configured {
		name = strings.TrimSpace(name)
		if !models.CommitType(name).IsValid() {
			return nil, fmt.Errorf("commit.types: unknown commit type %q (known types: %s)",
				name, strings.Join(known, ", "))
		}
		out = append(out, name)
	}
	return out, nil
}

func commitTypeNames(types []models.CommitType) []string {
	out := make([]string, len(types))
	for i, t := range types {
		out[i] = string(t)
	}
	return out
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
// the JSON output schema declared in the .prompt frontmatter (empty when none),
// with the allowed commit types enumerated on the schema's `type` property.
func renderCommitPrompt(commit models.CommitAnalysis, template string, maxBodyLines int, allowedTypes []string) (string, json.RawMessage, error) {
	data := commit.AsMap()
	for k, v := range commitPromptData(maxBodyLines, allowedTypes) {
		data[k] = v
	}
	req, _, err := prompt.Load(template).Render(data, nil)
	if err != nil {
		return "", nil, fmt.Errorf("render AI prompt template: %w", err)
	}
	schema, err := enumerateCommitType(req.Prompt.SchemaJSON, allowedTypes)
	if err != nil {
		return "", nil, err
	}
	return req.Prompt.User, schema, nil
}

// enumerateCommitType constrains the output schema's `type` property to the
// project's commit-type vocabulary, so the model picks from it rather than
// echoing a list. The enum is set on the parsed schema rather than templated
// into the frontmatter: frontmatter is parsed as YAML before any templating
// runs, so a Handlebars block inside it is a document that never parses — for
// the settings editor as much as for the run.
//
// A template that declares no output schema, or one whose schema has no `type`
// property, has nothing to constrain and is returned as it is; the response
// check in analyzeCommitMessageWithAI still refuses a type outside the
// vocabulary.
func enumerateCommitType(schemaJSON json.RawMessage, allowedTypes []string) (json.RawMessage, error) {
	if len(schemaJSON) == 0 {
		return schemaJSON, nil
	}
	var schema map[string]any
	if err := json.Unmarshal(schemaJSON, &schema); err != nil {
		return nil, fmt.Errorf("decode commit prompt output schema: %w", err)
	}
	properties, _ := schema["properties"].(map[string]any)
	typeProperty, ok := properties["type"].(map[string]any)
	if !ok {
		return schemaJSON, nil
	}
	typeProperty["enum"] = allowedTypes
	out, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("encode commit prompt output schema: %w", err)
	}
	return out, nil
}

func promptName(commit models.CommitAnalysis, suffix string) string {
	name := strings.TrimSpace(commit.PrettySubject().String())
	if name == "" {
		name = "commit diff"
	}
	return fmt.Sprintf("%s: %s", suffix, name)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
