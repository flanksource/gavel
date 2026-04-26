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

func TestEvaluateFileSizeViolations(t *testing.T) {
	t.Run("no files returns nothing", func(t *testing.T) {
		got, err := EvaluateFileSizeViolations(nil, nil, nil)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("non-executable at exactly 1 MB passes", func(t *testing.T) {
		got, err := EvaluateFileSizeViolations(
			[]stagedFileStat{{Path: "data.bin", Size: maxFileBytes}},
			nil, nil,
		)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("non-executable at 1 MB + 1 fails as oversize file", func(t *testing.T) {
		got, err := EvaluateFileSizeViolations(
			[]stagedFileStat{{Path: "data.bin", Size: maxFileBytes + 1}},
			nil, nil,
		)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "data.bin", got[0].File)
		assert.Equal(t, ReasonOversizeFile, got[0].Reason)
		assert.Equal(t, maxFileBytes, got[0].Limit)
		assert.False(t, got[0].IsExecutable)
	})

	t.Run("executable at exactly 10 KB passes", func(t *testing.T) {
		got, err := EvaluateFileSizeViolations(
			[]stagedFileStat{{Path: "run.sh", Size: maxExecutableBytes, IsExecutable: true}},
			nil, nil,
		)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("executable at 10 KB + 1 fails as oversize executable", func(t *testing.T) {
		got, err := EvaluateFileSizeViolations(
			[]stagedFileStat{{Path: "run.sh", Size: maxExecutableBytes + 1, IsExecutable: true}},
			nil, nil,
		)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, ReasonOversizeExecutable, got[0].Reason)
		assert.Equal(t, maxExecutableBytes, got[0].Limit)
		assert.True(t, got[0].IsExecutable)
	})

	t.Run("executable over file limit classified as oversize file", func(t *testing.T) {
		got, err := EvaluateFileSizeViolations(
			[]stagedFileStat{{Path: "big.sh", Size: 2 * maxFileBytes, IsExecutable: true}},
			nil, nil,
		)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, ReasonOversizeFile, got[0].Reason, "file limit wins over exec limit")
	})

	t.Run("non-executable small file passes", func(t *testing.T) {
		got, err := EvaluateFileSizeViolations(
			[]stagedFileStat{{Path: "notes.txt", Size: 15 * 1024}},
			nil, nil,
		)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("symlink over limit is skipped", func(t *testing.T) {
		got, err := EvaluateFileSizeViolations(
			[]stagedFileStat{{Path: "link", Size: 2 * maxFileBytes, IsSymlink: true}},
			nil, nil,
		)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("submodule over limit is skipped", func(t *testing.T) {
		got, err := EvaluateFileSizeViolations(
			[]stagedFileStat{{Path: "vendor", Size: 2 * maxFileBytes, IsSubmodule: true}},
			nil, nil,
		)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("allow pattern exempts match", func(t *testing.T) {
		got, err := EvaluateFileSizeViolations(
			[]stagedFileStat{{Path: "data.bin", Size: 2 * maxFileBytes}},
			nil,
			[]string{"data.bin"},
		)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("allow globstar exempts nested match", func(t *testing.T) {
		got, err := EvaluateFileSizeViolations(
			[]stagedFileStat{
				{Path: "assets/big.png", Size: 2 * maxFileBytes},
				{Path: "src/big.bin", Size: 2 * maxFileBytes},
			},
			nil,
			[]string{"assets/**"},
		)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "src/big.bin", got[0].File)
	})

	t.Run("whitespace-only allow pattern returns loud error", func(t *testing.T) {
		_, err := EvaluateFileSizeViolations(
			[]stagedFileStat{{Path: "data.bin", Size: 2 * maxFileBytes}},
			nil,
			[]string{"   "},
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "commit.allow")
	})

	t.Run("head already had equal-sized oversize file suppresses violation", func(t *testing.T) {
		got, err := EvaluateFileSizeViolations(
			[]stagedFileStat{{Path: "big.bin", Size: 2 * maxFileBytes}},
			[]stagedFileStat{{Path: "big.bin", Size: 2 * maxFileBytes}},
			nil,
		)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("head had same file but smaller flags the growth", func(t *testing.T) {
		got, err := EvaluateFileSizeViolations(
			[]stagedFileStat{{Path: "big.bin", Size: 3 * maxFileBytes}},
			[]stagedFileStat{{Path: "big.bin", Size: 2 * maxFileBytes}},
			nil,
		)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "big.bin", got[0].File)
	})

	t.Run("head had non-executable at same size but staged is now executable", func(t *testing.T) {
		// Above the exec threshold but below the file threshold; exec bit flipped.
		const size = maxExecutableBytes + 2048
		got, err := EvaluateFileSizeViolations(
			[]stagedFileStat{{Path: "tool", Size: size, IsExecutable: true}},
			[]stagedFileStat{{Path: "tool", Size: size, IsExecutable: false}},
			nil,
		)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, ReasonOversizeExecutable, got[0].Reason)
	})

	t.Run("new file with no head entry is flagged", func(t *testing.T) {
		got, err := EvaluateFileSizeViolations(
			[]stagedFileStat{{Path: "new.bin", Size: 2 * maxFileBytes}},
			nil,
			nil,
		)
		require.NoError(t, err)
		require.Len(t, got, 1)
	})
}

func TestHumanBytes(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1 KB"},
		{1536, "1.5 KB"},
		{maxExecutableBytes, "10 KB"},
		{maxFileBytes, "1 MB"},
		{2*maxFileBytes + 307*1024, "2.2 MB"},
		{3 * 1024 * 1024 * 1024, "3 GB"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, humanBytes(tc.in), "humanBytes(%d)", tc.in)
	}
}

func TestParseLsFilesRecord(t *testing.T) {
	e, err := parseLsFilesRecord("100755 abcd1234 0\tscripts/run.sh")
	require.NoError(t, err)
	assert.Equal(t, "scripts/run.sh", e.Path)
	assert.Equal(t, "abcd1234", e.sha)
	assert.True(t, e.IsExecutable)
	assert.False(t, e.IsSymlink)
	assert.False(t, e.IsSubmodule)

	e, err = parseLsFilesRecord("120000 beef 0\tlink")
	require.NoError(t, err)
	assert.True(t, e.IsSymlink)

	_, err = parseLsFilesRecord("bogus")
	require.Error(t, err)
}

// --- Integration tests ---

// writeLargeFile writes a file of the given size (bytes) under dir/name, with
// the executable bit set when exec is true. Content is deterministic zero
// bytes — the test only cares about size.
func writeLargeFile(t *testing.T, dir, name string, size int64, exec bool) {
	t.Helper()
	full := filepath.Join(dir, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	f, err := os.Create(full)
	require.NoError(t, err)
	defer f.Close()
	if size > 0 {
		_, err = f.Write(make([]byte, size))
		require.NoError(t, err)
	}
	mode := os.FileMode(0o644)
	if exec {
		mode = 0o755
	}
	require.NoError(t, os.Chmod(full, mode))
}

func staticFileSizeDecider(d FileSizeDecision) FileSizeDecider {
	return func(context.Context, FileSizeViolation) (FileSizeDecision, error) { return d, nil }
}

func TestRunFileSizeCheck_UnstageFile(t *testing.T) {
	repo := initCommitRepo(t)
	writeLargeFile(t, repo, "big.bin", 2*maxFileBytes, false)
	gitRun(t, repo, "add", "big.bin")

	outcome, err := RunFileSizeCheck(context.Background(), FileSizeParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"big.bin"},
		Decider:     staticFileSizeDecider(FileSizeDecisionGitIgnoreFile),
		SaveDir:     repo,
		Mode:        CheckModePrompt,
	})
	require.NoError(t, err)
	assert.False(t, outcome.Cancelled)
	assert.Equal(t, []string{"big.bin"}, outcome.Unstaged)
	assert.Equal(t, []string{"big.bin"}, outcome.GitIgnored)

	body := readFile(t, filepath.Join(repo, ".gitignore"))
	assert.Contains(t, body, "big.bin")

	staged := gitOutput(t, repo, "diff", "--cached", "--name-only")
	assert.NotContains(t, staged, "big.bin")
}

func TestRunFileSizeCheck_UnstageFolder(t *testing.T) {
	repo := initCommitRepo(t)
	writeLargeFile(t, repo, "big/a.bin", 2*maxFileBytes, false)
	writeLargeFile(t, repo, "big/b.bin", 2*maxFileBytes, false)
	gitRun(t, repo, "add", "big/a.bin", "big/b.bin")

	outcome, err := RunFileSizeCheck(context.Background(), FileSizeParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"big/a.bin", "big/b.bin"},
		Decider:     staticFileSizeDecider(FileSizeDecisionGitIgnoreFolder),
		SaveDir:     repo,
		Mode:        CheckModePrompt,
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"big/a.bin", "big/b.bin"}, outcome.Unstaged)
	assert.Equal(t, []string{"big/"}, outcome.GitIgnored)

	body := readFile(t, filepath.Join(repo, ".gitignore"))
	assert.Contains(t, body, "big/")

	staged := gitOutput(t, repo, "diff", "--cached", "--name-only")
	assert.NotContains(t, staged, "big/")
}

func TestRunFileSizeCheck_Allow(t *testing.T) {
	repo := initCommitRepo(t)
	writeLargeFile(t, repo, "fixture.bin", 2*maxFileBytes, false)
	gitRun(t, repo, "add", "fixture.bin")

	outcome, err := RunFileSizeCheck(context.Background(), FileSizeParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"fixture.bin"},
		Decider:     staticFileSizeDecider(FileSizeDecisionAllow),
		SaveDir:     repo,
		Mode:        CheckModePrompt,
	})
	require.NoError(t, err)
	assert.Empty(t, outcome.Unstaged)
	assert.Equal(t, []string{"fixture.bin"}, outcome.Allowed)
	assert.Equal(t, []string{"!fixture.bin"}, outcome.GitIgnored,
		"Allow must also write a !path negation into .gitignore as the explicit override")

	cfg, err := verify.LoadSingleGavelConfig(filepath.Join(repo, ".gavel.yaml"))
	require.NoError(t, err)
	assert.Equal(t, []string{"fixture.bin"}, cfg.Commit.Allow)

	gitignore := readFile(t, filepath.Join(repo, ".gitignore"))
	assert.Equal(t, "!fixture.bin\n", gitignore)

	staged := gitOutput(t, repo, "diff", "--cached", "--name-only")
	assert.Contains(t, staged, "fixture.bin")
}

func TestRunFileSizeCheck_AllowFolder(t *testing.T) {
	repo := initCommitRepo(t)
	require.NoError(t, os.MkdirAll(filepath.Join(repo, "assets"), 0o755))
	writeLargeFile(t, repo, "assets/big1.bin", 2*maxFileBytes, false)
	writeLargeFile(t, repo, "assets/big2.bin", 2*maxFileBytes, false)
	gitRun(t, repo, "add", "assets/big1.bin", "assets/big2.bin")

	outcome, err := RunFileSizeCheck(context.Background(), FileSizeParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"assets/big1.bin", "assets/big2.bin"},
		Decider:     staticFileSizeDecider(FileSizeDecisionAllowFolder),
		SaveDir:     repo,
		Mode:        CheckModePrompt,
	})
	require.NoError(t, err)
	assert.Empty(t, outcome.Unstaged)
	assert.Equal(t, []string{"assets/**"}, outcome.Allowed,
		"folder allow records the recursive glob (.gavel.yaml entry) so a single decision covers every nested file")
	assert.Equal(t, []string{"!assets/"}, outcome.GitIgnored)

	cfg, err := verify.LoadSingleGavelConfig(filepath.Join(repo, ".gavel.yaml"))
	require.NoError(t, err)
	assert.Equal(t, []string{"assets/**"}, cfg.Commit.Allow,
		"folder allow records the recursive glob in .gavel.yaml so EvaluateFileSizeViolations sees every nested file")

	gitignore := readFile(t, filepath.Join(repo, ".gitignore"))
	assert.Equal(t, "!assets/\n", gitignore)
}

func TestAppendGitIgnoreAllow_DedupesAndAddsBangPrefix(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.log\n!already.bin\n"), 0o644))

	written, err := appendGitIgnoreAllow(dir, []string{"new.bin", "!already.bin", "assets/"})
	require.NoError(t, err)
	assert.Equal(t, []string{"!new.bin", "!assets/"}, written,
		"already-present negation must dedupe; raw entries must get the ! prefix")

	body, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	require.NoError(t, err)
	assert.Equal(t, "*.log\n!already.bin\n!new.bin\n!assets/\n", string(body))
}

func TestRunFileSizeCheck_Cancel(t *testing.T) {
	repo := initCommitRepo(t)
	writeLargeFile(t, repo, "big.bin", 2*maxFileBytes, false)
	gitRun(t, repo, "add", "big.bin")

	outcome, err := RunFileSizeCheck(context.Background(), FileSizeParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"big.bin"},
		Decider:     staticFileSizeDecider(FileSizeDecisionCancel),
		SaveDir:     repo,
		Mode:        CheckModePrompt,
	})
	require.NoError(t, err)
	assert.True(t, outcome.Cancelled)

	_, err = os.Stat(filepath.Join(repo, ".gitignore"))
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(repo, ".gavel.yaml"))
	assert.True(t, os.IsNotExist(err))

	staged := gitOutput(t, repo, "diff", "--cached", "--name-only")
	assert.Contains(t, staged, "big.bin")
}

