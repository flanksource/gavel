package claude

import (
	"fmt"
	"strings"
	"time"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// SendFeedback resumes the inline claude session recorded on the TODO and asks
// it to fix the failures described in feedback (the post-completion check loop's
// test/lint summary), then commits whatever it changed. It implements
// todos.FeedbackExecutor.
//
// Unlike the cmux path, the inline agent is a one-shot subprocess, so each
// feedback round is a fresh `tsx agent.ts` run resumed via the recorded session
// id. It commits per round (like Execute) so the fixes persist for the next
// check pass; it does NOT stash, because the working tree already holds the
// agent's in-progress changes.
func (e *ClaudeExecutor) SendFeedback(ctx *todos.ExecutorContext, todosInGroup []*types.TODO, feedback string) (*todos.ExecutionResult, error) {
	result := &todos.ExecutionResult{
		ExecutorName: e.Name(),
		Transcript:   ctx.GetTranscript(),
	}
	startTime := time.Now()

	if len(todosInGroup) == 0 {
		return result, fmt.Errorf("no TODOs to feed back into")
	}
	feedback = strings.TrimSpace(feedback)
	if feedback == "" {
		return result, fmt.Errorf("empty feedback")
	}
	todo := todosInGroup[0]
	if todo.LLM == nil || todo.LLM.SessionId == "" {
		return result, fmt.Errorf("no prior claude session to resume for feedback")
	}

	agentDir, err := prepareAgentDir()
	if err != nil {
		return result, fmt.Errorf("failed to prepare agent dir: %w", err)
	}
	if err := ensureDependencies(agentDir); err != nil {
		return result, fmt.Errorf("failed to ensure dependencies: %w", err)
	}

	ctx.Notify(todos.Notification{
		Type:    todos.NotifyProgress,
		Message: "Resuming session to fix failing checks",
	})

	before, _ := gitSnapshot(e.config.WorkDir)
	if err := e.runAgent(ctx, agentDir, buildFeedbackPrompt(feedback), todo, result); err != nil {
		result.Duration = time.Since(startTime)
		result.ErrorMessage = err.Error()
		return result, err
	}

	if result.Success {
		sha, commitErr := gitCommitChanges(ctx.Context, e.config.WorkDir, todo, before)
		if commitErr != nil {
			ctx.Logger.Warnf("Failed to commit feedback changes: %v", commitErr)
		} else {
			result.CommitSHA = sha
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

func buildFeedbackPrompt(feedback string) string {
	return "The automated checks (tests and/or lint) failed after your last changes. " +
		"Fix the issues described below, then stop and wait for re-verification. " +
		"Address the root cause; do not disable or skip the failing checks.\n\n" + feedback
}

var _ todos.FeedbackExecutor = (*ClaudeExecutor)(nil)
