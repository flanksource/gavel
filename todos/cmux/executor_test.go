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
	fakeClaudeHome(t)
	runner := newScreenRunner(repo)
	exec := NewCmuxExecutor(CmuxExecutorConfig{
		WorkDir:                 repo,
		Resume:                  true,
		Timeout:                 2 * time.Second,
		Runner:                  runner.run,
		SendSettleDelay:         time.Millisecond,
		ScreenPollInterval:      time.Millisecond,
		ScreenStableDuration:    time.Millisecond,
		SessionLogPollInterval:  time.Millisecond,
		SessionLogAppearTimeout: time.Second,
	})

	// A resumed run seeks past the existing log, so it only completes on a turn
	// appended after tailing begins.
	logPath := sessionLogFile(t, repo, "prior-session")
	if err := os.WriteFile(logPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	appended := appendSessionLineAfter(logPath, completedSessionLine, 30*time.Millisecond)

	ctx := todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
	todo := &types.TODO{
		ID:              "abc123456789",
		TODOFrontmatter: types.TODOFrontmatter{Title: "Fix cmux", CWD: repo, LLM: &types.LLM{SessionId: "prior-session"}},
	}
	if _, err := exec.ExecuteGroup(ctx, []*types.TODO{todo}); err != nil {
		t.Fatalf("ExecuteGroup() error = %v", err)
	}
	if err := <-appended; err != nil {
		t.Fatalf("append session line: %v", err)
	}

	var agentSend string
	for _, call := range runner.calls {
		if len(call.args) > 0 && call.args[0] == "send" {
			payload := call.args[len(call.args)-1]
			if strings.Contains(payload, "claude --") {
				agentSend = payload
				break
			}
		}
	}
	if !strings.HasSuffix(agentSend, "claude --resume prior-session") {
		t.Fatalf("agent send = %q, want suffix %q", agentSend, "claude --resume prior-session")
	}
	// Resume reuses the prior id rather than minting a new one.
	if todo.LLM == nil || todo.LLM.SessionId != "prior-session" {
		t.Fatalf("session id changed on resume: %+v", todo.LLM)
	}
}

func TestCmuxExecutorRecordsSessionBeforeAgentLaunch(t *testing.T) {
	repo := t.TempDir()
	fakeClaudeHome(t)
	runner := newScreenRunner(repo)
	exec := NewCmuxExecutor(CmuxExecutorConfig{
		WorkDir:                 repo,
		SessionID:               "fixed-session",
		Timeout:                 2 * time.Second,
		Runner:                  runner.run,
		ScreenPollInterval:      time.Millisecond,
		ScreenStableDuration:    time.Millisecond,
		SessionLogPollInterval:  time.Millisecond,
		SessionLogAppearTimeout: time.Second,
	})

	ctx := todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
	var recorded string
	agentSentBeforeHook := false
	ctx.SetSessionIDHook(func(sid string) {
		recorded = sid
		for _, call := range runner.calls {
			if len(call.args) > 0 && call.args[0] == "send" && strings.Contains(call.args[len(call.args)-1], "claude --") {
				agentSentBeforeHook = true
			}
		}
		writeSessionLog(t, sessionLogFile(t, repo, sid), completedSessionLine)
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
	fakeClaudeHome(t)
	runner := newScreenRunner(repo)
	exec := NewCmuxExecutor(CmuxExecutorConfig{
		WorkDir:                 repo,
		SessionID:               "fixed-session",
		Timeout:                 2 * time.Second,
		Runner:                  runner.run,
		ScreenPollInterval:      time.Millisecond,
		ScreenStableDuration:    time.Millisecond,
		SessionLogPollInterval:  time.Millisecond,
		SessionLogAppearTimeout: time.Second,
	})

	ctx := todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
	completeSessionOnRecord(t, ctx, repo)
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
			if strings.Contains(payload, "claude --") {
				agentSend = payload
				break
			}
		}
	}
	if !strings.HasSuffix(agentSend, "claude --session-id fixed-session") {
		t.Fatalf("agent send = %q, want suffix %q", agentSend, "claude --session-id fixed-session")
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
		Timeout:                 2 * time.Second,
		Runner:                  runner.run,
		ScreenPollInterval:      time.Millisecond,
		ScreenStableDuration:    time.Millisecond,
		SessionLogPollInterval:  time.Millisecond,
		SessionLogAppearTimeout: time.Second,
	})

	ctx := todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
	completeSessionOnRecord(t, ctx, repo)
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

