package commit

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/flanksource/clicky/prompt"
)

func TestResolveAgentCommitDir(t *testing.T) {
	abs := filepath.Join(string(filepath.Separator), "abs", "path")
	cases := []struct {
		name    string
		workDir string
		cwd     string
		want    string
	}{
		{"empty cwd falls back to workDir", "/repo", "", "/repo"},
		{"relative cwd joins workDir", "/repo", "sub/dir", filepath.Clean("/repo/sub/dir")},
		{"absolute cwd is used verbatim", "/repo", abs, abs},
		{"relative cwd without workDir is cleaned", "", "sub/dir", filepath.Clean("sub/dir")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveAgentCommitDir(tc.workDir, tc.cwd); got != tc.want {
				t.Fatalf("resolveAgentCommitDir(%q, %q) = %q, want %q", tc.workDir, tc.cwd, got, tc.want)
			}
		})
	}
}

func TestAgentPromptScopeKeepsTheRunningTaskLabel(t *testing.T) {
	const taskID, issueID, sessionID = "task-1", "issue-1", "session-1"
	inTask := prompt.WithScope(context.Background(), prompt.Scope{Labels: map[string]string{prompt.TaskLabel: taskID}})

	got := prompt.ScopeFrom(agentPromptScope(inTask, AgentRunMetadata{IssueID: issueID, SessionID: sessionID}))

	want := prompt.Scope{Owner: issueID, Kind: "commit", Labels: map[string]string{prompt.TaskLabel: taskID, "session": sessionID}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("agentPromptScope() = %+v, want %+v", got, want)
	}
}

func TestAgentPromptScopeLeavesAnUnidentifiedRunUnscoped(t *testing.T) {
	if got := prompt.ScopeFrom(agentPromptScope(context.Background(), AgentRunMetadata{})); !got.IsZero() {
		t.Fatalf("agentPromptScope() without issue or session = %+v, want the zero scope", got)
	}
}
