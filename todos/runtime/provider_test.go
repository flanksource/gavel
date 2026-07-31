package runtime

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanResultContentPrefersReferencedPlanFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plan.md")
	const detailed = "# Detailed plan\n\n1. Inspect the parser.\n2. Add regression coverage.\n"
	require.NoError(t, os.WriteFile(path, []byte(detailed), 0o600))

	content, gotPath, err := planResultContent(&types.TODO{}, &todos.ExecutionResult{
		Plan: &types.PlanResult{
			Path:    path,
			Content: "The existing plan remains correct.",
		},
	})
	require.NoError(t, err)
	assert.Equal(t, path, gotPath)
	assert.Equal(t, strings.TrimSpace(detailed), content)
}

func TestTodoStatusKeepsDurableAndExecutionStateSeparate(t *testing.T) {
	tests := []struct {
		name      string
		status    native.IssueStatus
		execution native.ExecutionState
		want      types.Status
	}{
		{name: "draft", status: native.StatusDraft, execution: native.ExecutionRunning, want: types.StatusDraft},
		{name: "verified", status: native.StatusVerified, execution: native.ExecutionFailed, want: types.StatusVerified},
		{name: "closed", status: native.StatusClosed, execution: native.ExecutionRunning, want: types.StatusCompleted},
		{name: "cancelled", status: native.StatusCancelled, execution: native.ExecutionIdle, want: types.StatusCompleted},
		{name: "idle", status: native.StatusOpen, execution: native.ExecutionIdle, want: types.StatusPending},
		{name: "planning", status: native.StatusOpen, execution: native.ExecutionPlanning, want: types.StatusInProgress},
		{name: "running", status: native.StatusOpen, execution: native.ExecutionRunning, want: types.StatusInProgress},
		{name: "waiting", status: native.StatusOpen, execution: native.ExecutionWaiting, want: types.StatusAsk},
		{name: "stalled", status: native.StatusOpen, execution: native.ExecutionStalled, want: types.StatusFailed},
		{name: "failed", status: native.StatusOpen, execution: native.ExecutionFailed, want: types.StatusFailed},
		{name: "verifying", status: native.StatusOpen, execution: native.ExecutionVerifying, want: types.StatusInProgress},
		{name: "verification failed", status: native.StatusOpen, execution: native.ExecutionVerificationFailed, want: types.StatusUnverified},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, todoStatus(test.status, test.execution))
		})
	}
}

func TestTransientStatusesAreNotDurableIssueWrites(t *testing.T) {
	for _, status := range []types.Status{
		types.StatusInProgress,
		types.StatusReview,
		types.StatusAsk,
		types.StatusFailed,
		types.StatusUnverified,
	} {
		t.Run(string(status), func(t *testing.T) {
			_, durable, err := toDurableStatus(status)
			require.NoError(t, err)
			assert.False(t, durable)
		})
	}

	tests := map[types.Status]native.IssueStatus{
		types.StatusDraft:     native.StatusDraft,
		types.StatusPending:   native.StatusOpen,
		types.StatusVerified:  native.StatusVerified,
		types.StatusCompleted: native.StatusClosed,
		types.StatusSkipped:   native.StatusVerified,
	}
	for status, want := range tests {
		nativeStatus, durable, err := toDurableStatus(status)
		require.NoError(t, err)
		assert.True(t, durable)
		assert.Equal(t, want, nativeStatus)
	}
}

func TestRenderAttemptReportsResolvedRuntimeAndMultilineError(t *testing.T) {
	body := renderAttempt(&types.TODO{TODOFrontmatter: types.TODOFrontmatter{Attempts: 1}}, &todos.ExecutionResult{
		Runtime: todos.RunStartMetadata{
			Driver: "cli", Agent: "claude", Provider: "anthropic",
			Backend: "claude-agent", ResolvedModel: "claude-sonnet-5", Effort: "high",
		},
		ErrorMessage: "Claude Code process exited with code 1\nstderr:\nauthentication failed",
	})

	for _, expected := range []string{
		"- **Driver:** cli",
		"- **Agent:** claude",
		"- **Provider:** anthropic",
		"- **Backend:** claude-agent",
		"- **Model:** claude-sonnet-5",
		"- **Effort:** high",
		"- **Error:**\n\n```text\nClaude Code process exited with code 1\nstderr:\nauthentication failed\n```",
	} {
		assert.Contains(t, body, expected)
	}
	assert.NotContains(t, body, "- **Model:** headless-claude")
}

func TestTodoFromIssueUsesDatabaseIdentityAndVerification(t *testing.T) {
	workspaceID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	issueID := uuid.MustParse("20000000-0000-0000-0000-000000000001")
	provider := &Provider{workspace: &native.Workspace{
		ID: workspaceID, RootPath: "/work/gavel", DisplayName: "Gavel",
	}}
	issue := &native.Issue{
		ID: issueID, WorkspaceID: workspaceID, Version: 7,
		Title: "Native issue", Body: "Description\n", Verification: "```bash\ntrue\n```",
		Labels: []string{"database", "todos"}, Priority: native.PriorityCritical,
		Status: native.StatusOpen, ExecutionState: native.ExecutionWaiting,
	}

	todo, err := provider.todoFromIssue(context.Background(), issue, "/retained/gavel", false)
	require.NoError(t, err)
	assert.Equal(t, todos.ProviderDB, todo.Provider)
	assert.Equal(t, issueID.String(), todo.ID)
	assert.Equal(t, int64(7), todo.Version)
	assert.Equal(t, workspaceID.String(), todo.WorkspaceID)
	assert.Equal(t, string(native.ExecutionWaiting), todo.ExecutionState)
	assert.Equal(t, types.StatusAsk, todo.Status)
	assert.Equal(t, types.PriorityHigh, todo.Priority)
	assert.Equal(t, "Gavel", todo.Workspace)
	assert.Equal(t, "/retained/gavel", todo.CWD)
	assert.Equal(t, "Description", todo.MarkdownBody)
	assert.Equal(t, "```bash\ntrue\n```", todo.VerificationMarkdown)
}
