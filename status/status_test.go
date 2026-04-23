package status

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	clickyai "github.com/flanksource/clicky/ai"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/snapshots"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/flanksource/repomap"
	"github.com/flanksource/repomap/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStatusPorcelain(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []FileStatus
	}{
		{
			name: "modified unstaged",
			raw:  " M cmd/gavel/lint.go\x00",
			want: []FileStatus{{
				Path:     "cmd/gavel/lint.go",
				State:    StateUnstaged,
				WorkKind: KindModified,
			}},
		},
		{
			name: "modified staged",
			raw:  "M  cmd/gavel/lint.go\x00",
			want: []FileStatus{{
				Path:       "cmd/gavel/lint.go",
				State:      StateStaged,
				StagedKind: KindModified,
			}},
		},
		{
			name: "added staged",
			raw:  "A  cmd/gavel/lint_summary.go\x00",
			want: []FileStatus{{
				Path:       "cmd/gavel/lint_summary.go",
				State:      StateStaged,
				StagedKind: KindAdded,
			}},
		},
		{
			name: "deleted staged",
			raw:  "D  commit/ai-commit-group.md\x00",
			want: []FileStatus{{
				Path:       "commit/ai-commit-group.md",
				State:      StateStaged,
				StagedKind: KindDeleted,
			}},
		},
		{
			name: "both modified",
			raw:  "MM cmd/gavel/lint.go\x00",
			want: []FileStatus{{
				Path:       "cmd/gavel/lint.go",
				State:      StateBoth,
				StagedKind: KindModified,
				WorkKind:   KindModified,
			}},
		},
		{
			name: "untracked",
			raw:  "?? cmd/gavel/new.go\x00",
			want: []FileStatus{{
				Path:       "cmd/gavel/new.go",
				State:      StateUntracked,
				StagedKind: KindUntracked,
				WorkKind:   KindUntracked,
			}},
		},
		{
			name: "conflict UU",
			raw:  "UU commit/planner.go\x00",
			want: []FileStatus{{
				Path:       "commit/planner.go",
				State:      StateConflict,
				StagedKind: KindModified,
				WorkKind:   KindModified,
			}},
		},
		{
			name: "conflict AA",
			raw:  "AA newfile.go\x00",
			want: []FileStatus{{
				Path:       "newfile.go",
				State:      StateConflict,
				StagedKind: KindAdded,
				WorkKind:   KindAdded,
			}},
		},
		{
			name: "rename staged",
			raw:  "R  commit/new.go\x00commit/old.go\x00",
			want: []FileStatus{{
				Path:         "commit/new.go",
				PreviousPath: "commit/old.go",
				State:        StateStaged,
				StagedKind:   KindRenamed,
			}},
		},
		{
			name: "copy staged",
			raw:  "C  b.go\x00a.go\x00",
			want: []FileStatus{{
				Path:         "b.go",
				PreviousPath: "a.go",
				State:        StateStaged,
				StagedKind:   KindCopied,
			}},
		},
		{
			name: "multiple records",
			raw:  "M  a.go\x00 M b.go\x00?? c.go\x00",
			want: []FileStatus{
				{Path: "a.go", State: StateStaged, StagedKind: KindModified},
				{Path: "b.go", State: StateUnstaged, WorkKind: KindModified},
				{Path: "c.go", State: StateUntracked, StagedKind: KindUntracked, WorkKind: KindUntracked},
			},
		},
		{
			name: "empty output",
			raw:  "",
			want: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseStatusPorcelain([]byte(tc.raw))
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGatherFromRepo(t *testing.T) {
	repo := initStatusRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(repo, "staged.go"), []byte("package x\n"), 0o644))
	gitRun(t, repo, "add", "staged.go")
	require.NoError(t, os.WriteFile(filepath.Join(repo, "README.md"), []byte("# test\nchanged\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repo, "untracked.go"), []byte("package x\n"), 0o644))

	restore := stubFileMap(func(path, commit string) (*repomap.FileMap, error) {
		return &repomap.FileMap{Path: path, Language: "go"}, nil
	})
	defer restore()

	result, err := Gather(repo, Options{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Files, 3)

	byPath := map[string]FileStatus{}
	for _, f := range result.Files {
		byPath[f.Path] = f
	}
	assert.Equal(t, StateStaged, byPath["staged.go"].State)
	assert.Equal(t, StateUnstaged, byPath["README.md"].State)
	assert.Equal(t, StateUntracked, byPath["untracked.go"].State)
	assert.Equal(t, "go", byPath["staged.go"].FileMap.Language)
}

func TestGatherWithAIFileSummaries(t *testing.T) {
	repo := initStatusRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(repo, "staged.go"), []byte("package x\n\nfunc Added() {}\n"), 0o644))
	gitRun(t, repo, "add", "staged.go")

	prev := summarizeFileChangeWithAIFunc
	summarizeFileChangeWithAIFunc = func(_ context.Context, _ string, _ clickyai.Agent, file FileStatus) (string, error) {
		return "add helper function", nil
	}
	t.Cleanup(func() {
		summarizeFileChangeWithAIFunc = prev
	})

	result, err := Gather(repo, Options{
		NoRepomap: true,
		Agent:     stubStatusAgent{},
		Context:   context.Background(),
	})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)
	assert.Equal(t, "add helper function", result.Files[0].AISummary)
}

func TestSummarizeFileChangeWithAISendsDiffOnlyPrompt(t *testing.T) {
	restoreIO := stubAISummaryIO(
		func(string, string, bool) (string, error) {
			return "diff --git a/a.go b/a.go\n+added helper\n", nil
		},
		func(string, string) (string, error) {
			t.Fatal("readUntrackedStatusFileFunc should not be called for tracked files")
			return "", nil
		},
	)
	defer restoreIO()

	agent := &capturePromptAgent{}
	summary, err := summarizeFileChangeWithAI(context.Background(), "", agent, FileStatus{
		Path:  "a.go",
		State: StateStaged,
		Adds:  3,
		Dels:  1,
	})
	require.NoError(t, err)
	assert.Equal(t, "tighten handler flow", summary)
	assert.Contains(t, agent.prompt, "Staged diff:")
	assert.NotContains(t, agent.prompt, "File metadata:")
	assert.NotContains(t, agent.prompt, "path:")
	assert.NotContains(t, agent.prompt, "state:")
	assert.NotContains(t, agent.prompt, "adds:")
	assert.NotContains(t, agent.prompt, "dels:")
}

func TestStreamAISummariesPreservesFileOrdering(t *testing.T) {
	prev := summarizeFileChangeWithAIFunc
	summarizeFileChangeWithAIFunc = func(_ context.Context, _ string, _ clickyai.Agent, file FileStatus) (string, error) {
		switch file.Path {
		case "slow.go":
			time.Sleep(40 * time.Millisecond)
			return "slow summary", nil
		case "fast.go":
			return "fast summary", nil
		default:
			return "", nil
		}
	}
	t.Cleanup(func() {
		summarizeFileChangeWithAIFunc = prev
	})

	result := &Result{
		Files: []FileStatus{
			{Path: "slow.go", State: StateStaged, StagedKind: KindModified},
			{Path: "fast.go", State: StateStaged, StagedKind: KindModified},
		},
	}
	result.PrepareAISummaries()

	var running int
	for update := range StreamAISummaries(context.Background(), "", stubStatusAgent{}, result.Files, 2) {
		if update.Status == AISummaryStatusRunning {
			running++
		}
		result.ApplyAISummaryUpdate(update)
	}

	assert.Equal(t, 2, running)
	assert.Equal(t, "slow summary", result.Files[0].AISummary)
	assert.Equal(t, "fast summary", result.Files[1].AISummary)
	assert.Equal(t, AISummaryStatusDone, result.Files[0].AIStatus)
	assert.Equal(t, AISummaryStatusDone, result.Files[1].AIStatus)
}

func TestPrettyCleanRepo(t *testing.T) {
	r := &Result{Branch: "main"}
	out := stripANSI(r.Pretty().ANSI())
	assert.Contains(t, out, "main")
	assert.Contains(t, out, "clean")
}

func TestPrettyMixedStates(t *testing.T) {
	r := &Result{
		Branch: "main",
		Files: []FileStatus{
			{Path: "a.go", State: StateStaged, StagedKind: KindAdded, FileMap: &repomap.FileMap{Scopes: repomap.Scopes{"cli"}, Language: "go"}},
			{Path: "b.go", State: StateUnstaged, WorkKind: KindModified},
			{Path: "c.go", State: StateUntracked, StagedKind: KindUntracked, WorkKind: KindUntracked},
			{Path: "d.go", State: StateConflict, StagedKind: KindModified, WorkKind: KindModified},
			{Path: "e.go", PreviousPath: "old.go", State: StateStaged, StagedKind: KindRenamed},
		},
	}
	raw := r.Pretty().ANSI()
	clean := stripANSI(raw)

	assert.Contains(t, clean, "main")
	assert.Contains(t, clean, "+")
	assert.Contains(t, clean, "!")
	assert.Contains(t, clean, "?")
	assert.Contains(t, clean, "=")
	assert.Contains(t, clean, "»")
	assert.Contains(t, clean, "a.go")
	assert.Contains(t, clean, "cli")
	assert.Contains(t, clean, "go")
	assert.Contains(t, clean, "old.go → e.go")
	assert.Contains(t, clean, "⚠ conflict")
	assert.Contains(t, raw, "\x1b[")
}

func TestPrettyIncludesRepomapItems(t *testing.T) {
	r := &Result{
		Branch: "main",
		Files: []FileStatus{{
			Path: "chart.yaml", State: StateStaged, StagedKind: KindModified,
			FileMap: &repomap.FileMap{
				Scopes:         repomap.Scopes{"iac"},
				KubernetesRefs: make([]kubernetes.KubernetesRef, 3),
				Violations:     make([]repomap.Violation, 2),
			},
		}},
	}
	clean := stripANSI(r.Pretty().ANSI())
	assert.Contains(t, clean, "iac")
	assert.Contains(t, clean, "k8s:3")
	assert.Contains(t, clean, "viol:2")
}

func TestPrettyIncludesAISummary(t *testing.T) {
	r := &Result{
		Branch: "main",
		Files: []FileStatus{{
			Path:       "a.go",
			State:      StateStaged,
			StagedKind: KindModified,
			AISummary:  "tighten handler error handling",
		}},
	}
	clean := stripANSI(r.Pretty().ANSI())
	assert.Contains(t, clean, "tighten handler error handling")
}

func TestPrettyShowsAISummaryStatuses(t *testing.T) {
	r := &Result{
		Branch: "main",
		Files: []FileStatus{
			{Path: "pending.go", State: StateStaged, StagedKind: KindModified, AIStatus: AISummaryStatusPending},
			{Path: "running.go", State: StateStaged, StagedKind: KindModified, AIStatus: AISummaryStatusRunning},
			{Path: "failed.go", State: StateStaged, StagedKind: KindModified, AIStatus: AISummaryStatusFailed},
			{Path: "done.go", State: StateStaged, StagedKind: KindModified, AIStatus: AISummaryStatusDone, AISummary: "refactor handler flow"},
		},
	}

	clean := stripANSI(r.Pretty().ANSI())
	assert.Contains(t, clean, "⏳ ai")
	assert.Contains(t, clean, "⟳ ai")
	assert.Contains(t, clean, "⚠ ai summary failed")
	assert.Contains(t, clean, "refactor handler flow")
}

func TestPrettyGroupsFilesByScopeWithTestsLast(t *testing.T) {
	r := &Result{
		Branch: "main",
		Files: []FileStatus{
			{
				Path:       "z_test.go",
				State:      StateStaged,
				StagedKind: KindModified,
				FileMap:    &repomap.FileMap{Language: "go", Scopes: repomap.Scopes{repomap.ScopeTypeTest}},
			},
			{
				Path:       "docs.md",
				State:      StateStaged,
				StagedKind: KindModified,
				FileMap:    &repomap.FileMap{Language: "markdown", Scopes: repomap.Scopes{repomap.ScopeTypeDocs}},
			},
			{
				Path:       "app.go",
				State:      StateStaged,
				StagedKind: KindModified,
				FileMap:    &repomap.FileMap{Language: "go", Scopes: repomap.Scopes{repomap.ScopeTypeApp, repomap.ScopeTypeSecurity}},
			},
		},
	}

	clean := stripANSI(r.Pretty().ANSI())

	appHeader := strings.Index(clean, "\n go · app · security\n")
	docsHeader := strings.Index(clean, "\n markdown · docs\n")
	testHeader := strings.Index(clean, "\n go · test\n")
	appRow := strings.Index(clean, "app.go")
	docsRow := strings.Index(clean, "docs.md")
	testRow := strings.Index(clean, "z_test.go")

	require.NotEqual(t, -1, appHeader)
	require.NotEqual(t, -1, docsHeader)
	require.NotEqual(t, -1, testHeader)
	require.NotEqual(t, -1, appRow)
	require.NotEqual(t, -1, docsRow)
	require.NotEqual(t, -1, testRow)

	assert.Less(t, appHeader, testHeader)
	assert.Less(t, docsHeader, testHeader)
	assert.Less(t, appHeader, appRow)
	assert.Less(t, docsHeader, docsRow)
	assert.Less(t, testHeader, testRow)
	assert.Equal(t, 1, strings.Count(clean, "go · app · security"))
	assert.Equal(t, 1, strings.Count(clean, "markdown · docs"))
	assert.Equal(t, 1, strings.Count(clean, "go · test"))
}

func TestGatherWithFreshSnapshot(t *testing.T) {
	repo := initStatusRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(repo, "a.go"), []byte("package x\n"), 0o644))
	gitRun(t, repo, "add", "a.go")

	restore := stubSnapshot(func(string) (string, string, error) {
		return "deadbeef", "", nil
	}, func(string, string) (*snapshots.Pointer, error) {
		return &snapshots.Pointer{SHA: "deadbeef", Path: ".gavel/sha-deadbeef.json"}, nil
	}, func(string, *snapshots.Pointer) (*testui.Snapshot, error) {
		return &testui.Snapshot{
			Tests: []parsers.Test{{File: "a.go", Failed: true}},
			Lint: []*linters.LinterResult{{
				Violations: []models.Violation{{File: "a.go", Severity: "error"}},
			}},
		}, nil
	})
	defer restore()

	result, err := Gather(repo, Options{NoRepomap: true})
	require.NoError(t, err)
	require.Len(t, result.Files, 1)

	assert.False(t, result.ResultsStale)
	assert.Equal(t, 1, result.Files[0].TestStatus.Failed)
	assert.Equal(t, 1, result.Files[0].LintStatus.Errors)
}

