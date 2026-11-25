package main

import (
	"context"
	"io"

	"github.com/flanksource/clicky/prompt"
)

type commitFeedback struct {
	Context context.Context
	Output  io.Writer
}

func (f commitFeedback) lintAction() lintFindingsAction {
	if !prompt.PreferSink(f.Context) {
		return promptLintFindingsAction()
	}
	choices, ok := prompt.GlobalManager().Select(f.Context, "Lint reported violations. What now?", []string{
		"AI Fix: repair violations, then re-lint",
		"Continue commit once without lint checks",
		"Cancel commit",
	}, prompt.SelectOptions{})
	if !ok || len(choices) != 1 {
		return lintActionCancel
	}
	switch choices[0] {
	case 0:
		return lintActionAIFix
	case 1:
		return lintActionContinueOnce
	default:
		return lintActionCancel
	}
}
