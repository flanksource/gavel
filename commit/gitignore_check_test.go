package commit

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluateGitIgnoreMatches(t *testing.T) {
	t.Run("no patterns returns no violations", func(t *testing.T) {
		got, err := EvaluateGitIgnoreMatches([]string{"foo.log"}, nil, nil)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("glob matches and rejects non-matches", func(t *testing.T) {
		got, err := EvaluateGitIgnoreMatches(
			[]string{"foo.log", "foo.log.txt", "bar.txt"},
			[]string{"*.log"},
			nil,
		)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "foo.log", got[0].File)
		assert.Equal(t, "*.log", got[0].Pattern)
	})

	t.Run("globstar matches nested paths", func(t *testing.T) {
		got, err := EvaluateGitIgnoreMatches(
			[]string{"config/secrets.env", "secrets.env", "config/ok.yaml"},
			[]string{"**/*.env"},
			nil,
		)
		require.NoError(t, err)
		files := violationFiles(got)
		assert.ElementsMatch(t, []string{"config/secrets.env", "secrets.env"}, files)
	})

	t.Run("allow negation exempts a match", func(t *testing.T) {
		got, err := EvaluateGitIgnoreMatches(
			[]string{"foo.log", "bar.log"},
			[]string{"*.log"},
			[]string{"foo.log"},
		)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "bar.log", got[0].File)
	})

	t.Run("directory pattern matches files under that dir", func(t *testing.T) {
		got, err := EvaluateGitIgnoreMatches(
			[]string{"build/out.js", "mybuild.js"},
			[]string{"build/"},
			nil,
		)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "build/out.js", got[0].File)
	})

	t.Run("whitespace-only pattern returns loud error", func(t *testing.T) {
		_, err := EvaluateGitIgnoreMatches(
			[]string{"foo.log"},
			[]string{"*.log", "   "},
			nil,
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "commit.gitignore")
	})

	t.Run("multi-pattern match reports first matching pattern", func(t *testing.T) {
		got, err := EvaluateGitIgnoreMatches(
			[]string{"secrets.env"},
			[]string{"**/*.env", "secrets*"},
			nil,
		)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "**/*.env", got[0].Pattern)
	})
}

func violationFiles(vs []Violation) []string {
	out := make([]string, len(vs))
	for i, v := range vs {
		out[i] = v.File
	}
	return out
}

func TestRunGitIgnoreCheck_Unstage(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "debug.log", "log content\n")
	gitRun(t, repo, "add", "debug.log")

	outcome, err := RunGitIgnoreCheck(context.Background(), CheckParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"debug.log"},
		Config: verify.CommitConfig{
			GitIgnore: []string{"*.log"},
		},
		Decider: staticDecider(DecisionGitIgnore),
		SaveDir: repo,
		Mode:    IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	assert.False(t, outcome.Cancelled)
	assert.Equal(t, []string{"debug.log"}, outcome.Unstaged)
	assert.Equal(t, []string{"debug.log"}, outcome.GitIgnored)

	gitignoreBody := readFile(t, filepath.Join(repo, ".gitignore"))
	assert.Contains(t, gitignoreBody, "debug.log")

	staged := gitOutput(t, repo, "diff", "--cached", "--name-only")
	assert.NotContains(t, staged, "debug.log")
}

func TestRunGitIgnoreCheck_UnstagePattern(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "debug.log", "log content\n")
	gitRun(t, repo, "add", "debug.log")

	outcome, err := RunGitIgnoreCheck(context.Background(), CheckParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"debug.log"},
		Config: verify.CommitConfig{
			GitIgnore: []string{"*.log"},
		},
		Decider: staticDecider(DecisionGitIgnorePattern),
		SaveDir: repo,
		Mode:    IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"debug.log"}, outcome.Unstaged)
	assert.Equal(t, []string{"*.log"}, outcome.GitIgnored)

	gitignoreBody := readFile(t, filepath.Join(repo, ".gitignore"))
	assert.Contains(t, gitignoreBody, "*.log")
}

func TestRunGitIgnoreCheck_UnstageFolder(t *testing.T) {
	repo := initCommitRepo(t)
	require.NoError(t, os.MkdirAll(filepath.Join(repo, "logs"), 0o755))
	writeFile(t, repo, "logs/debug.log", "log content\n")
	gitRun(t, repo, "add", "logs/debug.log")

	outcome, err := RunGitIgnoreCheck(context.Background(), CheckParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"logs/debug.log"},
		Config: verify.CommitConfig{
			GitIgnore: []string{"*.log"},
		},
		Decider: staticDecider(DecisionGitIgnoreFolder),
		SaveDir: repo,
		Mode:    IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"logs/debug.log"}, outcome.Unstaged)
	assert.Equal(t, []string{"logs/"}, outcome.GitIgnored)

	gitignoreBody := readFile(t, filepath.Join(repo, ".gitignore"))
	assert.Contains(t, gitignoreBody, "logs/")
}

