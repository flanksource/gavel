package status

import (
	"github.com/flanksource/repomap"
	"github.com/flanksource/repomap/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"regexp"
	"strings"
	"testing"
)

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

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;:]*[A-Za-z]`)

func stripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}
