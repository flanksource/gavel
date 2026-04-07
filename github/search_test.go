package github

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSearchQuery(t *testing.T) {
	since := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		opts       Options
		searchOpts PRSearchOptions
		expect     string
		hasErr     bool
	}{
		{
			name:       "default single repo",
			opts:       Options{Repo: "flanksource/gavel"},
			searchOpts: PRSearchOptions{Author: "@me", Since: since, State: "open"},
			expect:     "is:pr author:@me is:open updated:>2026-03-31 repo:flanksource/gavel",
		},
		{
			name:       "all repos in org",
			opts:       Options{Repo: "flanksource/gavel"},
			searchOpts: PRSearchOptions{Author: "@me", Since: since, State: "open", All: true},
			expect:     "is:pr author:@me is:open updated:>2026-03-31 org:flanksource",
		},
		{
			name:       "explicit org",
			opts:       Options{Repo: "flanksource/gavel"},
			searchOpts: PRSearchOptions{Author: "@me", Since: since, All: true, Org: "other-org"},
			expect:     "is:pr author:@me is:open updated:>2026-03-31 org:other-org",
		},
		{
			name:       "state merged",
			opts:       Options{Repo: "flanksource/gavel"},
			searchOpts: PRSearchOptions{Author: "@me", State: "merged"},
			expect:     "is:pr author:@me is:merged repo:flanksource/gavel",
		},
		{
			name:       "state closed",
			opts:       Options{Repo: "flanksource/gavel"},
			searchOpts: PRSearchOptions{State: "closed"},
			expect:     "is:pr is:closed repo:flanksource/gavel",
		},
		{
			name:       "state all",
			opts:       Options{Repo: "flanksource/gavel"},
			searchOpts: PRSearchOptions{Author: "someuser", State: "all"},
			expect:     "is:pr author:someuser repo:flanksource/gavel",
		},
		{
			name:       "no since",
			opts:       Options{Repo: "flanksource/gavel"},
			searchOpts: PRSearchOptions{Author: "@me", State: "open"},
			expect:     "is:pr author:@me is:open repo:flanksource/gavel",
		},
		{
			name:       "no author",
			opts:       Options{Repo: "flanksource/gavel"},
			searchOpts: PRSearchOptions{State: "open", Since: since},
			expect:     "is:pr is:open updated:>2026-03-31 repo:flanksource/gavel",
		},
		{
			name:       "all without repo or org fails",
			opts:       Options{WorkDir: "/tmp"},
			searchOpts: PRSearchOptions{All: true},
			hasErr:     true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := buildSearchQuery(tc.opts, tc.searchOpts)
			if tc.hasErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expect, result)
		})
	}
}

func TestBuildSearchQueryForRepo(t *testing.T) {
	since := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	result := buildSearchQueryForRepo("flanksource/duty", PRSearchOptions{
		Author: "@me", Since: since, State: "open",
	})
	assert.Equal(t, "is:pr author:@me is:open updated:>2026-03-31 repo:flanksource/duty", result)
}

func TestPRListItemPretty(t *testing.T) {
	tests := []struct {
		name           string
		item           PRListItem
		expectContains []string
	}{
		{
			name: "open PR",
			item: PRListItem{
				Number: 42, Title: "feat: add widget", Author: "moshloop",
				Repo: "flanksource/gavel", Source: "feat/widget", Target: "main",
				State: "OPEN", ReviewDecision: "APPROVED",
			},
			expectContains: []string{"●", "gavel", "#42", "feat: add widget", "APPROVED", "feat/widget", "→", "main"},
		},
		{
			name: "draft PR",
			item: PRListItem{
				Number: 10, Title: "wip: refactor", Author: "dev",
				Repo: "flanksource/duty", Source: "refactor/api", Target: "main",
				State: "OPEN", IsDraft: true,
			},
			expectContains: []string{"○", "duty", "#10", "DRAFT", "refactor/api"},
		},
		{
			name: "merged PR",
			item: PRListItem{
				Number: 99, Title: "fix: nil pointer", Author: "dev",
				Repo: "flanksource/config-db", Source: "fix/nil", Target: "main",
				State: "MERGED",
			},
			expectContains: []string{"●", "config-db", "#99", "fix/nil"},
		},
		{
			name: "current branch PR with ahead/behind",
			item: PRListItem{
				Number: 15, Title: "feat: current work", Author: "moshloop",
				Repo: "flanksource/gavel", Source: "pr/test", Target: "main",
				State: "OPEN", IsCurrent: true, Ahead: 5, Behind: 2,
			},
			expectContains: []string{"→", "●", "#15", "pr/test", "main", "↑5↓2"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := tc.item.Pretty().String()
			for _, s := range tc.expectContains {
				assert.Contains(t, output, s)
			}
		})
	}
}

