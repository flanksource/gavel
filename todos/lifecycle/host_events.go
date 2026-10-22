package lifecycle

import (
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

// declaresVerify reports whether a workflow gives captain's verify registry
// anything to run. A bare `verify: {}` declares nothing.
func declaresVerify(workflow *api.Workflow) bool {
	if workflow == nil || workflow.Verify == nil {
		return false
	}
	v := workflow.Verify
	return strings.TrimSpace(v.Fixture) != "" || len(v.Commands) > 0 || len(v.Prompts) > 0
}

// handleEvent narrates the run into the transcript and the notification sink,
// and accounts the result event's usage.
func (h *Host) handleEvent(exec *todos.ExecutorContext, ev captainai.Event, execution *todos.ExecutionResult, todo *types.TODO, sawResult *bool, meta todos.RunStartMetadata) {
	transcript := exec.GetTranscript()
	switch ev.Kind {
	case captainai.EventText:
		if ev.Text == "" {
			return
		}
		transcript.AddExecutorMessage(truncate(ev.Text, 200), todos.EntryText, nil)
		exec.Notify(todos.Notification{Type: todos.NotifyProgress, Message: truncate(ev.Text, 100)})
	case captainai.EventThinking:
		transcript.AddExecutorMessage(ev.Text, todos.EntryThinking, nil)
		exec.Notify(todos.Notification{Type: todos.NotifyThinking, Message: truncate(ev.Text, 100)})
	case captainai.EventToolUse:
		action := toolSummary(ev)
		transcript.AddExecutorMessage(action, todos.EntryAction, map[string]any{"tool": ev.Tool})
		exec.Notify(todos.Notification{Type: todos.NotifyAction, Message: action})
	case captainai.EventPermission:
		message := "awaiting approval: " + toolSummary(ev)
		detail := map[string]any{"tool": ev.Tool}
		if ev.ApprovalID != "" {
			detail["approvalId"] = ev.ApprovalID
			detail["input"] = ev.Input
		}
		if ev.Request != nil {
			detail["kind"] = ev.Request.Kind
		}
		if ev.Reason != "" {
			message = "approval " + ev.Reason + ": " + ev.Tool
			detail["reason"] = ev.Reason
		}
		transcript.AddExecutorMessage(message, todos.EntryAction, detail)
		exec.Notify(todos.Notification{Type: todos.NotifyApproval, Message: message, Data: detail})
	case captainai.EventSystem:
		if ev.SessionID != "" {
			setSessionID(todo, ev.SessionID)
			exec.RecordSessionID(ev.SessionID)
			meta.SessionID = ev.SessionID
			execution.Runtime.SessionID = ev.SessionID
			exec.RecordRunStart(meta)
		}
		if ev.Text != "" {
			transcript.AddExecutorMessage(ev.Text, todos.EntryAction, map[string]any{"role": "system"})
			exec.Notify(todos.Notification{Type: todos.NotifyAction, Message: ev.Text})
		}
	case captainai.EventResult:
		*sawResult = true
		if !ev.Success {
			// A result that is not a success is an error even when the provider
			// attached no text: collect turns the message into a failed run, and a
			// blank one would have let the run read as succeeded.
			execution.ErrorMessage = firstNonEmpty(ev.Error, "agent reported an unsuccessful result")
		}
	case captainai.EventError:
		execution.ErrorMessage = ev.Error
		exec.Notify(todos.Notification{Type: todos.NotifyError, Message: ev.Error})
	}
}

func toolSummary(ev captainai.Event) string {
	for _, key := range []string{"command", "file_path", "path", "pattern", "query"} {
		if v, ok := ev.Input[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return ev.Tool + ": " + truncate(s, 120)
			}
		}
	}
	return ev.Tool
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
