package commit

import (
	"context"
	"testing"

	clickyai "github.com/flanksource/clicky/ai"
	"github.com/flanksource/gavel/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSingleCommitPrecommitAndCompatFalseBypassChecks(t *testing.T) {
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
		t.Fatal("AI agent should not be created when message is explicit and compat is false")
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
		CompatMode:    IgnoreCheckModeFalse,
		Config: verify.CommitConfig{
			GitIgnore: []string{"*.env"},
		},
	})
	require.NoError(t, err)
	require.Len(t, result.Commits, 1)
	assert.Equal(t, "chore: keep staged files", result.Commits[0].Message)
	assert.ElementsMatch(t, []string{".env", "package.json"}, result.Staged)
}

func TestResolvePrecommitModeUsesNewConfigBeforeLegacyFallback(t *testing.T) {
	mode, err := resolvePrecommitMode("", verify.CommitConfig{
		Precommit:  verify.PrecommitConfig{Mode: "fail"},
		LinkedDeps: verify.LinkedDepsConfig{Mode: "skip"},
	})
	require.NoError(t, err)
	assert.Equal(t, CheckModeFail, mode)

	mode, err = resolvePrecommitMode("", verify.CommitConfig{
		LinkedDeps: verify.LinkedDepsConfig{Mode: "false"},
	})
	require.NoError(t, err)
	assert.Equal(t, CheckModeSkip, mode)
}

func TestResolveCompatModeDefaultsToSkip(t *testing.T) {
	mode, err := resolveCompatMode("", verify.CommitConfig{})
	require.NoError(t, err)
	assert.Equal(t, CheckModeSkip, mode)
}

func TestResolveCompatModeUsesConfiguredValueWhenPresent(t *testing.T) {
	mode, err := resolveCompatMode("", verify.CommitConfig{
		Compatibility: verify.CompatibilityConfig{Mode: "fail"},
	})
	require.NoError(t, err)
	assert.Equal(t, CheckModeFail, mode)
}
