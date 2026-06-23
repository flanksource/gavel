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
		name string
		opts AgentCommandOpts
		want string
	}{
		{"claude default", AgentCommandOpts{Agent: "claude"}, "claude"},
		{"claude model", AgentCommandOpts{Agent: "claude", Model: "opus"}, "claude --model opus"},
		{"claude session+model", AgentCommandOpts{Agent: "claude", Model: "opus", SessionID: "sess-123"}, "claude --session-id sess-123 --model opus"},
		{"claude session", AgentCommandOpts{Agent: "claude", SessionID: "sess-123"}, "claude --session-id sess-123"},
		{"claude resume", AgentCommandOpts{Agent: "claude", SessionID: "sess-123", Resume: true}, "claude --resume sess-123"},
		{"claude resume+model", AgentCommandOpts{Agent: "claude", SessionID: "sess-123", Resume: true, Model: "opus"}, "claude --resume sess-123 --model opus"},
		{"claude resume without session", AgentCommandOpts{Agent: "claude", Resume: true}, "claude"},
		{"claude plan", AgentCommandOpts{Agent: "claude", SessionID: "sess-123", Plan: true}, "claude --session-id sess-123 --permission-mode plan"},
		{"claude plan+model", AgentCommandOpts{Agent: "claude", Model: "opus", Plan: true}, "claude --permission-mode plan --model opus"},
		{"codex default", AgentCommandOpts{Agent: "codex"}, "codex"},
		{"codex model", AgentCommandOpts{Agent: "codex", Model: "gpt-5"}, "codex -m gpt-5"},
		{"codex ignores session", AgentCommandOpts{Agent: "codex", Model: "gpt-5", SessionID: "sess-123"}, "codex -m gpt-5"},
		{"codex plan has no flag", AgentCommandOpts{Agent: "codex", Plan: true}, "codex"},
	}

	for _, tc := range cases {
		if got := AgentCommand(tc.opts); got != tc.want {
			t.Fatalf("%s: AgentCommand(%+v) = %q, want %q", tc.name, tc.opts, got, tc.want)
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

func TestSessionHistoryDirective(t *testing.T) {
	if got := SessionHistoryDirective(""); got != "" {
		t.Fatalf("SessionHistoryDirective(\"\") = %q, want empty", got)
	}
	got := SessionHistoryDirective("sess-abc")
	if !strings.Contains(got, "sess-abc") {
		t.Fatalf("SessionHistoryDirective(sess-abc) = %q, want it to reference the session id", got)
	}
}

func TestCmuxExecutorResumesPriorSession(t *testing.T) {
	repo := t.TempDir()
	runner := newScreenRunner(repo)
	exec := NewCmuxExecutor(CmuxExecutorConfig{
		WorkDir:                 repo,
		Resume:                  true,
		Timeout:                 100 * time.Millisecond,
		Runner:                  runner.run,
		ScreenPollInterval:      time.Millisecond,
		ScreenStableDuration:    time.Millisecond,
		SessionLogPollInterval:  time.Millisecond,
		SessionLogAppearTimeout: time.Millisecond, // no real session log → fall back to screen-idle
	})

	ctx := todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
	todo := &types.TODO{
		ID:              "abc123456789",
		TODOFrontmatter: types.TODOFrontmatter{Title: "Fix cmux", CWD: repo, LLM: &types.LLM{SessionId: "prior-session"}},
	}
	if _, err := exec.ExecuteGroup(ctx, []*types.TODO{todo}); err != nil {
		t.Fatalf("ExecuteGroup() error = %v", err)
	}

	var agentSend string
	for _, call := range runner.calls {
		if len(call.args) > 0 && call.args[0] == "send" {
			payload := call.args[len(call.args)-1]
			if strings.HasPrefix(payload, "claude") {
				agentSend = payload
				break
			}
		}
	}
	if agentSend != "claude --resume prior-session" {
		t.Fatalf("agent send = %q, want %q", agentSend, "claude --resume prior-session")
	}
	// Resume reuses the prior id rather than minting a new one.
	if todo.LLM == nil || todo.LLM.SessionId != "prior-session" {
		t.Fatalf("session id changed on resume: %+v", todo.LLM)
	}
}

func TestCmuxExecutorRecordsSessionBeforeAgentLaunch(t *testing.T) {
	repo := t.TempDir()
	runner := newScreenRunner(repo)
	exec := NewCmuxExecutor(CmuxExecutorConfig{
		WorkDir:                 repo,
		SessionID:               "fixed-session",
		Timeout:                 100 * time.Millisecond,
		Runner:                  runner.run,
		ScreenPollInterval:      time.Millisecond,
		ScreenStableDuration:    time.Millisecond,
		SessionLogPollInterval:  time.Millisecond,
		SessionLogAppearTimeout: time.Millisecond,
	})

	ctx := todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
	var recorded string
	agentSentBeforeHook := false
	ctx.SetSessionIDHook(func(sid string) {
		recorded = sid
		for _, call := range runner.calls {
			if len(call.args) > 0 && call.args[0] == "send" && strings.HasPrefix(call.args[len(call.args)-1], "claude") {
				agentSentBeforeHook = true
			}
		}
	})
	todo := &types.TODO{
		ID:              "abc123456789",
		TODOFrontmatter: types.TODOFrontmatter{Title: "Fix cmux", CWD: repo},
	}
	if _, err := exec.ExecuteGroup(ctx, []*types.TODO{todo}); err != nil {
		t.Fatalf("ExecuteGroup() error = %v", err)
	}
	if recorded != "fixed-session" {
		t.Fatalf("session id hook recorded %q, want fixed-session", recorded)
	}
	if agentSentBeforeHook {
		t.Fatal("session id must be recorded before the claude agent command is sent")
	}
}

func TestCmuxExecutorUsesConfiguredSessionID(t *testing.T) {
	repo := t.TempDir()
	runner := newScreenRunner(repo)
	exec := NewCmuxExecutor(CmuxExecutorConfig{
		WorkDir:                 repo,
		SessionID:               "fixed-session",
		Timeout:                 100 * time.Millisecond,
		Runner:                  runner.run,
		ScreenPollInterval:      time.Millisecond,
		ScreenStableDuration:    time.Millisecond,
		SessionLogPollInterval:  time.Millisecond,
		SessionLogAppearTimeout: time.Millisecond,
	})

	ctx := todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
	todo := &types.TODO{
		ID:              "abc123456789",
		TODOFrontmatter: types.TODOFrontmatter{Title: "Fix cmux", CWD: repo},
	}
	if _, err := exec.ExecuteGroup(ctx, []*types.TODO{todo}); err != nil {
		t.Fatalf("ExecuteGroup() error = %v", err)
	}

	var agentSend string
	for _, call := range runner.calls {
		if len(call.args) > 0 && call.args[0] == "send" {
			payload := call.args[len(call.args)-1]
			if strings.HasPrefix(payload, "claude") {
				agentSend = payload
				break
			}
		}
	}
	if agentSend != "claude --session-id fixed-session" {
		t.Fatalf("agent send = %q, want %q", agentSend, "claude --session-id fixed-session")
	}
	if todo.LLM == nil || todo.LLM.SessionId != "fixed-session" {
		t.Fatalf("configured session id not recorded: %+v", todo.LLM)
	}
}

func TestCmuxExecutorFreshSessionReferencesPriorInPrompt(t *testing.T) {
	repo := t.TempDir()
	runner := newScreenRunner(repo)
	// Resume disabled: a prior session exists but we start a fresh one and tell
	// the agent about the prior session id for history.
	exec := NewCmuxExecutor(CmuxExecutorConfig{
		WorkDir:                 repo,
		Timeout:                 100 * time.Millisecond,
		Runner:                  runner.run,
		ScreenPollInterval:      time.Millisecond,
		ScreenStableDuration:    time.Millisecond,
		SessionLogPollInterval:  time.Millisecond,
		SessionLogAppearTimeout: time.Millisecond,
	})

	ctx := todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
	todo := &types.TODO{
		ID:              "abc123456789",
		TODOFrontmatter: types.TODOFrontmatter{Title: "Fix cmux", CWD: repo, LLM: &types.LLM{SessionId: "prior-session"}},
	}
	if _, err := exec.ExecuteGroup(ctx, []*types.TODO{todo}); err != nil {
		t.Fatalf("ExecuteGroup() error = %v", err)
	}

	// A fresh session id is minted (not the prior one).
	if todo.LLM == nil || todo.LLM.SessionId == "" || todo.LLM.SessionId == "prior-session" {
		t.Fatalf("expected a fresh session id, got %+v", todo.LLM)
	}
	data, err := os.ReadFile(filepath.Join(repo, ".gavel", "cmux", "prompt-abc12345.md"))
	if err != nil {
		t.Fatalf("read prompt file: %v", err)
	}
	if !strings.Contains(string(data), "prior-session") {
		t.Fatalf("prompt should reference the prior session id for history:\n%s", data)
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
		WorkDir:                 repo,
		Model:                   "opus",
		Effort:                  "medium",
		Timeout:                 100 * time.Millisecond,
		Runner:                  runner.run,
		ScreenPollInterval:      time.Millisecond,
		ScreenStableDuration:    time.Millisecond,
		SessionLogPollInterval:  time.Millisecond,
		SessionLogAppearTimeout: time.Millisecond, // no real session log → fall back to screen-idle
	})

	ctx := todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
	todo := &types.TODO{
		ID:              "abc123456789",
		TODOFrontmatter: types.TODOFrontmatter{Title: "Fix cmux", CWD: repo},
		MarkdownBody:    "body",
	}
	result, err := exec.ExecuteGroup(ctx, []*types.TODO{todo})
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
	agentSend := sends[0]
	if len(agentSend) != 7 || agentSend[0] != "send" || agentSend[2] != "workspace:ws1" || agentSend[4] != "surface:sf1" || agentSend[5] != "--" {
		t.Fatalf("unexpected agent send args: %#v", agentSend)
	}
	if payload := agentSend[6]; !strings.HasPrefix(payload, "claude --session-id ") || !strings.HasSuffix(payload, " --model opus") {
		t.Fatalf("agent send payload = %q, want claude --session-id <uuid> --model opus", payload)
	}
	if todo.LLM == nil || todo.LLM.SessionId == "" {
		t.Fatalf("expected session id recorded on todo, got %+v", todo.LLM)
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
			if strings.HasPrefix(payload, "claude") {
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
		WorkDir:                 repo,
		Timeout:                 100 * time.Millisecond,
		Runner:                  runner,
		SendAttempts:            2,
		SendRetryDelay:          time.Millisecond,
		ScreenPollInterval:      time.Millisecond,
		ScreenStableDuration:    time.Millisecond,
		SessionLogPollInterval:  time.Millisecond,
		SessionLogAppearTimeout: time.Millisecond,
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