func TestGatherWithStaleSnapshot(t *testing.T) {
	repo := initStatusRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(repo, "a.go"), []byte("package x\n"), 0o644))
	gitRun(t, repo, "add", "a.go")

	restore := stubSnapshot(func(string) (string, string, error) {
		return "newsha", "ab12cd34", nil
	}, func(string, string) (*snapshots.Pointer, error) {
		return &snapshots.Pointer{SHA: "oldsha", Path: ".gavel/sha-oldsha.json"}, nil
	}, func(string, *snapshots.Pointer) (*testui.Snapshot, error) {
		return &testui.Snapshot{
			Tests: []parsers.Test{{File: "a.go", Passed: true}},
		}, nil
	})
	defer restore()

	result, err := Gather(repo, Options{NoRepomap: true})
	require.NoError(t, err)
	assert.True(t, result.ResultsStale)
	assert.Equal(t, "oldsha", result.ResultsSHA)
	assert.Equal(t, 1, result.Files[0].TestStatus.Passed)
	assert.True(t, result.Files[0].ResultsStale)
}

func TestGatherNoSnapshot(t *testing.T) {
	repo := initStatusRepo(t)
	require.NoError(t, os.WriteFile(filepath.Join(repo, "a.go"), []byte("package x\n"), 0o644))
	gitRun(t, repo, "add", "a.go")

	restore := stubSnapshot(func(string) (string, string, error) {
		return "deadbeef", "", nil
	}, func(string, string) (*snapshots.Pointer, error) {
		return nil, nil
	}, func(string, *snapshots.Pointer) (*testui.Snapshot, error) {
		t.Fatal("loadSnapshotFunc should not be called when pointer is nil")
		return nil, nil
	})
	defer restore()

	result, err := Gather(repo, Options{NoRepomap: true})
	require.NoError(t, err)
	assert.False(t, result.ResultsStale)
	assert.Empty(t, result.ResultsSHA)
	assert.Equal(t, 0, result.Files[0].TestStatus.Failed)
}