func TestBuildInstructionInlinesSmallPrompt(t *testing.T) {
	todo := &types.TODO{TODOFrontmatter: types.TODOFrontmatter{Title: "Fix the parser"}}
	prompt := "Think carefully.\n\n## Fix the parser\n\nThe parser drops trailing commas."

	got := buildInstruction([]*types.TODO{todo}, prompt, "/repo/.gavel/cmux/prompt-x.md", false)

	if !strings.HasPrefix(got, "# Fix the parser\n\n") {
		t.Fatalf("instruction should lead with the todo title: %q", got)
	}
	if !strings.Contains(got, "The parser drops trailing commas.") {
		t.Fatalf("small prompt should be inlined: %q", got)
	}
	if strings.Contains(got, "prompt-x.md") {
		t.Fatalf("small prompt should not reference the prompt file: %q", got)
	}
	if !strings.HasSuffix(got, "Implement all TODOs described above. When finished, stop and wait for verification.") {
		t.Fatalf("instruction should end with the run directive: %q", got)
	}
}

func TestBuildInstructionTruncatesLargePrompt(t *testing.T) {
	todo := &types.TODO{TODOFrontmatter: types.TODOFrontmatter{Title: "Big task"}}
	// Build a prompt well over the 10KB inline cap from whole lines.
	var sb strings.Builder
	for i := 0; sb.Len() < maxInlinePromptBytes*2; i++ {
		fmt.Fprintf(&sb, "line %d of the prompt body\n", i)
	}
	prompt := sb.String()
	path := "/repo/.gavel/cmux/prompt-big.md"

	got := buildInstruction([]*types.TODO{todo}, prompt, path, false)

	if !strings.Contains(got, "read "+path+" for the full prompt") {
		t.Fatalf("large prompt should reference the prompt file: %q", got)
	}
	body := strings.TrimPrefix(got, "# Big task\n\n")
	inlined := strings.SplitN(body, "\n\n... (prompt truncated", 2)[0]
	if len(inlined) > maxInlinePromptBytes {
		t.Fatalf("inlined body = %d bytes, want <= %d", len(inlined), maxInlinePromptBytes)
	}
	if !strings.HasSuffix(inlined, "body") {
		t.Fatalf("truncation should cut on a line boundary (ending a full line): %q", inlined[len(inlined)-20:])
	}
}

func TestBuildInstructionPlanMode(t *testing.T) {
	todo := &types.TODO{TODOFrontmatter: types.TODOFrontmatter{Title: "Plan it"}}
	got := buildInstruction([]*types.TODO{todo}, "do the thing", "/p.md", true)
	if !strings.Contains(got, "do NOT make any code changes — only plan") {
		t.Fatalf("plan instruction should forbid code changes: %q", got)
	}
	if strings.Contains(got, "Implement all TODOs described above") {
		t.Fatalf("plan instruction should not include the implement directive: %q", got)
	}
}

func TestPreviewInstructionMatchesRunPrompt(t *testing.T) {
	todo := &types.TODO{TODOFrontmatter: types.TODOFrontmatter{Title: "Fix the parser"}, MarkdownBody: "The parser drops trailing commas."}

	got := PreviewInstruction([]*types.TODO{todo}, "/repo", "high", false, false, "claude")

	if !strings.HasPrefix(got, "# Fix the parser\n\n") {
		t.Fatalf("preview should lead with the todo title: %q", got)
	}
	if !strings.Contains(got, EffortDirective("high")) {
		t.Fatalf("preview should include the effort directive: %q", got)
	}
	if !strings.Contains(got, "The parser drops trailing commas.") {
		t.Fatalf("preview should inline the todo body: %q", got)
	}
	if !strings.HasSuffix(got, "Implement all TODOs described above. When finished, stop and wait for verification.") {
		t.Fatalf("preview should end with the implement directive: %q", got)
	}
}