func TestRunGitIgnoreCheck_Allow(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "secrets.env", "SECRET=1\n")
	gitRun(t, repo, "add", "secrets.env")

	outcome, err := RunGitIgnoreCheck(context.Background(), CheckParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"secrets.env"},
		Config: verify.CommitConfig{
			GitIgnore: []string{"*.env"},
		},
		Decider: staticDecider(DecisionAllow),
		SaveDir: repo,
		Mode:    IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	assert.Empty(t, outcome.Unstaged)
	assert.Equal(t, []string{"secrets.env"}, outcome.Allowed)

	cfg, err := verify.LoadSingleGavelConfig(filepath.Join(repo, ".gavel.yaml"))
	require.NoError(t, err)
	assert.Equal(t, []string{"secrets.env"}, cfg.Commit.Allow)

	staged := gitOutput(t, repo, "diff", "--cached", "--name-only")
	assert.Contains(t, staged, "secrets.env")
}

func TestRunGitIgnoreCheck_Cancel(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "debug.log", "log\n")
	gitRun(t, repo, "add", "debug.log")

	outcome, err := RunGitIgnoreCheck(context.Background(), CheckParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"debug.log"},
		Config:      verify.CommitConfig{GitIgnore: []string{"*.log"}},
		Decider:     staticDecider(DecisionCancel),
		SaveDir:     repo,
		Mode:        IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	assert.True(t, outcome.Cancelled)

	_, err = os.Stat(filepath.Join(repo, ".gitignore"))
	assert.True(t, os.IsNotExist(err), "cancel must not write .gitignore")
	_, err = os.Stat(filepath.Join(repo, ".gavel.yaml"))
	assert.True(t, os.IsNotExist(err), "cancel must not write .gavel.yaml")

	staged := gitOutput(t, repo, "diff", "--cached", "--name-only")
	assert.Contains(t, staged, "debug.log", "cancel must not unstage")
}

func TestRunGitIgnoreCheck_PreExistingGitignore(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, ".gitignore", "*.tmp\n")
	gitRun(t, repo, "add", ".gitignore")
	gitRun(t, repo, "commit", "-m", "seed gitignore")

	writeFile(t, repo, "x.log", "x\n")
	gitRun(t, repo, "add", "x.log")

	_, err := RunGitIgnoreCheck(context.Background(), CheckParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"x.log"},
		Config:      verify.CommitConfig{GitIgnore: []string{"*.log"}},
		Decider:     staticDecider(DecisionGitIgnore),
		SaveDir:     repo,
		Mode:        IgnoreCheckModePrompt,
	})
	require.NoError(t, err)

	body := readFile(t, filepath.Join(repo, ".gitignore"))
	assert.Equal(t, "*.tmp\nx.log\n", body)

	// Running again with the same file already in .gitignore is a no-op on write.
	writeFile(t, repo, "y.log", "y\n")
	gitRun(t, repo, "add", "y.log")
	_, err = RunGitIgnoreCheck(context.Background(), CheckParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"y.log"},
		Config:      verify.CommitConfig{GitIgnore: []string{"*.log"}},
		Decider:     staticDecider(DecisionGitIgnore),
		SaveDir:     repo,
		Mode:        IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	body = readFile(t, filepath.Join(repo, ".gitignore"))
	assert.Equal(t, "*.tmp\nx.log\ny.log\n", body)
}

func TestRunGitIgnoreCheck_NoMatchIsNoOp(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "main.go", "package main\n")
	gitRun(t, repo, "add", "main.go")

	outcome, err := RunGitIgnoreCheck(context.Background(), CheckParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"main.go"},
		Config:      verify.CommitConfig{GitIgnore: []string{"*.log"}},
		Decider:     staticDecider(DecisionCancel), // should not be called
		SaveDir:     repo,
		Mode:        IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	assert.Equal(t, CheckOutcome{}, outcome)
	_, err = os.Stat(filepath.Join(repo, ".gitignore"))
	assert.True(t, os.IsNotExist(err))
}

func TestRunGitIgnoreCheck_FailMode(t *testing.T) {
	repo := initCommitRepo(t)

	called := false
	_, err := RunGitIgnoreCheck(context.Background(), CheckParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"a.log", "b.log"},
		Config:      verify.CommitConfig{GitIgnore: []string{"*.log"}},
		Decider: func(context.Context, Violation) (Decision, error) {
			called = true
			return DecisionCancel, nil
		},
		Mode: IgnoreCheckModeFail,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "a.log")
	assert.Contains(t, err.Error(), "b.log")
	assert.False(t, called, "decider must not run in fail mode")
}