func TestRunFileSizeCheck_ExecutableSmall(t *testing.T) {
	repo := initCommitRepo(t)
	writeLargeFile(t, repo, "run.sh", 15*1024, true)
	gitRun(t, repo, "add", "run.sh")

	outcome, err := RunFileSizeCheck(context.Background(), FileSizeParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"run.sh"},
		Decider:     staticFileSizeDecider(FileSizeDecisionGitIgnoreFile),
		SaveDir:     repo,
		Mode:        CheckModePrompt,
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"run.sh"}, outcome.Unstaged)
}

func TestRunFileSizeCheck_NonExecutableSmall(t *testing.T) {
	repo := initCommitRepo(t)
	writeLargeFile(t, repo, "notes.txt", 15*1024, false)
	gitRun(t, repo, "add", "notes.txt")

	outcome, err := RunFileSizeCheck(context.Background(), FileSizeParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"notes.txt"},
		Decider:     staticFileSizeDecider(FileSizeDecisionCancel), // would cancel if prompted
		SaveDir:     repo,
		Mode:        CheckModePrompt,
	})
	require.NoError(t, err)
	assert.False(t, outcome.Cancelled, "small non-executable should not trigger the prompt")
	assert.Empty(t, outcome.Unstaged)
}

func TestRunFileSizeCheck_StagedBlobDiffersFromWorkingTree(t *testing.T) {
	repo := initCommitRepo(t)
	writeLargeFile(t, repo, "big.bin", 2*maxFileBytes, false)
	gitRun(t, repo, "add", "big.bin")
	// Truncate the working-tree file AFTER staging; staged blob is still 2 MB.
	require.NoError(t, os.WriteFile(filepath.Join(repo, "big.bin"), nil, 0o644))

	outcome, err := RunFileSizeCheck(context.Background(), FileSizeParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"big.bin"},
		Decider:     staticFileSizeDecider(FileSizeDecisionGitIgnoreFile),
		SaveDir:     repo,
		Mode:        CheckModePrompt,
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"big.bin"}, outcome.Unstaged,
		"check must read staged blob size, not working-tree stat")
}

