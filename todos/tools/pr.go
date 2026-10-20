package tools

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/flanksource/captain/pkg/api"
)

// prTools are offered only to a run that pushes its turns to the TODO's linked
// PR: elsewhere there is no push for an addressed thread to be resolved at.
func (t *toolset) prTools() []api.ToolDefinition {
	if !t.opts.PRWrite {
		return nil
	}
	return []api.ToolDefinition{
		t.def("pr_thread_addressed", "Report a review thread of the linked PR that this turn's changes address. Gavel resolves it after the turn is pushed.",
			api.ToolPolicyAllow, object([]string{"threadId"}, map[string]any{"threadId": str("The thread's native threadId")}), t.threadAddressed),
	}
}

type threadInput struct {
	ThreadID string `json:"threadId"`
}

// threadAddressed only records the id. The turn's push validates it against the
// threads open when the run started, at the pushed head, before resolving it.
func (t *toolset) threadAddressed(_ context.Context, input map[string]any) (any, error) {
	var in threadInput
	if err := decode(input, &in); err != nil {
		return nil, err
	}
	if in.ThreadID == "" || in.ThreadID != strings.TrimSpace(in.ThreadID) {
		return nil, fmt.Errorf("threadId %q is not a thread id", in.ThreadID)
	}
	run := t.opts.Run
	run.mu.Lock()
	defer run.mu.Unlock()
	if slices.Contains(run.threads, in.ThreadID) {
		return nil, fmt.Errorf("thread %s is already reported addressed", in.ThreadID)
	}
	run.threads = append(run.threads, in.ThreadID)
	return fmt.Sprintf("thread %s will be resolved after this turn is pushed", in.ThreadID), nil
}
