package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strptr(s string) *string { return &s }

// rollupNode builds a CheckRun context node with the given conclusion so the
// classification table reads declaratively.
func checkRun(name, status, conclusion string) graphQLCheckNode {
	n := graphQLCheckNode{Typename: "CheckRun", Name: name, Status: status}
	if conclusion != "" {
		n.Conclusion = strptr(conclusion)
	}
	return n
}

// TestSummarizeRollup guards the classification logic shared by PR search and
// org branch status after it was extracted from computeCheckSummary. The
// expectations are computed independently from the input, not from the
// function's own output.
func TestSummarizeRollup(t *testing.T) {
	t.Run("nil rollup", func(t *testing.T) {
		assert.Nil(t, summarizeRollup(nil))
	})

	t.Run("mixed conclusions", func(t *testing.T) {
		rollup := &graphQLStatusCheckRollup{Contexts: graphQLContexts{Nodes: []graphQLCheckNode{
			checkRun("build", "COMPLETED", "SUCCESS"),
			checkRun("lint", "COMPLETED", "NEUTRAL"),
			checkRun("docs", "COMPLETED", "SKIPPED"),
			checkRun("test", "COMPLETED", "FAILURE"),
			checkRun("e2e", "COMPLETED", "TIMED_OUT"),
			checkRun("deploy", "IN_PROGRESS", ""),
			checkRun("queued", "QUEUED", ""),
		}}}

		cs := summarizeRollup(rollup)
		require.NotNil(t, cs)
		assert.Equal(t, 3, cs.Passed, "SUCCESS+NEUTRAL+SKIPPED count as passed")
		assert.Equal(t, 2, cs.Failed, "FAILURE+TIMED_OUT count as failed")
		assert.Equal(t, 1, cs.Running, "IN_PROGRESS counts as running")
		assert.Equal(t, 1, cs.Pending, "QUEUED falls through to pending")
		assert.Len(t, cs.Failures, 2)
	})
}

// TestComputeCheckSummaryDelegates confirms the PR-path wrapper still returns
// nil when there are no commits and otherwise matches summarizeRollup.
func TestComputeCheckSummaryDelegates(t *testing.T) {
	assert.Nil(t, computeCheckSummary(searchPRNode{}))

	node := searchPRNode{Commits: graphQLCommits{Nodes: []graphQLCommitNode{{
		Commit: graphQLCommit{StatusCheckRollup: &graphQLStatusCheckRollup{
			Contexts: graphQLContexts{Nodes: []graphQLCheckNode{
				checkRun("build", "COMPLETED", "SUCCESS"),
				checkRun("test", "COMPLETED", "FAILURE"),
			}},
		}},
	}}}}

	cs := computeCheckSummary(node)
	require.NotNil(t, cs)
	assert.Equal(t, 1, cs.Passed)
	assert.Equal(t, 1, cs.Failed)
}