func TestRunFileSizeCheck_PreExistingLargeFileInHeadIsSuppressed(t *testing.T) {
	repo := initCommitRepo(t)
	writeLargeFile(t, repo, "big.bin", 2*maxFileBytes, false)
	gitRun(t, repo, "add", "big.bin")
	gitRun(t, repo, "commit", "-m", "seed oversized file")

	// Touch an unrelated file so we have a non-empty commit, and re-stage the
	// same big.bin (same size) — HEAD already has this violation.
	writeFile(t, repo, "notes.txt", "small\n")
	gitRun(t, repo, "add", "notes.txt")
	// Rewrite big.bin with identical size + content so it stays staged-identical.
	writeLargeFile(t, repo, "big.bin", 2*maxFileBytes, false)
	gitRun(t, repo, "add", "big.bin")

	staged := strings.Fields(gitOutput(t, repo, "diff", "--cached", "--name-only"))

	outcome, err := RunFileSizeCheck(context.Background(), FileSizeParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: staged,
		Decider: func(_ context.Context, v FileSizeViolation) (FileSizeDecision, error) {
			t.Fatalf("decider should not be called; got violation for %q", v.File)
			return FileSizeDecisionCancel, nil
		},
		SaveDir: repo,
		Mode:    CheckModePrompt,
	})
	require.NoError(t, err)
	assert.False(t, outcome.Cancelled)
	assert.Empty(t, outcome.Unstaged)
}

