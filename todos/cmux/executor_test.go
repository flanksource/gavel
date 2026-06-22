package cmux

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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
		{"codex", "codex", ""},
		{"codex-pro", "codex", "codex-pro"},
	}

	for _, tc := range cases {
		agent, flag := ResolveAgent(tc.model)
		if agent != tc.agent || flag != tc.flag {
			t.Fatalf("ResolveAgent(%q) = (%q, %q), want (%q, %q)", tc.model, agent, flag, tc.agent, tc.flag)
		}
	}
}

func TestAgentCommandUsesDirectCLI(t *testing.T) {
	cases := []struct {
		agent string
		model string
		want  string
	}{
		{"claude", "", "claude"},
		{"claude", "opus", "claude --model opus"},
		{"codex", "", "codex"},
		{"codex", "gpt-5", "codex -m gpt-5"},
	}

	for _, tc := range cases {
		if got := AgentCommand(tc.agent, tc.model); got != tc.want {
			t.Fatalf("AgentCommand(%q, %q) = %q, want %q", tc.agent, tc.model, got, tc.want)
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

func TestScreenPollDelayBacksOffAndCaps(t *testing.T) {
	now := time.Now()
	base := time.Second
	max := 5 * time.Second

	cases := []struct {
		elapsed time.Duration
		want    time.Duration
	}{
		{0, time.Second},
		{9 * time.Second, time.Second},
		{10 * time.Second, 2 * time.Second},
		{20 * time.Second, 4 * time.Second},
		{30 * time.Second, 5 * time.Second},
		{2 * time.Minute, 5 * time.Second},
	}

	for _, tc := range cases {
		start := now.Add(-tc.elapsed)
		if got := screenPollDelay(start, base, max); got != tc.want {
			t.Fatalf("screenPollDelay(%s) = %s, want %s", tc.elapsed, got, tc.want)
		}
	}
}

func TestCmuxExecutorDispatchesPromptFileAndWaitsForIdle(t *testing.T) {
	repo := t.TempDir()
	runner := newScreenRunner(repo)
	exec := NewCmuxExecutor(CmuxExecutorConfig{
		WorkDir:              repo,
		Model:                "opus",
		Effort:               "medium",
		Timeout:              100 * time.Millisecond,
		Runner:               runner.run,
		ScreenPollInterval:   time.Millisecond,
		ScreenStableDuration: time.Millisecond,
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
	workspaceName := filepath.Base(repo) + "-claude"
	assertCalled(t, runner.calls, []string{"list-workspaces", "--json"})
	assertCalled(t, runner.calls, []string{"new-workspace", "--cwd", repo, "--name", workspaceName, "--focus", "true", "--description", fmt.Sprintf("gavel todos claude workspace for %s", repo), "--id-format", "both"})
	assertCalled(t, runner.calls, []string{"new-surface", "--type", "terminal", "--workspace", "workspace:ws1", "--working-directory", repo, "--focus", "true"})
	assertCalled(t, runner.calls, []string{"read-screen", "--workspace", "workspace:ws1", "--surface", "surface:sf1", "--lines", "120"})
	var sends [][]string
	var enterKeys [][]string
	for _, call := range runner.calls {
		if len(call.args) > 0 && call.args[0] == "send" {
			sends = append(sends, call.args)
		}
		if len(call.args) > 0 && call.args[0] == "send-key" {
			enterKeys = append(enterKeys, call.args)
		}
	}
	if len(sends) != 2 {
		t.Fatalf("send calls = %#v, want 2", sends)
	}
	if !reflect.DeepEqual(sends[0], []string{"send", "--workspace", "workspace:ws1", "--surface", "surface:sf1", "--", "claude --model opus"}) {
		t.Fatalf("agent send args = %#v", sends[0])
	}
	if len(sends[1]) != 7 || sends[1][0] != "send" || sends[1][2] != "workspace:ws1" || sends[1][4] != "surface:sf1" {
		t.Fatalf("unexpected prompt send args: %#v", sends[1])
	}
	if !strings.Contains(sends[1][6], ".gavel/cmux/prompt-abc12345.md") || strings.HasSuffix(sends[1][6], "\n") {
		t.Fatalf("send instruction should reference prompt file and submit it: %q", sends[1][6])
	}
	if len(enterKeys) != 2 {
		t.Fatalf("send-key calls = %#v, want 2", enterKeys)
	}
	for _, enter := range enterKeys {
		if !reflect.DeepEqual(enter, []string{"send-key", "--workspace", "workspace:ws1", "--surface", "surface:sf1", "Enter"}) {
			t.Fatalf("send-key args = %#v", enter)
		}
	}
	if _, err := os.Stat(filepath.Join(repo, ".gavel", "cmux", "prompt-abc12345.md")); err != nil {
		t.Fatalf("prompt file not written: %v", err)
	}
}

func TestCmuxExecutorRetriesInitialPromptSend(t *testing.T) {
	repo := t.TempDir()
	agentSends := 0
	promptSends := 0
	readScreens := 0
	runner := func(_ context.Context, _ string, _ string, _ time.Duration, args ...string) (string, error) {
		switch args[0] {
		case "list-workspaces":
			return `{"workspaces":[]}`, nil
		case "new-workspace":
			return "workspace:ws1", nil
		case "new-surface":
			return "surface:sf1", nil
		case "read-screen":
			readScreens++
			if agentSends == 0 {
				return "shell ready\n$ ", nil
			}
			if promptSends == 0 {
				return "Claude ready\n> ", nil
			}
			if readScreens < 5 {
				return "Claude is working", nil
			}
			return "Claude done\n> ", nil
		case "send":
			payload := args[len(args)-1]
			if payload == "claude" {
				agentSends++
				return "ok", nil
			}
			promptSends++
			if promptSends == 1 {
				return "", fmt.Errorf("surface not ready")
			}
		}
		return "ok", nil
	}
	exec := NewCmuxExecutor(CmuxExecutorConfig{
		WorkDir:              repo,
		Timeout:              100 * time.Millisecond,
		Runner:               runner,
		SendAttempts:         2,
		SendRetryDelay:       time.Millisecond,
		ScreenPollInterval:   time.Millisecond,
		ScreenStableDuration: time.Millisecond,
	})

	ctx := todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
	result, err := exec.ExecuteGroup(ctx, []*types.TODO{{
		ID:              "abc123456789",
		TODOFrontmatter: types.TODOFrontmatter{Title: "Fix cmux", CWD: repo},
	}})
	if err != nil {
		t.Fatalf("ExecuteGroup() error = %v", err)
	}
	if result == nil || !result.Success {
		t.Fatalf("unexpected result: %#v", result)
	}
	if agentSends != 1 {
		t.Fatalf("agent sends = %d, want 1", agentSends)
	}
	if promptSends != 2 {
		t.Fatalf("prompt sends = %d, want 2", promptSends)
	}
}

type screenRunner struct {
	calls         []runnerCall
	pendingSubmit string
	agentSent     bool
	promptSent    bool
}

func newScreenRunner(_ string) *screenRunner {
	return &screenRunner{}
}

func (r *screenRunner) run(_ context.Context, cwd, binary string, _ time.Duration, args ...string) (string, error) {
	r.calls = append(r.calls, runnerCall{cwd: cwd, binary: binary, args: append([]string(nil), args...)})
	switch args[0] {
	case "list-workspaces":
		return `{"workspaces":[]}`, nil
	case "new-workspace":
		return "workspace:ws1", nil
	case "new-surface":
		return "surface:sf1", nil
	case "send":
		r.pendingSubmit = args[len(args)-1]
		return "ok", nil
	case "send-key":
		if r.pendingSubmit == "" {
			return "ok", nil
		}
		if strings.HasPrefix(r.pendingSubmit, "claude") {
			r.agentSent = true
		} else {
			r.promptSent = true
		}
		r.pendingSubmit = ""
		return "ok", nil
	case "read-screen":
		if !r.agentSent {
			return "shell ready\n$ ", nil
		}
		if !r.promptSent {
			return "Claude ready\n> ", nil
		}
		return "Claude done\n> ", nil
	default:
		return "ok", nil
	}
}

func assertCalled(t *testing.T, calls []runnerCall, want []string) {
	t.Helper()
	for _, call := range calls {
		if len(call.args) == len(want) {
			match := true
			for i := range want {
				if call.args[i] != want[i] {
					match = false
					break
				}
			}
			if match {
				return
			}
		}
	}
	t.Fatalf("missing call %#v in %#v", want, calls)
}
