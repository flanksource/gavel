package cmux

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/flanksource/commons/logger"
	todopkg "github.com/flanksource/gavel/todos"
)

func testCtx() *todopkg.ExecutorContext {
	return todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
}

var testSurface = WorkspaceRef{WorkspaceID: "workspace:ws1", SurfaceID: "surface:sf1"}

func TestReplReadyRe(t *testing.T) {
	cases := []struct {
		name   string
		screen string
		want   bool
	}{
		{"bare prompt", "> ", true},
		{"unicode prompt", "❯ ", true},
		{"boxed prompt", "╭────────╮\n│ >      │\n╰────────╯", true},
		{"indented prompt", "  > ", true},
		{"shell dollar", "user@host:~$ ", false},
		{"plain banner", "Welcome to Claude Code\nStarting up…", false},
		{"prose with arrow midline", "Type > to send your message", false},
	}
	for _, tc := range cases {
		if got := replReadyRe.MatchString(tc.screen); got != tc.want {
			t.Errorf("%s: replReadyRe.MatchString(%q) = %v, want %v", tc.name, tc.screen, got, tc.want)
		}
	}
}

func TestDefaultSessionStartRetryDelaysEscalate(t *testing.T) {
	want := []time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second, 15 * time.Second}
	if len(defaultSessionStartRetryDelays) != len(want) {
		t.Fatalf("defaultSessionStartRetryDelays = %v, want %v", defaultSessionStartRetryDelays, want)
	}
	for i := range want {
		if defaultSessionStartRetryDelays[i] != want[i] {
			t.Fatalf("defaultSessionStartRetryDelays = %v, want %v", defaultSessionStartRetryDelays, want)
		}
	}
}

// timedRunner records each cmux invocation with the time it occurred so a test
// can assert the settle delay between the paste and the Enter.
type timedRunner struct {
	mu    sync.Mutex
	ops   []string
	times []time.Time
}

func (r *timedRunner) run(_ context.Context, _, _ string, _ time.Duration, args ...string) (string, error) {
	r.mu.Lock()
	r.ops = append(r.ops, args[0])
	r.times = append(r.times, time.Now())
	r.mu.Unlock()
	return "ok", nil
}

func TestSendSurfaceTextSettlesBetweenPasteAndEnter(t *testing.T) {
	runner := &timedRunner{}
	exec := NewCmuxExecutor(CmuxExecutorConfig{Runner: runner.run, SendSettleDelay: 50 * time.Millisecond})

	if err := exec.sendSurfaceText(testCtx(), "workspace:ws1", "surface:sf1", "prompt", "hello"); err != nil {
		t.Fatalf("sendSurfaceText() error = %v", err)
	}

	if len(runner.ops) != 2 || runner.ops[0] != "send" || runner.ops[1] != "send-key" {
		t.Fatalf("ops = %v, want [send send-key]", runner.ops)
	}
	gap := runner.times[1].Sub(runner.times[0])
	if gap < 45*time.Millisecond {
		t.Fatalf("paste→Enter gap = %s, want >= ~50ms settle delay", gap)
	}
}

func TestConfirmStartedInitialDetectsLogAppearance(t *testing.T) {
	repo := t.TempDir()
	fakeClaudeHome(t)
	logPath := sessionLogFile(t, repo, "sess")
	writeSessionLog(t, logPath, completedSessionLine)

	// read-screen returns the same content as postSend, so the only signal that can
	// fire is the log having appeared.
	exec := NewCmuxExecutor(CmuxExecutorConfig{Runner: constantScreenRunner("idle\n> ")})
	started, why := exec.confirmStarted(testCtx(), testSurface, submitConfirm{logPath: logPath}, "idle\n> ")
	if !started {
		t.Fatalf("confirmStarted() = false, want true once the log exists")
	}
	if why != "session log appeared" {
		t.Fatalf("confirmStarted reason = %q, want %q", why, "session log appeared")
	}
}

