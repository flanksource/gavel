package cmux

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	todopkg "github.com/flanksource/gavel/todos"
)

// replReadyRe matches the claude REPL's input prompt — the box-drawn "> " / "❯ "
// line claude renders once it is ready to accept a message. It is a positive
// readiness signal, stronger than screen-idle, which can settle on a half-drawn
// startup banner before the prompt is actually accepting input.
var replReadyRe = regexp.MustCompile(`(?m)^\s*(?:[│|]\s*)?[>❯](?:\s|$)`)

// submitConfirm describes how a submit confirms it actually took (the agent began
// the turn). The initial prompt waits for the session log to appear; feedback
// waits for it to grow past the byte offset captured before the send. Both also
// accept the terminal surface advancing past the just-submitted screen.
type submitConfirm struct {
	logPath string
	// baseOffset is the session-log size captured before the send. When growth is
	// set, confirmation requires the log to grow beyond it (feedback resumes an
	// existing log); otherwise the log merely appearing confirms (initial start).
	baseOffset int64
	growth     bool
}

// submitAndConfirm pastes text onto the surface, presses Enter, and confirms the
// submission actually started the turn, re-pressing Enter on the configured
// back-off when neither the session log nor the surface shows progress. cmux
// occasionally drops the Enter (REPL initializing, paste mode), leaving the text
// typed but unsent; this is the shared resilience the initial prompt and feedback
// both rely on. If the turn still hasn't started after the last re-press it
// returns nil and lets the downstream wait surface the loud failure rather than
// masking it here.
func (e *CmuxExecutor) submitAndConfirm(ctx *todopkg.ExecutorContext, ref WorkspaceRef, label, text string, sc submitConfirm) error {
	if err := e.sendSurfaceText(ctx, ref.String(), ref.SurfaceID, label, text); err != nil {
		return err
	}
	return e.confirmSubmitted(ctx, ref, label, sc)
}

// confirmSubmitted re-presses Enter until the submit is confirmed (or the
// back-off is exhausted). The post-send screen is the baseline the surface signal
// is compared against: if the Enter was dropped the typed text sits here
// unchanged; if it took, the screen advances past it.
func (e *CmuxExecutor) confirmSubmitted(ctx *todopkg.ExecutorContext, ref WorkspaceRef, label string, sc submitConfirm) error {
	postSend := e.readScreen(ctx, ref)
	if started, why := e.confirmStarted(ctx, ref, sc, postSend); started {
		ctx.Logger.V(1).Infof("cmux: %s confirmed (%s)", label, why)
		return nil
	}

	// Wait, re-check, then re-press only if still not started, so a submit whose
	// Enter merely took a moment to register isn't sent a spurious extra Enter.
	delays := e.sessionStartRetryDelays()
	for i, delay := range delays {
		ctx.Logger.Infof("cmux: %s not confirmed yet; waiting %s before re-pressing Enter (attempt %d/%d)", label, delay, i+1, len(delays))
		if err := sleepContext(ctx, delay); err != nil {
			return err
		}
		if started, why := e.confirmStarted(ctx, ref, sc, postSend); started {
			ctx.Logger.Infof("cmux: %s confirmed while waiting to re-press Enter (%s)", label, why)
			return nil
		}
		ctx.Logger.V(1).Infof("cmux command: cmux send-key --workspace %q --surface %q Enter", ref.String(), ref.SurfaceID)
		if err := e.client.SendKeySurface(ctx, ref.String(), ref.SurfaceID, "Enter"); err != nil {
			if ctx.Err() != nil {
				return err
			}
			ctx.Logger.Warnf("cmux: %s Enter re-press failed: %v", label, err)
		}
	}

	if started, why := e.confirmStarted(ctx, ref, sc, postSend); started {
		ctx.Logger.Infof("cmux: %s confirmed after re-pressing Enter (%s)", label, why)
		return nil
	}
	ctx.Logger.Warnf("cmux: %s not confirmed after re-pressing Enter %d time(s); relying on downstream timeout", label, len(delays))
	return nil
}

// confirmStarted reports whether the submit started the turn, from the session
// jsonl (authoritative — claude writes it as the turn progresses) and, as a
// fallback, the surface having advanced past the post-send baseline.
func (e *CmuxExecutor) confirmStarted(ctx *todopkg.ExecutorContext, ref WorkspaceRef, sc submitConfirm, postSend string) (bool, string) {
	if sc.growth {
		if fileSize(sc.logPath) > sc.baseOffset {
			return true, "session log grew past pre-send offset"
		}
	} else if _, err := os.Stat(sc.logPath); err == nil {
		return true, "session log appeared"
	}
	screen := e.readScreen(ctx, ref)
	if screen != "" && postSend != "" && screen != postSend {
		return true, "surface advanced past submission"
	}
	return false, ""
}

