package commit

import (
	"context"
	"fmt"
	"strings"

	clickyai "github.com/flanksource/clicky/ai"
	"github.com/flanksource/gavel/internal/prompting"
)

type prContentInput struct {
	commits []CommitResult
}

type prContent struct {
	Title string
	Body  string
}

type prContentSchema struct {
	Title string `json:"title" description:"PR title: imperative, <=70 chars, conventional-commit style when applicable"`
	Body  string `json:"body,omitempty" description:"Markdown body summarising what changed and why; may be empty for trivial PRs"`
}

const prContentPromptTemplate = `You are opening a GitHub pull request for the following local commits.
Generate a concise PR title and a short markdown body.

Guidelines:
- Title: imperative mood, <= 70 characters, prefer conventional-commit style when the commits share a type.
- Body: 1-4 short sections (What / Why / Notes). Bullet lists over prose. Omit sections that add no value.
- Do NOT invent context that isn't supported by the commit messages.

Commits (in order):
%s
`

func generatePRContent(ctx context.Context, agent clickyai.Agent, in prContentInput) (prContent, error) {
	if len(in.commits) == 0 {
		return prContent{}, fmt.Errorf("no commits to summarise")
	}

	var b strings.Builder
	for i, c := range in.commits {
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
		return prContent{}, fmt.Errorf("execute PR-content prompt: %w", err)
	}
	if resp.Error != "" {
		return prContent{}, fmt.Errorf("PR-content prompt returned error: %s", resp.Error)
	}
	if strings.TrimSpace(schema.Title) == "" {
		return prContent{}, fmt.Errorf("PR-content prompt returned empty title (raw: %q)", resp.Result)
	}

	return prContent{Title: strings.TrimSpace(schema.Title), Body: strings.TrimSpace(schema.Body)}, nil
}
