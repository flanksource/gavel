package commit

import (
	"context"
	"testing"

	clickyai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSingleCommitPrecommitFalseBypassChecks(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, ".env", "SECRET=1\n")
	writeFile(t, repo, "package.json", `{
  "name": "app",
  "dependencies": {
    "sibling": "file:../sibling"
  }
}
`)
	gitRun(t, repo, "add", ".env", "package.json")

	previousAgent := newAgentFunc
	newAgentFunc = func(clickyai.AgentConfig) (clickyai.Agent, error) {
		t.Fatal("AI agent should not be created when the message is explicit")
		return nil, nil
	}
	defer func() {
		newAgentFunc = previousAgent
	}()

	result, err := Run(context.Background(), Options{
		WorkDir:       repo,
		Message:       "chore: keep staged files",
		DryRun:        true,
		PrecommitMode: IgnoreCheckModeFalse,
		Config: verify.CommitConfig{
			GitIgnore: []string{"*.env"},
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Commits, 1)
	assert.Equal(t, "chore: keep staged files", result.Commits[0].Message)
	assert.ElementsMatch(t, []string{".env", "package.json"}, result.Staged)
}

func TestResolvePrecommitModeUsesConfigThenDefault(t *testing.T) {
	mode, err := resolvePrecommitMode("", verify.CommitConfig{
		Precommit: verify.PrecommitConfig{Mode: "fail"},
	})
	require.NoError(t, err)
	assert.Equal(t, CheckModeFail, mode)

	// A raw --precommit flag wins over the config value.
	mode, err = resolvePrecommitMode("skip", verify.CommitConfig{
		Precommit: verify.PrecommitConfig{Mode: "fail"},
	})
	require.NoError(t, err)
	assert.Equal(t, CheckModeSkip, mode)

	// The "false" alias maps to skip.
	mode, err = resolvePrecommitMode("", verify.CommitConfig{
		Precommit: verify.PrecommitConfig{Mode: "false"},
	})
	require.NoError(t, err)
	assert.Equal(t, CheckModeSkip, mode)

	// Empty flag and empty config default to prompt.
	mode, err = resolvePrecommitMode("", verify.CommitConfig{})
	require.NoError(t, err)
	assert.Equal(t, CheckModePrompt, mode)
}