func TestPreviewInstructionNumbersMultipleTodos(t *testing.T) {
	group := []*types.TODO{
		{TODOFrontmatter: types.TODOFrontmatter{Title: "Fix the parser"}},
		{TODOFrontmatter: types.TODOFrontmatter{Title: "Fix the writer"}},
	}

	got := PreviewInstruction(group, "/repo", "medium", false, false, "claude")

	if !strings.HasPrefix(got, "# 2 Todo Items\n\n") {
		t.Fatalf("multi-todo preview should be titled with the count: %q", got)
	}
	if !strings.Contains(got, "## 1. Fix the parser") || !strings.Contains(got, "## 2. Fix the writer") {
		t.Fatalf("multi-todo preview should number each item: %q", got)
	}
}

func TestPreviewInstructionPlanSuffix(t *testing.T) {
	todo := &types.TODO{TODOFrontmatter: types.TODOFrontmatter{Title: "Plan it"}}
	got := PreviewInstruction([]*types.TODO{todo}, "/repo", "medium", true, false, "claude")
	if !strings.Contains(got, "do NOT make any code changes — only plan") {
		t.Fatalf("plan preview should forbid code changes: %q", got)
	}
	if strings.Contains(got, "Implement all TODOs described above") {
		t.Fatalf("plan preview should not include the implement directive: %q", got)
	}
}

func TestPreviewInstructionPriorSessionNote(t *testing.T) {
	todo := &types.TODO{
		TODOFrontmatter: types.TODOFrontmatter{Title: "Resume me", LLM: &types.LLM{SessionId: "sess-123"}},
	}

	fresh := PreviewInstruction([]*types.TODO{todo}, "/repo", "medium", false, false, "claude")
	if !strings.Contains(fresh, "sess-123") {
		t.Fatalf("a fresh claude run over a prior session should reference it: %q", fresh)
	}

	resumed := PreviewInstruction([]*types.TODO{todo}, "/repo", "medium", false, true, "claude")
	if strings.Contains(resumed, "A previous agent session") {
		t.Fatalf("a resume run should not add the prior-session note: %q", resumed)
	}
}

func TestPromptTitleCountsGroupItems(t *testing.T) {
	group := []*types.TODO{
		{TODOFrontmatter: types.TODOFrontmatter{Title: "First"}},
		{TODOFrontmatter: types.TODOFrontmatter{Path: types.StringOrSlice{"pkg/foo.go:12"}}},
	}
	if got := promptTitle(group); got != "2 Todo Items" {
		t.Fatalf("group promptTitle = %q, want %q", got, "2 Todo Items")
	}
}

