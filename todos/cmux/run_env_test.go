package cmux

import (
	"strings"
	"testing"

	"github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos/types"
)

func TestWithRunEnv(t *testing.T) {
	todoList := []*types.TODO{{ID: "ISSUE-1"}, {ID: "ISSUE-2"}, {ID: ""}}
	got := withRunEnv("claude --session-id abc", todoList, "sess-9")

	wantIssue := commit.EnvIssueID + "='ISSUE-1,ISSUE-2'"
	wantSession := commit.EnvSessionID + "='sess-9'"
	if !strings.HasPrefix(got, wantIssue+" "+wantSession+" claude") {
		t.Fatalf("expected env-prefixed command, got: %q", got)
	}
}

func TestWithRunEnvNoMetadata(t *testing.T) {
	got := withRunEnv("codex", nil, "")
	if got != "codex" {
		t.Fatalf("expected command unchanged when no ids, got: %q", got)
	}
}

func TestShellSingleQuoteEscapes(t *testing.T) {
	if got := shellSingleQuote("a'b"); got != `'a'\''b'` {
		t.Fatalf("unexpected quoting: %q", got)
	}
}
