package commit

import (
	"bytes"
	"context"
	"strings"
	"testing"

	clickyai "github.com/flanksource/clicky/ai"
	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCompatibilityCheckContinue(t *testing.T) {
	outcome, err := RunCompatibilityCheck(context.Background(), CompatibilityParams{
		Commit: CommitResult{
			Message:              "feat(cli): remove legacy mode",
			FunctionalityRemoved: []string{"Removed the legacy CLI mode"},
		},
		Mode: IgnoreCheckModePrompt,
		Decider: func(context.Context, CommitResult) (CompatibilityDecision, error) {
			return CompatibilityDecisionContinue, nil
		},
	})
	require.NoError(t, err)
	assert.False(t, outcome.Cancelled)
}

func TestRunCompatibilityCheckCancel(t *testing.T) {
	outcome, err := RunCompatibilityCheck(context.Background(), CompatibilityParams{
		Commit: CommitResult{
			Message:             "feat(api): remove v1 endpoint",
			CompatibilityIssues: []string{"Clients must migrate to /v2 before deploying this change"},
		},
		Mode: IgnoreCheckModePrompt,
		Decider: func(context.Context, CommitResult) (CompatibilityDecision, error) {
			return CompatibilityDecisionCancel, nil
		},
	})
	require.NoError(t, err)
	assert.True(t, outcome.Cancelled)
}

func TestRunCompatibilityCheckFailMode(t *testing.T) {
	_, err := RunCompatibilityCheck(context.Background(), CompatibilityParams{
		Commit: CommitResult{
			Label:                "api",
			Message:              "feat(api): remove v1 endpoint",
			FunctionalityRemoved: []string{"Removed the v1 REST endpoint"},
			CompatibilityIssues:  []string{"Clients must migrate to /v2"},
		},
		Mode: IgnoreCheckModeFail,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `commit group "api"`)
	assert.Contains(t, err.Error(), "functionality removed:")
	assert.Contains(t, err.Error(), "compatibility issues:")
}

func TestRunCompatibilityCheckSkipMode(t *testing.T) {
	outcome, err := RunCompatibilityCheck(context.Background(), CompatibilityParams{
		Commit: CommitResult{
			Message:              "feat(cli): remove legacy mode",
			FunctionalityRemoved: []string{"Removed the legacy CLI mode"},
		},
		Mode: IgnoreCheckModeSkip,
	})
	require.NoError(t, err)
	assert.False(t, outcome.Cancelled)
}

func TestRunCompatibilityCheckNonTTYEscalatesToFail(t *testing.T) {
	previous := stdinIsTerminal
	stdinIsTerminal = func() bool { return false }
	defer func() {
		stdinIsTerminal = previous
	}()

	_, err := RunCompatibilityCheck(context.Background(), CompatibilityParams{
		Commit: CommitResult{
			Message:             "feat(api): remove v1 endpoint",
			CompatibilityIssues: []string{"Clients must migrate to /v2"},
		},
		Mode: IgnoreCheckModePrompt,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "compatibility warnings")
}

func TestRunSingleCommitCompatibilityCancel(t *testing.T) {
	repo := initCommitRepo(t)
	writeFileInDir(t, repo, "api.txt", "next\n")
	gitRun(t, repo, "add", "api.txt")

	restore := stubCommitAI(t,
		func(ctx context.Context, analysis models.CommitAnalysis, agent clickyai.Agent, opts git.AnalyzeOptions) (models.CommitAnalysis, error) {
			return models.CommitAnalysis{
				Commit: models.Commit{
					CommitType: models.CommitType("feat"),
					Scope:      models.ScopeType("api"),
					Subject:    "remove legacy endpoint",
				},
			}, nil
		},
		func(ctx context.Context, analysis models.CommitAnalysis, agent clickyai.Agent, opts git.AnalyzeOptions) (git.CommitPromptAnalysis, error) {
			return git.CommitPromptAnalysis{
				Commit:               analysis,
				FunctionalityRemoved: []string{"Removed the legacy endpoint"},
			}, nil
		},
	)
	defer restore()

	previousTTY := stdinIsTerminal
	stdinIsTerminal = func() bool { return true }
	defer func() {
		stdinIsTerminal = previousTTY
	}()

	previousDecider := interactiveCompatibilityDecider
	interactiveCompatibilityDecider = func(context.Context, CommitResult) (CompatibilityDecision, error) {
		return CompatibilityDecisionCancel, nil
	}
	defer func() {
		interactiveCompatibilityDecider = previousDecider
	}()

	_, err := Run(context.Background(), Options{
		WorkDir:    repo,
		CompatMode: IgnoreCheckModePrompt,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCompatibilityCancelled)
	assert.Equal(t, "1", strings.TrimSpace(gitOutput(t, repo, "rev-list", "--count", "HEAD")))
}

func TestRunCommitAllCompatibilityCancelLeavesHistoryUntouched(t *testing.T) {
	repo := initCommitRepo(t)
	writeFileInDir(t, repo, "alpha/a.txt", "one\n")
	writeFileInDir(t, repo, "beta/b.txt", "two\n")
	gitRun(t, repo, "add", "alpha/a.txt", "beta/b.txt")

	restore := stubCommitAI(t,
		func(ctx context.Context, analysis models.CommitAnalysis, agent clickyai.Agent, opts git.AnalyzeOptions) (models.CommitAnalysis, error) {
			return models.CommitAnalysis{
				Commit: models.Commit{
					CommitType: models.CommitType("feat"),
					Subject:    "split changes safely",
				},
			}, nil
		},
		func(ctx context.Context, analysis models.CommitAnalysis, agent clickyai.Agent, opts git.AnalyzeOptions) (git.CommitPromptAnalysis, error) {
			return git.CommitPromptAnalysis{
				Commit:              analysis,
				CompatibilityIssues: []string{"Downstream automation must handle the new split"},
			}, nil
		},
	)
	defer restore()

	previousTTY := stdinIsTerminal
	stdinIsTerminal = func() bool { return true }
	defer func() {
		stdinIsTerminal = previousTTY
	}()

	previousDecider := interactiveCompatibilityDecider
	interactiveCompatibilityDecider = func(context.Context, CommitResult) (CompatibilityDecision, error) {
		return CompatibilityDecisionCancel, nil
	}
	defer func() {
		interactiveCompatibilityDecider = previousDecider
	}()

	_, err := Run(context.Background(), Options{
		WorkDir:    repo,
		CommitAll:  true,
		CompatMode: IgnoreCheckModePrompt,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCompatibilityCancelled)
	assert.Equal(t, "1", strings.TrimSpace(gitOutput(t, repo, "rev-list", "--count", "HEAD")))
}

func TestRunSingleCommitDryRunShowsWarningsForExplicitMessage(t *testing.T) {
	repo := initCommitRepo(t)
	writeFileInDir(t, repo, "cli.txt", "next\n")
	gitRun(t, repo, "add", "cli.txt")

	restore := stubCommitAI(t, nil, func(ctx context.Context, analysis models.CommitAnalysis, agent clickyai.Agent, opts git.AnalyzeOptions) (git.CommitPromptAnalysis, error) {
		return git.CommitPromptAnalysis{
			Commit: analysis,
			FunctionalityRemoved: []string{
				"Removed the legacy CLI mode",
			},
			CompatibilityIssues: []string{
				"Shell scripts must switch to the new subcommand",
			},
		}, nil
	})
	defer restore()

	previousTTY := stdinIsTerminal
	stdinIsTerminal = func() bool { return false }
	defer func() {
		stdinIsTerminal = previousTTY
	}()

	previousDecider := interactiveCompatibilityDecider
	interactiveCompatibilityDecider = func(context.Context, CommitResult) (CompatibilityDecision, error) {
		t.Fatal("dry-run should not prompt")
		return CompatibilityDecisionCancel, nil
	}
	defer func() {
		interactiveCompatibilityDecider = previousDecider
	}()

	var buf bytes.Buffer
	previousOutput := dryRunOutput
	dryRunOutput = &buf
	defer func() {
		dryRunOutput = previousOutput
	}()

	result, err := Run(context.Background(), Options{
		WorkDir:    repo,
		Message:    "chore: keep explicit message",
		DryRun:     true,
		CompatMode: IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	require.Len(t, result.Commits, 1)
	assert.Equal(t, "chore: keep explicit message", result.Commits[0].Message)
	assert.Equal(t, []string{"Removed the legacy CLI mode"}, result.Commits[0].FunctionalityRemoved)
	assert.Equal(t, []string{"Shell scripts must switch to the new subcommand"}, result.Commits[0].CompatibilityIssues)

	clean := stripANSIForTest(buf.String())
	assert.Contains(t, clean, "Functionality removed:")
	assert.Contains(t, clean, "Compatibility issues:")
	assert.Contains(t, clean, "Removed the legacy CLI mode")
	assert.Contains(t, clean, "Shell scripts must switch to the new subcommand")
}

func TestGenerateCommitAnalysisWithAgent_ExplicitMessageAndCompatSkipSkipsAI(t *testing.T) {
	restore := stubCommitAI(t,
		func(context.Context, models.CommitAnalysis, clickyai.Agent, git.AnalyzeOptions) (models.CommitAnalysis, error) {
			t.Fatal("message AI should not run")
			return models.CommitAnalysis{}, nil
		},
		func(context.Context, models.CommitAnalysis, clickyai.Agent, git.AnalyzeOptions) (git.CommitPromptAnalysis, error) {
			t.Fatal("compatibility AI should not run")
			return git.CommitPromptAnalysis{}, nil
		},
	)
	defer restore()

	got, err := generateCommitAnalysisWithAgent(context.Background(), "diff", "chore: keep explicit message", IgnoreCheckModeSkip, nil)
	require.NoError(t, err)
	assert.Equal(t, "chore: keep explicit message", got.Message)
	assert.Empty(t, got.FunctionalityRemoved)
	assert.Empty(t, got.CompatibilityIssues)
}

func TestGenerateCommitAnalysisWithAgent_CompatSkipStillGeneratesMessage(t *testing.T) {
	restore := stubCommitAI(t,
		func(context.Context, models.CommitAnalysis, clickyai.Agent, git.AnalyzeOptions) (models.CommitAnalysis, error) {
			return models.CommitAnalysis{
				Commit: models.Commit{
					CommitType: models.CommitType("feat"),
					Scope:      models.ScopeType("api"),
					Subject:    "add stable endpoint",
				},
			}, nil
		},
		func(context.Context, models.CommitAnalysis, clickyai.Agent, git.AnalyzeOptions) (git.CommitPromptAnalysis, error) {
			t.Fatal("compatibility AI should not run when compat is skipped")
			return git.CommitPromptAnalysis{}, nil
		},
	)
	defer restore()

	got, err := generateCommitAnalysisWithAgent(context.Background(), "diff", "", IgnoreCheckModeSkip, nil)
	require.NoError(t, err)
	assert.Equal(t, "feat(api): add stable endpoint", got.Message)
	assert.Empty(t, got.FunctionalityRemoved)
	assert.Empty(t, got.CompatibilityIssues)
}

func TestRunSingleCommitCompatibilityAIFailureFailMode(t *testing.T) {
	repo := initCommitRepo(t)
	writeFileInDir(t, repo, "api.txt", "next\n")
	gitRun(t, repo, "add", "api.txt")

	restore := stubCommitAI(t,
		func(context.Context, models.CommitAnalysis, clickyai.Agent, git.AnalyzeOptions) (models.CommitAnalysis, error) {
			return models.CommitAnalysis{
				Commit: models.Commit{
					CommitType: models.CommitType("feat"),
					Subject:    "add endpoint",
				},
			}, nil
		},
		func(context.Context, models.CommitAnalysis, clickyai.Agent, git.AnalyzeOptions) (git.CommitPromptAnalysis, error) {
			return git.CommitPromptAnalysis{}, assert.AnError
		},
	)
	defer restore()

	_, err := Run(context.Background(), Options{
		WorkDir:    repo,
		CompatMode: IgnoreCheckModeFail,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "compatibility warnings")
	assert.Contains(t, err.Error(), "AI compatibility analysis failed")
	assert.Equal(t, "1", strings.TrimSpace(gitOutput(t, repo, "rev-list", "--count", "HEAD")))
}

func TestRunSingleCommitDryRunCompatibilityAIFailureShowsWarning(t *testing.T) {
	repo := initCommitRepo(t)
	writeFileInDir(t, repo, "api.txt", "next\n")
	gitRun(t, repo, "add", "api.txt")

	restore := stubCommitAI(t,
		func(context.Context, models.CommitAnalysis, clickyai.Agent, git.AnalyzeOptions) (models.CommitAnalysis, error) {
			return models.CommitAnalysis{
				Commit: models.Commit{
					CommitType: models.CommitType("feat"),
					Subject:    "add endpoint",
				},
			}, nil
		},
		func(context.Context, models.CommitAnalysis, clickyai.Agent, git.AnalyzeOptions) (git.CommitPromptAnalysis, error) {
			return git.CommitPromptAnalysis{}, assert.AnError
		},
	)
	defer restore()

	previousTTY := stdinIsTerminal
	stdinIsTerminal = func() bool { return false }
	defer func() {
		stdinIsTerminal = previousTTY
	}()

	previousDecider := interactiveCompatibilityDecider
	interactiveCompatibilityDecider = func(context.Context, CommitResult) (CompatibilityDecision, error) {
		t.Fatal("dry-run should not prompt on compatibility AI failures")
		return CompatibilityDecisionCancel, nil
	}
	defer func() {
		interactiveCompatibilityDecider = previousDecider
	}()

	var buf bytes.Buffer
	previousOutput := dryRunOutput
	dryRunOutput = &buf
	defer func() {
		dryRunOutput = previousOutput
	}()

	result, err := Run(context.Background(), Options{
		WorkDir:    repo,
		DryRun:     true,
		CompatMode: IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	require.Len(t, result.Commits, 1)
	assert.Equal(t, []string{"AI compatibility analysis failed: assert.AnError general error for testing"}, result.Commits[0].CompatibilityIssues)

	clean := stripANSIForTest(buf.String())
	assert.Contains(t, clean, "Compatibility issues:")
	assert.Contains(t, clean, "AI compatibility analysis failed")
}

func stubCommitAI(
	t *testing.T,
	messageFn func(context.Context, models.CommitAnalysis, clickyai.Agent, git.AnalyzeOptions) (models.CommitAnalysis, error),
	compatibilityFn func(context.Context, models.CommitAnalysis, clickyai.Agent, git.AnalyzeOptions) (git.CommitPromptAnalysis, error),
) func() {
	t.Helper()

	previousAgent := newAgentFunc
	previousMessage := analyzeCommitMessageWithAIFunc
	previousCompatibility := analyzeCompatibilityPromptsWithAIFunc

	newAgentFunc = func(cfg clickyai.AgentConfig) (clickyai.Agent, error) {
		return nil, nil
	}
	if messageFn == nil {
		messageFn = func(_ context.Context, analysis models.CommitAnalysis, _ clickyai.Agent, _ git.AnalyzeOptions) (models.CommitAnalysis, error) {
			return analysis, nil
		}
	}
	if compatibilityFn == nil {
		compatibilityFn = func(_ context.Context, analysis models.CommitAnalysis, _ clickyai.Agent, _ git.AnalyzeOptions) (git.CommitPromptAnalysis, error) {
			return git.CommitPromptAnalysis{Commit: analysis}, nil
		}
	}

	analyzeCommitMessageWithAIFunc = messageFn
	analyzeCompatibilityPromptsWithAIFunc = compatibilityFn

	return func() {
		newAgentFunc = previousAgent
		analyzeCommitMessageWithAIFunc = previousMessage
		analyzeCompatibilityPromptsWithAIFunc = previousCompatibility
	}
}