// sendSurfaceText pastes text onto the surface and presses Enter, retrying the
// whole paste+Enter on transient cmux errors. A settle delay separates the paste
// from the Enter so a fast paste→Enter doesn't submit a half-applied buffer or
// get swallowed while the surface is still in paste mode.
func (e *CmuxExecutor) sendSurfaceText(ctx *todopkg.ExecutorContext, workspaceRef, surfaceRef, label, text string) error {
	attempts := e.sendAttempts()
	delay := e.sendRetryDelay()
	settle := e.sendSettleDelay()
	text = strings.TrimRight(text, "\r\n")
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if attempt > 1 {
			ctx.Logger.V(1).Infof("cmux: waiting %s before retrying %s send", delay, label)
			if err := sleepContext(ctx, delay); err != nil {
				return err
			}
		}
		ctx.Logger.Infof("cmux: sending %s to workspace %s surface %s (attempt %d/%d)", label, workspaceRef, surfaceRef, attempt, attempts)
		ctx.Logger.V(1).Infof("cmux command: cmux send --workspace %q --surface %q -- <%s>", workspaceRef, surfaceRef, label)
		ctx.Logger.V(1).Infof("cmux command: cmux send-key --workspace %q --surface %q Enter", workspaceRef, surfaceRef)
		ctx.Logger.V(2).Infof("cmux send payload:\n%s", text)
		if err := e.client.SendSurface(ctx, workspaceRef, surfaceRef, text); err != nil {
			lastErr = err
			if ctx.Err() != nil {
				return err
			}
			if attempt < attempts {
				ctx.Logger.Warnf("cmux: %s send attempt %d/%d failed: %v; retrying in %s", label, attempt, attempts, err, delay)
			} else {
				ctx.Logger.Warnf("cmux: %s send attempt %d/%d failed: %v", label, attempt, attempts, err)
			}
			continue
		}
		if err := sleepContext(ctx, settle); err != nil {
			return err
		}
		if err := e.client.SendKeySurface(ctx, workspaceRef, surfaceRef, "Enter"); err != nil {
			lastErr = err
			if ctx.Err() != nil {
				return err
			}
			if attempt < attempts {
				ctx.Logger.Warnf("cmux: %s enter attempt %d/%d failed: %v; retrying in %s", label, attempt, attempts, err, delay)
			} else {
				ctx.Logger.Warnf("cmux: %s enter attempt %d/%d failed: %v", label, attempt, attempts, err)
			}
			continue
		}
		ctx.Logger.Infof("cmux: sent %s to workspace %s surface %s", label, workspaceRef, surfaceRef)
		return nil
	}
	return fmt.Errorf("send cmux %s after %d attempts: %w", label, attempts, lastErr)
}

// waitForREPLReady polls the surface until the claude input prompt appears,
// confirming the REPL will accept the initial prompt. On timeout it falls back to
// the screen-idle wait so a REPL whose prompt we failed to recognize (a theme
// change, a wrapped banner) still proceeds rather than failing the run.
func (e *CmuxExecutor) waitForREPLReady(ctx *todopkg.ExecutorContext, ref WorkspaceRef, timeout time.Duration, baseline string) (string, error) {
	readyTimeout := e.replReadyTimeout(timeout)
	poll := e.screenPollInterval()
	base := normalizeScreen(baseline)
	deadline := time.Now().Add(readyTimeout)
	for {
		screen := e.readScreen(ctx, ref)
		if screen != "" && screen != base && replReadyRe.MatchString(screen) {
			ctx.Logger.V(1).Infof("cmux: claude REPL ready (input prompt detected)")
			return screen, nil
		}
		if time.Now().After(deadline) {
			ctx.Logger.V(1).Infof("cmux: claude REPL prompt not detected within %s; falling back to screen-idle", readyTimeout)
			return e.waitForScreenIdle(ctx, ref, "after agent launch", timeout, baseline, true)
		}
		if err := sleepContext(ctx, poll); err != nil {
			return "", err
		}
	}
}

func (e *CmuxExecutor) sendSettleDelay() time.Duration {
	if e.config.SendSettleDelay > 0 {
		return e.config.SendSettleDelay
	}
	return defaultSendSettleDelay
}

func (e *CmuxExecutor) replReadyTimeout(timeout time.Duration) time.Duration {
	rt := e.config.REPLReadyTimeout
	if rt <= 0 {
		rt = defaultREPLReadyTimeout
	}
	if timeout > 0 && rt > timeout {
		rt = timeout
	}
	return rt
}

// fileSize returns the size of path in bytes, or 0 when it cannot be stat-ed (a
// missing log reads as size 0, the natural "not started" baseline).
func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