func TestRunGitIgnoreCheck_SkipMode(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "x.log", "x\n")
	gitRun(t, repo, "add", "x.log")

	called := false
	outcome, err := RunGitIgnoreCheck(context.Background(), CheckParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"x.log"},
		Config:      verify.CommitConfig{GitIgnore: []string{"*.log"}},
		Decider: func(context.Context, Violation) (Decision, error) {
			called = true
			return DecisionCancel, nil
		},
		Mode: IgnoreCheckModeSkip,
	})
	require.NoError(t, err)
	assert.Equal(t, CheckOutcome{}, outcome)
	assert.False(t, called)

	staged := gitOutput(t, repo, "diff", "--cached", "--name-only")
	assert.Contains(t, staged, "x.log", "skip mode must not unstage")
}

func TestRunGitIgnoreCheck_NonTTYEscalatesToFail(t *testing.T) {
	prev := stdinIsTerminal
	stdinIsTerminal = func() bool { return false }
	t.Cleanup(func() { stdinIsTerminal = prev })

	repo := initCommitRepo(t)

	// No Decider injected -> would fall through to the real interactive
	// prompt -> escalates to fail because stdin isn't a TTY.
	_, err := RunGitIgnoreCheck(context.Background(), CheckParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"bad.log"},
		Config:      verify.CommitConfig{GitIgnore: []string{"*.log"}},
		Mode:        IgnoreCheckModePrompt,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad.log")
}

func TestRunGitIgnoreCheck_PreservesExistingAllowEntries(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, ".gavel.yaml", "commit:\n  allow:\n    - existing.env\n")
	writeFile(t, repo, "new.env", "x\n")
	gitRun(t, repo, "add", "new.env")

	_, err := RunGitIgnoreCheck(context.Background(), CheckParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"new.env"},
		Config:      verify.CommitConfig{GitIgnore: []string{"*.env"}},
		Decider:     staticDecider(DecisionAllow),
		SaveDir:     repo,
		Mode:        IgnoreCheckModePrompt,
	})
	require.NoError(t, err)

	cfg, err := verify.LoadSingleGavelConfig(filepath.Join(repo, ".gavel.yaml"))
	require.NoError(t, err)
	assert.Equal(t, []string{"existing.env", "new.env"}, cfg.Commit.Allow)
}

func TestAppendGitIgnore_NoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.tmp"), 0o644))

	written, err := appendGitIgnore(dir, []string{"x.log"})
	require.NoError(t, err)
	assert.Equal(t, []string{"x.log"}, written)

	body, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	require.NoError(t, err)
	assert.Equal(t, "*.tmp\nx.log\n", string(body))
}

func TestAppendGitIgnore_SkipsDuplicate(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("x.log\n"), 0o644))

	written, err := appendGitIgnore(dir, []string{"x.log"})
	require.NoError(t, err)
	assert.Empty(t, written)

	body, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	require.NoError(t, err)
	assert.Equal(t, "x.log\n", string(body))
}

func TestGitIgnoreChoices(t *testing.T) {
	t.Run("nested file offers pattern, folder, file, allow, cancel", func(t *testing.T) {
		choices := gitIgnoreChoices(Violation{File: "logs/debug.log", Pattern: "*.log"})
		require.Len(t, choices, 5)
		assert.Equal(t, DecisionGitIgnorePattern, choices[0].Decision)
		assert.Equal(t, DecisionGitIgnoreFolder, choices[1].Decision)
		assert.Equal(t, DecisionGitIgnoreFile, choices[2].Decision)
		assert.Equal(t, DecisionAllow, choices[3].Decision)
		assert.Equal(t, DecisionCancel, choices[4].Decision)
		assert.Contains(t, choices[1].Text, `"logs/"`)
	})

	t.Run("root file omits folder option", func(t *testing.T) {
		choices := gitIgnoreChoices(Violation{File: "debug.log", Pattern: "*.log"})
		require.Len(t, choices, 4)
		assert.Equal(t, DecisionGitIgnorePattern, choices[0].Decision)
		assert.Equal(t, DecisionGitIgnoreFile, choices[1].Decision)
		assert.Equal(t, DecisionAllow, choices[2].Decision)
		assert.Equal(t, DecisionCancel, choices[3].Decision)
	})
}

func staticDecider(d Decision) Decider {
	return func(context.Context, Violation) (Decision, error) { return d, nil }
}

