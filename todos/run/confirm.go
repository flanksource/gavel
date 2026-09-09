package run

import (
	"context"

	"github.com/flanksource/clicky"
)

type concurrentChoice struct {
	Concurrent bool
	Label      string
}

func (c concurrentChoice) String() string { return c.Label }

// confirmSelectFunc routes the choice to clicky, which renders it on the TTY
// when one is attached and otherwise hands it to the installed interactive sink
// (the dashboard's prompt dialog). Tests override it.
var confirmSelectFunc = func(ctx context.Context, options []concurrentChoice, title string) (concurrentChoice, bool) {
	return clicky.PromptSelectCtx(ctx, options, clicky.PromptSelectOptions[concurrentChoice]{Title: title})
}

// ConfirmConcurrent asks whether to dispatch a run against a TODO that already
// has a live one. Declining — including a caller with nowhere to ask — leaves
// the dispatch refused, so an unattended run never doubles up by accident.
//
// Superseding is deliberately not offered here: the incumbent belongs to
// another process, and cancelling its database row would not stop the agent it
// is still driving. Stop that run where it is owned, or run alongside it.
func ConfirmConcurrent(ctx context.Context, conflict string) bool {
	choice, ok := confirmSelectFunc(ctx, []concurrentChoice{
		{Concurrent: false, Label: "Cancel this dispatch"},
		{Concurrent: true, Label: "Run alongside the existing run"},
	}, conflict)
	return ok && choice.Concurrent
}
