package prompting

import "github.com/flanksource/clicky"

// Prepare waits for the global clicky task manager to finish and stops its
// renderer before an interactive or AI prompt takes over the terminal.
func Prepare() {
	clicky.WaitForGlobalCompletionSilent()
}
