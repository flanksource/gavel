package cmux

import (
	"fmt"
	"strings"
	"time"

	todopkg "github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// SendFeedback resumes the agent's live session with feedback (e.g. the failing
// test/lint summary from the post-completion check loop) by sending it to the
// still-open REPL on the surface the last ExecuteGroup launched, then waiting
// for the agent's next turn to end. It implements todos.FeedbackExecutor.
//
// It requires a prior claude run on this executor (a tailable session id);
// codex runs, which manage their own sessions, return an error so the loop
// degrades to reporting the failures without iterating.
func (e *CmuxExecutor) SendFeedback(ctx *todopkg.ExecutorContext, _ []*types.TODO, feedback string) (*todopkg.ExecutionResult, error) {
	start := time.Now()
	if e.lastSessionID == "" || e.lastSurface.SurfaceID == "" {
		err := fmt.Errorf("cmux: no live claude session to feed back into")
		return failedResult(e.Name(), start, err), err
	}
	feedback = strings.TrimSpace(feedback)
	if feedback == "" {
		err := fmt.Errorf("cmux: empty feedback")
		return failedResult(e.Name(), start, err), err
	}

	timeout := e.timeout()
	ctx.Logger.Infof("cmux: sending check feedback to session %s", e.lastSessionID)
	if err := e.sendSurfaceText(ctx, e.lastSurface.String(), e.lastSurface.SurfaceID, "check feedback", feedback); err != nil {
		return failedResult(e.Name(), start, err), err
	}

	// seekToEnd (resume=true) skips the prior turn's end_turn so we wait for the
	// turn the feedback triggers, not the one that already completed.
	_, completed, err := e.awaitSessionCompletion(ctx, e.lastSessionID, e.lastWorkDir, timeout, true, nil)
	if err != nil {
		return failedResult(e.Name(), start, err), err
	}
	if !completed {
		err := fmt.Errorf("claude session %s did not complete the feedback turn within %s", e.lastSessionID, timeout)
		return failedResult(e.Name(), start, err), err
	}
	ctx.Logger.Infof("cmux: session %s completed feedback turn", e.lastSessionID)
	return &todopkg.ExecutionResult{
		Success:      true,
		ExecutorName: e.Name(),
		Duration:     time.Since(start),
		Transcript:   ctx.GetTranscript(),
	}, nil
}

var _ todopkg.FeedbackExecutor = (*CmuxExecutor)(nil)
