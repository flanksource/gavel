package commit

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/flanksource/captain/pkg/ai/prompt"
	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/internal/prompting"
)

// PRCommitInput is the minimal commit description GeneratePRContent needs.
// It exists so callers outside the commit package can build PR content
// without constructing a full CommitResult.
type PRCommitInput struct {
	Message string
	Files   []string
}

type PRContentInput struct {
	Commits []PRCommitInput
	// PromptOverride is the resolved .gavel.yaml commit.prContentPrompt override
	// (inline text or file contents); empty uses the embedded default template.
	PromptOverride string
}

type PRContent struct {
	Title  string
	Body   string
	Branch string
}

const maxPRTitleRunes = 40

func commitInputsFromResults(commits []CommitResult) []PRCommitInput {
	out := make([]PRCommitInput, len(commits))
	for i, c := range commits {
		out[i] = PRCommitInput{Message: c.Message, Files: c.Files}
	}
	return out
}

type prContentSchema struct {
	Title  string `json:"title" description:"PR title: imperative, <=40 characters, conventional-commit style when applicable"`
	Body   string `json:"body,omitempty" description:"Markdown body summarising what changed and why; may be empty for trivial PRs"`
	Branch string `json:"branch" description:"Suggested branch name: kebab-case, <=40 chars, conventional-commit type prefix (feat/, fix/, chore/, refactor/, docs/) when the commits share a type. Use only [a-z0-9/-]"`
}

//go:embed pr-content.prompt
var prContentPromptTemplate string

func GeneratePRContent(ctx context.Context, agent clickyai.Agent, in PRContentInput) (PRContent, error) {
	if len(in.Commits) == 0 {
		return PRContent{}, fmt.Errorf("no commits to summarise")
	}

	schema := &prContentSchema{}
	req, err := renderPRContentPrompt(in, schema)
	if err != nil {
		return PRContent{}, err
	}

	prompting.Prepare()
	resp, err := agent.ExecutePrompt(ctx, req)
	if err != nil {
		return PRContent{}, fmt.Errorf("execute PR-content prompt: %w", err)
	}
	if resp.Error != "" {
		return PRContent{}, fmt.Errorf("PR-content prompt returned error: %s", resp.Error)
	}

	title := strings.TrimSpace(schema.Title)
	if err := validatePRTitle(title, resp.Result); err != nil {
		return PRContent{}, err
	}

	branch := sanitizeBranchName(strings.TrimSpace(schema.Branch))
	if branch == "" {
		return PRContent{}, fmt.Errorf("PR-content prompt returned empty/invalid branch (raw: %q)", schema.Branch)
	}

	return PRContent{
		Title:  title,
		Body:   strings.TrimSpace(schema.Body),
		Branch: branch,
	}, nil
}

func renderPRContentPrompt(in PRContentInput, schema *prContentSchema) (clickyai.PromptRequest, error) {
	template := prContentPromptTemplate
	if strings.TrimSpace(in.PromptOverride) != "" {
		template = in.PromptOverride
	}
	req, _, err := prompt.Load(template).Render(prContentPromptData(in), schema)
	if err != nil {
		return clickyai.PromptRequest{}, fmt.Errorf("render PR-content prompt: %w", err)
	}
	return clickyai.PromptRequest{
		Name:             "PR title and body",
		Prompt:           req.Prompt.User,
		SystemPrompt:     req.Prompt.System,
		StructuredOutput: req.Prompt.Schema,
	}, nil
}

func prContentPromptData(in PRContentInput) map[string]any {
	commits := make([]map[string]any, 0, len(in.Commits))
	for i, c := range in.Commits {
		commits = append(commits, map[string]any{
			"index":   i + 1,
			"message": strings.TrimSpace(c.Message),
			"files":   strings.Join(c.Files, ", "),
		})
	}
	return map[string]any{"commits": commits}
}

func validatePRTitle(title, raw string) error {
	if title == "" {
		return fmt.Errorf("PR-content prompt returned empty title (raw: %q)", raw)
	}
	if got := utf8.RuneCountInString(title); got > maxPRTitleRunes {
		return fmt.Errorf("PR-content prompt returned title longer than %d characters (%d): %q", maxPRTitleRunes, got, title)
	}
	return nil
}

// sanitizeBranchName trims the AI-suggested branch to a safe git ref:
// lowercase [a-z0-9/-], no leading/trailing slash or dash, no double
// slash, max 60 chars. Returns "" if nothing usable is left.
func sanitizeBranchName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '/', r == '-':
			b.WriteRune(r)
		case r == ' ' || r == '_':
			b.WriteRune('-')
		}
	}
	cleaned := b.String()
	for strings.Contains(cleaned, "//") {
		cleaned = strings.ReplaceAll(cleaned, "//", "/")
	}
	for strings.Contains(cleaned, "--") {
		cleaned = strings.ReplaceAll(cleaned, "--", "-")
	}
	cleaned = strings.Trim(cleaned, "/-")
	if len(cleaned) > 60 {
		cleaned = strings.TrimRight(cleaned[:60], "/-")
	}
	return cleaned
}