func TestRunGitIgnoreCheck_FolderDecisionBatchesSiblings(t *testing.T) {
	repo := initCommitRepo(t)
	require.NoError(t, os.MkdirAll(filepath.Join(repo, "logs"), 0o755))
	writeFile(t, repo, "logs/a.log", "a\n")
	writeFile(t, repo, "logs/b.log", "b\n")
	writeFile(t, repo, "logs/c.log", "c\n")
	gitRun(t, repo, "add", "logs/a.log", "logs/b.log", "logs/c.log")

	var seen []Violation
	decider := func(_ context.Context, v Violation) (Decision, error) {
		seen = append(seen, v)
		return DecisionGitIgnoreFolder, nil
	}

	outcome, err := RunGitIgnoreCheck(context.Background(), CheckParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"logs/a.log", "logs/b.log", "logs/c.log"},
		Config:      verify.CommitConfig{GitIgnore: []string{"*.log"}},
		Decider:     decider,
		SaveDir:     repo,
		Mode:        IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	require.Len(t, seen, 1, "folder decision must batch all siblings in one prompt")
	assert.Equal(t, 2, seen[0].FolderSiblings, "sibling count should be reported to the decider")
	assert.ElementsMatch(t, []string{"logs/a.log", "logs/b.log", "logs/c.log"}, outcome.Unstaged)
	assert.Equal(t, []string{"logs/"}, outcome.GitIgnored)

	body := readFile(t, filepath.Join(repo, ".gitignore"))
	assert.Equal(t, "logs/\n", body)
}

func TestRunGitIgnoreCheck_PatternDecisionBatchesAcrossDirs(t *testing.T) {
	repo := initCommitRepo(t)
	require.NoError(t, os.MkdirAll(filepath.Join(repo, "logs"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(repo, "var"), 0o755))
	writeFile(t, repo, "logs/a.log", "a\n")
	writeFile(t, repo, "var/b.log", "b\n")
	writeFile(t, repo, "c.log", "c\n")
	gitRun(t, repo, "add", "logs/a.log", "var/b.log", "c.log")

	calls := 0
	decider := func(_ context.Context, v Violation) (Decision, error) {
		calls++
		return DecisionGitIgnorePattern, nil
	}

	outcome, err := RunGitIgnoreCheck(context.Background(), CheckParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"logs/a.log", "var/b.log", "c.log"},
		Config:      verify.CommitConfig{GitIgnore: []string{"*.log"}},
		Decider:     decider,
		SaveDir:     repo,
		Mode:        IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, calls, "pattern decision must batch every file matching the pattern")
	assert.ElementsMatch(t, []string{"logs/a.log", "var/b.log", "c.log"}, outcome.Unstaged)
	assert.Equal(t, []string{"*.log"}, outcome.GitIgnored)
}

func TestRunGitIgnoreCheck_MixedDecisionsRecalcAfterEach(t *testing.T) {
	repo := initCommitRepo(t)
	require.NoError(t, os.MkdirAll(filepath.Join(repo, "logs"), 0o755))
	writeFile(t, repo, "logs/a.log", "a\n")
	writeFile(t, repo, "logs/b.log", "b\n")
	writeFile(t, repo, "secrets.env", "S=1\n")
	gitRun(t, repo, "add", "logs/a.log", "logs/b.log", "secrets.env")

	var seen []Violation
	decider := func(_ context.Context, v Violation) (Decision, error) {
		seen = append(seen, v)
		if strings.HasSuffix(v.File, ".env") {
			return DecisionAllow, nil
		}
		return DecisionGitIgnoreFolder, nil
	}

	outcome, err := RunGitIgnoreCheck(context.Background(), CheckParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"logs/a.log", "logs/b.log", "secrets.env"},
		Config:      verify.CommitConfig{GitIgnore: []string{"*.log", "*.env"}},
		Decider:     decider,
		SaveDir:     repo,
		Mode:        IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	require.Len(t, seen, 2, "folder decision batches the two logs; env still prompts separately")
	assert.ElementsMatch(t, []string{"logs/a.log", "logs/b.log"}, outcome.Unstaged)
	assert.Equal(t, []string{"logs/"}, outcome.GitIgnored)
	assert.Equal(t, []string{"secrets.env"}, outcome.Allowed)
}

func TestRunGitIgnoreCheck_RootFileFolderSiblingsZero(t *testing.T) {
	repo := initCommitRepo(t)
	writeFile(t, repo, "debug.log", "x\n")
	gitRun(t, repo, "add", "debug.log")

	var seen Violation
	decider := func(_ context.Context, v Violation) (Decision, error) {
		seen = v
		return DecisionGitIgnoreFile, nil
	}
	_, err := RunGitIgnoreCheck(context.Background(), CheckParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"debug.log"},
		Config:      verify.CommitConfig{GitIgnore: []string{"*.log"}},
		Decider:     decider,
		SaveDir:     repo,
		Mode:        IgnoreCheckModePrompt,
	})
	require.NoError(t, err)
	assert.Equal(t, 0, seen.FolderSiblings)
}