func TestPrettyShowsTestLintBadgesAndStaleBanner(t *testing.T) {
	r := &Result{
		Branch:       "main",
		ResultsSHA:   "abcdef1234567890",
		ResultsStale: true,
		Files: []FileStatus{
			{Path: "a.go", State: StateUnstaged, WorkKind: KindModified,
				TestStatus: TestStatus{Failed: 2}, LintStatus: LintStatus{Errors: 3},
				ResultsStale: true},
			{Path: "b.go", State: StateStaged, StagedKind: KindModified,
				TestStatus: TestStatus{Passed: 5}, LintStatus: LintStatus{Warnings: 1}},
		},
	}
	clean := stripANSI(r.Pretty().ANSI())
	assert.Contains(t, clean, "fail:2")
	assert.Contains(t, clean, "err:3")
	assert.Contains(t, clean, "✓ 5")
	assert.Contains(t, clean, "warn:1")
	assert.Contains(t, clean, "stale results (sha abcdef12)")
}

func stubSnapshot(
	id func(string) (string, string, error),
	load func(string, string) (*snapshots.Pointer, error),
	snap func(string, *snapshots.Pointer) (*testui.Snapshot, error),
) func() {
	prevID, prevLoad, prevSnap := snapshotIDFunc, loadPointerFunc, loadSnapshotFunc
	snapshotIDFunc = id
	loadPointerFunc = load
	loadSnapshotFunc = snap
	return func() {
		snapshotIDFunc = prevID
		loadPointerFunc = prevLoad
		loadSnapshotFunc = prevSnap
	}
}

func initStatusRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init")
	gitRun(t, dir, "config", "user.email", "test@example.com")
	gitRun(t, dir, "config", "user.name", "Test User")
	gitRun(t, dir, "config", "commit.gpgsign", "false")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test\n"), 0o644))
	gitRun(t, dir, "add", "README.md")
	gitRun(t, dir, "commit", "-m", "initial commit")
	return dir
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, out)
}

func stubFileMap(fn func(path, commit string) (*repomap.FileMap, error)) func() {
	previous := fetchFileMapFunc
	fetchFileMapFunc = fn
	return func() {
		fetchFileMapFunc = previous
	}
}

type stubStatusAgent struct{}

func (stubStatusAgent) GetType() clickyai.AgentType { return clickyai.AgentTypeClaude }
func (stubStatusAgent) GetConfig() clickyai.AgentConfig {
	return clickyai.AgentConfig{}
}
func (stubStatusAgent) ListModels(context.Context) ([]clickyai.Model, error) { return nil, nil }
func (stubStatusAgent) ExecutePrompt(context.Context, clickyai.PromptRequest) (*clickyai.PromptResponse, error) {
	return nil, nil
}
func (stubStatusAgent) ExecuteBatch(context.Context, []clickyai.PromptRequest) (map[string]*clickyai.PromptResponse, error) {
	return nil, nil
}
func (stubStatusAgent) GetCosts() clickyai.Costs { return clickyai.Costs{} }
func (stubStatusAgent) Close() error             { return nil }