func TestPRSearchResultsPretty(t *testing.T) {
	t.Run("empty results", func(t *testing.T) {
		var results PRSearchResults
		output := results.Pretty().String()
		assert.Contains(t, output, "No pull requests found")
	})

	t.Run("multiple PRs", func(t *testing.T) {
		results := PRSearchResults{
			{Number: 1, Title: "first", Repo: "flanksource/a", State: "OPEN", Source: "br1"},
			{Number: 2, Title: "second", Repo: "flanksource/b", State: "MERGED", Source: "br2"},
		}
		output := results.Pretty().String()
		assert.Contains(t, output, "Pull Requests (2)")
		assert.Contains(t, output, "#1")
		assert.Contains(t, output, "#2")
	})
}

func TestCheckSummaryPretty(t *testing.T) {
	tests := []struct {
		name           string
		cs             CheckSummary
		expectContains []string
	}{
		{
			name:           "all passed",
			cs:             CheckSummary{Passed: 5},
			expectContains: []string{"✓5"},
		},
		{
			name:           "mixed",
			cs:             CheckSummary{Passed: 3, Failed: 1, Running: 2, Pending: 1},
			expectContains: []string{"✓3", "✗1", "●2", "○1"},
		},
		{
			name:           "only failed",
			cs:             CheckSummary{Failed: 2},
			expectContains: []string{"✗2"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := tc.cs.PrettySummary().String()
			for _, s := range tc.expectContains {
				assert.Contains(t, output, s)
			}
		})
	}
}

func TestFailedCheckPretty(t *testing.T) {
	t.Run("name only", func(t *testing.T) {
		f := FailedCheck{Name: "lint"}
		output := f.Pretty("  ").String()
		assert.Contains(t, output, "✗")
		assert.Contains(t, output, "lint")
	})

	t.Run("with steps and logs", func(t *testing.T) {
		f := FailedCheck{Name: "test", FailedSteps: []string{"Run tests"}, LogTail: "FAIL: TestFoo"}
		output := f.Pretty("  ").String()
		assert.Contains(t, output, "test")
		assert.Contains(t, output, "Run tests")
		assert.Contains(t, output, "FAIL: TestFoo")
	})
}

func TestComputeCheckSummary(t *testing.T) {
	success := "SUCCESS"
	failure := "FAILURE"
	node := searchPRNode{
		Commits: graphQLCommits{
			Nodes: []graphQLCommitNode{{
				Commit: graphQLCommit{
					StatusCheckRollup: &graphQLStatusCheckRollup{
						Contexts: graphQLContexts{
							Nodes: []graphQLCheckNode{
								{Typename: "CheckRun", Name: "build", Status: "COMPLETED", Conclusion: &success},
								{Typename: "CheckRun", Name: "lint", Status: "COMPLETED", Conclusion: &failure, DetailsURL: "https://github.com/org/repo/actions/runs/123/job/456"},
								{Typename: "CheckRun", Status: "IN_PROGRESS"},
								{Typename: "CheckRun", Status: "QUEUED"},
								{Typename: "StatusContext", State: "success", Context: "ci/ext"},
								{Typename: "StatusContext", State: "pending"},
							},
						},
					},
				},
			}},
		},
	}

	cs := computeCheckSummary(node)
	require.NotNil(t, cs)
	assert.Equal(t, 2, cs.Passed)
	assert.Equal(t, 1, cs.Failed)
	assert.Equal(t, 1, cs.Running)
	assert.Equal(t, 2, cs.Pending)
	require.Len(t, cs.Failures, 1)
	assert.Equal(t, "lint", cs.Failures[0].Name)
	assert.Contains(t, cs.Failures[0].DetailsURL, "runs/123")
}

func TestComputeCheckSummaryNil(t *testing.T) {
	assert.Nil(t, computeCheckSummary(searchPRNode{}))
}

func TestPRListItemWithCheckStatus(t *testing.T) {
	item := PRListItem{
		Number: 42, Title: "feat: widget", Author: "dev",
		Repo: "flanksource/gavel", Source: "feat/widget", Target: "main",
		State: "OPEN",
		CheckStatus: &CheckSummary{Passed: 3, Failed: 1},
	}
	output := item.Pretty().String()
	assert.Contains(t, output, "✓3")
	assert.Contains(t, output, "✗1")
}

func TestStateIcon(t *testing.T) {
	tests := []struct {
		state, expect string
		isDraft       bool
	}{
		{"OPEN", "●", false},
		{"MERGED", "●", false},
		{"CLOSED", "●", false},
		{"OPEN", "○", true},
	}
	for _, tc := range tests {
		icon := StateIcon(tc.state, tc.isDraft)
		assert.Contains(t, icon.String(), tc.expect,
			"state=%s isDraft=%v", tc.state, tc.isDraft)
	}
}