func TestConfirmStartedFeedbackDetectsLogGrowth(t *testing.T) {
	repo := t.TempDir()
	fakeClaudeHome(t)
	logPath := sessionLogFile(t, repo, "sess")
	writeSessionLog(t, logPath, completedSessionLine) // the prior turn
	base := fileSize(logPath)

	// read-screen matches the post-send baseline so the surface signal never fires,
	// isolating log-growth as the only confirmation signal.
	exec := NewCmuxExecutor(CmuxExecutorConfig{Runner: constantScreenRunner("idle prompt")})
	sc := submitConfirm{logPath: logPath, baseOffset: base, growth: true}

	if started, _ := exec.confirmStarted(testCtx(), testSurface, sc, "idle prompt"); started {
		t.Fatal("confirmStarted() = true before the log grew, want false")
	}

	appendSessionLine(t, logPath, `{"type":"user","message":{"content":[{"type":"text","text":"feedback"}]}}`)
	started, why := exec.confirmStarted(testCtx(), testSurface, sc, "idle prompt")
	if !started {
		t.Fatal("confirmStarted() = false after the log grew, want true")
	}
	if why != "session log grew past pre-send offset" {
		t.Fatalf("confirmStarted reason = %q, want growth reason", why)
	}
}

func TestConfirmStartedDetectsSurfaceAdvance(t *testing.T) {
	// No log on disk and growth disabled, so only a surface change can confirm.
	exec := NewCmuxExecutor(CmuxExecutorConfig{Runner: constantScreenRunner("claude is working…")})
	sc := submitConfirm{logPath: "/nonexistent/session.jsonl"}

	started, why := exec.confirmStarted(testCtx(), testSurface, sc, "prompt waiting to submit")
	if !started {
		t.Fatal("confirmStarted() = false, want true when the surface advanced past submission")
	}
	if why != "surface advanced past submission" {
		t.Fatalf("confirmStarted reason = %q, want surface reason", why)
	}
}

func TestWaitForREPLReadyDetectsPrompt(t *testing.T) {
	exec := NewCmuxExecutor(CmuxExecutorConfig{
		Runner:             constantScreenRunner("Claude ready\n> "),
		ScreenPollInterval: time.Millisecond,
	})
	got, err := exec.waitForREPLReady(testCtx(), testSurface, 2*time.Second, "shell ready\n$ ")
	if err != nil {
		t.Fatalf("waitForREPLReady() error = %v", err)
	}
	if got == "" || !replReadyRe.MatchString(got) {
		t.Fatalf("waitForREPLReady() = %q, want the ready screen matching the prompt", got)
	}
}

func TestWaitForREPLReadyFallsBackOnTimeout(t *testing.T) {
	// The surface never renders a recognizable prompt; readiness must fall back to
	// the screen-idle wait (a changed, stable screen) rather than failing.
	exec := NewCmuxExecutor(CmuxExecutorConfig{
		Runner:               constantScreenRunner("compiling, no prompt yet"),
		REPLReadyTimeout:     10 * time.Millisecond,
		ScreenPollInterval:   time.Millisecond,
		ScreenStableDuration: time.Millisecond,
	})
	got, err := exec.waitForREPLReady(testCtx(), testSurface, time.Second, "shell ready\n$ ")
	if err != nil {
		t.Fatalf("waitForREPLReady() fallback error = %v", err)
	}
	if got != "compiling, no prompt yet" {
		t.Fatalf("waitForREPLReady() fallback = %q, want the idle screen", got)
	}
}

// constantScreenRunner is a cmux Runner whose read-screen always returns screen
// and whose other commands succeed, for tests that drive surface logic directly.
func constantScreenRunner(screen string) Runner {
	return func(_ context.Context, _, _ string, _ time.Duration, args ...string) (string, error) {
		if args[0] == "read-screen" {
			return screen, nil
		}
		return "ok", nil
	}
}

// appendSessionLine appends one line to an existing session log on disk.
func appendSessionLine(t *testing.T, path, line string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.WriteString(line + "\n"); err != nil {
		t.Fatal(err)
	}
}
