package commit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flanksource/captain/pkg/api"
	clickyai "github.com/flanksource/gavel/ai"
)

type prContentPromptAgent struct {
	t    *testing.T
	req  clickyai.PromptRequest
	fill prContentSchema
}

func (a *prContentPromptAgent) ExecutePrompt(_ context.Context, req clickyai.PromptRequest) (*clickyai.PromptResponse, error) {
	a.req = req
	require.NotEmpty(a.t, req.SchemaJSON, "PR-content schema should come from the .prompt frontmatter")
	raw, err := json.Marshal(a.fill)
	require.NoError(a.t, err)
	return &clickyai.PromptResponse{StructuredData: json.RawMessage(raw)}, nil
}

func (a *prContentPromptAgent) ExecuteBatch(context.Context, []clickyai.PromptRequest) (map[string]*clickyai.PromptResponse, error) {
	return nil, fmt.Errorf("unexpected batch execution")
}

func (a *prContentPromptAgent) GetCosts() clickyai.Costs { return nil }
func (a *prContentPromptAgent) Close() error             { return nil }

func TestGeneratePRContentRendersCaptainPromptAndSchema(t *testing.T) {
	agent := &prContentPromptAgent{
		t: t,
		fill: prContentSchema{
			Title:  "fix: use prompts",
			Body:   "## What\n- uses Captain prompts",
			Branch: "fix/use-prompts",
		},
	}

	got, err := GeneratePRContent(context.Background(), agent, PRContentInput{Commits: []PRCommitInput{
		{
			Message: "fix: use Captain prompt for PRs",
			Files:   []string{"commit/push_prompt.go", "commit/pr-content.prompt"},
		},
	}})
	require.NoError(t, err)

	assert.Equal(t, PRContent{
		Title:  "fix: use prompts",
		Body:   "## What\n- uses Captain prompts",
		Branch: "fix/use-prompts",
	}, got)

	assert.Equal(t, "PR title and body", agent.req.Name)
	assert.Contains(t, agent.req.Prompt, "Return structured output matching the provided output schema.")
	assert.Contains(t, agent.req.Prompt, "Title: imperative mood, <= 40 characters")
	assert.Contains(t, agent.req.Prompt, "--- commit 1 ---")
	assert.Contains(t, agent.req.Prompt, "fix: use Captain prompt for PRs")
	assert.Contains(t, agent.req.Prompt, "files: commit/push_prompt.go, commit/pr-content.prompt")
	assert.NotContains(t, agent.req.Prompt, "%s")
	assert.NotEmpty(t, agent.req.SchemaJSON, "schema should be carried as SchemaJSON from the frontmatter")
	assert.Equal(t, api.SchemaStrictnessRetry, agent.req.SchemaStrictness,
		"schemaStrictness: retry from the frontmatter must be forwarded so a schema violation is fixed by re-asking the model")
}

func TestGeneratePRContentRejectsLongTitle(t *testing.T) {
	agent := &prContentPromptAgent{
		t: t,
		fill: prContentSchema{
			Title:  "fix: this title is definitely longer than forty characters",
			Body:   "## What\n- too long",
			Branch: "fix/long-title",
		},
	}

	_, err := GeneratePRContent(context.Background(), agent, PRContentInput{Commits: []PRCommitInput{
		{Message: "fix: test title limit"},
	}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "longer than 40 characters")
}

func TestValidatePRTitleCountsRunes(t *testing.T) {
	assert.NoError(t, validatePRTitle(strings.Repeat("界", maxPRTitleRunes), "{}"))
	assert.ErrorContains(t, validatePRTitle(strings.Repeat("界", maxPRTitleRunes+1), "{}"), "longer than 40 characters")
}