func TestRunFileSizeCheck_GrowthOverHeadIsFlagged(t *testing.T) {
	repo := initCommitRepo(t)
	writeLargeFile(t, repo, "data.bin", maxFileBytes/2, false) // well under limit
	gitRun(t, repo, "add", "data.bin")
	gitRun(t, repo, "commit", "-m", "seed small data.bin")

	writeLargeFile(t, repo, "data.bin", 2*maxFileBytes, false) // grew past limit
	gitRun(t, repo, "add", "data.bin")

	outcome, err := RunFileSizeCheck(context.Background(), FileSizeParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"data.bin"},
		Decider:     staticFileSizeDecider(FileSizeDecisionGitIgnoreFile),
		SaveDir:     repo,
		Mode:        CheckModePrompt,
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"data.bin"}, outcome.Unstaged)
}

func TestRunFileSizeCheck_SkipMode(t *testing.T) {
	repo := initCommitRepo(t)
	writeLargeFile(t, repo, "big.bin", 2*maxFileBytes, false)
	gitRun(t, repo, "add", "big.bin")

	outcome, err := RunFileSizeCheck(context.Background(), FileSizeParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"big.bin"},
		Decider: func(context.Context, FileSizeViolation) (FileSizeDecision, error) {
			t.Fatalf("skip mode must not call decider")
			return FileSizeDecisionCancel, nil
		},
		SaveDir: repo,
		Mode:    CheckModeSkip,
	})
	require.NoError(t, err)
	assert.Empty(t, outcome.Unstaged)
	assert.Empty(t, outcome.Allowed)
}

