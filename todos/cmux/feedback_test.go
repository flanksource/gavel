package cmux

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/flanksource/commons/logger"
	todopkg "github.com/flanksource/gavel/todos"
)

// TestSendFeedbackRepressesUntilTurnStarts asserts SendFeedback gets the same
// dropped-Enter resilience as the initial prompt: when the feedback's Enter is
// dropped (the session log does not grow), it re-presses Enter until the turn
// demonstrably starts, then waits for it to complete.
func TestSendFeedbackRepressesUntilTurnStarts(t *testing.T) {
	repo := t.TempDir()
	fakeClaudeHome(t)
	logPath := sessionLogFile(t, repo, "fb-sess")
	writeSessionLog(t, logPath, completedSessionLine) // the prior turn already in the log

	var mu sync.Mutex
	pending := ""
	repressed := 0
	runner := func(_ context.Context, _, _ string, _ time.Duration, args ...string) (string, error) {
		mu.Lock()
		defer mu.Unlock()
		switch args[0] {
		case "send":
			pending = args[len(args)-1]
		case "send-key":
			if args[len(args)-1] == "Enter" {
				if pending != "" {
					// The feedback submit: simulate a dropped Enter — the log does not grow.
					pending = ""
				} else {
					// A re-press finally starts the turn; grow the log past the pre-send
					// offset with the feedback's user line.
					repressed++
					if repressed == 1 {
						appendSessionLine(t, logPath, `{"type":"user","message":{"content":[{"type":"text","text":"fix the failing tests"}]}}`)
					}
				}
			}
		case "read-screen":
			return "static surface", nil // never advances; isolate the log-growth signal
		}
		return "ok", nil
	}

	exec := NewCmuxExecutor(CmuxExecutorConfig{
		WorkDir:                 repo,
		Runner:                  runner,
		SendSettleDelay:         time.Millisecond,
		SessionStartRetryDelays: []time.Duration{time.Millisecond, time.Millisecond},
		SessionLogPollInterval:  time.Millisecond,
		SessionLogAppearTimeout: time.Second,
		StallTimeout:            time.Minute, // must not interfere
		Timeout:                 2 * time.Second,
	})
	exec.lastSessionID = "fb-sess"
	exec.lastWorkDir = repo
	exec.lastSurface = testSurface

	// The feedback turn completes only after tailing begins — seekToEnd skips the
	// prior turn and the re-press user line, so the end_turn must arrive afterwards.
	appended := appendSessionLineAfter(logPath, completedSessionLine, 40*time.Millisecond)

	ctx := todopkg.NewExecutorContext(context.Background(), logger.StandardLogger(), nil)
	result, err := exec.SendFeedback(ctx, nil, "fix the failing tests")
	if err != nil {
		t.Fatalf("SendFeedback() error = %v", err)
	}
	if result == nil || !result.Success {
		t.Fatalf("unexpected result: %#v", result)
	}
	if err := <-appended; err != nil {
		t.Fatalf("append completed turn: %v", err)
	}

	mu.Lock()
	got := repressed
	mu.Unlock()
	if got != 1 {
		t.Fatalf("feedback re-presses = %d, want 1", got)
	}
}