type capturePromptAgent struct {
	prompt string
}

func (a *capturePromptAgent) GetType() clickyai.AgentType { return clickyai.AgentTypeClaude }
func (a *capturePromptAgent) GetConfig() clickyai.AgentConfig {
	return clickyai.AgentConfig{}
}
func (a *capturePromptAgent) ListModels(context.Context) ([]clickyai.Model, error) { return nil, nil }
func (a *capturePromptAgent) ExecutePrompt(_ context.Context, req clickyai.PromptRequest) (*clickyai.PromptResponse, error) {
	a.prompt = req.Prompt
	if schema, ok := req.StructuredOutput.(*fileSummarySchema); ok {
		schema.Summary = "tighten handler flow"
	}
	return &clickyai.PromptResponse{}, nil
}
func (a *capturePromptAgent) ExecuteBatch(context.Context, []clickyai.PromptRequest) (map[string]*clickyai.PromptResponse, error) {
	return nil, nil
}
func (a *capturePromptAgent) GetCosts() clickyai.Costs { return clickyai.Costs{} }
func (a *capturePromptAgent) Close() error             { return nil }

func stubAISummaryIO(
	diff func(workDir, path string, cached bool) (string, error),
	read func(workDir, path string) (string, error),
) func() {
	prevDiff := diffForStatusFileFunc
	prevRead := readUntrackedStatusFileFunc
	diffForStatusFileFunc = diff
	readUntrackedStatusFileFunc = read
	return func() {
		diffForStatusFileFunc = prevDiff
		readUntrackedStatusFileFunc = prevRead
	}
}

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;:]*[A-Za-z]`)

func stripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}