func TestRunFileSizeCheck_FailMode(t *testing.T) {
	repo := initCommitRepo(t)
	writeLargeFile(t, repo, "big.bin", 2*maxFileBytes, false)
	gitRun(t, repo, "add", "big.bin")

	_, err := RunFileSizeCheck(context.Background(), FileSizeParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"big.bin"},
		Decider: func(context.Context, FileSizeViolation) (FileSizeDecision, error) {
			t.Fatalf("fail mode must not call decider")
			return FileSizeDecisionCancel, nil
		},
		SaveDir: repo,
		Mode:    CheckModeFail,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "big.bin")
	assert.Contains(t, err.Error(), "1 MB")
}

func TestRunFileSizeCheck_NonTTYEscalatesToFail(t *testing.T) {
	prev := stdinIsTerminal
	stdinIsTerminal = func() bool { return false }
	defer func() { stdinIsTerminal = prev }()

	repo := initCommitRepo(t)
	writeLargeFile(t, repo, "big.bin", 2*maxFileBytes, false)
	gitRun(t, repo, "add", "big.bin")

	_, err := RunFileSizeCheck(context.Background(), FileSizeParams{
		WorkDir:     repo,
		GitRoot:     repo,
		StagedFiles: []string{"big.bin"},
		SaveDir:     repo,
		Mode:        CheckModePrompt,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "big.bin")
}

func TestReadStagedStats_ExecutableBitPreserved(t *testing.T) {
	repo := initCommitRepo(t)
	writeLargeFile(t, repo, "run.sh", 12*1024, true)
	gitRun(t, repo, "add", "run.sh")

	stats, err := readStagedStats(repo, []string{"run.sh"})
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, "run.sh", stats[0].Path)
	assert.True(t, stats[0].IsExecutable)
	assert.Equal(t, int64(12*1024), stats[0].Size)
}

func TestReadHeadStats_NoHeadYet(t *testing.T) {
	dir := t.TempDir()
	gitRun(t, dir, "init")
	gitRun(t, dir, "config", "user.email", "test@example.com")
	gitRun(t, dir, "config", "user.name", "Test User")
	gitRun(t, dir, "config", "commit.gpgsign", "false")

	stats, err := readHeadStats(dir, []string{"never.bin"})
	require.NoError(t, err)
	assert.Empty(t, stats)
}