func TestPromptTitleSingleUsesTitleWithPathFallback(t *testing.T) {
	if got := promptTitle([]*types.TODO{{TODOFrontmatter: types.TODOFrontmatter{Title: "Fix it"}}}); got != "Fix it" {
		t.Fatalf("single promptTitle = %q, want %q", got, "Fix it")
	}
	if got := promptTitle([]*types.TODO{{TODOFrontmatter: types.TODOFrontmatter{Path: types.StringOrSlice{"pkg/foo.go:12"}}}}); got != "pkg/foo.go:12" {
		t.Fatalf("single promptTitle fallback = %q, want %q", got, "pkg/foo.go:12")
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
	fakeClaudeHome(t)
	runner := newScreenRunner(repo)
	exec := NewCmuxExecutor(CmuxExecutorConfig{
		WorkDir:                 repo,
		Model:                   "opus",
		Effort:                  "medium",
		Timeout:                 2 * time.Second,
		Runner:                  runner.run,
		ScreenPollInterval:      time.Millisecond,
		ScreenStableDuration:    time.Millisecond,
		SessionLogPollInterval:  time.Millisecond,
		SessionLogAppearTimeout: time.Second,
	})

	ctx := todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
	completeSessionOnRecord(t, ctx, repo)
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
	if payload := agentSend[6]; !strings.Contains(payload, "claude --session-id ") || !strings.HasSuffix(payload, " --model opus") {
		t.Fatalf("agent send payload = %q, want ...claude --session-id <uuid> --model opus", payload)
	}
	if todo.LLM == nil || todo.LLM.SessionId == "" {
		t.Fatalf("expected session id recorded on todo, got %+v", todo.LLM)
	}
	if len(sends[1]) != 7 || sends[1][0] != "send" || sends[1][2] != "workspace:ws1" || sends[1][4] != "surface:sf1" {
		t.Fatalf("unexpected prompt send args: %#v", sends[1])
	}
	instruction := sends[1][6]
	if !strings.HasPrefix(instruction, "# Fix cmux") {
		t.Fatalf("instruction should lead with the todo title: %q", instruction)
	}
	if !strings.Contains(instruction, "body") || !strings.Contains(instruction, "Implement all TODOs described above") {
		t.Fatalf("instruction should inline the prompt body and run directive: %q", instruction)
	}
	if strings.Contains(instruction, "prompt-abc12345.md") {
		t.Fatalf("a small prompt should be inlined, not reference the prompt file: %q", instruction)
	}
	if strings.HasSuffix(instruction, "\n") {
		t.Fatalf("instruction should be submitted without a trailing newline: %q", instruction)
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
	fakeClaudeHome(t)
	agentSends := 0
	promptSends := 0
	runner := func(_ context.Context, _ string, _ string, _ time.Duration, args ...string) (string, error) {
		switch args[0] {
		case "list-workspaces":
			return `{"workspaces":[]}`, nil
		case "new-workspace":
			return "workspace:ws1", nil
		case "new-surface":
			return "surface:sf1", nil
		case "read-screen":
			if agentSends == 0 {
				return "shell ready\n$ ", nil
			}
			return "Claude ready\n> ", nil
		case "send":
			payload := args[len(args)-1]
			if strings.Contains(payload, "claude --") {
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
		Timeout:                 2 * time.Second,
		Runner:                  runner,
		SendAttempts:            2,
		SendRetryDelay:          time.Millisecond,
		ScreenPollInterval:      time.Millisecond,
		ScreenStableDuration:    time.Millisecond,
		SessionLogPollInterval:  time.Millisecond,
		SessionLogAppearTimeout: time.Second,
	})

	ctx := todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
	completeSessionOnRecord(t, ctx, repo)
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

// TestCmuxExecutorRepressesEnterUntilSessionStarts asserts that when the initial
// prompt's Enter is dropped (no session log, static surface) the executor
// re-presses Enter until the session log appears, then proceeds normally.
func TestCmuxExecutorRepressesEnterUntilSessionStarts(t *testing.T) {
	repo := t.TempDir()
	fakeClaudeHome(t)
	logPath := sessionLogFile(t, repo, "start-session")

	var pendingSubmit string
	agentSent := false
	promptSubmitted := false
	repressCount := 0
	runner := func(_ context.Context, _ string, _ string, _ time.Duration, args ...string) (string, error) {
		switch args[0] {
		case "list-workspaces":
			return `{"workspaces":[]}`, nil
		case "new-workspace":
			return "workspace:ws1", nil
		case "new-surface":
			return "surface:sf1", nil
		case "send":
			pendingSubmit = args[len(args)-1]
			return "ok", nil
		case "send-key":
			switch {
			case strings.Contains(pendingSubmit, "claude --"):
				agentSent = true
			case pendingSubmit != "":
				promptSubmitted = true
			default:
				// A re-press with nothing freshly typed: the first one finally
				// submits the prompt, so the session log is created here.
				repressCount++
				if repressCount == 1 {
					writeSessionLog(t, logPath, completedSessionLine)
				}
			}
			pendingSubmit = ""
			return "ok", nil
		case "read-screen":
			switch {
			case !agentSent:
				return "shell ready\n$ ", nil
			case !promptSubmitted:
				return "Claude ready\n> ", nil
			default:
				// Static while the prompt sits unsent, so the screen signal never
				// fires and the session-start check must rely on the jsonl.
				return "prompt waiting to be submitted\n> ", nil
			}
		}
		return "ok", nil
	}
	exec := NewCmuxExecutor(CmuxExecutorConfig{
		WorkDir:                 repo,
		SessionID:               "start-session",
		Timeout:                 2 * time.Second,
		Runner:                  runner,
		ScreenPollInterval:      time.Millisecond,
		ScreenStableDuration:    time.Millisecond,
		SessionLogPollInterval:  time.Millisecond,
		SessionLogAppearTimeout: time.Second,
		SessionStartRetryDelays: []time.Duration{time.Millisecond, time.Millisecond},
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
	if repressCount != 1 {
		t.Fatalf("Enter re-presses = %d, want 1", repressCount)
	}
}

// TestCmuxExecutorFailsWhenSessionLogNeverAppears asserts a claude run fails the
// run (rather than falling back to screen-idle detection) when the pre-generated
// session log never materializes.
func TestCmuxExecutorFailsWhenSessionLogNeverAppears(t *testing.T) {
	repo := t.TempDir()
	fakeClaudeHome(t)
	runner := newScreenRunner(repo)
	exec := NewCmuxExecutor(CmuxExecutorConfig{
		WorkDir:                 repo,
		SessionID:               "missing-session",
		Timeout:                 200 * time.Millisecond,
		Runner:                  runner.run,
		ScreenPollInterval:      time.Millisecond,
		ScreenStableDuration:    time.Millisecond,
		SessionLogPollInterval:  time.Millisecond,
		SessionLogAppearTimeout: 5 * time.Millisecond,
		SessionStartRetryDelays: []time.Duration{time.Millisecond, time.Millisecond},
	})

	// No session log is ever written, so the tailer reports errSessionLogNotFound.
	ctx := todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
	todo := &types.TODO{
		ID:              "abc123456789",
		TODOFrontmatter: types.TODOFrontmatter{Title: "Fix cmux", CWD: repo},
	}
	result, err := exec.ExecuteGroup(ctx, []*types.TODO{todo})
	if err == nil {
		t.Fatal("ExecuteGroup() error = nil, want a failure when the session log never appears")
	}
	if result == nil || result.Success {
		t.Fatalf("expected a failed result, got %#v", result)
	}
	if !strings.Contains(err.Error(), "missing-session") || !strings.Contains(err.Error(), "did not appear") {
		t.Fatalf("error should name the session and the missing log: %v", err)
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
		if strings.Contains(r.pendingSubmit, "claude --") {
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

// completedSessionLine is one assistant end_turn entry; a tailer that reads it
// reports the session complete.
const completedSessionLine = `{"type":"assistant","sessionId":"s","message":{"stop_reason":"end_turn","content":[{"type":"text","text":"done"}]}}`

// fakeClaudeHome redirects the claude projects directory (resolved from $HOME) to
// a temp dir so cmux runs read and write session logs inside the test sandbox
// instead of the developer's real ~/.claude.
func fakeClaudeHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

// sessionLogFile resolves the session log a cmux run tails for sessionID under
// workDir, creating its parent directory tree.
func sessionLogFile(t *testing.T, workDir, sessionID string) string {
	t.Helper()
	path, err := SessionLogPath(workDir, sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// completeSessionOnRecord writes a finished session log the moment the run records
// its session id (before it begins tailing), so a fresh cmux run reports
// completion from the log rather than the screen.
func completeSessionOnRecord(t *testing.T, ctx *todopkg.ExecutorContext, workDir string) {
	t.Helper()
	ctx.SetSessionIDHook(func(sid string) {
		writeSessionLog(t, sessionLogFile(t, workDir, sid), completedSessionLine)
	})
}

// appendSessionLineAfter appends line to path after delay, simulating the agent
// emitting a new turn once a resumed run (which seeks past the existing log) has
// begun tailing. Any write error surfaces on the returned channel.
func appendSessionLineAfter(path, line string, delay time.Duration) <-chan error {
	done := make(chan error, 1)
	go func() {
		time.Sleep(delay)
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			done <- err
			return
		}
		_, werr := f.WriteString(line + "\n")
		if cerr := f.Close(); werr == nil {
			werr = cerr
		}
		done <- werr
	}()
	return done
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
