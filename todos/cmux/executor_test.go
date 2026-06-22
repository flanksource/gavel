package cmux

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/commons/logger"
	todopkg "github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

func TestResolveAgent(t *testing.T) {
	cases := []struct {
		model string
		agent string
		flag  string
	}{
		{"", "claude", ""},
		{"opus", "claude", "opus"},
		{"claude", "claude", ""},
		{"gpt-5", "codex", "gpt-5"},
		{"codex", "codex", "codex"},
		{"codex-pro", "codex", "codex-pro"},
	}

	for _, tc := range cases {
		agent, flag := ResolveAgent(tc.model)
		if agent != tc.agent || flag != tc.flag {
			t.Fatalf("ResolveAgent(%q) = (%q, %q), want (%q, %q)", tc.model, agent, flag, tc.agent, tc.flag)
		}
	}
}

func TestBuildPromptAddsEffortDirective(t *testing.T) {
	prompt := BuildPrompt([]*types.TODO{{TODOFrontmatter: types.TODOFrontmatter{Title: "Fix it"}, MarkdownBody: "Body"}}, "/repo", "high")
	if !strings.HasPrefix(prompt, "Think hard") {
		t.Fatalf("prompt missing high effort directive: %q", prompt)
	}
	if !strings.Contains(prompt, "Fix it") || !strings.Contains(prompt, "Body") {
		t.Fatalf("prompt missing todo content: %q", prompt)
	}
}

func TestWritePromptFile(t *testing.T) {
	dir := t.TempDir()
	todo := &types.TODO{ID: "abc123456789", TODOFrontmatter: types.TODOFrontmatter{Title: "Fix"}}

	path, err := WritePromptFile(dir, []*types.TODO{todo}, "hello")
	if err != nil {
		t.Fatalf("WritePromptFile() error = %v", err)
	}
	if filepath.Base(path) != "prompt-abc12345.md" {
		t.Fatalf("prompt path = %s", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello\n" {
		t.Fatalf("prompt file = %q", string(data))
	}
}

func TestCmuxExecutorDispatchesPromptFileAndWaitsForIdle(t *testing.T) {
	repo := t.TempDir()
	sessionDir := t.TempDir()
	if err := os.WriteFile(
		filepath.Join(sessionDir, "claude-hook-sessions.json"),
		[]byte(`[{"sessionId":"s1","workspaceId":"ws1","cwd":`+strconv.Quote(repo)+`,"lifecycle":"idle","updatedAt":"2026-06-22T00:00:00Z"}]`),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	runner := &recordingRunner{out: map[string]string{
		joinArgs([]string{"new-workspace", "--cwd", repo, "--name", filepath.Base(repo), "--focus", "true", "--command", "cmux claude-teams --model opus", "--id-format", "both"}): "workspace=ws1 surface=sf1",
	}}
	exec := NewCmuxExecutor(CmuxExecutorConfig{
		WorkDir: repo,
		Model:   "opus",
		Effort:  "medium",
		Timeout: 50 * time.Millisecond,
		Runner:  runner.run,
		Store:   &SessionStore{Dir: sessionDir, PollInterval: time.Millisecond},
	})

	ctx := todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
	result, err := exec.ExecuteGroup(ctx, []*types.TODO{{
		ID:              "abc123456789",
		TODOFrontmatter: types.TODOFrontmatter{Title: "Fix cmux", CWD: repo},
		MarkdownBody:    "body",
	}})
	if err != nil {
		t.Fatalf("ExecuteGroup() error = %v", err)
	}
	if result == nil || !result.Success || result.ExecutorName != "cmux-claude" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if len(runner.calls) != 3 {
		t.Fatalf("runner calls = %d, want 3: %#v", len(runner.calls), runner.calls)
	}
	send := runner.calls[2].args
	if len(send) != 5 || send[0] != "send" || send[2] != "ws1" {
		t.Fatalf("unexpected send args: %#v", send)
	}
	if !strings.Contains(send[4], ".gavel/cmux/prompt-abc12345.md") || !strings.HasSuffix(send[4], "\n") {
		t.Fatalf("send instruction should reference prompt file and submit it: %q", send[4])
	}
	if _, err := os.Stat(filepath.Join(repo, ".gavel", "cmux", "prompt-abc12345.md")); err != nil {
		t.Fatalf("prompt file not written: %v", err)
	}
}
