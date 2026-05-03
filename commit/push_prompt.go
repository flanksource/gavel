package commit

import (
	"context"
	"fmt"
	"strings"

	clickyai "github.com/flanksource/clicky/ai"
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
}

type PRContent struct {
	Title  string
	Body   string
	Branch string
}

func commitInputsFromResults(commits []CommitResult) []PRCommitInput {
	out := make([]PRCommitInput, len(commits))
	for i, c := range commits {
		out[i] = PRCommitInput{Message: c.Message, Files: c.Files}
	}
	return out
}

type prContentSchema struct {
	Title  string `json:"title" description:"PR title: imperative, <=70 chars, conventional-commit style when applicable"`
	Body   string `json:"body,omitempty" description:"Markdown body summarising what changed and why; may be empty for trivial PRs"`
	Branch string `json:"branch" description:"Suggested branch name: kebab-case, <=40 chars, conventional-commit type prefix (feat/, fix/, chore/, refactor/, docs/) when the commits share a type. Use only [a-z0-9/-]"`
}

const prContentPromptTemplate = `You are opening a GitHub pull request for the following local commits.
Generate a concise PR title, a short markdown body, and a branch name.

Guidelines:
- Title: imperative mood, <= 70 characters, prefer conventional-commit style when the commits share a type.
- Body: 1-4 short sections (What / Why / Notes). Bullet lists over prose. Omit sections that add no value.
- Branch: kebab-case, <= 40 characters, conventional-commit type prefix (feat/, fix/, chore/, refactor/, docs/) when the commits share a type. Use only [a-z0-9/-]. Example: "feat/user-auth-rate-limit".
- Do NOT invent context that isn't supported by the commit messages.

Commits (in order):
%s
`

func GeneratePRContent(ctx context.Context, agent clickyai.Agent, in PRContentInput) (PRContent, error) {
	if len(in.Commits) == 0 {
		return PRContent{}, fmt.Errorf("no commits to summarise")
	}

	var b strings.Builder
	for i, c := range in.Commits {
		fmt.Fprintf(&b, "--- commit %d ---\n%s\n", i+1, strings.TrimSpace(c.Message))
		if len(c.Files) > 0 {
			fmt.Fprintf(&b, "files: %s\n", strings.Join(c.Files, ", "))
		}
		b.WriteString("\n")
	}

	prompt := fmt.Sprintf(prContentPromptTemplate, b.String())

	schema := &prContentSchema{}
	prompting.Prepare()
	resp, err := agent.ExecutePrompt(ctx, clickyai.PromptRequest{
		Name:             "PR title and body",
		Prompt:           prompt,
		StructuredOutput: schema,
	})
	if err != nil {
		return PRContent{}, fmt.Errorf("execute PR-content prompt: %w", err)
	}
	if resp.Error != "" {
		return PRContent{}, fmt.Errorf("PR-content prompt returned error: %s", resp.Error)
	}
	if strings.TrimSpace(schema.Title) == "" {
		return PRContent{}, fmt.Errorf("PR-content prompt returned empty title (raw: %q)", resp.Result)
	}

	branch := sanitizeBranchName(strings.TrimSpace(schema.Branch))
	if branch == "" {
		return PRContent{}, fmt.Errorf("PR-content prompt returned empty/invalid branch (raw: %q)", schema.Branch)
	}

	return PRContent{
		Title:  strings.TrimSpace(schema.Title),
		Body:   strings.TrimSpace(schema.Body),
		Branch: branch,
	}, nil
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
